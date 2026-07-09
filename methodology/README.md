# Methodology — how to work, not just what's configured

The main differentiator of OFFICINA: reusable **rules and guides** for working with Claude Code and
your tooling — writing style, CLI authoring, skill & subagent authoring, loop engineering. Each is
kept readable on its own, meant to be read before the matching task rather than skimmed once. Part of
the [OFFICINA](../) repository.

These are working guides distilled from real projects, not abstract advice — copy what fits and adapt
the conventions to your setup. No absolute paths or secrets are included.

## Guides

| Guide | What it covers | Read before |
|-------|----------------|-------------|
| [`writing-style.md`](writing-style.md) | Writing rules for chat replies and document prose: straight-to-the-point tone, active voice, and a hygiene list against AI-slop (banned antithesis, dead transitions, dead vocabulary). | Producing any prose — a reply, a doc, a commit body. |
| [`cli-authoring-guide.md`](cli-authoring-guide.md) | Building personal CLI tools: CLI-vs-skill-vs-loop boundary, bash-vs-Python choice, colocated folder structure, behavioral conventions (exit codes, streams, flags), `uv`/PEP 723 dependencies, testing, and copy-ready templates. | A task that writes a CLI or utility. |
| [`skill-agent-authoring-guide.md`](skill-agent-authoring-guide.md) | Designing Claude Code skills and subagents: when to use which, full frontmatter reference, `description` rules, progress indication, checklists, output-structure files, subagent best practices, and a pre-deploy validation checklist. | Creating or editing a skill or subagent. |
| [`loop-engineering-rules.md`](loop-engineering-rules.md) | Building autonomous work loops: the five mandatory steps of a turn, the six parts that implement them, generator-vs-evaluator separation, five anti-patterns, scheduler choice (cloud/desktop/`/loop`), and operational discipline. | Designing a loop for a project or process. |

`cli-authoring-guide.md` and `skill-agent-authoring-guide.md` are a pair: the first draws the line for
deterministic terminal tools, the second for model-driven artifacts. `writing-style.md` underpins all
four — the other guides defer to it for tone and prose.

## How to use

Read the guide that matches the task before you start, not after. They are written as durable
references, not tutorials — scan the relevant section, apply it, and use the validation checklist at
the end (where present) before calling the work done.

New guides are added here over time — one file per guide, one row per guide in the table above.

---

**Keywords:** developer methodology, Claude Code skills, subagents, SKILL.md, loop engineering,
autonomous agents, CLI authoring, writing style, AI-slop, AI coding workflow.
**Topics:** `claude-code` · `methodology` · `claude-code-skills` · `subagents` · `developer-tools`
