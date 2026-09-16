# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/apiservicesac/dooprint/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/apiservicesac/dooprint/releases/tag/v1.0.0
