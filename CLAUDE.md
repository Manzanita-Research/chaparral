# CLAUDE.md

## What This Is

Chaparral is a Go CLI tool (Bubble Tea TUI) for managing brand identity across multiple organizations and repos. It's the connective tissue between sibling projects that share a design language, voice, and set of Claude Code skills.

Two modes, one tool:

- **Local skills** — symlink shared skills from a brand repo into every sibling project. Edit once, propagated instantly. The fast path for active development.
- **Marketplace awareness** — see which Claude Code plugin marketplaces and plugins are installed across your repos. The bridge between local iteration and published distribution.

This is a brand design tool first. It exists because creative orgs working across multiple repos need shared identity to stay alive in every project, not just the one you're currently editing.

## Multi-Org by Default

Chaparral assumes you work across multiple orgs or clients. Discovery scans a base directory (default `~/code/`) and finds every org that has a `chaparral.json` manifest. You might manage three orgs with completely different brand identities from one dashboard.

## Tech Stack

- **Go** with **Bubble Tea** (charmbracelet/bubbletea) for TUI
- **Lip Gloss** (charmbracelet/lipgloss) for styling
- **Bubbles** (charmbracelet/bubbles) for TUI components (spinner, etc.)
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
- **Local skills**: Symlinked from brand repo, instant propagation, for active development
- **Marketplace plugins**: Installed via Claude Code's plugin system, cached copies, for stable distribution

## Commands

- `chaparral` — Launch TUI dashboard (default)
- `chaparral sync` — Sync all orgs, link skills to sibling repos
- `chaparral status` — Show current link state across orgs
- `chaparral unlink` — Remove all managed symlinks

## TUI Views

The dashboard has two tabs (toggle with `tab`):

- **Skills view**: lists shared skills across an org, with linked/total counts
- **Repos view**: lists each sibling repo, with its skill statuses underneath

Both views show CLAUDE.md link status and support syncing selected or all orgs.

## Build & Run

```bash
go build -o chaparral ./cmd/chaparral
go run ./cmd/chaparral
```

## Conventions

- Keep the TUI warm and readable. Use Lip Gloss styles that feel human, not corporate.
- Error messages should be helpful and plain-spoken. No "Error:" prefix.
- Symlink operations must be idempotent and safe — never overwrite non-symlink files.
- All filesystem operations should handle missing directories gracefully.
- Respect `NO_COLOR` environment variable.
- Cap content width at 80 characters, use Lip Gloss padding instead of manual string prefixes.

## Relationship to Claude Code Plugins

Claude Code has a plugin marketplace system for distributing skills. Chaparral is complementary:

- **Plugins** solve distribution — getting skills to other people, machines, orgs. Plugins copy to a cache and require per-repo configuration.
- **Chaparral** solves local development sync — keeping your own sibling repos in lockstep while you iterate. Symlinks are instant, no ceremony, no per-repo setup.
- **The bridge**: Chaparral can generate a `marketplace.json` from a brand repo, turning local skills into publishable plugins when they're stable. The TUI can show both local symlink status and installed marketplace plugin status in one view.
