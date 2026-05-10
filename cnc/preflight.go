package cnc

// Pre-flight tool check — parse the NC source for tool references,
// compare against the machine's latest persisted tool table, and
// classify each tool as ok / warn / empty / offline / missing.
//
// Surfaced through GET /api/cnc/preflight by the SendWizard. Read-only;
// the streamer is not touched.

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ToolUsage is one tool the NC program references. Fields after Comment
// describe what the program EXPECTS (parsed from the comment block at
// the top of the file); fields after InTable describe what the
// controller's last tool-table read reported.
type ToolUsage struct {
	Tool                 int      `json:"tool"`
	ReferenceCount       int      `json:"reference_count"`
	Comment              string   `json:"comment,omitempty"`
	ExpectedDiameter     *float64 `json:"expected_diameter,omitempty"`
	ExpectedCornerRadius *float64 `json:"expected_corner_radius,omitempty"`
	// Tool-table state for the slot whose number matches Tool.
	InTable        bool     `json:"in_table"`
	Loaded         bool     `json:"loaded"`       // table read OK + non-zero length
	EmptyPocket    bool     `json:"empty_pocket"` // table read OK but every value 0
	Offline        bool     `json:"offline"`      // table read errored on this slot
	ActualDiameter *float64 `json:"actual_diameter,omitempty"`
	DiameterDelta  *float64 `json:"diameter_delta,omitempty"`
	Status         string   `json:"status"`        // ok / warn / empty / offline / missing
	StatusReason   string   `json:"status_reason,omitempty"`
}

// PreflightSummary counts each status across Tools.
type PreflightSummary struct {
	OK      int `json:"ok"`
	Warn    int `json:"warn"`
	Empty   int `json:"empty"`
	Offline int `json:"offline"`
	Missing int `json:"missing"`
}

// Preflight is the response shape of GET /api/cnc/preflight.
type Preflight struct {
	FilePath     string           `json:"file_path"`
	MachineID    string           `json:"machine_id"`
	Tools        []ToolUsage      `json:"tools"`
	TableReadAt  string           `json:"table_read_at,omitempty"`
	TableMissing bool             `json:"table_missing,omitempty"`
	Summary      PreflightSummary `json:"summary"`
}

// DiameterTolerance is how far an effective_diameter may drift from
// the expected D= in the comment before we flag it warn. 0.005" is
// generous for a setup pre-flight (cutter wear on a long-running
// tool can creep that far without anyone noticing). Operators who
// want tighter can grep for "DiameterTolerance" and rebuild — env
// override would be future work.
const DiameterTolerance = 0.005

// Top-of-file CAM tool-list line:
//
//	(T5 D=0.5 CR=0.06 - ZMIN=-1.87 - bullnose end mill)
//
// Single capture group around the inside of the parens.
var commentLine = regexp.MustCompile(`\(([^)]+)\)`)

// One tool reference outside comments:
//
//	N30 T14 M6
//	    ^^^
//
// Word-boundary on both sides so "F100T0" / "M6T1" both still match.
// Excludes `Txxx` inside comments because the comment text is
// stripped first.
var toolCall = regexp.MustCompile(`(?i)\bT(\d{1,3})\b`)

// Tool comment header inside a comment block:
//
//	T5 D=0.5 CR=0.06 - bullnose end mill
//	^^^^^
//
// Anchored at start to distinguish from inline references.
var toolCommentHeader = regexp.MustCompile(`(?i)^\s*T(\d{1,3})\s*(.*)$`)

var diamRe = regexp.MustCompile(`(?i)\bD\s*=\s*([0-9.]+)`)
var cornerRe = regexp.MustCompile(`(?i)\bCR\s*=\s*([0-9.]+)`)

// BuildPreflight reads the NC file at absPath, parses tool usage, and
// joins it against `table` (typically the most recent persisted
// tool-table dump for `machineID`). table may be nil — in that case
// every used tool reports status="missing" with TableMissing=true on
// the result, which the wizard surfaces as "no tool-table read on
// file; you're flying blind". The function never errors on parse —
// malformed NC just produces an empty Tools list.
func BuildPreflight(absPath, displayPath, machineID string, table *ToolTable) (*Preflight, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	usage := map[int]*ToolUsage{}
	get := func(n int) *ToolUsage {
		if u, ok := usage[n]; ok {
			return u
		}
		u := &ToolUsage{Tool: n}
		usage[n] = u
		return u
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()

		// Pass 1 — pull comments out for the tool-list header parse.
		for _, c := range commentLine.FindAllStringSubmatch(line, -1) {
			body := strings.TrimSpace(c[1])
			if m := toolCommentHeader.FindStringSubmatch(body); m != nil {
				n, err := strconv.Atoi(m[1])
				if err != nil {
					continue
				}
				u := get(n)
				if u.Comment == "" {
					u.Comment = strings.TrimSpace(m[2])
				}
				if d := diamRe.FindStringSubmatch(body); d != nil && u.ExpectedDiameter == nil {
					if v, err := strconv.ParseFloat(d[1], 64); err == nil {
						u.ExpectedDiameter = &v
					}
				}
				if c2 := cornerRe.FindStringSubmatch(body); c2 != nil && u.ExpectedCornerRadius == nil {
					if v, err := strconv.ParseFloat(c2[1], 64); err == nil {
						u.ExpectedCornerRadius = &v
					}
				}
			}
		}

		// Pass 2 — count actual T-references on a comment-stripped
		// version of the line. Avoids "(T5 D=0.5)" producing two refs.
		stripped := commentLine.ReplaceAllString(line, "")
		for _, m := range toolCall.FindAllStringSubmatch(stripped, -1) {
			n, err := strconv.Atoi(m[1])
			if err != nil {
				continue
			}
			get(n).ReferenceCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", absPath, err)
	}

	// Drop tools that ONLY appeared in a comment block but were never
	// called — they're documentation, not actually used.
	for n, u := range usage {
		if u.ReferenceCount == 0 {
			delete(usage, n)
		}
	}

	// Sort by tool number for stable output.
	tools := make([]ToolUsage, 0, len(usage))
	for _, u := range usage {
		tools = append(tools, *u)
	}
	sort.Slice(tools, func(i, j int) bool { return tools[i].Tool < tools[j].Tool })

	pf := &Preflight{
		FilePath:  displayPath,
		MachineID: machineID,
		Tools:     tools,
	}

	if table == nil {
		pf.TableMissing = true
		// Without a table, every tool is missing. Don't classify
		// further — wizard renders the "no read on file" hint.
		for i := range pf.Tools {
			pf.Tools[i].Status = "missing"
			pf.Tools[i].StatusReason = "no tool-table read on file"
			pf.Summary.Missing++
		}
		return pf, nil
	}

	pf.TableReadAt = table.ReadAt.Format("2006-01-02T15:04:05Z07:00")
	tableSlots := map[int]*ToolTableSlot{}
	for i := range table.Slots {
		s := &table.Slots[i]
		tableSlots[s.Slot] = s
	}

	for i := range pf.Tools {
		t := &pf.Tools[i]
		slot, ok := tableSlots[t.Tool]
		if !ok {
			t.Status = "missing"
			t.StatusReason = fmt.Sprintf("slot %d not covered by last read (slots_requested=%d)", t.Tool, table.SlotsRequested)
			pf.Summary.Missing++
			continue
		}
		t.InTable = true
		// Offline beats empty: an errored read for this slot means we
		// don't really know anything about the pocket.
		if hasErrors(slot) {
			t.Offline = true
			t.Status = "offline"
			t.StatusReason = firstSlotError(slot)
			pf.Summary.Offline++
			continue
		}
		if slot.Empty {
			t.EmptyPocket = true
			t.Status = "empty"
			t.StatusReason = "table read OK, every value 0 — no tool in pocket"
			pf.Summary.Empty++
			continue
		}
		t.Loaded = true
		t.ActualDiameter = slot.EffectiveDiameter
		if t.ExpectedDiameter != nil && t.ActualDiameter != nil {
			delta := *t.ActualDiameter - *t.ExpectedDiameter
			t.DiameterDelta = &delta
			if math.Abs(delta) > DiameterTolerance {
				t.Status = "warn"
				t.StatusReason = fmt.Sprintf(
					"expected ⌀%.4f, table reports ⌀%.4f (Δ %+.4f)",
					*t.ExpectedDiameter, *t.ActualDiameter, delta,
				)
				pf.Summary.Warn++
				continue
			}
		}
		t.Status = "ok"
		pf.Summary.OK++
	}

	return pf, nil
}

func hasErrors(s *ToolTableSlot) bool {
	if s.Errors == nil {
		return false
	}
	// "Loaded but with one base errored" still counts as offline-ish.
	for _, msg := range s.Errors {
		if msg != "" {
			return true
		}
	}
	return false
}

func firstSlotError(s *ToolTableSlot) string {
	for k, v := range s.Errors {
		return fmt.Sprintf("%s: %s", k, v)
	}
	return ""
}
