package cnc

// Fusion 360 tool library import — operator exports their library as
// JSON (Tools tab → Export Tool Library) and uploads it here. We keep
// the entries verbatim so the SVG profile renderer (future) can draw
// the revolved silhouette directly from the holder.segments + tool
// geometry. The dashboard surfaces:
//   - The richer description per slot (vendor + part + dimensions)
//   - Vendor product-link when available
//   - Tool type, flute count, OAL — fields the live Q-code read
//     doesn't expose
//
// Storage: single JSON file at $XDG_CONFIG_HOME/filebrowser-NC/
// tool-library.json. Best-effort; if the file's missing the lookup
// just returns nil.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FusionToolGeometry mirrors the geometry block in a Fusion tool
// library export. Fields are optional — some tool types (face mill,
// face indexable) skip the corner radius, taper angle, etc.
type FusionToolGeometry struct {
	CSP                 bool    `json:"CSP,omitempty"`
	DC                  float64 `json:"DC,omitempty"`
	DCX                 float64 `json:"DCX,omitempty"`
	HAND                bool    `json:"HAND,omitempty"`
	LB                  float64 `json:"LB,omitempty"`
	LCF                 float64 `json:"LCF,omitempty"`
	NOF                 int     `json:"NOF,omitempty"`
	OAL                 float64 `json:"OAL,omitempty"`
	RE                  float64 `json:"RE,omitempty"`
	SFDM                float64 `json:"SFDM,omitempty"`
	TA                  float64 `json:"TA,omitempty"`
	AssemblyGaugeLength float64 `json:"assemblyGaugeLength,omitempty"`
	ShoulderDiameter    float64 `json:"shoulder-diameter,omitempty"`
	ShoulderLength      float64 `json:"shoulder-length,omitempty"`
	TipDiameter         float64 `json:"tip-diameter,omitempty"`
}

// FusionHolderSegment is one step in the holder's revolved profile.
// The full profile is a sequence of these, walked tip-up. Diameters
// in the tool's `unit` field (inches or millimeters).
type FusionHolderSegment struct {
	Height         float64 `json:"height"`
	LowerDiameter  float64 `json:"lower-diameter"`
	UpperDiameter  float64 `json:"upper-diameter"`
}

// FusionHolder is the assembly piece between the spindle taper and
// the cutting tool. Segments form the revolved profile rendered in
// the magazine view.
type FusionHolder struct {
	Description string                `json:"description"`
	GaugeLength float64               `json:"gaugeLength"`
	GUID        string                `json:"guid,omitempty"`
	ProductID   string                `json:"product-id,omitempty"`
	ProductLink string                `json:"product-link,omitempty"`
	Segments    []FusionHolderSegment `json:"segments,omitempty"`
	Type        string                `json:"type,omitempty"`
	Unit        string                `json:"unit,omitempty"`
	Vendor      string                `json:"vendor,omitempty"`
}

// FusionPostProcess is the controller-side mapping — `number` is the
// tool pocket number on the machine. Zero means the tool exists in
// the library but isn't currently loaded.
type FusionPostProcess struct {
	BreakControl     bool   `json:"break-control,omitempty"`
	Comment          string `json:"comment,omitempty"`
	DiameterOffset   int    `json:"diameter-offset,omitempty"`
	LengthOffset     int    `json:"length-offset,omitempty"`
	Live             bool   `json:"live,omitempty"`
	ManualToolChange bool   `json:"manual-tool-change,omitempty"`
	Number           int    `json:"number,omitempty"`
	Turret           int    `json:"turret,omitempty"`
}

// FusionTool is one entry in the library. Top-level fields preserve
// what Fusion exports; we don't lose anything on round-trip.
type FusionTool struct {
	BMC         string             `json:"BMC,omitempty"`
	Description string             `json:"description,omitempty"`
	Geometry    FusionToolGeometry `json:"geometry,omitempty"`
	GUID        string             `json:"guid,omitempty"`
	Holder      FusionHolder       `json:"holder,omitempty"`
	PostProcess FusionPostProcess  `json:"post-process,omitempty"`
	ProductID   string             `json:"product-id,omitempty"`
	ProductLink string             `json:"product-link,omitempty"`
	// StartValues is preserved as raw JSON — the speeds/feeds presets
	// are voluminous and we don't surface them yet. Keeping the bytes
	// means a future feature can decode without re-uploading.
	StartValues json.RawMessage `json:"start-values,omitempty"`
	Type        string          `json:"type,omitempty"`
	Unit        string          `json:"unit,omitempty"`
	Vendor      string          `json:"vendor,omitempty"`
}

// IsHolderOnly returns true for entries that are bare holders (no
// cutting tool). Fusion exports these as part of the holder catalog.
// Excluded from per-slot lookups.
func (t FusionTool) IsHolderOnly() bool {
	return t.Type == "holder" || (t.Geometry == FusionToolGeometry{} && t.Type == "")
}

// FusionLibrary is the file shape Fusion exports.
type FusionLibrary struct {
	Data    []FusionTool    `json:"data"`
	Version int             `json:"version,omitempty"`
	// UploadedAt is set on import for the GET response — not present
	// in the source file.
	UploadedAt time.Time `json:"uploaded_at,omitempty"`
}

// ToolLibrary indexes a parsed FusionLibrary by tool pocket number
// for fast per-slot lookups. Holder-only entries are dropped.
type ToolLibrary struct {
	raw   FusionLibrary
	byNum map[int]FusionTool
}

// NewToolLibrary builds the index. Tools whose post-process.number is
// 0 are kept in `raw` but not in `byNum` (not currently loaded).
func NewToolLibrary(raw FusionLibrary) *ToolLibrary {
	idx := map[int]FusionTool{}
	for _, t := range raw.Data {
		if t.IsHolderOnly() {
			continue
		}
		n := t.PostProcess.Number
		if n <= 0 {
			continue
		}
		// First-write-wins on duplicate numbers. Fusion allows
		// duplicates (multiple tools assigned to the same pocket from
		// different programs); the first entry is usually the active
		// one.
		if _, exists := idx[n]; !exists {
			idx[n] = t
		}
	}
	return &ToolLibrary{raw: raw, byNum: idx}
}

// Lookup returns the library entry for the given pocket number, or
// (zero, false) when no tool is assigned to that slot.
func (l *ToolLibrary) Lookup(number int) (FusionTool, bool) {
	if l == nil {
		return FusionTool{}, false
	}
	t, ok := l.byNum[number]
	return t, ok
}

// AssignedSlots returns the sorted list of pocket numbers that have
// a tool entry. Useful for the upload-confirmation UI.
func (l *ToolLibrary) AssignedSlots() []int {
	if l == nil {
		return nil
	}
	out := make([]int, 0, len(l.byNum))
	for n := range l.byNum {
		out = append(out, n)
	}
	// Simple insertion sort — assigned-slot counts are typically <200.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

// Raw returns the underlying library (with UploadedAt set). Callers
// shouldn't mutate the returned slice — pass through to JSON.
func (l *ToolLibrary) Raw() FusionLibrary {
	if l == nil {
		return FusionLibrary{}
	}
	return l.raw
}

// ── Persistence ────────────────────────────────────────────────────────

// LibraryStore loads/saves a single tool library to a config file.
// Concurrent reads are safe; writes serialize via mu.
type LibraryStore struct {
	path string

	mu  sync.RWMutex
	lib *ToolLibrary
}

// NewLibraryStore returns a store rooted at path, eager-loading the
// existing file if present. Pass "" to use the default config dir.
// A missing file is not an error — the store is just empty.
func NewLibraryStore(path string) (*LibraryStore, error) {
	if path == "" {
		path = resolveToolLibraryPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create lib dir: %w", err)
	}
	s := &LibraryStore{path: path}
	if err := s.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		// Bad JSON shouldn't brick the install; log via the caller's
		// error handler (the cnc.Registry wires logf for this).
		return s, err
	}
	return s, nil
}

func resolveToolLibraryPath() string {
	if cfg, err := os.UserConfigDir(); err == nil && cfg != "" {
		return filepath.Join(cfg, "filebrowser-NC", "tool-library.json")
	}
	return filepath.Join(os.TempDir(), "filebrowser-NC-tool-library.json")
}

func (s *LibraryStore) load() error {
	buf, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var fl FusionLibrary
	if err := json.Unmarshal(buf, &fl); err != nil {
		return fmt.Errorf("parse tool library: %w", err)
	}
	s.mu.Lock()
	s.lib = NewToolLibrary(fl)
	s.mu.Unlock()
	return nil
}

// Replace persists the new library and atomically swaps the in-memory
// pointer. The input is parsed eagerly so a malformed JSON fails the
// upload rather than corrupting on-disk state.
func (s *LibraryStore) Replace(raw []byte) (*ToolLibrary, error) {
	var fl FusionLibrary
	if err := json.Unmarshal(raw, &fl); err != nil {
		return nil, fmt.Errorf("parse tool library: %w", err)
	}
	if len(fl.Data) == 0 {
		return nil, fmt.Errorf("tool library has no entries (expected Fusion-style data: [...])")
	}
	fl.UploadedAt = time.Now().UTC()
	// Re-marshal so the on-disk file is canonical (pretty-printed,
	// uploaded_at populated) rather than whatever the upload byte-for-
	// byte was.
	out, err := json.MarshalIndent(fl, "", "  ")
	if err != nil {
		return nil, err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return nil, fmt.Errorf("write tool library: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return nil, fmt.Errorf("rename tool library: %w", err)
	}
	lib := NewToolLibrary(fl)
	s.mu.Lock()
	s.lib = lib
	s.mu.Unlock()
	return lib, nil
}

// Library returns the current in-memory index (or nil when none).
func (s *LibraryStore) Library() *ToolLibrary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lib
}

// Clear empties the library and removes the file. Used by the "Reset"
// button in Settings → Tool Library.
func (s *LibraryStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lib = nil
	err := os.Remove(s.path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
