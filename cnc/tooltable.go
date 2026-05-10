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

// Tool-table macro-var bases. Canonical NGC mapping confirmed by the
// probe (probe.go). baseForKey() is the only consumer.
const (
	toolTableBaseLengthGeom   = 2001
	toolTableBaseLengthWear   = 2201
	toolTableBaseDiameterGeom = 2401
	toolTableBaseDiameterWear = 2601
)

// ReadToolTable reads `slots` consecutive tool slots and returns a
// structured table. slots is clamped to [1, 200] (Haas tool table max).
// Refuses while a streaming job is running — Q-codes during a stream
// take the per-line write turn and would slow the stream.
//
// Strategy: two-pass to avoid blowing 4× round-trips on empty pockets.
//
//  1. First pass reads length-geom for every slot. ~150 ms/round-trip
//     × N slots. Length-geom == 0 is the canonical "no tool" marker
//     on Haas — controllers don't carry orphan wear/diameter values
//     for an unset pocket, so skipping the other 3 bases is safe.
//  2. Second pass fetches length-wear / diameter-geom / diameter-wear
//     ONLY for slots whose length-geom came back non-zero.
//
// For a 200-slot table with 14 populated tools that drops the read
// from 800 round-trips to 200 + 42 = 242 — roughly 3× faster at any
// baud, and the difference between "I'll wait" and "no thanks" at
// 9600.
//
// On context cancellation (timeout or operator dismiss) we return the
// partial table built so far rather than nil — the operator gets to
// see whatever did read, and the persisted JSON marks every unreached
// slot's errors map with the cancel reason.
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

	// Per-slot row built up across the two passes. Index 0 unused so
	// slot N maps to rows[N] for readability.
	rows := make([]ToolTableSlot, slots+1)
	for i := 1; i <= slots; i++ {
		rows[i] = ToolTableSlot{Slot: i}
	}

	// ── Pass 1: length-geom across every slot ────────────────────────
	s.logf("info", "tool-table pass 1 — length-geom across %d slots", slots)
	populated := make([]bool, slots+1)
	for slot := 1; slot <= slots; slot++ {
		if ctxErr := ctx.Err(); ctxErr != nil {
			markRemainingCancelled(rows, slot, slots, ctxErr)
			break
		}
		v := toolTableBaseLengthGeom + (slot - 1)
		res, qerr := s.Query(ctx, 600, &v)
		applyBase(&rows[slot], "length_geom", res, qerr, &populated[slot])
	}

	// ── Pass 2: the other three bases, populated slots only ──────────
	pcount := 0
	for _, ok := range populated[1:] {
		if ok {
			pcount++
		}
	}
	s.logf("info", "tool-table pass 2 — 3 bases × %d populated slots", pcount)
	for slot := 1; slot <= slots; slot++ {
		if !populated[slot] {
			continue
		}
		for _, key := range [...]string{"length_wear", "diameter_geom", "diameter_wear"} {
			if ctxErr := ctx.Err(); ctxErr != nil {
				markRemainingCancelled(rows, slot, slots, ctxErr)
				goto FINISH
			}
			base := baseForKey(key)
			v := base + (slot - 1)
			res, qerr := s.Query(ctx, 600, &v)
			applyBase(&rows[slot], key, res, qerr, nil)
		}
	}

FINISH:
	// Stitch row metadata + effective values, copy to the table.
	for slot := 1; slot <= slots; slot++ {
		row := rows[slot]
		anyOK := row.LengthGeom != nil || row.LengthWear != nil ||
			row.DiameterGeom != nil || row.DiameterWear != nil
		anyNonZero := nonZero(row.LengthGeom) || nonZero(row.LengthWear) ||
			nonZero(row.DiameterGeom) || nonZero(row.DiameterWear)
		if anyOK {
			tbl.SlotsRead++
		}
		if anyOK && !anyNonZero {
			row.Empty = true
		}
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
	if ctxErr := ctx.Err(); ctxErr != nil {
		// Partial result is still useful — surface the cancel reason
		// in the table envelope and return non-nil so the caller can
		// persist whatever did read.
		return tbl, fmt.Errorf("tool-table read cancelled: %w", ctxErr)
	}
	return tbl, nil
}

func applyBase(row *ToolTableSlot, key string, res *QueryResult, qerr error, lengthPopulated *bool) {
	if qerr != nil {
		row.errSet(key, qerr.Error())
		return
	}
	if !res.OK {
		row.errSet(key, res.Error)
		return
	}
	n, ok := parseFloatTail(res.Value)
	if !ok {
		row.errSet(key, "unparseable: "+res.Value)
		return
	}
	switch key {
	case "length_geom":
		row.LengthGeom = &n
		if lengthPopulated != nil && n != 0 {
			*lengthPopulated = true
		}
	case "length_wear":
		row.LengthWear = &n
	case "diameter_geom":
		row.DiameterGeom = &n
	case "diameter_wear":
		row.DiameterWear = &n
	}
}

func (s *ToolTableSlot) errSet(key, msg string) {
	if s.Errors == nil {
		s.Errors = map[string]string{}
	}
	s.Errors[key] = msg
}

func baseForKey(key string) int {
	switch key {
	case "length_geom":
		return toolTableBaseLengthGeom
	case "length_wear":
		return toolTableBaseLengthWear
	case "diameter_geom":
		return toolTableBaseDiameterGeom
	case "diameter_wear":
		return toolTableBaseDiameterWear
	}
	return 0
}

func nonZero(p *float64) bool { return p != nil && *p != 0 }

// markRemainingCancelled stamps a "cancelled" error onto every base
// of every slot from `from` through `slots` that we didn't actually
// reach. The frontend renders these as red, distinct from "empty
// pocket". Idempotent — never overwrites a real read.
func markRemainingCancelled(rows []ToolTableSlot, from, slots int, ctxErr error) {
	msg := "cancelled: " + ctxErr.Error()
	for slot := from; slot <= slots; slot++ {
		row := &rows[slot]
		for _, key := range [...]string{"length_geom", "length_wear", "diameter_geom", "diameter_wear"} {
			if row.Errors != nil {
				if _, already := row.Errors[key]; already {
					continue
				}
			}
			// Only stamp keys that didn't read.
			switch key {
			case "length_geom":
				if row.LengthGeom != nil {
					continue
				}
			case "length_wear":
				if row.LengthWear != nil {
					continue
				}
			case "diameter_geom":
				if row.DiameterGeom != nil {
					continue
				}
			case "diameter_wear":
				if row.DiameterWear != nil {
					continue
				}
			}
			row.errSet(key, msg)
		}
	}
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
