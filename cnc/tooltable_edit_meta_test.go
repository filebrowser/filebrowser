package cnc

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// JSON round-trip — ManuallyEdited + EditedAt + Source must survive
// the marshal that persists a dump and the unmarshal that loads it
// for diff / display. omitempty on the slot fields means an unedited
// dump stays as small as it always was.
func TestToolTableEditMetaRoundTrip(t *testing.T) {
	when := time.Date(2026, 5, 17, 14, 30, 0, 0, time.UTC)
	tbl := ToolTable{
		MachineID: "mill-1",
		ReadAt:    when,
		Source:    "edit",
		Slots: []ToolTableSlot{
			{Slot: 1, LengthGeom: ptrF(2.500)},
			{Slot: 2, LengthGeom: ptrF(3.250), ManuallyEdited: true, EditedAt: when},
		},
	}
	buf, err := json.Marshal(&tbl)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(buf)
	if !strings.Contains(body, `"source":"edit"`) {
		t.Fatalf("source field missing: %s", body)
	}
	if !strings.Contains(body, `"manually_edited":true`) {
		t.Fatalf("manually_edited missing on slot 2: %s", body)
	}
	// Slot 1 is unedited — must NOT carry the flag (omitempty).
	if strings.Contains(body, `"slot":1,"length_geom":2.5,"manually_edited"`) {
		t.Fatalf("unedited slot leaked manually_edited: %s", body)
	}
	var back ToolTable
	if err := json.Unmarshal(buf, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Source != "edit" {
		t.Fatalf("source lost: %q", back.Source)
	}
	if !back.Slots[1].ManuallyEdited {
		t.Fatal("manually_edited lost")
	}
	if !back.Slots[1].EditedAt.Equal(when) {
		t.Fatalf("edited_at lost: %v", back.Slots[1].EditedAt)
	}
	if back.Slots[0].ManuallyEdited {
		t.Fatal("unedited slot 1 deserialised as edited")
	}
}

// An older dump (no Source, no ManuallyEdited) must still parse. This
// is the contract for backwards-compatibility against the file-on-disk
// history dating from before the edit feature shipped.
func TestToolTableLegacyDumpParses(t *testing.T) {
	body := []byte(`{
		"machine_id": "mill-1",
		"read_at": "2026-04-01T10:00:00Z",
		"slots": [{"slot": 5, "length_geom": 4.125}]
	}`)
	var tbl ToolTable
	if err := json.Unmarshal(body, &tbl); err != nil {
		t.Fatalf("unmarshal legacy: %v", err)
	}
	if tbl.Source != "" {
		t.Fatalf("expected empty source, got %q", tbl.Source)
	}
	if tbl.Slots[0].ManuallyEdited {
		t.Fatal("legacy slot deserialised as edited")
	}
}
