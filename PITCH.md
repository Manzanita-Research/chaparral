# Chaparral

Chaparral keeps your shared Claude Code skills and brand identity alive across every repo in an organization. One brand repo holds the truth. Chaparral symlinks it everywhere else. Edit a skill once, and it's instantly live in every sibling project — no copies, no drift, no ceremony.

## The problem

If you work across multiple repos that share an identity — a design language, a set of AI coding skills, a voice — you know what happens. You update the brand voice skill in one place and forget to propagate. You add a new slash command and three repos never get it. The org-level CLAUDE.md is four versions behind in half your projects.

Claude Code discovers skills per-project (`.claude/skills/`) or globally (`~/.claude/skills/`). There's no org level. So the connective tissue between your projects becomes copy-paste and good intentions.

Claude Code has a plugin marketplace now, and it's useful for distribution — getting stable skills to other people and machines. But marketplace plugins copy to a cache, require per-repo configuration, and need manual updates after changes. When you're actively iterating on shared skills — tuning a design system, writing the slash commands your team actually uses — you need something that moves at the speed of a file save.

## What we're building

A small CLI with a warm TUI dashboard. You put a `chaparral.json` manifest in your brand repo — it declares which skills to share and where your org-level CLAUDE.md lives. Run `chaparral sync` and every sibling repo gets symlinks to those shared assets. New repo appears? Sync once and it's connected. New skill? Same.

The dashboard shows you everything at a glance. Two views — skills (which skills exist and how many repos have them) and repos (which repos exist and which skills they carry). Navigate with `j`/`k`, sync one org or all of them, check health, move on.

It's multi-org by default. If you're a studio, an agency, a freelancer juggling clients — each org gets its own brand repo, its own manifest, its own identity. Chaparral scans your code directory, finds every org with a manifest, and manages them all from one place.

The feeling is: you open the dashboard, glance at the state of things, hit `s` to sync, and close it. Thirty seconds. Everything's connected. Back to work.

## How it works

Symlinks. That's the key decision. No copies, no caches, no sync engines. A symlink means the skill file in your brand repo *is* the skill file in every sibling repo. One source of truth, rooted in version control where it belongs. Edit it and the change is live before you switch terminals.

Chaparral walks your base directory (defaults to `~/code/`), finds every subdirectory that contains a repo with a `chaparral.json`, and treats it as an org. It reads the manifest, discovers sibling repos, creates `.claude/skills/` directories where needed, and symlinks each shared skill folder. It also links the org-level CLAUDE.md into the parent directory so every project inherits it.

All operations are idempotent. Existing symlinks are left alone. Non-symlink files are never overwritten. Running sync twice does nothing the second time.

## Who it's for

Someone who maintains multiple related repos under one org — a creative studio with a dozen projects, an agency with separate client codebases, a solo developer with a growing constellation of tools. They use Claude Code seriously. They've built custom skills for their design language, their coding conventions, their brand voice. And they're tired of those skills living in one project while the others fall out of sync.

More specifically: the person who just created a new repo, opened Claude Code, and realized none of their shared context came with them. Chaparral is the answer to that moment.

## Why us

We built this because we needed it. Manzanita Research has a growing family of projects — Audiotree, Toneword, Glory — each with its own purpose, all needing to speak the same language, carry the same design sensibility, use the same skills when Claude Code works with the code. We were the person in the paragraph above, creating new repos and losing our shared context every time.

We also care about the shape of the tool. Chaparral isn't a platform or a framework. It's a small program that creates symlinks and tells you what it did. That restraint is intentional.

## Where this is going

What works now: the TUI dashboard with skills and repos views, `sync` / `status` / `unlink` commands, multi-org discovery, manifest-driven configuration. It's functional and we use it daily.

What's next: marketplace bridge — the ability to generate a `marketplace.json` from a brand repo, turning local skills into publishable Claude Code plugins when they're stable. This closes the loop between local development and distribution. We want the TUI to show both symlink status and installed marketplace plugin status in one view.

What's still a question: how far to take the dashboard. Right now it's a status-and-sync tool. It could grow into something that helps you author and test skills, visualize dependencies between them, or manage skill versions. We're not sure yet. We're using it and seeing what we reach for.

## The name

Chaparral — the dense, fire-adapted brushland that covers California's coastal hills. Manzanita, ceanothus, sage, toyon — they grow together in this ecosystem, their roots intertwined beneath the surface. Different plants, same soil. That's the idea.

---

With love from California.
