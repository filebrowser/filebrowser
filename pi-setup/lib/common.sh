# Shared helpers for setup-pi.sh. Sourced, never executed.
# shellcheck shell=bash

CONF_PATH="/etc/cnc-pi.conf"

c_red()   { printf '\033[31m%s\033[0m' "$*"; }
c_green() { printf '\033[32m%s\033[0m' "$*"; }
c_yellow(){ printf '\033[33m%s\033[0m' "$*"; }
c_bold()  { printf '\033[1m%s\033[0m' "$*"; }

log()  { printf '  %s\n' "$*"; }
ok()   { printf '  %s %s\n' "$(c_green '✓')" "$*"; }
warn() { printf '  %s %s\n' "$(c_yellow '!')" "$*" >&2; }
die()  { printf '  %s %s\n' "$(c_red '✗')" "$*" >&2; exit 1; }

step() { printf '\n%s\n' "$(c_bold "==> $*")"; }

# require_root — re-exec under sudo if not already root.
require_root() {
  if [[ $EUID -ne 0 ]]; then
    log "Re-running with sudo…"
    exec sudo --preserve-env=HOME -E -- "$0" "$@"
  fi
}

# ask <var-name> <prompt> [default]
# Reads a line; if empty, uses default. Echoes the value via the named variable.
ask() {
  local __varname=$1 __prompt=$2 __default=${3:-} __reply
  if [[ -n $__default ]]; then
    read -r -p "  $__prompt [$__default]: " __reply || true
    __reply=${__reply:-$__default}
  else
    read -r -p "  $__prompt: " __reply || true
  fi
  printf -v "$__varname" '%s' "$__reply"
}

# ask_yes_no <var-name> <prompt> <default-y-or-n>
ask_yes_no() {
  local __varname=$1 __prompt=$2 __default=$3 __reply __hint
  case $__default in y|Y) __hint='Y/n' ;; *) __hint='y/N' ;; esac
  while :; do
    read -r -p "  $__prompt [$__hint]: " __reply || true
    __reply=${__reply:-$__default}
    case $__reply in
      y|Y|yes|YES) printf -v "$__varname" 'y'; return 0 ;;
      n|N|no|NO)   printf -v "$__varname" 'n'; return 0 ;;
      *) printf '  (please answer y or n)\n' ;;
    esac
  done
}

# ask_choice <var-name> <prompt> <option1> <option2> ...
# Numbered menu. Sets var to the selected option (1-indexed value).
ask_choice() {
  local __varname=$1 __prompt=$2; shift 2
  local __options=("$@") __reply __i
  printf '  %s\n' "$__prompt"
  for __i in "${!__options[@]}"; do
    printf '    %d) %s\n' "$((__i+1))" "${__options[$__i]}"
  done
  while :; do
    read -r -p "  choice [1]: " __reply || true
    __reply=${__reply:-1}
    if [[ $__reply =~ ^[0-9]+$ ]] && (( __reply >= 1 && __reply <= ${#__options[@]} )); then
      printf -v "$__varname" '%s' "${__options[$((__reply-1))]}"
      return 0
    fi
    printf '  (enter 1..%d)\n' "${#__options[@]}"
  done
}

# load_conf — sources $CONF_PATH if it exists, defining vars per the file.
load_conf() {
  if [[ -r $CONF_PATH ]]; then
    # shellcheck disable=SC1090
    . "$CONF_PATH"
    return 0
  fi
  return 1
}

# write_conf — atomically write /etc/cnc-pi.conf. Pass key=value pairs.
# Values get shell-safe quoting (printf %q) so spaces, quotes, $, etc. round-trip
# correctly through `. /etc/cnc-pi.conf` in setup-pi.sh and the watcher.
write_conf() {
  local tmp
  tmp=$(mktemp)
  {
    printf '# /etc/cnc-pi.conf — written by setup-pi.sh\n'
    printf '# Re-run setup-pi.sh to change these values.\n'
    local kv k v
    for kv in "$@"; do
      k=${kv%%=*}
      v=${kv#*=}
      printf '%s=%q\n' "$k" "$v"
    done
  } > "$tmp"
  install -m 0644 "$tmp" "$CONF_PATH"
  rm -f "$tmp"
  ok "wrote $CONF_PATH"
}

# render_template <template-path> <output-path> [VAR=VALUE …]
# Replaces @@VAR@@ placeholders. No quoting magic — values must not contain @@.
render_template() {
  local tmpl=$1 out=$2; shift 2
  local content
  content=$(< "$tmpl")
  for kv in "$@"; do
    local k=${kv%%=*} v=${kv#*=}
    content=${content//"@@${k}@@"/$v}
  done
  printf '%s' "$content" > "$out"
}

# ensure_pkgs <pkg> <pkg> …
ensure_pkgs() {
  local missing=()
  for p in "$@"; do
    dpkg -s "$p" &>/dev/null || missing+=("$p")
  done
  if (( ${#missing[@]} )); then
    log "installing: ${missing[*]}"
    DEBIAN_FRONTEND=noninteractive apt-get update -qq
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends "${missing[@]}"
  fi
}

# enable_now <unit>
enable_now() {
  local unit=$1
  systemctl daemon-reload
  systemctl enable --now "$unit"
}
