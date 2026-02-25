---
status: passed
phase: 04-polish
verified: 2026-02-24
---

# Phase 4: Polish — Verification

## Goal
Fix display bugs and UX issues found during milestone audit

## Requirements Verified

### DISC-03: Installed plugins display with correct sage enabled icon
**Status: PASS**
- `internal/marketplace/scanner.go:78` sets `Enabled: true` on all InstalledPlugin structs in ScanInstalled
- This flows through `MergeStatus` in `marketplace.go` to `PluginStatus.Enabled`
- TUI's `ps.Installed && ps.Enabled` branch (tui.go:618) renders `pluginInstalled` (sage icon)
- The `pluginDisabled` (ochre) branch only fires if Enabled is false, which no longer happens for installed plugins

### DISC-04: TUI install targets the repo the user is browsing via repoCursor
**Status: PASS**
- `repoCursor int` field added to Model struct at tui.go:46
- Up/down navigation in repos tab moves repoCursor (tui.go:216-232)
- Visible `>` cursor indicator rendered at tui.go:590-596
- `startInstallPick` uses `repoOrder[m.repoCursor]` at tui.go:342 (not `org.Repos[0]`)
- repoCursor resets to 0 on tab switch and org cursor movement

### GEN-04: Generate preview shows the version that would actually be published
**Status: PASS**
- `BumpVersion` exported from publisher.go:39
- `runGenerate` in main.go:300 applies `publisher.BumpVersion` to each plugin.json preview
- Marketplace preview at main.go:324 also applies BumpVersion per plugin
- For unpublished skills, BumpVersion returns "0.1.0" (initial version); for published skills, it reads the existing version and bumps patch

## Success Criteria

| # | Criterion | Status |
|---|-----------|--------|
| 1 | Installed plugins show correct enabled/disabled icon in TUI (not always disabled) | PASS |
| 2 | TUI install targets the repo the user is browsing, not always the first repo | PASS |
| 3 | `chaparral generate` shows the next version that would be published, not always 0.1.0 | PASS |

## Build Verification

- `go build ./...` — passes
- `go test ./... -count=1` — all tests pass (generator, marketplace, publisher, skillmeta, validator)
- `go vet ./...` — clean

## Human Verification Recommended

The following items benefit from visual/interactive testing but are not blocking:
- Launch `go run ./cmd/chaparral` and confirm sage icons appear for installed plugins
- Navigate repos tab with j/k and confirm `>` cursor moves between repos
- Press `i` on a non-first repo and confirm the install picker shows the correct repo name

## Score

**3/3 must-haves verified. Phase goal achieved.**
