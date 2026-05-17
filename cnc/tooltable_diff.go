package cnc

// Tool-table history diff — joins two ToolTable snapshots slot-by-slot
// and classifies each row's change since the older read. Operators
// tracking tool wear use this to spot "tool 5 wore 0.001" / "tool 14
// was changed" without eyeballing two JSON dumps side-by-side.
//
// Pure data — no Haas access. The HTTP layer feeds in two persisted
// reads from the /cnc-tool-tables share and renders the result.

import (
	"math"
	"sort"
	"time"
)

// SlotChange tags how slot N transitioned between two reads. Operators
// are watching for tool swaps (added / removed) and wear (drift).
type SlotChange string

const (
	// SlotChangeUnchanged — both reads agree within the diameter
	// tolerance and length tolerance. Excluded from the wire payload
	// unless the caller asks for full output.
	SlotChangeUnchanged SlotChange = "unchanged"
	// SlotChangeAdded — old read was empty/missing, new read has a tool.
	SlotChangeAdded SlotChange = "added"
	// SlotChangeRemoved — old read had a tool, new read is empty.
	SlotChangeRemoved SlotChange = "removed"
	// SlotChangeDriftDiameter — same tool, diameter delta exceeds
	// tolerance. Almost always wear (rarely a re-grind).
	SlotChangeDriftDiameter SlotChange = "drift_diameter"
	// SlotChangeDriftLength — same tool, length delta exceeds tolerance.
	// Often a re-touch-off (operator re-zeroed Z).
	SlotChangeDriftLength SlotChange = "drift_length"
	// SlotChangeDriftBoth — both diameter and length drifted.
	SlotChangeDriftBoth SlotChange = "drift_both"
	// SlotChangeOfflineThen / SlotChangeOfflineNow — one of the two
	// reads had errors on this slot, so we can't compare cleanly.
	SlotChangeOfflineThen SlotChange = "offline_then"
	SlotChangeOfflineNow  SlotChange = "offline_now"
)

// SlotDiff is one row in the diff result.
type SlotDiff struct {
	Slot          int        `json:"slot"`
	Change        SlotChange `json:"change"`
	OldDiameter   *float64   `json:"old_diameter,omitempty"`
	NewDiameter   *float64   `json:"new_diameter,omitempty"`
	DiameterDelta *float64   `json:"diameter_delta,omitempty"`
	OldLength     *float64   `json:"old_length,omitempty"`
	NewLength     *float64   `json:"new_length,omitempty"`
	LengthDelta   *float64   `json:"length_delta,omitempty"`
	Note          string     `json:"note,omitempty"`
}

// DiffSummary counts each non-unchanged transition. Renders as the
// header line on the diff page ("3 wore, 1 added").
type DiffSummary struct {
	Added         int `json:"added"`
	Removed       int `json:"removed"`
	DriftDiameter int `json:"drift_diameter"`
	DriftLength   int `json:"drift_length"`
	DriftBoth     int `json:"drift_both"`
	OfflineThen   int `json:"offline_then"`
	OfflineNow    int `json:"offline_now"`
	Unchanged     int `json:"unchanged"`
}

// ToolTableDiff is the response shape from GET /api/cnc/tool-table/diff.
type ToolTableDiff struct {
	MachineID         string      `json:"machine_id"`
	OldReadAt         time.Time   `json:"old_read_at"`
	NewReadAt         time.Time   `json:"new_read_at"`
	DiameterTolerance float64     `json:"diameter_tolerance"`
	LengthTolerance   float64     `json:"length_tolerance"`
	Summary           DiffSummary `json:"summary"`
	// Slots contains every slot that's NOT unchanged, sorted ascending.
	// Operators want a tight summary — adding 30 "unchanged" rows is
	// noise. The Summary.Unchanged count covers the rest.
	Slots []SlotDiff `json:"slots"`
}

// Default tolerances. Diameter mirrors preflight's DiameterTolerance
// (0.005"); length is slightly tighter at 0.002" since Z-zero drift
// shows up earlier than diameter wear.
const (
	DiffDefaultDiameterTolerance = 0.005
	DiffDefaultLengthTolerance   = 0.002
)

// DiffToolTables joins old and new slot-by-slot and returns a SlotDiff
// per slot that changed. Tolerances <= 0 fall back to the defaults.
// Either snapshot may be nil-safe — a nil snapshot is treated as an
// empty table, so a brand-new install with one read returns every
// populated slot as "added".
func DiffToolTables(oldT, newT *ToolTable, diaTol, lenTol float64) *ToolTableDiff {
	if diaTol <= 0 {
		diaTol = DiffDefaultDiameterTolerance
	}
	if lenTol <= 0 {
		lenTol = DiffDefaultLengthTolerance
	}
	out := &ToolTableDiff{
		DiameterTolerance: diaTol,
		LengthTolerance:   lenTol,
		// Init as empty slice (not nil) so the JSON response always
		// has slots: [] when there are no changes — frontend reads
		// .length on it and a null crashes the diff panel.
		Slots: []SlotDiff{},
	}
	oldSlots := indexSlots(oldT)
	newSlots := indexSlots(newT)
	if oldT != nil {
		out.OldReadAt = oldT.ReadAt
		out.MachineID = oldT.MachineID
	}
	if newT != nil {
		out.NewReadAt = newT.ReadAt
		if out.MachineID == "" {
			out.MachineID = newT.MachineID
		}
	}

	// Union of slot numbers from both snapshots — covers slots that
	// exist in one but not the other.
	slotNums := map[int]bool{}
	for n := range oldSlots {
		slotNums[n] = true
	}
	for n := range newSlots {
		slotNums[n] = true
	}
	keys := make([]int, 0, len(slotNums))
	for n := range slotNums {
		keys = append(keys, n)
	}
	sort.Ints(keys)

	for _, n := range keys {
		d := classifySlot(n, oldSlots[n], newSlots[n], diaTol, lenTol)
		switch d.Change {
		case SlotChangeUnchanged:
			out.Summary.Unchanged++
		case SlotChangeAdded:
			out.Summary.Added++
			out.Slots = append(out.Slots, d)
		case SlotChangeRemoved:
			out.Summary.Removed++
			out.Slots = append(out.Slots, d)
		case SlotChangeDriftDiameter:
			out.Summary.DriftDiameter++
			out.Slots = append(out.Slots, d)
		case SlotChangeDriftLength:
			out.Summary.DriftLength++
			out.Slots = append(out.Slots, d)
		case SlotChangeDriftBoth:
			out.Summary.DriftBoth++
			out.Slots = append(out.Slots, d)
		case SlotChangeOfflineThen:
			out.Summary.OfflineThen++
			out.Slots = append(out.Slots, d)
		case SlotChangeOfflineNow:
			out.Summary.OfflineNow++
			out.Slots = append(out.Slots, d)
		}
	}
	return out
}

func indexSlots(t *ToolTable) map[int]*ToolTableSlot {
	if t == nil {
		return map[int]*ToolTableSlot{}
	}
	out := make(map[int]*ToolTableSlot, len(t.Slots))
	for i := range t.Slots {
		out[t.Slots[i].Slot] = &t.Slots[i]
	}
	return out
}

func classifySlot(n int, oldS, newS *ToolTableSlot, diaTol, lenTol float64) SlotDiff {
	d := SlotDiff{Slot: n}
	switch {
	case oldS == nil && newS == nil:
		d.Change = SlotChangeUnchanged
		return d
	case oldS == nil && newS != nil:
		fillNew(&d, newS)
		if isLoaded(newS) {
			d.Change = SlotChangeAdded
			return d
		}
		// New side reports empty/offline — no actionable change.
		d.Change = SlotChangeUnchanged
		return d
	case oldS != nil && newS == nil:
		fillOld(&d, oldS)
		if isLoaded(oldS) {
			d.Change = SlotChangeRemoved
			return d
		}
		d.Change = SlotChangeUnchanged
		return d
	}

	// Both present. Check error states first — a slot we couldn't
	// read isn't comparable.
	if len(oldS.Errors) > 0 {
		d.Change = SlotChangeOfflineThen
		fillOld(&d, oldS)
		fillNew(&d, newS)
		return d
	}
	if len(newS.Errors) > 0 {
		d.Change = SlotChangeOfflineNow
		fillOld(&d, oldS)
		fillNew(&d, newS)
		return d
	}

	oldLoaded := isLoaded(oldS)
	newLoaded := isLoaded(newS)
	fillOld(&d, oldS)
	fillNew(&d, newS)
	switch {
	case !oldLoaded && newLoaded:
		d.Change = SlotChangeAdded
		return d
	case oldLoaded && !newLoaded:
		d.Change = SlotChangeRemoved
		return d
	case !oldLoaded && !newLoaded:
		d.Change = SlotChangeUnchanged
		return d
	}

	// Both loaded. Compute deltas and classify drift.
	var diaDrift, lenDrift bool
	if oldS.EffectiveDiameter != nil && newS.EffectiveDiameter != nil {
		dd := *newS.EffectiveDiameter - *oldS.EffectiveDiameter
		d.DiameterDelta = &dd
		if math.Abs(dd) > diaTol {
			diaDrift = true
		}
	}
	if oldS.EffectiveLength != nil && newS.EffectiveLength != nil {
		dl := *newS.EffectiveLength - *oldS.EffectiveLength
		d.LengthDelta = &dl
		if math.Abs(dl) > lenTol {
			lenDrift = true
		}
	}
	switch {
	case diaDrift && lenDrift:
		d.Change = SlotChangeDriftBoth
	case diaDrift:
		d.Change = SlotChangeDriftDiameter
	case lenDrift:
		d.Change = SlotChangeDriftLength
	default:
		d.Change = SlotChangeUnchanged
	}
	return d
}

func isLoaded(s *ToolTableSlot) bool {
	if s == nil || s.Empty {
		return false
	}
	if s.EffectiveLength != nil && *s.EffectiveLength != 0 {
		return true
	}
	if s.EffectiveDiameter != nil && *s.EffectiveDiameter != 0 {
		return true
	}
	return false
}

func fillOld(d *SlotDiff, s *ToolTableSlot) {
	d.OldDiameter = s.EffectiveDiameter
	d.OldLength = s.EffectiveLength
}

func fillNew(d *SlotDiff, s *ToolTableSlot) {
	d.NewDiameter = s.EffectiveDiameter
	d.NewLength = s.EffectiveLength
}
