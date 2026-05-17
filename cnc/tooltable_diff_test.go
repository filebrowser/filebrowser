package cnc

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func ptrF(v float64) *float64 { return &v }

func mkSlot(n int, dia, length float64) ToolTableSlot {
	return ToolTableSlot{
		Slot:              n,
		EffectiveDiameter: ptrF(dia),
		EffectiveLength:   ptrF(length),
	}
}

func mkEmpty(n int) ToolTableSlot {
	return ToolTableSlot{Slot: n, Empty: true}
}

func mkErr(n int, msg string) ToolTableSlot {
	return ToolTableSlot{
		Slot:   n,
		Errors: map[string]string{"length_geom": msg},
	}
}

func TestDiffEmptyVsEmpty(t *testing.T) {
	out := DiffToolTables(&ToolTable{}, &ToolTable{}, 0, 0)
	if len(out.Slots) != 0 {
		t.Fatalf("empty diff should be empty, got %+v", out.Slots)
	}
	if out.Summary != (DiffSummary{}) {
		t.Fatalf("empty diff summary should be zero, got %+v", out.Summary)
	}
}

// TestDiffSlotsAlwaysMarshalsAsArray protects against the regression
// where a no-change diff produced slots: null in the JSON, crashing
// the frontend's .length read with "Cannot read properties of null".
func TestDiffSlotsAlwaysMarshalsAsArray(t *testing.T) {
	out := DiffToolTables(&ToolTable{}, &ToolTable{}, 0, 0)
	buf, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(buf)
	if strings.Contains(got, `"slots":null`) {
		t.Fatalf("slots must marshal as [] not null: %s", got)
	}
	if !strings.Contains(got, `"slots":[]`) {
		t.Fatalf("expected slots:[] in output, got: %s", got)
	}
}

func TestDiffAddedTool(t *testing.T) {
	oldT := &ToolTable{Slots: []ToolTableSlot{mkEmpty(5)}}
	newT := &ToolTable{Slots: []ToolTableSlot{mkSlot(5, 0.500, 4.000)}}
	out := DiffToolTables(oldT, newT, 0, 0)
	if out.Summary.Added != 1 {
		t.Fatalf("expected 1 added, got %+v", out.Summary)
	}
	if len(out.Slots) != 1 || out.Slots[0].Change != SlotChangeAdded {
		t.Fatalf("expected added slot, got %+v", out.Slots)
	}
}

func TestDiffRemovedTool(t *testing.T) {
	oldT := &ToolTable{Slots: []ToolTableSlot{mkSlot(7, 0.250, 2.500)}}
	newT := &ToolTable{Slots: []ToolTableSlot{mkEmpty(7)}}
	out := DiffToolTables(oldT, newT, 0, 0)
	if out.Summary.Removed != 1 {
		t.Fatalf("expected 1 removed, got %+v", out.Summary)
	}
}

func TestDiffDiameterWear(t *testing.T) {
	// Wear: 0.5000 → 0.4940 (delta 0.006, exceeds 0.005 tolerance)
	oldT := &ToolTable{Slots: []ToolTableSlot{mkSlot(3, 0.5000, 3.000)}}
	newT := &ToolTable{Slots: []ToolTableSlot{mkSlot(3, 0.4940, 3.000)}}
	out := DiffToolTables(oldT, newT, 0, 0)
	if out.Summary.DriftDiameter != 1 || out.Summary.DriftBoth != 0 {
		t.Fatalf("expected diameter-only drift, got %+v", out.Summary)
	}
	d := out.Slots[0]
	if d.DiameterDelta == nil || *d.DiameterDelta > -0.0055 || *d.DiameterDelta < -0.0065 {
		t.Fatalf("delta should be ~-0.006, got %v", d.DiameterDelta)
	}
}

func TestDiffLengthOnlyWithinDiameterTolerance(t *testing.T) {
	oldT := &ToolTable{Slots: []ToolTableSlot{mkSlot(2, 0.500, 4.000)}}
	newT := &ToolTable{Slots: []ToolTableSlot{mkSlot(2, 0.500, 3.997)}}
	out := DiffToolTables(oldT, newT, 0, 0)
	if out.Summary.DriftLength != 1 {
		t.Fatalf("expected length-only drift, got %+v", out.Summary)
	}
}

func TestDiffDriftBoth(t *testing.T) {
	oldT := &ToolTable{Slots: []ToolTableSlot{mkSlot(1, 0.500, 4.000)}}
	newT := &ToolTable{Slots: []ToolTableSlot{mkSlot(1, 0.490, 3.995)}}
	out := DiffToolTables(oldT, newT, 0, 0)
	if out.Summary.DriftBoth != 1 {
		t.Fatalf("expected both drift, got %+v", out.Summary)
	}
}

func TestDiffUnchangedBelowTolerance(t *testing.T) {
	// 0.001 diameter delta — below 0.005 default. Should not surface.
	oldT := &ToolTable{Slots: []ToolTableSlot{mkSlot(8, 0.500, 4.000)}}
	newT := &ToolTable{Slots: []ToolTableSlot{mkSlot(8, 0.501, 4.001)}}
	out := DiffToolTables(oldT, newT, 0, 0)
	if out.Summary.Unchanged != 1 || len(out.Slots) != 0 {
		t.Fatalf("expected unchanged, got %+v / slots %v", out.Summary, out.Slots)
	}
}

func TestDiffOfflineInOldRead(t *testing.T) {
	oldT := &ToolTable{Slots: []ToolTableSlot{mkErr(4, "timeout")}}
	newT := &ToolTable{Slots: []ToolTableSlot{mkSlot(4, 0.500, 4.000)}}
	out := DiffToolTables(oldT, newT, 0, 0)
	if out.Summary.OfflineThen != 1 {
		t.Fatalf("expected offline_then, got %+v", out.Summary)
	}
}

func TestDiffNilOldTreatedAsEmpty(t *testing.T) {
	newT := &ToolTable{
		MachineID: "abc",
		ReadAt:    time.Now(),
		Slots:     []ToolTableSlot{mkSlot(1, 0.5, 4)},
	}
	out := DiffToolTables(nil, newT, 0, 0)
	if out.Summary.Added != 1 {
		t.Fatalf("expected nil-old to register additions, got %+v", out.Summary)
	}
	if out.MachineID != "abc" {
		t.Fatalf("machine id should populate from new when old missing")
	}
}

func TestDiffMixedSlotSets(t *testing.T) {
	oldT := &ToolTable{Slots: []ToolTableSlot{mkSlot(1, 0.500, 4.000), mkSlot(2, 0.250, 3.000)}}
	newT := &ToolTable{Slots: []ToolTableSlot{mkSlot(2, 0.250, 3.000), mkSlot(3, 0.125, 2.000)}}
	out := DiffToolTables(oldT, newT, 0, 0)
	// Slot 1 removed, slot 3 added, slot 2 unchanged.
	if out.Summary.Added != 1 || out.Summary.Removed != 1 || out.Summary.Unchanged != 1 {
		t.Fatalf("union of slot sets miscounted: %+v", out.Summary)
	}
}
