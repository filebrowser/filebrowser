package cnc

// Per-machine NC job queue — staging area between "I want to send
// this file" and "send it now." See docs/QUEUE_DESIGN.md for the
// flow. Persists across restarts as a JSON file per machine,
// loaded eagerly at Registry boot.
//
// Shared across operators (not per-user) because the queue represents
// the machine's pending workload — a second operator viewing /machine
// sees what the first operator has staged.

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// randHex returns n hex chars from crypto/rand. Used only for queue
// item IDs — collision-resistant at queue cadence; if rand fails we
// fall back to a timestamp suffix rather than bricking enqueue.
func randHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())[:n]
	}
	return hex.EncodeToString(buf)[:n]
}

// QueueItem is one entry on a machine's queue. Identity is the
// generated ID; FilePath is user-scope-relative so the same path
// works for any operator with read access to that share.
type QueueItem struct {
	ID   string `json:"id"`
	// FilePath is the user-scope-relative path the queue knows about.
	// Resolved to an absolute on disk per operator at send time —
	// shared queue, multi-user share.
	FilePath string `json:"file_path"`
	// JobName is the parent folder name. Used in the UI as the
	// human-readable "Job" line under the file name. Empty for files
	// at the share root.
	JobName string `json:"job_name,omitempty"`
	// OnumberHint is the O-number parsed from the NC source at enqueue
	// time, used to auto-match a running program to its queue row.
	// Stored as a normalized string ("O00057") so a leading-zero
	// difference between operator typing and CAM output doesn't break
	// the join.
	OnumberHint string `json:"onumber_hint,omitempty"`
	// SizeBytes is the file size at enqueue time — handy UI metadata.
	SizeBytes int64 `json:"size_bytes,omitempty"`
	// State is one of "queued" / "sending" / "running". Only one item
	// per machine can be "sending" or "running" at a time.
	State string `json:"state"`
	// Method (mem/dnc) — set when the operator clicks send.
	Method string `json:"method,omitempty"`
	// AddedAt is the enqueue timestamp.
	AddedAt time.Time `json:"added_at"`
	// LineCurrent / LineTotal mirror the streamer's progress for the
	// item currently sending/running. Updated by status emit.
	LineCurrent int `json:"line_current,omitempty"`
	LineTotal   int `json:"line_total,omitempty"`
}

// QueueStates the UI cares about. Constants here so the JSON shape is
// stable; spec calls them "queued", "sending", "running".
const (
	QueueStateQueued  = "queued"
	QueueStateSending = "sending"
	QueueStateRunning = "running"
)

// QueueStore is per-Registry singleton. One in-memory map keyed by
// machine ID; each mutation persists the affected machine's queue
// file. Concurrency: a single sync.Mutex covers the map + each
// machine's slice (the contention is low — enqueue is operator-paced).
type QueueStore struct {
	mu     sync.Mutex
	dir    string
	queues map[string][]QueueItem
}

// NewQueueStore loads any persisted per-machine queues from dir.
// dir is created if missing. Pass an empty string to use the default
// resolveQueueDir() location. Errors loading a single machine's file
// are logged via the optional logf and skipped (a corrupt JSON file
// shouldn't brick the install).
func NewQueueStore(dir string, logf func(level, format string, args ...any)) (*QueueStore, error) {
	if dir == "" {
		dir = resolveQueueDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create queue dir: %w", err)
	}
	qs := &QueueStore{dir: dir, queues: map[string][]QueueItem{}}
	// Eagerly load — small file, single read.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		machineID := strings.TrimSuffix(e.Name(), ".json")
		buf, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			if logf != nil {
				logf("warn", "queue load %s: %v", e.Name(), err)
			}
			continue
		}
		var items []QueueItem
		if err := json.Unmarshal(buf, &items); err != nil {
			if logf != nil {
				logf("warn", "queue parse %s: %v", e.Name(), err)
			}
			continue
		}
		// Reset transient states — a process restart while sending /
		// running means the streamer no longer has that job, but the
		// queue file might. Demote those rows back to "queued" so the
		// operator can retry rather than seeing a phantom in-flight
		// entry.
		for i := range items {
			if items[i].State == QueueStateSending || items[i].State == QueueStateRunning {
				items[i].State = QueueStateQueued
				items[i].LineCurrent = 0
			}
		}
		qs.queues[machineID] = items
	}
	return qs, nil
}

// resolveQueueDir picks a writable directory for queue persistence.
// Default: $XDG_CONFIG_HOME/filebrowser-NC/queues. Falls back to
// $HOME/.config/filebrowser-NC/queues then os.TempDir() —- the
// install must always have a writable location, even when running
// in a stripped-down container.
func resolveQueueDir() string {
	if cfg, err := os.UserConfigDir(); err == nil && cfg != "" {
		return filepath.Join(cfg, "filebrowser-NC", "queues")
	}
	return filepath.Join(os.TempDir(), "filebrowser-NC-queues")
}

// List returns a snapshot copy of machineID's queue. Empty slice if
// the machine has no entries yet (never nil — callers can range
// without a guard).
func (qs *QueueStore) List(machineID string) []QueueItem {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	src := qs.queues[machineID]
	out := make([]QueueItem, len(src))
	copy(out, src)
	return out
}

// Add appends a new item to the back of the queue and persists.
// Returns the assigned ID. OnumberHint is parsed from the file at
// absPath; failures fall through to an empty hint (auto-match just
// won't fire for that item).
func (qs *QueueStore) Add(machineID string, item QueueItem, absPath string) (QueueItem, error) {
	if machineID == "" {
		return QueueItem{}, errors.New("machine_id required")
	}
	item.ID = newQueueID()
	if item.State == "" {
		item.State = QueueStateQueued
	}
	if item.AddedAt.IsZero() {
		item.AddedAt = time.Now().UTC()
	}
	if absPath != "" {
		if onum, _ := readONumber(absPath); onum != "" {
			item.OnumberHint = onum
		}
		if fi, err := os.Stat(absPath); err == nil {
			item.SizeBytes = fi.Size()
		}
	}
	if item.JobName == "" && item.FilePath != "" {
		// "/folder/file.nc" → "folder". Empty when the file lives at
		// the share root.
		dir := filepath.Dir(item.FilePath)
		if dir != "." && dir != "/" {
			item.JobName = filepath.Base(dir)
		}
	}
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.queues[machineID] = append(qs.queues[machineID], item)
	if err := qs.persistLocked(machineID); err != nil {
		return item, err
	}
	return item, nil
}

// Remove deletes the item with id from machineID's queue. Returns
// nil even when the id isn't found — idempotent so repeated DELETE
// calls don't trip an error.
func (qs *QueueStore) Remove(machineID, id string) error {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	src := qs.queues[machineID]
	out := src[:0]
	for _, it := range src {
		if it.ID == id {
			continue
		}
		out = append(out, it)
	}
	qs.queues[machineID] = out
	return qs.persistLocked(machineID)
}

// Reorder replaces machineID's queue with the supplied id ordering.
// Items whose IDs aren't in the supplied list are dropped (treated
// as "removed in the same reorder"). Items in `ids` but not in the
// current queue are ignored.
func (qs *QueueStore) Reorder(machineID string, ids []string) error {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	src := qs.queues[machineID]
	byID := make(map[string]QueueItem, len(src))
	for _, it := range src {
		byID[it.ID] = it
	}
	out := make([]QueueItem, 0, len(ids))
	for _, id := range ids {
		if it, ok := byID[id]; ok {
			out = append(out, it)
		}
	}
	qs.queues[machineID] = out
	return qs.persistLocked(machineID)
}

// MarkSending flips one row to "sending" and clears any prior
// sending/running rows back to "queued" (only one in-flight item per
// machine). Returns the updated item or an error if id is unknown.
func (qs *QueueStore) MarkSending(machineID, id, method string) (*QueueItem, error) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	src := qs.queues[machineID]
	var hit *QueueItem
	for i := range src {
		if src[i].ID == id {
			src[i].State = QueueStateSending
			src[i].Method = method
			hit = &src[i]
			continue
		}
		if src[i].State == QueueStateSending || src[i].State == QueueStateRunning {
			src[i].State = QueueStateQueued
			src[i].LineCurrent = 0
		}
	}
	if hit == nil {
		return nil, fmt.Errorf("queue item %s not found", id)
	}
	return hit, qs.persistLocked(machineID)
}

// FindByONumber returns a copy of the first queue item whose
// OnumberHint matches onum (after normalization), or nil when no
// queue row matches. Used by the registry auto-attach watcher to
// resolve a controller-reported O-number to a known file path
// without scanning the filesystem.
func (qs *QueueStore) FindByONumber(machineID, onum string) *QueueItem {
	if onum == "" {
		return nil
	}
	onum = normalizeONumber(onum)
	qs.mu.Lock()
	defer qs.mu.Unlock()
	for _, it := range qs.queues[machineID] {
		if it.OnumberHint == onum {
			out := it
			return &out
		}
	}
	return nil
}

// PromoteByONumber finds the item matching onum and promotes it to
// "running". Used by the streamer when a status frame reports a new
// program is executing — covers the SD-card / USB / pre-existing
// program case so the queue still tracks the work. No-op when no
// item matches.
func (qs *QueueStore) PromoteByONumber(machineID, onum string) {
	if onum == "" {
		return
	}
	onum = normalizeONumber(onum)
	qs.mu.Lock()
	defer qs.mu.Unlock()
	src := qs.queues[machineID]
	for i := range src {
		if src[i].OnumberHint == onum {
			src[i].State = QueueStateRunning
		} else if src[i].State == QueueStateRunning {
			// Another row was already in running; demote it. Only one
			// program runs at a time.
			src[i].State = QueueStateQueued
		}
	}
	_ = qs.persistLocked(machineID)
}

// MarkProgress updates the line counters on whatever row is currently
// sending or running. No-op when no row is in-flight.
func (qs *QueueStore) MarkProgress(machineID string, lineCurrent, lineTotal int) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	src := qs.queues[machineID]
	for i := range src {
		if src[i].State == QueueStateSending || src[i].State == QueueStateRunning {
			src[i].LineCurrent = lineCurrent
			if lineTotal > 0 {
				src[i].LineTotal = lineTotal
			}
		}
	}
	_ = qs.persistLocked(machineID)
}

// ClearInFlight drops sending/running state back to "queued". Used
// when the streamer reports a clean stop / idle transition.
func (qs *QueueStore) ClearInFlight(machineID string) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	src := qs.queues[machineID]
	changed := false
	for i := range src {
		if src[i].State == QueueStateSending || src[i].State == QueueStateRunning {
			src[i].State = QueueStateQueued
			src[i].LineCurrent = 0
			changed = true
		}
	}
	if changed {
		_ = qs.persistLocked(machineID)
	}
}

func (qs *QueueStore) persistLocked(machineID string) error {
	if qs.dir == "" {
		return nil
	}
	if machineID == "" {
		return errors.New("machine_id required")
	}
	if !safeMachineIDForFilename(machineID) {
		return fmt.Errorf("unsafe machine id for filename: %q", machineID)
	}
	path := filepath.Join(qs.dir, machineID+".json")
	buf, err := json.MarshalIndent(qs.queues[machineID], "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func newQueueID() string {
	// Time-prefixed + 6-char random suffix. Sortable; collision
	// resistant at queue cadence; short enough to read in logs.
	return fmt.Sprintf("q%d%s", time.Now().UnixNano(), randHex(6))
}

// safeMachineIDForFilename rejects ids that could traverse paths.
// Registry-issued IDs are URL-safe base64 (a-zA-Z0-9_-) so this is a
// belt-and-braces check against operator-supplied or migrated values.
func safeMachineIDForFilename(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}

// readONumber pulls the O-number ("O00057" or "O123") out of the first
// few non-comment lines of an NC file. Haas-typical layout puts O on
// line 1 or 2; we scan up to 20 lines so leading-percent / blank /
// comment lines don't trip us up. Returns "" on miss.
var onumberRe = regexp.MustCompile(`(?i)\bO(\d{1,5})\b`)

func readONumber(absPath string) (string, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for i := 0; i < 20 && sc.Scan(); i++ {
		line := sc.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "(") {
			continue
		}
		// Drop inline comments before matching so "(O42 reference)"
		// inside an active line doesn't fool the regex.
		stripped := strings.ReplaceAll(line, "(", " ")
		if i := strings.IndexByte(stripped, ')'); i >= 0 {
			stripped = stripped[:i]
		}
		if m := onumberRe.FindStringSubmatch(stripped); m != nil {
			return normalizeONumber("O" + m[1]), nil
		}
	}
	return "", sc.Err()
}

// normalizeONumber pads O-numbers to 5 digits so "O57" and "O00057"
// from the same program don't show as a queue mismatch.
func normalizeONumber(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	if !strings.HasPrefix(s, "O") {
		return s
	}
	digits := strings.TrimLeft(s[1:], "0")
	if digits == "" {
		digits = "0"
	}
	for len(digits) < 5 {
		digits = "0" + digits
	}
	return "O" + digits
}
