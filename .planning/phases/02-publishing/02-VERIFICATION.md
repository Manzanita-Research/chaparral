---
phase: 02-publishing
status: passed
verified: 2026-02-24
---

# Phase 2: Publishing - Verification

## Phase Goal
**Users can write generated manifests to the brand repo and push them to GitHub after an explicit two-step confirmation**

## Success Criteria Verification

### 1. User can see a diff of exactly which files will change before committing to publish
**Status: PASSED**
- `publisher.DiffManifests()` returns `[]FileChange` with Kind: "new", "modified", or "unchanged"
- `runPublishFull()` and `runPublishWriteOnly()` display changes with +/~ prefix before any writes
- Verified: `DiffManifests` tested in `TestDiffManifests_NewFiles`, `TestDiffManifests_ModifiedFiles`, `TestDiffManifests_UnchangedFiles`

### 2. User can write the `.claude-plugin/` directory structure to the brand repo without triggering a push
**Status: PASSED**
- `chaparral publish --write-only` calls `WriteManifests` without any git operations
- `WriteManifests` creates `.claude-plugin/marketplace.json` and per-skill `plugin.json` files
- Verified: `TestWriteManifests_NewFiles`, `TestWriteManifests_ExistingFiles`, `TestWriteManifests_VersionBump`

### 3. User can push the marketplace to GitHub only after a two-step confirmation that shows version, target remote, and changed files
**Status: PASSED**
- `runPublishFull()` shows diff preview, then version and remote URL, then two `confirm()` calls
- First prompt: "Push to GitHub? [y/N]"
- Second prompt: "Confirm push (this will update the live marketplace): [y/N]"
- `CommitAndPush` stages files via go-git, commits with version message, pushes with GITHUB_TOKEN
- Verified: code inspection of `runPublishFull()` in main.go lines 380-457

### 4. User can check whether local skills are newer than the published marketplace before deciding to republish
**Status: PASSED**
- `chaparral publish --check` calls `CheckFreshness` and shows per-skill freshness
- Mtime comparison with 1-second tolerance + content comparison fallback for post-clone
- Live test: `./chaparral publish --check` correctly shows 5 stale skills
- Verified: `TestCheckFreshness_NeverPublished`, `TestCheckFreshness_Stale`, `TestCheckFreshness_Fresh`

## Requirement Traceability

| Requirement | Plan | Status | Evidence |
|-------------|------|--------|----------|
| PUB-01 | 02-01, 02-02 | Complete | WriteManifests creates .claude-plugin/ + per-skill plugin.json |
| PUB-02 | 02-01, 02-02 | Complete | DiffManifests returns new/modified/unchanged with content |
| PUB-03 | 02-02 | Complete | CommitAndPush + two-step confirm in runPublishFull |
| PUB-04 | 02-01, 02-02 | Complete | CheckFreshness via mtime + content fallback, --check flag |

## Must-Have Verification

### Plan 02-01 Must-Haves
- [x] WriteManifests creates .claude-plugin/marketplace.json and per-skill plugin.json files
- [x] DiffManifests returns new/modified/unchanged status without writing
- [x] CheckFreshness detects when skill source files are newer than published plugin.json
- [x] Version bumps patch number when existing plugin.json has a version
- [x] Version defaults to 0.1.0 when no published plugin.json exists

### Plan 02-02 Must-Haves
- [x] User can run `chaparral publish --check` to see staleness report
- [x] User can run `chaparral publish --write-only` to write manifests without pushing
- [x] User can run `chaparral publish` and see diff preview, then two-step confirmation before push
- [x] Push requires GITHUB_TOKEN env var and shows clear message if missing
- [x] Second publish with no changes prints "already up to date" instead of erroring

## Artifact Verification

| Artifact | Min Lines | Actual | Status |
|----------|-----------|--------|--------|
| internal/publisher/publisher.go | 100 | 348 | PASSED |
| internal/publisher/publisher_test.go | 80 | 406 | PASSED |
| internal/publisher/git.go | 40 | 94 | PASSED |
| internal/publisher/git_test.go | 30 | 128 | PASSED |
| cmd/chaparral/main.go (contains runPublish) | - | Yes | PASSED |

## Key Links Verification

| From | To | Pattern | Found |
|------|----|---------|-------|
| publisher.go | generator.go | `generator\.Generate` | Yes |
| publisher.go | config.go | `config\.(Org\|Skill)` | Yes |
| main.go | publisher.go | `publisher\.(Check\|Diff\|Write)` | Yes |
| main.go | git.go | `publisher\.CommitAndPush` | Yes |
| git.go | go-git/v5 | `git\.Plain` | Yes |

## Test Results

```
16 tests pass:
- 3 bumpVersion tests
- 4 WriteManifests tests
- 3 DiffManifests tests
- 3 CheckFreshness tests
- 3 git tests (no-changes, staging+commit, missing remote)
```

All existing tests continue to pass (generator, skillmeta, validator).

## Self-Check: PASSED

All success criteria met. All requirements covered. All must-haves verified. All artifacts meet minimum line counts. All key links present.
