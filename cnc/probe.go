package cnc

// Tool-table discovery probe — see docs/TOOL_TABLE_RESEARCH.md.
//
// Operator runs this once at the controller to confirm which Haas
// macro-variable mapping the firmware uses for tool length / wear /
// diameter. Output drives the next phase of the tool-table feature
// (live tool index, life warnings, diameter-check on Send).
//
// Implementation: fire Q600 over the four canonical NGC bases
// (length=2001, length-wear=2201, diameter=2401, diameter-wear=2601)
// for a sample of tool slots. Each query goes through Streamer.Query
// so queryMu serialization + minQuerySpacing apply — the bridge sees
// at most one in-flight query at a time, same as normal polling.
//
// At 150 ms spacing × 4 bases × N slots, the probe takes
// ~0.6 s per slot. Sampling 10 slots is ~6 seconds; 30 slots is
// ~18 seconds. Operator-triggered, not a steady-state poller.

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ToolProbeBaseResult is per-macro-var-base summary of the probe
// across the sampled slots.
type ToolProbeBaseResult struct {
	Base       int                   `json:"base"`
	Label      string                `json:"label"`
	OK         int                   `json:"ok"`         // count of successful round-trips
	Empty      int                   `json:"empty"`      // count returning 0.0 (empty pocket)
	Errors     int                   `json:"errors"`     // count with non-empty error
	Samples    []ToolProbeSlotResult `json:"samples"`    // first few raw values for inspection
	FirstError string                `json:"first_error,omitempty"`
}

// ToolProbeSlotResult is one slot's reading on one base.
type ToolProbeSlotResult struct {
	Slot  int    `json:"slot"`
	Var   int    `json:"var"`   // base + (slot-1)
	Value string `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

// ToolProbeReport is the full structured response from POST /api/cnc/probe-tools.
type ToolProbeReport struct {
	Slots          int                   `json:"slots_probed"`
	DurationMs     float64               `json:"duration_ms"`
	BridgeAddress  string                `json:"bridge_address"`
	Bases          []ToolProbeBaseResult `json:"bases"`
	Verdict        string                `json:"verdict"`
	Recommendation string                `json:"recommendation"`
}

// toolProbeBases is the set we sample. Order matters for output
// readability — length first, then diameter, with wear after each.
var toolProbeBases = []struct {
	Base  int
	Label string
}{
	{2001, "Tool length geometry"},
	{2201, "Tool length wear"},
	{2401, "Tool diameter geometry"},
	{2601, "Tool diameter wear"},
}

// ProbeTools runs the discovery probe across `slots` tool slots and
// classifies the result. slots must be in [1, 200] (Haas tool table
// max). The streamer must be idle — running during a job would
// contend with the streaming socket; the caller is expected to gate
// on Streamer.IsRunning().
func (s *Streamer) ProbeTools(ctx context.Context, slots int) (*ToolProbeReport, error) {
	if slots < 1 || slots > 200 {
		return nil, fmt.Errorf("slots must be 1..200, got %d", slots)
	}
	set, err := s.settings.Get()
	if err != nil {
		return nil, err
	}
	if set.Cnc.HaasHost == "" {
		return nil, ErrConfigMissing
	}
	if s.IsRunning() {
		return nil, fmt.Errorf("can't probe during a streaming job")
	}

	t0 := time.Now()
	rep := &ToolProbeReport{
		Slots:         slots,
		BridgeAddress: fmt.Sprintf("%s:%d", set.Cnc.HaasHost, set.Cnc.HaasPort),
	}
	s.logf("info", "starting tool-table probe over %d slots × %d bases", slots, len(toolProbeBases))

	const sampleCap = 5 // raw samples per base in the response (keeps payload small)

	for _, b := range toolProbeBases {
		baseRes := ToolProbeBaseResult{Base: b.Base, Label: b.Label}
		for slot := 1; slot <= slots; slot++ {
			v := b.Base + (slot - 1)
			res, qerr := s.Query(ctx, 600, &v)
			sample := ToolProbeSlotResult{Slot: slot, Var: v}
			switch {
			case qerr != nil:
				baseRes.Errors++
				sample.Error = qerr.Error()
				if baseRes.FirstError == "" {
					baseRes.FirstError = qerr.Error()
				}
			case !res.OK:
				baseRes.Errors++
				sample.Error = res.Error
				if baseRes.FirstError == "" {
					baseRes.FirstError = res.Error
				}
			default:
				baseRes.OK++
				sample.Value = res.Value
				if isEmptyMacroValue(res.Value) {
					baseRes.Empty++
				}
			}
			if slot <= sampleCap {
				baseRes.Samples = append(baseRes.Samples, sample)
			}
		}
		rep.Bases = append(rep.Bases, baseRes)
	}

	rep.DurationMs = sinceMs(t0)
	rep.Verdict, rep.Recommendation = classifyProbe(rep, slots)
	s.logf("info", "tool-table probe done in %.1f ms — verdict: %s", rep.DurationMs, rep.Verdict)
	return rep, nil
}

// isEmptyMacroValue returns true for a "MACRO, 2401, 0.0000"-style
// frame. Empty pockets read as numerically 0; we don't treat that as
// an error but we count it separately.
func isEmptyMacroValue(value string) bool {
	parts := splitAndTrim(value, ",")
	if len(parts) == 0 {
		return false
	}
	last := parts[len(parts)-1]
	last = strings.TrimSpace(last)
	if last == "" {
		return false
	}
	if last == "0" || last == "0.0" || last == "0.0000" || last == "0.000000" {
		return true
	}
	if v, ok := parseNumber(last); ok {
		switch n := v.(type) {
		case float64:
			return n == 0
		case int:
			return n == 0
		}
	}
	return false
}

// classifyProbe maps the per-base counts into a human verdict.
// Possible outcomes:
//
//   - "ngc-mapping-confirmed" — all four bases answered with at
//     least one non-zero value. NGC firmware, mapping is the one in
//     TOOL_TABLE_RESEARCH.md, scaffold the live tool index against
//     these ranges.
//   - "ngc-mapping-empty" — bases answer cleanly but every slot
//     reads 0. Tool table is empty (fresh controller / probe before
//     setup). Re-run with tools loaded.
//   - "legacy-mapping-suspected" — bases ERROR consistently. Likely
//     pre-NGC firmware using a different range. Open a bug with the
//     `first_error` text + Haas firmware version.
//   - "partial-coverage" — some bases work, some don't. Either
//     mixed-firmware quirk or a Setting 143 issue. Requires
//     investigation; report exact counts.
//   - "no-bridge" — every query timed out. Bridge unreachable.
//     Run /api/cnc/check first.
func classifyProbe(rep *ToolProbeReport, slots int) (string, string) {
	allErr := 0
	allOK := 0
	allEmpty := 0
	for _, b := range rep.Bases {
		allErr += b.Errors
		allOK += b.OK
		allEmpty += b.Empty
	}
	total := slots * len(rep.Bases)
	switch {
	case allErr == total:
		return "no-bridge", "Every query failed. Run /api/cnc/check before re-probing."
	case allOK == 0 && allErr > 0:
		return "legacy-mapping-suspected",
			"All bases errored but the bridge is reachable. Likely pre-NGC firmware. Capture firmware version and check Haas operator manual for that vintage's macro-var map."
	case allOK == total && allEmpty == total:
		return "ngc-mapping-empty",
			"NGC mapping confirmed but every slot reads 0 — tool table is empty. Load tools, re-probe."
	case allOK == total && allEmpty < total:
		return "ngc-mapping-confirmed",
			"NGC mapping confirmed; live tool index can be scaffolded against bases 2001/2201/2401/2601."
	default:
		return "partial-coverage",
			fmt.Sprintf("%d ok, %d errors, %d empty across %d queries — investigate per-base counts; check Setting 143 (Machine Data Collect) is ON.", allOK, allErr, allEmpty, total)
	}
}
