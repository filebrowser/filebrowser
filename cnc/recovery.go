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

// markerStateDir resolves the directory holding per-machine markers.
// Honors RUNTIME_DIRECTORY-style $CNC_STATE_DIR for systemd installs;
// falls back to UserCacheDir / TempDir.
func markerStateDir() string {
	dir := os.Getenv("CNC_STATE_DIR")
	if dir != "" {
		return dir
	}
	if cache, err := os.UserCacheDir(); err == nil {
		return filepath.Join(cache, "filebrowser-cnc")
	}
	return filepath.Join(os.TempDir(), "filebrowser-cnc")
}

// markerPathFor returns the per-machine marker path. Two machines
// running concurrent jobs each get their own marker file so a crash
// of one doesn't block the other.
func markerPathFor(machineID string) string {
	if machineID == "" {
		machineID = "default"
	}
	return filepath.Join(markerStateDir(), "active_job_"+machineID+".json")
}

func writeMarkerFor(machineID string, j *job) error {
	p := markerPathFor(machineID)
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

func clearMarkerFor(machineID string) {
	_ = os.Remove(markerPathFor(machineID))
	// Also clear the pre-multi-machine path so an old marker from
	// before this binary doesn't haunt forever.
	legacy := filepath.Join(markerStateDir(), "active_job.json")
	_ = os.Remove(legacy)
}

// readMarkerFor is non-fatal — missing marker is normal idle state.
// A malformed marker is treated like a present one (better to surface
// the "didn't end cleanly" warning than ignore it because of a parse
// glitch). Falls back to the pre-multi-machine path so an existing
// orphan marker from before the migration still triggers the prompt.
func readMarkerFor(machineID string) *activeJobMarker {
	for _, p := range []string{
		markerPathFor(machineID),
		filepath.Join(markerStateDir(), "active_job.json"), // legacy
	} {
		buf, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var m activeJobMarker
		if err := json.Unmarshal(buf, &m); err != nil {
			return &activeJobMarker{}
		}
		return &m
	}
	return nil
}

// AckRecovery clears the pending-recovery flag and removes the marker.
// Idempotent — safe to call when no recovery is pending.
func (s *Streamer) AckRecovery() {
	s.mu.Lock()
	s.pendingRecovery = nil
	s.mu.Unlock()
	clearMarkerFor(s.machineID)
}
