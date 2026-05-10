# Tool-life cycle counts — open research

**Status:** open question, no read code yet. The `ToolTableSlot.cycle_count`
field exists with `omitempty` so the JSON schema is forward-compatible
once the macro mapping is confirmed against a live machine.

## What we want

For each loaded tool, surface:

- **Cycles since last reset** — how many times `M06 T<n>` has selected
  the tool (or, equivalently, how many tool changes the tool has been
  through). Lets the operator spot "T5 has 1247 cycles, oldest in the
  magazine" before a long unattended run.
- **Cumulative use time** (optional) — total spindle-on time the tool
  has accumulated, when Setting 22 (Tool Life Monitoring) is enabled.
- **Life remaining** (optional) — when an operator has set a per-tool
  life limit, the remaining count / time before the controller will
  alarm out and refuse the next selection.

## Why this isn't shipped yet

The macro mapping is **not in the Haas docs we have on hand** and the
ranges most commonly cited online vary by Next Gen Control firmware
revision. Candidates that have been mentioned in scattered references:

| Range (rumored)   | Notes                                                |
|-------------------|------------------------------------------------------|
| `#3196` / `#3197` | Tool life set / max time — single-tool, not per-slot |
| `#5300`–`#5403`   | Probing macros — likely **not** tool life            |
| `#8500`–`#8552`   | Cutter-comp / stop / restart info — not life         |
| `#3122` etc.      | Sometimes cited as per-tool use count — unverified   |

Reading speculative macros over the bridge during a real job is
unsafe (every Q600 round-trip costs 150 ms of write turn). So the
right next step is an operator-triggered **discovery probe**, modeled
on `cnc/probe.go` for the tool-table, that scans a candidate range
and reports which macros populate.

## Plan

1. Add `cnc/probe_life.go` with `ProbeToolLife(ctx, slots, baseStart, baseEnd, step)`.
   Same shape as `ProbeTools`: serializes through `queryMu`, reports per-base
   sample values + ok/empty/error counts.
2. Wire a `POST /api/cnc/probe-tool-life` handler.
3. Add a "Probe tool life" button to Settings → Machine, next to the
   existing "Probe tool table" button.
4. Run on the live Haas. Confirm which range carries non-zero values
   when known tools have been used. Pin the base in code.
5. Extend `tooltable.go` with a third pass that reads the confirmed
   life macro for populated slots only (mirrors the geom passes).
6. Surface the count in `ToolTablePanel.vue` (per-row column +
   magazine label) and in the preflight comparison ("T5 has 1247
   cycles — second-oldest in the magazine").

## Interim placeholder

The schema field is reserved so a future PR can land the actual read
without a JSON migration:

```go
type ToolTableSlot struct {
    // …existing fields…
    // CycleCount is the tool's accumulated select-count since last
    // reset. Populated only after the tool-life probe lands; absent
    // means "not read" (NOT zero).
    CycleCount *int `json:"cycle_count,omitempty"`
}
```

The frontend renders the column only when at least one slot has the
field set, so the UI doesn't show empty cycle columns until the read
is real.
