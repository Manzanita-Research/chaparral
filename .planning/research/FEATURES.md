# Feature Research

**Domain:** Local-to-registry bridge tooling (marketplace bridge for Claude Code plugin system)
**Researched:** 2026-02-24
**Confidence:** MEDIUM — WebSearch and WebFetch unavailable. Analysis draws on well-established patterns from npm publish, cargo publish, vsce (VS Code Extension Manager), brew tap creation, and GitHub Actions publishing workflows. Patterns in this space are stable and well-documented in training data.

---

## Context: What Kind of Tool This Is

Chaparral's marketplace bridge is a **local-to-registry packager/publisher** — the same category as:

- `npm publish` — packages a local directory into a tarball, pushes to npm registry
- `cargo publish` — validates, packages, and uploads a Rust crate
- `vsce package` / `vsce publish` — bundles a VS Code extension and submits to marketplace
- `brew tap` — points Homebrew at a GitHub repo as a formula source
- `gh release create` — packages local artifacts and creates a versioned GitHub release

The specific analogy that maps most closely to chaparral's design is **brew tap + gh release**: the "registry" is a GitHub repo, not a hosted service, and there's no auth service to deal with beyond git push. This significantly simplifies the feature set.

---

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Generate marketplace.json from skills | Core reason the bridge exists — turns local skills into a publishable catalog | LOW | Format is Claude Code's native `.claude-plugin/marketplace.json`. Skills already discovered by existing discovery package. Main work is JSON generation. |
| Generate per-plugin plugin.json manifests | Claude Code requires `plugin.json` per plugin — without it nothing installs | LOW | One file per skill directory. Fields: name, version, description, skills array. Straightforward struct-to-JSON. |
| Version stamping on publish | Every publish tool versions its output. Users expect semver or date-based versions. Without versions, consumers can't pin or upgrade reliably | MEDIUM | Must embed version into plugin.json at publish time. Need to decide: auto-increment, explicit flag, or git tag. |
| Dry run / preview mode | Every mature publish tool (npm publish --dry-run, cargo publish --dry-run) has this. Users want to see what would happen before committing | LOW | Show what files would be written/pushed without writing them. Straightforward flag. |
| Status check before publish | Show current state: are local skills newer than published? Is marketplace.json in sync? | MEDIUM | Requires comparing local skills list against what's currently in `.claude-plugin/` on disk or in git. Extends existing `status` command concept. |
| Git-based publishing (push to GitHub) | The brand repo IS the marketplace repo. "Publish" means writing the `.claude-plugin/` structure and pushing | MEDIUM | Shell out to git or use go-git. Must respect the explicit-push-only convention from PROJECT.md — never push without user confirmation. |
| List what would be published | Show skills that will be included, skipped, and why | LOW | Pre-publish report. Maps naturally to existing skills discovery. |
| Idempotent file generation | Running generate twice should produce identical output if nothing changed | LOW | Deterministic JSON serialization. Sort keys/arrays. Already a convention in the codebase. |
| Validate skills before packaging | Catch missing SKILL.md, missing required fields, broken directory structure before attempting publish | LOW | Extends existing FindSkills validation. Add field-level checks. |
| Clear error messages when git push fails | If push fails (no remote, no auth, merge conflict), the error must be actionable | LOW | Error message quality, not a feature per se. Part of the existing codebase convention. |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| TUI publish flow with confirmation step | vs. fire-and-forget CLI commands, showing a summary and asking "publish these 4 skills? [y/n]" before any git operations | MEDIUM | Fits naturally into existing Bubble Tea TUI. Confirmation model mirrors the "never push without explicit user action" constraint in PROJECT.md. |
| Side-by-side local vs. published diff in TUI | Show which skills are ahead of what's published, which are in sync, which have been removed locally but still published | HIGH | Requires tracking published state (what's in `.claude-plugin/`) vs. local skills. Valuable for multi-person teams or multi-machine workflows. |
| Multi-marketplace support per org | Each brand repo becomes its own marketplace. Chaparral already supports multiple orgs — extending to multiple brand repos per org is architecturally natural | MEDIUM | PROJECT.md explicitly calls this out as a requirement. The differentiator is surfacing this clearly in the TUI rather than hiding it. |
| Scan sibling repos for installed marketplace plugins | Show which repos have which marketplace plugins installed — not just local symlinks | MEDIUM | Requires reading `.claude/plugins/` or settings files in each sibling repo. Clarifies what `claude plugin install` has already been run. |
| Install marketplace plugins from TUI | Shell out to `claude plugin install plugin-name@marketplace-name` from within the dashboard | MEDIUM | Keeps user in one tool. Dependency: must know which marketplace a plugin belongs to, and that the repo has the marketplace added. |
| Changelog / release notes generation | Auto-draft release notes from git log between last publish and now | HIGH | npm version + conventional commits pattern. Valuable but high complexity. Defer to v2. |
| `chaparral init` for new brand repos | Bootstrap a new brand repo with correct directory structure, chaparral.json, and empty .claude-plugin/ scaffolding | LOW | One-time setup helper. Reduces friction for new orgs adopting the tool. Strong first-run experience. |
| Show marketplace plugin install status per repo | In the Repos TUI tab, show not just symlink status but also which marketplace plugins are installed | MEDIUM | Unified view of both distribution modes in one screen is a core UX promise from PROJECT.md and CLAUDE.md. |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Auto-push on sync | "One command does everything" feels fast and clean | Pushes without user review violate the core safety principle. Users have burned themselves with auto-push in git workflows (git push --force accidents, pushing credentials). PROJECT.md explicitly forbids this. | Two-step: generate + explicit publish. Show what will be pushed, confirm, then push. |
| Writing Claude Code settings files directly | Installing plugins by editing `~/.claude/settings.json` directly seems faster than shelling out | Bypasses Claude Code's own install system. When Claude Code updates its settings format, chaparral silently breaks. Shell out to `claude plugin install` as PROJECT.md specifies. | Shell out to `claude plugin install`. Stay in sync with the system. |
| Custom plugin format / plugin registry | "Why use Claude Code's format? We could make our own." | Inventing a format means maintaining a spec, writing parsers on the consumer side, and diverging from the ecosystem. Every consumer would need chaparral installed. | Generate Claude Code's native format. Zero consumer-side requirements. |
| Automatic version bumping | "Bump the version automatically on every publish" | Removes developer judgment from versioning. Leads to 47 patch releases with identical behavior. Breaks semver semantics. | Require explicit version flag or prompt. Default to showing current version and asking. |
| Plugin dependency resolution | "Skills should be able to declare deps on other skills" | Adds a dependency graph that must be resolved, stored, and kept in sync. This is the npm dependency hell problem for a much smaller ecosystem. | Keep skills flat and independent. If two skills must ship together, they're one plugin. |
| Analytics / telemetry on skill usage | "Know which skills are actually being used" | Surveillance of creative work. Directly violates Manzanita's no-data-collection value. Would require network calls from a local tool. | Trust users. If a skill is bad, they'll stop using it. Collect feedback via GitHub issues. |
| Cloud-hosted marketplace registry | "Host a chaparral.io so others can discover our marketplace" | Adds infrastructure, hosting costs, a service to maintain, and a central point of failure. Contradicts the local-first, no-subscription ethos. | GitHub IS the registry. A public GitHub repo with marketplace.json is discoverable via `claude plugin marketplace add`. |
| Interactive skill editor in TUI | "Edit skill content from within the dashboard" | Scope creep. Chaparral packages skills, it doesn't author them. Adding an editor conflates two concerns. | Open the skill directory in the user's editor via `$EDITOR`. Keep authoring outside chaparral. |

---

## Feature Dependencies

```
[Validate skills]
    └──required by──> [Generate plugin.json per skill]
                          └──required by──> [Generate marketplace.json]
                                                └──required by──> [Publish to GitHub]

[Git status check]
    └──required by──> [Publish to GitHub]
                          └──enables──> [Side-by-side local vs. published diff]

[Scan sibling repos for installed plugins]
    └──required by──> [Show marketplace plugin install status in Repos TUI tab]
                          └──required by──> [Install marketplace plugins from TUI]

[Generate marketplace.json]
    └──enhances──> [chaparral init] (init scaffolds the structure generate fills)

[Version stamping]
    └──required by──> [Publish to GitHub]
    └──required by──> [Status check before publish] (need to compare versions)
```

### Dependency Notes

- **Validate skills requires nothing new**: Extends existing `FindSkills()` with field-level checks. Must happen before any generation.
- **Generate plugin.json and marketplace.json are sequential**: Plugin manifests must exist before marketplace catalog references them.
- **Publish requires explicit git operations**: Writing files to `.claude-plugin/` and pushing are separate steps. The TUI confirmation step sits between them.
- **Scan for installed plugins is independent**: Does not require the publish pipeline. Can be added to the status/discovery layer without affecting generation.
- **Install from TUI depends on scan**: You can only install what you've discovered is available. Must know the marketplace name and plugin name before shelling out to `claude plugin install`.
- **Side-by-side diff depends on publish having happened at least once**: Requires a `.claude-plugin/marketplace.json` baseline to compare against. If no baseline exists, falls back to "nothing published yet."

---

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the bridge concept.

- [ ] Validate skills before packaging — catches broken structure early, builds trust
- [ ] Generate plugin.json per skill — required by Claude Code's format, zero value without this
- [ ] Generate marketplace.json — the catalog that makes a brand repo a marketplace
- [ ] Dry run / preview mode — required before anyone will use publish in a real org
- [ ] Publish to GitHub (write files + explicit git push step) — the actual bridge. Without this, nothing is distributed
- [ ] Status check (are skills in sync with published marketplace.json?) — table stakes for day-2 use
- [ ] TUI publish flow with confirmation step — keeps with the existing TUI character; also enforces the safety convention

### Add After Validation (v1.x)

Features to add once core publish flow is working and trusted.

- [ ] Scan sibling repos for installed marketplace plugins — adds value once there's something to install
- [ ] Show marketplace plugin install status in Repos TUI tab — requires scan to be working
- [ ] Install marketplace plugins from TUI — high convenience, depends on scan and confirmed stable publish format
- [ ] `chaparral init` for new brand repos — reduces friction for new adopters, low complexity

### Future Consideration (v2+)

Features to defer until publish flow is proven stable.

- [ ] Side-by-side local vs. published diff — high value but high complexity; needs publish baseline to exist first
- [ ] Changelog / release notes generation — valuable but requires git log parsing and formatting decisions; scope creep risk for v1
- [ ] Multi-marketplace support per org — architecturally supported already, but UI surface area is meaningful; wait for a real use case

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Validate skills | HIGH | LOW | P1 |
| Generate plugin.json per skill | HIGH | LOW | P1 |
| Generate marketplace.json | HIGH | LOW | P1 |
| Dry run / preview | HIGH | LOW | P1 |
| Version stamping | HIGH | MEDIUM | P1 |
| Publish to GitHub (write + push) | HIGH | MEDIUM | P1 |
| Status check before publish | HIGH | MEDIUM | P1 |
| TUI publish flow with confirmation | HIGH | MEDIUM | P1 |
| `chaparral init` | MEDIUM | LOW | P2 |
| Scan installed marketplace plugins | MEDIUM | MEDIUM | P2 |
| Plugin install status in Repos tab | MEDIUM | MEDIUM | P2 |
| Install from TUI | MEDIUM | MEDIUM | P2 |
| Side-by-side local vs. published diff | HIGH | HIGH | P3 |
| Changelog generation | MEDIUM | HIGH | P3 |
| Multi-marketplace UI | MEDIUM | HIGH | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

---

## Competitor Feature Analysis

The closest analogues and what chaparral can learn from each:

| Feature | npm publish | vsce (VS Code) | cargo publish | brew tap | Our Approach |
|---------|-------------|----------------|---------------|----------|--------------|
| Dry run | `--dry-run` flag | `--dry-run` flag | `--dry-run` flag | n/a | Dry run flag + TUI preview |
| Version management | Requires version in package.json, `npm version` to bump | Requires version in package.json | Requires version in Cargo.toml | Tags in GitHub | Explicit version flag on publish; show current on confirm |
| Auth | npm token via env or .npmrc | Personal access token via vsce login | crates.io API token | git push auth (SSH/HTTPS) | git push auth only — no new auth system |
| Validation pre-publish | Checks package.json fields | Checks package.json, extension entry | Checks Cargo.toml | n/a | Validate SKILL.md presence, required fields |
| File inclusion control | .npmignore / files field | .vscodeignore | Cargo.toml include/exclude | n/a | Manifest exclude list (already exists in chaparral.json) |
| Diff / status | `npm outdated` (consumer side) | n/a | n/a | `brew outdated` | Status command comparing local vs. published |
| Confirmation | None (fires immediately) | None (fires immediately) | None (fires immediately) | n/a | TUI confirmation step — our differentiator |
| Registry type | Hosted service (npm Inc.) | Hosted service (Microsoft) | Hosted service (crates.io) | GitHub repo | GitHub repo — our differentiator |

**Key insight:** Every hosted-registry tool fires immediately with no confirmation. Chaparral's TUI confirmation step is genuinely unusual and fits Manzanita's "unhurried" brand voice. It's not a compromise — it's a feature.

**Key insight:** brew tap is the closest structural analogue (GitHub repo as registry, no custom hosting), but brew tap requires no tooling to consume — just `brew tap owner/repo`. Claude Code's `claude plugin marketplace add owner/repo` works the same way. Chaparral is the tooling that creates and maintains the tap-equivalent structure.

---

## Confidence Assessment

| Claim | Confidence | Basis |
|-------|------------|-------|
| npm/cargo/vsce all have dry-run flags | HIGH | Well-documented, stable features in training data |
| All these tools do pre-publish validation | HIGH | Standard practice, documented in official sources |
| None of these tools have confirmation steps | HIGH | Direct CLI experience, not a contested claim |
| brew tap analogy is structurally accurate | HIGH | Homebrew architecture is well-understood |
| Claude Code plugin format (plugin.json / marketplace.json) | MEDIUM | From PROJECT.md context provided. Not independently verified via current docs. |
| `claude plugin install` is the install command | MEDIUM | From PROJECT.md. Not verified against current Claude Code docs. |
| Plugin cache at `~/.claude/plugins/cache` | MEDIUM | From PROJECT.md. Should be verified before implementation. |

---

## Sources

- Analysis based on established patterns from: npm CLI documentation, VS Code extension publishing docs (vsce), Cargo publish documentation, Homebrew tap architecture
- Claude Code plugin system details sourced from: `.planning/PROJECT.md` (project context provided by user)
- No live web sources available during this research session (WebSearch and WebFetch unavailable)
- Confidence levels reflect training data recency (knowledge cutoff August 2025) for general patterns, and MEDIUM for Claude Code-specific claims that should be verified against current docs before implementation

---

*Feature research for: Chaparral marketplace bridge — local-to-registry publishing for Claude Code plugins*
*Researched: 2026-02-24*
