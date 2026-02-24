# Chaparral — Marketplace Bridge

## What This Is

Chaparral is a Go CLI/TUI tool that manages brand identity across multiple orgs and repos. It already handles local skill syncing via symlinks. This milestone adds the marketplace bridge: the ability to package local skills into Claude Code's native plugin marketplace format, publish them to GitHub, see what marketplace plugins are installed across repos, and install marketplace plugins into sibling repos — all from the same dashboard that already shows local skill status.

## Core Value

One place to see and manage both local skills and marketplace plugins across all your orgs — plus the bridge that turns local iteration into published distribution.

## Requirements

### Validated

- Local skill syncing via symlinks — existing
- Multi-org discovery from base directory — existing
- TUI dashboard with Skills/Repos tabs — existing
- CLI commands: sync, status, unlink — existing
- chaparral.json manifest for org/brand repo config — existing
- CLAUDE.md org-level symlink management — existing

### Active

- [ ] Generate marketplace.json from brand repo skills (Claude Code's native format)
- [ ] Package skills as versioned plugin snapshots
- [ ] Publish marketplace to GitHub (push to the brand repo itself)
- [ ] Scan sibling repos for installed marketplace plugins
- [ ] Query remote marketplace registry for available plugins
- [ ] Show marketplace plugin status alongside local skills in the TUI
- [ ] Install marketplace plugins into sibling repos (shell out to `claude plugin install`)
- [ ] Support multiple brand repos per org (each can be its own marketplace)
- [ ] Start private, option to go public later per marketplace

### Out of Scope

- Custom plugin format — we generate Claude Code's native format, don't invent our own
- npm/pip plugin sources — GitHub-hosted marketplaces only for now
- Marketplace hosting service — we publish to GitHub repos, not a custom registry
- Plugin authoring tools — Chaparral packages existing skills, doesn't help write new ones
- Auto-update management — Claude Code handles plugin updates natively

## Context

Claude Code has a mature plugin system (as of 2026):
- Plugins live in directories with `.claude-plugin/plugin.json` manifests
- Marketplaces are catalogs (`.claude-plugin/marketplace.json`) listing plugins with sources
- Users add marketplaces with `/plugin marketplace add owner/repo`
- Users install plugins with `/plugin install plugin-name@marketplace-name`
- Plugins are copied to cache at `~/.claude/plugins/cache` (snapshot model)
- Skills are namespaced: `/plugin-name:skill-name`

The brand repo itself becomes the marketplace repo. Chaparral adds the `.claude-plugin/` directory structure to it, so `/plugin marketplace add manzanita-research/chaparral` would work directly.

An org can have multiple brand repos (e.g., one for tone/voice skills, one for public toolkits), each becoming its own marketplace.

Existing codebase is Go with Bubble Tea TUI, ~1200 lines across 5 internal packages. Architecture is layered: discovery → config → linker → TUI/CLI.

## Constraints

- **Format**: Must generate Claude Code's native plugin/marketplace JSON format exactly
- **Install method**: Shell out to `claude plugin install` for installs — don't write settings directly
- **Versioning**: Snapshot model — published plugins freeze at a version, not live references
- **Tech stack**: Go, Bubble Tea, Lip Gloss (existing stack)
- **Safety**: Never overwrite non-symlink files (existing convention), never push to remote without explicit user action

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Generate native Claude Code format | Avoid reinventing the wheel — we're a generator, not a format owner | — Pending |
| Brand repo = marketplace repo | Simpler than maintaining separate repos; skills and marketplace live together | — Pending |
| Shell out to claude CLI for installs | Stay in sync with Claude Code's system; don't write settings directly | — Pending |
| Snapshot versioning for published skills | Marketplace plugins should be stable releases, not live files | — Pending |

---
*Last updated: 2026-02-24 after initialization*
