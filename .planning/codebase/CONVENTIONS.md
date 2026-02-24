# Coding Conventions

**Analysis Date:** 2026-02-24

## Naming Patterns

**Files:**
- Package files: lowercase with no underscores (`config.go`, `linker.go`, `discovery.go`)
- Styles and constants in separate files when substantial (`styles.go`)
- Entry point: `main.go` in `cmd/chaparral/`

**Functions:**
- Public functions: PascalCase (exported) - `LoadManifest`, `FindOrgs`, `SyncOrg`, `StatusOrg`
- Private functions: camelCase (unexported) - `scanOrgDir`, `discoverRepos`, `isRepo`, `linkSkill`, `createSymlink`, `checkLink`
- Receiver methods on structs use PascalCase: `(o *Org) SkillsPath()`, `(m Model) View()`
- Constructor functions: `NewModel`, `NewStyle` pattern

**Variables:**
- Local variables: camelCase - `orgs`, `results`, `statuses`, `repos`, `skills`, `manifest`
- Package-level variables (styles): camelCase - `titleStyle`, `orgNameStyle`, `repoStyle`
- Constants: PascalCase or UPPER_SNAKE_CASE - `maxContentWidth`, `manifestFile`
- Receiver names: single letter (standard Go) - `m` for Model, `o` for Org, `b` for *strings.Builder

**Types:**
- Struct names: PascalCase - `Manifest`, `Org`, `Skill`, `LinkResult`, `LinkStatus`, `Model`
- Interface-like types using iota: PascalCase enum members - `viewDashboard`, `tabSkills`
- Private message types for Bubble Tea: camelCase - `orgsLoaded`, `syncDone`

## Code Style

**Formatting:**
- Standard Go formatting (gofmt)
- Line breaks use consistent spacing:
  - One blank line between top-level declarations
  - Two blank lines between major sections within functions
- Indentation: one tab character (standard Go)
- Max line length: practical limit of ~100 chars, but TUI content capped at `maxContentWidth = 80`

**Linting:**
- No explicit linter config detected; assumes standard `go vet` and `golangci-lint` defaults
- Follow Go standard conventions: unexported package-level functions before exported ones
- Comment exported items according to Go conventions (required for exported symbols)

## Import Organization

**Order:**
1. Standard library imports (stdlib)
2. Third-party imports (charmbracelet, github.com packages)

**Pattern from actual code** (`cmd/chaparral/main.go`):
```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manzanita-research/chaparral/internal/config"
	"github.com/manzanita-research/chaparral/internal/discovery"
	"github.com/manzanita-research/chaparral/internal/linker"
	"github.com/manzanita-research/chaparral/internal/tui"
)
```

**Path Aliases:**
- No path aliases used; full import paths with module prefix: `github.com/manzanita-research/chaparral/internal/...`

## Error Handling

**Patterns:**
- Immediate check: `if err != nil { ... }`
- Named returns for errors - functions return `(Type, error)` or `([]Type, error)`
- Error wrapping with context: `fmt.Errorf("operation: %w", err)` to preserve error chain
- Non-critical errors are logged to stderr and loop continues (graceful degradation in discovery)
- Critical errors trigger `os.Exit()` from main

**Pattern from** `internal/config/config.go`:
```go
if err := json.Unmarshal(data, &m); err != nil {
    return Manifest{}, fmt.Errorf("parsing manifest: %w", err)
}
```

**Pattern from** `cmd/chaparral/main.go` for CLI errors:
```go
if err := tui.Run(basePath); err != nil {
    fmt.Fprintf(os.Stderr, "error: %v\n", err)
    os.Exit(1)
}
```

## Logging

**Framework:** Standard library `fmt` package (no external logger)

**Patterns:**
- Informational output: `fmt.Println()`, `fmt.Printf()`
- Errors to stderr: `fmt.Fprintf(os.Stderr, "...")`
- No logging inside library packages (`internal/`); callers decide output
- TUI handles display through Bubble Tea's View model pattern

**Error message style:**
- Plain-spoken, no "Error:" prefix (per project conventions in CLAUDE.md)
- Includes context: "no orgs found. add a chaparral.json to a brand repo to get started."

## Comments

**When to Comment:**
- Required: All exported functions and types must have doc comments
- Recommended: Complex logic or non-obvious state transitions
- Avoided: Self-documenting code (well-named functions/variables)

**Documentation pattern from** `internal/discovery/discovery.go`:
```go
// FindOrgs scans a base directory for organization directories.
// An org is any directory containing a repo with a chaparral.json.
func FindOrgs(basePath string) ([]config.Org, error) {
```

**Block comments for sections:**
```go
// Link org-level CLAUDE.md to parent directory
claudeResults := linkClaudeMD(org)
results = append(results, claudeResults...)

// Find available skills
skills, err := discovery.FindSkills(org.SkillsPath())
```

## Function Design

**Size:** Functions average 10-40 lines; longer functions (50-100 lines) like `renderDashboard()` have clear sections separated by blank lines

**Parameters:**
- Prefer concrete types (structs) over many individual parameters
- Receiver methods on Model pass receiver as first parameter implicitly
- Helper functions receive only what they need (linker functions: `org config.Org, repo string, skill config.Skill`)

**Return Values:**
- Most functions return either single value or `(Value, error)` pair
- Slices returned directly (not pointers) for iteration
- Status/lookup functions return either result or zero-value with boolean: `(config.Org, bool, error)` in `scanOrgDir()`

**Pattern from** `internal/linker/linker.go`:
```go
func SyncOrg(org config.Org) ([]LinkResult, error) {
    var results []LinkResult
    // ... operations accumulate results
    return results, nil
}
```

## Module Design

**Exports:**
- Each package exports only public API needed by consumers
- `config/`: exports `Manifest`, `Org`, `Skill` types and `LoadManifest()` function
- `discovery/`: exports `FindOrgs()` and `FindSkills()` functions
- `linker/`: exports result types (`LinkResult`, `LinkStatus`) and operations (`SyncOrg`, `UnlinkOrg`, `StatusOrg`)
- `tui/`: exports `Run()` function and `Model` type for Bubble Tea

**Internal Packages:**
- No re-export patterns; internal packages do not expose other packages' types
- Dependencies flow: `main.go` → `{config, discovery, linker, tui}` → stdlib/charmbracelet

**Package Organization:**
- Each package has single file (except `tui/` which has `tui.go` and `styles.go` for clarity)
- Styles kept separate in `styles.go` for visual customization
- Constants defined near point of use within packages

## TUI-Specific Conventions

**State Management:**
- Model struct holds all state; no global variables
- Message types (Bubble Tea) use unexported struct types: `orgsLoaded`, `syncDone`
- State transitions managed through view enum and string comparisons

**Styling:**
- All colors defined in `internal/tui/styles.go` at package level
- Styles use Manzanita brand palette: terracotta, sage, ochre, cream, rust, redwood, lavender
- Render functions return strings to View model for composition
- Width constrained to `maxContentWidth = 80` characters per project convention

---

*Convention analysis: 2026-02-24*
