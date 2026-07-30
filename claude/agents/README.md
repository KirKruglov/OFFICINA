# Agents — subagents (Claude Code-native, portable to Cursor / Codex / Grok)

Subagents you delegate a scoped task to: each runs with its own role, model, and behavioral
contract, and returns a focused result instead of you steering a general model through the whole
job. One of them, `architect`, works the other way round — it owns a session from the root rather
than being dispatched into one. Part of the [OFFICINA](../../) repository.

The canonical form is Claude Code's `.md` (YAML frontmatter + Markdown body). That same file works
**as-is in Cursor 2.0** (it reads `.claude/agents/` directly). Codex and Grok use different formats —
convert the body once (see [Portability](#portability)). These are examples, not a drop-in set — copy
what fits and adapt the persona. No absolute paths or secrets are included.

## Agents

| Agent | Model | What it does |
|-------|-------|--------------|
| [`ai-dev`](ai-dev.md) | `inherit` | Agentic-systems architect and senior staff engineer. Designs, builds, and ships production AI agents across the full stack — surface, orchestration, tools (MCP/A2A), knowledge/RAG, memory, model/inference, runtime/isolation, evals + observability, governance. Thinks in trade-offs, insists on evals before deploy, treats cost/latency/security as first-class, verifies fast-moving facts before asserting them, and pushes back rather than agreeing by default. |
| [`architect`](architect.md) | `inherit` | Software architect persona for a whole session (`claude --agent architect`). Skeptical minimalist: verifies versions and compatibility against primary sources, builds on the existing stack, objects to over-engineering instead of agreeing silently, and insists every decision lands in repository files. Routes each situation to the matching [`arch-*` skill](../../skills/arch/). |
| [`arch-critic-runner`](arch-critic-runner.md) | `inherit` | Fresh-context critic of one architecture decision package. Preloads the `arch-critic` skill, receives the decision without the author's reasoning, and returns a four-axis verdict. Dispatched by `arch-new` and `arch-change`; not for direct use. |

The `model` field is intentionally unset (or `inherit`) — Claude Code and Cursor default to
`inherit`, so the agent adopts the session model. Force one per harness only if you want to (see
[Model per harness](#model-per-harness)).

`architect` and `arch-critic-runner` are two halves of one design. The first owns a session and
never gets delegated to; the second exists only to be delegated to, and its whole value is the
context it does **not** have — a critic that has read the author's justification tends to agree
with it. Both belong to the architecture system documented in
[`skills/arch/`](../../skills/arch/README.md).

## Portability

The persona body is harness-neutral; only the wrapper format differs. Same subagent, four targets:

| Harness | Install location | Format | Persona field | Notes |
|---------|------------------|--------|---------------|-------|
| **Claude Code** | `~/.claude/agents/ai-dev.md` or `.claude/agents/` | Markdown + YAML frontmatter | md body | Drop in as-is. `name`, `description` required. |
| **Cursor 2.0** | reads `.claude/agents/` directly, or `~/.cursor/agents/ai-dev.md` | Markdown + YAML frontmatter | md body | No `tools` field — inherits all; only `readonly` / `is_background` restrict. `.cursor/` wins on name conflict. |
| **Codex** | `~/.codex/agents/ai-dev.toml` or `.codex/agents/` | **TOML** | `developer_instructions` | `name`, `description`, `developer_instructions` required. |
| **Grok** (`@vibe-kit/grok-cli`) | `~/.grok/user-settings.json` → `subAgents[]` | **JSON** entry | `instruction` | `name`, `model`, `instruction` all required. Reserved names: `general`, `explore`, `vision`, `verify`, `computer`. |

Two frontmatter fields do not survive the trip. `tools` exists in Claude Code and Cursor only —
Codex and Grok inherit the full pool, so a read-only agent like `arch-critic-runner` loses its
restriction and needs the prohibition restated in the body. `skills` is Claude Code-specific
entirely: it preloads the text of a skill into the subagent's context at start, and no other
harness resolves it. Porting `arch-critic-runner` elsewhere means pasting the body of
[`arch-critic/SKILL.md`](../../skills/arch/arch-critic/SKILL.md) into the persona itself.

Official **Grok Build** (xAI) has no documented custom named-persona mechanism — put the body in
`AGENTS.md` instead (all four harnesses read `AGENTS.md` as instruction text, but that is project
policy, not a delegatable subagent).

### Codex conversion

```toml
name = "ai-dev"
description = "Agentic systems architect - design, build, and ship production AI agents..."
model = "gpt-5.4"                 # optional; omit to inherit the session model
model_reasoning_effort = "high"   # optional
developer_instructions = """
<paste the Markdown body of ai-dev.md here>
"""
```

### Grok conversion (`@vibe-kit/grok-cli`)

```json
{
  "subAgents": [
    {
      "name": "ai-dev",
      "model": "grok-4.3",
      "instruction": "<paste the Markdown body of ai-dev.md here>"
    }
  ]
}
```

## Model per harness

The agent runs best on a strong reasoning model. Recommended per target (adjust for cost):

| Harness | Field | Value | Notes |
|---------|-------|-------|-------|
| Claude Code | `model` frontmatter | `opus` (or omit for `inherit`) | Alias resolves to the current Opus tier. |
| Cursor 2.0 | `model` frontmatter | full ID + params, e.g. `claude-opus-4-8[effort=high]` or `composer-2` | The `opus` alias does **not** resolve in Cursor — use an ID. |
| Codex | `model` + `model_reasoning_effort` | e.g. `gpt-5.4` + `high` | |
| Grok | `model` (required) | e.g. `grok-4.3` | |

## How to use

Pick your harness from [Portability](#portability), install the file at the listed location (or
convert the body), then adapt the persona and constraints to your standards. New agents are added
here over time — one file per agent, one row per agent in the table above.
