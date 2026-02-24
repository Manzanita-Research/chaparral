# External Integrations

**Analysis Date:** 2026-02-24

## APIs & External Services

**Claude Code:**
- **Claude Code skills system** - Integration point for distributed skills via marketplace
  - Implementation: Reads and symlinks `.claude/skills/` directories
  - Files: `internal/config/config.go`, `internal/linker/linker.go`
  - Purpose: Manages local skill symlinks and tracks marketplace plugin status (planned feature)

## Data Storage

**Databases:**
- None - This is a CLI tool that operates on the filesystem only

**File Storage:**
- Local filesystem only
  - JSON manifest files: `chaparral.json` (org configuration)
  - CLAUDE.md files (org-wide instructions)
  - Skill directories (SKILL.md format)
  - Symlink operations: Creates symlinks in `{repo}/.claude/skills/`
  - Client: Go stdlib `os`, `os/exec`, `path/filepath`

**Caching:**
- None

## Authentication & Identity

**Auth Provider:**
- None - Local filesystem operations only, no authentication required

**Git Integration:**
- Detection only: Identifies directories as repos by checking for `.git/` directory
- Files: `internal/discovery/discovery.go:isRepo()`
- Purpose: Distinguishes valid repos from other directories during org scanning

## Monitoring & Observability

**Error Tracking:**
- None - Local tool with no telemetry

**Logs:**
- Console output only (stdout/stderr)
- TUI error display in modal windows
- No persistent logging
- Respects `NO_COLOR` environment variable for accessibility

## CI/CD & Deployment

**Hosting:**
- GitHub (source code): `github.com/manzanita-research/chaparral`
- No cloud deployment required

**CI Pipeline:**
- Not detected (no GitHub Actions, GitLab CI, or other automation files)

**Distribution:**
- Go module: `github.com/manzanita-research/chaparral/cmd/chaparral`
- Installation: `go install github.com/manzanita-research/chaparral/cmd/chaparral@latest`

## Environment Configuration

**Required env vars:**
- None - Tool works with defaults (`~/code/` base path)

**Optional env vars:**
- `NO_COLOR` - Disables colored output (standard convention for CLI tools)
  - Checked in `internal/tui/styles.go:hasNoColor()`

**Secrets location:**
- Not applicable - No secrets required for local operation

## Webhooks & Callbacks

**Incoming:**
- None

**Outgoing:**
- None

## Filesystem Conventions

**Organization Structure Expectations:**
- `~/code/{org-name}/` - Organization directory
- `~/code/{org-name}/{brand-repo}/chaparral.json` - Brand repo with manifest
- `~/code/{org-name}/{brand-repo}/{skills_dir}/` - Shared skills directory
- `~/code/{org-name}/{brand-repo}/{claude_md}` - Org-level CLAUDE.md file
- `~/code/{org-name}/{sibling-repos}/.claude/skills/` - Symlink targets

**Manifest Format:**
```json
{
  "org": "organization-name",
  "claude_md": "path/to/CLAUDE.md",
  "skills_dir": "path/to/skills",
  "exclude": ["repo-name-to-skip"]
}
```

---

*Integration audit: 2026-02-24*
