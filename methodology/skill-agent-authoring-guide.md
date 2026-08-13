# Claude Code skill & agent authoring guide

A general spec-and-checklist for building skills and subagents in this repository. Use it at the
start of any task that creates or edits a skill or an agent. Based on the official Claude Code
documentation (`code.claude.com/docs` — skills, sub-agents, memory). Skill sections verified on
**2026-07-31**; subagent sections retain the **2026-07-10** revision.

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

**Granularity and composition**

- One skill — one workflow with one result. Several results or several domains — several skills plus
  an orchestrator.
- The orchestrator invokes stages explicitly (via an executor subagent or a fork, §2.5). A text
  pointer ("see skill X") leaves invocation to Claude and fires unreliably. Fork the stage, not the
  orchestrator: the orchestrator keeps the dialogue and the gates in the main context.
- Run stages through a subagent rather than loading them into the current context: an invoked skill
  body stays until the session ends, and on compaction early invocations drop wholesale (§2.4).
  Inline calls: at most 2.

---

## 2. Skills

### 2.1. Folder structure

The only allowed layout is catalog-style: the skill root holds `SKILL.md` and nothing else;
companions go into directories by role. Do not place companion files in the root.

```
<skill-name>/
├── SKILL.md          # required — entry point (frontmatter + instructions)
├── references/       # opt. — reference and examples, loaded on demand
│   ├── <topic>.md    #        one file per topic; a single topic → reference.md
│   └── examples.md   #        examples of the expected output
├── templates/        # opt. — output-structure files (§2.9)
│   ├── <artifact>-template.md  #  layout with placeholders
│   └── <artifact>-format.md    #  OR a format reference — one kind per artifact
├── scripts/          # opt. — scripts (executed, not loaded into context)
│   └── helper.py
├── assets/           # opt. — immutable resources (images, fonts, stubs);
│                     #        read by a program, not by Claude
└── evals/            # opt. — skill verification scenarios (§2.11)
    └── evals.json
```

- Create a directory only when it holds files. Do not add empty `scripts/`, `assets/`, or the like.
- `references/examples.md` is required when the result is prose whose quality a template cannot
  catch (an acceptance criterion, a scenario step, the wording of a finding). For tabular or
  structural output, do not add examples. Contents: 2–3 "weak → strong" pairs **on a fragment**,
  each with a one-line rationale. A whole-document example is forbidden: it is reproduced as
  content — domain words leak into a foreign artifact.
- An example does not introduce requirements and does not repeat the step algorithm — it only shows
  the quality bar. A requirement that appears only in an example is lost at run time: examples are
  read selectively and last.
- Both kinds of output-structure file (§2.9) live in `templates/` — the naming rule and the
  placement rule must not live in different sections.
- Folder name = invocation command (`<skill-name>/SKILL.md` → `/skill-name`).
- Keep `SKILL.md` **< 500 lines**; move large reference material into separate files and link to them
  from `SKILL.md` (progressive disclosure: name + description always load, ≈100 tokens; the body on
  invocation; extra files on demand).
- Every companion file is linked directly from `SKILL.md`. Link chains (`SKILL.md` → `advanced.md` →
  `details.md`) are forbidden: Claude follows a nested link with a partial read (`head -100`) and
  gets incomplete data.
- A reference file longer than 100 lines opens with a table of contents — otherwise a partial read
  cannot see the file's makeup.
- The file name states the contents: `form-validation-rules.md`, not `doc2.md`.

### 2.2. `SKILL.md` frontmatter

In Claude Code every field is optional; `description` is the recommended minimum. The Agent Skills
standard requires `name` and `description` — see the hard validation limits below.

| Field | Purpose |
|-------|---------|
| `name` | Display name in listings. Defaults to the folder name. Does **not** change the invocation command (except plugin-root). |
| `description` | What it does and **when to use it**. Claude uses this field to decide whether to load the skill. Key case first. |
| `when_to_use` | Extra triggers/example requests. Appended to `description`. |
| `argument-hint` | Hint for expected arguments in autocomplete, e.g. `[issue-number]`. |
| `arguments` | Named positional arguments for `$name` substitution. |
| `disable-model-invocation` | `true` — the skill is invoked only manually via `/name` (Claude does not launch it itself). For actions with side effects. |
| `user-invocable` | `false` — hide from the `/` menu; only Claude invokes it (background knowledge). |
| `allowed-tools` | Tools pre-approved (no permission prompt) **on the turn the skill is invoked**. The grant lifts on the next user message; a later invocation grants it again. Does not restrict the pool, only pre-approves. |
| `disallowed-tools` | Tools removed from the pool **on the invocation turn**; the restriction lifts on the next user message. |
| `model` | Model for the current turn; not saved to settings, the next prompt returns to the session model. `inherit` — keep the current one. |
| `effort` | Effort level: `low`/`medium`/`high`/`xhigh`/`max`. |
| `context` | `fork` — run in a forked subagent. |
| `agent` | Subagent type under `context: fork`: `Explore` (search, read-only), `Plan` (design), `general-purpose`, or a custom one from `.claude/agents/`. |
| `background` | Only with `context: fork`. Default `true` — the fork goes to the background. `false` — wait for the result in the invocation turn. |
| `shell` | Shell for the `` !`commands` `` in the body (§2.5): `bash` (default) or `powershell`. |
| `hooks` | Hooks on the skill's lifecycle. |
| `paths` | Glob patterns: auto-load the skill only when working with matching files. |

> **Permissions last one turn; the body lasts the session.** `allowed-tools` and `disallowed-tools`
> lift on the next user message, even though the skill text stays in context. A multi-step
> interactive skill only has pre-approvals on the first turn; lasting permissions go through
> `permissions` in settings.

> **Description limit.** The `description` + `when_to_use` pair is truncated at **1536 characters**
> in the listing (configurable via `skillListingMaxDescChars`). Key case in the first sentence.

> **Listing budget.** The full skill listing is capped at 1% of the context window
> (`skillListingBudgetFraction`, or a character cap via `SLASH_COMMAND_TOOL_CHAR_BUDGET`). On
> overflow, descriptions are dropped wholesale, starting with the least-used skills: names remain,
> auto-invocation breaks. Cost is visible in `/doctor` and the Skills line of `/context`; unload via
> `skillOverrides` in settings (`"on"`, `"name-only"`, `"user-invocable-only"`, `"off"`).

| Field | Constraint |
|-------|------------|
| `name` | Required by the Agent Skills standard. ≤ 64 chars; lowercase Latin, digits, hyphens; no XML tags; the words `anthropic` and `claude` are forbidden. |
| `description` | Required by the standard. Non-empty, ≤ **1024** chars, no XML tags. |

A skill whose description exceeds 1024 chars still loads in Claude Code, but fails to load in the
API and on claude.ai.

### 2.3. `description` rules (the most important field)

- Write in the **third person**: the description is injected into the system prompt, and a change
  of person breaks skill selection. "Assembles a feature spec from a PRD" is correct; "I will help
  assemble a spec", "You can assemble a spec" are not.
- State **what** it does AND **when to apply it** (trigger conditions, not just a summary).
- Split the fields: `description` — what it does and the key trigger; `when_to_use` — extra request
  phrasings.
- Include the words the user actually says.
- Phrase the trigger concretely: "when `npm test` fails". General wording ("helps with code")
  almost never fires.
- Name the skill with a verb or a gerund (`to-spec`, `processing-pdfs`) that names the action.
  Names `helper`, `utils`, `tools`, `data`, `files` are forbidden: the skill will not be found.

### 2.4. Lifecycle

- On invocation, the rendered `SKILL.md` enters the conversation as one message and **stays until the
  end of the session**; the file is not re-read. Write it as permanent instructions, not one-off
  steps.
- Permissions from `allowed-tools` / `disallowed-tools` do not follow the body — they last one turn
  (§2.2).
- A re-invocation with the same rendered contents does not duplicate the text: the conversation
  gets a marker that the skill is already loaded. Contents are appended again only if the arguments
  or the `` !`command` `` output changed.
- On auto-compaction the last invocation of each skill is re-attached, the first 5000 tokens of
  each, total budget 25 000 tokens, filled from the newest invocation. Skills invoked long ago drop
  wholesale; if behavior slumps after compaction — invoke the skill again.
- The body is a recurring token cost: say **what to do**, without the "how and why" narrative.
- **The body is self-contained — do not reference a companion `spec.md` from it** (or other
  dev-artifacts prepared before the skill is created). The `spec.md` lives in a dev workspace (e.g.
  `skill-dev/<name>/`); on deploy to `.claude/skills/<name>/` it is not alongside — the link would
  dangle. Keep rationale in `spec.md`, but do not point the body at it.
- **The spec is normative, the body is procedural.** The spec states requirements ("the skill does
  not ship a finding without a source"); the body states actions ("attach a source; if attaching
  failed — drop the finding"). Do not copy spec paragraphs into the body verbatim: each requirement
  becomes an action in a step (§2.8), a line in the "Boundaries" block, or a reference file.

### 2.5. Dynamics, substitutions, and fork

**Inline shell**

- `` !`command` `` in the body — runs before the skill is sent, the output is substituted into the
  prompt (preprocessing, not a Claude call). Multi-line — a ` ```! ` block.
- `!` is recognized only at the start of a line or immediately after a space: in `` KEY=!`cmd` ``
  the substitution does not fire and the line stays as text.
- Substitution is one-shot: the output is inserted as text and is not scanned again for
  placeholders — a command cannot emit a placeholder for a later pass.
- Disabled by the `disableSkillShellExecution` setting; in managed settings the user cannot
  override it.

**String substitutions**

| Substitution | Meaning |
|--------------|---------|
| `$ARGUMENTS` | All invocation arguments. If the body has no placeholder — arguments are appended at the end as `ARGUMENTS: <value>`. |
| `$ARGUMENTS[N]`, `$N` | Argument by position, zero-based. Quote a multi-word value at invocation. |
| `$name` | A named argument declared in the `arguments` field; names map to positions in order. |
| `${CLAUDE_SKILL_DIR}` | The folder that holds this skill's `SKILL.md`. |
| `${CLAUDE_PROJECT_DIR}` | The project root. |
| `${CLAUDE_SESSION_ID}` | The current session id. |
| `${CLAUDE_EFFORT}` | The current effort level (`low`…`max`). |

`${CLAUDE_SKILL_DIR}` and `${CLAUDE_PROJECT_DIR}` are substituted in the body and in
`allowed-tools`. The same path in both places removes the permission prompt for launching a skill
script:

```yaml
allowed-tools: Bash(${CLAUDE_SKILL_DIR}/scripts/render.sh *)
```

and in the body — "run `${CLAUDE_SKILL_DIR}/scripts/render.sh <file>`".

**Fork (`context: fork`)**

- Only for skills with an **explicit task**: without an action the subagent receives instructions
  and returns empty.
- The fork does not see the conversation history: the subagent gets the skill body as its prompt
  and starts from scratch. Do not fork a skill with `⏸️` gates, `❓` clarifications, or an interview
  — there is no one to ask, and the user's answers never reach the subagent.
- `agent: Explore` and `agent: Plan` load neither `CLAUDE.md` nor git status: the subagent context
  holds only the `SKILL.md` text and the agent's system prompt. A skill that takes paths or project
  rules from `CLAUDE.md` cannot use these agents — it will not find them and will start inventing.
  For that skill use `general-purpose` or a custom agent.
- By default the fork goes to the background — the turn is not blocked, the result arrives in the
  conversation later. `background: false` — wait for the result in the invocation turn.
- A wait is forced regardless of `background` in 4 cases: non-interactive mode (`-p`, Agent SDK);
  `CLAUDE_CODE_DISABLE_BACKGROUND_TASKS=1`; a repeated fork invocation while the previous one is
  still running; a scheduled launch. Do not treat background execution as a given.
- A background fork runs with a narrowed tool set. If a skill step depends on a tool outside that
  set — set `background: false`.
- Edits from a background fork bypass checkpoints: `/rewind` does not roll them back; rollback is
  through git only.
- A fork breaks command chaining: parsing of `/skill-a /skill-b …` stops at the first forked skill,
  and the rest becomes its arguments. A skill designed as a link in a command chain must not be
  forked — whether it is background or `background: false`.

### 2.6. Progress indication (mandatory for multi-step skills)

> A local repository requirement: absent from the official Claude Code docs, and not in conflict
> with them.

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

> A local repository requirement. The official guide offers a copy-paste text checklist; here the
> harness tasks are `TaskCreate` / `TaskUpdate` / `TaskList` / `TaskGet`. `TodoWrite` is not defined
> in this harness and does not resolve in `allowed-tools` — do not use it.

A skill whose algorithm has **more than 3 steps** **must**, at the very start of the run (before the
first step), print the plan to the user and create tasks via `TaskCreate` (status via `TaskUpdate`) —
**one task per algorithm step**. This gives the user a forward-looking map of progress (what is ahead
and where they are now) and keeps the skill from skipping steps over a long interactive process.

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

### 2.8. Step anatomy

A step is written as a procedure, not as a set of rules. A skill is executed from the body
literally: a list of norms with no action order produces different runs on the same input.

**Step form:**

```markdown
## Step N — Title
Status: `[skill-name] <emoji> <action>: <dynamic>`

1. **Action.** What to do and with what.
2. **Action.** …
3. Repeat items 1–2 until <condition>.

→ check: <exit criterion for the step>
```

**Rules:**

- Every item starts with an action: read, fill in, ask, attach, write. An item with no action goes
  to the table below.
- Item order matches execution order. An item that fires before its neighbors stands above them.
- Loops, branches, and returns are separate items: "repeat items 2–4 until the list is exhausted",
  "invariant broken — return to step 5". A caveat at the end of the step does not set the order.
- The condition that skips the whole step is the first item of the step.
- One rule lives at one address. A rule used in several steps is placed where it first applies;
  later — a reference.
- A reference table (reason lists, limits, codes) sits under the step items with a label of what it
  is, or moves to a reference file.
- The criterion in `→ check` is an observable sign: the file exists, every ID from the list
  appeared, the section count matches the template. A judgment ("looks complete", "detailed
  enough") is not a criterion — it can be neither confirmed nor refuted at run time.
- A step that produces an artifact must have a failure branch with a return point: "criterion not
  met — return to item 2 and fill in what is missing". At most 2 repeats; after the second — print
  `🛑`, name what does not match, and stop.
- A step that creates more than one file or overwrites an existing one prints a write plan (file →
  action) before writing and waits for confirmation at a `⏸️` gate (§2.6).

**Where text that is not an action goes:**

| Text | Place |
|------|-------|
| A term definition used by several steps | Body header, once |
| A threshold, a criterion, an anchor example | A reference file (`references/*.md`) |
| A prohibition ("do not invent", "do not overwrite silently") | The "Boundaries" block |
| The shape of the result | The output-structure file (§2.9) |
| The skill's purpose, the rationale for decisions | `description` and the spec; not carried into the body (§2.4) |

Item check: "can I execute this and see the result?" An item that answers "what is X" is a
definition; "what is forbidden" is a prohibition; "why this way" is a rationale. None of the three
stays in the step.

### 2.9. Output-structure file: template vs format reference

A companion file that defines the structure of the produced artifact comes in two **different** kinds.
Mixing the registers is not allowed — it is a recurring source of confusion.

| Kind | What is inside | Name | When |
|------|----------------|------|------|
| **Template (layout)** | Headers/tables of the final document + `<…>` placeholders; at most one orienting line at the top | `templates/<artifact>-template.md` | Output compact and uniform (one screen, fixed blocks) |
| **Format reference (canon)** | A list of sections: what each contains, mandatory/conditional | `templates/<artifact>-format.md` | Output large/variable (structure depends on the input, conditional sections) |

**Iron rule: the file describes the OUTPUT, not the PROCESS.** The fill-in logic — where to take data
from (input sections), what to include/omit, gates, step order — lives **only in `SKILL.md`** (in the
assembly step), not duplicated into the structure file.

**Boundary with examples.** A template is the skeleton of the result, a format reference is the
section makeup, `references/examples.md` (§2.1) is the quality bar of the fill-in; the skeleton is
not repeated in examples, and examples are not moved into the template.

**Litmus test before saving.** Read any line of the file: does it tell Claude *what to do* (a verb:
"extract", "verify", "if…") → that is an instruction, move it to `SKILL.md`. Does it show *what the
result looks like* (a header, a placeholder, a field) → keep it.

### 2.10. Body content rules

- **No time binding.** Do not write "until August 2026 — the old way, after — the new": the text
  will go stale in silence. The current way lives in the main text; the obsolete way — in a folded
  `<details>` block labeled "Obsolete", or delete it.
- **One term per concept** across the whole skill: do not mix "field" and "element", "extract" and
  "get", "spec" and "specification".
- **One default method.** A list of alternatives stalls execution: name the primary tool; a fallback
  only with the condition under which it is needed.
- **Forward slashes in paths** (`scripts/helper.py`, `references/guide.md`), including for Windows.
- **Broken YAML in the frontmatter** — the skill loads with empty metadata: `/name` works,
  auto-invocation does not, the listing has no description. The parse error shows up when launched
  with `--debug`.
- **Project skills** and their `allowed-tools` turn on only after the folder-trust dialogue is
  accepted. Read a skill in a foreign repository before trusting the folder: it can grant itself
  broad permissions.

### 2.11. Verification before deploy

Firing and output quality are checked separately: a skill can load correctly and still produce the
wrong result.

1. Collect 3 realistic requests on which the skill must fire, and 2–3 on which it must not.
2. Run each in a **fresh session**: the skill-development context masks gaps in the text.
3. Repeat the same requests with the skill turned off via `skillOverrides` (`"<name>": "off"`) and
   compare with the run that has the skill.
4. Did not fire on a needed request — fix `description` (§2.3). Fired on an extra one — narrow the
   description or set `disable-model-invocation: true`.
5. The skill enforces discipline (prohibitions, gates, refusals) — run scenarios first without the
   skill, then with it, and compare.

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
- [ ] `description` states what + when, key case first.
- [ ] The skill/subagent body has no references to a companion `spec.md` / dev-artifacts (the body is
      self-contained, §2.4).

**Skill**
- [ ] `name` is at most 64 characters; no words `anthropic` or `claude` (§2.2).
- [ ] `description` — third person; ≤ 1024 characters (standard ceiling); paired with `when_to_use`
      ≤ 1536 characters (listing truncation) (§2.2, §2.3).
- [ ] `SKILL.md` < 500 lines; large reference moved to extra files and linked.
- [ ] Companion links are one level from `SKILL.md`; a reference file > 100 lines opens with a
      table of contents (§2.1).
- [ ] The body is permanent instructions (not one-off steps), without excess narrative.
- [ ] For actions with side effects — `disable-model-invocation: true`.
- [ ] `allowed-tools` — only the tools the skill actually calls; paths in scripts via
      `${CLAUDE_SKILL_DIR}`.
- [ ] Under `context: fork` — the body has an explicit task.
- [ ] For skills with side effects — a "Boundaries" block (explicit prohibitions gathered in one
      place).
- [ ] For multi-step skills — status lines per §2.6 (tag, one emoji per stage, reserved
      `🛑/❓/⏸️/✅`).
- [ ] For skills with >3 steps — a "Checklist" block per §2.7 (plan at the start + `TaskCreate`, one
      task per step, status via `TaskUpdate`; `TaskCreate, TaskUpdate` in `allowed-tools`).
- [ ] Catalog layout: only `SKILL.md` in the root; companions in `references/`, `templates/`,
      `scripts/`, `assets/`, `evals/` (§2.1).
- [ ] The output-structure file is of one register (template `templates/<artifact>-template.md` OR
      reference `templates/<artifact>-format.md`, §2.9); describes the output, not the process.
- [ ] The fill-in logic (data sources, conditions, gates) — only in `SKILL.md`, not duplicated in the
      structure file.
- [ ] Every step is a numbered list of actions per §2.8; items start with an action.
- [ ] Item order matches execution order; loops, branches, and returns are separate items.
- [ ] The skill covers one workflow with one result; several domains are split across skills (§1).
- [ ] The criterion in `→ check` is an observable sign, not a judgment; a step that produces an
      artifact has a failure branch with a return point and a ceiling of 2 repeats (§2.8).
- [ ] A step that creates more than one file or overwrites an existing one prints a write plan at a
      `⏸️` gate (§2.8).
- [ ] `context: fork` is not set on a skill with gates, clarifications, or an interview; the forked
      agent has enough context (`Explore` and `Plan` do not load `CLAUDE.md`) (§2.5).
- [ ] Steps contain no definitions, prohibitions, or rationales — those live in the body header,
      the "Boundaries" block, and reference files.
- [ ] No rule is repeated in two places in the body.
- [ ] The body contains no paragraphs copied from the spec verbatim (§2.4).
- [ ] Paths use forward slashes; the body has no time bindings; terminology is uniform (§2.10).
- [ ] The skill was run on 3 firing scenarios in a fresh session and compared with a run without it
      (§2.11).

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

## Step 1 — [title]
Status: `[skill-name] <emoji> <action>: <dynamic>`

1. **[Action].** [What to do and with what.] [Branch — print `❓` and ask.]
2. **[Action].** [What to do.]
3. Repeat items 1–2 until [condition].

→ check: [observable sign — the file exists, every ID appeared, the section count matches]
→ failed: return to item [N] and fill in what is missing; after 2 repeats — `🛑` and stop (§2.8)

## Step 2 — [title] (write step)
Status: `[skill-name] <emoji> <action>: <dynamic>`

1. **Print the write plan.** [File → action] for each file; wait for `⏸️` confirmation — required
   if there is more than one file or an existing file is overwritten (§2.8).
2. **[Action].** …

→ check: [observable sign]
→ failed: [return point]

## Boundaries (what the skill does not do)
- [Prohibition/boundary — explicit runtime prohibitions, destructive operations, scope limits]

## Further material
- Details: [references/reference.md](references/reference.md)
```

> **The "Boundaries" block.** Required for skills with side effects (writing/editing files, deploy,
> delegation): gather the hard prohibitions in one place (what not to touch, where to stop, what not
> to do silently) instead of scattering them across steps.

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
