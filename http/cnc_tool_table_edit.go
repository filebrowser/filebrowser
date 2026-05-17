package fbhttp

// Local-only tool-table edits.
//
// Operators sometimes need to tweak an offset without round-tripping
// through the controller: copying a sister-tool's offsets into a fresh
// pocket, recording a hand-measurement when the spindle probe is out,
// or restoring a known-good value after a chip strike. None of those
// touch the machine — this endpoint only writes the edit into a new
// history dump so the dashboard reflects the operator's intent.
//
// Write-back to the controller (G10 emission) is a deliberate future
// phase: see project_filebrowser_nc_tooltable_edit_todo.md. Until that
// lands, an edit is purely a dashboard override that the next
// controller read will replace.

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/filebrowser/filebrowser/v2/cnc"
)

// toolTableEditBody is the wire shape for POST /api/cnc/tool-table/edit.
//
// CopyFromSlot, when non-zero, takes precedence over the explicit
// numeric fields — the four offsets of the named slot in the latest
// dump get copied into Slot. The frontend's "copy from slot N" picker
// uses this path; the inline-edit form uses the per-field path.
type toolTableEditBody struct {
	Slot         int      `json:"slot"`
	LengthGeom   *float64 `json:"length_geom,omitempty"`
	LengthWear   *float64 `json:"length_wear,omitempty"`
	DiameterGeom *float64 `json:"diameter_geom,omitempty"`
	DiameterWear *float64 `json:"diameter_wear,omitempty"`
	CopyFromSlot int      `json:"copy_from_slot,omitempty"`
}

func cncToolTableEditHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}
		st, machineID, code, err := resolveStreamer(registry, r)
		if err != nil {
			return code, err
		}
		// Refuse mid-stream — same gate as a real read. An edit doesn't
		// touch the controller, but persisting one in the middle of a
		// job is confusing: the operator scrolls back later and sees a
		// dump dated mid-cycle that the controller never produced.
		if st.IsRunning() {
			return http.StatusConflict, errors.New("can't edit tool table during a streaming job")
		}

		var req toolTableEditBody
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return http.StatusBadRequest, err
		}
		if req.Slot < 1 || req.Slot > 200 {
			return http.StatusBadRequest, errors.New("slot must be 1-200")
		}
		if req.CopyFromSlot != 0 && (req.CopyFromSlot < 1 || req.CopyFromSlot > 200) {
			return http.StatusBadRequest, errors.New("copy_from_slot must be 1-200")
		}
		if req.CopyFromSlot == req.Slot {
			return http.StatusBadRequest, errors.New("copy_from_slot must differ from slot")
		}
		if req.CopyFromSlot == 0 &&
			req.LengthGeom == nil && req.LengthWear == nil &&
			req.DiameterGeom == nil && req.DiameterWear == nil {
			return http.StatusBadRequest, errors.New("nothing to apply: pass copy_from_slot or at least one numeric field")
		}

		dir := toolTableDirAbs(d, machineID)
		latestPath, err := newestJSONIn(dir)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		if latestPath == "" {
			return http.StatusBadRequest, errors.New("no tool-table read on file yet — read at least once before editing")
		}
		buf, err := os.ReadFile(latestPath)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		var prev cnc.ToolTable
		if err := json.Unmarshal(buf, &prev); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("parse latest dump: %w", err)
		}

		// Resolve copy-from source against the prior dump. Returning a
		// 404 here (rather than silently no-oping) helps the operator
		// catch a stale dropdown where they're trying to copy from a
		// slot that's since been cleared.
		var copySrc *cnc.ToolTableSlot
		if req.CopyFromSlot != 0 {
			copySrc = findSlot(&prev, req.CopyFromSlot)
			if copySrc == nil {
				return http.StatusNotFound, fmt.Errorf("copy_from_slot %d not present in latest read", req.CopyFromSlot)
			}
		}

		// Build the new table as a copy of the previous, then mutate
		// the target slot. Other slots carry their original measured
		// values + their original ManuallyEdited / EditedAt markers,
		// so a row edited two reads ago still surfaces as edited until
		// the controller re-reads it.
		next := prev
		next.Slots = make([]cnc.ToolTableSlot, len(prev.Slots))
		copy(next.Slots, prev.Slots)

		target := -1
		for i := range next.Slots {
			if next.Slots[i].Slot == req.Slot {
				target = i
				break
			}
		}
		if target < 0 {
			// Edit landed on a slot the previous read didn't cover.
			// Append a fresh row so the dump still includes it — the
			// frontend renders rows in slot order regardless of source.
			next.Slots = append(next.Slots, cnc.ToolTableSlot{Slot: req.Slot})
			target = len(next.Slots) - 1
		}

		now := time.Now().UTC()
		row := &next.Slots[target]
		// Clear errors on the edited row — manual override implies
		// "I'm telling you what this is" and the operator wouldn't see
		// stale "offline" badges hanging around alongside their values.
		row.Errors = nil
		row.Empty = false

		if copySrc != nil {
			row.LengthGeom = clonePtr(copySrc.LengthGeom)
			row.LengthWear = clonePtr(copySrc.LengthWear)
			row.DiameterGeom = clonePtr(copySrc.DiameterGeom)
			row.DiameterWear = clonePtr(copySrc.DiameterWear)
		} else {
			if req.LengthGeom != nil {
				row.LengthGeom = clonePtr(req.LengthGeom)
			}
			if req.LengthWear != nil {
				row.LengthWear = clonePtr(req.LengthWear)
			}
			if req.DiameterGeom != nil {
				row.DiameterGeom = clonePtr(req.DiameterGeom)
			}
			if req.DiameterWear != nil {
				row.DiameterWear = clonePtr(req.DiameterWear)
			}
		}
		// Recompute effectives — geom + wear when both present.
		if row.LengthGeom != nil && row.LengthWear != nil {
			v := *row.LengthGeom + *row.LengthWear
			row.EffectiveLength = &v
		} else {
			row.EffectiveLength = nil
		}
		if row.DiameterGeom != nil && row.DiameterWear != nil {
			v := *row.DiameterGeom + *row.DiameterWear
			row.EffectiveDiameter = &v
		} else {
			row.EffectiveDiameter = nil
		}
		row.ManuallyEdited = true
		row.EditedAt = now

		// Recompute slots_read against the edited table.
		next.SlotsRead = 0
		for _, s := range next.Slots {
			if s.LengthGeom != nil || s.LengthWear != nil ||
				s.DiameterGeom != nil || s.DiameterWear != nil {
				next.SlotsRead++
			}
		}
		next.ReadAt = now
		next.DurationMs = 0
		next.Source = "edit"

		if err := persistToolTable(d, machineID, &next); err != nil {
			return http.StatusInternalServerError, err
		}
		return renderJSON(w, r, map[string]any{"table": &next})
	})
}

func findSlot(t *cnc.ToolTable, slot int) *cnc.ToolTableSlot {
	for i := range t.Slots {
		if t.Slots[i].Slot == slot {
			return &t.Slots[i]
		}
	}
	return nil
}

func clonePtr(p *float64) *float64 {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}
