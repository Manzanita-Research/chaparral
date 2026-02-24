# Codebase Structure

**Analysis Date:** 2026-02-24

## Directory Layout

```
chaparral/
├── cmd/
│   └── chaparral/
│       └── main.go              # CLI entry point, command dispatcher
├── internal/
│   ├── config/
│   │   └── config.go            # Org, Manifest, Skill types; path helpers
│   ├── discovery/
│   │   └── discovery.go         # Org/skill filesystem scanning
│   ├── linker/
│   │   └── linker.go            # Symlink creation/removal/status
│   └── tui/
│       ├── tui.go               # Bubble Tea model, views, event handling
│       └── styles.go            # Lip Gloss styles, colors, symbols
├── .planning/
│   └── codebase/                # Architecture documentation
├── go.mod                        # Go module manifest
├── go.sum                        # Go dependency checksums
├── CLAUDE.md                     # Org-wide Claude instructions
├── PITCH.md                      # Product positioning
├── README.md                     # User documentation
├── .gitignore                    # Git exclusions
└── chaparral                     # Compiled binary (built artifact)
```

## Directory Purposes

**`cmd/chaparral/`:**
- Purpose: CLI binary entry point
- Contains: Main function, command dispatcher, output formatters
- Key files: `cmd/chaparral/main.go` (150+ lines)

**`internal/config/`:**
- Purpose: Type definitions and config parsing for org manifests
- Contains: `Manifest` struct (chaparral.json schema), `Org` struct (discovered org metadata), `Skill` struct (skill directory metadata), path helper methods
- Key files: `internal/config/config.go` (~67 lines)

**`internal/discovery/`:**
- Purpose: Filesystem scanning to find orgs, repos, and skills
- Contains: `FindOrgs()` (top-level discovery), `FindSkills()` (enumerate skills in a directory), `scanOrgDir()` (org-level scanning), `discoverRepos()` (sibling repo enumeration), repo detection via `.git`
- Key files: `internal/discovery/discovery.go` (~129 lines)

**`internal/linker/`:**
- Purpose: Symlink operations: creation, removal, status checking, safety validation
- Contains: `SyncOrg()` (full org sync), `UnlinkOrg()` (remove all managed links), `StatusOrg()` (check state), `LinkResult` and `LinkStatus` types, per-skill linking, conflict detection
- Key files: `internal/linker/linker.go` (~202 lines)

**`internal/tui/`:**
- Purpose: Interactive TUI dashboard built with Bubble Tea
- Contains: Bubble Tea `Model`, event handling, four view states (dashboard, syncing, done, help), tabbed views (skills and repos), styling and color definitions
- Key files: `internal/tui/tui.go` (~551 lines), `internal/tui/styles.go` (~62 lines)

**`.planning/codebase/`:**
- Purpose: Architecture and codebase documentation consumed by GSD commands
- Contains: ARCHITECTURE.md, STRUCTURE.md, and future CONVENTIONS.md, TESTING.md, CONCERNS.md

## Key File Locations

**Entry Points:**
- `cmd/chaparral/main.go`: Binary entry point; routes to TUI (no args) or CLI commands (sync, status, unlink, help)

**Configuration:**
- `go.mod`: Go module dependencies (Charmbracelet, stdlib)
- `CLAUDE.md`: Org-wide Claude Code instructions and brand voice
- `chaparral.json`: (User-created in brand repo) Org manifest with skills path and excluded repos

**Core Logic:**
- `internal/discovery/discovery.go`: Org and skill discovery from filesystem
- `internal/config/config.go`: Manifest parsing and path resolution
- `internal/linker/linker.go`: Symlink state management and operations
- `internal/tui/tui.go`: Interactive dashboard and state machine

**Styling and Display:**
- `internal/tui/styles.go`: Color palette (terracotta, sage, ochre, cream, rust) and Lip Gloss style definitions

**Testing:**
- No test files detected in the codebase

## Naming Conventions

**Files:**
- Go source files: lowercase with underscores only in package names (e.g., `main.go`, `discovery.go`)
- No test files found; if tests are added, use `*_test.go` convention

**Directories:**
- Package directories: lowercase, no underscores (e.g., `cmd`, `internal`, `config`, `discovery`)
- Binary packages always use `cmd/` subdirectory

**Functions:**
- Exported (public): PascalCase (e.g., `FindOrgs`, `SyncOrg`, `LoadManifest`, `StatusOrg`)
- Unexported (private): camelCase (e.g., `scanOrgDir`, `discoverRepos`, `linkSkill`, `createSymlink`, `checkLink`)
- Main entry function: lowercase `main()`

**Types:**
- Struct types: PascalCase (e.g., `Manifest`, `Org`, `Skill`, `Model`, `LinkResult`, `LinkStatus`)
- Constants: camelCase for local constants (e.g., `maxContentWidth`, `manifestFile`); PascalCase for exported (none in this codebase)

**Variables:**
- Local vars: camelCase (e.g., `basePath`, `orgs`, `skills`, `skillMap`, `linkedCount`)
- Function parameters: camelCase (e.g., `org`, `repo`, `skill`, `source`, `dest`)
- Unexported receiver variables in methods: shorthand camelCase (e.g., `m` for Model, `o` for Org, `b` for strings.Builder)

**Interfaces:**
- No interfaces defined explicitly; methods added to types as needed (duck typing)

## Where to Add New Code

**New Feature (e.g., filter by org type, exclude pattern matching):**
- Primary code: `internal/discovery/discovery.go` (extend scanning logic) or `internal/linker/linker.go` (extend sync operations)
- Entry point: `cmd/chaparral/main.go` (add CLI flag or subcommand)
- Display: `internal/tui/tui.go` (add view state or tab if interactive)

**New Component/Module (e.g., config validation, dry-run mode):**
- Implementation: Create new file in relevant package (e.g., `internal/validator/validator.go`)
- Import: Update importing files to call new functions
- Tests: Create `internal/validator/validator_test.go` (see Testing.md when available)

**New TUI View (e.g., detail view for single org):**
- Implementation: Add view constant in `internal/tui/tui.go`; add case in `Update()` and `View()` switch statements; create new `render*()` method
- Styling: Add colors/styles to `internal/tui/styles.go` if needed

**New CLI Command (e.g., init, validate):**
- Implementation: Add case in `cmd/chaparral/main.go` switch statement
- Handler: Create `runNewCommand(basePath string)` function in same file
- Help: Update `printHelp()` output

**Utilities and Helpers:**
- Shared filesystem helpers: `internal/discovery/discovery.go`
- Shared symlink helpers: `internal/linker/linker.go`
- Output formatting: `cmd/chaparral/main.go` (CLI) or `internal/tui/tui.go` (TUI)

## Special Directories

**`cmd/`:**
- Purpose: Executable packages
- Generated: No
- Committed: Yes
- Note: Go requires binaries to be in `cmd/packagename/` directory

**`internal/`:**
- Purpose: Private packages not importable by external Go modules
- Generated: No
- Committed: Yes
- Note: Enforces API boundary — everything in `internal/` is off-limits to outside consumers

**.planning/codebase/:**
- Purpose: GSD codebase documentation
- Generated: Yes (by Claude agent via `/gsd:map-codebase`)
- Committed: Yes (checked into git)

**`.claude/`:**
- Purpose: Claude Code project-level skills and settings
- Generated: Partially (user-created .claude/skills/ in sibling repos by chaparral)
- Committed: Symlinks to .claude/skills/ are committed; actual skill directories are in brand repo

## File Organization Patterns

**Layout for Discovery:**
```go
// discovery.go: public functions first
func FindOrgs(basePath string) ([]config.Org, error)
func FindSkills(skillsDir string) ([]config.Skill, error)

// unexported helpers follow
func scanOrgDir(orgPath string) (config.Org, bool, error)
func discoverRepos(...) []string
func isRepo(path string) bool
```

**Layout for Operations:**
```go
// linker.go: high-level operations first
func SyncOrg(org config.Org) ([]LinkResult, error)
func UnlinkOrg(org config.Org) ([]LinkResult, error)
func StatusOrg(org config.Org) ([]LinkStatus, error)

// result/status types
type LinkResult struct { ... }
type LinkStatus struct { ... }

// unexported symlink helpers
func linkClaudeMD(org config.Org) []LinkResult
func linkSkill(org config.Org, repo string, skill config.Skill) LinkResult
func createSymlink(source, dest, repo, name string) LinkResult
func checkLink(linkPath, expectedTarget, repo, name string) LinkStatus
func isOurSymlink(path string) bool
```

**Layout for TUI:**
```go
// tui.go: model definition and interface
type Model struct { ... }
func NewModel(basePath string) Model
func (m Model) Init() tea.Cmd
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m Model) View() string

// custom message types
type orgsLoaded struct { ... }
type syncDone struct { ... }

// view rendering methods
func (m Model) renderDashboard() string
func (m Model) renderSyncing() string
func (m Model) renderResults() string
func (m Model) renderHelp() string

// helper methods
func (m Model) syncAll() tea.Cmd
func (m Model) syncOrg(index int) tea.Cmd

// styles.go: colors, styles, symbols
var colorTerracotta = ...
var titleStyle = ...
var statusLinked = ...
```

## Build and Dependency Management

**Build:**
- `go build -o chaparral ./cmd/chaparral` produces binary
- No build configuration files (no Makefile, no build scripts)

**Dependencies:**
- Charmbracelet packages: bubbletea, bubbles, lipgloss (TUI framework and styling)
- All others are transitive dependencies of Charmbracelet
- No external dependencies for core logic (config parsing, discovery, symlink operations use stdlib)

**Module:**
- Go 1.24.4 minimum (specified in go.mod)
- Module path: `github.com/manzanita-research/chaparral`

---

*Structure analysis: 2026-02-24*
