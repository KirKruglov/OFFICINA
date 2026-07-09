# Claude Code skill & agent authoring guide

A general spec-and-checklist for building skills and subagents in this repository. Use it at the
start of any task that creates or edits a skill or an agent. Based on the official Claude Code
documentation (`code.claude.com/docs` — skills, sub-agents, memory); verified on **2026-07-10**.

Repository conventions and the deploy flow are documented separately (a project `CLAUDE.md` in your
dev workspace). This file describes **how artifacts are built and written**; the deploy doc — **where
they land and how they ship**.

---

## 1. When to use what: skill vs subagent vs CLAUDE.md

| Need | Choice |
|------|--------|
| A fact/rule that always applies (on every prompt) | `CLAUDE.md` |
| A reusable instruction, checklist, or multi-step process in the **main context** | **Skill** |
| An action launched by the `/name` command (deploy, commit) | **Skill** with `disable-model-invocation: true` |
| An isolated subtask with large output (search, tests, logs) → only the summary is needed | **Subagent** |
| Restricting the tool set / a separate model / separate permissions | **Subagent** |
| A persona-role with its own system prompt | **Subagent** |

Key distinction: a **skill** loads into the main context and stays there until the end of the
session; a **subagent** works in a separate context window and returns only a summary. CLAUDE.md
loads on every prompt — facts only, not procedures.

Skills and subagents compose:
- A skill with `context: fork` — the skill's content becomes the task for the named `agent`.
- A subagent with a `skills` field — the full text of the skills is injected into its context at
  start.

---

## 2. Skills

### 2.1. Folder structure

```
<skill-name>/
├── SKILL.md          # required — entry point (frontmatter + instructions)
├── reference.md      # opt. — detailed reference, loaded on demand
├── examples.md       # opt. — examples of the expected output
├── <artifact>-template.md  # opt. — layout of the output artifact (placeholders), §2.8
│                           #        OR <artifact>-format.md — format reference
└── scripts/
    └── helper.py     # opt. — scripts (executed, not loaded into context)
```

- Folder name = invocation command (`<skill-name>/SKILL.md` → `/skill-name`).
- Keep `SKILL.md` **< 500 lines**; move large reference material into separate files and link to them
  from `SKILL.md` (progressive disclosure: name + description always load, ≈100 tokens; the body on
  invocation; extra files on demand).

### 2.2. `SKILL.md` frontmatter

All fields are optional; `description` is the recommended minimum.

| Field | Purpose |
|-------|---------|
| `name` | Display name in listings. Defaults to the folder name. Does **not** change the invocation command (except plugin-root). |
| `description` | What it does and **when to use it**. Claude uses this field to decide whether to load the skill. Key case first. |
| `when_to_use` | Extra triggers/example requests. Appended to `description`. |
| `argument-hint` | Hint for expected arguments in autocomplete, e.g. `[issue-number]`. |
| `arguments` | Named positional arguments for `$name` substitution. |
| `disable-model-invocation` | `true` — the skill is invoked only manually via `/name` (Claude does not launch it itself). For actions with side effects. |
| `user-invocable` | `false` — hide from the `/` menu; only Claude invokes it (background knowledge). |
| `allowed-tools` | Tools pre-approved (no permission prompt) while the skill is active. Does not restrict the pool, only pre-approves. |
| `disallowed-tools` | Tools removed from the pool while the skill is active. |
| `model` | Model for the duration of the skill; `inherit` — keep the current one. |
| `effort` | Effort level: `low`/`medium`/`high`/`xhigh`/`max`. |
| `context` | `fork` — run in a forked subagent. |
| `agent` | Subagent type under `context: fork` (`Explore`, `Plan`, `general-purpose`, or custom). |
| `hooks` | Hooks on the skill's lifecycle. |
| `paths` | Glob patterns: auto-load the skill only when working with matching files. |
| `shell` | Shell for the `` !`commands` `` in the body (§2.5): `bash` (default) or `powershell`. |

> **Description limit.** The `description` + `when_to_use` pair is truncated at **1536 characters**
> in the listing (configurable via `skillListingMaxDescChars`). Key case in the first sentence.

### 2.3. `description` rules (the most important field)

- State **what** it does AND **when to apply it** (trigger conditions, not just a summary).
- Include the words the user actually says.
- Too general ("helps with code") — rarely fires; concrete ("when `npm test` fails") — fires
  reliably.
- Fires too often → narrow the description or add `disable-model-invocation: true`.

### 2.4. Lifecycle

- On invocation, the rendered `SKILL.md` enters the conversation as one message and **stays until the
  end of the session**; the file is not re-read. Write it as permanent instructions, not one-off
  steps.
- The body is a recurring token cost: say **what to do**, without the "how and why" narrative.
- **The body is self-contained — do not reference a companion `spec.md` from it** (or other
  dev-artifacts prepared before the skill is created). The `spec.md` lives in a dev workspace (e.g.
  `skill-dev/<name>/`); on deploy to `.claude/skills/<name>/` it is not alongside — the link would
  dangle. Keep rationale in `spec.md`, but do not point the body at it.

### 2.5. Dynamics and fork

- `` !`command` `` in the body — runs before the skill is sent, the output is substituted into the
  prompt (preprocessing, not a Claude call). Multi-line — a ` ```! ` block.
- `context: fork` — for skills with an **explicit task** (otherwise the subagent gets instructions
  with no action and returns empty).

### 2.6. Progress indication (mandatory for multi-step skills)

A multi-step/interactive skill **must** print a status line at the start of each stage — for
transparency of progress and visibility at gates.

**Format:** `[<skill-name>] <emoji> <present-tense action>: <dynamic> [<context>]`

- The `[<skill-name>]` tag — the skill name, the same across all lines.
- One emoji per stage, semantically tied to the action.
- The skill substitutes the dynamic value (`<value>`) on the fly; `[<context>]` — optional.

**Reserved emoji** (use strictly for their purpose, do not dilute):

| Emoji | Purpose |
|-------|---------|
| `🛑` | Stop / blocking error (no input, cannot continue) |
| `❓` | Clarification from the user (ambiguity, a decision is needed) |
| `⏸️` | A wait-for-confirmation point (gate) |
| `✅` | Successful finish of a stage/skill |

Working actions (read, generate, verify, write) — arbitrary themed emoji (`📥 🗺️ ✍️ 🔎 📋 🔗` etc.),
one per stage.

**Example:** `[user-story-decompose] 🔎 Checking coverage completeness [uc-reviewer]`

### 2.7. Checklist (mandatory for skills with >3 steps in the algorithm)

A skill whose algorithm has **more than 3 steps** **must**, at the very start of the run (before the
first step), print the plan to the user and create tasks via `TaskCreate` (status via `TaskUpdate`) —
**one task per algorithm step**. This gives the user a forward-looking map of progress (what is ahead
and where they are now) and keeps the skill from skipping steps over a long interactive process.

> **The task tool in this harness is the `Task*` family** (`TaskCreate` / `TaskUpdate` / `TaskList` /
> `TaskGet`). There is no `TodoWrite` here — that is a tool from a different Claude Code harness; it
> does not resolve in `allowed-tools`, and the plan silently drops into text. Do not use `TodoWrite`.

**Rules:**
- The plan is a **fixed list of steps**, independent of the input; print it before resolving
  arguments (do not tune the task wording to the input data at the end).
- Granularity — **the algorithm-step level**, no deeper. Do not split conditional/interactive
  sub-stages within a step into separate tasks — their transparency comes from the `⏸️` sub-confirmations
  (see §2.6).
- Keep a conditional step (which may be skipped) as a single task for its step, not as a separate
  candidate task — otherwise a skip looks like unfinished work.
- Mark each task done as you pass its step; execute strictly in order.
- Add `TaskCreate, TaskUpdate` to `allowed-tools`.
- The checklist and the status lines (§2.6) **do not duplicate** each other: the checklist = the
  arc/progress of the whole skill, the status line = the current dynamic action with value
  substitution. Do not introduce a separate text plan on top of the tasks (`TaskCreate`).

**Placement in `SKILL.md`:** a separate "Checklist (run plan)" block between "Progress indication"
and the step descriptions; it holds the list of steps and the mandate to create `TaskCreate` tasks
for them.

### 2.8. Output-structure file: template vs format reference

A companion file that defines the structure of the produced artifact comes in two **different** kinds.
Mixing the registers is not allowed — it is a recurring source of confusion.

| Kind | What it is | What is inside | Name |
|------|------------|----------------|------|
| **Template (layout)** | A literal skeleton of the output artifact that the skill fills in | Headers/tables of the final document + `<…>` placeholders; at most one orienting line at the top | `<artifact>-template.md` |
| **Format reference (canon)** | A description of the artifact's structure | A list of sections: what each contains, mandatory/conditional | `<artifact>-format.md` |

**Iron rule: the file describes the OUTPUT, not the PROCESS.** The fill-in logic — where to take data
from (input sections), what to include/omit, gates, step order — lives **only in `SKILL.md`** (in the
assembly step), not duplicated into the structure file.

**What goes where:**
- In a template — only the shape of the result. No "extract from §N of the input", "if X — state it",
  "no hex/px": that is procedure → into `SKILL.md`.
- In a format reference — a description of the sections, not a step-by-step algorithm.
- One rule in one place: if "where to take a block from" is already in a `SKILL.md` step, do not
  repeat it in the structure file.

**How to choose the kind:**
- Output compact and uniform (one screen, fixed blocks) → **template** (`-template.md`).
- Output large/variable (structure depends on the input, conditional sections) → **reference**
  (`-format.md`).

**Litmus test before saving.** Read any line of the file: does it tell Claude *what to do* (a verb:
"extract", "verify", "if…") → that is an instruction, move it to `SKILL.md`. Does it show *what the
result looks like* (a header, a placeholder, a field) → keep it.

---

## 3. Subagents

### 3.1. File format

File `<agent-name>.md`: YAML frontmatter + body (system prompt). A subagent gets only this system
prompt + a base environment (not the whole Claude Code system prompt). It starts in a separate clean
context. As of Claude Code v2.1.172 a subagent can spawn nested subagents if `Agent` is listed in its
`tools` (max nesting depth — 5); by default, without `Agent` in `tools`, it does not spawn any.

```markdown
---
name: agent-name
description: When Claude should delegate to this subagent
tools: Read, Grep, Glob
model: inherit
---

System prompt: role, workflow, output format.
```

### 3.2. Subagent frontmatter

Only `name` and `description` are mandatory.

| Field | Purpose |
|-------|---------|
| `name` | Unique id (lowercase + hyphens). The filename need not match. |
| `description` | When to delegate. Write it as triggers; for proactivity — "use proactively". |
| `tools` | Allowlist of tools. If omitted — inherits all. To preload skills use `skills`, not `Skill` here. |
| `disallowedTools` | Denylist (applied before `tools`). |
| `model` | `sonnet`/`opus`/`haiku`/`fable`, a full ID, or `inherit` (default `inherit`). |
| `skills` | Skills to preload — the full text is injected at start. |
| `permissionMode` | `default`/`acceptEdits`/`auto`/`dontAsk`/`bypassPermissions`/`plan`. |
| `maxTurns` | Limit on agent turns. |
| `memory` | Persistent memory: `user`/`project`/`local`. `project` — the recommended default. |
| `mcpServers` | MCP servers available to the subagent (inline or by name). |
| `hooks` | Subagent lifecycle hooks. |
| `isolation` | `worktree` — an isolated copy of the repo in a git worktree. |
| `effort` | Effort level. |
| `background` | `true` — always run in the background. |
| `color` | Color in the task list. |

### 3.3. Best practices (official)

- **One subagent — one task**, a narrow specialization.
- **A detailed `description`** — Claude uses it to decide whether to delegate.
- **Minimum tools.** The default for reviewers/researchers is read-only: drop `Edit`, `Write`,
  `NotebookEdit`; leave all writing to the parent.
- **Version** project subagents (`.claude/agents/`) for the team.
- Unavailable to subagents even if listed in `tools`: `AskUserQuestion`, `EnterPlanMode`,
  `ExitPlanMode` (except under `permissionMode: plan`), `ScheduleWakeup`, `WaitForMcpServers`. `Agent`
  is available — it enables spawning nested subagents (§3.1).
- **The system prompt is self-contained — do not reference `spec.md`/dev-artifacts** (see §2.4): the
  subagent does not see them, and on deploy they are not alongside.

### 3.4. When a subagent fits

Search over a large codebase; tasks with verbose output; independent parallel tasks; isolating
experimental edits (worktree). When you need frequent dialogue/shared context/a quick small edit —
the main context, not a subagent.

---

## 4. Validation checklist before deploy

**Common**
- [ ] Name — `kebab-case`, Latin; unique in its scope.
- [ ] Frontmatter filled in **English** (except necessary native-language trigger phrases).
- [ ] `description` states what + when, key case first, within 1536 chars.
- [ ] The skill/subagent body has no references to a companion `spec.md` / dev-artifacts (the body is
      self-contained, §2.4).

**Skill**
- [ ] `SKILL.md` < 500 lines; large reference moved to extra files and linked.
- [ ] The body is permanent instructions (not one-off steps), without excess narrative.
- [ ] For actions with side effects — `disable-model-invocation: true`.
- [ ] `allowed-tools` — only what is needed; paths in scripts via `${CLAUDE_SKILL_DIR}`.
- [ ] Under `context: fork` — the body has an explicit task.
- [ ] For skills with side effects — a "Boundaries" block (explicit prohibitions gathered in one
      place).
- [ ] For multi-step skills — status lines per §2.6 (tag, one emoji per stage, reserved
      `🛑/❓/⏸️/✅`).
- [ ] For skills with >3 steps — a "Checklist" block per §2.7 (plan at the start + `TaskCreate`, one
      task per step, status via `TaskUpdate`; `TaskCreate, TaskUpdate` in `allowed-tools`).
- [ ] The output-structure file is of one register (template `-template.md` OR reference
      `-format.md`, §2.8); describes the output, not the process.
- [ ] The fill-in logic (data sources, conditions, gates) — only in `SKILL.md`, not duplicated in the
      structure file.

**Subagent**
- [ ] `tools` restricted by the minimum principle (read-only if no writing is needed).
- [ ] `model` chosen deliberately (`inherit`/`haiku` for cheap tasks / `sonnet`+ for analysis).
- [ ] System prompt: role, step-by-step workflow, output format.
- [ ] Accounted for the subagent not seeing the conversation history — critical rules duplicated in
      the prompt.

**Deploy** (per your deploy conventions)
- [ ] Skill → `.claude/skills/<skill-name>/`; subagent → `.claude/agents/<agent-name>.md`.
- [ ] Deploy — only after an explicit user command.

---

## 5. Templates (copy-ready)

### 5.1. `SKILL.md`

```markdown
---
name: skill-name
description: What it does and when to use it (English). Key case first. Triggers — "phrase1", "phrase2".
allowed-tools: Read Grep
---

# Skill Name

## When to apply
[Trigger conditions]

## Progress indication
Status-line format: `[skill-name] <emoji> <action>: <value> [<context>]`. Emoji `🛑/❓/⏸️/✅` —
stop / clarify / wait / finish (§2.6).

## Checklist (run plan) — only if >3 steps (§2.7)
At the start of the run, before Step 1, print the plan and create `TaskCreate` tasks, one per step
(status via `TaskUpdate`). The list is fixed, independent of the input:
1. Step 1 — [title]
2. Step 2 — [title]
...
Execute in order, mark done as you pass each. `TaskCreate, TaskUpdate` — in `allowed-tools`.

## Instructions
1. Status: `[skill-name] <emoji> ...` → [Step] → check: [criterion]
2. Status: `[skill-name] <emoji> ...` → [Step] → check: [criterion]

## Boundaries (what the skill does not do)
- [Prohibition/boundary — recommended block: explicit runtime prohibitions, destructive operations,
  scope limits]

## Further material
- Details: [reference.md](reference.md)
```

> **The "Boundaries" block.** Recommended for skills with side effects (writing/editing files,
> deploy, delegation): gather the hard prohibitions in one place (what not to touch, where to stop,
> what not to do silently) instead of scattering them across steps.

### 5.2. `agent.md`

```markdown
---
name: agent-name
description: When Claude should delegate. Use proactively when [trigger].
tools: Read, Grep, Glob
model: inherit
---

You are [role]. [Area of responsibility.]

When invoked:
1. [Step]
2. [Step]

[Process / checklist.]

Output format:
- [Structure of the result]
```
