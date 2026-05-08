# pi-setup/lib/prereqs.sh — install build toolchain on a fresh Pi.
#
# Installs Node 24 (NodeSource), pnpm via corepack, and Go 1.25 (official
# tarball — apt's Go is too old). Idempotent: skips anything already at
# the right version.
#
# shellcheck shell=bash

GO_VERSION_REQUIRED="1.25.0"
GO_DOWNLOAD_VERSION="1.25.4"   # what we actually fetch if installing
NODE_MAJOR_REQUIRED=24

# version_ge <a> <b>  → 0 if a >= b, else 1. Compares dotted versions.
version_ge() {
  printf '%s\n%s\n' "$2" "$1" | sort -C -V
}

ensure_apt_basics() {
  ensure_pkgs ca-certificates curl gnupg git build-essential
}

# ── Node 24 + corepack ──────────────────────────────────────────────────────
ensure_node() {
  step "Checking Node.js"

  local current_major=0
  if command -v node &>/dev/null; then
    current_major=$(node -v 2>/dev/null | sed -E 's/^v([0-9]+).*/\1/')
  fi

  if (( current_major >= NODE_MAJOR_REQUIRED )); then
    ok "Node $(node -v) already installed"
  else
    log "installing Node $NODE_MAJOR_REQUIRED via NodeSource…"
    # NodeSource repo for Node 24. Their script handles arch + distro.
    curl -fsSL "https://deb.nodesource.com/setup_${NODE_MAJOR_REQUIRED}.x" | bash -
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq nodejs
    ok "Node $(node -v) installed"
  fi

  # corepack ships with Node — make sure pnpm resolves from package.json's
  # packageManager pin without manual installs.
  if command -v corepack &>/dev/null; then
    corepack enable >/dev/null 2>&1 || true
    ok "corepack enabled (pnpm will auto-resolve from packageManager pin)"
  else
    warn "corepack not found in this Node install — pnpm may not resolve"
  fi
}

# ── Go 1.25 from official tarball ───────────────────────────────────────────
ensure_go() {
  step "Checking Go"

  local current=""
  # Check both /usr/local/go (our install) and PATH go (apt's)
  if [[ -x /usr/local/go/bin/go ]]; then
    current=$(/usr/local/go/bin/go version 2>/dev/null | awk '{print $3}' | sed 's/^go//')
  elif command -v go &>/dev/null; then
    current=$(go version 2>/dev/null | awk '{print $3}' | sed 's/^go//')
  fi

  if [[ -n $current ]] && version_ge "$current" "$GO_VERSION_REQUIRED"; then
    ok "Go $current already installed (>= $GO_VERSION_REQUIRED)"
    # Make sure PATH picks it up if it's our /usr/local/go install
    if [[ -x /usr/local/go/bin/go ]]; then
      export PATH="/usr/local/go/bin:$PATH"
    fi
    return 0
  fi

  log "installing Go $GO_DOWNLOAD_VERSION from go.dev/dl…"

  local arch
  case "$(dpkg --print-architecture 2>/dev/null || uname -m)" in
    arm64|aarch64)        arch=arm64 ;;
    armhf|armv7l|armv6l)  arch=armv6l ;;  # works on armv7 too
    amd64|x86_64)         arch=amd64 ;;
    *) die "unsupported architecture: $(uname -m). Install Go $GO_VERSION_REQUIRED manually and re-run." ;;
  esac

  local tarball="go${GO_DOWNLOAD_VERSION}.linux-${arch}.tar.gz"
  local url="https://go.dev/dl/${tarball}"
  local tmpdir
  tmpdir=$(mktemp -d)

  curl -fsSL "$url" -o "${tmpdir}/${tarball}" || die "could not download $url"
  rm -rf /usr/local/go
  tar -C /usr/local -xzf "${tmpdir}/${tarball}"
  rm -rf "$tmpdir"

  # Make it discoverable for this script and future shells.
  export PATH="/usr/local/go/bin:$PATH"
  if ! grep -q '/usr/local/go/bin' /etc/profile.d/go.sh 2>/dev/null; then
    # shellcheck disable=SC2016 # literal $PATH for the login shell to expand
    printf 'export PATH="/usr/local/go/bin:$PATH"\n' > /etc/profile.d/go.sh
    chmod 0644 /etc/profile.d/go.sh
  fi

  ok "Go $(/usr/local/go/bin/go version | awk '{print $3}') installed at /usr/local/go"
}

install_build_prereqs() {
  step "Installing build prerequisites (one-time)"
  ensure_apt_basics
  ensure_node
  ensure_go
  ok "all prerequisites ready"
}

# ── Build the filebrowser binary as the invoking (non-root) user ────────────
build_filebrowser() {
  step "Building filebrowser binary"

  local build_user="${SUDO_USER:-${USER}}"
  if [[ -z $build_user || $build_user == root ]]; then
    warn "running as root with no SUDO_USER — building as root (caches go in /root)"
    build_user=root
  fi

  # Make sure Node + Go are on PATH for the build user's bash.
  # shellcheck disable=SC2016 # literal $PATH for the bash -c subshell to expand
  local prelude='export PATH="/usr/local/go/bin:$PATH"'

  log "  frontend: pnpm install + build  (this may take a few minutes on a fresh Pi)…"
  if [[ $build_user == root ]]; then
    bash -c "$prelude; cd '$REPO_DIR/frontend' && corepack pnpm install --silent && corepack pnpm run build"
  else
    runuser -u "$build_user" -- bash -c "$prelude; cd '$REPO_DIR/frontend' && corepack pnpm install --silent && corepack pnpm run build"
  fi
  ok "frontend built"

  log "  backend: go build…"
  if [[ $build_user == root ]]; then
    bash -c "$prelude; cd '$REPO_DIR' && go build -o filebrowser"
  else
    runuser -u "$build_user" -- bash -c "$prelude; cd '$REPO_DIR' && go build -o filebrowser"
  fi
  ok "backend built ($REPO_DIR/filebrowser)"
}
