#!/bin/sh
# install.sh — Build, install, configure and start dns-proxy-api + dns-proxy-cli.
# Run from the root of the acme-dns-challange-proxy repository as root.
set -e

INSTALL_DIR="/usr/local/bin"
CONF_DIR="/etc/acme-dns-tools"
API_CONF="$CONF_DIR/dns-proxy-api.conf"
CLI_CONF="$CONF_DIR/dns-proxy-cli.conf"
OPENRC_INIT="/etc/init.d/dns-proxy-api"
SYSTEMD_UNIT="/etc/systemd/system/dns-proxy-api.service"
SERVICE_NAME="dns-proxy-api"

# ============================================================
# Helpers
# ============================================================

info()    { echo "[INFO]  $*"; }
ok()      { echo "[OK]    $*"; }
warn()    { echo "[WARN]  $*"; }
die()     { echo "[ERROR] $*" >&2; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "Required command not found: $1"
}

# ============================================================
# Preflight
# ============================================================

preflight() {
  echo "=========================================="
  echo " dns-proxy install"
  echo "=========================================="
  echo ""

  if [ "$(id -u)" -ne 0 ]; then
    die "This script must be run as root (or via sudo)."
  fi

  require_cmd go
  ok "go compiler: $(go version)"

  # Detect init system
  if command -v rc-service >/dev/null 2>&1; then
    INIT_SYSTEM="openrc"
  elif command -v systemctl >/dev/null 2>&1; then
    INIT_SYSTEM="systemd"
  else
    INIT_SYSTEM="none"
    warn "No supported init system detected (OpenRC / systemd). Service will not be registered."
  fi
  info "Init system : $INIT_SYSTEM"
  echo ""
}

# ============================================================
# Build
# ============================================================

build() {
  info "Building binaries..."
  make clean
  make all
  ok "dns-proxy-api built"
  ok "dns-proxy-cli built"
  echo ""
}

# ============================================================
# Install binaries
# ============================================================

install_bins() {
  info "Installing binaries to $INSTALL_DIR..."
  cp -f dns-proxy-api "$INSTALL_DIR/dns-proxy-api"
  chmod +x "$INSTALL_DIR/dns-proxy-api"
  ok "Installed: $INSTALL_DIR/dns-proxy-api"

  cp -f dns-proxy-cli "$INSTALL_DIR/dns-proxy-cli"
  chmod +x "$INSTALL_DIR/dns-proxy-cli"
  ok "Installed: $INSTALL_DIR/dns-proxy-cli"
  echo ""
}

# ============================================================
# Config files
# ============================================================

install_configs() {
  mkdir -p "$CONF_DIR"
  chmod 700 "$CONF_DIR"

  # --- dns-proxy-api.conf ---
  if [ -f "$API_CONF" ]; then
    info "Config already exists (not overwritten): $API_CONF"
  else
    info "Creating sample config: $API_CONF"
    cat > "$API_CONF" << 'EOF'
# dns-proxy-api configuration
# /etc/acme-dns-tools/dns-proxy-api.conf

# --- DNS management API (existing) ---
# Bearer token for the /set_txt endpoint (used by certbot hooks on remote hosts)
API_KEY=REPLACE_WITH_RANDOM_API_KEY

# --- cPanel credentials (used internally by dns-proxy-cli) ---
# These are only needed if dns-proxy-cli is invoked by the API process.
# They are typically configured in /etc/dns-proxy-cli.conf instead.

# --- Cert serving (pull model) ---
# Bearer token that remote hosts must present to GET /certs/{domain}/{file}
CERT_BEARER_TOKEN=REPLACE_WITH_RANDOM_CERT_BEARER_TOKEN

# Comma-separated list of hostnames allowed to pull certificates (FCrDNS check)
CERT_DNS_ALLOWLIST=REPLACE_WITH_ALLOWED_HOSTNAME

# Optional: override the base directory for certificate files
# Defaults to /etc/letsencrypt/live if omitted
# CERT_BASE_DIR=/etc/letsencrypt/live
EOF
    chmod 600 "$API_CONF"
    ok "Created: $API_CONF"
    warn "Edit $API_CONF and replace all REPLACE_WITH_* placeholders before starting the service."
  fi

  # --- dns-proxy-cli.conf ---
  if [ -f "$CLI_CONF" ]; then
    info "Config already exists (not overwritten): $CLI_CONF"
  else
    info "Creating sample config: $CLI_CONF"
    cat > "$CLI_CONF" << 'EOF'
# dns-proxy-cli configuration
# /etc/acme-dns-tools/dns-proxy-cli.conf

# cPanel base URL (include port, e.g. :2083)
cpanel_url=https://YOUR_CPANEL_HOST:2083

# cPanel username
cpanel_user=YOUR_CPANEL_USERNAME

# cPanel API token (created in cPanel → Manage API Tokens)
cpanel_apikey=YOUR_CPANEL_API_TOKEN
EOF
    chmod 600 "$CLI_CONF"
    ok "Created: $CLI_CONF"
    warn "Edit $CLI_CONF with your cPanel credentials."
  fi
  echo ""
}

# ============================================================
# Service registration
# ============================================================

install_service_openrc() {
  if [ -f "$OPENRC_INIT" ]; then
    info "OpenRC init script already exists (not overwritten): $OPENRC_INIT"
  else
    info "Creating OpenRC init script: $OPENRC_INIT"
    cat > "$OPENRC_INIT" << 'EOF'
#!/sbin/openrc-run

name="dns-proxy-api"
description="DNS Proxy API Service"
command="/usr/local/bin/dns-proxy-api"
command_background="yes"
pidfile="/var/run/dns-proxy-api.pid"
output_log="/var/log/dns-proxy-api.log"
error_log="/var/log/dns-proxy-api.log"

depend() {
    need net
    use logger
}

start_pre() {
    checkpath --directory /var/run
    checkpath --file --mode 0644 /var/log/dns-proxy-api.log
}
EOF
    chmod +x "$OPENRC_INIT"
    ok "Created: $OPENRC_INIT"
  fi

  rc-update add "$SERVICE_NAME" default 2>/dev/null && ok "Enabled $SERVICE_NAME on boot" || \
    info "$SERVICE_NAME already enabled or rc-update failed (non-fatal)"
}

install_service_systemd() {
  if [ -f "$SYSTEMD_UNIT" ]; then
    info "systemd unit already exists (not overwritten): $SYSTEMD_UNIT"
  else
    info "Creating systemd unit: $SYSTEMD_UNIT"
    cat > "$SYSTEMD_UNIT" << 'EOF'
[Unit]
Description=DNS Proxy API Service
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/dns-proxy-api
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
    ok "Created: $SYSTEMD_UNIT"
  fi

  systemctl daemon-reload
  systemctl enable "$SERVICE_NAME" && ok "Enabled $SERVICE_NAME on boot"
}

install_service() {
  case "$INIT_SYSTEM" in
    openrc)  install_service_openrc ;;
    systemd) install_service_systemd ;;
    *)       warn "Skipping service registration (no supported init system)."; return ;;
  esac
  echo ""
}

# ============================================================
# Start service
# ============================================================

start_service() {
  # If either config still contains REPLACE_WITH_ placeholders, skip start
  if grep -q 'REPLACE_WITH_' "$API_CONF" 2>/dev/null; then
    warn "Config $API_CONF still has placeholder values."
    warn "Service NOT started. Edit the config then run:"
    case "$INIT_SYSTEM" in
      openrc)  echo "  rc-service $SERVICE_NAME start" ;;
      systemd) echo "  systemctl start $SERVICE_NAME" ;;
      *)       echo "  $INSTALL_DIR/dns-proxy-api" ;;
    esac
    return
  fi

  info "Starting $SERVICE_NAME..."
  case "$INIT_SYSTEM" in
    openrc)
      rc-service "$SERVICE_NAME" start
      rc-service "$SERVICE_NAME" status
      ok "$SERVICE_NAME started (OpenRC)"
      ;;
    systemd)
      systemctl start "$SERVICE_NAME"
      systemctl status "$SERVICE_NAME" --no-pager
      ok "$SERVICE_NAME started (systemd)"
      ;;
    *)
      warn "Start the service manually: $INSTALL_DIR/dns-proxy-api"
      ;;
  esac
}

# ============================================================
# Summary
# ============================================================

summary() {
  echo ""
  echo "=========================================="
  echo " Installation complete"
  echo "=========================================="
  echo ""
  echo "  Binaries  : $INSTALL_DIR/dns-proxy-api"
  echo "              $INSTALL_DIR/dns-proxy-cli"
  echo "  API config: $API_CONF"
  echo "  CLI config: $CLI_CONF"
  if [ "$INIT_SYSTEM" = "openrc" ];  then echo "  Init      : $OPENRC_INIT"; fi
  if [ "$INIT_SYSTEM" = "systemd" ]; then echo "  Unit      : $SYSTEMD_UNIT"; fi
  echo ""
  if grep -q 'REPLACE_WITH_' "$API_CONF" 2>/dev/null || grep -q 'YOUR_' "$CLI_CONF" 2>/dev/null; then
    echo "  ⚠️  NEXT STEPS:"
    echo "  1. Edit $API_CONF  — set API_KEY, CERT_BEARER_TOKEN, CERT_DNS_ALLOWLIST"
    echo "  2. Edit $CLI_CONF  — set cpanel_url, cpanel_user, cpanel_apikey"
    echo "  3. Start the service:"
    case "$INIT_SYSTEM" in
      openrc)  echo "       rc-service $SERVICE_NAME start" ;;
      systemd) echo "       systemctl start $SERVICE_NAME" ;;
      *)       echo "       $INSTALL_DIR/dns-proxy-api" ;;
    esac
  else
    echo "  Service is running. Verify with:"
    case "$INIT_SYSTEM" in
      openrc)  echo "    rc-service $SERVICE_NAME status" ;;
      systemd) echo "    systemctl status $SERVICE_NAME" ;;
    esac
  fi
  echo ""
}

# ============================================================
# Entry point
# ============================================================

preflight
build
install_bins
install_configs
install_service
start_service
summary
