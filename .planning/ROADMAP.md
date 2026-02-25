# Roadmap: Chaparral — Marketplace Bridge

## Overview

The existing Chaparral tool handles local skill syncing. This milestone adds the bridge: generate Claude Code-native plugin manifests from brand repo skills, publish them to GitHub as a marketplace, then scan and install from that marketplace — all from the same dashboard. Three phases, each delivering a coherent, independently verifiable capability.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Generation** - Users can validate and generate Claude Code plugin manifests from brand repo skills
- [x] **Phase 2: Publishing** - Users can write and push a versioned marketplace to GitHub with explicit confirmation
- [x] **Phase 3: Discovery and Install** - Users can see marketplace plugin status and install plugins from the TUI
- [x] **Phase 4: Polish** - Fix display bugs and UX issues found during milestone audit (completed 2026-02-25)

## Phase Details

### Phase 1: Generation
**Goal**: Users can validate brand repo skills and generate Claude Code-native plugin manifests locally, without any network calls or filesystem side effects
**Depends on**: Nothing (first phase)
**Requirements**: GEN-01, GEN-02, GEN-03, GEN-04
**Success Criteria** (what must be TRUE):
  1. User can run `chaparral validate` and see which skills are correctly structured and which have problems
  2. User can generate a `plugin.json` manifest for each skill and see the output before any files are written
  3. User can generate a `marketplace.json` catalog listing all skills, viewable as a dry run to stdout
  4. Each generated manifest includes a semantic version that increments on publish
**Plans**: 2 plans
- [x] 01-01-PLAN.md — SKILL.md frontmatter parsing, skill validation, and `chaparral validate` command
- [x] 01-02-PLAN.md — Plugin and marketplace manifest generation and `chaparral generate` command

### Phase 2: Publishing
**Goal**: Users can write generated manifests to the brand repo and push them to GitHub after an explicit two-step confirmation
**Depends on**: Phase 1
**Requirements**: PUB-01, PUB-02, PUB-03, PUB-04
**Success Criteria** (what must be TRUE):
  1. User can see a diff of exactly which files will change before committing to publish
  2. User can write the `.claude-plugin/` directory structure to the brand repo without triggering a push
  3. User can push the marketplace to GitHub only after a two-step confirmation that shows version, target remote, and changed files
  4. User can check whether local skills are newer than the published marketplace before deciding to republish
**Plans**: 2 plans
- [x] 02-01-PLAN.md — Publisher package: write manifests, diff preview, staleness check, version bumping (TDD)
- [x] 02-02-PLAN.md — go-git integration and `chaparral publish` command with three modes

### Phase 3: Discovery and Install
**Goal**: Users can see which marketplace plugins are installed across sibling repos and install plugins directly from the TUI
**Depends on**: Phase 2
**Requirements**: DISC-01, DISC-02, DISC-03, DISC-04
**Success Criteria** (what must be TRUE):
  1. The TUI shows which marketplace plugins each sibling repo has installed, alongside local skill link status
  2. User can see which plugins are available in the published marketplace from within the dashboard
  3. User can install a marketplace plugin into a sibling repo from the TUI, with full error output if the install fails
**Plans**: 2 plans
- [x] 03-01-PLAN.md — Marketplace package: types, scanner, installer, and CLI status integration (TDD)
- [x] 03-02-PLAN.md — TUI integration: plugin display in repos tab and interactive install flow

### Phase 4: Polish
**Goal**: Fix display bugs and UX issues found during milestone audit
**Depends on**: Phase 3
**Requirements**: DISC-03, DISC-04, GEN-04 (gap closure)
**Gap Closure:** Closes gaps from v1-MILESTONE-AUDIT.md
**Success Criteria** (what must be TRUE):
  1. Installed plugins show correct enabled/disabled icon in TUI (not always disabled)
  2. TUI install targets the repo the user is browsing, not always the first repo
  3. `chaparral generate` shows the next version that would be published, not always 0.1.0
**Plans**: 1 plan
- [ ] 04-01-PLAN.md — Fix enabled icon, repo cursor targeting, and version-aware generate preview

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Generation | 2/2 | Complete | 2026-02-24 |
| 2. Publishing | 2/2 | Complete | 2026-02-24 |
| 3. Discovery and Install | 2/2 | Complete | 2026-02-25 |
| 4. Polish | 0/1 | Complete    | 2026-02-25 |
