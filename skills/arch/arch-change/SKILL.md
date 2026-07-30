---
name: arch-change
description: Architectural delta (stack and declared structure) for a feature in an existing product. Default is reusing the current stack; a new dependency, replacement, or version bump is allowed only for a named product need the stack does not cover. Verifies versions against primary sources, verifies package-level resolvability with a lock-only resolve, records the delta in tech-stack.md, architecture.md and an ADR. Manual invocation only.
argument-hint: [feature-description]
disable-model-invocation: true
allowed-tools: Read, Write, Edit, Bash, WebFetch, WebSearch, Skill, Agent, TaskCreate, TaskUpdate
---

# Arch Change

A point architectural decision for a feature inside a live product. Input — a feature description
(argument or conversation). Output — a delta against the recorded architecture: updated
`tech-stack.md` entries, an ADR, a lockfile; `architecture.md` only when the delta is structural.
The full set of architecture files is not recreated.

The target files are `docs/architecture/` and the project manifests in the current session
directory (cwd).

## References

| File | Role |
|---|---|
| `${CLAUDE_SKILL_DIR}/../../arch/architecture-contract.md` | the artifact contract: file roles, schemas of the `stack`/`resolve`/`rules` blocks, the compact version-verification procedure |
| `${CLAUDE_SKILL_DIR}/../../arch/house-stack.md` | the house stack — the default choice space for candidates |
| `${CLAUDE_SKILL_DIR}/../../arch/sources.md` | the registry of primary sources for version verification |

The skill has no other references; do not use other paths.

## Principles

- **Build on what exists.** The default is reusing the current stack. A new dependency, a
  replacement, or a version bump is allowed only for a specific named need the stack does not
  cover. The insufficiency of what exists is proven before anything new is proposed.
- **Renovate's territory.** A version bump is this skill's delta only when it leaves the recorded
  decision (the exact version or the range in the `version` field of a `stack` entry). Movement
  inside the recorded range (patches, minors under `^`) is maintenance: no `arch-change`, no ADR.
- **Versions come from primary sources only.** A version is not stated as reliable without
  confirmation from a source in `sources.md` (procedure — the compact section in
  `architecture-contract.md`). The model's own knowledge is good only for narrowing candidates.
- **A delta, not a recreation.** Files are updated pointwise; entries the feature does not touch
  are left alone.

## Process indication

Every step opens with a status line:

`[arch-change] <emoji> <present-tense action>: <value>`

Reserved emoji: `🛑` stop / blocking error · `❓` clarification from the user · `⏸️` confirmation
gate · `✅` finish. Working actions take a themed emoji, one per step (`📥 🎯 🔎 🧩 🔍 🧪 ✍️`).

## Gates

`🛑`, `❓`, `⏸️` mean a text status line with the necessary context and **the end of the turn**: the
skill prints the question or the package and stops; the answer arrives in the user's next message.
Do not use `AskUserQuestion`. Project files are not changed before the confirmation at a gate.

## Checklist (run plan)

At the start of the run, before Step 1, print the plan and create `TaskCreate` tasks — one per
step; status via `TaskUpdate`, executed strictly in order. Do not use `TodoWrite`: it belongs to a
different harness and does not resolve in `allowed-tools`. The list is fixed and does not depend on
the input:

1. Reading the architecture and the reality
2. The need and the delta classification
3. Sufficiency of what exists
4. Candidates and the structural delta
5. Self-check
6. Candidate resolve
7. Confirmation gate
8. Write
9. Summary

A conditionally skipped step (for example, the resolve on a purely structural delta) is marked done
with the note "skipped: <reason>"; the number of tasks is always nine.

## Workflow

### Step 1 — Reading the architecture and the reality

Status: `[arch-change] 📥 Reading the architecture and the manifests: <project>`

1. Read `docs/architecture/architecture.md` (the hub), and through its links — `tech-stack.md` and
   the ADRs relevant to the feature. Entry is always through the hub.
2. Read the project manifests (`package.json` / `pyproject.toml` / `Cargo.toml` / `go.mod`) — that
   is the reality.
3. Validate the contract (schemas — `architecture-contract.md`):
   - no `architecture.md` → `🛑` the architecture is not recorded; recommend running `/arch-new`;
     finish, touch no files;
   - the hub exists but `tech-stack.md` is missing, or the `stack`/`resolve` blocks do not parse →
     `🛑` with a report on the broken contract and the routes: manual restoration, or `/arch-new`
     as a deliberate redesign; finish, touch no files.
4. `house-stack.md` or `sources.md` unavailable → `❓`: continue without it (then every choice goes
   through explicit confirmation), or stop. Check the `arch-critic-runner` subagent in the same
   step — `test -f ~/.claude/agents/arch-critic-runner.md ||
   test -f .claude/agents/arch-critic-runner.md` — before any `WebFetch`, so a partial installation
   surfaces at the start of the run rather than at Step 5. Missing → `❓` with two routes: copy the
   file into `~/.claude/agents/` and continue, or proceed **without the self-check** — Step 5 is then
   marked "skipped: critic unavailable" and the fact travels into the Step 7 gate, the ADR
   "Consequences", and the summary. An inline critique by this skill is not one of the routes: the
   independence of `arch-critic` rests on context isolation.
5. For the layers the feature touches, compare intent (`tech-stack.md`) against reality (the
   manifests). Drift found → `❓`: fold the reconciliation into the current delta, or continue on
   top of the named drift (and then record it in the summary). Continuing silently is forbidden.

### Step 2 — The need and the delta classification

Status: `[arch-change] 🎯 Formulating the need: <feature>`

1. State which product goal the feature serves and what it therefore requires of the architecture.
   The goal is not named in the request and does not follow from the context (for example, "add X"
   with no need) → `❓` about the product goal; work stops until it is answered.
2. Classify the delta:
   - **stack** — dependencies, versions;
   - **structural** — the recorded form, `rules` entries;
   - **mixed**;
   - **none** → `✅` finish with no write, and the reasoning in chat.

### Step 3 — Sufficiency of what exists

Status: `[arch-change] 🔎 Checking the stack's sufficiency: <need>`

1. The stack side: does the current stack cover the need — the installed libraries and their
   capabilities per the `tech-stack.md` entries and the documentation from `sources.md`. It does →
   the decision is "no new dependencies".
2. The structural side: does the need require changing the recorded form or the rules.
3. Routes:
   - the request carried a candidate to add or replace and it was rejected, or the decision changes
     how something recorded earlier is read → **the "rejection" route**: Steps 5, 7, 8 (an ADR of
     rejection via `adr-write`, stack entries unchanged; skip Step 6 with a note);
   - the status quo is confirmed — no changes, no rejected candidates, no structural delta → `✅`
     finish with no write, and the reasoning for sufficiency in chat;
   - the stack is insufficient, or there is a structural delta → Step 4.

### Step 4 — Candidates and the structural delta

Status: `[arch-change] 🧩 Selecting candidates: <layer/need>`

1. A stack delta takes candidates from `house-stack.md` (the default choice space).
2. A candidate outside the house stack is admissible only for an explicitly named need → a `⏸️`
   gate: an explicit **"outside the house stack"** mark, the reasoning, and a wait for the user's
   confirmation. Without confirmation the candidate is dropped.
3. Verify each candidate's version by the procedure in `architecture-contract.md` (the compact
   "Version verification against primary sources"). A hard gate:
   - only sources from `sources.md` count as proof; no entry → `WebSearch` for the official source,
     show the URL, `❓`: use it and add it to `sources.md` (edited only after confirmation), or
     reject the technology;
   - `WebFetch` the release page; record the name, the stable version, the verification date, and
     the source URL;
   - fetch failed or the version was not found → `🛑` for that technology: the version is not
     stated, plausible generation is forbidden.
4. The structural delta: state the change to the recorded form and/or to the `rules` entries. Rules
   are only those explicitly declared and checkable: each with an executable read-only `check`
   command (exit 0 = satisfied) and an `adr` field (the `rules` block schema is in
   `architecture-contract.md`). The `check` command follows the form fixed in the contract (allowed
   prefix, single command, no shell metacharacters); a command outside that form is refused
   unexecuted by the auditing skills, so such a rule is never checked and is not worth recording.

### Step 5 — Self-check

Status: `[arch-change] 🔍 Running the delta review: arch-critic-runner`

1. Launch the `arch-critic-runner` subagent via `Agent`: pass the feature's need (Step 2) and the
   whole delta — the stack side and the structural side (the decisions from Steps 3–4). A fresh
   context is mandatory: the critic must not see the reasoning of the delta's author — pass only
   the need and the resulting delta, with no discussion history.
2. Validate the form of the reply before using it: the heading `Decision review: <name>`, the
   verdict line, and the table of four axes. Missing → the subagent ran without the `arch-critic`
   procedure (the `skills:` preload did not resolve, usually because the skill is not installed);
   the run counts as "self-check not performed" by the second route of Step 1.4, and the improvised
   text is not accepted as a verdict.
3. Take the verdict into the delta. Every objection is either resolved by amending the delta (return
   to Step 4) or carried into the Step 7 gate, the ADR, and the summary with an explicit
   **"review objection, unresolved"** mark. Ignoring objections silently is forbidden.

### Step 6 — Candidate resolve

Status: `[arch-change] 🧪 Verifying installability: <ecosystem>`

Only for a stack delta; a purely structural delta → mark the task "skipped: structural delta" and
move to Step 7.

1. Copy the project manifest(s) into a system temp directory (`mktemp -d`); do not edit the real
   manifest.
2. Apply the delta to the copy and run **a single lock-only pass over the whole manifest**:
   installability is proven for the complete dependency set, never library by library. Commands by
   ecosystem:
   - node: `npm install --package-lock-only`
   - python (uv): `uv lock`
   - rust: `cargo generate-lockfile`
   - go — a special case: `go mod download` (the lockfile is `go.sum`); do not use `go mod tidy` —
     it scans code imports. Isolate the candidate pass with a temporary `GOMODCACHE`; nothing is
     installed into the project.
3. Install nothing: no `node_modules`, no venv, no binaries — and no toolchains. Before the pass,
   check that the ecosystem's resolver is on the machine (`command -v npm` / `uv` / `cargo` / `go`).
   Absent → `❓` gate naming the ecosystem and the missing command, with three routes: the user
   installs it and the run continues; the candidate is replaced by one from an available ecosystem;
   or the delta proceeds with **installability unproven**. A missing resolver is never treated as a
   passed resolve.
4. The "unproven" route is recorded, not swallowed: the `resolve` entry of that ecosystem is left
   untouched in Step 8, and the wording "installability not proven: <ecosystem> resolver unavailable
   on <date>" goes into the ADR "Consequences", into the Step 7 gate package, and into the Step 9
   summary.
5. Conflict → read the resolver output, adjust the version, retry. The limit is 3 iterations, then
   `❓` with the resolver output and the question of what to do.

### Step 7 — Confirmation gate

Status: `[arch-change] ⏸️ Waiting for confirmation of the delta package`

The single point of consent to write. Print the full package as one block:

1. Structured decision blocks — following the `adr-write` input contract fields: title, context,
   decision (with version verification dates), alternatives, consequences, affected entries, the
   superseded ADR (if any).
2. The structural part: the change to the form, the `rules` entries.
3. Unresolved review objections (marked as such).
4. The list of files to be changed: manifest, lockfile, `tech-stack.md`; `architecture.md` on a
   structural delta; the new ADR.

The "rejection" route: the package is an ADR of rejection; the file list is the new ADR only.

Confirmation is a single one for the whole package, in the user's next message. Rejection → return
to Steps 3–6, or finish with no write. Project files are not changed before confirmation.

### Step 8 — Write

Status: `[arch-change] ✍️ Writing the delta: <files>`

**The "rejection" route:** leave the manifest, the lockfile, and the `stack` and `resolve` entries
alone. Invoke only `adr-write` via `Skill` with `action: reject` — an ADR of rejection where the
alternative is the rejected candidate with the reason it was refused.

**The delta route** — the order is fixed:

1. Apply the delta to the real manifest; run the final lock-only resolve in the project directory
   (commands — Step 6). The resulting lockfile is an artifact. On the "installability unproven"
   route of Step 6 the manifest is still updated, the resolve is skipped, and the `resolve` entry of
   that ecosystem stays as it was — an entry is never written from a resolve that did not run.
2. Invoke the `adr-write` skill via `Skill`; take the fields verbatim from the confirmed gate
   block.
3. In `tech-stack.md`, update only the affected entries of the `stack` block (fields
   `layer, name, kind, version, verified, source, adr`; `verified` — the verification date, `adr` —
   the number of the ADR just created) and the entry of the affected ecosystem in the `resolve`
   block: `lockfile`, `proof` — `lockfile:sha256-<first 12 characters>` from `shasum -a 256`,
   `resolved` — today's date. Schemas — `architecture-contract.md`.
4. Edit `architecture.md` if and only if the delta changes a specific entry of its normative
   content: it adds or changes a `rules` entry (with an `adr` field), or it changes the recorded
   structural form. The litmus test: name the rule `arch-review` should check once the feature
   lands; no rule can be named and the form does not change → leave the file alone.

### Step 9 — Summary

Status: `[arch-change] ✅ Delta recorded: ADR <number>`

The recap in chat: the changed files, the ADR number, the substance of the delta, the resolve proof
(the lockfile hash); and the drift named in Step 1, if the user chose to continue on top of it. Do
not list entries the feature did not touch.

## Boundaries (what the skill does not do)

- Does not recreate `docs/architecture/` wholesale and does not edit stack entries the feature does
  not touch.
- Does not install dependencies: lock-only resolve, no `node_modules` / venv / binaries. Does not
  install toolchains either — a missing resolver is a gate for the user, not a task for the skill.
- Does not write the feature's code and does not judge code quality.
- Does not fix the drift it finds between intent and reality — it only names it.
- Does not write to project files before the confirmation at the gate (Step 7); writing happens
  only in Step 8; the candidate resolve runs on a copy in temp.
- Does not state a version as reliable without confirmation from a primary source; sources
  unavailable → `🛑`.
- Does not propose candidates outside `house-stack.md` without a named need and a `⏸️` gate.
- Does not handle bumps inside the recorded version range — that is Renovate's territory.
- Does not use `AskUserQuestion`: every gate is a status line and the end of the turn.
- Is not invoked by the model on its own initiative.
- Does not edit the target project's `CLAUDE.md` (or `AGENTS.md`); the user maintains the
  navigation pointer line by hand.
- Bash: does not use `git -C <path>`; avoids compound commands for read-only tasks.
