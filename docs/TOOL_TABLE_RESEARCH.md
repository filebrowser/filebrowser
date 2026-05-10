# Tool table integration — Q-code research

Status: research-only. The integration plan describes Phase A as
"confirm Q-code/macro-var access"; this doc captures the candidate
queries before we sit at the controller and try them.

User direction (2026-05-09): _"have an offline DB of tools and do a
tool swap action on the gcode IF cutter comp is applied before
sending. no other tool on the planet really does this, but it is
really pretty trivial."_

---

## What we want to read

For each tool slot 1..N (Haas typically supports up to 200):

| Datum | Why |
|-------|-----|
| Length offset (H[n]) | confirm physical setup matches program |
| Length wear | flag a tool wearing past acceptable |
| Diameter / radius offset (D[n]) | **load-bearing for diameter check on Send** |
| Diameter wear | flag a worn cutter |
| Tool life % remaining | warn before in-flight failure |
| Tool count (cycles) | cross-reference with life |
| Active tool number | read-out (we have this — Q201) |

## Haas macro variable map (per public Haas Mill / Lathe operator manual)

The Haas controller exposes its tool-table fields through the macro
variable space. Q600 with `<var>` returns one variable's value. Numbers
below are for **NGC controls** (post-2014 firmware); pre-2014 classic
controls use a different range — we'll need to detect by polling Q104
+ checking for "MACRO" frame shape.

### Tool length (mill: H register)

```
2001..2200   Tool length geometry, slots 1..200
2201..2400   Tool length wear, slots 1..200
```

`Q600 2001` reads tool 1's length geometry, `Q600 2201` reads its
wear.

### Tool diameter / radius (mill: D register)

```
2401..2600   Tool diameter (radius * 2) geometry, slots 1..200
2601..2800   Tool diameter wear, slots 1..200
```

`Q600 2401` reads tool 1's geometry diameter — **this is the value
we'd compare against the program's D-register.**

### Tool life (Advanced Tool Management — ATM)

ATM is optional on Haas mills. The data lives in 8000-range variables
when ATM is enabled, indexed by the active tool group. Likely needs
both:

```
8500..8699   Tool group ID per tool slot (0 if not in a group)
8550..8649   Group's allowed-life metric
8650..8749   Group's accumulated count
```

These ranges need verification per controller — the Haas manual's
ATM appendix is the authoritative source. If ATM isn't installed,
queries return blank/error and we fall back to tool wear thresholds
as a proxy.

### Active tool

```
Q201         current tool number  (already polled — name `tool`)
4120         current tool, alternate read via macro
```

## Discovery probe

Before scaffolding the live tool index, validate each range exists on
the user's TM-2P. Run this once during Phase A:

```go
// pseudo, against streamer.Query
for slot := 1; slot <= 30; slot++ {  // sample, not all 200
    for _, base := range []int{2001, 2201, 2401, 2601} {
        v := base + (slot - 1)
        res := streamer.Query(ctx, 600, &v)
        log.Printf("slot %d base %d → ok=%v value=%q", slot, base, res.OK, res.Value)
    }
}
```

Outcomes:
- All four return clean MACRO frames → NGC mapping confirmed.
- Some ranges blank/zero → empty pockets, fine.
- All return errors → classic-control range, switch to legacy map.
- Cross-talk (e.g. tool length showing diameter values) → mapping bug
  in our parser, not real data.

Run with the controller idle. **Don't** run during a job — even with
PR #36's polling pause we shouldn't fan out 100+ Q-codes during DNC.

## Candidate features (in order of risk)

### A. Live tool index (read-only)

A panel on /machine that lists slot # → diameter / length / wear /
active. Aggregator gets new metric specs (Q600 over each slot's
diameter), polled at a slow interval (60 s). Memory: 200 slots × ~80
bytes each = 16 KB, trivial.

### B. Tool-life warnings

Cron-like check (every 30 s) compares each tool's wear / life count
vs. configurable thresholds. Surfaces in Activity log:
`[warn] tool 5 at 92% life — replace soon`.

Threshold config goes in `settings.Cnc` per Machine (per-machine
because tool inventory varies).

### C. Diameter-check on Send (the novel piece)

Before posting `/api/cnc/start`, we already fetch the file content for
the NC mirror. New step: parse the program for `D` register references,
look up each in the live tool table, reject Send if any D value
disagrees with the slot's actual diameter beyond a tolerance.

Edge cases:
- **Tool not in table**: hard reject. "D5 referenced; slot 5 empty."
- **Comp mode never invoked** (`G40` only): skip the check.
- **Wear-only mismatch** (small): warn but allow — operator sets wear
  intentionally.

UI: Send confirm dialog grows a "Tool diameter check" section with
green checks per D-register or red errors. Operator can override
("Send anyway") for non-critical mismatches; hard rejects (empty slot)
have no override.

### D. Offline tool DB + swap-on-send

Operator maintains a JSON catalog of physical tools (name, diameter,
length, vendor, in-stock). On Send, if a D mismatches:
- Suggest "swap to slot X" if a matching physical tool sits in another
  slot
- Patch the program in-flight (NOT on disk — bytes substituted as
  they go out the wire to the bridge)

This is the bit the user called out as "no other tool on the planet
does this". Implementation: extend the streamer's per-line write to
optionally pass through a substitution callback that rewrites
`Dn` references for the active swap.

### E. Cutter-comp scan

Static analysis of the program: track G41/G42 (cutter comp on/off),
flag invalid entry/exit, missing G40 before tool change, etc. Needs
a real G-code parser (not just regex). Defer until A-D are proven.

## Storage / state

Live tool data is per-machine and ephemeral (mirrors the controller's
state). Stash in a per-Machine `ToolTable` field on the Aggregator,
populated by polling, exposed via `GET /api/cnc/tools?machine_id=...`.
Don't persist — it's always re-readable from the controller, and a
stale persist could ship the wrong values.

Offline DB (Phase D) is operator-owned, persisted as JSON in the share
or BoltDB. Path TBD.

## Open questions

- **Polling cost.** 200 tool slots × 4 fields = 800 macro queries.
  At 150ms spacing = 2 minutes per full sweep. Probably fine if we
  only sweep every 60-120 s, but worth confirming the Haas can sustain
  that traffic without choking the user's display.
- **Lathe vs mill.** Haas lathes use different macro ranges. The user
  has a TM-2P (mill); deferring lathe support until a lathe shows up.
- **Subprograms / call.** Diameter-check needs to follow `M97/M98`
  calls and look at D-references in subprograms, not just the main.
  Probably means staging the file content + the called-program
  contents through the parser before Send.
