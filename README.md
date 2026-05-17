# filebrowser-NC

A fork of [filebrowser/filebrowser](https://github.com/filebrowser/filebrowser) tailored for shuttling NC programs to a Haas mill (or similar) over RS-232 from a Pi sitting on the shop LAN. The upstream file browser is intact — login, file tree, editing, sharing all work the same — and on top of it there's a complete machine-integration layer at `/machine` that owns the bridge to the controller.

```
   ┌──────────┐    HTTPS/WS    ┌──────────────┐    RS-232↔TCP    ┌─────────┐
   │ Operator │ ─────────────► │  filebrowser-│ ──────────────► │  Haas   │
   │ (browser)│ ◄───────────── │   NC on Pi   │ ◄────────────── │ control │
   └──────────┘                └──────┬───────┘                 └─────────┘
                                      │
                          Waveshare TCP↔RS-232 bridge
```

## What this fork adds

### Editing & visualization
- **G-code editor + 3D toolpath viewer** — split-pane Ace editor (G-code syntax highlighting) next to a live Three.js viewer. Z-up rendering matches machinist convention so a `Z-0.1` plunge visibly descends on screen. Click in the 3D view to jump the editor cursor.
- **NC chapter parser** — CAM-emitted operation headers ("OP1: ROUGH XY", "DRILL D=0.25") render as a clickable TOC. Click a chapter to jump the editor; the "Current: …" pill tracks the live line.
- **Mfg part-number auto-linking** — vendor + part-number tokens in CAM comments (`HARVEY 50050`, `OSG-A-7-1234`, `Helical EUDP-3-30000-3`) become click-through links to a Google site-search. 25 vendors curated.

### Sending programs to the controller
- **Multi-machine config** — configure N controllers (host:port for the Waveshare, tool slot count, axes, drift tolerance), switch between them on `/machine` with a dropdown. All state is per-machine.
- **Persistent send queue** — drag-reorder, stage files in advance, survives daemon restarts. Rows auto-promote when the controller starts running their O-number (catches the SD-card-loaded case).
- **Two send methods** — MEM (Receive into program memory then operator hits Cycle Start) and DNC (drip-feed). Both are RS-232 byte streams from the Pi.
- **Auto-send pipeline (opt-in)** — preflight + Start in one click when everything is green. ⚡ button on each queue row.
- **Manual "Attach to running"** — operator marks a file as the program currently running (loaded from SD card / Ethernet drop). The dashboard follows along without sending bytes via the bridge.

### Safety + visibility
- **Pre-flight tool check** — parses the NC for T-references, compares against the latest tool-table dump, flags missing / empty pocket / diameter drift / cutter-comp-with-zero-diameter. Optional hard block on Send.
- **Swap-on-send check** — surfaces "spindle currently has T7 → program will swap to T5" before the M06.
- **Live tool table** — Q-code reads of length / wear / diameter / effective values, persisted as JSON history dumps per machine.
- **Tool-table history diff** — compare any two reads. Surfaces tools added / removed / diameter wear / length re-touches between probes.
- **Tool-life discovery probe** — operator-triggered scan over a macro range to find the per-slot life counters on the specific Haas firmware.
- **Haas alarm/setting code catalog** — curated lookup for the codes operators see (Setting 414, alarms 150-154, etc.). Lines like "Alarm 152 raised" in the activity log auto-resolve into a hover-card with the title + fix hint; a typed lookup field on the connection modal handles codes you didn't get a log line for.

### Live stream + telemetry
- **WebSocket event stream** — line counter, status, log messages, queue mutations, DPRNT lines — all push to the dashboard so nothing polls.
- **G-code follow with snap-back** — the editor scrolls to the active line as the streamer advances. Operator scrolling detaches; a ⏎ live button snaps back.
- **3D machine cursor** — a green dot in the 3D viewer tracks the controller's reported XYZ position.
- **DPRNT capture** — `DPRNT[…]` macro output is scavenged between line writes and surfaces as `DPRNT:` entries in the activity log; per-job sidecar log file (`<file>.<job-id>.dprnt.log`) lands next to the NC source.

### Notifications + integration
- **Discord push notifications** — admin sets up a bot once (token + channel ID + opt-in categories: machine info / failures / operation starts), Send/Attach/error events fan out to Discord.
- **`/api/cnc/state` + bearer token** — long-lived token authenticates S2S calls so Home Assistant or other dashboards can pull the current metrics without a filebrowser session.

### Pi setup helpers (`pi-setup/`)
- **USB-gadget mass-storage** — Pi shows up as a USB stick to the controller, so machines without networking can still pull files. Debounced eject/reattach watcher so the controller picks up new uploads without operator action.
- **go2rtc re-streamer** — turn an RTSP/RTSPS camera into a stable HLS URL the dashboard can iframe. `pi-setup/scripts/install-go2rtc` drops the binary, config, and systemd unit; URL goes straight into Settings → Machine → Camera URL.

## Quickstart

### Local dev
```sh
./setup.sh           # pick a folder to serve, builds + runs
./rebuild-filebrowser.sh   # rebuild + restart after code changes
```

Frontend dev (hot reload) lives under `frontend/` — `pnpm install && pnpm dev`.

### On the Pi
```sh
# Backend
git clone https://github.com/jasongainor/filebrowser-NC ~/filebrowser-NC
cd ~/filebrowser-NC && ./setup.sh

# USB-gadget storage (optional — for controllers without networking)
sudo bash pi-setup/scripts/install-usb-gadget

# RTSP re-streamer (optional — for non-HLS cameras)
sudo bash pi-setup/scripts/install-go2rtc
```

After it boots, open the filebrowser UI, log in, navigate to **Settings → Machine** and configure:
- **Host / Port** — the Waveshare RS-232↔TCP bridge address
- **Tool slots** — your magazine pocket count (clamp for tool-table reads)
- **Axes enabled** — XYZ always; check ABC for 4th/5th-axis controllers
- **Camera** — RTSP / HLS / MJPEG / UniFi iframe URL

Verify with **Connection check** in the modal — it dials the bridge and fires a Q104 mode query.

## Haas controller prerequisites

- **Setting 143 (Machine Data Collect)** — must be **ON**. Q-codes are no-ops otherwise.
- **Settings 11-14, 37 (RS-232 framing)** — 9600 baud / 8 data / 1 stop / no parity / XON-XOFF handshake is the default the Pi side speaks.
- **Setting 187 (Echo)** — either way works; parser tolerates both.
- Look up any unfamiliar code in the dashboard's **Connection → Activity → Look up** field, or hit `/api/cnc/codes/lookup?kind=setting&number=414`.

## Architecture pointers

| Concern | File |
| --- | --- |
| Per-machine streamer + job lock | `cnc/streamer.go` |
| Q-code wire protocol | `cnc/qcode.go` |
| Background metric aggregator | `cnc/state.go` |
| Persistent send queue | `cnc/queue.go` |
| Pre-flight + cutter-comp + swap-on-send | `cnc/preflight.go` |
| Tool-table read (two-pass) | `cnc/tooltable.go` |
| Tool-table diff | `cnc/tooltable_diff.go` |
| Tool-life probe + cluster analysis | `cnc/probe_life.go` |
| DPRNT capture | `cnc/dprnt.go` |
| NC chapter parser | `cnc/chapters.go` |
| Haas codes catalog | `cnc/haas_codes.go` |
| Discord notifier | `cnc/notify.go` |
| Per-machine registry | `cnc/registry.go` |
| `/machine` view | `frontend/src/views/Machine.vue` |
| G-code follow | `frontend/src/components/machine/GcodeFollow.vue` |
| 3D viewer | `frontend/src/components/GCode3DViewer.vue` |
| Tool table panel | `frontend/src/components/ToolTablePanel.vue` |
| Queue panel | `frontend/src/components/machine/QueuePanel.vue` |

## Upstream filebrowser

Everything below this point comes from upstream. File-browsing primitives, auth, sharing, settings shell are all unchanged.

> File Browser provides a file managing interface within a specified directory and it can be used to upload, delete, preview and edit your files. It is a **create-your-own-cloud**-kind of software where you can just install it on your server, direct it to a path and access your files through a nice web interface.

Upstream is on maintenance-only mode — see [@hacdias' note](https://hacdias.com/2026/03/11/filebrowser/). Pulls into this fork need to be evaluated for compatibility with the CNC layer.

## License

Apache License 2.0 — same as upstream. © File Browser Contributors + filebrowser-NC contributors.
