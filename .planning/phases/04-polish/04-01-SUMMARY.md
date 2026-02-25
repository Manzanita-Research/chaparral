---
phase: 04-polish
plan: 01
subsystem: tui, marketplace, publisher
tags: [bubbletea, lipgloss, plugin-status, version-bumping]

requires:
  - phase: 03-discovery-and-install
    provides: TUI repos tab, marketplace scanner, publisher with bumpVersion
provides:
  - Correct enabled/disabled icon rendering in TUI repos tab
  - Repo-level cursor navigation for targeted plugin install
  - Version-aware generate preview matching publish behavior
affects: []

tech-stack:
  added: []
  patterns:
    - "repoOrderForOrg as single source of truth for repos tab rendering and navigation"

key-files:
  created: []
  modified:
    - internal/marketplace/scanner.go
    - internal/tui/tui.go
    - internal/publisher/publisher.go
    - internal/publisher/publisher_test.go
    - cmd/chaparral/main.go

key-decisions:
  - "Export BumpVersion rather than duplicating version logic in main.go"
  - "Leave pluginDisabled render branch in TUI as dead code for future disabled state"

patterns-established:
  - "repoOrderForOrg: shared method for both rendering and navigation in repos tab"

requirements-completed: [DISC-03, DISC-04, GEN-04]

duration: 3min
completed: 2026-02-24
---

# Phase 4 Plan 1: Polish Summary

**Fixed enabled icon rendering, added repo cursor targeting for install, and wired version-aware generate preview using exported BumpVersion**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-24T00:00:00Z
- **Completed:** 2026-02-24T00:03:00Z
- **Tasks:** 3
- **Files modified:** 5

## Accomplishments
- Installed plugins now show sage icon (enabled) instead of always ochre (disabled) in TUI repos tab
- Repos tab has visible cursor with up/down navigation, and install targets the selected repo
- `chaparral generate` shows the next version that would be published, not always 0.1.0

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix enabled icon and add repo cursor to TUI** - `8ee8cee` (fix)
2. **Task 2: Make chaparral generate show correct next version** - `258808f` (fix)
3. **Task 3: Verify all fixes together** - verification only, no code changes

## Files Created/Modified
- `internal/marketplace/scanner.go` - Added Enabled: true to InstalledPlugin in ScanInstalled
- `internal/tui/tui.go` - Added repoCursor field, repoOrderForOrg method, repo-level navigation
- `internal/publisher/publisher.go` - Exported BumpVersion (was bumpVersion)
- `internal/publisher/publisher_test.go` - Updated test calls to BumpVersion
- `cmd/chaparral/main.go` - Applied BumpVersion in runGenerate for plugin.json and marketplace.json

## Decisions Made
- Exported BumpVersion rather than duplicating version-reading logic in main.go
- Left pluginDisabled render branch as dead code for potential future disabled state
- Used repoOrderForOrg as single source of truth for both rendering and cursor navigation

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All three milestone audit gaps closed
- Phase 4 complete, milestone ready for final verification

---
*Phase: 04-polish*
*Completed: 2026-02-24*
