# Claude Code — curated skills, subagents, and settings

The core of OFFICINA: a deliberately selected set of Claude Code artifacts — **skills** (reusable
modes and procedures), **subagents**, and **settings** — chosen on purpose rather than
mirrored wholesale. Part of the [OFFICINA](../) repository.

## Layout

| Folder | What it holds | Status |
|--------|---------------|--------|
| [`skills/`](skills/) | Skills — reusable modes and procedures invoked in a session | 🚧 Growing |
| [`agents/`](agents/) | Subagents — specialized agents for delegated tasks | 🚧 Growing |
| [`settings/`](settings/) | Sanitized global config — `CLAUDE.md` examples, `settings.json`, status line | 🚧 Growing |

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

## Agents

Specialized subagents you delegate a scoped task to — see [`agents/README.md`](agents/README.md) for
per-agent notes.

| Agent | Model | What it does |
|-------|-------|--------------|
| [`ai-dev`](agents/ai-dev.md) | `opus` | Agentic-systems architect: designs, builds, and ships production AI agents across the full stack — orchestration, tools, RAG, memory, inference, runtime, evals/observability, governance — with cost, latency, and security treated as first-class. Blunt, trade-off-driven, evals-before-deploy. |

Each agent is a single Markdown file with YAML frontmatter (`name`, `description`, `model`). Drop it
into `~/.claude/agents/` (or a project's `.claude/agents/`) to make it available.

## Settings

The global configuration layer — see [`settings/README.md`](settings/README.md) for per-file notes.

| File | Layer | What it is |
|------|-------|------------|
| [`CLAUDE.global.md`](settings/CLAUDE.global.md) | Global | Cross-project preferences loaded in every session |
| [`CLAUDE.project.md`](settings/CLAUDE.project.md) | Project | Reusable template for a project's root `CLAUDE.md` |
| [`settings.json`](settings/settings.json) | Global | Permissions, status line, plugins, effort level |
| [`settings.local.json`](settings/settings.local.json) | Project-local | Example of machine-/repo-local overrides |
| [`statusline-command.sh`](settings/statusline-command.sh) | Global | Custom status line (model, context, rate limits, branch) |

---

**Keywords:** Claude Code skills, subagents, agent skills, SKILL.md, discussion mode, git, conventional commits, merge, AI coding workflow, developer methodology.
**Topics:** `claude-code` · `claude-code-skills` · `subagents` · `developer-tools`
