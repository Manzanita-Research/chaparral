# Project Research Summary

**Project:** Chaparral — marketplace bridge milestone
**Domain:** Go CLI/TUI — local-to-registry publisher for Claude Code plugins
**Researched:** 2026-02-24
**Confidence:** MEDIUM (codebase analysis high-confidence; Claude Code plugin format details medium-confidence; external search unavailable)

## Executive Summary

Chaparral is adding a marketplace bridge: the tooling that turns a brand repo's local skills into a published Claude Code plugin marketplace hosted on GitHub. The closest structural analogue from the wider ecosystem is `brew tap` — the "registry" is a GitHub repo, there is no hosted service to deal with, and the publish step means writing `.claude-plugin/marketplace.json` plus versioned plugin snapshots and pushing. This is a significantly simpler problem than hosted registries (npm, crates.io), which means the MVP surface area is small and the implementation risk is low if the sequencing is right.

The recommended approach is a four-phase build: config types first, then pure generation, then the GitHub publishing layer, then the TUI and CLI surface. Every phase has a clean compile boundary. The generation logic stays pure (no I/O, fully testable), while the publisher package owns all side-effectful operations. Two new dependencies are added: `google/go-github` for GitHub API calls and `golang.org/x/oauth2` for token auth. Everything else is stdlib or existing charmbracelet packages.

The critical risks are not technical — they are behavioral. Accidental publish (no confirmation step before git push), silent format drift (generated manifests breaking when Claude Code changes schema), and TUI model bloat (adding too many fields to a struct that already has 12) are the pitfalls most likely to cause real pain. All three are preventable with structure decisions made at the start of the first phase, not patched at the end.

## Key Findings

### Recommended Stack

The core Go 1.24.4 / Bubble Tea 1.3.10 / Lip Gloss 1.1.0 stack is unchanged. Two new libraries are needed. `google/go-github` (v68.x, verify before adding) handles all GitHub API operations: creating releases, writing `.claude-plugin/` files to the remote repo, and fetching published marketplace state. `golang.org/x/oauth2` accompanies it for `GITHUB_TOKEN`-based auth. All other new functionality — JSON generation, subprocess shell-outs to `claude plugin install`, git operations — is handled by Go stdlib packages already present in the codebase.

**Core technologies (additions only):**
- `google/go-github/v68`: GitHub REST API client — typed, handles auth and rate limiting, no `gh` CLI dependency
- `golang.org/x/oauth2`: Token auth for go-github — stable companion library, standard pattern
- `encoding/json` (stdlib): Manifest generation — already used in codebase, `json.MarshalIndent` is sufficient
- `os/exec` (stdlib): Shell out to `claude plugin install` and git commands — already present in codebase patterns

The decision to use go-github over shelling out to `gh` is correct: `gh` may not be installed, its output format is fragile, and there is no type safety. The decision to shell out to `claude plugin install` rather than writing settings files directly is also correct: Claude Code's settings format is internal; going around it risks silent breakage on any Claude Code update.

### Expected Features

The feature set maps cleanly to the `brew tap` + `gh release create` pattern. Generation and validation come first, publish comes second, TUI install flow comes third.

**Must have (table stakes):**
- Validate skills before packaging — catches broken structure, builds user trust
- Generate `plugin.json` per skill — required by Claude Code's format; nothing installs without it
- Generate `marketplace.json` — the catalog that makes a brand repo a marketplace
- Version stamping on publish — consumers cannot pin or upgrade without versions
- Dry run / preview mode — no serious user will run publish without it
- Publish to GitHub (write files + explicit git push) — the core bridge operation
- Status check (local vs. published sync state) — required for day-two use
- TUI publish flow with two-step confirmation — enforces the "never push without review" safety principle

**Should have (competitive, add post-validation):**
- Scan sibling repos for installed marketplace plugins
- Show marketplace plugin install status in Repos tab
- Install marketplace plugins from TUI
- `chaparral init` for new brand repos

**Defer (v2+):**
- Side-by-side local vs. published diff — high value, needs a publish baseline to exist first
- Changelog / release notes generation — scope creep risk for v1
- Multi-marketplace UI per org — architecturally supported; wait for a real use case

The TUI confirmation step before any git push is genuinely unusual versus every hosted-registry tool (npm, cargo, vsce all fire immediately). This is not a compromise — it is the differentiator, and it fits Manzanita's unhurried brand voice.

### Architecture Approach

The existing four-layer architecture (presentation → operations → domain → external) extends cleanly with two new internal packages. `internal/marketplace` owns pure generation: it takes `[]config.Skill` and returns JSON-serializable structs — no filesystem writes, fully testable. `internal/publisher` owns all side-effectful operations: git subprocess calls, GitHub API calls, and the `claude plugin install` shell-out. Existing packages (`internal/config`, `internal/discovery`) are extended in-place with new types and one new function. The TUI gains a third tab and new async message types following the existing `syncDone`/`orgsLoaded` pattern.

**Major components:**
1. `internal/config` (extended) — PluginManifest, MarketplaceCatalog, PluginEntry, InstalledPlugin types
2. `internal/discovery` (extended) — ScanInstalledPlugins(repoPath) for marketplace status
3. `internal/marketplace` (new) — GenerateMarketplace, GeneratePlugin, PackageSkills; pure transforms
4. `internal/publisher` (new) — PublishToGitHub, GitCommitAndPush, InstallPlugin, QueryRemoteMarketplace
5. `internal/tui` (extended) — third tab, viewPublishConfirm, new async message types
6. `cmd/chaparral` (extended) — generate, publish, install subcommands

### Critical Pitfalls

1. **Format assumption drift** — Generated manifests silently break when Claude Code updates its plugin schema. Prevention: define a single `PluginManifest` struct as the one source of truth; add a `chaparral validate` command; write a test that round-trips generated output through `json.Unmarshal`. Address this in Phase 1 before any publish capability exists.

2. **Accidental publish without confirmation** — A TUI keypress without a dedicated confirmation view can push to a public remote. Prevention: publish requires a new `viewPublishConfirm` view (not a modal) that shows exactly which files change, the version being published, and the target remote. A single `p` keypress must never trigger git push directly.

3. **TUI model bloat** — The existing `Model` has 12 fields. Adding marketplace state naively produces a 20+ field struct with an unreadable `Update()`. Prevention: before adding any marketplace fields, extract a `marketplaceState` struct. Route `Update()` by view/tab rather than handling all cases inline.

4. **Async operations without cancellation** — GitHub API queries and `claude plugin install` shell-outs block longer than local filesystem ops. Without `context.WithTimeout`, a hung network call leaves the TUI frozen. Prevention: wrap all network and subprocess operations with a 30-second context timeout; handle cancellation on `ctrl+c`.

5. **Shell-out failure invisible to user** — If `claude plugin install` fails, a generic error message tells the user nothing. Prevention: capture `CombinedOutput()` on every exec call; surface full stderr in the TUI results view; run `exec.LookPath("claude")` before any install attempt.

6. **Snapshot versioning skipped** — Without version tracking, consumers cannot distinguish stale installs from updated ones. Prevention: store `plugin_version` in `chaparral.json`; auto-increment patch on publish when content changes; show version prominently in the confirmation view.

## Implications for Roadmap

Based on research, the build order is determined by the dependency graph: types must exist before generation, generation before publishing, publishing before TUI surface. Each phase has a clean compile-check boundary.

### Phase 1: Config Types and Manifest Generation

**Rationale:** Everything downstream depends on correct typed structs and deterministic JSON generation. Format drift (Pitfall 1) and version management (Pitfall 6) must be addressed here, before any I/O or network code exists. Generation being pure (no I/O) makes this phase fully unit-testable.

**Delivers:** `internal/config` extended with plugin types; `internal/marketplace` package with GenerateMarketplace, GeneratePlugin, PackageSkills; a `chaparral validate` command; version tracking field in chaparral.json

**Addresses:** validate skills, generate plugin.json per skill, generate marketplace.json, version stamping, dry run preview (manifest output to stdout)

**Avoids:** format assumption drift, snapshot versioning without enforcement

### Phase 2: Publisher — GitHub and Git Operations

**Rationale:** Publishing has side effects (git push, network calls) that must be isolated from generation. The confirmation view (Pitfall 2) and context cancellation (Pitfall 4) must be implemented before any real push capability ships. This phase introduces the two new dependencies.

**Delivers:** `internal/publisher` package with GitCommitAndPush, PublishToGitHub, QueryRemoteMarketplace; `GITHUB_TOKEN` token management; `viewPublishConfirm` TUI view; `chaparral publish` CLI command

**Uses:** `google/go-github/v68`, `golang.org/x/oauth2`, `os/exec` for git subprocess

**Implements:** publisher layer, external GitHub boundary, explicit-push safety convention

**Avoids:** accidental publish without confirmation, async operations without cancellation

### Phase 3: TUI Marketplace Tab

**Rationale:** The TUI surface is the last layer to add. Extending the model before the underlying packages are stable risks churn. Model structure (Pitfall 3) must be addressed as the first task of this phase — extract `marketplaceState` struct before adding any new fields.

**Delivers:** third Marketplace tab in the dashboard; plugin status display; `viewPublishing` and `viewPublishDone` view states; new async message types (marketplaceGenerated, publishDone); keyboard bindings for generate/publish within TUI

**Implements:** async TUI operations following existing syncDone pattern; model routing by view/tab

**Avoids:** TUI model bloat, message type collision, subprocess failure invisible

### Phase 4: Plugin Install and Scan

**Rationale:** Install capability depends on knowing what is available (scan), which depends on the publish flow being stable. This is a v1.x addition — adds polish and closes the loop from publish to install in one tool.

**Delivers:** `discovery.ScanInstalledPlugins()`; install status in Repos TUI tab; `viewInstalling` state; `chaparral install` CLI command; shell-out to `claude plugin install` with full stderr capture

**Addresses:** scan sibling repos for installed plugins, show marketplace plugin install status, install from TUI

**Avoids:** shell-out failure invisible to user

### Phase Ordering Rationale

- Types before generation before publishing enforces the clean compile-boundary build order from ARCHITECTURE.md.
- Generation must be pure before publishing exists — otherwise the only way to test generation is through a live publish, which is slow and side-effectful.
- The confirmation view ships in Phase 2 (same phase as git push) because it cannot be separated from the operation it guards.
- TUI extensions come after the underlying packages are stable to avoid building UI on moving foundations.
- Install scan comes last because it has no value until there is something to install (requires Phase 2 published output).

### Research Flags

Phases likely needing deeper research during planning:

- **Phase 2 (Publisher):** go-github version needs verification at `pkg.go.dev/github.com/google/go-github` before `go get`. Claude Code plugin format (plugin.json required fields, marketplace.json schema) needs verification against current Claude Code docs — this information came from PROJECT.md, not an independent source.
- **Phase 4 (Install):** The exact location of installed plugin cache (`~/.claude/plugins/cache` or equivalent) needs verification. The `claude plugin install` command syntax (`plugin-name@marketplace-name`) is from PROJECT.md; should be tested against the actual CLI before shipping.

Phases with standard patterns (research not needed):

- **Phase 1 (Config + Generation):** Pure Go struct definition and JSON marshaling. No new patterns. Well-understood stdlib usage already present in codebase.
- **Phase 3 (TUI Extension):** Direct extension of existing Bubble Tea patterns documented in the codebase. Third tab follows the same structure as the first two.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | MEDIUM | go-github is canonical; version number needs verification before go get. oauth2 is stable HIGH. |
| Features | MEDIUM | Table stakes and differentiators well-grounded in npm/vsce/brew tap analogues. Claude Code-specific format details from PROJECT.md only — not independently verified. |
| Architecture | HIGH | Based on direct codebase analysis. Layer boundaries, data flows, and build order are derived from reading actual Go source. |
| Pitfalls | MEDIUM | Grounded in real patterns from registry tooling and Bubble Tea app architecture. No external verification available. |

**Overall confidence:** MEDIUM

### Gaps to Address

- **Claude Code plugin format:** plugin.json and marketplace.json schemas are described in PROJECT.md but not verified against current Claude Code documentation. Before Phase 1 ships, confirm exact required fields with a test install.
- **go-github version:** v68 is current as of late 2025 per training knowledge. Run `pkg.go.dev/github.com/google/go-github` lookup before `go get` — major version in import path changes with each major release.
- **Installed plugin location:** `~/.claude/plugins/cache` is the expected path from PROJECT.md. Verify this before implementing ScanInstalledPlugins in Phase 4.
- **Model refactor scope:** The existing `Model` struct has 12 fields and a non-trivial `Update()`. Before Phase 3, review whether a `marketplaceState` sub-struct is sufficient or whether a full parent/child model delegation is needed.

## Sources

### Primary (HIGH confidence)
- `/Users/jem/code/manzanita-research/chaparral/internal/` — all Go source files; direct codebase analysis
- `/Users/jem/code/manzanita-research/chaparral/.planning/PROJECT.md` — milestone requirements, plugin format spec
- `/Users/jem/code/manzanita-research/chaparral/.planning/codebase/ARCHITECTURE.md` — existing layer boundaries

### Secondary (MEDIUM confidence)
- `google/go-github` at pkg.go.dev — canonical Go GitHub client; version confirmed as v68.x as of late 2025 training data
- `golang.org/x/oauth2` — official Go OAuth2 library; stable API
- npm CLI, vsce, cargo publish, Homebrew tap — analogues for feature and pattern research; well-documented in training data

### Tertiary (needs validation before implementation)
- Claude Code plugin format (plugin.json / marketplace.json required fields) — from PROJECT.md only; not verified against live docs
- `claude plugin install` command syntax — from PROJECT.md; not verified against current Claude Code CLI
- `~/.claude/plugins/cache` installed plugin location — from PROJECT.md; needs verification

---
*Research completed: 2026-02-24*
*Ready for roadmap: yes*
