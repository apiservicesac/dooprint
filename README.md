# dooprint

Print service for Odoo. It runs on a computer next to the printers, as a background service with a
web interface, and prints what Odoo sends to USB and network receipt printers.

- Receives ePOS-Print requests and turns them into ESC/POS for any receipt printer, or passes raw
  ZPL to label printers.
- Pairs with an Odoo database running the `dooprint` module, reports its printers and prints the
  jobs queued there.
- Configured from its web interface; the command line only sets the port.

## How it connects to Odoo

| Mode | When | How |
|---|---|---|
| **Agent** | Odoo is outside the printers network (cloud, VPS) | The device picks up its jobs from Odoo. Odoo wakes it up through its bus, so tickets print right away with no open ports or tunnels. |
| **Local** | Odoo is on the same network | Odoo sends each job straight to the device. |

To pair a device:

1. In Odoo, open **dooprint › Devices** and click **Connect**. Copy the pairing token.
2. Open the device web interface at `http://<computer-ip>:4547`, go to the **Odoo** tab, paste the
   token, name the device, choose the mode and click **Connect**.

The device and its printers show up in Odoo right away.

Besides print jobs, Odoo can send the device commands: restart the service, rescan printers, or
make an HTTP request on the device network and return the answer, to reach equipment Odoo cannot
see. The Odoo modules live in [dooprint-odoo](https://github.com/apiservicesac/dooprint-odoo).

## Routes

```
POST /p/<id>/cgi-bin/epos/service.cgi   print on a network or USB printer by id
POST /cgi-bin/epos/service.cgi          print on the first USB printer found
POST /p/<id>/pstprnt                    raw ZPL for label printers
GET  /                                  web interface
GET  /api/...                           API used by the web interface
```

Network printers need no setup: their id carries the IP address.

## Install

Download the files of the [latest release](https://github.com/apiservicesac/dooprint/releases/latest).

### Windows

Run `dooprint-<version>-windows-amd64-setup.exe` as administrator. It:

- installs Dooprint in `Program Files` as the **dooprint** Windows service, started with Windows
  and restarted if it stops,
- asks for the port (4547 by default) and opens it in the firewall,
- keeps the configuration and logs in `C:\ProgramData\Dooprint`,
- opens the web interface at the end to pair the device with Odoo.

Running the installer of a newer version updates it in place and keeps the pairing. Uninstall it
from **Settings › Apps**: the service and the firewall rule are removed, the configuration stays.

### Linux

```bash
tar -xzf dooprint-<version>-linux-amd64.tar.gz
cd dooprint
sudo ./install.sh
```

It installs `/usr/local/bin/dooprint` as the **dooprint** systemd service, running as its own
`dooprint` user with the configuration in `/var/lib/dooprint`, and gives that user access to USB
receipt printers.

On a server without a browser, pair it from the command line:

```bash
sudo ./install.sh --pair "https://odoo.example.com?token=...&db_name=..." --mode agent --name "Canteen"
```

| Command | What it does |
|---|---|
| `sudo ./install.sh --port 8080` | Install or update on another port |
| `sudo ./install.sh --uninstall` | Remove the service, keep the configuration |
| `sudo ./install.sh --uninstall --purge` | Also delete the configuration and the user |
| `journalctl -u dooprint -f` | Follow the logs |

## Run without installing

```bash
./dooprint                            # web interface at http://<ip>:4547
./dooprint -port 8080
./dooprint -data-dir /srv/dooprint    # configuration folder (default: user config folder)
./dooprint -printer-id 192.168.1.80   # id and address of a network printer
./dooprint -version
```

## Development

Only Docker and `make` are needed: every target builds inside containers.

```bash
make help            # list the targets
make web             # web interface
make linux windows   # dist/dooprint and dist/dooprint.exe
make installer       # Windows installer from dist/dooprint.exe
make package-linux   # tar.gz with the binary and install.sh
make dist            # all of the above plus SHA256SUMS
make check test      # gofmt, go vet and tests
```

## Releases

Versions follow [Semantic Versioning](https://semver.org). The version lives in `VERSION` and
the notes in `CHANGELOG.md`: write the changes under **Unreleased** as you go, then:

```bash
make release-patch   # 1.0.0 -> 1.0.1
make release-minor   # 1.0.0 -> 1.1.0
make release-major   # 1.0.0 -> 2.0.0
make release VERSION=1.2.0-rc.1
make release-dry-run
```

The release script checks that `main` is clean and in sync, runs the checks, moves the
Unreleased notes to the new version, commits, tags `vX.Y.Z` and pushes. The **Release** workflow
then builds everything and publishes the GitHub release with:

| File | Contents |
|---|---|
| `dooprint-X.Y.Z-windows-amd64-setup.exe` | Windows installer |
| `dooprint-X.Y.Z-windows-amd64.exe` | Portable Windows executable |
| `dooprint-X.Y.Z-linux-amd64.tar.gz` | Linux binary with `install.sh` |
| `SHA256SUMS` | Checksums |

Versions with a suffix (`1.2.0-rc.1`) are published as pre-releases. The **CI** workflow checks
formatting, vets, tests and builds the web interface and the installer on every push.

## Project layout

| Path | Contents |
|---|---|
| `cmd/dooprint` | Entry point |
| `internal/server` | ePOS print routes |
| `internal/escpos` | ePOS-Print XML to ESC/POS |
| `internal/printer` | USB and network printers, one queue per printer |
| `internal/app` | Device logic and the Odoo link |
| `internal/agent` | Pairing, heartbeat, jobs and the Odoo bus |
| `internal/webui` | API and embedded web interface |
| `frontend` | Web interface (React, Tailwind, i18next) |
| `packaging` | Windows (Inno Setup) and Linux (systemd) installers |
| `scripts` | Release script |

## Origin

The printing core (`internal/server`, `internal/escpos`, `internal/printer`) comes from
[odoo/obox-app](https://github.com/odoo/obox-app), without its desktop application.
