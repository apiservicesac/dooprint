#!/usr/bin/env bash
# Installs Dooprint as a systemd service.
#
#   sudo ./install.sh                                  install or update
#   sudo ./install.sh --port 4547                      use another port
#   sudo ./install.sh --pair "<token>" --mode agent    also pair with Odoo (servers without a browser)
#   sudo ./install.sh --uninstall                      remove the service (keeps the configuration)
#   sudo ./install.sh --uninstall --purge              also delete the configuration
set -euo pipefail
cd "$(dirname "$0")"

PORT=4547
PAIRING=""
MODE="agent"
NAME=""
UNINSTALL=0
PURGE=0

while [[ $# -gt 0 ]]; do
    case "$1" in
        --port) PORT="$2"; shift 2 ;;
        --pair) PAIRING="$2"; shift 2 ;;
        --mode) MODE="$2"; shift 2 ;;
        --name) NAME="$2"; shift 2 ;;
        --uninstall) UNINSTALL=1; shift ;;
        --purge) PURGE=1; shift ;;
        -h|--help) sed -n '2,9p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

if [[ $EUID -ne 0 ]]; then
    echo "Run it as root: sudo $0 $*" >&2
    exit 1
fi

if [[ $UNINSTALL -eq 1 ]]; then
    systemctl disable --now dooprint 2>/dev/null || true
    rm -f /etc/systemd/system/dooprint.service /usr/local/bin/dooprint /etc/udev/rules.d/60-dooprint.rules
    systemctl daemon-reload
    if [[ $PURGE -eq 1 ]]; then
        rm -rf /var/lib/dooprint
        userdel dooprint 2>/dev/null || true
    fi
    echo "Dooprint removed."
    exit 0
fi

if [[ ! -f dooprint ]]; then
    echo "The dooprint binary must be next to this script." >&2
    exit 1
fi

# The old obox-headless used the same port.
if [[ -f /etc/systemd/system/obox-headless.service ]]; then
    echo "Removing the old obox-headless service..."
    systemctl disable --now obox-headless || true
    rm -f /etc/systemd/system/obox-headless.service /usr/local/bin/obox-headless
fi

echo "Installing libusb and curl..."
if command -v apt-get >/dev/null; then
    apt-get install -y -qq libusb-1.0-0 curl >/dev/null
elif command -v dnf >/dev/null; then
    dnf install -y -q libusb1 curl
fi

id dooprint >/dev/null 2>&1 || useradd --system --home-dir /var/lib/dooprint --shell /usr/sbin/nologin dooprint
for group in plugdev lp; do
    getent group "$group" >/dev/null || groupadd --system "$group"
done
install -d -o dooprint -g dooprint -m 0750 /var/lib/dooprint

# USB printers (class 07) readable and writable by the plugdev group.
if [[ -d /etc/udev/rules.d ]]; then
    echo 'SUBSYSTEM=="usb", ENV{ID_USB_INTERFACES}=="*:07*", MODE="0664", GROUP="plugdev"' > /etc/udev/rules.d/60-dooprint.rules
    udevadm control --reload-rules 2>/dev/null && udevadm trigger --subsystem-match=usb 2>/dev/null || true
fi

systemctl stop dooprint 2>/dev/null || true
install -m 0755 dooprint /usr/local/bin/dooprint
sed "s/-port 4547/-port $PORT/" dooprint.service > /etc/systemd/system/dooprint.service
systemctl daemon-reload
systemctl enable --now dooprint

echo -n "Waiting for Dooprint"
for _ in $(seq 1 20); do
    if curl -fs "http://127.0.0.1:$PORT/api/variable" >/dev/null 2>&1; then break; fi
    echo -n "."; sleep 1
done
echo

if [[ -n "$PAIRING" ]]; then
    echo "Pairing with Odoo..."
    body=$(printf '{"pairing":"%s","mode":"%s","name":"%s"}' "$PAIRING" "$MODE" "$NAME")
    if ! curl -fsS -X POST "http://127.0.0.1:$PORT/api/odoo/pair" -H 'Content-Type: application/json' -d "$body" >/dev/null; then
        echo "Pairing failed: check the token (Odoo creates a new one each time Connect is opened)." >&2
    fi
fi

ip=$(hostname -I 2>/dev/null | awk '{print $1}')
echo
echo "Dooprint $(/usr/local/bin/dooprint -version) is running."
echo "  Web interface: http://${ip:-<this-computer-ip>}:$PORT"
echo "  Logs:          journalctl -u dooprint -f"
[[ -z "$PAIRING" ]] && echo "  Open the web interface to pair it with Odoo."
exit 0
