# chaparral

The connective tissue between your projects.

Chaparral manages shared [Claude Code](https://claude.ai/code) skills and brand identity across every repo in an organization. One brand repo holds the truth. Chaparral links it everywhere it needs to go.

Built for people who work across multiple orgs and clients. Multi-org by default.

## The problem

You have an org directory — say `~/code/manzanita-research/` — with a dozen repos inside. They share a design language, a brand voice, a set of Claude Code skills that shape how AI works with your code. But Claude Code discovers skills per-project (`.claude/skills/`) or globally (`~/.claude/skills/`). There's no org level.

So you copy things around. You forget. Projects drift. The new repo never gets the frontend design skill. The brand voice doc is three versions behind in half your projects.

Claude Code has a plugin marketplace now, but it solves a different problem — distribution to other people and machines. Plugins copy to a cache, require per-repo configuration, and need manual updates after changes. When you're actively developing shared skills across sibling repos, you need something faster.

## The fix

Put a `chaparral.json` manifest in your brand repo. Chaparral finds it, reads it, and symlinks your shared skills and org-level `CLAUDE.md` into every sibling repo. Edit a skill in the brand repo and it's instantly live everywhere. No reinstalls, no cache invalidation, no per-repo setup.

```
~/code/manzanita-research/
├── CLAUDE.md                    ← symlink, managed by chaparral
├── brand/
│   ├── chaparral.json           ← manifest
│   ├── org/
│   │   ├── CLAUDE.md            ← org-wide Claude instructions
│   │   └── skills/
│   │       ├── frontend-design/
│   │       └── brand-voice/
│   └── ...
├── toyon/
│   └── .claude/skills/
│       ├── frontend-design/     ← symlink
│       └── brand-voice/         ← symlink
├── ceanothus/
│   └── .claude/skills/
│       ├── frontend-design/     ← symlink
│       └── brand-voice/         ← symlink
└── ...
```

## Install

```bash
go install github.com/manzanita-research/chaparral/cmd/chaparral@latest
```

Or build from source:

```bash
git clone https://github.com/manzanita-research/chaparral.git
cd chaparral
go build -o chaparral ./cmd/chaparral
```

## Usage

### TUI dashboard

```bash
chaparral
```

Launch the interactive dashboard. Toggle between skills view and repos view with `tab`. See all your orgs at a glance, sync interactively, check link health.

### Sync everything

```bash
chaparral sync
```

Discovers all org directories, finds brand repos (by `chaparral.json`), and links skills into every sibling. Idempotent — safe to run anytime.

### Check status

```bash
chaparral status
```

Shows what's linked, what's stale, what's new and unlinked.

### Clean up

```bash
chaparral unlink
```

Removes all chaparral-managed symlinks. Only touches symlinks it created.

## The manifest

Your brand repo needs a `chaparral.json` at its root:

```json
{
  "org": "manzanita-research",
  "claude_md": "org/CLAUDE.md",
  "skills_dir": "org/skills",
  "exclude": ["brand"]
}
```

| Field | What it does |
|-------|-------------|
| `org` | Human-readable org name (for display) |
| `claude_md` | Path to the org-level CLAUDE.md, relative to brand repo root |
| `skills_dir` | Directory containing shared skills, relative to brand repo root |
| `exclude` | Repos to skip when linking (the brand repo itself, forks, archives) |

## How discovery works

Chaparral looks for org directories in `~/code/`. Any subdirectory that contains a repo with a `chaparral.json` is treated as an org. This means you can manage multiple orgs — different clients, different brands, all from one tool:

```
~/code/
├── manzanita-research/    ← org (brand/ has chaparral.json)
├── temple-of-silicon/     ← org (identity/ has chaparral.json)
├── cosmic-computation-lab/ ← org (brand/ has chaparral.json)
└── personal-projects/     ← not an org (no chaparral.json anywhere)
```

## Local skills vs marketplace plugins

Chaparral and Claude Code's plugin marketplace are complementary:

| | Local skills (chaparral) | Marketplace plugins |
|---|---|---|
| **Propagation** | Symlinks — instant | Cache copies — requires update |
| **Setup** | One manifest, auto-discovered | Per-repo `.claude/settings.json` |
| **Best for** | Active development, fast iteration | Stable distribution to others |
| **Scope** | Org directory | Per-user or per-project |

Chaparral is your workbench. The marketplace is your storefront. Develop locally with symlinks, publish via marketplace when stable.

## What it doesn't do

- No file copying. Symlinks only. One source of truth.
- No git operations. It doesn't pull, push, or commit anything.
- No global installs into `~/.claude/`. Everything stays org-scoped.
- No magic. It creates symlinks and tells you what it did.

## Named for

The chaparral — dense, fire-adapted brushland that covers California's coastal hills. Manzanita, ceanothus, sage, toyon — they grow together in this ecosystem, their roots intertwined beneath the surface. Different plants, same soil. That's the idea.

---

With love from California.
