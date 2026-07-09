# Agents — specialized Claude Code subagents

Subagents you delegate a scoped task to: each runs with its own role, model, and behavioral
contract, and returns a focused result instead of you steering a general model through the whole
job. Part of the [OFFICINA](../../) repository.

These are examples, not a drop-in set — copy what fits and adapt the persona to how you work. No
absolute paths or secrets are included.

## Agents

| Agent | Model | What it does |
|-------|-------|--------------|
| [`ai-dev`](ai-dev.md) | `opus` | Agentic-systems architect and senior staff engineer. Designs, builds, and ships production AI agents across the full stack — surface, orchestration, tools (MCP/A2A), knowledge/RAG, memory, model/inference, runtime/isolation, evals + observability, governance. Thinks in trade-offs, insists on evals before deploy, treats cost/latency/security as first-class, verifies fast-moving facts before asserting them, and pushes back rather than agreeing by default. |

## How to use

Copy an agent file into `~/.claude/agents/` (available in every project) or a project's
`.claude/agents/` (that project only), then adapt the persona to your context. Adjust the `model`
field to trade capability for cost, and edit the constraints to match your standards.

New agents are added here over time — one file per agent, one row per agent in the table above.
