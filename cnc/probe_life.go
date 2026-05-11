package cnc

// Tool-life discovery probe — see docs/TOOL_LIFE_RESEARCH.md.
//
// Scans a macro-variable range and reports what populates. Run after
// known tool changes so the operator can spot which macros carry
// cycle counts / life timers on their specific Haas firmware. Once
// confirmed, the macro range gets pinned in tooltable.go's pass logic
// and per-tool cycle counts can be surfaced in the tool table.
//
// Implementation: operator-triggered, fires Q600 #N for each macro in
// [start, end] stepping by step. Same queryMu serialization +
// minQuerySpacing as the rest of the read path; refuses while a
// streaming job is running. Range size is clamped to 500 macros so a
// runaway "1..30000" probe can't wedge the bridge for an hour.

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// ToolLifeProbeSample is one macro's readout.
type ToolLifeProbeSample struct {
	Macro int    `json:"macro"`
	Value string `json:"value,omitempty"` // raw "MACRO, N, x.xxxx" frame
	// Parsed numeric tail when the frame was a clean MACRO response.
	// Pointer so the JSON distinguishes "could not parse" (omitted)
	// from "read 0" (present, value 0).
	Number *float64 `json:"number,omitempty"`
	Error  string   `json:"error,omitempty"`
}

// ToolLifeProbeReport is the response from POST /api/cnc/probe-tool-life.
type ToolLifeProbeReport struct {
	Start         int     `json:"start"`
	End           int     `json:"end"`
	Step          int     `json:"step"`
	Probed        int     `json:"probed"`   // count of macros queried
	OK            int     `json:"ok"`       // count of clean reads
	Empty         int     `json:"empty"`    // count of clean reads where number == 0
	NonZero       int     `json:"non_zero"` // count of clean reads where number != 0
	Errors        int     `json:"errors"`   // count of failed queries
	DurationMs    float64 `json:"duration_ms"`
	BridgeAddress string  `json:"bridge_address"`
	// Samples are sorted with non-zero entries first (most interesting
	// for the operator) followed by errors, then zeros, then anything
	// else. Capped at SampleCap entries to keep the payload small
	// even for a 500-macro probe — the counts above describe the full
	// scan.
	Samples []ToolLifeProbeSample `json:"samples"`
	// Clusters describes contiguous runs of non-zero macros (gap <= step).
	// A per-slot tool-life counter table will show up here as one long
	// cluster, which is the single most useful summary for the operator
	// trying to pin a macro range.
	Clusters       []MacroCluster `json:"clusters,omitempty"`
	Verdict        string         `json:"verdict"`
	Recommendation string         `json:"recommendation"`
}

// MacroCluster describes a contiguous run of macros that carry non-zero
// values. Operators reading the probe report want runs like "3122..3141
// (20 entries)" surfaced as one row, not 20 individual hits.
type MacroCluster struct {
	Start int `json:"start"`
	End   int `json:"end"`
	Count int `json:"count"`
}

// FindNonZeroClusters scans an UNSORTED-by-macro samples slice (the
// probe report sorts by interest, so we re-sort by macro here) and
// returns contiguous runs where each non-zero sample is within `step`
// of the previous non-zero entry. Zero values and errors break a
// cluster. step <= 0 is treated as 1.
func FindNonZeroClusters(samples []ToolLifeProbeSample, step int) []MacroCluster {
	if step <= 0 {
		step = 1
	}
	// Copy + sort ascending by macro so detection works against the
	// raw scan order, not the by-interest sort the report applies.
	type entry struct {
		macro int
		hot   bool
	}
	entries := make([]entry, 0, len(samples))
	for _, s := range samples {
		hot := s.Number != nil && *s.Number != 0
		entries = append(entries, entry{macro: s.Macro, hot: hot})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].macro < entries[j].macro })
	out := []MacroCluster{}
	var cur *MacroCluster
	for _, e := range entries {
		if !e.hot {
			cur = nil
			continue
		}
		if cur != nil && e.macro-cur.End <= step {
			cur.End = e.macro
			cur.Count++
			continue
		}
		out = append(out, MacroCluster{Start: e.macro, End: e.macro, Count: 1})
		cur = &out[len(out)-1]
	}
	return out
}

// Per-call ceilings. The bridge is shared with the streamer's poll
// rail, so a probe that scans the entire macro space would starve
// every other consumer for the duration. 500 macros × 150 ms ≈ 75 s,
// which is the long edge of acceptable for an operator-triggered
// scan.
const (
	toolLifeProbeMaxRange  = 500
	toolLifeProbeSampleCap = 60
	// Default candidates when the caller doesn't pass a range. Covers
	// the most-cited Haas tool-monitor macros (3120..3199) plus the
	// 3000-3030 region (alarm + timers + parts counters) so the
	// operator can compare a known-populated counter (e.g. #3022,
	// M30 parts) against the unknown tool-life ones.
	toolLifeProbeDefaultStart = 3000
	toolLifeProbeDefaultEnd   = 3199
)

// ProbeToolLife scans macros [start, end] stepping by step and
// returns per-macro samples. start <= 0 / end <= 0 / step <= 0 trip
// defaults.
func (s *Streamer) ProbeToolLife(ctx context.Context, start, end, step int) (*ToolLifeProbeReport, error) {
	if start <= 0 {
		start = toolLifeProbeDefaultStart
	}
	if end <= 0 {
		end = toolLifeProbeDefaultEnd
	}
	if step <= 0 {
		step = 1
	}
	if end < start {
		return nil, fmt.Errorf("end (%d) must be >= start (%d)", end, start)
	}
	count := ((end - start) / step) + 1
	if count > toolLifeProbeMaxRange {
		return nil, fmt.Errorf(
			"range too large: %d macros at step %d would scan %d entries (cap %d) — narrow the range and re-run",
			end-start+1, step, count, toolLifeProbeMaxRange)
	}

	m, port, rerr := s.resolveMachine()
	if rerr != nil {
		return nil, rerr
	}
	if s.IsRunning() {
		return nil, fmt.Errorf("can't probe during a streaming job")
	}

	t0 := time.Now()
	rep := &ToolLifeProbeReport{
		Start:         start,
		End:           end,
		Step:          step,
		BridgeAddress: fmt.Sprintf("%s:%d", m.Host, port),
	}
	s.logf("info", "starting tool-life probe over macros %d..%d step %d (%d entries)", start, end, step, count)

	all := make([]ToolLifeProbeSample, 0, count)
	for v := start; v <= end; v += step {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return rep, fmt.Errorf("tool-life probe cancelled: %w", ctxErr)
		}
		macro := v
		res, qerr := s.Query(ctx, 600, &macro)
		samp := ToolLifeProbeSample{Macro: macro}
		switch {
		case qerr != nil:
			rep.Errors++
			samp.Error = qerr.Error()
		case !res.OK:
			rep.Errors++
			samp.Error = res.Error
		default:
			rep.OK++
			samp.Value = res.Value
			if n, ok := parseFloatTail(res.Value); ok {
				samp.Number = &n
				if n == 0 {
					rep.Empty++
				} else {
					rep.NonZero++
				}
			}
		}
		all = append(all, samp)
	}
	rep.Probed = len(all)

	// Compute clusters BEFORE the interest-sort + cap so a long run of
	// populated macros is reflected even if the operator's sample list
	// gets truncated for payload size.
	rep.Clusters = FindNonZeroClusters(all, step)

	// Sort: non-zero numbers first (most interesting for the operator),
	// then errors, then zero values, then anything else. Stable sort
	// so ties keep ascending macro order — easier to spot a base+slot
	// arithmetic series in the report.
	sort.SliceStable(all, func(i, j int) bool {
		ri, rj := lifeSampleRank(all[i]), lifeSampleRank(all[j])
		if ri != rj {
			return ri < rj
		}
		return all[i].Macro < all[j].Macro
	})
	if len(all) > toolLifeProbeSampleCap {
		all = all[:toolLifeProbeSampleCap]
	}
	rep.Samples = all

	rep.DurationMs = sinceMs(t0)
	rep.Verdict, rep.Recommendation = classifyLifeProbe(rep)
	s.logf("info", "tool-life probe done in %.1f ms — %d non-zero, %d empty, %d errors",
		rep.DurationMs, rep.NonZero, rep.Empty, rep.Errors)
	return rep, nil
}

// lifeSampleRank — lower rank sorts first. 0 = non-zero number
// (interesting), 1 = error (legacy / unsupported macro), 2 = zero
// (unused), 3 = parsed-but-unclassified.
func lifeSampleRank(s ToolLifeProbeSample) int {
	switch {
	case s.Number != nil && *s.Number != 0:
		return 0
	case s.Error != "":
		return 1
	case s.Number != nil && *s.Number == 0:
		return 2
	default:
		return 3
	}
}

func classifyLifeProbe(rep *ToolLifeProbeReport) (string, string) {
	if rep.Probed == 0 {
		return "empty-range", "No macros probed."
	}
	if rep.Errors == rep.Probed {
		return "no-bridge",
			"Every query failed. Run /api/cnc/check before re-probing."
	}
	if rep.NonZero == 0 {
		return "all-zero",
			"Range answered cleanly but every macro reads 0. Either the wrong range, or the controller hasn't accumulated any tool-life data yet. Run a known tool change and re-probe, or try a different range."
	}
	hot := strings.Builder{}
	for i, s := range rep.Samples {
		if s.Number == nil || *s.Number == 0 {
			break
		}
		if i > 0 {
			hot.WriteString(", ")
		}
		fmt.Fprintf(&hot, "#%d=%g", s.Macro, *s.Number)
		if i >= 4 {
			hot.WriteString(", …")
			break
		}
	}
	// Cluster line — if there's a contiguous run of populated macros,
	// surface it explicitly. A per-slot counter table almost always
	// shows up as one cluster spanning the magazine size.
	clusterNote := ""
	if len(rep.Clusters) > 0 {
		best := rep.Clusters[0]
		for _, c := range rep.Clusters[1:] {
			if c.Count > best.Count {
				best = c
			}
		}
		if best.Count >= 3 {
			clusterNote = fmt.Sprintf(
				" Longest run: #%d..#%d (%d entries) — likely a per-slot table.",
				best.Start, best.End, best.Count)
		}
	}
	return "candidates-found",
		fmt.Sprintf("Non-zero macros: %s.%s Cross-reference each against a known tool change (run a few cycles, re-probe, look for the macro that incremented).",
			hot.String(), clusterNote)
}
