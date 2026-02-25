# Requirements: Chaparral — Marketplace Bridge

**Defined:** 2026-02-24
**Core Value:** One place to see and manage both local skills and marketplace plugins across all your orgs — plus the bridge that turns local iteration into published distribution.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Generation

- [x] **GEN-01**: User can validate brand repo skills are correctly structured before packaging
- [x] **GEN-02**: User can generate a `plugin.json` manifest for each skill in the brand repo
- [x] **GEN-03**: User can generate a `marketplace.json` catalog listing all skills as plugins
- [x] **GEN-04**: Each publish stamps a semantic version on the generated plugin manifests

### Publishing

- [ ] **PUB-01**: User can write `.claude-plugin/` directory structure to the brand repo
- [ ] **PUB-02**: User sees a diff preview of what will be published before confirming
- [ ] **PUB-03**: User can push the marketplace to GitHub after two-step confirmation
- [ ] **PUB-04**: User can check if local skills are newer than the published marketplace

### Discovery

- [ ] **DISC-01**: Chaparral scans sibling repos for installed marketplace plugins
- [ ] **DISC-02**: Chaparral queries the published marketplace to show available plugins
- [ ] **DISC-03**: TUI shows marketplace plugin install status alongside local skill link status
- [ ] **DISC-04**: User can install marketplace plugins into sibling repos from the TUI

## v2 Requirements

### Scaffolding

- **INIT-01**: `chaparral init` scaffolds a new brand repo with chaparral.json

### Visibility

- **VIS-01**: Private/public marketplace toggle per brand repo
- **SYNC-01**: Auto-detect when published marketplace is out of date on TUI launch

## Out of Scope

| Feature | Reason |
|---------|--------|
| Custom plugin format | We generate Claude Code's native format, don't invent our own |
| npm/pip plugin sources | GitHub-hosted marketplaces only for now |
| Hosted registry service | GitHub repos are the registry — no custom hosting |
| Plugin authoring tools | We package existing skills, don't help write new ones |
| Auto-update management | Claude Code handles plugin updates natively |
| Multi-marketplace per brand repo | One brand repo = one marketplace; multiple brand repos per org is supported |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| GEN-01 | Phase 1 | Complete |
| GEN-02 | Phase 1 | Complete |
| GEN-03 | Phase 1 | Complete |
| GEN-04 | Phase 1 | Complete |
| PUB-01 | Phase 2 | Pending |
| PUB-02 | Phase 2 | Pending |
| PUB-03 | Phase 2 | Pending |
| PUB-04 | Phase 2 | Pending |
| DISC-01 | Phase 3 | Pending |
| DISC-02 | Phase 3 | Pending |
| DISC-03 | Phase 3 | Pending |
| DISC-04 | Phase 3 | Pending |

**Coverage:**
- v1 requirements: 12 total
- Mapped to phases: 12
- Unmapped: 0

---
*Requirements defined: 2026-02-24*
*Last updated: 2026-02-24 — traceability filled after roadmap creation*
