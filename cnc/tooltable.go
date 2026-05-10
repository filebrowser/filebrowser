package cnc

// Live tool-table readout — the next phase after the discovery probe
// (probe.go). Reads length-geometry / length-wear / diameter-geometry /
// diameter-wear for every requested slot and returns a structured
// table the dashboard can render and persist for history.
//
// Same constraints as the probe: serialized through queryMu with the
// 150 ms minQuerySpacing, refuses while a streaming job is running.
// 4 bases × N slots × 150 ms = ~0.6 s/slot, so 30 slots is ~18 s.

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ToolTableSlot is one slot's full readout. Numeric fields are
// optional pointers because a slot can read cleanly but be empty
// (the controller returns 0.0000), in which case we want the
// frontend to render "—" rather than "0.0000".
type ToolTableSlot struct {
	Slot         int      `json:"slot"`
	LengthGeom   *float64 `json:"length_geom,omitempty"`     // macro 2001+(slot-1)
	LengthWear   *float64 `json:"length_wear,omitempty"`     // macro 2201+(slot-1)
	DiameterGeom *float64 `json:"diameter_geom,omitempty"`   // macro 2401+(slot-1)
	DiameterWear *float64 `json:"diameter_wear,omitempty"`   // macro 2601+(slot-1)
	// EffectiveDiameter is geom + wear when both are present. Pre-
	// computed so the operator-side diameter check on Send doesn't
	// have to repeat the math (and so the JSON history dumps record
	// what the operator actually saw).
	EffectiveDiameter *float64 `json:"effective_diameter,omitempty"`
	EffectiveLength   *float64 `json:"effective_length,omitempty"`
	// Empty is true when every field that DID read came back 0.0
	// — distinguishes "pocket has no tool" from "we couldn't read".
	Empty bool `json:"empty,omitempty"`
	// Errors holds any per-field error strings, keyed by base name
	// ("length_geom", "diameter_wear", …). Populated only on failure
	// so the typical row stays small in the JSON dump.
	Errors map[string]string `json:"errors,omitempty"`
}

// ToolTable is the full structured readout from POST /api/cnc/tool-table.
type ToolTable struct {
	MachineID      string          `json:"machine_id"`
	MachineName    string          `json:"machine_name,omitempty"`
	BridgeAddress  string          `json:"bridge_address"`
	ReadAt         time.Time       `json:"read_at"`
	DurationMs     float64         `json:"duration_ms"`
	SlotsRequested int             `json:"slots_requested"`
	SlotsRead      int             `json:"slots_read"` // count of slots with at least one OK field
	Slots          []ToolTableSlot `json:"slots"`
}

// toolTableBases maps a JSON field key to the macro-var base. Same
// canonical NGC mapping the probe confirms — kept here so that if the
// probe ever has to switch to a legacy mapping, this file is the only
// place that needs to change.
var toolTableBases = []struct {
	Key  string
	Base int
}{
	{"length_geom", 2001},
	{"length_wear", 2201},
	{"diameter_geom", 2401},
	{"diameter_wear", 2601},
}

// ReadToolTable reads `slots` consecutive tool slots and returns a
// structured table. slots is clamped to [1, 200] (Haas tool table max).
// Refuses while a streaming job is running — Q-codes during a stream
// take the per-line write turn and would slow the stream.
func (s *Streamer) ReadToolTable(ctx context.Context, slots int) (*ToolTable, error) {
	if slots < 1 {
		slots = 30
	}
	if slots > 200 {
		slots = 200
	}
	m, port, err := s.resolveMachine()
	if err != nil {
		return nil, err
	}
	if s.IsRunning() {
		return nil, fmt.Errorf("can't read tool table during a streaming job")
	}

	t0 := time.Now()
	tbl := &ToolTable{
		MachineID:      s.machineID,
		MachineName:    m.Name,
		BridgeAddress:  fmt.Sprintf("%s:%d", m.Host, port),
		ReadAt:         t0.UTC(),
		SlotsRequested: slots,
		Slots:          make([]ToolTableSlot, 0, slots),
	}
	s.logf("info", "reading tool table — %d slots × %d bases", slots, len(toolTableBases))

	for slot := 1; slot <= slots; slot++ {
		row := ToolTableSlot{Slot: slot}
		// Track whether anything READ on this row, to distinguish
		// "empty pocket" (zeros across the board) from "we couldn't
		// reach this slot at all" (errors across the board).
		anyOK := false
		anyNonZero := false

		for _, b := range toolTableBases {
			v := b.Base + (slot - 1)
			res, qerr := s.Query(ctx, 600, &v)
			if qerr != nil {
				if row.Errors == nil {
					row.Errors = map[string]string{}
				}
				row.Errors[b.Key] = qerr.Error()
				continue
			}
			if !res.OK {
				if row.Errors == nil {
					row.Errors = map[string]string{}
				}
				row.Errors[b.Key] = res.Error
				continue
			}
			anyOK = true
			n, ok := parseFloatTail(res.Value)
			if !ok {
				if row.Errors == nil {
					row.Errors = map[string]string{}
				}
				row.Errors[b.Key] = "unparseable: " + res.Value
				continue
			}
			if n != 0 {
				anyNonZero = true
			}
			switch b.Key {
			case "length_geom":
				row.LengthGeom = &n
			case "length_wear":
				row.LengthWear = &n
			case "diameter_geom":
				row.DiameterGeom = &n
			case "diameter_wear":
				row.DiameterWear = &n
			}
		}

		if anyOK {
			tbl.SlotsRead++
		}
		if anyOK && !anyNonZero {
			row.Empty = true
		}
		// Pre-compute effective values when we have both halves.
		if row.LengthGeom != nil && row.LengthWear != nil {
			v := *row.LengthGeom + *row.LengthWear
			row.EffectiveLength = &v
		}
		if row.DiameterGeom != nil && row.DiameterWear != nil {
			v := *row.DiameterGeom + *row.DiameterWear
			row.EffectiveDiameter = &v
		}
		tbl.Slots = append(tbl.Slots, row)
	}

	tbl.DurationMs = sinceMs(t0)
	s.logf("info", "tool table read in %.1f ms — %d/%d slots populated",
		tbl.DurationMs, tbl.SlotsRead, slots)
	return tbl, nil
}

// parseFloatTail parses the trailing numeric token from a Q600 frame
// like "MACRO,2001,3.5400". Returns the float and true on success.
func parseFloatTail(value string) (float64, bool) {
	parts := splitAndTrim(value, ",")
	if len(parts) == 0 {
		return 0, false
	}
	last := strings.TrimSpace(parts[len(parts)-1])
	if last == "" {
		return 0, false
	}
	v, ok := parseNumber(last)
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	}
	return 0, false
}
