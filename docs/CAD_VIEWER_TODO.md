# CAD viewer + setup-sheet bundle — research and plan

Research scratchpad for adding STEP / x_t / sldprt previewing to the
filebrowser-NC editor view, and pairing CAD + drawing + NC together as
a "setup sheet bundle" that follows a part through the shop.

This is a TODO doc. Nothing in here is implemented yet — see
`frontend/src/components/CadViewer.vue` for the current scaffolding.

## What we want

When a machinist opens a part folder in filebrowser, they should see
all the artifacts that belong to that part — the CAD model, the 2D
drawing, the NC file(s) — and be able to preview each one inline
without launching another app. Click the CAD; it renders. Click the
PDF drawing; it renders. Click the `.nc`; the existing toolpath
viewer renders. Same window, same workflow.

That's the bundle. The shop floor's "setup sheet" used to be a
printed page with toolpath screenshots, drawing thumbnails, and a
tool list. We can make it a folder + a viewer.

## Format support — what's realistic

Researched 2026-05-01.

| Format | Browser-native? | Path |
|---|---|---|
| `.step` / `.stp` | ✅ Yes | [occt-import-js](https://github.com/kovacsv/occt-import-js) (OpenCascade WASM port). LGPL, ~5–8 MB WASM, mature. **This is the chosen library.** |
| `.iges` / `.igs` | ✅ Yes | Same library — occt handles IGES too. |
| `.brep` | ✅ Yes | Same. |
| `.stl` | ✅ Yes | Three.js has `STLLoader`. Trivial. |
| `.3mf` | ✅ Yes | Three.js has `3MFLoader`. |
| `.x_t` (Parasolid) | ❌ No | Format kernel is closed; Siemens controls implementations. **No open-source browser reader exists.** Must convert backend-side. |
| `.sldprt` (SolidWorks) | ❌ No | Proprietary. eDrawings web requires their plugin. **No open-source browser reader.** Must convert backend-side. |
| `.f3d` (Fusion) | ❌ No | Proprietary. Same story. |

For the closed formats, the answer is **server-side conversion to STEP**
when the file lands in the share folder, then serve the STEP through
the same browser viewer.

## Server-side conversion — options

| Tool | Cost | Headless | Reliability |
|---|---|---|---|
| **FreeCAD** macro CLI | Free (LGPL) | Yes (with quirks) | ~85–95% on simple parts; falls over on complex assemblies. Acceptable for MVP. |
| **CAD Exchanger SDK** | Commercial, license per machine | Yes | 99%+ on all three formats. The grown-up answer. |
| **OpenCascade direct** | Free (LGPL) | Yes (C++) | STEP/IGES/BREP only. Doesn't help with x_t/sldprt. |

For a Pi-resident shop server, **FreeCAD headless** is the realistic
path. Install via apt, run a python macro on file change (hooked into
the same `inotifywait` watcher already in `pi-setup/`), drop the
converted `.step` next to the source file. Cache the conversion so
re-opens are fast.

## Plan — staged

### Stage 1 — STEP preview in browser (small)
- `frontend/src/components/CadViewer.vue` — Three.js scene, lazy-loads
  `occt-import-js` via dynamic `import()`. Don't bloat the main bundle
  for users who never open CAD.
- Wire into the file preview path for `.step` / `.stp` / `.iges` /
  `.igs` / `.brep` / `.stl` / `.3mf`. The text-based Editor view
  doesn't apply — these are binary, so it's a Preview-style mount.
- File size cap (say 50 MB raw) with a "download instead" fallback.

### Stage 2 — drawing PDF inline
- `.pdf` already gets previewed by upstream filebrowser. Confirm it
  still works after the upstream merge. No work expected.

### Stage 3 — setup-sheet bundle UI
- A "part folder" convention: a folder with a CAD file, a drawing PDF,
  and one or more NC files is treated specially.
- Folder view shows a card layout: CAD on top, drawing + tool list on
  the side, NC files in a list. Click any to expand; the viewers from
  Stage 1 + the existing NC viewer handle the rendering.
- This is mostly a `FileListing` styling pass with a couple of
  metadata tweaks. No new viewers needed.

### Stage 4 — backend conversion for closed formats
- Add an optional `cad-converter` systemd service (separate from the
  USB-bridge watcher), watches the share path for `.x_t` / `.sldprt` /
  `.f3d`, runs FreeCAD headless to emit `<name>.step` next to it.
- Cache by source-file hash so we don't re-convert unchanged files.
- Surface a "converted from .sldprt" badge in the UI so users know
  they're looking at the conversion, not the source.
- Pi 4/5 should handle this fine for typical shop part files. Pi Zero
  will struggle.

### Stage 5 — assembly support (later)
- STEP assemblies (multi-body) should explode/collapse with a
  toggle. occt-import-js returns a tree; the viewer just needs UI.

## Library choice rationale

**Why occt-import-js over `online-3d-viewer`?**
Online 3D Viewer is a full app — it bundles its own viewer UI, file
input, etc. We want to render into our own Three.js scene in our own
component. occt-import-js is the pure parser; we control the viewer.
Smaller surface, better fit.

**License caveat:** occt-import-js is LGPL because OpenCascade is.
Static linking the WASM into our bundle is fine for a self-hosted
server (LGPL allows use), but if filebrowser-NC ever ships as a
binary distribution, the WASM should be a separate download or
available on a CDN to keep the LGPL boundary clean. Not a problem
today.

## Open questions

- Does upstream filebrowser have a clean way to register a custom
  preview component for arbitrary extensions, or do we have to fork
  `Preview.vue` directly? (Stage 1 needs this.)
- For STEP files >20 MB, is the parse fast enough to keep on the main
  thread, or do we need to push occt into a Worker? Test before
  committing to a path.
- Setup-sheet bundle: file naming convention vs. metadata file. The
  cleanest version is "if a folder contains exactly one CAD + one PDF
  + one or more NC, treat as bundle"; the explicit version is a
  `bundle.yaml` next to the files. Pick after we have one real shop
  using it.

## Anti-goals

- We are not building a CAD editor. Read-only viewing.
- We are not building a CAM tool. Read-only NC, write-through editor.
- We are not solving Parasolid licensing. If the customer uses a
  closed format, they get the conversion path or they wait.
