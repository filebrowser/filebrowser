# DPRNT log capture (Z-14) — research + design

Status: design-only. Z-14 is hardware-gated; needs an operator at the
controller running a real probe / DPRNT-emitting program to validate.

User flagged this as "Z-14: DPRNT log capture — needs hardware to
verify; refactors the streaming socket's read path."

---

## What DPRNT is

`DPRNT[<format>]` is a Haas (and most other CNC) macro statement that
formats a string from macro variables and **emits it on the active
serial channel**. Operators use it to:

- Log probe results during in-process measurement
- Dump tool-life or process telemetry mid-program
- Send custom messages to a paired logger/data-collection PC

When the controller is in DNC mode (Setting 14 = 9 = "DNC via
RS-232"), DPRNT output goes back **over the same RS-232 link we're
streaming G-code to**. The Waveshare bridge is bidirectional, so
those bytes show up on our TCP socket interleaved with our writes.

Wire format (from public Haas DPRNT documentation):
```
DPRNT[PROBE*Z*#5063[44]]
→ controller writes ASCII: "PROBE Z -1.2345\r\n"
```

The exact framing varies — DPRNT can include any printable ASCII +
CRLF terminator. Unlike Q-code responses, DPRNT output is NOT wrapped
in STX/ETB. So our existing framing-required parser (`stripEchoAndFraming`)
correctly rejects DPRNT lines as not-a-Q-response, but we currently
have no path to **capture** them.

## Why the read path needs a refactor

Today's streamer.run():
```go
for scanner.Scan() {
    select {
    case req := <-j.queryCh:
        s.serviceQuery(conn, req)        // writes Q + reads framed response
    default:
    }
    line := scanner.Text()
    conn.Write([]byte(line + "\r\n"))    // writes G-code; never reads
}
```

We never read the socket between writes. DPRNT output piles up in the
TCP receive buffer until the next serviceQuery or until the socket
buffer overflows. PR #36 paused polling during streams to keep
Q-responses clean; that means DPRNT bytes accumulate for the entire
job and we likely lose them when the connection closes.

Fix shape: a dedicated read goroutine that:
- Owns the socket's read side for the duration of a job
- Buffers incoming bytes between line writes
- Emits any `\r\n`-terminated line that doesn't match a Q-response
  shape as a `dprnt` event on the WS feed
- Keeps a rolling tail in the streamer (last ~1000 lines) for
  postmortem retrieval

Tricky bit: when the streaming side wants to fire a Q-code (rare
during streams now, but `/api/cnc/qcode` still goes through the run
loop), the read goroutine has to step aside while serviceQuery does
its write+read dance. Use a per-conn read mutex.

## Proposed architecture

```go
type dprntReader struct {
    conn       net.Conn
    mu         sync.Mutex          // held while a Q-response read is in flight
    onLine     func(string)        // emit callback (DPRNT event + tail buffer)
    stop       chan struct{}
    lastBytes  *ringBuffer         // rolling 64 KB for diagnostics
}

func (r *dprntReader) loop() {
    buf := make([]byte, 4096)
    for {
        select { case <-r.stop: return; default: }
        n, err := r.conn.Read(buf)  // blocks
        if n > 0 {
            r.mu.Lock()
            r.lastBytes.Write(buf[:n])
            // Split on CRLF, emit each non-empty non-prompt line
            r.mu.Unlock()
        }
        if err != nil { return }
    }
}
```

When the run loop wants to do a Q-code round-trip:

```go
func (s *Streamer) serviceQueryWithDPRNT(conn net.Conn, req *queryReq) {
    s.dprntReader.mu.Lock()
    defer s.dprntReader.mu.Unlock()
    // exchangeOnConn proceeds; the dprntReader is paused on its next iteration
    raw, err := exchangeOnConn(conn, req.q, req.macroV)
    ...
}
```

Limitation: the dprntReader holds the read side, so when Q-response
bytes arrive interleaved with DPRNT output, we have to disambiguate.
Heuristic: STX-framed payload → Q-response, anything else → DPRNT.
That works because DPRNT explicitly never emits STX.

## Event shape

Add `dprnt` to the event types in `cnc/events.go`:

```go
type Event struct {
    Type   string  `json:"type"`   // "line" | "status" | "metric" | "log" | "dprnt"
    ...
    DPRNT  string  `json:"dprnt,omitempty"`
}
```

Frontend renders DPRNT as a separate stream in the Activity log,
maybe with a different color tag, and exposes a "Save log" button
that downloads the rolling tail buffer.

For probe workflows specifically, the operator usually wants to:
- See live DPRNT messages as they fire
- Save the full DPRNT log of a run to a `.dprnt.txt` file in the same
  share folder as the NC

Both fall out of the rolling tail + per-line WS emit.

## Edge cases

- **Lines split across reads**: buffer until newline. Standard.
- **Prompt characters** (`>`, `?`): filter — these are interactive-mode
  noise from controllers with Setting 187 echo on.
- **Q-response races DPRNT**: covered by the read-mutex pattern; a
  Q-response read sees only its STX-framed payload because the
  dprntReader is paused while the mutex is held.
- **Read deadline**: dprntReader uses no deadline (blocks indefinitely);
  serviceQuery sets/clears its own deadline as today.
- **Buffer overflow**: TCP receive buffer is fine; we'd only overflow
  if we never read, which the dprntReader guarantees against.

## Hardware validation plan

When the user is at the controller:

1. Pick a probe macro that DPRNTs (Renishaw cycles all do).
2. Add a small marker line to a test program:
   ```
   DPRNT[ZINC-MARKER*1234]
   ```
3. Send the program through filebrowser-NC.
4. Confirm the Activity log shows a `dprnt` row with text
   `ZINC-MARKER 1234`.
5. Confirm the rolling tail download has the same line.

If the marker doesn't show up, likely causes:
- Setting 14 not in DNC mode (DPRNT goes to a different channel)
- Setting 26 (DPRNT comma swap) interfering
- Read mutex not releasing between Q-codes

## Open questions

- **Persist DPRNT to disk?** Yes — write to
  `<share>/dprnt/<job-id>.txt` as the job runs. Append-only. Operators
  shouldn't have to remember to download.
- **Per-line vs batched WS emit?** Per-line. DPRNT is human-readable
  and operators want immediate feedback during probing.
- **Other brands?** Fanuc has `POPEN/PCLOS/DPRNT` family with similar
  semantics. The MachineProtocol abstraction (see MULTI_MACHINE_DESIGN.md)
  should let each brand provide its own filter/parser for "what is and
  isn't a Q-response on this socket".
