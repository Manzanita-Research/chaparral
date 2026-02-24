# Technology Stack

**Analysis Date:** 2026-02-24

## Languages

**Primary:**
- Go 1.24.4 - All application code, CLI, and TUI

## Runtime

**Environment:**
- Go runtime (compiled binary distribution)

**Package Manager:**
- Go modules (`go.mod`, `go.sum`)
- Lockfile: Present

## Frameworks

**Core:**
- Bubble Tea 1.3.10 - TUI event loop and state management
- Bubbles 1.0.0 - TUI components (spinner, interactive elements)
- Lip Gloss 1.1.0 - Terminal styling and layout

**CLI:**
- Go stdlib `fmt`, `os`, `path/filepath`, `strings` - Command parsing and argument handling

**JSON:**
- Go stdlib `encoding/json` - Manifest and config file parsing

**Filesystem:**
- Go stdlib `os`, `os/exec` - File I/O, symlink operations, directory traversal

## Key Dependencies

**Critical:**
- `github.com/charmbracelet/bubbletea` 1.3.10 - TUI rendering and event handling. Core to interactive dashboard.
- `github.com/charmbracelet/lipgloss` 1.1.0 - Terminal styling for the warm, muted design language
- `github.com/charmbracelet/bubbles` 1.0.0 - Spinner and other reusable TUI components

**Terminal Utilities:**
- `github.com/charmbracelet/x/ansi` 0.11.6 - ANSI escape code handling
- `github.com/charmbracelet/x/term` 0.2.2 - Terminal capability detection
- `github.com/charmbracelet/x/cellbuf` 0.0.15 - Terminal cell buffer management
- `github.com/charmbracelet/colorprofile` 0.4.1 - Color capability detection
- `github.com/muesli/termenv` 0.16.0 - Terminal environment and color support
- `github.com/mattn/go-isatty` 0.0.20 - TTY detection
- `github.com/mattn/go-runewidth` 0.0.19 - Unicode rune width calculation
- `github.com/muesli/cancelreader` 0.2.2 - Input cancellation
- `github.com/erikgeiser/coninput` 0.0.0-20211004153227-1c3628e74d0f - Windows console input

**Text Processing:**
- `github.com/lucasb-eyer/go-colorful` 1.3.0 - Color space conversions
- `github.com/clipperhouse/displaywidth` 0.9.0 - Display width calculation
- `github.com/clipperhouse/stringish` 0.1.1 - String utilities
- `github.com/clipperhouse/uax29/v2` 2.5.0 - Unicode segmentation
- `github.com/rivo/uniseg` 0.4.7 - Unicode text segmentation
- `github.com/muesli/ansi` 0.0.0-20230316100256-276c6243b2f6 - ANSI code parsing

**System:**
- `golang.org/x/sys` 0.38.0 - OS-specific system calls
- `golang.org/x/text` 0.3.8 - Unicode and locale support

## Configuration

**Environment:**
- `NO_COLOR` env var - Respected for color output control (see `internal/tui/styles.go:hasNoColor()`)
- Default base path: `~/code/` (overridable via command-line arguments)

**Build:**
- Standard Go build: `go build -o chaparral ./cmd/chaparral`
- No custom build configuration files (Makefile, build scripts)

## Manifest Configuration

**Manifest File:** `chaparral.json` (per brand repo)
- Specifies org name, Claude.md path, skills directory, and excluded repos
- Parsed via Go stdlib `encoding/json` in `internal/config/config.go`

## Platform Requirements

**Development:**
- Go 1.24.4 or compatible
- Unix-like shell (macOS, Linux) or Windows with WSL
- Git (for discovery of `.git` directories when identifying repos)

**Production:**
- No external dependencies beyond compiled binary
- Requires filesystem access to org directories
- Respects `NO_COLOR` for accessibility

**Deployment:**
- Distributed as compiled static binary via `go install`
- Cross-platform: builds to any GOOS/GOARCH supported by Go
- No runtime dependencies or DLLs required

---

*Stack analysis: 2026-02-24*
