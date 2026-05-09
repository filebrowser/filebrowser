# Haas Dashboard ↔ Zinc (filebrowser-NC) Integration Plan

> **Audience:** the Claude session working on `jasongainor/filebrowser-NC` (Zinc).
> This doc is self-contained — you don't need to read the haas-dashboard chat history.
> **Status:** decisions resolved 2026-05-09. **Zinc side is shipped as of 2026-05-09**; see "Status as of 2026-05-09" below for what landed and what changed in flight.

---

## Status as of 2026-05-09

Most of the original plan is in. One architectural pivot in flight (Z-11): the operator-facing dashboard is now a **native Vue page on Zinc** rather than an embedded iframe pointing at haas-dashboard. The iframe-specific dashboard-side todos (D-3 CORS, D-4 token-in-URL) are therefore obsolete; D-1 (proxy mode for haas-dashboard) is the only piece that keeps the bearer token earning its keep.

| Item | Status | PR(s) |
|---|---|---|
| Z-1 / Z-2 / Z-3 — Settings tab + GET/PUT + status stub | ✓ shipped | #11 |
| Z-4 — Streamer skeleton (single-job lock, start/stop/status) | ✓ shipped | #12 |
| Z-5 — Q-code multiplexer (`POST /api/cnc/qcode`) | ✓ shipped | #13 |
| Z-8 — WS event stream (`/api/cnc/stream`, line + status events) | ✓ shipped | #14 |
| Z-9 — Send-to-Machine button on the .nc viewer | ✓ shipped | #17 |
| Z-12 — Global status pill on the header | ✓ shipped | #18 |
| Z-10 — Machine tracker + follow-machine toggle | ✓ shipped | #19 |
| Z-11 — `/machine` page (architecture pivot — see below) | ✓ shipped natively | #20 → #24 |
| Z-13 — Camera embed (HLS / snapshot / RTSP-hint) | ✓ shipped | #20 |
| Z-15 — Crash-recovery prompt + ack | ✓ shipped (backend + UI) | #21, #22 |
| Z-14 — DPRNT log capture | ⏸ deferred — needs hardware to verify safely |  |
| `GET /api/cnc/state` — curated telemetry snapshot | ✓ shipped (new, not in original plan) | #24 |
| Branding assets out of share folder | ✓ shipped (new, not in original plan) | #23 |

Pi-side bugfixes that landed alongside (not in the original plan but blocking):

| Item | PR |
|---|---|
| `dr_mode=peripheral` forced under `[all]` (Bookworm imager seeds `[cm5]`) | #15 |
| `cnc-usb-watcher` initial sync at startup | #16 |
| Stop filebrowser before BoltDB writes (admin password) | #9 |

### Z-11 architectural pivot

Original plan: `/machine` is an iframe pointing at haas-dashboard with `?token=` for auth. Hung on D-3 + D-4 to ship.

What we shipped instead: `/machine` is a native filebrowser Vue page that polls `GET /api/cnc/state` (also new). The state endpoint is backed by a `cnc.Aggregator` background poller that hits each Q-code on its own ticker and caches the result, so the request path doesn't hit the Haas. **The Pi is now the sole owner of the "live state" view AND the data source.** External services that want the same telemetry call `/api/cnc/state` (or `/api/cnc/qcode` for ad-hoc) with the bearer token; haas-dashboard becomes one such consumer rather than the canonical UI.

That makes the bearer-token feature still useful (D-1 — haas-dashboard's HaasBridge can route through the Pi during a streaming job) but **removes the urgency of D-3 / D-4**. Marked obsolete below.

---

## TL;DR

Two systems exist today:

1. **`haas-dashboard`** (`/home/ubuntu/Repos/haas-dashboard/`) — FastAPI app that
   polls a Haas TM-2P over a Waveshare RS-232↔TCP bridge at
   `192.168.20.200:4196` and renders live telemetry. Reads `Q104` (mode),
   `Q500` (program/status/parts), `Q600 #5021/22/23` (machine pos),
   `Q600 #5041/42/43` (work pos), `Q600 #5221/22/23` (G54 offsets),
   `Q600 #3027/1815` (spindle), etc.
2. **Zinc / filebrowser-NC** (`/home/ubuntu/Repos/filebrowser-NC/`) — Go fork
   of filebrowser running on a Pi. Today it serves files to the Haas via **USB
   mass-storage emulation** (`g_mass_storage`, `cnc-usb-watcher`). It does
   **not** stream over RS-232 yet — `pi-setup/README.md` lists "G-code
   streaming" as a stretch stub.

Goal: add RS-232 drip-feed (DNC streaming), surface "current job" everywhere,
embed the dashboard inside filebrowser-NC, and let the operator click a "Send
to Machine" button on any `.nc` file.

---

## Resolved decisions (per user, 2026-05-09)

| # | Decision | Detail |
|---|---|---|
| 1 | **RS-232 always via the Waveshare bridge** | The Pi opens a TCP connection to `192.168.20.200:4196` and streams the file over that. Same wire path the dashboard uses today. **The Pi becomes the sole owner of that TCP port.** USB mass-storage mode (existing) stays — the operator can still copy files from the panel as if the Pi were a USB stick. The new RS-232 path is *additional*. |
| 2 | **Pause = Phase 2** | v1 ships without pause. Stop button can still cancel a job (simply stops feeding bytes; Haas drains its look-ahead and halts within seconds). |
| 3 | **Camera = whatever Unifi gives us** | Render a `<video>` (HLS) or `<img>` (MJPEG) tag with whatever URL the user pastes. If they paste an RTSP URL, we surface a "RTSP isn't browser-renderable, set up `go2rtc`" hint instead of trying to be clever. |
| 4 | **Filebrowser auth applies** | All `/api/cnc/*` endpoints check the filebrowser session. The embedded haas-dashboard inherits that auth — either by Zinc reverse-proxying it, or by filebrowser stamping a short-lived bearer token into the iframe URL. See API contract below. |
| 5 | **Config lives in filebrowser settings** | New **"Machine"** tab on the filebrowser settings page. Fields: `haas_host` / `haas_port` (defaults `192.168.20.200` / `4196`), `camera_url`, `haas_dashboard_url`. Persists to filebrowser-NC's settings store. |

**Follow-up confirmation needed:** the user said *"served via ssh to
waveshare"* — interpreting this as plain TCP from the Pi to the Waveshare
(SSH described how the user *used to* kick off streaming manually before
this UI). If actually an SSH tunnel is in the path, this plan needs a tweak.

---

## Architecture

```
                    ┌────────────────────────────────────────────┐
                    │  Pi @ zinc.local                           │
                    │                                            │
   USB-OTG ◄────────┤  filebrowser-NC                            │
   (existing,        │   ├─ files/                  (existing)   │
    parallel         │   ├─ /3d viewer              (existing)   │
    channel — the    │   ├─ /machine                (NEW iframe) │
    operator can     │   ├─ /settings#machine       (NEW tab)    │
    still copy       │   └─ /api/cnc/*              (NEW)        │
    files from the  │        ├─ status                          │
    pendant)         │        ├─ start / stop                    │
                    │        ├─ qcode  (proxy for haas-dashbd)  │
                    │        ├─ settings                         │
                    │        └─ /ws/cnc/stream                   │
                    │                                            │
                    └─────┬──────────────────────────────────────┘
                          │ TCP (sole client)
                          ▼
                    ┌──────────────────────────────┐
                    │ Waveshare bridge             │
                    │ 192.168.20.200:4196          │
                    └─────┬────────────────────────┘
                          │ RS-232
                          ▼
                    ┌──────────────────────────────┐
                    │ Haas TM-2P                   │
                    └──────────────────────────────┘

   ┌───────────────────────────────────────────┐
   │ haas-dashboard (home server)              │
   │ FastAPI /api/state, /ws                   │
   │ Q-code reads → POST zinc.local/api/cnc/   │ ◄── all serial access
   │                       qcode               │     proxied through Pi
   └───────────────────────────────────────────┘
```

**Why Pi-as-broker:** the Waveshare typically accepts only one TCP client at
a time. If the Pi holds it open during a 30-minute streaming job, the
dashboard's direct-to-Waveshare polling would fail for 30 minutes. Routing
the dashboard's Q-code reads through the Pi lets the Pi multiplex streaming
+ status queries on its single TCP socket — the dashboard stays live during
a job.

When **no job is running**, the Pi has no TCP connection to the Waveshare
held open. It opens a fresh transient connection per Q-code request (same
pattern the dashboard's `haas_bridge.py` uses today).

---

## API contracts (Zinc exposes these)

All endpoints require an authenticated filebrowser session. Browser clients
get this for free via cookies. The haas-dashboard backend gets a long-lived
bearer token from filebrowser settings (operator pastes it into
haas-dashboard env or settings UI — see `D-1` below).

```
GET  /api/cnc/status
  → 200 application/json
  {
    "running":      true,
    "file_path":    "/programs/part_0429_v3.nc",   // null when idle
    "file_url":     "/files/programs/part_0429_v3.nc",
    "line_current": 482,
    "line_total":   2043,
    "started_at":   "2026-05-09T18:42:11Z",
    "haas_ok":      true,
    "haas_last_error": null
  }

POST /api/cnc/start
  Body: { "file_path": "/programs/part_0429_v3.nc" }
  → 202 { "job_id": "..." }
  Streams the file over the Waveshare TCP connection. 409 if a job is
  already running. 404 if file doesn't exist under the share folder.

POST /api/cnc/stop
  → 200 { "stopped": true }
  Stops the stream. The Haas's look-ahead drains and the program halts
  within seconds. (Phase 2: also send M30 at the next line boundary.)

POST /api/cnc/qcode
  Body: { "q": 600, "var": 5021 }       // var optional
  → 200 {
      "q": 600, "var": 5021,
      "raw":    "?\r\n>Q600 5021\r\r\nMACRO, 5021,    -4.633600...",
      "value":  "MACRO, 5021,    -4.633600",
      "parsed": -4.6336,
      "duration_ms": 47.2
    }
  Single Q-code query. While a job is streaming, the Pi serializes this
  against the streaming write (pause feed → query → resume feed). When
  idle, opens a transient TCP connection per call.

GET  /api/cnc/settings
  → 200 {
      "haas_host":          "192.168.20.200",
      "haas_port":          4196,
      "camera_url":         "https://unifi.local/proxy/...",
      "haas_dashboard_url": "http://homeserver.tail.../"
    }

PUT  /api/cnc/settings
  Body: same shape as GET. Persists to filebrowser-NC config.

WS   /api/cnc/stream
  Pushes events as they happen:
    { "type": "line",   "n": 483, "text": "G1 X12.345 Y-3.0 F500" }
    { "type": "status", "running": true, ... }
    { "type": "log",    "level": "warn", "msg": "..." }
```

### Haas-dashboard endpoints (already exist; reused by Zinc)

```
GET /api/state           // full snapshot — Zinc embeds this in /machine
GET /api/raw?q=104       // diagnostic
WS  /ws                  // metric updates
POST /api/query          // {"q": 104} or {"q": 600, "var": 5021}
```

---

## TODOs — Zinc side (filebrowser-NC)

> The Zinc chat owns these. Pick them off in order — later items depend on
> earlier ones.

### Phase 1 — Settings + status stub

- [x] **Z-1. Add "Machine" tab to filebrowser settings.** Look at how the
  existing `/settings` page is composed (likely `frontend/src/views/settings/`)
  and add a tab next to whatever's already there. Fields:
  - `haas_host` (text, default `192.168.20.200`)
  - `haas_port` (number, default `4196`)
  - `camera_url` (text, optional; placeholder: `https://… HLS or MJPEG URL`)
  - `haas_dashboard_url` (text, default `http://homeserver.tail.../:8080`)
  - **Long-lived bearer token** (read-only display + "regenerate" button) —
    so the operator can paste it into the haas-dashboard's env to authorize
    its API calls.
  Persist via `settings.Storage`.
- [x] **Z-2. `GET/PUT /api/cnc/settings`** backed by the same store as Z-1.
  Both endpoints require `filebrowser.User.Perm.Admin` (matches the existing
  settings convention; check how the rest of `/settings` is gated).
- [x] **Z-3. Stub `GET /api/cnc/status`** that returns `{"running": false}`.
  Unblocks haas-dashboard development before streaming is built.

### Phase 2 — DNC streaming + Q-code proxy

- [x] **Z-4. Implement RS-232 streaming.** New service (`runner/cnc-stream/`?)
  that opens a TCP connection to `<haas_host>:<haas_port>` and feeds a file
  line-by-line.
  - **Wire format:** Same as the existing dashboard:
    `<line>\r\n` per line. The Haas DNC mode uses XON/XOFF flow control
    (Setting 14 governs sync mode — confirm with user that it's set to a
    DNC-compatible mode before testing).
  - **Send loop:** write bytes; if the Haas pushes XOFF (0x13), pause until
    XON (0x11). Track `line_current` only after `Write()` returns and the
    OS buffer is drained.
  - **Reference:** the haas-dashboard's
    [`haas_bridge.py`](../haas_bridge.py) has the framing/parser logic for
    Q-code responses (`STX 0x02 ... ETB 0x17`). Re-use the protocol details
    in Go.
- [x] **Z-5. Q-code multiplexing.** Implement `POST /api/cnc/qcode`.
  - **When idle (no job running):** open a transient TCP connection,
    send `?Q<n> [<var>]\r\n`, read until `\x17`, close. Same as
    `haas_bridge._round_trip()`.
  - **When streaming:** queue the request. Between line writes, drain the
    socket, send the Q-code query, read the framed response, then resume
    streaming. Bound the pause window so the Haas's look-ahead doesn't
    starve.
  - Return `{ raw, value, parsed }` — same shape as the dashboard's
    `/api/query`.
- [x] **Z-6. `POST /api/cnc/start`.** Validates the file path is under the
  share folder, kicks off the streamer, returns a `job_id`. 409 if a job is
  already running. Logs to `journalctl -u cnc-stream`.
- [x] **Z-7. `POST /api/cnc/stop`.** Stops feeding bytes. The Haas stalls
  on look-ahead drain.
- [x] **Z-8. `WS /api/cnc/stream`** — pushes `line` events as the streamer
  advances and `status` events on state change. Match the message shapes in
  the API contract above.

### Phase 3 — UI

- [x] **Z-9. "Send to Machine" button** on the `.nc` file viewer (the
  existing 3D viewer / Ace editor split-pane). POSTs to `/api/cnc/start`.
  Show a confirmation dialog first: *"Stream `part_0429_v3.nc` (2,043
  lines) to 192.168.20.200? — Confirm"*. Disabled if a job is already
  running (poll `/api/cnc/status`).
- [x] **Z-10. Two trackers in the 3D viewer.**
  - **Machine tracker** — colored marker that follows the line index in the
    `line` WebSocket events.
  - **User tracker** — already exists (the cursor that scrubs through the
    toolpath when the user clicks in the 3D view or moves the editor cursor).
  - **"Follow machine" button** — sticky-ON while a job is running. When
    ON, the user tracker is locked to the machine tracker. When the user
    clicks anywhere in the editor or 3D view, the lock breaks and the
    button reappears as **"Resume follow"** to re-lock. Approximation is
    fine — the user explicitly said: *"doesn't have to be exact correct
    atm, just a reactive of what the connection machine says."*
- [x] **Z-11. "Machine Status" entry in the left sidebar.** ~~Iframe + ?token=~~ **Pivoted (PR #24)**: `/machine` is now a native Vue page that polls `GET /api/cnc/state` (also new). Sidebar entry is in (PR #20).
- [x] **Z-12. Currently-served-file breadcrumb.** When `/api/cnc/status`
  returns `running:true`, surface the file path in:
  - The **header bar** (across the top of every filebrowser page) — pill
    showing `Running: part_0429_v3.nc · line 482 / 2043`. Clicking
    navigates to `/files/<that-path>`.
  - The **top of the left sidebar** — same info, more space.
  Updates live from the WS stream when open, falls back to a 2-second
  poll otherwise.
- [x] **Z-13. Camera embed** on the `/machine` page. If `camera_url` set:
  - URL ends in `.m3u8` → `<video>` with HLS.js (or native HLS on Safari).
  - URL ends in `.jpg` / `/snapshot` → `<img>` reloaded every 200 ms.
  - URL starts with `rtsp://` → render *"RTSP isn't browser-renderable;
    set up an `go2rtc` proxy to convert to HLS, then paste the `.m3u8` URL
    here."*
  Place above or beside the dashboard iframe.

### Phase 4 — Polish

- [ ] **Z-14. DPRNT / output capture during streaming.** When the Haas
  emits DPRNT logs back over the same RS-232, capture them and append to
  a sibling file in the share folder (`<jobname>.log`). Same pattern as
  today's USB-mode `cnc-usb-watcher`.
- [x] **Z-15. Crash recovery prompt.** If filebrowser-NC restarts mid-job,
  the next start of the streamer should refuse to auto-resume — surface a
  banner asking the operator to confirm.

---

## TODOs — Haas-dashboard side (this repo)

- [ ] **D-1. Switch the bridge to proxy mode.** New env vars:
  - `ZINC_BASE_URL` (e.g. `http://zinc.local:8080`)
  - `ZINC_BEARER_TOKEN` (issued by Zinc settings page; see `Z-1`)
  When set, `HaasBridge.query()` POSTs to
  `<ZINC_BASE_URL>/api/cnc/qcode` with `Authorization: Bearer …` instead of
  opening TCP to the Waveshare. When unset (offline / standalone testing),
  current direct-TCP behavior is preserved.
- [ ] **D-2. Add a "Current Job" tile.** Between the hero row and the
  metric tiles. Polls `<ZINC_BASE_URL>/api/cnc/status` every 2 s. Shows
  file name, line `current/total`, progress bar, and a "view in
  filebrowser" link to `<ZINC_BASE_URL>/files/<file_path>`.
- ~~D-3. CORS for iframe embed.~~ **Obsolete (Z-11 pivot)** — no iframe.
- [ ] **D-4. Token-bearer auth on the dashboard's API.** ~~`?token=…` query string for iframe~~ — drop the query-string acceptor; `Authorization: Bearer …` header is enough now. Validated against the value Zinc minted; unset → fall back to LAN-open (dev mode).
- [ ] **D-5. Hide direct-TCP details when proxied.** Footer "Host: 192.168.20.200:4196" becomes "via zinc.local" when running through the proxy. Still applies — useful operator hint.

---

## Cross-cutting / coordination

- [ ] **X-1. Auth contract.** Zinc's settings page mints a long-lived
  bearer token (Z-1). User pastes it into haas-dashboard's env
  (`ZINC_BEARER_TOKEN`). Dashboard sends it on every `/api/cnc/qcode` call.
  Token regeneration on the Zinc side invalidates the old token; dashboard
  reports a clear "Pi rejected our token, re-paste" error.
- [ ] **X-2. API versioning.** Header `X-API-Version: 1` on every
  `/api/cnc/*` request and response. Dashboard logs a warning if the
  version Zinc returns isn't 1.
- [ ] **X-3. Failure-mode UX.** Dashboard distinguishes between three
  states:
  - **Pi reachable, Haas reachable** — green dot, normal.
  - **Pi reachable, Haas unresponsive** — yellow dot,
    `Pi up, machine not responding`.
  - **Pi unreachable** — red dot, `Pi offline`.
  Implement in `static/app.js` `updateConnection()`.

---

## Out of scope for v1 (revisit when ready)

- Hardware feed-hold pause via Pi GPIO + opto-isolated relay → Haas's
  external feed-hold input on the I/O board. (User: *"we can table that
  for now."*)
- M00 injection mid-stream as a software pause.
- Multi-machine support (one Pi serving multiple Haases).
- Job queueing — submit a list of files to run sequentially.
- Probing-result capture beyond simple DPRNT logs.
- Fine-grained per-line ETA (would need per-line travel-time estimation).

---

## Quick reference for the Zinc chat

- **This repo (haas-dashboard):** `/home/ubuntu/Repos/haas-dashboard/`
- **Filebrowser-NC repo:** `/home/ubuntu/Repos/filebrowser-NC/`
- **Existing haas-dashboard endpoints:**
  `GET /api/state`, `GET /api/raw?q=…`, `POST /api/query`, `WS /ws`.
- **Haas wire protocol (already implemented in
  [`haas_bridge.py`](../haas_bridge.py)):**
  - Send: `?Q<code>[ <var>]\r\n`
  - Receive: `<echo>\r\n\x02<data>\x17\r\n>\n` (STX-framed payload, then
    next-input prompt)
  - **End-of-data marker:** `\x17` (ETB), **not** `>?` (`>?` is the
    next-input idle prompt and arrives unreliably).
- **Useful Q-codes:**

  | Q-code     | Returns                              |
  |------------|--------------------------------------|
  | Q104       | Mode (LISTPROG, MEM, MDI, JOG…)       |
  | Q201       | Current tool                          |
  | Q300       | Power-on time                         |
  | Q301       | Motion time                           |
  | Q303       | Last cycle time                       |
  | Q402       | Parts counter                         |
  | Q500       | `PROGRAM,<O#>,<status>,PARTS,<n>`     |
  | Q600 N     | Macro variable read                   |
  | • #3027    |   spindle actual RPM                  |
  | • #1815    |   commanded spindle                   |
  | • #5021–23 |   machine X / Y / Z                   |
  | • #5041–43 |   work coord X / Y / Z                |
  | • #5221–23 |   G54 offset X / Y / Z                |
  | • #4014    |   active modal G-code (which WCS)     |

- **Haas Settings expected ON:**
  - `143` — Machine Data Collect (enables Q-codes)
  - `187` — Echo (the dashboard tolerates either; not critical)
  - `14`  — Synchronization mode — **CONFIRM** with user before streaming
            tests; needs to be a DNC-compatible mode for XON/XOFF feed.
