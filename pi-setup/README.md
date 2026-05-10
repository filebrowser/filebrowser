# pi-setup — CNC USB bridge installer

A Pi that pretends to be a USB stick to your CNC controller, with
filebrowser-NC on the LAN as the upload UI.

**The pain this solves.** When the operator changes a file, the controller
won't see it without a manual unmount + remount on the panel — and on
some controllers, even that doesn't refresh the directory cache. This
installer wires up a debounced eject + reattach, so any file write
through filebrowser shows up on the machine's screen a few seconds later
with no panel input.

## How it fits together

- **filebrowser-NC** runs as a systemd service rooted at the share folder.
  The share folder is a regular Linux directory — **not** a mount of
  anything.
- **A FAT32 image file** lives next to the share. It's the file
  `g_mass_storage` exports to the controller as a USB drive. Linux
  **never** loop-mounts it. The host only reads/writes the image
  through `mtools`, and only while the LUN is detached.
- **`cnc-usb-watcher`** orchestrates the bidirectional sync. On every
  file change in the share folder, after `WATCH_DEBOUNCE_SECONDS` of
  quiescence and at least `WATCH_MIN_INTERVAL_SECONDS` since the last
  cycle:
  1. Detach the LUN — controller's USB stack sees the stick unplug.
  2. Pull any controller-side new files into the share folder
     (DPRNT logs, output files the machine wrote).
  3. Atomically rebuild the image from the share (build to a temp
     `.new`, swap into place).
  4. Reattach the LUN — controller re-mounts and sees the new contents.

### Why this design

The kernel's `Documentation/usb/mass-storage.rst` is explicit:

> If the file is opened for both reading and writing and is accessed
> via the host and via the local Linux system at the same time then
> the contents of the file may be corrupted.

An earlier version of this fork loop-mounted the image AND exported
it as read-write USB. Both sides cached the FAT independently and
fought, corrupting directory entries and crossing file contents. We
had files come back garbled.

The current design sidesteps the race entirely: only one side ever
touches the image at a time. While the LUN is attached, the
controller is the only writer (host doesn't touch the image at all).
While the LUN is detached, the host syncs through `mtools` (the
controller literally can't see the device). No shared cache, no
fight.

### Conflict policy

| Situation | Outcome |
|---|---|
| File on both sides | Filebrowser wins — the rebuild from the share overwrites whatever the controller wrote |
| File only on the controller, not in last snapshot | Pulled into the share (controller-created) |
| File only on the controller, *was* in last snapshot | Dropped — the host deleted it, deletion sticks |
| File only on the host | Pushed into image via the rebuild |

The watcher keeps a snapshot of the image's contents at the last
successful sync (`/var/lib/cnc-usb-watcher/last_sync_listing`) so it
can tell "controller created a new file" apart from "host deleted a
file that was previously on the stick".

## First run

Fresh Pi, Bookworm or later, OTG-capable hardware (Zero / Zero 2 W / 4 / 5):

```bash
git clone https://github.com/jasongainor/filebrowser-NC.git
cd filebrowser-NC
./rebuild-filebrowser.sh        # builds the binary
sudo bash pi-setup/setup-pi.sh  # interactive installer
sudo reboot                     # first run only — enables dwc2 OTG
```

After reboot, plug the Pi into the controller's USB-OTG port (USB-C on
Zero 2 / Pi 4 / Pi 5; inner micro-USB on Zero W). Filebrowser is on
`http://<pi-ip>:8080`.

## Re-running

Re-run `setup-pi.sh` any time. It reads previous answers from
`/etc/cnc-pi.conf` and pre-fills them — just hit Enter to keep, or type
a new value to change. Safe to re-run with the same answers.

To change one knob without re-prompting through everything, edit
`/etc/cnc-pi.conf` directly and `sudo systemctl restart cnc-usb-watcher`.

## Modes

| Mode | What it does | Status |
|---|---|---|
| **USB mass-storage** | Pi looks like a thumb drive to the CNC controller. | ✅ implemented |
| **G-code streaming** | Pi acts as a sender to a simpler router (cncjs etc). | 🚧 stretch — stub only |

## Defaults

| Setting | Default | Notes |
|---|---|---|
| `SHARE_PATH` | `~/cnc/files` | Regular folder filebrowser serves. Avoid spaces in the path. |
| `IMAGE_PATH` | `~/cnc/cnc-usb.img` | FAT32 image exported as USB. Never mounted on the host. |
| `IMAGE_SIZE_MB` | `4096` | 4 GB. Only used when creating a new image |
| `WATCH_DEBOUNCE_SECONDS` | `8` | Quiet seconds before re-export |
| `WATCH_MIN_INTERVAL_SECONDS` | `30` | Min gap between two re-exports |

## Logs

```bash
journalctl -u filebrowser -f          # filebrowser web app
journalctl -u cnc-usb-watcher -f      # debounced watcher activity
journalctl -u cnc-usb-mass-storage -f # gadget module load/unload
```

## Troubleshooting

**Controller doesn't see fresh files.** Check `journalctl -u
cnc-usb-watcher -f` — you should see `sync complete` followed by
`LUN reattached` after edits. If you don't, the watcher isn't being
triggered. If you do, the controller's USB stack may be caching too
aggressively; some Haas controllers need the operator to actually go
to the directory listing screen for the re-mount to register.

**`could not find LUN file under /sys`** in watcher logs. The
`g_mass_storage` module isn't loaded, usually because dwc2 isn't
available. Confirm with `lsmod | grep dwc2` and
`lsmod | grep g_mass_storage`. If dwc2 is missing, the dwc2 overlay
edit didn't take effect — check `/boot/firmware/config.txt` for
`dtoverlay=dwc2` and reboot.

**Files in filebrowser don't show up on the controller.** Verify
the share folder is *not* a mount point: `mountpoint -q $SHARE_PATH`
should return non-zero (it's just a folder now). If it IS a mount
point, you're on a stale v1 install — re-run `setup-pi.sh`, the
migration step will umount and clean up the fstab line.

**Path with spaces breaks setup (legacy).** v1 wrote an unescaped
fstab entry which would parse-fail on paths like
`/home/admin/Desktop/cnc files`. v2 doesn't write any fstab entry, so
spaces in the share path are now fine — but if you upgraded from a
half-broken v1 install, the migration step in `setup-pi.sh` removes
the bad fstab line for you.

**I want to wipe everything.** `sudo systemctl disable --now
filebrowser cnc-usb-watcher cnc-usb-mass-storage`, then delete the
unit files in `/etc/systemd/system/`, the image file, the share
folder, `/etc/cnc-pi.conf`, and `/var/lib/cnc-usb-watcher/`.

## Optional: go2rtc — persistent UniFi Protect / RTSP camera embed

UniFi Protect's "Share Live View" links default to 24-hour expiry on
UniFi OS 3.x and earlier; pasting one into Settings → Machine →
Camera URL works for a day, then breaks. Browsers also can't play
raw RTSP/RTSPS. The fix is a re-streamer on the Pi that translates
the camera's never-expiring RTSPS feed into HLS — which the camera
tile renders natively.

```bash
sudo bash ~/filebrowser-NC/pi-setup/scripts/install-go2rtc
# Paste the camera's RTSP/RTSPS URL when prompted.
```

Then in Settings → Machine, set Camera URL to:

```
http://<pi-host>:1984/api/stream.m3u8?src=mill-cam
```

Camera type = HLS (or Auto — the .m3u8 suffix routes correctly).

The script is idempotent — re-run after editing `/etc/go2rtc/go2rtc.yaml`
to refresh the systemd unit. To add more cameras, edit the YAML
directly and `sudo systemctl restart go2rtc`.

go2rtc's own web UI lives at `http://<pi-host>:1984` for quick
diagnostics.

## Files installed

| Path | Purpose |
|---|---|
| `/etc/cnc-pi.conf` | All knobs in one place. Source of truth for re-runs. |
| `/etc/systemd/system/filebrowser.service` | Web file manager |
| `/etc/systemd/system/cnc-usb-mass-storage.service` | Loads `g_mass_storage` at boot |
| `/etc/systemd/system/cnc-usb-watcher.service` | The bidirectional sync watcher |
| `/usr/local/bin/cnc-usb-watcher` | The watcher script itself |
| `/var/lib/cnc-usb-watcher/last_sync_listing` | Snapshot of the image's contents at the last successful sync. Lets the watcher distinguish controller-created files from host-deleted ones. |
| `/boot/firmware/config.txt` | `dtoverlay=dwc2` line appended (backup written) |
| `/boot/firmware/cmdline.txt` | `modules-load=dwc2` appended (backup written) |
