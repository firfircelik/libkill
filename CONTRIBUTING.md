# Contributing to LibKill

Thanks for your interest in improving supply-chain security.

## Development Setup

```bash
git clone https://github.com/firfircelik/libkill.git
cd libkill

# Requirements
# - Go 1.23+
# - golangci-lint (optional, for linting)

go mod download
```

## Build, Test, Lint

```bash
make build      # Build binary
make test       # Run all tests with race detector
make lint       # Run golangci-lint
make install    # Build + copy to ~/.local/bin
make build-all  # Cross-compile for all platforms
```

## Project Structure

```
cmd/libkill/        CLI entry point, commands, menus
internal/
  config/           Configuration loading
  db/               SQLite threat database
  feed/             Threat intelligence fetchers (Socket, OSV, GitHub)
  scanner/          Package scanners (npm, pip)
  daemon/           Background service + OS service installers
  notify/           Cross-platform desktop notifications
  tui/              Bubble Tea terminal UI
seed.json           Embedded threat database (2672 entries)
```

## Commit Messages

- `add: feature description`
- `fix: bug description`
- `threat: add compromised packages from <campaign>`
- `docs: documentation change`

## Pull Requests

1. Fork the repo
2. Create a feature branch
3. Run `make test` and `make lint`
4. Submit PR with description

## Adding Threat Data

To add new compromised packages to the seed database:

```bash
# Add entries to internal/feed/seed.json
# Format:
{
  "package": "malicious-pkg",
  "version": "1.0.0",
  "ecosystem": "npm",
  "feed": "research",
  "severity": "critical",
  "reason": "Brief description of the compromise"
}

# Rebuild
go build ./cmd/libkill
```
