# Claude Code — curated subagents and settings

The Claude Code layer of OFFICINA: a deliberately selected set of Claude-specific artifacts —
**subagents** and **settings** — chosen on purpose rather than mirrored wholesale. Part of the
[OFFICINA](../) repository.

## Layout

| Folder | What it holds | Status |
|--------|---------------|--------|
| [`agents/`](agents/) | Subagents — specialized agents for delegated tasks | 🚧 Growing |
| [`settings/`](settings/) | Sanitized global config — `CLAUDE.md` examples, `settings.json`, status line | 🚧 Growing |

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

**Keywords:** Claude Code subagents, agents, settings, CLAUDE.md, status line, AI coding workflow, developer methodology.
**Topics:** `claude-code` · `subagents` · `developer-tools`
