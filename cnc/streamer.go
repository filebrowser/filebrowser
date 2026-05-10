// Package cnc holds the singleton streamer that owns the TCP socket to
// the Waveshare RS-232↔TCP bridge. See docs/INTEGRATION_WITH_HAAS_DASHBOARD.md
// for the wider design — Pi-as-broker, single-job lock, multiplexed
// streaming + Q-code queries.
//
// Phase 2.1 (this file) wires Start / Stop / Status with a basic
// line-by-line writer. XON/XOFF flow control and Q-code multiplexing
// land in 2.2; the live WS event stream in 2.3.
package cnc

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/filebrowser/filebrowser/v2/settings"
)

// ErrJobAlreadyRunning signals a Start call collided with an in-flight job.
// The HTTP layer maps this to 409 Conflict.
var ErrJobAlreadyRunning = errors.New("a CNC job is already running")

// ErrConfigMissing signals the Machine settings tab hasn't been filled
// in yet (HaasHost is empty). The HTTP layer maps this to 400.
var ErrConfigMissing = errors.New("haas host/port not configured (Settings → Machine)")

// settingsReader is the slice of *storage.Storage we actually need.
// Keeping it as an interface here means the streamer is trivially
// fakeable in tests without dragging the whole storage package in.
type settingsReader interface {
	Get() (*settings.Settings, error)
}

// Status is a point-in-time snapshot of streamer state. Serialized as
// the response body of GET /api/cnc/status.
type Status struct {
	Running       bool      `json:"running"`
	JobID         string    `json:"job_id,omitempty"`
	FilePath      string    `json:"file_path,omitempty"` // share-relative
	LineCurrent   int64     `json:"line_current"`
	LineTotal     int       `json:"line_total"`
	StartedAt     time.Time `json:"started_at,omitempty"`
	HaasOK        bool      `json:"haas_ok"`
	HaasLastError string    `json:"haas_last_error,omitempty"`

	// Z-15: a previous instance left an active-job marker behind.
	// Frontend renders the warning until the operator POSTs to
	// /api/cnc/recovery/ack.
	RecoveryPending  bool   `json:"recovery_pending,omitempty"`
	RecoveryFilePath string `json:"recovery_file_path,omitempty"`
}

// Streamer is the long-lived singleton. One per process. Holds the
// single-job invariant the user asked for: any Start while a job is
// running returns ErrJobAlreadyRunning, never spawns a second TCP
// connection to the Waveshare.
type Streamer struct {
	settings  settingsReader
	machineID string // immutable; identifies which Machine in settings this Streamer owns

	mu  sync.Mutex // guards job + pendingRecovery
	job *job       // nil when idle

	// last* are kept across jobs so /status can show the most
	// recent error after the streamer goes idle. Read with mu held.
	lastError string

	// pendingRecovery is set when New() finds an orphaned active-job
	// marker (i.e. previous instance crashed mid-job). Start refuses
	// while this is non-nil — see recovery.go (Z-15). Cleared by
	// AckRecovery().
	pendingRecovery *activeJobMarker

	// Event broadcast — see events.go. Independent mutex so a slow
	// subscriber's drop logic can't block job state changes.
	subsMu sync.Mutex
	subs   []*subscriber

	// queryMu serializes all transient Q-code queries against the
	// Waveshare bridge. The bridge accepts ONE TCP client at a time;
	// without this, the aggregator's 16 polling goroutines fan out
	// concurrent dials and responses cross-contaminate. lastQueryAt
	// gates a min-spacing pause between back-to-back queries so the
	// RS-232 side has room to drain.
	queryMu      sync.Mutex
	lastQueryAt  time.Time
}

// queryQueueDepth caps in-flight queries against the streaming socket.
// 4 is plenty — the dashboard polls one Q-code at a time.
const queryQueueDepth = 4

// minQuerySpacing is the floor on time between consecutive Q-code
// round-trips against the Waveshare RS-232↔TCP bridge. The bridge
// only serves one TCP client at a time AND the underlying RS-232 link
// has finite bandwidth — without spacing the aggregator's 16 polling
// goroutines pile responses up in the RS-232 receive buffer and they
// bleed across connection boundaries (mode==program, work.X==machine.X,
// G54 X==Y==Z, etc). Mirrors haas_bridge.py's MIN_QUERY_SPACING.
const minQuerySpacing = 150 * time.Millisecond

type job struct {
	id          string
	displayPath string // share-relative, what /status echoes back
	absPath     string // absolute filesystem path
	startedAt   time.Time
	lineCurrent atomic.Int64
	lineTotal   int
	cancel      context.CancelFunc
	done        chan struct{} // closed when the streaming goroutine exits
	queryCh     chan *queryReq
}

// New builds a Streamer for one Machine. machineID identifies which
// entry in settings.Cnc.Machines this Streamer owns. Picks up any
// active-job marker left behind by a previous instance for THIS
// machine and stashes it in pendingRecovery so /status surfaces it
// and Start refuses until ack.
//
// One Streamer per Machine — see cnc/registry.go.
func New(s settingsReader, machineID string) *Streamer {
	st := &Streamer{settings: s, machineID: machineID}
	if m := readMarkerFor(machineID); m != nil {
		st.pendingRecovery = m
	}
	return st
}

// resolveMachine looks up THIS streamer's Machine in settings. Returns
// the Machine + the resolved port (defaulting if zero) + nil err if
// found; ErrConfigMissing if not. All Streamer methods that need
// host/port go through here so the answer is always live (operator
// edits in settings take effect on the next call).
func (s *Streamer) resolveMachine() (settings.Machine, int, error) {
	set, err := s.settings.Get()
	if err != nil {
		return settings.Machine{}, 0, err
	}
	m, ok := set.Cnc.MachineByID(s.machineID)
	if !ok {
		return settings.Machine{}, 0, ErrConfigMissing
	}
	if m.Host == "" {
		return settings.Machine{}, 0, ErrConfigMissing
	}
	port := m.Port
	if port == 0 {
		port = settings.DefaultHaasPort
	}
	return m, port, nil
}

// Start kicks off a streaming job. absPath must already be path-validated
// against the user's scope — the streamer doesn't second-guess that.
// displayPath is share-relative and what shows up in /status.
//
// Returns ErrJobAlreadyRunning if a job is already in flight.
func (s *Streamer) Start(absPath, displayPath string) (*Status, error) {
	m, port, err := s.resolveMachine()
	if err != nil {
		return nil, err
	}
	host := m.Host

	lineTotal, err := countLines(absPath)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	if s.pendingRecovery != nil {
		s.mu.Unlock()
		return nil, ErrRecoveryPending
	}
	if s.job != nil {
		s.mu.Unlock()
		return nil, ErrJobAlreadyRunning
	}

	id, err := newJobID()
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	j := &job{
		id:          id,
		displayPath: displayPath,
		absPath:     absPath,
		startedAt:   time.Now().UTC(),
		lineTotal:   lineTotal,
		cancel:      cancel,
		done:        make(chan struct{}),
		queryCh:     make(chan *queryReq, queryQueueDepth),
	}
	s.job = j
	s.lastError = ""
	s.mu.Unlock()

	// Persist the active-job marker AFTER the job is in s.job so a
	// crash between Start returning and the marker write would still
	// be observed via the inhibit (defensive — the window is tiny).
	if err := writeMarkerFor(s.machineID, j); err != nil {
		// Don't fail Start over a marker write — log via lastError so
		// /status surfaces it but the job runs. Operators usually
		// care more about "the spindle is on" than "we couldn't
		// write a recovery hint."
		s.mu.Lock()
		s.lastError = "marker write failed: " + err.Error()
		s.mu.Unlock()
	}

	st := s.Status()
	s.emit(Event{Type: "status", Status: st})
	s.logf("info", "start job %s: %s (%d lines) → %s:%d", j.id, j.displayPath, j.lineTotal, host, port)
	go s.run(ctx, j, host, port)

	return st, nil
}

// Stop cancels the in-flight job (if any). Returns true if a job was
// running. Blocks until the streaming goroutine exits so callers know
// the socket is freed.
func (s *Streamer) Stop() bool {
	s.mu.Lock()
	j := s.job
	s.mu.Unlock()
	if j == nil {
		return false
	}
	s.logf("info", "stop requested for job %s at line %d/%d", j.id, j.lineCurrent.Load(), j.lineTotal)
	j.cancel()
	<-j.done
	return true
}

// CheckBridge does a TCP dial to the configured host:port. Returns
// (ok, latencyMs, addr, err). Does NOT send any Q-code — only verifies
// network reachability of the Waveshare. Caller should skip during
// streaming to avoid contending with the job's socket.
func (s *Streamer) CheckBridge() (bool, float64, string, error) {
	m, port, err := s.resolveMachine()
	if err != nil {
		return false, 0, "", err
	}
	addr := net.JoinHostPort(m.Host, strconv.Itoa(port))
	t0 := time.Now()
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	latency := sinceMs(t0)
	if err != nil {
		return false, latency, addr, err
	}
	_ = conn.Close()
	return true, latency, addr, nil
}

// CheckController sends one Q104 (mode) and validates the response
// frame. Returns (ok, latencyMs, mode, err). Routes through Query so
// queryMu serialization + minQuerySpacing apply.
func (s *Streamer) CheckController(ctx context.Context) (bool, float64, string, error) {
	t0 := time.Now()
	res, err := s.Query(ctx, 104, nil)
	latency := sinceMs(t0)
	if err != nil {
		return false, latency, "", err
	}
	if !res.OK {
		return false, latency, "", fmt.Errorf("%s", res.Error)
	}
	return true, latency, res.Value, nil
}

// IsRunning returns true while a streaming job holds the socket. The
// aggregator uses this to pause polling during streams — Q-code reads
// on the streaming socket pick up G-code line bytes / flow-control
// chatter instead of clean Q responses, so polling produces garbage
// and risks consuming bytes the controller depends on (XON/XOFF).
func (s *Streamer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.job != nil
}

// Status returns a snapshot. Safe to call concurrently with Start/Stop.
func (s *Streamer) Status() *Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := &Status{
		HaasOK:        s.lastError == "",
		HaasLastError: s.lastError,
	}
	if s.job != nil {
		st.Running = true
		st.JobID = s.job.id
		st.FilePath = s.job.displayPath
		st.LineCurrent = s.job.lineCurrent.Load()
		st.LineTotal = s.job.lineTotal
		st.StartedAt = s.job.startedAt
	}
	if s.pendingRecovery != nil {
		st.RecoveryPending = true
		st.RecoveryFilePath = s.pendingRecovery.DisplayPath
	}
	return st
}

// Query runs one Q-code round-trip. When idle, opens a transient TCP
// connection. When a streaming job is in flight, queues the request on
// the streaming worker so we don't try to open a second client to the
// Waveshare (it typically only accepts one).
//
// Honors ctx for cancellation while the request is queued; once the
// streaming worker has the conn locked the queryTimeout (3s default)
// bounds the read.
func (s *Streamer) Query(ctx context.Context, qCode int, macroVar *int) (*QueryResult, error) {
	m, port, err := s.resolveMachine()
	if err != nil {
		return nil, err
	}
	host := m.Host

	s.mu.Lock()
	j := s.job
	s.mu.Unlock()
	if j == nil {
		return s.runTransient(ctx, host, port, qCode, macroVar)
	}

	req := &queryReq{
		q:      qCode,
		macroV: macroVar,
		ctx:    ctx,
		respCh: make(chan *QueryResult, 1),
	}
	select {
	case j.queryCh <- req:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-j.done:
		// Job ended while we were trying to enqueue. Fall back to a
		// transient query — through runTransient so we still get the
		// queryMu serialization vs. any other concurrent callers.
		return s.runTransient(ctx, host, port, qCode, macroVar)
	}

	select {
	case res := <-req.respCh:
		return res, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// runTransient is the gated path for transient (no-job) queries. The
// bridge serves one TCP client at a time, so we serialize on queryMu
// and enforce a min-spacing pause so the RS-232 side can drain between
// round-trips. ctx cancellation aborts the spacing wait but never an
// in-flight transientQuery (those have their own queryTimeout).
func (s *Streamer) runTransient(ctx context.Context, host string, port, qCode int, macroVar *int) (*QueryResult, error) {
	s.queryMu.Lock()
	defer s.queryMu.Unlock()
	if !s.lastQueryAt.IsZero() {
		if wait := minQuerySpacing - time.Since(s.lastQueryAt); wait > 0 {
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	res := transientQuery(host, port, qCode, macroVar)
	s.lastQueryAt = time.Now()
	return res, nil
}

// run is the worker goroutine. Owns the TCP socket for the duration of
// the job and clears s.job on exit so subsequent Starts succeed.
//
// Each iteration: optionally service one Q-code query (so /api/cnc/qcode
// stays responsive during a stream), then write the next line. Per-line
// write keeps cancel + line counters honest; flow control (XON/XOFF) is
// the next iteration if testing on the real Haas needs it.
func (s *Streamer) run(ctx context.Context, j *job, host string, port int) {
	defer close(j.done)
	defer func() {
		s.mu.Lock()
		s.job = nil
		s.mu.Unlock()
		// Marker is cleared on clean exit (whether the source EOFed,
		// the user clicked Stop, or the dial/write failed and we're
		// returning the error). A crash between here and the next
		// New() leaves the marker in place — exactly the case Z-15
		// is designed to catch.
		clearMarkerFor(s.machineID)
		// Emit the idle status event AFTER s.job has been cleared so
		// subscribers see the post-job snapshot, not a stale "running".
		s.emit(Event{Type: "status", Status: s.Status()})
	}()

	// net.JoinHostPort handles bracketing IPv6 addresses; "%s:%d" doesn't.
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	s.logf("info", "dialing bridge %s…", addr)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		s.recordError(fmt.Errorf("dial %s: %w", addr, err))
		return
	}
	defer conn.Close()
	s.logf("info", "bridge connected, opening %s", j.displayPath)

	f, err := os.Open(j.absPath)
	if err != nil {
		s.recordError(fmt.Errorf("open %s: %w", j.absPath, err))
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Allow long G-code lines just in case.
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Drain at most one pending Q-code query before the next line
		// so the stream still makes forward progress under heavy
		// polling. queryQueueDepth bounds the worst case.
		select {
		case req := <-j.queryCh:
			s.serviceQuery(conn, req)
		default:
		}

		line := strings.TrimRight(scanner.Text(), "\r\n")
		if _, err := conn.Write([]byte(line + "\r\n")); err != nil {
			s.recordError(fmt.Errorf("write line %d: %w", j.lineCurrent.Load()+1, err))
			return
		}
		n := j.lineCurrent.Add(1)
		s.emit(Event{Type: "line", N: n, Text: line})
		// Periodic progress beacon — every line on the per-line WS feed
		// is great for the ticker but would drown a 100k-line program
		// in the system log AND the activity panel. Every 100 lines we
		// also push to the log channel so journal + activity panel get
		// regular "we're at line N" updates.
		if n%100 == 0 {
			s.logf("info", "wrote line %d/%d", n, j.lineTotal)
		}
	}
	if err := scanner.Err(); err != nil {
		s.recordError(fmt.Errorf("read source: %w", err))
	} else {
		s.logf("info", "stream complete: %d lines sent", j.lineCurrent.Load())
	}

	// Drain any queries enqueued during the final line so callers don't
	// hang waiting for a response.
	for {
		select {
		case req := <-j.queryCh:
			s.serviceQuery(conn, req)
		default:
			return
		}
	}
}

// serviceQuery executes one Q-code request on the streaming socket and
// fulfils its response channel. Errors are returned via QueryResult.OK
// = false rather than thrown — the streaming run loop must keep going.
func (s *Streamer) serviceQuery(conn net.Conn, req *queryReq) {
	if err := req.ctx.Err(); err != nil {
		req.respCh <- &QueryResult{Q: req.q, Var: req.macroV, Error: err.Error()}
		return
	}
	t0 := time.Now()
	raw, err := exchangeOnConn(conn, req.q, req.macroV)
	res := &QueryResult{
		Q:          req.q,
		Var:        req.macroV,
		DurationMs: sinceMs(t0),
	}
	if err != nil {
		res.Error = err.Error()
		req.respCh <- res
		return
	}
	res.Raw = raw
	v := stripEchoAndFraming(raw)
	if err := validateResponseShape(req.q, req.macroV, v); err != nil {
		// Keep Raw, drop Value/Parsed — see same pattern in qcode.go.
		res.Error = err.Error()
		req.respCh <- res
		return
	}
	res.Value = v
	res.Parsed = parseValue(v, req.q, req.macroV)
	res.OK = true
	req.respCh <- res
}

func (s *Streamer) recordError(err error) {
	s.mu.Lock()
	s.lastError = err.Error()
	s.mu.Unlock()
	s.logf("error", "%v", err)
}

// logf emits one structured log event AND writes to the standard
// logger (journalctl when running under systemd). Two destinations
// because operators want different things — `journalctl -u filebrowser`
// for permanent record, and the WS feed for live UI without an SSH.
func (s *Streamer) logf(level, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("[cnc:%s] %s", level, msg)
	s.emit(Event{Type: "log", Level: level, Msg: msg})
}

func newJobID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	n := 0
	for scanner.Scan() {
		n++
	}
	return n, scanner.Err()
}
