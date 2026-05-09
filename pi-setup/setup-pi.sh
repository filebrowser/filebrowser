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
# shellcheck source=lib/smb_share.sh
. "$LIB_DIR/smb_share.sh"

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
ADMIN_USER="admin"
ADMIN_PASSWORD="cncadmin1234"   # 12+ chars to satisfy upstream's minimum
ENABLE_SMB="y"                  # serve $SHARE_PATH as SMB so Finder/Explorer can mount it
SMB_GUEST="y"                   # no-auth (guest writable) — fine on a shop LAN

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

# Filebrowser admin password — set deterministically so the user knows what
# it is from the moment setup ends. Must be at least 12 chars (upstream
# minimum). Default is fine for shop-LAN use; user changes it from the UI
# once they're in.
while :; do
  ask ADMIN_PASSWORD "Filebrowser admin password (min 12 chars)" "$ADMIN_PASSWORD"
  (( ${#ADMIN_PASSWORD} >= 12 )) && break
  warn "must be at least 12 characters"
done

# SMB share — Pi shows up under "Network" in Finder / Explorer with the
# share folder mountable as a regular drive. Reuses the filebrowser admin
# password for the SMB user, so you don't need to remember two.
ask_yes_no ENABLE_SMB "Expose share over SMB (Finder / Explorer network drive)?" "${ENABLE_SMB:-y}"
if [[ $ENABLE_SMB == y ]]; then
  # Guest mode = no password prompt when mounting. Right answer for a
  # shop-LAN appliance; wrong answer if the box is exposed to a network
  # you don't trust.
  ask_yes_no SMB_GUEST "Allow SMB guest access (no password)?" "${SMB_GUEST:-y}"
fi

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
  "FB_DB=$FB_DB" \
  "ADMIN_USER=$ADMIN_USER" \
  "ADMIN_PASSWORD=$ADMIN_PASSWORD" \
  "ENABLE_SMB=$ENABLE_SMB" \
  "SMB_GUEST=$SMB_GUEST"
# Conf has the admin password — restrict to root.
chmod 0600 "$CONF_PATH" 2>/dev/null || true

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

# Pre-create the admin user with the chosen password so the box is usable
# the moment the service starts. Idempotent on re-runs (updates the password
# if the user already exists) — no journal-grepping for a random string.
ensure_admin_user() {
  if [[ ! -x $FB_BIN ]]; then
    return 0
  fi
  step "Configuring filebrowser admin user"
  # Make sure the DB schema exists (config init is a no-op if it already does).
  runuser -u "$FB_USER" -- "$FB_BIN" config init --database "$FB_DB" >/dev/null 2>&1 || true
  # Try update first (covers re-runs where admin already exists).
  if runuser -u "$FB_USER" -- "$FB_BIN" users update "$ADMIN_USER" \
       --password "$ADMIN_PASSWORD" --database "$FB_DB" >/dev/null 2>&1; then
    ok "updated admin password"
  elif runuser -u "$FB_USER" -- "$FB_BIN" users add "$ADMIN_USER" "$ADMIN_PASSWORD" \
       --perm.admin --database "$FB_DB" >/dev/null 2>&1; then
    ok "created admin user"
  else
    warn "could not create or update admin user — first start will auto-init with a random password (look in journalctl -u filebrowser)"
  fi
}
ensure_admin_user

if [[ -x $FB_BIN ]]; then
  enable_now filebrowser.service
  ok "filebrowser running"
else
  systemctl daemon-reload
  warn "skipping filebrowser start — binary not built (build step failed?)"
fi

# ── SMB share (after filebrowser, so $SHARE_PATH is owned + populated) ──────
if [[ ${ENABLE_SMB:-y} == y ]]; then
  install_smb_share
fi

# ── Done ────────────────────────────────────────────────────────────────────
step "Done"
log ""
PI_IP=$(hostname -I | awk '{print $1}')
PI_HOST=$(hostname)
log "Filebrowser:"
log "  URL:      http://$PI_IP:8080"
log "  Username: $ADMIN_USER"
log "  Password: $ADMIN_PASSWORD"
log ""
if [[ ${ENABLE_SMB:-y} == y ]]; then
  log "SMB share (Mac Finder / Windows Explorer):"
  log "  Finder:   smb://$PI_HOST.local/cnc   (or smb://$PI_IP/cnc)"
  log "  Explorer: \\\\$PI_HOST\\cnc            (or \\\\$PI_IP\\cnc)"
  if [[ ${SMB_GUEST:-y} == y ]]; then
    log "  Auth:     guest (no password)"
  else
    log "  Username: $FB_USER"
    log "  Password: $ADMIN_PASSWORD   (same as filebrowser)"
  fi
  log ""
fi
log "Change the password from the user menu once you're logged in (or"
log "re-run this script to set a new one)."
log ""
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
