# LibKill

Supply-chain compromise scanner and cleaner. Detects and removes compromised npm/pip packages using threat intelligence from Socket.dev, GitHub Advisory DB, OSV.dev, and security research.

```
  LibKill — Supply-Chain Compromise Scanner
  ─────────────────────────────────────────
  [1] Scan system for compromised packages
  [2] Update threat database
  [3] List known threats
  [4] Launch TUI (interactive)
  [q] Quit
```

## Features

- **2672+ compromised package artifacts** in the threat database
- Scans npm global, pip global, and Bun caches
- Cross-platform: macOS, Linux, Windows
- Interactive terminal UI with Bubble Tea (optional)
- Background daemon with desktop notifications
- Auto-updating threat feeds from multiple sources
- One-command install

## Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/firfircelik/libkill/main/install.sh | bash
```

Or manually:

```bash
git clone https://github.com/firfircelik/libkill.git
cd libkill
make install
```

## Usage

```bash
# Interactive menu (no arguments)
libkill

# Direct scan
libkill scan

# Auto-clean without confirmation
libkill scan --auto

# Interactive terminal UI
libkill tui

# Run as background daemon
libkill daemon

# Install as OS service (launchd/systemd)
libkill install

# List known threats
libkill list

# Update threat database
libkill update
```

### Scan output

```
LibKill — Supply-Chain Compromise Scanner
──────────────────────────────────────────────────
Scanning npm packages...
  867 scanned → clean
Scanning pip packages...
  8 scanned → clean
──────────────────────────────────────────────────

✓ No compromised packages found. Your system is clean.
```

When threats are found:

```
1 compromised package(s) found:

  [1] rc@1.2.8
      ecosystem: npm
      location:  bun cache: rc@1.2.8@@@1
      reason:    rc hijacked (2021)

Remove? [1-1 / all / none / quit] (default: all):
```

## Threat Coverage

| Source | Entries |
|--------|---------|
| Socket.dev (Mini Shai-Hulud v1 + v2) | 1500 |
| Contagious Interview (North Korea) | 534 |
| 2025 Compromises (Qix, DuckDB, Nx, CrowdStrike, Tinycolor) | 353 |
| GitHub Advisory DB (GHSA) | 110 |
| 2026 Attacks (dYdX, SANDWORM, StegaBin, CanisterWorm) | 104 |
| Historical research (event-stream, ua-parser-js, W4SP, typosquats) | 71 |

Known campaigns: Shai Hulud, Contagious Interview, StegaBin, SANDWORM_MODE, CanisterWorm, TeamPCP, W4SP, and dozens more.

## Supported Ecosystems

- **npm** — global packages, Bun cache
- **pip** — global packages, virtual environments

More coming: cargo, gem, nuget.

## Building from Source

```bash
# Requires Go 1.23+
go build -o libkill ./cmd/libkill

# Run tests
go test ./... -race

# Cross-compile
make build-all
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
