# Architecture

**Analysis Date:** 2026-02-24

## Pattern Overview

**Overall:** Layered CLI with pluggable TUI frontend and command-line commands.

**Key Characteristics:**
- Multi-org discovery and single-source-of-truth symlink management
- Separation between discovery (filesystem scanning), configuration (manifest parsing), operations (symlink creation/removal), and presentation (CLI vs TUI)
- Idempotent symlink operations with conflict detection and safety guards
- Stateless command execution with TUI state management via Bubble Tea

## Layers

**Discovery Layer:**
- Purpose: Scan filesystem for organizations, brand repos, and available skills
- Location: `internal/discovery/discovery.go`
- Contains: `FindOrgs()` scans base path for orgs containing `chaparral.json`; `FindSkills()` enumerates skills in a skills directory; `scanOrgDir()` and `discoverRepos()` handle directory traversal with exclusion rules
- Depends on: stdlib filesystem APIs, `internal/config` types
- Used by: Main entry point (`cmd/chaparral/main.go`), TUI initialization

**Configuration Layer:**
- Purpose: Parse and represent org manifests, skills, and paths
- Location: `internal/config/config.go`
- Contains: `Manifest` (chaparral.json structure), `Org` (org metadata and sibling repo list), `Skill` (skill directory metadata)
- Depends on: stdlib JSON unmarshaling
- Used by: Discovery layer, Linker layer, TUI state

**Linker Layer (Operations):**
- Purpose: Safely create, remove, and verify symlinks; perform org-level sync operations
- Location: `internal/linker/linker.go`
- Contains: `SyncOrg()` performs full sync of skills and CLAUDE.md; `UnlinkOrg()` removes managed symlinks; `StatusOrg()` checks current link state; helper functions for symlink creation/validation
- Depends on: `internal/config`, `internal/discovery`, stdlib filesystem APIs
- Used by: Main CLI commands, TUI sync operations

**CLI Layer (Command Dispatcher):**
- Purpose: Route commands and provide text-based output for batch operations
- Location: `cmd/chaparral/main.go`
- Contains: `main()` dispatches to `runSync()`, `runStatus()`, `runUnlink()`; helper functions for formatting output (`actionIcon()`, `stateIcon()`, `countAction()`)
- Depends on: All internal layers
- Used by: Invoked directly when subcommands are provided (sync, status, unlink, help)

**TUI Layer (Interactive Frontend):**
- Purpose: Provide interactive dashboard for multi-org management with tabbed views and async operations
- Location: `internal/tui/tui.go`, `internal/tui/styles.go`
- Contains: `Model` (Bubble Tea model with state), `Update()` (event handling and state transitions), `View()` (render current screen), `renderDashboard()` (tabbed skills/repos view), `renderSyncing()` (progress spinner), `renderResults()` (sync outcome summary), `renderHelp()` (keybindings), color definitions
- Depends on: Charmbracelet (bubbletea, lipgloss, bubbles), all internal layers
- Used by: Invoked when no subcommand is provided (default dashboard mode)

## Data Flow

**Discovery and Sync Flow:**

1. User runs `chaparral` (no args) or `chaparral sync`
2. `main.defaultBasePath()` returns `~/code` (home-relative default)
3. `discovery.FindOrgs(basePath)` scans each org directory
4. `discovery.scanOrgDir()` looks for `chaparral.json` in each repo within org
5. `config.LoadManifest()` parses manifest, extracts org name, skills path, CLAUDE.md path
6. `discovery.discoverRepos()` finds sibling repos (excluding brand repo and excluded names)
7. For each repo and skill, `linker.linkSkill()` creates or updates symlink at `repo/.claude/skills/skillname`
8. `linker.linkClaudeMD()` creates org-level symlink at `org/CLAUDE.md`
9. Results returned as `[]LinkResult` with action ("created", "exists", "skipped", "error") and optional detail message

**Status Check Flow:**

1. User runs `chaparral status`
2. `discovery.FindOrgs()` loads all orgs
3. `linker.StatusOrg()` checks each symlink using `checkLink()`
4. For each link, examines symlink target vs expected source
5. Returns `[]LinkStatus` with state ("linked", "missing", "stale", "conflict")
6. CLI formats and displays per org, grouping by skill or repo

**TUI Interaction Flow:**

1. `tui.Run()` creates Bubble Tea program with `NewModel(basePath)`
2. `Model.Init()` spawns goroutine calling `discovery.FindOrgs()` and `linker.StatusOrg()` for each org
3. Results trigger `orgsLoaded` custom message, updating model state
4. User navigates orgs with j/k, tabs between skills view and repos view with `tab`
5. On `enter` or `s` key, `Model.syncOrg()` or `Model.syncAll()` spawns goroutine calling `linker.SyncOrg()`
6. Results trigger `syncDone` message, rendering `viewDone` with summary
7. User presses `esc` to return to dashboard and refresh

**State Management:**

- TUI state lives in `Model` struct: orgs, statuses (map[orgName][]LinkStatus), results, cursor position, current view
- Async operations return custom tea.Msg types (`orgsLoaded`, `syncDone`) to trigger state updates
- No shared state between CLI commands and TUI — each invocation is stateless
- Discovery results cached in `Model.statuses` to avoid refetching until user presses `r` to refresh

## Key Abstractions

**Org Manifest (`chaparral.json`):**
- Purpose: Declare which skills to share, where CLAUDE.md lives, which repos to exclude
- Examples: `internal/config/config.go` (types), README.md (example manifest)
- Pattern: JSON file with `org`, `claude_md`, `skills_dir`, `exclude` fields; enables zero-config discovery

**Skill Directory:**
- Purpose: A directory containing a `SKILL.md` file that Claude Code recognizes
- Examples: Validated by `discovery.FindSkills()` checking for SKILL.md presence
- Pattern: Symlinked into sibling repos at `.claude/skills/skillname`

**Link State Enum:**
- Purpose: Represent symlink health without modifying files
- Examples: "linked" (correct), "missing" (absent), "stale" (wrong target), "conflict" (non-symlink exists)
- Pattern: Used in `LinkStatus` to communicate what needs fixing without changing state

**Link Result Enum:**
- Purpose: Report what action was taken during sync
- Examples: "created", "exists", "skipped", "error", "removed"
- Pattern: Returned from sync/unlink operations with optional detail message for reporting

**Org Directory Structure:**
- Purpose: Assume parent directory contains brand repo (with manifest) and sibling repos
- Examples: `~/code/manzanita-research/` with `brand/`, `toyon/`, `ceanothus/`
- Pattern: Enables discovery without global config — any dir with a chaparral.json becomes discoverable

## Entry Points

**TUI (Default):**
- Location: `cmd/chaparral/main.go` → `tui.Run(basePath)`
- Triggers: Running `chaparral` with no arguments
- Responsibilities: Interactive dashboard; load orgs asynchronously; dispatch sync commands; display status and results

**CLI Sync:**
- Location: `cmd/chaparral/main.go` → `runSync(basePath)`
- Triggers: Running `chaparral sync`
- Responsibilities: Load orgs; iterate and sync each; print per-repo results and summary

**CLI Status:**
- Location: `cmd/chaparral/main.go` → `runStatus(basePath)`
- Triggers: Running `chaparral status`
- Responsibilities: Load orgs; check all link states; group and print by skill or repo; show missing/linked counts

**CLI Unlink:**
- Location: `cmd/chaparral/main.go` → `runUnlink(basePath)`
- Triggers: Running `chaparral unlink`
- Responsibilities: Load orgs; iterate and unlink each; print removed link summary

## Error Handling

**Strategy:** Fail gracefully per org; report errors to user without stopping other operations.

**Patterns:**
- `discovery.FindOrgs()` skips dirs it can't read, returns partial results
- `discovery.scanOrgDir()` skips repos without chaparral.json, continues searching
- `linker.SyncOrg()` catches per-skill errors and returns them in results; doesn't stop on first failure
- `linker.linkSkill()` returns `LinkResult` with action "error" and detail message instead of panicking
- TUI shows error state in `Model.err`; CLI prints to stderr and exits with status 1 on fatal errors
- Symlink operations use `os.Lstat()` not `os.Stat()` to avoid following symlinks and missing stale links

## Cross-Cutting Concerns

**Logging:** No logging framework. `fmt.Fprintf(os.Stderr, ...)` for errors; `fmt.Printf(...)` for status output. TUI uses Lip Gloss styles for colored output.

**Validation:**
- Manifest parsing validates JSON structure in `config.LoadManifest()`
- Symlink validation uses `os.Readlink()` to check target matches expected source
- Repo detection checks for `.git` directory presence using `isRepo(path)`
- Skill detection requires `SKILL.md` file presence using `discovery.FindSkills()`

**Authentication:** Not applicable — purely filesystem-based operations.

**Idempotency:**
- `createSymlink()` detects existing correct symlinks and returns "exists" without modification
- `SyncOrg()` safe to run multiple times; updates stale symlinks, skips correct ones
- `UnlinkOrg()` only touches symlinks (`isOurSymlink()` checks mode); leaves real files alone
- NO_COLOR environment variable respected via `hasNoColor()` for CI/script-friendly output

**Safety Guards:**
- Never overwrites non-symlink files; returns "skipped" with detail instead
- `os.MkdirAll()` used before creating symlinks to ensure .claude/skills/ exists
- `os.Symlink()` creates or replaces only symlinks, not real files
- `os.Remove()` only called on symlinks (`isOurSymlink()` guard)

---

*Architecture analysis: 2026-02-24*
