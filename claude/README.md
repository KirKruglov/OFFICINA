# Claude Code — curated skills, subagents, and settings

The core of OFFICINA: a deliberately selected set of Claude Code artifacts — **skills** (reusable
modes and procedures), **subagents**, and sanitized **settings** — chosen on purpose rather than
mirrored wholesale. Part of the [OFFICINA](../) repository.

Artifacts are exported from a private source of truth: copy + sanitization (no absolute paths, no
secrets, no private Claude state) on each one. See the repo [README](../README.md) for the model.

## Layout

| Folder | What it holds | Status |
|--------|---------------|--------|
| [`skills/`](skills/) | Skills — reusable modes and procedures invoked in a session | 🚧 Growing |
| [`agents/`](agents/) | Subagents — specialized agents for delegated tasks | 🚧 Pending |
| [`settings/`](settings/) | Sanitized global `CLAUDE.md` and `settings.json` | 🚧 Pending |

## Skills

| Skill | Invoke | What it does |
|-------|--------|--------------|
| [`committing-changes`](skills/committing-changes/SKILL.md) | auto — before any commit | Conventional Commits procedure: reads project/global `CLAUDE.md` conventions, scans staged files for secrets and garbage, proposes atomic splits, drafts a `type(scope): subject` message, and waits for approval before committing. Never pushes or amends unasked. |
| [`discuss`](skills/discuss/SKILL.md) | `/discuss` (manual only) | Persistent, read-only thinking-partner mode for Q&A, ideation, analysis, and design debate — concise, one question at a time, argues its own view, proposes 3+ alternatives, writes no files until you exit. |
| [`merging-branches`](skills/merging-branches/SKILL.md) | auto — before merge/rebase | Safe branch integration: determines strategy from `CLAUDE.md`, runs pre-merge checks (clean tree, up-to-date with remote, version-aware conflict pre-scan), and requires a typed phrase to touch protected branches. Never pushes or deletes branches unasked. |

The two git skills chain: on a squash merge, `merging-branches` hands off to `committing-changes`
to produce the single commit — one message convention, one safety scan, no duplication.

Each skill lives in its own folder as a `SKILL.md` with YAML frontmatter (`name`, `description`,
and — for manual-only skills — `disable-model-invocation: true`). Drop the folder into
`~/.claude/skills/` (or a project's `.claude/skills/`) to use it.

---

**Keywords:** Claude Code skills, subagents, agent skills, SKILL.md, discussion mode, git, conventional commits, merge, AI coding workflow, developer methodology.
**Topics:** `claude-code` · `claude-code-skills` · `subagents` · `developer-tools`
