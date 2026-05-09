package cnc

// Q-code wire protocol — Go port of haas-dashboard's haas_bridge.py.
//
// Wire format (verified against a TM-2P over a Waveshare RS-232↔TCP):
//   send: "?Q<code>[ <var>]\r\n"
//   recv: "<echo>\r\r\n\x02<payload>\x17\r\n>\n"
// The payload is wrapped in STX (0x02) … ETB (0x17). The trailing ">?" /
// ">\n" is the *next* idle prompt — unreliable as an end marker. ETB is
// the truth; idle-after-data is the fallback for controls that don't
// frame.
//
// Pre-conditions on the Haas:
//   Setting 143 (Machine Data Collect) — ON. Without it Q-codes are no-ops.
//   Setting 187 (Echo) — either way is fine; the parser tolerates both.

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	stxByte = 0x02
	etbByte = 0x17

	// queryTimeout bounds a single Q-code round-trip (dial + write + read + close).
	queryTimeout = 3 * time.Second
	// idleAfterData backs off the read once we've already seen bytes — covers
	// the small handful of Haas controls that emit data without a closing ETB.
	idleAfterData = 1 * time.Second
)

// QueryResult mirrors the dashboard's contract so /api/cnc/qcode is a
// drop-in replacement for haas-dashboard's POST /api/query.
type QueryResult struct {
	Q          int     `json:"q"`
	Var        *int    `json:"var,omitempty"`
	Raw        string  `json:"raw"`
	Value      string  `json:"value"`
	Parsed     any     `json:"parsed,omitempty"`
	OK         bool    `json:"ok"`
	Error      string  `json:"error,omitempty"`
	DurationMs float64 `json:"duration_ms"`
}

// payloadFor builds the bytes we put on the wire. macroVar is optional;
// pass nil for plain queries like Q104 (mode) or Q500 (program/parts).
func payloadFor(qCode int, macroVar *int) []byte {
	var b strings.Builder
	b.WriteString("?Q")
	b.WriteString(strconv.Itoa(qCode))
	if macroVar != nil {
		b.WriteByte(' ')
		b.WriteString(strconv.Itoa(*macroVar))
	}
	b.WriteString("\r\n")
	return []byte(b.String())
}

// transientQuery opens a one-shot TCP connection, sends the query,
// reads the response, closes. Used when no streaming job holds the
// socket. Mirrors HaasBridge._round_trip in haas_bridge.py.
func transientQuery(host string, port, qCode int, macroVar *int) *QueryResult {
	t0 := time.Now()
	res := &QueryResult{Q: qCode, Var: macroVar}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, queryTimeout)
	if err != nil {
		res.Error = fmt.Sprintf("dial %s: %v", addr, err)
		res.DurationMs = sinceMs(t0)
		return res
	}
	defer conn.Close()

	raw, err := exchangeOnConn(conn, qCode, macroVar)
	res.DurationMs = sinceMs(t0)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.Raw = raw
	v := stripEchoAndFraming(raw)
	if err := validateResponseShape(qCode, macroVar, v); err != nil {
		// Keep Raw for postmortems but DON'T expose Value/Parsed —
		// rawValue() in the UI would otherwise render a contaminated
		// frame as if it were truth.
		res.Error = err.Error()
		return res
	}
	res.Value = v
	res.Parsed = parseValue(v, qCode, macroVar)
	res.OK = true
	return res
}

// exchangeOnConn writes one query and reads one framed response on an
// already-open connection. The streaming loop calls this between line
// writes so /api/cnc/qcode keeps working during a job (the Waveshare
// only accepts one client at a time, so we share the streaming socket).
func exchangeOnConn(conn net.Conn, qCode int, macroVar *int) (string, error) {
	deadline := time.Now().Add(queryTimeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return "", err
	}
	defer func() {
		_ = conn.SetDeadline(time.Time{}) // clear so the streaming loop isn't capped
	}()

	if _, err := conn.Write(payloadFor(qCode, macroVar)); err != nil {
		return "", fmt.Errorf("write: %w", err)
	}

	br := bufio.NewReader(conn)
	var buf bytes.Buffer
	chunk := make([]byte, 512)
	for {
		// Once we've already buffered something, shorten the per-read
		// deadline so an ETB-less control doesn't keep us blocked all
		// the way to queryTimeout.
		if buf.Len() > 0 {
			next := time.Now().Add(idleAfterData)
			if next.Before(deadline) {
				_ = conn.SetReadDeadline(next)
			}
		}
		n, err := br.Read(chunk)
		if n > 0 {
			buf.Write(chunk[:n])
			if bytes.IndexByte(buf.Bytes(), etbByte) >= 0 {
				return buf.String(), nil
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				if buf.Len() > 0 {
					return buf.String(), nil
				}
				return "", fmt.Errorf("read: %w", err)
			}
			// Timeout? If we have buffered bytes treat it as idle-done,
			// otherwise propagate.
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				if buf.Len() > 0 {
					return buf.String(), nil
				}
				return "", fmt.Errorf("no response within %s", queryTimeout)
			}
			return "", err
		}
	}
}

// queryReq is the channel-passed request carried from Streamer.Query
// into the streaming run loop when a job is active.
type queryReq struct {
	q       int
	macroV  *int
	ctx     context.Context
	respCh  chan *QueryResult
}

// frameRe matches `\x02 … \x17` — the canonical Haas STX-framed payload.
var frameRe = regexp.MustCompile("\x02([^\x17]*)\x17")

// stripEchoAndFraming pulls the meaningful payload out of a raw response.
// Requires STX/ETB framing — historically we tolerated unframed line-mode
// responses for non-Haas controllers, but in practice the only way that
// path triggered was when the read buffer picked up G-code line bytes
// during a stream and we'd return them as if they were a Q response.
// Returning "" on no-frame lets validateResponseShape reject cleanly.
func stripEchoAndFraming(raw string) string {
	if raw == "" {
		return ""
	}
	if m := frameRe.FindStringSubmatch(raw); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// validateResponseShape returns nil if `value` plausibly answers the
// query we sent, or an error describing the mismatch. Catches bridge
// cross-talk: the Waveshare buffers RS-232 responses across TCP
// connection boundaries, so a fresh dial can return data left over
// from someone else's last query (or our own previous one, before
// minQuerySpacing kicked in). Without this check those stale frames
// would parse and display as truth.
//
// Rules are intentionally narrow — match the obvious tag word for
// each Q-code per haas_bridge.py + Haas docs. Macro queries are
// stricter: the response must carry the exact macro var number we
// asked for.
func validateResponseShape(qCode int, macroVar *int, value string) error {
	if value == "" {
		return fmt.Errorf("empty response")
	}
	upper := strings.ToUpper(value)
	switch qCode {
	case 104: // Mode — single token like MEM/MDI/JOG
		if strings.Contains(upper, "PROGRAM") ||
			strings.Contains(upper, "MACRO") ||
			strings.Contains(upper, "PARTS") ||
			strings.Contains(upper, "LAST CYCLE") ||
			strings.Contains(upper, "TOOL") {
			return fmt.Errorf("Q104 (mode) got cross-talk frame: %q", value)
		}
	case 201:
		if !strings.Contains(upper, "TOOL") {
			return fmt.Errorf("Q201 expected TOOL prefix, got %q", value)
		}
	case 303:
		if !strings.Contains(upper, "LAST CYCLE") && !strings.Contains(upper, "PREVIOUS CYCLE") {
			return fmt.Errorf("Q303 expected LAST CYCLE prefix, got %q", value)
		}
	case 402:
		if !strings.Contains(upper, "PARTS") {
			return fmt.Errorf("Q402 expected PARTS prefix, got %q", value)
		}
	case 500:
		if !strings.Contains(upper, "PROGRAM") {
			return fmt.Errorf("Q500 expected PROGRAM prefix, got %q", value)
		}
	case 600:
		if !strings.Contains(upper, "MACRO") {
			return fmt.Errorf("Q600 expected MACRO prefix, got %q", value)
		}
		if macroVar != nil {
			tag := strconv.Itoa(*macroVar)
			// Tolerate either ", N," or ",N," spacing the bridge happens
			// to emit. The MACRO frame is comma-delimited.
			if !strings.Contains(value, ", "+tag+",") && !strings.Contains(value, ","+tag+",") {
				return fmt.Errorf("Q600 macro var mismatch: asked %d, got %q", *macroVar, value)
			}
		}
	}
	return nil
}

// parseValue is best-effort structured parsing — same shape contract as
// the Python dashboard returns. Callers that need more detail should
// look at the raw + value fields.
func parseValue(value string, qCode int, macroVar *int) any {
	if value == "" {
		return nil
	}
	parts := splitAndTrim(value, ",")
	switch {
	case qCode == 500:
		// PROGRAM,<O#>,<status>,PARTS,<n>
		if len(parts) >= 5 &&
			strings.EqualFold(parts[0], "PROGRAM") &&
			strings.EqualFold(parts[3], "PARTS") {
			return map[string]string{
				"program": parts[1],
				"status":  parts[2],
				"parts":   parts[4],
			}
		}
		// fallback pair-based dict
		out := map[string]string{}
		for i := 0; i+1 < len(parts); i += 2 {
			k := strings.ReplaceAll(strings.ToLower(parts[i]), " ", "_")
			out[k] = parts[i+1]
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case qCode == 600 && macroVar != nil:
		// "MACRO, <var>, <value>" — value is usually last; walk backward.
		for i := len(parts) - 1; i >= 0; i-- {
			if v, ok := parseNumber(parts[i]); ok {
				return v
			}
		}
		return nil
	}
	if len(parts) >= 2 {
		if v, ok := parseNumber(parts[len(parts)-1]); ok {
			return v
		}
		return parts[len(parts)-1]
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return nil
}

func parseNumber(s string) (any, bool) {
	if strings.Contains(s, ".") {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f, true
		}
		return nil, false
	}
	if i, err := strconv.Atoi(s); err == nil {
		return i, true
	}
	return nil, false
}

func splitAndTrim(s, sep string) []string {
	out := []string{}
	for _, p := range strings.Split(s, sep) {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func sinceMs(t time.Time) float64 {
	return float64(time.Since(t).Microseconds()) / 1000.0
}
