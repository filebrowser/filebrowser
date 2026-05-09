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
}

// Streamer is the long-lived singleton. One per process. Holds the
// single-job invariant the user asked for: any Start while a job is
// running returns ErrJobAlreadyRunning, never spawns a second TCP
// connection to the Waveshare.
type Streamer struct {
	settings settingsReader

	mu  sync.Mutex // guards job
	job *job       // nil when idle

	// last* are kept across jobs so /status can show the most
	// recent error after the streamer goes idle. Read with mu held.
	lastError string
}

type job struct {
	id          string
	displayPath string // share-relative, what /status echoes back
	absPath     string // absolute filesystem path
	startedAt   time.Time
	lineCurrent atomic.Int64
	lineTotal   int
	cancel      context.CancelFunc
	done        chan struct{} // closed when the streaming goroutine exits
}

// New builds the singleton.
func New(s settingsReader) *Streamer {
	return &Streamer{settings: s}
}

// Start kicks off a streaming job. absPath must already be path-validated
// against the user's scope — the streamer doesn't second-guess that.
// displayPath is share-relative and what shows up in /status.
//
// Returns ErrJobAlreadyRunning if a job is already in flight.
func (s *Streamer) Start(absPath, displayPath string) (*Status, error) {
	set, err := s.settings.Get()
	if err != nil {
		return nil, err
	}
	if set.Cnc.HaasHost == "" {
		return nil, ErrConfigMissing
	}
	host := set.Cnc.HaasHost
	port := set.Cnc.HaasPort
	if port == 0 {
		port = settings.DefaultHaasPort
	}

	lineTotal, err := countLines(absPath)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
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
	}
	s.job = j
	s.lastError = ""
	s.mu.Unlock()

	go s.run(ctx, j, host, port)

	return s.Status(), nil
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
	j.cancel()
	<-j.done
	return true
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
	return st
}

// run is the worker goroutine. Owns the TCP socket for the duration of
// the job and clears s.job on exit so subsequent Starts succeed.
func (s *Streamer) run(ctx context.Context, j *job, host string, port int) {
	defer close(j.done)
	defer func() {
		s.mu.Lock()
		s.job = nil
		s.mu.Unlock()
	}()

	// net.JoinHostPort handles bracketing IPv6 addresses; "%s:%d" doesn't.
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		s.recordError(fmt.Errorf("dial %s: %w", addr, err))
		return
	}
	defer conn.Close()

	f, err := os.Open(j.absPath)
	if err != nil {
		s.recordError(fmt.Errorf("open %s: %w", j.absPath, err))
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Allow long G-code lines just in case.
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	w := bufio.NewWriter(conn)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			_ = w.Flush()
			return
		default:
		}

		line := strings.TrimRight(scanner.Text(), "\r\n")
		if _, err := w.WriteString(line); err != nil {
			s.recordError(fmt.Errorf("write line %d: %w", j.lineCurrent.Load()+1, err))
			return
		}
		if _, err := w.WriteString("\r\n"); err != nil {
			s.recordError(fmt.Errorf("write CRLF: %w", err))
			return
		}
		// Flush per line so cancel + line counters track real wire progress.
		// Phase 2.2 will replace this with XON/XOFF-aware pacing.
		if err := w.Flush(); err != nil {
			s.recordError(fmt.Errorf("flush: %w", err))
			return
		}
		j.lineCurrent.Add(1)
	}
	if err := scanner.Err(); err != nil {
		s.recordError(fmt.Errorf("read source: %w", err))
	}
}

func (s *Streamer) recordError(err error) {
	s.mu.Lock()
	s.lastError = err.Error()
	s.mu.Unlock()
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
