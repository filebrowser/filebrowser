package cnc

// Crash-recovery (Z-15) — when filebrowser-NC restarts mid-job, the
// next Start call refuses to auto-resume and the operator has to
// explicitly acknowledge before a new job can begin. This guards
// against "the Pi rebooted while my part was being cut" leaving the
// system in a state where a fresh job silently picks up after a
// previous job's failure (which could rewind the program, double-cut,
// or run on a wrong WCS without anyone noticing).
//
// Mechanism:
//   - On Start: write a marker JSON to <state>/active_job.json.
//   - On clean exit (run loop returns OR Stop is called): remove it.
//   - On streamer New(): if the marker is present, set the
//     pendingRecovery flag. Status() reflects it; Start refuses with
//     ErrRecoveryPending until AckRecovery clears it.

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// ErrRecoveryPending signals a previous job didn't end cleanly. The
// HTTP layer maps this to 409 — the operator must hit
// /api/cnc/recovery/ack before kicking off a new job.
var ErrRecoveryPending = errors.New("previous job did not end cleanly — acknowledge recovery before starting a new one")

// activeJobMarker is the persisted shape of the in-flight marker.
// Time is RFC3339 so it survives the JSON round-trip cleanly.
type activeJobMarker struct {
	JobID       string    `json:"job_id"`
	DisplayPath string    `json:"display_path"`
	AbsPath     string    `json:"abs_path"`
	StartedAt   time.Time `json:"started_at"`
	LineTotal   int       `json:"line_total"`
}

// markerPath honors the same RUNTIME_DIRECTORY systemd-friendly env var
// the watcher uses, but on a per-instance fallback so dev runs don't
// step on each other.
func markerPath() string {
	dir := os.Getenv("CNC_STATE_DIR")
	if dir == "" {
		// Default to a user-writable spot that survives a restart but
		// not a reboot — we WANT the marker to survive a process
		// crash, but if the host rebooted the operator's already
		// going to be re-checking work, so persistence is a bonus
		// rather than a requirement.
		if cache, err := os.UserCacheDir(); err == nil {
			dir = filepath.Join(cache, "filebrowser-cnc")
		} else {
			dir = filepath.Join(os.TempDir(), "filebrowser-cnc")
		}
	}
	return filepath.Join(dir, "active_job.json")
}

func writeMarker(j *job) error {
	p := markerPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	m := activeJobMarker{
		JobID:       j.id,
		DisplayPath: j.displayPath,
		AbsPath:     j.absPath,
		StartedAt:   j.startedAt,
		LineTotal:   j.lineTotal,
	}
	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, buf, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func clearMarker() {
	_ = os.Remove(markerPath())
}

// readMarker is non-fatal — a missing marker is the normal idle state.
// A malformed marker is treated like a present one (better to surface
// the "didn't end cleanly" warning than ignore it because of a parse
// glitch).
func readMarker() *activeJobMarker {
	buf, err := os.ReadFile(markerPath())
	if err != nil {
		return nil
	}
	var m activeJobMarker
	if err := json.Unmarshal(buf, &m); err != nil {
		// Treat as orphan-present so the operator gets the prompt.
		return &activeJobMarker{}
	}
	return &m
}

// AckRecovery clears the pending-recovery flag and removes the marker.
// Idempotent — safe to call when no recovery is pending.
func (s *Streamer) AckRecovery() {
	s.mu.Lock()
	s.pendingRecovery = nil
	s.mu.Unlock()
	clearMarker()
}
