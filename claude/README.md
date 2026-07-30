# Claude Code setup — subagents, `CLAUDE.md`, `settings.json`, and a custom status line

The **Claude Code** layer of OFFICINA: the Claude-specific artifacts of a working setup — **subagents**
you delegate scoped tasks to, two layers of `CLAUDE.md`, a `settings.json` with a curated permission
allowlist, and a status line that shows the model, context usage, rate limits, and branch. Hand-picked
rather than mirrored wholesale: what's here is in daily use. Part of the [OFFICINA](../) repository.

Everything below is an **example to adapt, not a drop-in profile**. Paths are genericized to
`~/.claude/…`, no secrets are included, and the safety-sensitive flags are called out explicitly.

## What's in this layer

| Folder | What it holds | Status |
|--------|---------------|--------|
| [`agents/`](agents/) | Subagents — specialized agents for delegated tasks, portable to Cursor / Codex / Grok | 🚧 Growing |
| [`settings/`](settings/) | Global config — `CLAUDE.md` examples, `settings.json`, status line script | 🚧 Growing |

The two folders answer different questions. `agents/` is about **who** does the work — a role with its
own context and its own standards. `settings/` is about **the ground rules** every session runs on:
what Claude may do without asking, what it should know about you before you type anything, and what
you see at the bottom of the terminal while it works.

## Subagents

### What a subagent is

A subagent is a delegated worker with its own **context window**, its own **model**, and its own
**behavioral contract**. You hand it a scoped task; it does the work in isolation and returns a
result. Nothing it read along the way lands in your main conversation — which is the point. A research
sweep across forty files costs you one paragraph of context instead of forty file dumps.

The second thing a subagent buys you is a **persona that holds**. A general session drifts toward
agreeableness; a subagent whose contract says *argue and defend your position, and never quote a
version from memory without verifying it* keeps doing that on turn thirty. That behavioral contract
is the whole file. Everything else is frontmatter.

Each subagent is a single Markdown file with YAML frontmatter — `name`, `description`, and an optional
`model`. Drop it into `~/.claude/agents/` to make it available everywhere, or into a project's
`.claude/agents/` to scope it to that repository.

### `ai-dev` — agentic-systems architect

| Agent | Model | What it does |
|-------|-------|--------------|
| [`ai-dev`](agents/ai-dev.md) | `inherit` | Designs, builds, and ships production AI agents across the full stack |

A senior staff engineer persona for **building AI agents**, not for talking about them. It thinks in
terms of the agent stack rather than prompts: surface, orchestration, tools (MCP as the default
agent-to-tool protocol, A2A between agents), knowledge and RAG, memory, model and inference routing,
runtime isolation, evals and observability as a vertical rail, governance. Cost, latency, and security
are treated as first-class constraints rather than afterthoughts.

Its refusals are as specific as its recommendations: no autonomous agent where a deterministic
pipeline would be more reliable, no architecture handed over without an evals and observability layer,
no tool named without a trade-off and a recommendation, and no version or price quoted from memory
without checking it first. It pushes back rather than agreeing by default.

The `model` field is deliberately unset — Claude Code and Cursor default to `inherit`, so the agent
adopts the session model. Per-harness recommendations are in
[`agents/README.md`](agents/README.md#model-per-harness).

### `architect` and `arch-critic-runner` — a session owner and its blind critic

| Agent | Model | What it does |
|-------|-------|--------------|
| [`architect`](agents/architect.md) | `inherit` | Runs a whole session as a skeptical-minimalist software architect |
| [`arch-critic-runner`](agents/arch-critic-runner.md) | `inherit` | Reviews one decision package with no access to the reasoning behind it |

`architect` inverts the usual pattern: you start a session with `claude --agent architect` instead
of dispatching it into one. Its contract is a short list of refusals — no version stated from
memory, no dependency added without a named need, no decision left in the chat log instead of a
repository file — and a routing table that points each situation at the matching
[`arch-*` skill](../skills/arch/).

`arch-critic-runner` is the opposite shape: a subagent that exists purely to be delegated to. What
makes it work is subtraction. It receives the decision package and nothing else — no candidate
list rationale, no discussion history — because a reviewer who has read the justification tends to
ratify it. Its `tools` are `Read, Grep, Glob`, and the `skills` field preloads the `arch-critic`
procedure into its context at start.

Subtraction that thorough needs a check on the other end. A subagent whose `skills:` preload did not
resolve — the usual cause being that the skill was never installed — still answers, fluently and
plausibly, from no procedure at all. So the skills that dispatch it verify the file exists before the
run starts and verify the shape of the report when it comes back: the heading, the verdict line, the
table of four axes. Anything else counts as no review having happened, which is recorded as such.
Silence would have been safer than an improvised verdict, and a check on the format is what tells the
two apart.

Both are documented with the system they serve in
[`skills/arch/`](../skills/arch/README.md).

### Portability beyond Claude Code

The persona body is harness-neutral; only the wrapper differs. The same subagent runs on four targets:

| Harness | Install location | Format | Persona field |
|---------|------------------|--------|---------------|
| **Claude Code** | `~/.claude/agents/` or `.claude/agents/` | Markdown + YAML frontmatter | the Markdown body |
| **Cursor 2.0** | reads `.claude/agents/` directly | Markdown + YAML frontmatter | the Markdown body |
| **Codex** | `~/.codex/agents/*.toml` | TOML | `developer_instructions` |
| **Grok** (`@vibe-kit/grok-cli`) | `~/.grok/user-settings.json` → `subAgents[]` | JSON entry | `instruction` |

Claude Code and Cursor take the file as-is. Codex and Grok need the body pasted into a different
wrapper once — conversion snippets for both are in [`agents/README.md`](agents/README.md#portability),
along with the reserved-name list for Grok and the per-harness model IDs.

## Two `CLAUDE.md` layers, and what belongs in each

Claude Code reads `CLAUDE.md` at two levels, and mixing them up is the most common way a setup goes
stale. The distinction is stable:

- **`~/.claude/CLAUDE.md`** — cross-project *preferences*. How you want Claude to behave everywhere,
  regardless of what you're working on. Loaded in every session, so it stays short: communication
  style, default permission posture, plan-before-changes, git conventions. See
  [`CLAUDE.global.md`](settings/CLAUDE.global.md) — it fits on one screen, on purpose.
- **A project's root `CLAUDE.md`** — project-specific *navigation and rules*. What the repository is,
  where things live, which conventions apply per area, what must never happen here. Loaded only in
  that project. See [`CLAUDE.project.md`](settings/CLAUDE.project.md).

They compose: global sets the baseline, project layers specifics on top. A rule in the wrong layer
either bloats every unrelated session or silently fails to apply where you needed it.

`CLAUDE.project.md` is a **template**, not a finished file. Placeholders (`<PROJECT_NAME>`,
`<CODE_DIR>`, `<VISION_DOC>`, …) get filled per project, optional blocks marked "delete if …" get
removed to match the layout, and the methodological core — parallel-agent rules, analysis discipline,
discussion-session behavior, mandatory rules — carries over unchanged. That core is the part worth
copying; the rest is scaffolding.

## `settings.json` — permissions and defaults

[`settings.json`](settings/settings.json) is the global configuration file at `~/.claude/`. The keys
that matter most:

| Key | Value here | What it does |
|-----|------------|--------------|
| `permissions.allow` | read-only + safe git/curl | Auto-approves a curated allowlist so routine calls don't prompt |
| `permissions.defaultMode` | `auto` | Auto-approves allowlisted calls; still prompts for anything outside the list |
| `effortLevel` | `high` | Reasoning depth per turn — more thorough, slower |
| `autoCompactEnabled` | `true` | Auto-summarizes context when the window fills, so long sessions continue |
| `enabledPlugins` | code-review, frontend-design, superpowers | Plugins from the official marketplace declared in `extraKnownMarketplaces` |
| `tui` | `fullscreen` | Full-screen terminal UI instead of inline |

The full annotated table, including every flag not listed here, is in
[`settings/README.md`](settings/README.md).

### How to build a permission allowlist

The allowlist is where a setup either saves you a hundred prompts a day or quietly hands over more
than you meant. The rule that keeps it safe: **allow reads freely, allow writes never**.

Everything in the allowlist here either reads or fetches — `Read`, `WebFetch`, `WebSearch`, `find`,
`ls`, `grep`, `rg`, `cat`, `head`, `curl -s`. The git entries are read-only by construction:
`git status`, `git log`, `git diff`, `git show`, `git blame`, `git branch`, `git rev-parse`,
`git ls-files`. There is no `git add`, no `git commit`, no `git push`, no `rm`, no shell redirection.
Prompts survive exactly where an action changes something.

[`settings.local.json`](settings/settings.local.json) shows the project-local layer — normally
git-ignored — where a repository you trust can add `git add` and `git commit` to the allowlist without
that permission leaking into every other project. The same file demonstrates the `deny` list:
`Read(.env)`, `Bash(cat .env*)`, `Bash(echo * > .env*)`. Deny wins over allow, which makes it the
right place for anything that must never happen regardless of mode.

### Safety-sensitive flags

Three keys in `settings.json` remove confirmation steps rather than add capability:

| Key | Effect |
|-----|--------|
| `skipDangerousModePermissionPrompt` | Suppresses the confirmation before entering bypass-permissions mode |
| `skipAutoPermissionPrompt` | Suppresses the auto-approval prompt |
| `skipWorkflowUsageWarning` | Hides the multi-agent workflow token-cost warning |

They lower friction at the cost of guardrails, and they fit a solo workflow in a trusted environment.
Copy them only if that describes yours — they're listed separately here precisely so they aren't
copied by accident.

## The status line

[`statusline-command.sh`](settings/statusline-command.sh) replaces the default status line with one
that answers the four questions you actually ask mid-session — which model is running, how full the
context is, how much of the rate limit is left, and which branch you're on:

```
Opus 4.8 (1M context) (high) | █░░░░░░░░░ 9% (89k/1000k) | 5h:6% ~4h11m | wk:48% ~7h31m | ⎇ release/0.1.0
```

Each `|`-separated segment is color-coded, and the usage segments shift green → yellow → red as they
fill, so a context window about to compact is visible before it happens. The per-segment breakdown is
in [`settings/README.md`](settings/README.md#status-line).

It needs two things on the machine: [`jq`](https://jqlang.github.io/jq/) for parsing, and a
[Nerd Font](https://www.nerdfonts.com/) for the `⎇` branch glyph. `statusLine.command` points at
`~/.claude/statusline-command.sh` — a `~`-relative path, so the script has to live in `~/.claude/` for
it to resolve. Absolute paths in that key are the single most common way a personal `settings.json`
leaks a home directory into a public repository.

## Subagent, skill, or CLI?

These three artifacts get confused constantly, and the distinction is about **what varies**:

| Need | Artifact | Why |
|---|---|---|
| A deterministic action, no judgment | **CLI** | A script is cheaper, faster, and can't improvise. See [`cli/`](../cli/) |
| A fixed procedure with situational decisions inside it | **Skill** | Same steps every time; the calls within them depend on context. See [`skills/`](../skills/) |
| A whole role, or a subtask that shouldn't pollute your context | **Subagent** | Its own context window, its own model, its own standards |

A skill changes *how the current agent behaves*. A subagent *is a different agent*. If what you're
writing starts with "always do X before Y", it's a skill; if it starts with "you are a…", it's a
subagent. The full reasoning and the authoring guides live in [`methodology/`](../methodology/).

## FAQ

### Where do Claude Code subagents live?

`~/.claude/agents/<name>.md` for subagents available in every project, or
`<project>/.claude/agents/<name>.md` for ones scoped to a single repository. One Markdown file per
agent — there's no folder structure to build.

### What's the difference between global and project `CLAUDE.md`?

`~/.claude/CLAUDE.md` holds preferences that apply everywhere and is loaded in every session, so it
should stay short. A project's root `CLAUDE.md` holds that repository's map and conventions and loads
only there. They compose rather than override wholesale: global sets the baseline, project layers on
top.

### Do these subagents work in Cursor or Codex?

In Cursor 2.0, yes — as-is. It reads `.claude/agents/` directly, and the frontmatter format matches.
Codex and Grok need the persona body moved into a different wrapper (TOML and JSON respectively);
conversion snippets are in [`agents/README.md`](agents/README.md#portability).

### How do I stop Claude Code from asking permission for every command?

Add the calls you trust to `permissions.allow` in `settings.json` and set `permissions.defaultMode`
to `auto`. Keep the list read-only — reads, searches, fetches, and the read-only git subcommands.
Anything that writes, commits, pushes, or deletes should still prompt. Repository-specific exceptions
belong in a project-local `settings.local.json`, not the global file.

### Can I use a custom status line without a Nerd Font?

Yes — the script needs a Nerd Font only for the `⎇` branch glyph. Replace that character with a plain
ASCII label and the rest of the line renders in any terminal font. `jq` is the one hard dependency.

### Is this a drop-in profile I can copy entirely?

No, and it isn't meant to be. Permissions, plugins, effort level, and the safety-sensitive flags
encode one person's risk tolerance and workflow. Take the structure — two `CLAUDE.md` layers, a
read-only allowlist, a project-local override file — and fill it with your own decisions.

## Related

- [OFFICINA](../) — the method and the rest of the toolkit
- [`skills/`](../skills/) — portable `SKILL.md` procedures that work alongside these subagents
- [`methodology/`](../methodology/) — the guides behind these artifacts, subagent authoring included
- [`cli/`](../cli/) — deterministic tools for the work that needs no model
- [Claude Code](https://claude.com/claude-code) — the official product page
