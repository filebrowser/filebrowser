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

- **filebrowser-NC** runs as a systemd service rooted at your share folder.
- A **FAT32 image file** is loop-mounted at that folder, so filebrowser
  writes land directly in the image — no rsync, no drift between "what
  you uploaded" and "what the controller sees", they're the same bytes.
- **`g_mass_storage`** exports that image to the controller as a USB drive,
  **read-only to the controller** (more on this below).
- **`cnc-usb-watcher`** debounces file events and re-exports the LUN
  (`echo "" > …/lun0/file` then `echo $IMAGE_PATH > …/lun0/file`), which
  the controller's USB stack handles like an unplug + replug. New
  contents show up automatically.

### Why the controller mounts the stick read-only

This is the part that bites hard if you don't get it right. The kernel
docs (`Documentation/usb/mass-storage.rst`) say:

> If the file is opened for both reading and writing and is accessed
> via the host and via the local Linux system at the same time then
> the contents of the file may be corrupted.

We need to write to the image from Linux (filebrowser uploads), and
the controller needs to read it. If we let the controller also write,
both sides cache the FAT separately and fight — directory entries get
corrupted, file contents end up at the wrong sectors, files look
garbled when you try to open them.

The fix the kernel docs describe is the one we use: pass `ro=1` to
`g_mass_storage`. Linux writes freely, the controller reads only.
Edits happen at the office workstation and travel through filebrowser;
the controller is a consumer.

If your workflow needs the controller to write back to the stick
(rare but possible — DPRNT logs, edited offsets), you'll need to
either flip `ro=1` → `ro=0` in
`/etc/systemd/system/cnc-usb-mass-storage.service` and accept the
corruption risk, or wait for a v2 that detaches the LUN, syncs via
`mtools`, and re-attaches (no shared mount, no race).

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
| `SHARE_PATH` | `~/cnc/files` | Where filebrowser is rooted |
| `IMAGE_PATH` | `~/cnc/cnc-usb.img` | The FAT32 image, loop-mounted at SHARE_PATH |
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

**Controller doesn't see fresh files.** Either the watcher isn't
re-exporting, or the controller's USB stack is too aggressive about
caching. Check `journalctl -u cnc-usb-watcher -f` — you should see
`re-export complete` lines after edits. Try shortening
`WATCH_DEBOUNCE_SECONDS` or lengthening `WATCH_MIN_INTERVAL_SECONDS`
if the controller is rejecting back-to-back re-mounts.

**`could not find LUN file under /sys`** in watcher logs. The
`g_mass_storage` module isn't loaded, usually because dwc2 isn't
available. Confirm with `lsmod | grep dwc2` and
`lsmod | grep g_mass_storage`. If dwc2 is missing, the dwc2 overlay
edit didn't take effect — check `/boot/firmware/config.txt` for
`dtoverlay=dwc2` and reboot.

**Filebrowser writes don't show up in the image.** Check that
`mountpoint -q "$SHARE_PATH"` returns true. If the image isn't mounted,
filebrowser is writing to a plain folder of the same name and the
controller will never see those bytes. `sudo mount "$SHARE_PATH"`.

**I want to wipe everything.** `sudo systemctl disable --now
filebrowser cnc-usb-watcher cnc-usb-mass-storage`, then delete the
unit files in `/etc/systemd/system/`, the loop image, and
`/etc/cnc-pi.conf`.

## Files installed

| Path | Purpose |
|---|---|
| `/etc/cnc-pi.conf` | All knobs in one place. Source of truth for re-runs. |
| `/etc/systemd/system/filebrowser.service` | Web file manager |
| `/etc/systemd/system/cnc-usb-mass-storage.service` | Loads `g_mass_storage` at boot |
| `/etc/systemd/system/cnc-usb-watcher.service` | The debounced watcher loop |
| `/usr/local/bin/cnc-usb-watcher` | The watcher script itself |
| `/etc/fstab` | Adds the loop-mount entry for the FAT32 image |
| `/boot/firmware/config.txt` | `dtoverlay=dwc2` line appended (backup written) |
| `/boot/firmware/cmdline.txt` | `modules-load=dwc2` appended (backup written) |
