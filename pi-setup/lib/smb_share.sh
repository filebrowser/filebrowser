# Samba share so Finder / Explorer see the Pi as a network location.
#
# Why this exists:
#   The shop wants the same files visible three ways — through filebrowser
#   on the LAN, as a USB stick to the controller, and as a regular network
#   drive in Finder/Explorer. SMB is the network-drive piece. Avahi is what
#   makes the Pi auto-appear under "Network" in Finder.
#
# Design:
#   - We OWN /etc/samba/smb.conf — single [global] + single [cnc] share.
#     Distro default (which auto-shares [homes]/[printers]) gets backed up
#     to .cnc-pi.bak. With exactly one share visible, Finder skips the
#     share-picker and drops you straight into the folder.
#   - $SMB_GUEST=y → no auth (guest writable). On a shop LAN this is the
#     "open the box and go" path. $SMB_GUEST=n → $FB_USER + $ADMIN_PASSWORD.
#   - Files written via SMB are force-owned by $FB_USER:$FB_USER so
#     filebrowser and the watcher can read them either way.
#   - Avahi service file advertises _smb._tcp + _device-info._tcp so the
#     Pi shows up with a server icon under "Network" in Finder's sidebar.
#
# shellcheck shell=bash

install_smb_share() {
  step "Installing SMB share (Finder / Explorer network location)"
  ensure_pkgs samba avahi-daemon

  # Make sure the share folder exists and is owned by FB_USER before
  # samba tries to serve it.
  mkdir -p "$SHARE_PATH"
  chown -R "$FB_USER:$FB_USER" "$SHARE_PATH" 2>/dev/null || true

  # Always set the SMB password from ADMIN_PASSWORD — even in guest mode
  # we want auth-mode to be one Enter-key away if the user toggles it later.
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

  # Take full ownership of smb.conf. Backup once so the distro default is
  # recoverable if someone ever wants the stock Debian setup back.
  local conf=/etc/samba/smb.conf
  if [[ -f $conf && ! -f ${conf}.cnc-pi.bak ]]; then
    cp -a "$conf" "${conf}.cnc-pi.bak"
    log "backed up distro smb.conf → ${conf}.cnc-pi.bak"
  fi

  local guest=${SMB_GUEST:-n}
  local global_extras share_auth
  if [[ $guest == y ]]; then
    global_extras='   map to guest = bad user'
    share_auth=$'   guest ok = yes\n   guest only = yes\n   force user = '"$FB_USER"$'\n   force group = '"$FB_USER"
  else
    global_extras='   map to guest = never'
    share_auth=$'   guest ok = no\n   valid users = '"$FB_USER"$'\n   force user = '"$FB_USER"$'\n   force group = '"$FB_USER"
  fi

  cat > "$conf" <<EOF
# /etc/samba/smb.conf — written by setup-pi.sh
# Single-purpose: expose the CNC share folder as [cnc].
# Original distro default backed up at ${conf}.cnc-pi.bak.
# Re-run setup-pi.sh to change settings.

[global]
   workgroup = WORKGROUP
   server string = %h CNC
   security = user
$global_extras
   load printers = no
   disable spoolss = yes
   printcap name = /dev/null
   log file = /var/log/samba/log.%m
   max log size = 1000
   logging = file

[cnc]
   comment = CNC files
   path = $SHARE_PATH
   browseable = yes
   read only = no
$share_auth
   create mask = 0664
   directory mask = 0775
   veto files = /._*/.DS_Store/.AppleDouble/.Trashes/
   delete veto files = yes
EOF

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

  if [[ $guest == y ]]; then
    ok "samba share '\\\\$(hostname)\\cnc' configured (no auth — guest writable)"
  else
    ok "samba share '\\\\$(hostname)\\cnc' configured (user: $FB_USER)"
  fi
}
