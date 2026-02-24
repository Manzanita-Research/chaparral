# Architecture Research

**Domain:** Go CLI/TUI — marketplace bridge layer added to existing layered architecture
**Researched:** 2026-02-24
**Confidence:** HIGH (existing codebase fully analyzed; Claude Code plugin format described explicitly in PROJECT.md)

## Standard Architecture

### System Overview

The existing architecture is a clean four-layer stack. The marketplace bridge adds two new internal packages that sit between the existing layers and a thin operations layer that calls the external `claude` CLI.

```
┌─────────────────────────────────────────────────────────────────┐
│                     Presentation Layer                           │
│  ┌────────────────────────┐  ┌───────────────────────────────┐  │
│  │   TUI (Bubble Tea)     │  │   CLI (cmd/chaparral/main.go) │  │
│  │  tui.go / styles.go    │  │   sync, status, unlink,       │  │
│  │  + marketplace tab     │  │   + generate, publish, install │  │
│  └────────────┬───────────┘  └──────────────┬────────────────┘  │
├───────────────┴──────────────────────────────┴──────────────────┤
│                     Operations Layer                             │
│  ┌──────────────────┐  ┌──────────────┐  ┌────────────────────┐ │
│  │ internal/linker  │  │internal/     │  │ internal/          │ │
│  │ (existing)       │  │marketplace   │  │ publisher          │ │
│  │ symlink ops      │  │ generate,    │  │ git push, github   │ │
│  │                  │  │ scan plugins │  │ API calls          │ │
│  └────────┬─────────┘  └──────┬───────┘  └─────────┬──────────┘ │
├───────────┴────────────────────┴──────────────────────┴──────────┤
│                      Domain Layer                                 │
│  ┌──────────────────────────┐  ┌─────────────────────────────┐   │
│  │  internal/config         │  │  internal/discovery          │   │
│  │  (existing + extended)   │  │  (existing + plugin scan)    │   │
│  │  Manifest, Org, Skill    │  │  FindOrgs, FindSkills,       │   │
│  │  + PluginManifest,       │  │  + ScanInstalledPlugins      │   │
│  │  MarketplaceCatalog      │  │                              │   │
│  └──────────────────────────┘  └─────────────────────────────┘   │
├───────────────────────────────────────────────────────────────────┤
│                    External Boundaries                            │
│  ┌──────────────┐  ┌───────────────────┐  ┌────────────────────┐ │
│  │  Filesystem  │  │ github.com API    │  │  claude CLI        │ │
│  │  (existing)  │  │ (remote registry) │  │  plugin install    │ │
│  └──────────────┘  └───────────────────┘  └────────────────────┘ │
└───────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Status |
|-----------|----------------|--------|
| `internal/config` | Types for manifests, orgs, skills; parse chaparral.json | Existing — extend with plugin types |
| `internal/discovery` | Filesystem scanning for orgs, repos, skills | Existing — extend with plugin scanning |
| `internal/linker` | Symlink create/remove/status | Existing — no changes needed |
| `internal/marketplace` | Generate plugin.json, marketplace.json; version packaging | New package |
| `internal/publisher` | Git operations, GitHub API calls, remote registry queries | New package |
| `internal/tui` | Interactive dashboard | Existing — add marketplace tab and install flow |
| `cmd/chaparral` | Command dispatcher | Existing — add generate, publish, install commands |

## Recommended Project Structure

```
internal/
├── config/
│   └── config.go          # existing: Manifest, Org, Skill
│                          # extend: PluginManifest, MarketplaceCatalog, PluginEntry
├── discovery/
│   └── discovery.go       # existing: FindOrgs, FindSkills
│                          # extend: ScanInstalledPlugins(repoPath)
├── linker/
│   └── linker.go          # no changes — symlink ops unchanged
├── marketplace/
│   └── marketplace.go     # new: GeneratePlugin, GenerateMarketplace, PackageSkill
├── publisher/
│   ├── publisher.go       # new: PublishToGitHub, GitCommitAndPush
│   └── registry.go        # new: QueryRemoteMarketplace, FetchPluginList
└── tui/
    ├── tui.go             # extend: marketplace tab, install flow, new view states
    └── styles.go          # extend: styles for marketplace states
```

### Structure Rationale

- **`internal/marketplace/`**: Generation is pure data transformation — takes `[]config.Skill` and produces JSON structures. No I/O beyond reading skill directories. Isolated so it's testable without touching the filesystem.
- **`internal/publisher/`**: Separated from marketplace generation because it has side effects (git push, network calls). Clean split between "build the artifact" and "ship the artifact."
- **`internal/discovery/`**: Extended rather than forked because plugin scanning follows the same filesystem-walking pattern as skill discovery. Keeps the single-responsibility boundary: discovery knows where things are.
- **`internal/config/`**: Extended in-place because new types (PluginManifest, MarketplaceCatalog) are still just parsed representations of JSON files on disk.

## Architectural Patterns

### Pattern 1: Extend Config Types, Don't Replace Them

**What:** Add new structs to `internal/config/config.go` for Claude Code's native JSON formats. Keep `Manifest`, `Org`, and `Skill` unchanged.

**When to use:** When new data structures represent files on disk that map 1:1 to existing concepts (a skill becomes a plugin entry; a brand repo becomes a marketplace).

**Trade-offs:** Keeps the domain layer thin and cohesive. Risks config.go growing large — split into `config/manifest.go` and `config/plugin.go` if it exceeds ~150 lines.

**Example:**
```go
// internal/config/config.go — new additions

// PluginManifest represents .claude-plugin/plugin.json
type PluginManifest struct {
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Description string   `json:"description"`
    Skills      []string `json:"skills"`
}

// MarketplaceCatalog represents .claude-plugin/marketplace.json
type MarketplaceCatalog struct {
    Name    string        `json:"name"`
    Plugins []PluginEntry `json:"plugins"`
}

// PluginEntry is one entry in a marketplace catalog
type PluginEntry struct {
    Name    string `json:"name"`
    Source  string `json:"source"` // e.g. "github:manzanita-research/chaparral"
    Version string `json:"version"`
}

// InstalledPlugin represents a discovered installed plugin in a sibling repo
type InstalledPlugin struct {
    Name    string
    RepoPath string
    Source  string
    State   string // "installed", "missing", "stale"
}
```

### Pattern 2: Pure Generation in `internal/marketplace`

**What:** `marketplace.go` takes existing config types as input and returns JSON-serializable structs. No filesystem writes — callers write the output. No git operations — publisher handles that.

**When to use:** Keeps the generation logic independently testable. The TUI preview, CLI generate command, and publisher all call the same generation function.

**Trade-offs:** Caller must handle file I/O. This is the right tradeoff — generation is deterministic and testable; I/O is the side effect to push to the boundary.

**Example:**
```go
// internal/marketplace/marketplace.go

// GenerateMarketplace builds a MarketplaceCatalog from org skills.
// Does not write any files — returns the catalog for caller to serialize.
func GenerateMarketplace(org config.Org, skills []config.Skill, version string) config.MarketplaceCatalog

// GeneratePlugin builds a PluginManifest for a single skill snapshot.
func GeneratePlugin(skill config.Skill, version string) config.PluginManifest

// PackageSkills copies skill directories into a versioned snapshot path.
// This IS a filesystem operation — it creates the snapshot at destDir.
func PackageSkills(skills []config.Skill, destDir string) error
```

### Pattern 3: Shell Out for `claude plugin install`

**What:** When installing a marketplace plugin into a sibling repo, run `claude plugin install plugin-name@marketplace-name` as a subprocess. Do not write Claude Code settings files directly.

**When to use:** Any time Chaparral needs to perform a Claude Code native operation. This keeps Chaparral in sync with however Claude Code evolves its install mechanics.

**Trade-offs:** Requires `claude` CLI to be installed and in `$PATH`. Must handle subprocess errors. Cannot show real-time progress without streaming output. This is worth it — writing settings directly would couple Chaparral to Claude Code internals.

**Example:**
```go
// internal/publisher/publisher.go

// InstallPlugin shells out to `claude plugin install <plugin>@<marketplace>`.
// Returns stdout+stderr on failure for display in TUI or CLI.
func InstallPlugin(pluginRef string) error {
    cmd := exec.Command("claude", "plugin", "install", pluginRef)
    out, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("install failed: %s", string(out))
    }
    return nil
}
```

### Pattern 4: Async TUI Operations via `tea.Cmd`

**What:** All marketplace operations that touch the network or shell out (publish, install, remote query) must run as goroutines returning `tea.Msg` — same pattern as existing `syncDone` and `orgsLoaded`.

**When to use:** Any operation that could block for >50ms. This already applies to sync; extend the same pattern to marketplace operations.

**Trade-offs:** Requires new message types (`marketplaceGenerated`, `publishDone`, `installDone`, `remoteLoaded`). Each needs a case in `Update()`. Worth the boilerplate — blocking the TUI event loop is a worse outcome.

**Example:**
```go
// internal/tui/tui.go — new message types

type marketplaceGenerated struct {
    catalog config.MarketplaceCatalog
    err     error
}

type publishDone struct {
    url string // GitHub URL of published marketplace
    err error
}

type installDone struct {
    plugin string
    err    error
}

type remoteLoaded struct {
    plugins []config.PluginEntry
    err     error
}
```

## Data Flow

### Generate Flow

```
User: chaparral generate  (CLI) or 'g' key (TUI)
    ↓
discovery.FindOrgs(basePath)  →  []config.Org
    ↓
discovery.FindSkills(org.SkillsPath())  →  []config.Skill
    ↓
marketplace.GenerateMarketplace(org, skills, version)  →  config.MarketplaceCatalog
marketplace.GeneratePlugin(skill, version)  →  config.PluginManifest  [per skill]
    ↓
Write .claude-plugin/marketplace.json and .claude-plugin/plugin.json
    ↓
marketplace.PackageSkills(skills, snapshotDir)  →  versioned snapshot at .claude-plugin/v{version}/
    ↓
CLI: print confirmation  |  TUI: marketplaceGenerated msg → viewDone
```

### Publish Flow

```
User: chaparral publish  (CLI) or 'p' key (TUI)
    ↓
[Generate flow above, if not already generated]
    ↓
publisher.GitCommitAndPush(brandRepoPath, message)
    ↓
[optional] publisher.CreateGitHubRelease(org, version)  →  GitHub API
    ↓
CLI: print GitHub URL  |  TUI: publishDone msg → viewDone with URL
```

### Plugin Status Scan Flow

```
TUI Init() or CLI status
    ↓
discovery.FindOrgs(basePath)  →  []config.Org
    ↓
for each org, for each sibling repo:
    discovery.ScanInstalledPlugins(repoPath)  →  []config.InstalledPlugin
    ↓
Merged with existing linker.StatusOrg() results
    ↓
TUI: orgsLoaded msg with combined local + plugin statuses
CLI: print unified status table
```

### Remote Registry Query Flow

```
User: chaparral status --remote  (CLI)  or 'R' key (TUI)
    ↓
config.Manifest.MarketplaceSource  →  "github:owner/repo"
    ↓
publisher.QueryRemoteMarketplace(source)
    → fetch raw GitHub URL for .claude-plugin/marketplace.json
    → parse into []config.PluginEntry
    ↓
Compare against locally installed plugins
    ↓
TUI: remoteLoaded msg → dashboard shows available vs installed
CLI: print diff table
```

### Install Flow

```
User: 'i' key in TUI marketplace view, or chaparral install plugin-name
    ↓
publisher.InstallPlugin("plugin-name@org-marketplace")
    → exec.Command("claude", "plugin", "install", ...)
    ↓
TUI: installDone msg → refresh plugin status
CLI: print success/failure
```

### State Management

```
TUI Model struct — new fields:

type Model struct {
    // existing
    basePath  string
    orgs      []config.Org
    statuses  map[string][]linker.LinkStatus
    results   []linker.LinkResult
    cursor    int
    view      view
    ...

    // new
    pluginStatuses  map[string][]config.InstalledPlugin  // keyed by org name
    remotePlugins   map[string][]config.PluginEntry       // keyed by org name
    marketplaceTab  marketplaceTabView                    // skills | marketplace
    publishURL      string                                // last published URL
}
```

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| GitHub raw content | HTTP GET to `raw.githubusercontent.com/{owner}/{repo}/main/.claude-plugin/marketplace.json` | Used for remote registry queries. No auth needed for public repos. Private repos require token — out of scope initially. |
| GitHub API | Optional: `POST /repos/{owner}/{repo}/releases` for tagged releases | Can use `gh` CLI subprocess instead of direct API to avoid token management complexity. |
| `claude` CLI | `exec.Command("claude", "plugin", "install", ...)` | Must be in `$PATH`. Fail gracefully with helpful message if not found. |
| git CLI | `exec.Command("git", "add", ...)`, `git commit`, `git push` | Subprocess model. Avoids go-git dependency. Consistent with how developers already work. |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| `config` ↔ `marketplace` | Direct function calls; `marketplace` imports `config` types | marketplace never imports linker or publisher |
| `discovery` ↔ `config` | Existing pattern; extend `ScanInstalledPlugins` to return `[]config.InstalledPlugin` | discovery remains read-only |
| `marketplace` ↔ `publisher` | Publisher calls marketplace generation functions before writing/pushing | publisher is the only package with side-effectful I/O |
| `tui` ↔ all internal packages | TUI imports everything; calls packages inside `tea.Cmd` goroutines | TUI remains the only presenter |
| `cmd/chaparral` ↔ all internal packages | CLI calls packages directly (synchronous); prints results | CLI remains the batch-mode presenter |

## Build Order

The dependency graph determines build order. Each phase can only start when its dependencies compile cleanly.

**Phase 1 — Config types (no dependencies)**
- Add `PluginManifest`, `MarketplaceCatalog`, `PluginEntry`, `InstalledPlugin` to `internal/config/config.go`
- No other code changes. Types available to all downstream packages immediately.
- Build check: `go build ./internal/config/...`

**Phase 2 — Discovery extension (depends on Phase 1)**
- Add `ScanInstalledPlugins(repoPath string) ([]config.InstalledPlugin, error)` to `internal/discovery/discovery.go`
- Reads `.claude/plugins/` directories in sibling repos (or equivalent installed location)
- Build check: `go build ./internal/discovery/...`

**Phase 3 — Marketplace generation (depends on Phase 1)**
- Create `internal/marketplace/marketplace.go`
- Pure generation: `GenerateMarketplace`, `GeneratePlugin`, `PackageSkills`
- Testable without network or subprocess
- Build check: `go build ./internal/marketplace/...`

**Phase 4 — Publisher (depends on Phases 1 and 3)**
- Create `internal/publisher/publisher.go` and `internal/publisher/registry.go`
- `publisher.go`: git subprocess ops, `claude plugin install` subprocess
- `registry.go`: HTTP fetch for remote marketplace JSON, parse into config types
- Build check: `go build ./internal/publisher/...`

**Phase 5 — CLI commands (depends on Phases 1-4)**
- Extend `cmd/chaparral/main.go` with `generate`, `publish`, `install` subcommands
- Wire up `runGenerate`, `runPublish`, `runInstall` handlers
- Build check: `go build ./cmd/chaparral`

**Phase 6 — TUI extension (depends on Phases 1-4)**
- Extend `internal/tui/tui.go`: new view states, new tab, new message types, new key bindings
- Add marketplace-specific styles to `internal/tui/styles.go`
- Build check: `go build ./internal/tui/...`

## Anti-Patterns

### Anti-Pattern 1: Writing Claude Code Settings Directly

**What people do:** Parse `~/.claude/settings.json` and write plugin configuration directly, avoiding the subprocess dependency.

**Why it's wrong:** Claude Code's settings format is internal and undocumented. A schema change in Claude Code silently breaks Chaparral. Users end up with corrupted settings. The whole point of `claude plugin install` is that it's the stable public interface.

**Do this instead:** Always shell out to `claude plugin install`. Detect if `claude` is not in `$PATH` and print a clear message telling the user how to install Claude Code.

### Anti-Pattern 2: Blocking the TUI Event Loop

**What people do:** Call `publisher.PublishToGitHub()` directly inside `Update()` when the user presses `p`.

**Why it's wrong:** Git push and network operations block the TUI goroutine, freezing the spinner and making the UI unresponsive. The terminal may appear hung.

**Do this instead:** Wrap all I/O operations in `tea.Cmd` goroutines returning custom `tea.Msg` types. This is how the existing `syncDone` pattern works — extend it.

### Anti-Pattern 3: Coupling Generation to Filesystem State

**What people do:** Have `GenerateMarketplace()` read the skills directory and write marketplace.json in one step.

**Why it's wrong:** Cannot test generation logic without touching the filesystem. Cannot preview what would be generated. Cannot reuse generation in both CLI and TUI contexts cleanly.

**Do this instead:** Keep `GenerateMarketplace()` pure — it receives `[]config.Skill` and returns a `config.MarketplaceCatalog`. Caller does the filesystem read and write. The `marketplace` package only owns the transformation logic.

### Anti-Pattern 4: Adding a New Discovery Scan to TUI Init Without Caching

**What people do:** Add plugin status scanning to `Model.Init()` sequentially, making TUI startup noticeably slower as it scans every repo for installed plugins.

**Why it's wrong:** Adding blocking operations to Init directly slows first paint. Users see a spinner for longer before the dashboard appears.

**Do this instead:** Run plugin scanning in the same goroutine as `FindOrgs`, but return all results together in `orgsLoaded`. The combined scan still happens once and is still async — it just includes plugin state in the same load batch.

## Scalability Considerations

This is a local CLI tool. Scale here means: works well for an org with 50 repos and 20 skills, and handles multiple orgs (up to ~10) without noticeable slowness.

| Concern | Approach |
|---------|----------|
| Many sibling repos | Plugin scanning is O(repos × installed plugins) — fast filesystem reads, no issue at 50 repos |
| Multiple orgs | Existing pattern: scan all orgs in one goroutine in Init; no parallelism needed at this scale |
| Remote registry fetch | HTTP GET to GitHub raw content; single request per marketplace; cache result in Model for the TUI session |
| Large skill snapshots | Snapshot copy at publish time; only matters at publish, not at scan or status time |

## Sources

- `internal/config/config.go` — Manifest, Org, Skill types; established type patterns
- `internal/discovery/discovery.go` — FindOrgs, FindSkills; established discovery patterns
- `internal/linker/linker.go` — SyncOrg, StatusOrg; established operations patterns
- `internal/tui/tui.go` — Model, Update, tea.Cmd async pattern; orgsLoaded/syncDone message types
- `.planning/PROJECT.md` — Claude Code plugin format specification: plugin.json, marketplace.json, `.claude-plugin/` directory, `claude plugin install` command, `~/.claude/plugins/cache` snapshot model
- `.planning/codebase/ARCHITECTURE.md` — Existing layer boundaries, data flow, error handling strategy

---
*Architecture research for: Chaparral marketplace bridge (Go CLI/TUI)*
*Researched: 2026-02-24*
