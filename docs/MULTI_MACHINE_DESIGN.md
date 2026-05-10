# Multi-machine support — design + migration plan

Status: design-only. No code change yet. Ship Phase 1 when there's a
second physical machine on the network to test against.

---

## Why

Today: `settings.Cnc` is a single struct (`HaasHost`/`HaasPort`/etc).
One Streamer, one Aggregator, one Q-code dialect baked through the
codebase. User direction (2026-05-09): _"should eventually be machine
agnostic and allow multiple and different brands. Along with
destination selection on send."_

## Constraints / non-goals

- **Backwards compat for existing single-machine installs.** Old DBs
  must auto-migrate to a single-entry array on first boot of the new
  binary. No operator intervention.
- **No multi-tenant access control.** Every logged-in operator can
  see/send to every machine. Per-machine permissions are a separate
  problem if it ever comes up.
- **Same-network assumption.** Each machine has a network-reachable
  bridge (Waveshare or equivalent). Different brands behind the same
  bridge type all good; different bridge types behind one machine is
  weird and not in scope.

## Schema

`settings.Cnc` becomes:

```go
type Cnc struct {
    // Long-lived bearer for S2S API access. NOT per-machine —
    // external consumers see all machines under the same auth.
    MachineToken string `json:"machineToken"`

    // The active machine list. First entry is treated as default
    // for any endpoint called without a machine_id.
    Machines []Machine `json:"machines"`
}

type Machine struct {
    ID       string `json:"id"`        // ulid; stable across renames
    Name     string `json:"name"`      // human label, freely editable
    Brand    string `json:"brand"`     // "haas" | "fanuc" | "mazak" | "generic"
    Host     string `json:"host"`
    Port     int    `json:"port"`
    CameraURL string `json:"cameraUrl,omitempty"`

    // Capabilities — populated at config time so the UI knows what
    // tiles to show. Brand-specific quirks (eg. STEP-NC dialect)
    // hang off here without leaking into the struct surface.
    Capabilities MachineCaps `json:"capabilities,omitempty"`
}

type MachineCaps struct {
    QCode  bool `json:"qCode"`   // can answer ?Q-style queries
    DPRNT  bool `json:"dprnt"`   // can stream DPRNT log lines
    Stream bool `json:"stream"`  // accepts inbound DNC G-code
}
```

Migration: on Settings load, if `Machines` is nil/empty AND the legacy
fields are populated, synthesise one Machine entry from
`{HaasHost, HaasPort, CameraURL}` with brand=`haas`, ID=`legacy`,
all caps true. Save back. Existing API behavior is unchanged on
single-machine setups.

## Protocol abstraction

```go
// MachineProtocol is the wire-level contract every brand
// implementation must satisfy. The Streamer + Aggregator only talk
// to this interface — brand-specific code lives behind it.
type MachineProtocol interface {
    // Idle round-trip: open conn, send query, read response, close.
    // Caller serializes via the Streamer's queryMu; impl just does
    // the wire dance.
    Query(ctx context.Context, host string, port int, key string, arg *string) (*QueryResult, error)

    // Streaming a job: caller hands over an open net.Conn; impl
    // writes the DNC payload (line-by-line for Haas, possibly
    // chunked or with handshakes for others) and reports progress
    // via the lineEmit callback.
    Stream(ctx context.Context, conn net.Conn, src io.Reader, lineEmit func(n int64, text string)) error

    // Validate that a response value plausibly matches the query
    // we sent. Catches cross-talk per #35. Brand-specific because
    // each protocol has different framing tags.
    Validate(key string, arg *string, value string) error

    // Default metric specs to populate the Aggregator with. Brand
    // chooses what's worth polling.
    DefaultMetrics() []MetricSpec
}
```

The Haas implementation is what's already in `cnc/qcode.go` +
`cnc/state.go::defaultMetricSpecs`, just lifted behind the interface.
`key` for Haas is the Q-code as a string (`"104"`, `"600"`); `arg` is
the macro var. For Fanuc / Mazak the keys would be different
(probably mnemonic instead of numeric), but the Streamer doesn't care.

## Per-machine Streamer + Aggregator

Today's Streamer owns one socket. Multi-machine needs one Streamer
**instance** per Machine, keyed by ID. A simple registry:

```go
type Registry struct {
    mu       sync.RWMutex
    streamers map[string]*Streamer    // by Machine.ID
    aggregators map[string]*Aggregator
    settings settingsReader
}
```

`Registry.Get(id)` returns the per-machine pair. On settings change
(Machine added/removed/edited), Registry diffs and creates/destroys
instances accordingly. Aggregators live forever as long as the Machine
exists; Streamers live for the lifetime of a job.

## API surface

Add `?machine_id=` (or path param) everywhere. Default to first machine
when omitted, for backwards compat.

```
GET  /api/cnc/state?machine_id=zinc-tm2p
POST /api/cnc/qcode?machine_id=zinc-tm2p
POST /api/cnc/start            { machine_id, file_path }
POST /api/cnc/stop?machine_id=zinc-tm2p
WS   /api/cnc/stream?machine_id=zinc-tm2p
GET  /api/cnc/check?machine_id=zinc-tm2p
GET  /api/cnc/siblings?path=...&machine_id=zinc-tm2p
GET  /api/cnc/machines       (new — list configured Machines)
```

## UI

- **Settings → Machines**: list view with add/edit/delete. Per-row
  brand + host/port + camera + capability toggles.
- **Editor's Send button**: today is single-target. Becomes a split
  button when `machines.length > 1`: primary action sends to the
  default; dropdown picks any other.
- **/machine**: top-level switcher (tab strip or dropdown) to change
  which machine the dashboard renders. Per-machine route segment:
  `/machine/:id` (default route redirects to first machine).
- **CNC store** (`stores/cnc.ts`): becomes a `Map<machineId, CncState>`.
  Each WS opens against its own `?machine_id=` URL.

## Order of operations

1. **MachineProtocol interface + Haas adapter** behind it. No behavior
   change. Single Streamer/Aggregator continues working through the
   adapter. Refactor PR; should be net-zero in functionality.
2. **Schema migration** to `Machines[]` with auto-upgrade from the
   legacy single-machine struct. API adds an optional `?machine_id=`
   param; missing param → default to first machine.
3. **Registry of Streamer + Aggregator instances** keyed by Machine.ID.
   Refactor of route handlers to look up the right instance.
4. **Frontend store keyed by machine ID**. Add the `/machine/:id`
   route segment and the switcher UI.
5. **Send-destination dropdown** in the editor. Toast/route to the
   selected machine.
6. **Second brand**: pick a real one (Fanuc most likely — has documented
   DPRNT-equivalent + macro vars). Implement MachineProtocol for it.
   Validate on hardware.

## Open questions

- **Brand auto-detect**: probe Q104 on first connect; if it answers
  with a Haas-shape frame, set brand=haas. Worth doing? Probably not
  — operator sets it once at config time. Detection helps surface
  errors ("you marked this as Fanuc but it answers Haas-style").
- **Per-machine bearer tokens?** Today there's one MachineToken for
  S2S. Could split per-machine if HA wants only mode-X access. Defer
  until someone asks.
- **Fleet view**: dashboard tile per machine showing running/idle +
  current line. Different from the per-machine /machine view.
  Probably ships as `/fleet` route. After Phase 5.
