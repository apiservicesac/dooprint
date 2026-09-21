# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2026-09-22

### Added
- Update from the web interface. **System &rsaquo; Update** looks for the latest published release
  and, when there is a newer one, a strip on every tab installs it: the device downloads the
  release for its system, checks it against the published `SHA256SUMS`, replaces the running
  binary and restarts the service. It only looks when asked, and only downloads once the update
  is confirmed.

## [1.0.1] - 2026-09-22

### Fixed
- Local mode keeps the device online in Odoo. The agent stayed idle when the device was paired
  in local mode, so Odoo only knew it was alive while someone had its form open and nothing
  refreshed its printers. It now sends the same heartbeat as in agent mode, every 30 s, without
  listening to the bus, and picks up the restart and rescan commands that used to wait forever.

## [1.0.0] - 2026-09-16

### Added
- ePOS-Print server that turns requests into ESC/POS for USB and network receipt printers, and
  passes raw ZPL to label printers.
- Web interface in English and Spanish to manage printers and the Odoo link.
- Pairing with the `dooprint` Odoo module in agent or local mode. In agent mode the device keeps
  the Odoo bus open and prints right away, with Odoo on a remote server.
- Remote commands from Odoo: restart, rescan printers and HTTP requests on the device network.
- Windows service support and the `-data-dir` and `-version` flags.
- Windows installer (service, firewall rule, port page) and Linux installer (systemd service,
  USB permissions, optional pairing from the command line).
- Makefile and release script; GitHub workflows that build, test and publish tagged releases.

[Unreleased]: https://github.com/apiservicesac/dooprint/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/apiservicesac/dooprint/compare/v1.0.1...v1.1.0
[1.0.1]: https://github.com/apiservicesac/dooprint/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/apiservicesac/dooprint/releases/tag/v1.0.0
