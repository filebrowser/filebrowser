package cnc

// DPRNT capture — the Haas DPRNT[…] macro statement emits formatted text
// over the RS-232 port at execution time. Operators use it for in-cycle
// probing output (size readings, hole positions) and homemade telemetry.
// Without a listener the bytes sit in the bridge's receive buffer and
// eventually contaminate the next Q-code response.
//
// This file owns the scavenger that pulls those bytes between G-code
// line writes. Lives behind Machine.DPRNTCapture so installs that don't
// use DPRNT don't pay the per-line read cost.
//
// Lifecycle: a dprntBuffer is created when a job starts (if enabled),
// hung off the job struct, and drained between writes inside run().
// Anything not framed as a Q-code response (STX … ETB) is treated as
// DPRNT and emitted on the WS feed.

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"
)

// dprntSidecarPath returns the on-disk path where the streamer should
// persist DPRNT output for one job. Naming: "<absPath>.<job-id>.dprnt.log".
// Lives next to the NC source so it's reachable through the regular
// file browser without a separate UI surface.
func dprntSidecarPath(absPath, jobID string) string {
	return absPath + "." + jobID + ".dprnt.log"
}

// dprntSink returns the per-line emit callback the scavenger calls. It
// fans out to both the WS feed and (when set) the per-job sidecar file.
// File writes are best-effort; failures log a warn but don't unwind
// the streaming loop.
func (s *Streamer) dprntSink(j *job) func(text string) {
	return func(text string) {
		s.emit(Event{Type: "dprnt", Text: text})
		if j.dprntLog == nil {
			return
		}
		if _, err := fmt.Fprintln(j.dprntLog, text); err != nil {
			s.logf("warn", "dprnt sidecar write failed: %v", err)
			// Don't try again — keep the activity log clean and let the
			// WS event continue to carry the data.
			_ = j.dprntLog.Close()
			j.dprntLog = nil
		}
	}
}

// dprntScavengeDeadline is how long we'll wait per scavenge before
// giving up and returning to the line-write loop. Kept tiny because
// run() services scavenges between every line — a streamed program
// at 9600 baud sees ~one scavenge per ~30-100ms anyway.
const dprntScavengeDeadline = 3 * time.Millisecond

// dprntMaxLineBytes bounds a single DPRNT line. Haas DPRNT[…] output is
// human-scale (a dozen formatted numbers); anything past 4 KiB without
// a newline is almost certainly cross-talk, so we flush the buffer.
const dprntMaxLineBytes = 4 * 1024

// dprntBuffer accumulates partial bytes across multiple scavenges. The
// scavenger emits whenever a complete line is seen (terminated by \r,
// \n, or \r\n).
type dprntBuffer struct {
	buf bytes.Buffer
}

// scavengeOnce performs a single non-blocking read on conn. Returns the
// number of bytes read. Pulls any DPRNT-looking lines out of the
// accumulated buffer and emits them; discards Q-code-framed crosstalk
// silently (with a debug log so the streamer log shows what was lost).
//
// Safe to call between line writes inside run(). Sets and clears its
// own read deadline so the streamer's outer deadline is untouched.
func (d *dprntBuffer) scavengeOnce(conn net.Conn, emit func(text string), debug func(level, msg string)) (int, error) {
	prev := time.Now().Add(dprntScavengeDeadline)
	if err := conn.SetReadDeadline(prev); err != nil {
		return 0, err
	}
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }()

	tmp := make([]byte, 1024)
	n, err := conn.Read(tmp)
	if n > 0 {
		d.buf.Write(tmp[:n])
		d.drain(emit, debug)
	}
	// Timeouts on a non-blocking scavenge are expected and not an error.
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return n, nil
		}
		return n, err
	}
	return n, nil
}

// drain pulls every complete line out of the buffer. Q-code STX … ETB
// frames are discarded (they're cross-talk that should have been read
// by exchangeOnConn). Bare ASCII lines are emitted as DPRNT.
//
// Caller holds no locks; the scavenger is single-threaded inside run().
func (d *dprntBuffer) drain(emit func(text string), debug func(level, msg string)) {
	for {
		raw := d.buf.Bytes()
		if len(raw) == 0 {
			return
		}
		// Q-code frame? Drain through the trailing ETB and continue.
		if idx := bytes.IndexByte(raw, stxByte); idx >= 0 {
			etb := bytes.IndexByte(raw[idx:], etbByte)
			if etb < 0 {
				// Frame is mid-flight — wait for more bytes.
				return
			}
			// Emit nothing; just consume.
			if debug != nil {
				debug("warn", "DPRNT scavenger discarded a stray Q-code frame (cross-talk)")
			}
			d.consume(idx + etb + 1)
			continue
		}
		// Line terminator? Emit everything up to (and including) it.
		nl := bytes.IndexAny(raw, "\r\n")
		if nl < 0 {
			// No complete line yet. Cap the buffer.
			if d.buf.Len() > dprntMaxLineBytes {
				d.buf.Reset()
				if debug != nil {
					debug("warn", "DPRNT scavenger flushed an oversized line buffer")
				}
			}
			return
		}
		line := strings.TrimRight(string(raw[:nl+1]), "\r\n")
		d.consume(nl + 1)
		// Collapse runs of \r and \n that follow the matched terminator
		// so a CRLF doesn't emit twice.
		for d.buf.Len() > 0 {
			next := d.buf.Bytes()[0]
			if next != '\r' && next != '\n' {
				break
			}
			d.consume(1)
		}
		if line != "" {
			emit(line)
		}
	}
}

// consume removes n bytes from the front of the buffer. bytes.Buffer's
// Next() returns a slice view we'd have to copy; this is the simplest
// safe way to discard.
func (d *dprntBuffer) consume(n int) {
	if n <= 0 {
		return
	}
	rest := d.buf.Bytes()
	if n >= len(rest) {
		d.buf.Reset()
		return
	}
	// Manual shift; bytes.Buffer doesn't expose a "skip first n" op.
	keep := make([]byte, len(rest)-n)
	copy(keep, rest[n:])
	d.buf.Reset()
	d.buf.Write(keep)
}
