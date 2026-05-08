#!/usr/bin/env bash
#
# setup-pi.sh — interactive installer for the CNC-USB-bridge Pi.
#
# Single-command bring-up on a fresh Pi. Walks through:
#   1. Share folder on the Pi (default ~/cnc/files)
#   2. Mode:
#        a) USB mass-storage gadget — pretend to be a thumb drive to a CNC controller
#        b) G-code streaming server — stretch, stub
#   3. Auto-installs build prereqs (Node 24, corepack, Go 1.25) — first run only
#   4. Builds the filebrowser binary (frontend + Go backend)
#   5. For mode (a): dwc2 OTG, FAT32 backing image, g_mass_storage gadget,
#                    debounced eject+reattach watcher
#   6. filebrowser systemd service, points at the share, starts on boot
#   7. Optional reboot (required first run only, for dwc2 OTG to take effect)
#
# Idempotent: re-running prefills answers from /etc/cnc-pi.conf, skips
# already-installed prereqs, and is safe to re-run with the same answers.
#
# Usage:  sudo bash pi-setup/setup-pi.sh   (auto-sudo-elevates if needed)

set -euo pipefail

REPO_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
LIB_DIR="$REPO_DIR/pi-setup/lib"

# shellcheck source=lib/common.sh
. "$LIB_DIR/common.sh"
# shellcheck source=lib/prereqs.sh
. "$LIB_DIR/prereqs.sh"
# shellcheck source=lib/usb_mass_storage.sh
. "$LIB_DIR/usb_mass_storage.sh"
# shellcheck source=lib/gcode_stream.sh
. "$LIB_DIR/gcode_stream.sh"

require_root "$@"

REBOOT_REQUIRED=0

# ── Defaults (overridden by /etc/cnc-pi.conf if present) ────────────────────
DEFAULT_USER=${SUDO_USER:-${USER}}
DEFAULT_HOME=$(getent passwd "$DEFAULT_USER" | cut -d: -f6)
SHARE_PATH="${DEFAULT_HOME}/cnc/files"
MODE="usb"            # usb | stream
IMAGE_PATH="${DEFAULT_HOME}/cnc/cnc-usb.img"
IMAGE_SIZE_MB=4096    # 4 GB
WATCH_DEBOUNCE_SECONDS=8
WATCH_MIN_INTERVAL_SECONDS=30
USB_VENDOR="filebrowser-NC"
USB_PRODUCT="CNC USB"
USB_SERIAL="$(hostname | tr -d '\n')"
FB_USER="$DEFAULT_USER"
FB_DB="${DEFAULT_HOME}/.config/filebrowser/filebrowser.db"

# Load existing config if present (overrides the defaults above).
load_conf || true

step "filebrowser-NC :: Pi USB-bridge setup"
log "This will configure filebrowser + (optionally) USB mass-storage gadget."
log "Existing config: ${CONF_PATH} ($([[ -e $CONF_PATH ]] && echo found || echo none))"

# ── Prompts ─────────────────────────────────────────────────────────────────

ask SHARE_PATH "Path on the Pi where CNC files live" "$SHARE_PATH"

ask_choice MODE_LABEL "Pick a mode" \
  "USB mass-storage (Pi acts as a thumb drive to the controller)" \
  "G-code streaming server (stretch — not implemented)"
case $MODE_LABEL in
  USB*)   MODE=usb ;;
  G-code*) MODE=stream ;;
esac

if [[ $MODE == usb ]]; then
  ask IMAGE_PATH               "Path to the FAT32 backing image" "$IMAGE_PATH"
  ask IMAGE_SIZE_MB            "Image size (MB) — only used if creating new" "$IMAGE_SIZE_MB"
  ask WATCH_DEBOUNCE_SECONDS   "Debounce — quiet seconds before re-export" "$WATCH_DEBOUNCE_SECONDS"
  ask WATCH_MIN_INTERVAL_SECONDS \
                               "Min seconds between two re-exports (no flapping)" "$WATCH_MIN_INTERVAL_SECONDS"
  ask USB_VENDOR  "USB vendor string"   "$USB_VENDOR"
  ask USB_PRODUCT "USB product string"  "$USB_PRODUCT"
  ask USB_SERIAL  "USB serial number"   "$USB_SERIAL"
fi

ask FB_USER "User filebrowser will run as" "$FB_USER"

# ── Resolve binary location ────────────────────────────────────────────────
FB_BIN="$REPO_DIR/filebrowser"
FB_WORKDIR="$REPO_DIR"

# ── Persist config ──────────────────────────────────────────────────────────
step "Saving configuration to $CONF_PATH"
write_conf \
  "SHARE_PATH=$SHARE_PATH" \
  "MODE=$MODE" \
  "IMAGE_PATH=$IMAGE_PATH" \
  "IMAGE_SIZE_MB=$IMAGE_SIZE_MB" \
  "WATCH_DEBOUNCE_SECONDS=$WATCH_DEBOUNCE_SECONDS" \
  "WATCH_MIN_INTERVAL_SECONDS=$WATCH_MIN_INTERVAL_SECONDS" \
  "USB_VENDOR=$USB_VENDOR" \
  "USB_PRODUCT=$USB_PRODUCT" \
  "USB_SERIAL=$USB_SERIAL" \
  "FB_USER=$FB_USER" \
  "FB_BIN=$FB_BIN" \
  "FB_WORKDIR=$FB_WORKDIR" \
  "FB_DB=$FB_DB"

# ── Build prereqs + filebrowser binary (slow, automated) ────────────────────
# Done after prompts so the user can walk away while the install runs, and
# before the systemd unit step so we know the binary exists when we enable it.
install_build_prereqs
build_filebrowser

# ── Mode-specific install ───────────────────────────────────────────────────
case $MODE in
  usb)    install_usb_mass_storage_mode ;;
  stream) install_gcode_stream_mode ;;
  *)      die "unknown MODE=$MODE" ;;
esac

# ── filebrowser service (last, so it sees the mounted share) ────────────────
step "Installing filebrowser systemd service"
mkdir -p "$(dirname "$FB_DB")"
chown -R "$FB_USER:$FB_USER" "$(dirname "$FB_DB")" || true
mkdir -p "$SHARE_PATH"

render_template "$REPO_DIR/pi-setup/systemd/filebrowser.service.tmpl" \
                /etc/systemd/system/filebrowser.service \
                FB_USER="$FB_USER" \
                FB_BIN="$FB_BIN" \
                FB_WORKDIR="$FB_WORKDIR" \
                SHARE_PATH="$SHARE_PATH" \
                FB_DB="$FB_DB"
ok "wrote /etc/systemd/system/filebrowser.service"

if [[ -x $FB_BIN ]]; then
  enable_now filebrowser.service
  ok "filebrowser running on http://$(hostname -I | awk '{print $1}'):8080"
else
  systemctl daemon-reload
  warn "skipping filebrowser start — binary not built (build step failed?)"
fi

# ── Done ────────────────────────────────────────────────────────────────────
step "Done"
log "Re-run this script any time to change the share folder, mode, or timings."
log ""
log "Logs:"
log "  filebrowser:        journalctl -u filebrowser -f"
log "  USB watcher:        journalctl -u cnc-usb-watcher -f"
log "  USB mass-storage:   journalctl -u cnc-usb-mass-storage -f"
log ""
if (( REBOOT_REQUIRED )); then
  warn "A reboot is required for dwc2 OTG changes to take effect."
  ask_yes_no REBOOT_NOW "Reboot now?" y
  if [[ $REBOOT_NOW == y ]]; then
    log "rebooting…"
    systemctl reboot
  else
    warn "Reboot when ready:  sudo reboot"
  fi
fi
