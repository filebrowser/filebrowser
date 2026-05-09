# Samba share so Finder / Explorer see the Pi as a network location.
#
# Why this exists:
#   The shop wants the same files visible three ways — through filebrowser
#   on the LAN, as a USB stick to the controller, and as a regular network
#   drive in Finder/Explorer. SMB is the network-drive piece. Avahi is what
#   makes the Pi auto-appear under "Network" in Finder.
#
# Design:
#   - Reuse $FB_USER as the SMB user (it already owns $SHARE_PATH).
#   - SMB password = $ADMIN_PASSWORD (one credential to remember).
#   - Share definition lives in /etc/samba/smb.conf inside a marked block
#     we own. Re-runs replace the block in place.
#   - Avahi service file advertises _smb._tcp + _device-info._tcp so the
#     Pi shows up with a server icon in Finder's sidebar.
#
# shellcheck shell=bash

install_smb_share() {
  step "Installing SMB share (Finder / Explorer network location)"
  ensure_pkgs samba avahi-daemon

  # Make sure the share folder exists and is owned by FB_USER before
  # samba tries to serve it.
  mkdir -p "$SHARE_PATH"
  chown -R "$FB_USER:$FB_USER" "$SHARE_PATH" 2>/dev/null || true

  # SMB user: reuse FB_USER. Samba keeps its own password DB (smbpasswd),
  # so we have to set the SMB password explicitly even though FB_USER is
  # already a Linux user. Drive both add and update from $ADMIN_PASSWORD
  # so re-runs converge.
  if pdbedit -L 2>/dev/null | grep -q "^${FB_USER}:"; then
    log "updating samba password for $FB_USER"
    printf '%s\n%s\n' "$ADMIN_PASSWORD" "$ADMIN_PASSWORD" \
      | smbpasswd -s "$FB_USER" >/dev/null
  else
    log "creating samba user $FB_USER"
    printf '%s\n%s\n' "$ADMIN_PASSWORD" "$ADMIN_PASSWORD" \
      | smbpasswd -s -a "$FB_USER" >/dev/null
  fi
  smbpasswd -e "$FB_USER" >/dev/null 2>&1 || true

  # Render the share block. Markers so re-runs replace it in place.
  local conf=/etc/samba/smb.conf
  local begin='# BEGIN cnc-share (managed by setup-pi.sh)'
  local end='# END cnc-share'
  local block
  block=$(cat <<EOF
$begin
[cnc]
   comment = CNC files
   path = $SHARE_PATH
   browseable = yes
   read only = no
   guest ok = no
   valid users = $FB_USER
   create mask = 0664
   directory mask = 0775
   force user = $FB_USER
   force group = $FB_USER
   veto files = /._*/.DS_Store/.AppleDouble/.Trashes/
   delete veto files = yes
$end
EOF
)

  if [[ ! -f $conf ]]; then
    # Fresh samba install always lays one down, but be defensive.
    printf '[global]\n   workgroup = WORKGROUP\n   server string = %%h CNC\n   security = user\n   map to guest = never\n' > "$conf"
  fi

  if grep -qF "$begin" "$conf"; then
    local tmp
    tmp=$(mktemp)
    awk -v b="$begin" -v e="$end" -v block="$block" '
      BEGIN { skip = 0 }
      $0 == b { print block; skip = 1; next }
      skip && $0 == e { skip = 0; next }
      skip { next }
      { print }
    ' "$conf" > "$tmp"
    install -m 0644 "$tmp" "$conf"
    rm -f "$tmp"
  else
    printf '\n%s\n' "$block" >> "$conf"
  fi

  # Validate before we restart — testparm exits non-zero on syntax errors.
  if ! testparm -s "$conf" >/dev/null 2>&1; then
    warn "samba config did not pass testparm; SMB may not start"
  fi

  # Avahi service for Finder discovery. Registers _smb._tcp so the Pi
  # appears under "Network", and _device-info._tcp with model=RackMac so
  # Finder draws a server icon next to it.
  mkdir -p /etc/avahi/services
  cat > /etc/avahi/services/cnc-smb.service <<'EOF'
<?xml version="1.0" standalone='no'?>
<!DOCTYPE service-group SYSTEM "avahi.dtd">
<service-group>
  <name replace-wildcards="yes">%h CNC</name>
  <service>
    <type>_smb._tcp</type>
    <port>445</port>
  </service>
  <service>
    <type>_device-info._tcp</type>
    <port>0</port>
    <txt-record>model=RackMac</txt-record>
  </service>
</service-group>
EOF

  systemctl enable smbd avahi-daemon >/dev/null 2>&1 || true
  systemctl enable nmbd >/dev/null 2>&1 || true
  systemctl restart smbd avahi-daemon
  systemctl restart nmbd 2>/dev/null || true

  ok "samba share '\\\\$(hostname)\\cnc' configured (user: $FB_USER)"
}
