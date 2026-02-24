# CLAUDE.md

## What This Is

Chaparral is a Go CLI tool (Bubble Tea TUI) that manages shared Claude Code skills and configuration across multiple repos within an organization directory. Think of it as the connective tissue between sibling projects that share a brand identity.

## Tech Stack

- **Go** with **Bubble Tea** (charmbracelet/bubbletea) for TUI
- **Lip Gloss** (charmbracelet/lipgloss) for styling
- **Bubbles** (charmbracelet/bubbles) for TUI components
- No external config libraries — uses Go stdlib for JSON/filesystem

## Project Structure

```
cmd/chaparral/       — CLI entry point
internal/
  config/            — Manifest parsing, org config types
  linker/            — Symlink creation, validation, cleanup
  discovery/         — Repo and skill discovery, filesystem walking
  tui/               — Bubble Tea models, views, styles
```

## Key Concepts

- **Org directory**: A parent directory (e.g. `~/code/manzanita-research/`) containing multiple repos
- **Brand repo**: The repo within an org that contains the shared skills and CLAUDE.md (identified by `chaparral.json` manifest)
- **Manifest** (`chaparral.json`): Lives in the brand repo root, declares which skills to share and the org-level CLAUDE.md path
- **Sibling repos**: Other repos in the same org directory that receive symlinked skills

## Commands

- `chaparral` — Launch TUI dashboard (default)
- `chaparral sync` — Sync all orgs, link skills to sibling repos
- `chaparral status` — Show current link state across orgs
- `chaparral unlink` — Remove all managed symlinks

## Build & Run

```bash
go build -o chaparral ./cmd/chaparral
go run ./cmd/chaparral
```

## Conventions

- Keep the TUI warm and readable. Use Lip Gloss styles that feel human, not corporate.
- Error messages should be helpful and plain-spoken.
- Symlink operations must be idempotent and safe — never overwrite non-symlink files.
- All filesystem operations should handle missing directories gracefully.
