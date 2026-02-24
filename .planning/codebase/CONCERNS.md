# Codebase Concerns

**Analysis Date:** 2026-02-24

## File Operation Error Handling

**Silent `os.Remove()` failures:**
- Issue: `os.Remove()` calls are executed without error checking in cleanup operations
- Files: `internal/linker/linker.go` (lines 52, 68, 150)
- Impact: Failed removal of stale symlinks won't be reported. Users may think a link was updated when it actually failed to be removed first, leaving conflicting files on disk
- Fix approach: Check and return errors from `os.Remove()` calls. Include removal errors in `LinkResult` structs or return them explicitly

**`StatusOrg()` error silently ignored in TUI initialization:**
- Issue: `linker.StatusOrg(org)` error is discarded with blank `_` assignment during TUI init
- Files: `internal/tui/tui.go` (line 87)
- Impact: If status check fails (permissions, missing files), the TUI will display incomplete/stale data without warning the user. Sync operations may then act on incomplete state
- Fix approach: Propagate the error and either display it or retry. At minimum, log it or set a "status unknown" state

## Bounds Checking Gaps

**Potential cursor overflow in tabbed view:**
- Issue: When switching tabs with `tab` key, cursor is reset to 0. If the repos tab has fewer items than the skills tab, users cannot navigate properly if they had cursor > items in new tab
- Files: `internal/tui/tui.go` (lines 123-131)
- Impact: Cursor position becomes incoherent when switching between views with different numbers of items
- Fix approach: When switching tabs, clamp cursor to `min(cursor, len(items)-1)` for the new tab rather than always resetting to 0

## Path Construction Edge Cases

**Non-absolute base path handling:**
- Issue: Default base path is derived from `os.UserHomeDir()` but CLI can accept custom `basePath`. If custom path is relative (not absolute), downstream `filepath.Join()` calls will produce relative paths used in symlink targets
- Files: `cmd/chaparral/main.go` (lines 16-49), `internal/tui/tui.go` (lines 538-545)
- Impact: Symlinks created with relative paths may break when working directory changes. Expected absolute paths would be safer
- Fix approach: Normalize custom basePath to absolute before passing to discovery and linker functions. Use `filepath.Abs()` early

**Skill path validation missing:**
- Issue: Skills directory must contain `SKILL.md` to be recognized, but there's no validation that this is actually a file (not a directory with that name)
- Files: `internal/discovery/discovery.go` (lines 111-117)
- Impact: A symlink named `SKILL.md` in a directory would pass validation. Unlikely but not impossible after manual filesystem edits
- Fix approach: Add `!info.IsDir()` check when validating SKILL.md

## Race Conditions in Concurrent Operations

**Symlink state can change between check and create:**
- Issue: `createSymlink()` checks for existing destination, then creates. Between `os.Lstat()` and `os.Symlink()`, another process could modify the destination
- Files: `internal/linker/linker.go` (lines 140-168)
- Impact: In shared team environments or CI pipelines running chaparral simultaneously, a stale symlink could be removed by one process while another is trying to create it, causing "file exists" errors or broken links
- Fix approach: Add retry logic with a small backoff. Consider file locking for critical sections. Document that chaparral is not safe for concurrent orgs execution

## Missing Manifest Validation

**No validation of manifest.Exclude entries:**
- Issue: `Exclude` list in chaparral.json is used directly without validating that excluded repos actually exist or that the brand repo is in the list
- Files: `internal/config/config.go` (lines 59-66), `internal/discovery/discovery.go` (lines 74-79)
- Impact: Typos in exclude list will silently not match, potentially linking to wrong repos. Missing the brand repo from exclude list will create symlinks inside the brand repo itself
- Fix approach: Validate manifest.Exclude against discovered repos. At minimum, warn if an excluded repo is never found

**No validation that skills_dir and claude_md exist:**
- Issue: Manifest paths are used directly without existence checks until operations fail
- Files: `internal/linker/linker.go` (lines 29-32), elsewhere
- Impact: User won't know about missing manifest paths until trying to sync. Error message will be generic "finding skills" error
- Fix approach: Validate paths when loading manifest. Return clear "source skill directory not found at [path]" errors

## Fragile Test Dependency Path

**Repository detection relies solely on .git directory:**
- Issue: `isRepo()` checks for `.git/` to identify repos. Worktrees, shallow clones, or bare repos may not have this structure
- Files: `internal/discovery/discovery.go` (lines 122-128)
- Impact: Git worktrees or repositories in unusual configurations will be skipped. Similarly, cloned repos without .git (e.g., exported archives) will be treated as non-repos
- Fix approach: Make repo detection configurable. Accept a repo manifest file (e.g., `.chaparral-link`) or check for multiple git indicators

## TUI State Mutation After Errors

**Error state persists across refresh:**
- Issue: When `orgsLoaded` message sets `m.err`, that error is rendered but the view changes to `viewDashboard`. Refreshing with `r` calls `Init()` but doesn't clear previous error
- Files: `internal/tui/tui.go` (lines 167-171, 152-154)
- Impact: Old error message may flash when dashboard is shown after a failed operation. Confusing UX if user takes corrective action and retries
- Fix approach: Clear `m.err` when transitioning between views or receiving new status

## Incomplete Cleanup on Partial Failures

**`UnlinkOrg()` continues after `FindSkills()` fails:**
- Issue: If finding skills fails partway through, `UnlinkOrg` returns the partial results without unlinking. Next unlink call will detect the same skills again
- Files: `internal/linker/linker.go` (lines 59-62)
- Impact: If a skills directory becomes inaccessible mid-operation, some symlinks won't be unlinked. Re-running unlink won't know what to remove because skill list is incomplete
- Fix approach: Store the list of linked repos somewhere (e.g., checking which `.claude/skills/` dirs exist directly) to make unlink more resilient. Or fail fast with clear error

## Symlink Target Format Assumptions

**Readlink comparison uses string equality, not path equivalence:**
- Issue: `os.Readlink()` returns the target path as stored in the symlink. Comparison with `expectedTarget` is string-based
- Files: `internal/linker/linker.go` (lines 183-192)
- Impact: If paths differ only in `.` or `..` resolution, or symlink was created with relative path, they won't match even though they point to the same file. Will incorrectly report "stale"
- Fix approach: Resolve both paths to absolute canonical form before comparison using `filepath.EvalSymlinks()` or similar

## Discovery Performance on Large Orgs

**Recursive symlink resolution not prevented:**
- Issue: `FindOrgs` walks org directories with `ReadDir`, which will follow symlinks if present. A symlink loop won't be detected
- Files: `internal/discovery/discovery.go` (lines 15-38)
- Impact: On large org directories with symlinks back to parent dirs, discovery could hang or exhaust memory
- Fix approach: Track visited inodes, or add a max depth limit to directory walking

**No caching of manifest between discovery calls:**
- Issue: Every TUI refresh re-discovers all orgs and re-reads all manifests and link statuses
- Files: `internal/tui/tui.go` (lines 76-94), entire discovery flow
- Impact: With many repos and skills, refresh (`r` key) will be slow. TUI responsiveness decreases with org size
- Fix approach: Implement simple file-mtime-based cache. Only re-scan if manifest or skills dir changed. Could be opt-in for safety

## Missing Input Validation

**No CLI argument count validation beyond first arg:**
- Issue: Commands like `sync`, `status`, `unlink` don't validate that no extra arguments are passed
- Files: `cmd/chaparral/main.go` (lines 27-40)
- Impact: `chaparral sync --help` will silently be treated as `chaparral sync`. User might think a flag was accepted when it wasn't
- Fix approach: Validate `os.Args` length matches expected count per command. Return helpful error if extra args present

## Security Considerations

**Symlink source paths not validated:**
- Issue: Symlink targets in `createSymlink()` use paths directly from config without checking they're within expected boundaries
- Files: `internal/linker/linker.go` (lines 125-136, 138-168)
- Impact: Malicious `chaparral.json` manifest could create symlinks to files outside the brand repo (e.g., `/etc/passwd`). Limited impact since only affects user's own org dir, but worth noting
- Fix approach: Validate that skill paths are within `skillsDir`. Use `filepath.Abs()` and `filepath.Rel()` with bounds checking

**No verification of symlink targets before creation:**
- Issue: `createSymlink()` doesn't verify the source actually exists as a valid skill before creating the link
- Files: `internal/linker/linker.go` (lines 125-136)
- Impact: Can create broken symlinks if skill was moved/deleted between skill discovery and link creation
- Fix approach: Re-verify skill path exists immediately before creating symlink

## Platform-Specific Issues

**Windows symlink support:**
- Issue: Code uses `os.Symlink()` which on Windows requires admin privileges or Developer Mode
- Files: `internal/linker/linker.go` (entire file)
- Impact: Windows users without proper permissions will get cryptic errors. No graceful degradation
- Fix approach: Detect Windows and either warn upfront or fall back to hard links / file copying. Document Windows limitations clearly

---

*Concerns audit: 2026-02-24*
