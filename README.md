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

## Build

Only Docker is needed:

```bash
./build.sh      # dist/dooprint (Linux) and dist/dooprint.exe (Windows)
```

Both binaries include USB support and the web interface.

On GitHub, the **CI** workflow vets and tests every push and pull request, and the **Release**
workflow builds `dooprint-linux-amd64` and `dooprint-windows-amd64.exe` and attaches them to a
release when a `v*` tag is pushed:

```bash
git tag v1.0.0 && git push origin v1.0.0
```

## Run

```bash
./dist/dooprint                 # web interface at http://<ip>:4547
./dist/dooprint -port 8080
./dist/dooprint -printer-id 192.168.1.80   # id and address of a network printer
```

Settings (paired Odoo, network printers, device id) are stored in the user configuration
directory, under `dooprint/config.json`.

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

## Origin

The printing core (`internal/server`, `internal/escpos`, `internal/printer`) comes from
[odoo/obox-app](https://github.com/odoo/obox-app), without its desktop application.
