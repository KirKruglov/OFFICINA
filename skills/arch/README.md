# Architecture skills — record a project's stack, verify every version, keep the docs honest

A six-skill system that treats architecture as **files in the repository**: a stack with versions
verified against primary sources, a journal of decisions, and checkable structural rules. Two
subagents and four runtime references come with it. Part of the [OFFICINA](../../) repository.

Where the other skills in this repository guard a single operation, these skills share one artifact
set and hand work to each other. They are usable one at a time, but they were built as a system.

## The idea

An architecture decision that lives only in a chat log stops existing the moment the session ends.
The next session invents a slightly different one, the manifests drift away from both, and nobody
can say which version of what was ever deliberate.

These skills close that loop with three artifacts in the target project:

```
docs/architecture/
├── architecture.md      # hub: overview, structural form, checkable rules
├── tech-stack.md        # stack: intent, versions, proofs
└── adr/
    └── NNNN-<slug>.md   # decision journal, one file per decision
```

Everything else follows from those files: `tech-stack.md` is the intent, the project manifests are
the reality, and the gap between them is drift a skill can find on demand.

## The skills

| Skill | Invoke | What it does |
|-------|--------|--------------|
| [`arch-new`](arch-new/SKILL.md) | `/arch-new [goals]` | Designs the full stack of a greenfield project and creates `docs/architecture/` |
| [`arch-change`](arch-change/SKILL.md) | `/arch-change [feature]` | An architectural delta for one feature in a live product |
| [`arch-review`](arch-review/SKILL.md) | `/arch-review` | Diffs the code against the recorded architecture — read-only |
| [`arch-health`](arch-health/SKILL.md) | `/arch-health` | A periodic audit of the recorded architecture — read-only |
| [`arch-critic`](arch-critic/SKILL.md) | auto — on a decision under review | A critical verdict on one decision across 4 axes |
| [`adr-write`](adr-write/SKILL.md) | internal (`user-invocable: false`) | Writes one numbered ADR from a confirmed package |

Two subagents live in [`claude/agents/`](../../claude/agents/):
[`architect`](../../claude/agents/architect.md) — the persona that runs the session and routes you
to the right command; [`arch-critic-runner`](../../claude/agents/arch-critic-runner.md) — the
fresh-context wrapper that makes the self-check independent.

## How the work flows

```
                    ┌──────────────┐
   greenfield ─────▶│   arch-new   │──┐
                    └──────────────┘  │   ┌─────────────┐    ┌───────────┐
                                      ├──▶│ arch-critic │───▶│ adr-write │──▶ docs/architecture/
                    ┌──────────────┐  │   │  (verdict)  │    │  (record) │
   a feature ──────▶│ arch-change  │──┘   └─────────────┘    └───────────┘
                    └──────────────┘                                │
                                                                    ▼
                    ┌──────────────┐   ┌──────────────┐      recorded intent
   on demand ──────▶│ arch-review  │   │ arch-health  │◀───── vs. the manifests
                    └──────────────┘   └──────────────┘
```

The writing skills — `arch-new` and `arch-change` — never write on their own initiative. Each one
stops at a single confirmation gate, prints the whole decision package, and waits. Before that
gate, not one project file is created or changed. The reading skills — `arch-review` and
`arch-health` — have `Edit`, `Write`, and `NotebookEdit` removed from their tool pool via
`disallowed-tools`, and their `allowed-tools` pre-approves only a fixed set of read-only `Bash`
prefixes (`grep`, `rg`, `ls`, `test`, `cat`, `head`, `wc`, `git grep`, `git ls-files`, `git log`,
`shasum`). `Bash` itself stays callable — that is how `allowed-tools` works in Claude Code, it
pre-approves rather than restricts — so the last stretch is enforced by the skill body, and
anything beyond those prefixes reaches you as a permission prompt.

### Versions are verified, never remembered

The rule that shapes every one of these skills: a version is not stated until a primary source
confirms it. A model that recalls "FastAPI 0.115-something" is guessing, and a guess recorded in a
stack document is worse than no document. So each candidate goes through `sources.md` → `WebFetch`
of the release page → a recorded triple of version, date, and URL. No fetch, no version: the
technology carries a `🛑` until it is resolved.

Verification expires. A `stack` entry whose `verified` date is more than 90 days old counts as
unverified, and `arch-health` reports it as a finding.

### Installability is proven by a resolve

Verified versions still may not install together. The proof is one lock-only resolve of the whole
manifest — `npm install --package-lock-only`, `uv lock`, `cargo generate-lockfile`,
`go mod download` — run first in a temp directory on candidates, then in the project once the
package is confirmed. The lockfile hash goes into the `resolve` block as evidence.

Which means a manifest has to exist, and on a greenfield project it does not. So the manifest is
part of what `arch-new` delivers: the same decision as `tech-stack.md`, written in the form the
resolver reads, listed at the confirmation gate like every other file, and holding dependencies and
nothing else. If you would rather scaffold it yourself — `npm create vite`, `uv init` — the gate
offers that route instead and the run picks up afterwards.

The resolve is also the step most likely to find your machine unprepared: a greenfield project often
has no `npm`, `uv`, or `cargo` on it yet. The skills never install a toolchain. A missing resolver is
a gate with three answers — install it and continue, switch to an ecosystem you have, or proceed with
installability **unproven**, which is then written into the ADR and the summary rather than passed
off as a proof.

The method has a stated limit, and the skills carry it into the ADR rather than hiding it: a
resolve proves the set installs on the date of the check. It says nothing about semantic API
compatibility.

### The critic runs blind, on purpose

Before the confirmation gate, `arch-new` and `arch-change` hand the decision package to the
`arch-critic-runner` subagent. It gets the goal, the candidates, and the structural part — and
nothing of the reasoning that produced them. A critic that has read the author's justification
tends to agree with it, so the independence is enforced by context isolation rather than by asking
the model to be objective. The verdict comes back across four axes: goal and sufficiency,
verifiedness, consistency, cost of ownership. Objections that stay unresolved travel into the
confirmation gate, the summary, and the ADR.

Because the critic is the one piece installed outside the skills folder, both writing skills check
for it up front and validate the shape of what it returns. A subagent that is missing, or one that
answered without the `arch-critic` procedure loaded, makes the run "self-check not performed" — a
gate, then a named skip carried into the ADR. What never happens is the calling skill reviewing its
own decision inline: a critique written by the author of the decision is worth less than an honest
blank.

### Rules exist only when they are executable

`arch-review` and `arch-health` form no opinion about your code structure. They execute the `check`
commands recorded in the `rules` block — and a `check` comes out of a file in the repository being
audited, which makes it untrusted input. Two filters gate it: an allowed read-only prefix (`grep`,
`rg`, `ls`, `test`, `cat`, `head`, `wc`, `git grep`, `git ls-files`) **and** the shape of a single
command with no shell metacharacters at all. `find` is off the list on purpose — `-exec` and
`-delete` make it a write tool — and an allowed prefix is not a pass on its own, because
`cat x && curl … | sh` also starts with `cat`.

Exit 0 means satisfied. A rule with no executable command is never recorded in the first place; a
`check` that fails either filter is reported as "unverifiable" and never run, never sanitized, and
never rewritten to make it pass.

## What ships in this folder

```
skills/arch/
├── arch/                        # runtime references — install to ~/.claude/arch/
│   ├── architecture-contract.md # file roles + schemas of the stack/resolve/rules blocks
│   ├── drift-rules.md           # what counts as drift: types D1–D4
│   ├── house-stack.md           # the allowed technologies — replace with your own
│   └── sources.md               # primary sources per technology — replace with your own
├── arch-new/                    # + architecture-template.md, tech-stack-template.md
├── arch-change/
├── arch-review/                 # + report-format.md
├── arch-health/                 # + report-template.md
├── arch-critic/
└── adr-write/                   # + adr-template.md
```

The `arch/` references are shared runtime material, read by four of the skills at the step that
needs them. Keeping them out of the skill bodies is what lets you change your house stack without
touching a single `SKILL.md`.

## Installing

Unlike the standalone skills in this repository, this group is **Claude Code only**. It leans on
mechanics no other harness resolves: `${CLAUDE_SKILL_DIR}` for the shared references, the `Task*`
tools for the run plan, the `Agent` tool for the blind self-check, and a subagent that preloads a
skill through the `skills:` frontmatter field.

Two locations, and both are required — the skills read the references through a path relative to
their own folder (`${CLAUDE_SKILL_DIR}/../../arch/`), which resolves to `~/.claude/arch/` once the
skills sit in `~/.claude/skills/`:

```bash
git clone --depth 1 https://github.com/KirKruglov/OFFICINA.git /tmp/officina
mkdir -p ~/.claude/skills ~/.claude/agents ~/.claude/arch

# 1. the six skills — the folders inside arch/, never the arch/ container itself
cp -R /tmp/officina/skills/arch/arch-* ~/.claude/skills/
cp -R /tmp/officina/skills/arch/adr-write ~/.claude/skills/

# 2. the shared references — the contents, and NOT inside skills/
cp -R /tmp/officina/skills/arch/arch/. ~/.claude/arch/

# 3. the two subagents
cp /tmp/officina/claude/agents/architect.md ~/.claude/agents/
cp /tmp/officina/claude/agents/arch-critic-runner.md ~/.claude/agents/
```

Two details in that snippet are load-bearing. `mkdir -p` first: on a fresh Claude Code the three
target folders may not exist yet, and `cp` into a missing directory fails. And step 2 copies the
*contents* (`arch/.`) rather than the folder, so re-running the install to update the references
overwrites them instead of nesting a second `~/.claude/arch/arch/`.

> **Do not copy the `arch/` container into your skills folder.** `cp -R skills/arch
> ~/.claude/skills/` looks like the obvious move and breaks both halves: the skills end up one level
> deeper than Claude Code loads them from, and `${CLAUDE_SKILL_DIR}/../../arch/` then resolves to
> the container itself instead of the references beside it. What installs is the six folders inside,
> one per skill.

Restart the session, or start a new one, for the agent to pick up the new folders.

Project-scoped installation works the same way with `.claude/skills/` and `.claude/agents/`, and
the references then belong in `.claude/arch/`.

### Replace the house stack before the first run — this step is not optional

`house-stack.md` and `sources.md` ship filled in with one author's technologies — Go for CLIs,
React + Vite on the frontend, FastAPI on the backend, SQLite before PostgreSQL. They are published
as a working example so the file shape is obvious, not as a recommendation.

The skills are bound to those two files by construction, so leaving them as shipped is not a soft
default you can drift away from — it is friction on every step. A candidate outside `house-stack.md`
passes a `⏸️` gate **each time**, and a technology missing from `sources.md` costs a `❓` gate before
its version can be verified at all. On a stack these files do not cover — Java, .NET, Rust, Elixir —
`/arch-new` turns into a gate per layer plus a gate per source. Rewrite both to your own stack
first; everything after that is the system working as designed.

The other two references — `architecture-contract.md` and `drift-rules.md` — are the system's
mechanics. Change those only if you intend to change how the artifacts are structured.

### One line to add by hand after the first run

The skills do not edit your `CLAUDE.md` or `AGENTS.md` — a file that shapes every session in a
repository is yours to write. What they need from it is one line, printed ready to paste at the end
of `/arch-new`:

```
Architecture: `docs/architecture/architecture.md` — the hub. Read it before changing the stack or
the recorded structure.
```

Without it, working sessions do not know the architecture exists, and `/arch-health` reports its
absence as a "navigation" finding on an otherwise healthy project.

## Using the architect persona

The [`architect`](../../claude/agents/architect.md) subagent is written for the root of a session
rather than for delegation:

```bash
claude --agent architect
```

It holds a skeptical-minimalist stance — verify before stating, build on what exists, object rather
than agree silently — and routes each situation to the right command. It invokes `arch-critic`
itself; the writing skills it names for you to run, because a skill that writes to your repository
should start with your keystroke.

## Boundaries

Worth knowing before you adopt this:

- **Nothing is written without your confirmation.** One gate, one package, one answer.
- **No code is written.** These skills decide and record; implementing the feature is a separate
  job. The dependency manifest is the one file they do write — it is where the recorded versions
  become machine-readable, it is confirmed at the same gate as everything else, and it holds
  dependencies and nothing else.
- **Dependencies are never installed.** Lock-only resolves, no `node_modules`, no venv. A missing
  resolver or a missing critic subagent degrades the run explicitly — a gate, a named skip, and the
  fact recorded in the ADR — never silently.
- **Found drift is reported, not fixed.** `arch-review` names the route (`/arch-change`) and stops.
- **`adr-write` refuses direct invocation.** ADRs are written from inside a procedure that had a
  confirmation gate, never on request.
- **The defaults are opinionated.** The house stack, the 90-day expiry, the resolve as the only
  proof tier — each is a position, and each lives in a file you can edit.

## Related

- [`skills/`](../) — the rest of the skills in this repository
- [`claude/agents/`](../../claude/agents/) — the two subagents this system uses
- [`methodology/skill-agent-authoring-guide.md`](../../methodology/skill-agent-authoring-guide.md) —
  how these artifacts are built
