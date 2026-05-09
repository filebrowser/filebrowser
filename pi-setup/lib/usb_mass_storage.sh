# Mode A: Pi presents itself as a USB mass-storage device to the connected CNC controller.
#
# Architecture (v2 — bidirectional, corruption-safe):
#
#   - $SHARE_PATH is a regular Linux folder. Filebrowser writes to it directly.
#     It is NOT a loop mount of the image (was, in v1, and that caused FAT
#     corruption when both Linux and the controller wrote at the same time).
#   - $IMAGE_PATH is the FAT32 image exported via g_mass_storage. Linux NEVER
#     mounts it. It's only ever read or written through mtools, and only
#     while the LUN is detached.
#   - The watcher orchestrates the sync: detach LUN → pull controller-new
#     files into the share → atomically rebuild image from share → reattach
#     LUN. See pi-setup/scripts/cnc-usb-watcher for the full algorithm.
#
# This is the documented-safe pattern from
# Documentation/usb/mass-storage.rst — the host and the controller never
# touch the file at the same time, so the FAT cache fight that corrupts
# directory entries can't happen.
#
# shellcheck shell=bash

# enable_dwc2 — make sure dwc2 OTG controller is enabled on this Pi.
# Edits /boot/firmware/config.txt and /boot/firmware/cmdline.txt (Bookworm path).
# Falls back to /boot/* on older Raspberry Pi OS.
enable_dwc2() {
  step "Configuring dwc2 USB OTG"

  local cfg="/boot/firmware/config.txt" cmd="/boot/firmware/cmdline.txt"
  [[ -f $cfg ]] || cfg="/boot/config.txt"
  [[ -f $cmd ]] || cmd="/boot/cmdline.txt"
  [[ -f $cfg ]] || die "could not find config.txt — is this a Raspberry Pi?"
  [[ -f $cmd ]] || die "could not find cmdline.txt — is this a Raspberry Pi?"

  # Gadget mode requires dwc2 in peripheral mode in a section that
  # actually applies on this hardware. The Bookworm imager seeds a
  # [cm5] block with `dtoverlay=dwc2,dr_mode=host` — fine on a Compute
  # Module 5, completely inert on a Pi 4 Model B. Earlier versions of
  # this script edited that line in place; the result was an "already
  # configured" log line and a Pi where dwc2 never bound. We always
  # ensure a peripheral-mode line lives under [all] regardless.
  local want='dtoverlay=dwc2,dr_mode=peripheral'
  local want_block
  want_block=$'\n# Added by setup-pi.sh — USB OTG for mass-storage gadget. Lives under\n# [all] so it applies to Pi 4 Model B (the [cm5] block from the imager\n# does not).\n[all]\n'"$want"$'\n'

  # Already correct under an applicable section? grep checks that the
  # peripheral line is preceded somewhere by an [all] / [pi4] header
  # rather than buried in [cm4]/[cm5]/[pi5] etc.
  if awk '
      /^[[:space:]]*\[/ { section = $0; next }
      /^[[:space:]]*dtoverlay=dwc2,dr_mode=peripheral[[:space:]]*$/ {
        if (section ~ /^\[(all|pi4|pi04)\]/) { found = 1 }
      }
      END { exit found ? 0 : 1 }
    ' "$cfg"; then
    ok "dwc2 overlay already in peripheral mode under [all]/[pi4] in $cfg"
  else
    [[ -f ${cfg}.cnc-pi.bak ]] || cp -a "$cfg" "${cfg}.cnc-pi.bak"
    # Strip every existing dtoverlay=dwc2 line — they're either wrong
    # (host mode) or in a section that doesn't apply. Then append a
    # fresh peripheral-mode line under [all].
    sed -i -E '/^\s*dtoverlay=dwc2(\b|,).*/d' "$cfg"
    printf '%s' "$want_block" >> "$cfg"
    ok "wrote $want under [all] in $cfg (backup: ${cfg}.cnc-pi.bak)"
    # shellcheck disable=SC2034
    REBOOT_REQUIRED=1
  fi

  if ! grep -q 'modules-load=.*dwc2' "$cmd"; then
    cp -a "$cmd" "${cmd}.cnc-pi.bak"
    # cmdline.txt is single-line; append modules-load=dwc2.
    sed -i 's/$/ modules-load=dwc2/' "$cmd"
    ok "added modules-load=dwc2 to $cmd (backup: ${cmd}.cnc-pi.bak)"
    # shellcheck disable=SC2034
    REBOOT_REQUIRED=1
  else
    ok "dwc2 module-load already set in $cmd"
  fi
}

# create_backing_image <path> <size_mb>
# Creates a sparse FAT32 image. Never re-formats an existing one — that
# would clobber whatever's currently on the stick.
create_backing_image() {
  local image=$1 size_mb=$2
  step "Creating backing image at $image (${size_mb} MB, FAT32)"

  if [[ -f $image ]]; then
    log "image already exists; keeping it (delete the file and re-run setup if you want a fresh one)"
    return 0
  fi

  mkdir -p "$(dirname "$image")"
  # Sparse via dd seek — fast even for large sizes.
  dd if=/dev/zero of="$image" bs=1M count=0 seek="$size_mb" status=none
  # FAT32 with a friendly volume label. -I needed for raw .img files
  # (mkfs.vfat otherwise refuses because the file isn't a partition device).
  mkfs.vfat -F 32 -n CNC -I "$image" >/dev/null
  ok "created $image"
}

# migrate_v1_loop_mount — clean up the v1 architecture if found.
# v1 loop-mounted $IMAGE_PATH at $SHARE_PATH and added an fstab line. v2
# wants $SHARE_PATH to be a regular folder and no fstab entry. Handles
# three cases:
#   1. v1 install in good state (mounted): umount, copy contents, drop fstab.
#   2. v1 install half-broken (fstab line present but never mounted —
#      e.g. unescaped space in path): just drop the fstab line.
#   3. Fresh install: nothing to do.
migrate_v1_loop_mount() {
  local mounted=0 had_fstab=0
  if mountpoint -q "$SHARE_PATH" 2>/dev/null; then
    mounted=1
  fi
  # Match any fstab line that references the image file. Robust to
  # paths with spaces (escaped or not) and quoted paths.
  if [[ -r /etc/fstab ]] && grep -qF "$IMAGE_PATH" /etc/fstab; then
    had_fstab=1
  fi

  if (( mounted == 0 && had_fstab == 0 )); then
    return 0
  fi

  step "Migrating from v1 loop-mounted layout"

  systemctl stop filebrowser cnc-usb-watcher cnc-usb-mass-storage 2>/dev/null || true

  if (( mounted )); then
    local stash
    stash=$(mktemp -d /tmp/cnc-pi-migrate-XXXXXX)
    log "preserving share contents at $stash"
    cp -a "$SHARE_PATH"/. "$stash/" 2>/dev/null || true

    if ! umount "$SHARE_PATH"; then
      rm -rf "$stash"
      die "could not umount $SHARE_PATH — close any process holding it (lsof) and re-run"
    fi
    ok "umounted $SHARE_PATH"

    cp -a "$stash"/. "$SHARE_PATH"/ 2>/dev/null || true
    rm -rf "$stash"
    ok "share contents preserved into the regular folder at $SHARE_PATH"
  fi

  if (( had_fstab )); then
    cp -a /etc/fstab /etc/fstab.cnc-pi.bak 2>/dev/null || true
    grep -vF "$IMAGE_PATH" /etc/fstab > /etc/fstab.tmp && mv /etc/fstab.tmp /etc/fstab
    ok "removed legacy fstab entries referencing $IMAGE_PATH (backup: /etc/fstab.cnc-pi.bak)"
    systemctl daemon-reload 2>/dev/null || true
  fi
}

install_usb_mass_storage_mode() {
  # mtools = the host-side FAT32 toolkit the watcher uses to read/write
  # the image without mounting it. dosfstools = mkfs.vfat. inotify-tools
  # = the watcher's change detector.
  ensure_pkgs dosfstools inotify-tools mtools

  migrate_v1_loop_mount

  enable_dwc2

  create_backing_image "$IMAGE_PATH" "$IMAGE_SIZE_MB"
  mkdir -p "$SHARE_PATH"
  if [[ -n "${FB_USER:-}" ]]; then
    chown -R "$FB_USER:$FB_USER" "$SHARE_PATH" 2>/dev/null || true
  fi

  step "Installing USB-gadget systemd units"

  render_template "$REPO_DIR/pi-setup/systemd/cnc-usb-mass-storage.service.tmpl" \
                  /etc/systemd/system/cnc-usb-mass-storage.service \
                  IMAGE_PATH="$IMAGE_PATH" \
                  USB_VENDOR="$USB_VENDOR" \
                  USB_PRODUCT="$USB_PRODUCT" \
                  USB_SERIAL="$USB_SERIAL"
  ok "wrote /etc/systemd/system/cnc-usb-mass-storage.service"

  install -m 0755 "$REPO_DIR/pi-setup/scripts/cnc-usb-watcher" /usr/local/bin/cnc-usb-watcher
  ok "installed /usr/local/bin/cnc-usb-watcher"

  render_template "$REPO_DIR/pi-setup/systemd/cnc-usb-watcher.service.tmpl" \
                  /etc/systemd/system/cnc-usb-watcher.service \
                  WATCHER_BIN=/usr/local/bin/cnc-usb-watcher
  ok "wrote /etc/systemd/system/cnc-usb-watcher.service"

  enable_now cnc-usb-mass-storage.service
  enable_now cnc-usb-watcher.service
  ok "USB mass-storage gadget + watcher enabled"
}
