# Methodology — how to work with an AI coding assistant, not just what to configure

The core of OFFICINA: four reusable **guides** for working with Claude Code and your own tooling —
writing style and AI-slop hygiene, CLI authoring, skill and subagent authoring, and loop engineering.
Each is written as a durable reference to read *before* the matching task, not as a tutorial to skim
once. Part of the [OFFICINA](../) repository.

These are working guides distilled from real projects, not abstract advice. Copy what fits, adapt the
conventions to your setup. No absolute paths or secrets are included.

## The question the method starts from

Every task begins the same way: **what should solve this?** Four answers, and picking the wrong one is
the most expensive mistake in the whole workflow — a model where a script would do costs tokens,
latency, and variance; a script where judgment was required produces confident nonsense.

| The task | Artifact | Test |
|---|---|---|
| A deterministic action, no judgment at run time | **CLI** | Can every step be an `if`? |
| A fixed procedure whose decisions depend on context | **Skill** | Does a step need "understand and decide"? |
| An isolated subtask with verbose output, or a whole persona-role | **Subagent** | Do you need only the summary, or a separate tool set and model? |
| A recurring cycle that finds its own work | **Loop** | Is there a real trigger and something able to say "no"? |
| A fact or rule that always applies | **`CLAUDE.md`** | Should it be in context on *every* prompt? |

A CLI may call an agent as a subprocess, but it never stands in for the model's reasoning. A skill
loads into the main context and stays there for the session; a subagent works in its own context
window and returns a summary. `CLAUDE.md` carries facts, never procedures — it's paid for on every
single prompt.

Once the artifact is chosen, the matching guide below covers structure, conventions, and a
pre-deploy checklist.

## The guides

| Guide | What it covers | Read before |
|-------|----------------|-------------|
| [`writing-style.md`](writing-style.md) | Prose rules for chat and documents; a banned list against AI-slop | Producing any prose — a reply, a doc, a commit body |
| [`cli-authoring-guide.md`](cli-authoring-guide.md) | Building CLI tools: boundary, language choice, structure, conventions, dependencies, testing | A task that writes a CLI or utility |
| [`skill-agent-authoring-guide.md`](skill-agent-authoring-guide.md) | Designing skills and subagents: frontmatter, `description` rules, lifecycle, validation checklist | Creating or editing a skill or subagent |
| [`loop-engineering-rules.md`](loop-engineering-rules.md) | Autonomous loops: five steps, six parts, anti-patterns, schedulers, operational discipline | Designing a loop for a project or process |

### `writing-style.md` — hygiene against AI-slop

The shortest guide and the one that applies most often, because every other artifact produces prose.
Its core is a **banned list**, not a style aspiration:

- **Antithesis** in every form — "It's not X, it's Y", "Not X. Y.", "X is dead, Y is the future". Any
  construction that knocks one thing down to prop up another.
- **Dead transitions** — "moreover", "furthermore", "in addition", "that said", "with that in mind".
- **Dead phrases** — "it's important to note", "it's worth noting", "in order to", "in other words",
  "at the end of the day", "simply put".
- **Dead AI vocabulary** — delve, leverage, utilize, paradigm, cutting-edge, revolutionize, crucial,
  underscore, synergy, seamless, holistic, proactive, robust, intuitive, turnkey. "Serves as", "is a
  testament to", and "boasts" become plain "is" and "has".

Around the list sit the positive rules: straight to the point with no preamble, active voice,
numerals as digits, bold sparingly, no closing paragraph restating what was just said, no inflating
significance, and no listing things in threes for rhythm. Chat takes a position rather than hedging;
documents stay impersonal and structural, with no CTAs or hook openers carried over from publication
writing.

### `cli-authoring-guide.md` — deterministic tools

Where the line falls between a script and a model, and everything that follows from being on the
script side of it. Language choice is a decision tree rather than a preference: **bash** up to about
50 lines of gluing commands, **Python** once there are data structures, JSON parsing, or more than
three branches, **Go** when the tool must ship as one static binary or embed its own assets via
`//go:embed`. Node is off by default.

The behavioral conventions are the reusable part: exit codes `0` / `1` / `2` (success, expected error,
usage error) with no silent failures; the result to `stdout` and everything else to `stderr` so output
survives a pipe; a mandatory side-effect-free `--help`; destructive actions behind an explicit flag or
a confirmation, defaulting to dry-run or refusal. Structure is colocation — one folder per tool,
holding its entry point, private logic, tests, and README, with shared code promoted to a common
library only on the *second* real consumer.

Dependencies are zero by default (stdlib only), pinned at the tool level and never globally, wired
through `uv` — PEP 723 inline metadata for single-file tools, a per-tool `venv` for the rest. The
guide ends with copy-ready templates. See [`cli/`](../cli/) for two tools built to it.

### `skill-agent-authoring-guide.md` — skills and subagents

The longest guide, and the one that turns "I wrote a skill and it never fires" into a solved problem.
The answer is almost always the **`description` field** — the only text always in context, and the
sole basis on which the model decides to load anything. It has to state *what* the skill does **and
when to apply it**, in the words a user actually says. "Helps with code" rarely fires; "when
`npm test` fails" fires reliably. Firing too often is the opposite failure, fixed by narrowing the
description or setting `disable-model-invocation: true`.

The rest follows from the **lifecycle**: an invoked `SKILL.md` enters the conversation as one message
and stays until the session ends — the file is never re-read. So the body is written as permanent
instructions, not one-off steps, and every line is a recurring token cost, which is why it says what
to do and skips the narrative of how and why. The body must also be self-contained: no references to
a companion `spec.md` that won't exist next to the deployed file.

Beyond that: full frontmatter reference for both artifact types, `context: fork` and the
`` !`command` `` preprocessing form, mandatory progress indication for multi-step skills, a mandatory
checklist block past three steps, the template-versus-format-reference distinction for output files,
subagent best practices (restrict `tools` to the minimum, choose `model` deliberately, and remember a
subagent never sees the conversation history — critical rules get duplicated into its prompt), and a
**validation checklist to run before deploy**. See [`skills/`](../skills/) and
[`claude/agents/`](../claude/agents/) for the results.

### `loop-engineering-rules.md` — autonomous loops that can stop themselves

Rules for building a loop that discovers, executes, verifies, persists, and reschedules work with no
human in the inner cycle. The premise sets the tone: **the hard part isn't building the loop, it's
putting something inside it able to say "no"**. The cost of a mistake scales with the number of turns
it survives before detection, and a loop by definition maximizes turns — so every safety mechanism
exists to shorten the distance between a mistake and its discovery. Reliability comes from the
quality of the constraints, not the size of the model.

The structure is five mandatory steps per turn — **discovery, handoff, verification, persistence,
scheduling** — implemented by six parts: automations, worktrees, skills, connectors, subagents, and
memory. Discovery lives in a skill rather than a wall of instructions in cron, and it sets the quality
ceiling of the whole loop. Verification means a *second* subagent with different instructions and
sometimes a different model, because one agent acting as both player and judge always favors its own
side. Memory is a file on disk, not context: context resets on compaction, memory survives between
days.

Each missing step has a named failure mode:

| Anti-pattern | Missing step | Symptom |
|---|---|---|
| **Nodding loop** | Verification | Hundreds of turns, never once said "no" to itself |
| **Amnesiac loop** | Persistence | Every morning starts from the same place |
| **Manual loop** | Scheduling | The last run was on demo day |
| **Blind loop** | Discovery | A human still spends the morning deciding what the loop should do |
| **Tangled loop** | Handoff | Under parallelism the edits collide and the merge is a mess |

A rushed loop installs the two that produce visible output — discovery and handoff — and skips the
three that provide safety. The guide also covers choosing a scheduler (cloud, desktop, or an in-session
`/loop`, decided by whether the work is tied to local files), where to put deterministic gates between
creative LLM steps, and the operational discipline that keeps a human in the loop: read a
representative sample every day and be able to explain each change, set hard budget and retry ceilings
*before* the first unattended run, and keep at least one permanent checkpoint where the loop stops and
waits for a person.

The guide adapts the principles of Anthropic's *Loop Engineering* playbook; the wording, structure,
and conventions are original to this repository.

## How the guides relate

`cli-authoring-guide.md` and `skill-agent-authoring-guide.md` are a **pair**: the first draws the line
for deterministic terminal tools, the second for model-driven artifacts. Between them they cover the
decision table at the top of this page. `loop-engineering-rules.md` sits above both — a loop is
assembled out of skills, subagents, and deterministic gates, so it assumes the other two.
`writing-style.md` underpins all three; the others defer to it for tone and prose rather than
repeating the rules.

The artifacts in this repository are these guides applied: [`skills/`](../skills/) and
[`claude/`](../claude/) follow the authoring guide, [`cli/`](../cli/) follows the CLI guide, and
[`jig`](../cli/jig/) scaffolds the very doc structure the guides describe — the method turned into a
binary.

## How to use them

Read the guide that matches the task **before** you start, not after. They're durable references:
scan the relevant section, apply it, and run the validation checklist at the end (where present)
before calling the work done. The checklists are the part worth stealing even if you keep none of the
conventions.

New guides are added over time — one file per guide, one row in the table above.

## FAQ

### When should I write a skill instead of putting the rule in `CLAUDE.md`?

`CLAUDE.md` is loaded on every prompt, so it should hold facts and rules that always apply — repo
layout, conventions, hard prohibitions. A procedure with steps belongs in a skill, which loads only
when it's needed. Putting a multi-step process in `CLAUDE.md` means paying for it on every turn,
including the thousands where it's irrelevant.

### What's the difference between a skill and a subagent?

A skill loads into your main context and stays there for the rest of the session — it changes how the
current agent behaves. A subagent runs in its own context window with its own tool set and model, and
returns only a summary. Use a subagent when the output would be verbose (searches, test runs, logs),
when you need different permissions, or when you want a persona-role. Use a skill when the work
belongs in the conversation you're already having.

### Why doesn't my Claude Code skill fire automatically?

The `description` is the only text in context before invocation, so it's the only thing the decision
can be based on. Rewrite it to state the trigger conditions explicitly — including the phrases a user
would actually type — rather than summarizing what the skill does. If it fires too often instead,
narrow the description or add `disable-model-invocation: true`.

### What words make writing sound AI-generated?

The list in [`writing-style.md`](writing-style.md) covers four families: antithesis constructions
("it's not X, it's Y"), dead transitions ("moreover", "that said"), dead phrases ("it's important to
note", "at the end of the day"), and dead vocabulary (delve, leverage, seamless, robust, crucial,
holistic). Add the structural tells — a preamble before the answer, three-item lists for rhythm, and a
closing paragraph that restates the opening.

### How do I stop an autonomous agent loop from running away?

Three mechanisms, all installed before the first unattended run: a separate evaluator subagent that
can reject the generator's output, hard budget and retry ceilings that turn open-ended risk into
bounded risk, and at least one permanent checkpoint where the loop stops for a human. A loop with no
real check is an agent nodding at itself.

### Are these guides specific to Claude Code?

The skill and subagent guide is Claude Code-specific in its frontmatter details, though the design
rules carry over to any harness reading the `SKILL.md` format. The CLI guide, the writing rules, and
the loop principles are tool-agnostic — they're about where to draw the boundary between deterministic
and probabilistic work, which is the same problem everywhere.

## Related

- [OFFICINA](../) — the method and the rest of the toolkit
- [`skills/`](../skills/) — portable `SKILL.md` procedures built to the authoring guide
- [`claude/`](../claude/) — subagents, `CLAUDE.md` layers, and settings
- [`cli/`](../cli/) — deterministic tools built to the CLI guide

---

**Keywords:** developer methodology, AI coding workflow, how to write Claude Code skills, SKILL.md,
skill vs subagent, CLAUDE.md rules, skill description not firing, subagents, loop engineering,
autonomous agent loops, generator and evaluator, agent anti-patterns, agent scheduler, CLI authoring,
CLI conventions, exit codes, writing style, AI slop, AI writing tells, banned words list.
**Topics:** `claude-code` · `methodology` · `claude-code-skills` · `subagents` · `developer-tools`
