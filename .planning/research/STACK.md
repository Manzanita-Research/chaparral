# Stack Research

**Domain:** Go CLI/TUI — marketplace bridge for Claude Code plugin distribution
**Researched:** 2026-02-24
**Confidence:** MEDIUM (web fetch and Context7 denied; based on codebase analysis + training knowledge for Go ecosystem; versions flagged where uncertain)

---

## Context

This is a **subsequent milestone** on an existing Go 1.24.4 / Bubble Tea 1.3.10 / Lip Gloss 1.1.0 codebase (~1200 lines across 5 packages). The core stack is not changing. This document covers only the **new libraries and patterns** needed to add:

1. JSON manifest generation (`plugin.json`, `marketplace.json`)
2. GitHub API interaction (creating releases, committing files, pushing to repos)
3. Shelling out to `claude plugin install`
4. Extending the Bubble Tea TUI with a third tab and new async operations

---

## Recommended Stack

### Core Technologies (Existing — No Change)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.24.4 | All application code | Existing; compiled binary distribution |
| Bubble Tea | 1.3.10 | TUI event loop | Existing; already used for dashboard |
| Lip Gloss | 1.1.0 | Terminal styling | Existing; design language baked in |
| Bubbles | 1.0.0 | TUI components (spinner) | Existing |

### New Libraries

| Library | Version | Purpose | Confidence |
|---------|---------|---------|------------|
| `github.com/google/go-github/v68` | v68.x | GitHub REST API — creating releases, pushing file contents, reading repo metadata | MEDIUM — v68 is current as of late 2025 based on training; verify latest at pkg.go.dev before adding |
| `golang.org/x/oauth2` | v0.x (latest) | GitHub token auth (required by go-github) | HIGH — standard OAuth2 companion to go-github, stable API |
| Go stdlib `encoding/json` | built-in | Write `plugin.json` and `marketplace.json` manifests | HIGH — already used in codebase, `json.MarshalIndent` covers write path |
| Go stdlib `os/exec` | built-in | Shell out to `claude plugin install <name@marketplace>` | HIGH — already used in codebase patterns; correct choice for delegating to external CLI |

---

## Detailed Recommendations

### 1. JSON Manifest Generation — encoding/json (stdlib)

**Verdict: Use stdlib only. No external library needed.**

The existing codebase already uses `encoding/json` with struct tags for reading `chaparral.json`. The same pattern extends cleanly to writing `plugin.json` and `marketplace.json` via `json.MarshalIndent`.

Define structs matching Claude Code's native format:

```go
// internal/marketplace/types.go

type PluginManifest struct {
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Description string   `json:"description"`
    Skills      []string `json:"skills"`
    // extend as Claude Code plugin.json spec requires
}

type MarketplaceEntry struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Source      string `json:"source"` // e.g. "github:owner/repo"
    Version     string `json:"version"`
}

type MarketplaceManifest struct {
    Name    string             `json:"name"`
    Plugins []MarketplaceEntry `json:"plugins"`
}
```

Write with:

```go
data, err := json.MarshalIndent(manifest, "", "  ")
if err != nil { return err }
err = os.WriteFile(destPath, data, 0644)
```

**Why not a schema validation library:** Claude Code's plugin format is well-defined and not complex. A validation library adds a dependency without benefit for a tool generating a known fixed schema. If the schema changes, the struct fields change — simple.

**Confidence: HIGH** — stdlib pattern, matches existing codebase style.

---

### 2. GitHub API — google/go-github

**Verdict: Use google/go-github for API calls. Do NOT shell out to `gh` CLI.**

Two options exist for GitHub API interaction:

**Option A: `github.com/google/go-github/v68`**
- Strongly-typed Go client for the GitHub REST API
- Handles authentication, rate limiting, pagination
- Idiomatic Go error handling
- No external process dependency

**Option B: Shell out to `gh` CLI**
- Simpler code surface initially
- Requires `gh` to be installed and authenticated separately from the tool
- Parsing output is fragile; `gh` output format changes between versions
- No type safety; errors are string matching

**Recommendation: go-github.** Here is why:

The tool is already a compiled binary. Users authenticate once by setting `GITHUB_TOKEN` (or a config entry). go-github uses that token directly — no assumption that the user has `gh` installed. The tool needs specific operations (create release, upload asset, commit a file tree) that are straightforward API calls but verbose to script reliably via `gh`. The go-github client handles all of this cleanly.

**Authentication pattern:**

```go
import (
    "context"
    "github.com/google/go-github/v68/github"
    "golang.org/x/oauth2"
)

func newGitHubClient(token string) *github.Client {
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
    tc := oauth2.NewClient(context.Background(), ts)
    return github.NewClient(tc)
}
```

**Key API surface used:**

```go
// Create/update a file in the repo (for .claude-plugin/marketplace.json)
client.Repositories.CreateOrUpdateFileContents(ctx, owner, repo, path, opts)

// Create a release tag
client.Repositories.CreateRelease(ctx, owner, repo, release)

// Upload a release asset (zipped plugin snapshot)
client.Repositories.UploadReleaseAsset(ctx, owner, repo, releaseID, opts, file)

// Get current file SHA (required for update operations)
client.Repositories.GetContents(ctx, owner, repo, path, opts)
```

**Token sourcing:** Check `GITHUB_TOKEN` env var first, then fall back to a `~/.config/chaparral/config.json` entry. Never hard-code. Never read from `~/.config/gh/hosts.yml` (fragile, gh-internal format).

**Version note:** `v68` is the version current as of late 2025. By Feb 2026, a higher minor may exist. Verify at `pkg.go.dev/github.com/google/go-github` before `go get`. The import path changes with major versions (`v68` → `v69` would require import path update).

**Confidence: MEDIUM** — go-github is the canonical Go GitHub client with a stable, long-running API. Version number needs verification.

---

### 3. Shelling Out to `claude plugin install` — os/exec (stdlib)

**Verdict: Use `os/exec` directly. No wrapper library needed.**

The existing codebase uses `os/exec` patterns already (via stdlib imports). The `claude plugin install` operation is a one-shot shell-out: run the command, capture stdout/stderr, report outcome.

```go
// internal/publisher/installer.go

import "os/exec"

type InstallResult struct {
    Plugin string
    Repo   string
    Output string
    Err    error
}

func InstallPlugin(pluginName, marketplaceRepo string) InstallResult {
    arg := fmt.Sprintf("%s@%s", pluginName, marketplaceRepo)
    cmd := exec.Command("claude", "plugin", "install", arg)
    out, err := cmd.CombinedOutput()
    return InstallResult{
        Plugin: pluginName,
        Repo:   marketplaceRepo,
        Output: string(out),
        Err:    err,
    }
}
```

**Key patterns to follow:**

- Use `cmd.CombinedOutput()` to capture both stdout and stderr — useful for surfacing `claude`'s own error messages in the TUI
- Check `exec.LookPath("claude")` before attempting to install and surface a helpful error ("claude CLI not found — install from claude.ai/code") rather than letting exec fail with a raw error
- Run via `tea.Cmd` (anonymous func returning `tea.Msg`) to avoid blocking the TUI event loop — same pattern as existing `syncOrg` and `syncAll`

**Why not a subprocess management library:** The interaction is simple (one-shot, no streaming stdin, no long-lived process). subprocess/pty libraries add complexity with no benefit here.

**Confidence: HIGH** — os/exec is the correct stdlib tool; pattern is already present in the codebase.

---

### 4. Extending Bubble Tea TUI — New Tab + Views

**Verdict: Extend existing Model in-place. Add `tabMarketplace dashTab` constant. No new library needed.**

The existing TUI uses a `dashTab` int enum (`tabSkills`, `tabRepos`). Adding a third tab is straightforward:

```go
const (
    tabSkills     dashTab = iota
    tabRepos
    tabMarketplace // new
)
```

**Tab cycling pattern:**

```go
case "tab":
    if m.view == viewDashboard {
        m.tab = (m.tab + 1) % 3 // cycles through all three tabs
        m.cursor = 0
    }
```

**New async messages** follow existing `orgsLoaded` / `syncDone` pattern:

```go
type marketplaceStatusLoaded struct {
    statuses map[string][]MarketplacePluginStatus
    err      error
}

type publishDone struct {
    results []PublishResult
    err     error
}

type installDone struct {
    result InstallResult
}
```

**New view states:**

```go
const (
    viewDashboard view = iota
    viewSyncing
    viewDone
    viewHelp
    viewPublishing  // new — show spinner while GitHub API call runs
    viewPublishDone // new — show publish result summary
    viewInstalling  // new — show spinner while claude plugin install runs
)
```

**Render pattern** for marketplace tab follows existing `renderSkillsTab` / `renderReposTab` style — a `renderMarketplaceTab(*strings.Builder, statuses)` method on `Model`.

**Model struct additions:**

```go
type Model struct {
    // ... existing fields ...
    marketplaceStatuses map[string][]MarketplacePluginStatus // keyed by org name
    publishResults      []PublishResult
    installResults      []InstallResult
    githubToken         string // loaded from env or config on init
}
```

**Why not a separate sub-model per tab:** The existing design is a single flat `Model` with view state. Splitting into sub-models would require a parent/child message delegation pattern (common in larger Bubble Tea apps but an overhead rewrite for this codebase). Keep it flat and consistent with what exists.

**Confidence: HIGH** — this is a direct extension of the existing pattern documented in the codebase.

---

## New Internal Package

Add one new package for the marketplace operations:

```
internal/
  marketplace/   — NEW
    types.go     — PluginManifest, MarketplaceManifest, MarketplacePluginStatus structs
    generator.go — Generate plugin.json and marketplace.json from org skills
    publisher.go — Push to GitHub via go-github; create releases; manage tokens
    scanner.go   — Scan sibling repos for installed marketplace plugins
    installer.go — Shell out to claude plugin install
```

This fits the existing layered architecture: `marketplace` is a new operations layer, consumed by the TUI layer (new tab) and a new `chaparral publish` CLI command.

---

## Alternatives Considered

| Recommended | Alternative | Why Not |
|-------------|-------------|---------|
| `google/go-github` | Shell out to `gh` CLI | gh may not be installed; output parsing is fragile; no type safety |
| `google/go-github` | Raw `net/http` calls to GitHub API | Reinvents auth, rate limiting, pagination, struct mapping — all solved by go-github |
| `os/exec` for `claude plugin install` | Writing to `~/.claude/settings.json` directly | Claude Code's plugin install does more than write a JSON entry (cache copy, namespace setup); going around the CLI risks breaking things |
| Extend existing `Model` with new tab | New separate Bubble Tea program for marketplace | Two programs = two binaries or complex launch handoff; one cohesive dashboard is the product vision |
| `encoding/json` stdlib | `goccy/go-json` or `json-iterator` | Performance is irrelevant for manifest files; stdlib is simpler, no dependency |

---

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `go-git` (in-process git) | Would require implementing auth, commit graph, push — much higher complexity than a GitHub API call for the operations needed | `google/go-github` for API-level operations; for local git inspection (if needed), shell out to `git` via `os/exec` |
| `github.com/cli/go-gh` | gh's internal Go library — not designed for use outside the gh CLI itself; undocumented stability guarantees | `google/go-github` |
| Goreleaser or similar | Build tooling for releasing Chaparral itself, not for publishing plugin content | not applicable |
| OAuth device flow / web flow | Too much ceremony for a CLI; users are developers who can set a token | `GITHUB_TOKEN` env var with `golang.org/x/oauth2` static token |

---

## Installation

```bash
# Add to existing module (verify latest v-number at pkg.go.dev first)
go get github.com/google/go-github/v68@latest
go get golang.org/x/oauth2@latest
```

No other new dependencies needed. All other requirements are covered by stdlib or existing charmbracelet packages.

---

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| `go-github/v68` | `golang.org/x/oauth2` (any recent) | go-github's oauth2 dep constraint is loose; latest oauth2 is fine |
| `go-github/v68` | Go 1.24.4 | go-github supports Go 1.21+; no compatibility issue |
| New stdlib patterns | Bubble Tea 1.3.10 | `os/exec` runs in goroutines wrapped as `tea.Cmd` — same pattern as existing sync ops |

---

## Stack Patterns by Variant

**If GitHub token is not set:**
- Surface an actionable error in the TUI: "set GITHUB_TOKEN to publish to GitHub"
- Disable publish/install keybindings in the marketplace tab UI
- Allow browsing marketplace status without a token (read-only scan of local plugin cache)

**If `claude` CLI is not in PATH:**
- Surface: "claude CLI not found — install from claude.ai"
- Disable install keybinding
- Allow everything else (generate manifests, publish to GitHub)

**If running in CI / non-interactive:**
- `chaparral publish` CLI command (no TUI) handles this case
- Use `NO_COLOR` convention (already respected)
- Exit code 1 on publish failure, 0 on success

---

## Sources

- `/Users/jem/code/manzanita-research/chaparral/go.mod` — existing dependencies, Go version
- `/Users/jem/code/manzanita-research/chaparral/internal/tui/tui.go` — existing TUI patterns (tab enum, view enum, tea.Cmd async pattern)
- `/Users/jem/code/manzanita-research/chaparral/internal/config/config.go` — existing JSON parsing pattern
- `/Users/jem/code/manzanita-research/chaparral/internal/linker/linker.go` — existing operation layer pattern
- `/Users/jem/code/manzanita-research/chaparral/.planning/PROJECT.md` — milestone requirements
- `google/go-github` — canonical Go GitHub client, pkg.go.dev; training knowledge (v68 is current as of late 2025; verify before use) — MEDIUM confidence on version
- `golang.org/x/oauth2` — official Go OAuth2 library; stable API — HIGH confidence
- `os/exec` — Go stdlib, stable — HIGH confidence

---

*Stack research for: Chaparral marketplace bridge milestone*
*Researched: 2026-02-24*
