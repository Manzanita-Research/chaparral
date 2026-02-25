# CLAUDE.md

Chaparral is a brand identity sync tool. It symlinks shared Claude Code skills and CLAUDE.md files from a brand repo into every sibling repo in an org, so you edit once and propagate instantly. It also shows marketplace plugin status alongside local symlinks. It's multi-org — discovery scans `~/code/` and finds every org with a `chaparral.json` manifest.

## Domain Vocabulary

- **Org directory**: Parent dir (e.g. `~/code/manzanita-research/`) containing multiple repos
- **Brand repo**: The repo with the `chaparral.json` manifest — source of truth for shared skills and CLAUDE.md
- **Manifest** (`chaparral.json`): Declares which skills to share and the org-level CLAUDE.md path
- **Sibling repos**: Other repos in the same org directory that receive symlinked skills
- **Local skills** vs **marketplace plugins**: Local = symlinked from brand repo, instant, for active development. Marketplace = installed via Claude Code's plugin system, cached copies, for stable distribution. Chaparral bridges both — it can generate `marketplace.json` from local skills when they're ready to publish.

## Conventions

- Symlink operations must be idempotent and safe — never overwrite non-symlink files.
- All filesystem operations should handle missing directories gracefully.
- Error messages should be helpful and plain-spoken. No "Error:" prefix.
- Keep the TUI warm and readable. Lip Gloss styles should feel human, not corporate.
- Cap content width at 80 characters. Use Lip Gloss padding, not manual string prefixes.
- Respect `NO_COLOR` environment variable.
