# Skills — portable modes and procedures

Reusable **skills**: self-contained modes and procedures a coding agent invokes mid-session. Each is a
folder with a `SKILL.md` — plain Markdown plus YAML frontmatter — so it isn't tied to any one runtime.
It runs in Claude Code and any harness that reads the `SKILL.md` format. Part of the
[OFFICINA](../) repository.

| Skill | Invoke | What it does |
|-------|--------|--------------|
| [`committing-changes`](committing-changes/SKILL.md) | auto — before any commit | Conventional Commits procedure: reads project/global conventions (`CLAUDE.md` or `AGENTS.md`), scans staged files for secrets and garbage, proposes atomic splits, drafts a `type(scope): subject` message, and waits for approval before committing. Never pushes or amends unasked. |
| [`discuss`](discuss/SKILL.md) | `/discuss` (manual only) | Persistent, read-only thinking-partner mode for Q&A, ideation, analysis, and design debate — concise, one question at a time, argues its own view, proposes 3+ alternatives, writes no files until you exit. |
| [`merging-branches`](merging-branches/SKILL.md) | auto — before merge/rebase | Safe branch integration: determines strategy from the project conventions file (`CLAUDE.md` or `AGENTS.md`), runs pre-merge checks (clean tree, up-to-date with remote, version-aware conflict pre-scan), and requires a typed phrase to touch protected branches. Never pushes or deletes branches unasked. |

The two git skills chain: on a squash merge, `merging-branches` hands off to `committing-changes`
to produce the single commit — one message convention, one safety scan, no duplication.

## Using a skill

Each skill lives in its own folder as a `SKILL.md` with YAML frontmatter (`name`, `description`,
and — for manual-only skills — `disable-model-invocation: true`). Drop the folder into your agent's
skills directory to make it available:

- **Claude Code** — `~/.claude/skills/<name>/` (global) or a project's `.claude/skills/<name>/`.
- **Other harnesses** — the equivalent skills location for your runtime; the `SKILL.md` content is
  harness-neutral.

---

**Keywords:** agent skills, SKILL.md, reusable skills, discussion mode, git, conventional commits, merge, AI coding workflow.
**Topics:** `agent-skills` · `claude-code-skills` · `developer-tools`
