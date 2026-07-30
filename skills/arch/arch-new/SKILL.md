---
name: arch-new
description: Design the full architecture and tech stack for a new (greenfield) project — languages, frameworks, libraries with versions verified against primary sources and package-level resolvability verified by a single lockfile resolve. Creates docs/architecture/ (architecture.md, tech-stack.md, adr/). Manual-only.
argument-hint: [product-goals | goals-file]
disable-model-invocation: true
allowed-tools: Read, Write, Edit, Bash, WebFetch, WebSearch, Skill, Agent, TaskCreate, TaskUpdate
---

# Arch New

Designs the architecture and the full stack of a new (greenfield) project: languages, frameworks,
libraries, versions. The result is the canonical `docs/architecture/` set in the target project
(`architecture.md`, `tech-stack.md`, `adr/`) plus a lockfile as proof that the chosen set of
versions installs. Invocation is manual only, via `/arch-new`.

## References

- `${CLAUDE_SKILL_DIR}/../../arch/house-stack.md` — the house stack by layer.
- `${CLAUDE_SKILL_DIR}/../../arch/sources.md` — the registry of primary sources for version
  verification.
- `${CLAUDE_SKILL_DIR}/../../arch/architecture-contract.md` — the artifact contract: the roles of
  the `docs/architecture/` files, the schemas of the `stack`, `resolve`, and `rules` yaml blocks,
  the version-verification procedure. Block schemas come from the contract and nowhere else.

Templates for the output files: `${CLAUDE_SKILL_DIR}/architecture-template.md`,
`${CLAUDE_SKILL_DIR}/tech-stack-template.md`.

## Process indication

Print a status line at the start of every step:

`[arch-new] <emoji> <present-tense action>: <value> [<context>]`

Reserved emoji: `🛑` stop / blocking error · `❓` clarification from the user · `⏸️` waiting for
confirmation (gate) · `✅` successful finish. Working actions take a themed emoji, one per step.

## Gates

A gate (`❓` or `⏸️`) is executed as text: print the status line and the question or the package,
then end the turn; the user's answer arrives in the next message. Do not use `AskUserQuestion`. The
single point of consent to write is the gate in Step 8: before it is confirmed, not one project
file is created or changed.

## Checklist (run plan)

At the start of the run, before Step 1, print the plan and create tasks via `TaskCreate` — one per
step; status via `TaskUpdate`. Do not use `TodoWrite`: it belongs to a different harness, does not
resolve in `allowed-tools`, and the plan would silently degrade into plain text. The list is fixed
and does not depend on the input:

1. Input and context
2. References and preflight
3. Candidates by layer
4. Structural form and rules
5. Version verification
6. Self-check
7. Candidate resolve
8. Confirmation gate
9. Write
10. Summary

Execute strictly in order and mark each task done as you pass it. A conditionally skipped step is
marked done with the note "skipped: <reason>"; the number of tasks always equals the number of
steps in the plan.

## Workflow

### Step 1 — Input and context

Status: `[arch-new] 📥 Gathering input: <source of goals>`

1. Obtain the product goals (purpose, load, target platform, constraints) in this order of
   priority: the command argument (text or a file path), a file named in the conversation, the
   conversation itself. A file was given → read it, extract the goals, show a digest for
   confirmation; missing fields → a targeted `❓` about those fields only. Goals are never
   auto-discovered across the repository — the user names the source. No goals in a file and none
   in the conversation → `❓`; selection does not start.
2. Read the actual state of the repository: manifests, structure, existing code. The repository is
   non-empty → the technologies found enter the candidate stack marked "existing"; replacing them
   requires an explicitly named need, as in Step 3.
3. Check `docs/architecture/`. The documents count as populated if at least one holds: the `stack`
   block in `tech-stack.md` has ≥1 entry; the `rules` block in `architecture.md` has ≥1 entry;
   `adr/` holds at least one ADR. Routes: no folder or files → creation in Step 9; a skeleton
   exists but is not populated (empty files, placeholders) → populate in Step 9 without recreating;
   populated → `❓`: exit and use `/arch-change`, or continue as a deliberate redesign (overwriting
   happens only through the Step 8 gate).

### Step 2 — References and preflight

Status: `[arch-new] 📚 Reading references: house-stack, sources, contract [+ critic subagent]`

Read `house-stack.md`, `sources.md`, and `architecture-contract.md` (paths — the "References"
section). Any of them missing → `❓`: continue without that reference (every affected choice then
goes through explicit user confirmation), or stop and create the references separately. Versions
are never stated from memory in either variant.

The same step checks the one part of the system that lives outside the skills folder — the
`arch-critic-runner` subagent: `test -f ~/.claude/agents/arch-critic-runner.md ||
test -f .claude/agents/arch-critic-runner.md`. Checking it here, before a single `WebFetch`, is what
makes a partial installation surface at the start of the run instead of at Step 6. Missing → `❓`
with two routes:

- install it — copy `arch-critic-runner.md` into `~/.claude/agents/` from wherever the system was
  cloned, and the run continues;
- proceed **without the self-check** — Step 6 is marked "skipped: critic unavailable", and the fact
  travels into the Step 8 gate, the "Consequences" of every ADR of the run, and the summary.

The second route is the absence of a review, never a stand-in for one: an inline critique by this
skill is forbidden — the independence of `arch-critic` rests on context isolation — and no verdict is
recorded as if a review had happened.

### Step 3 — Candidates by layer

Status: `[arch-new] 🗺️ Selecting candidates by layer: <number of layers>`

For each layer, pick a candidate from `house-stack.md` against the goals from Step 1. A candidate
outside the house stack is admissible only for an explicitly named product need the house stack
does not cover → a `⏸️` gate for each such case, before verification; never fold one into the stack
silently. The result is a table of candidates by layer, with existing technologies marked.

### Step 4 — Structural form and rules

Status: `[arch-new] 🧱 Recording the structural form: <form>`

Record the high-level structural form for the goals from Step 1 (monolith/services, the approach to
layering, the chosen patterns) — as prose for `architecture.md`. Formulate checkable structural
rules — entries of the `rules` block per the contract schema: explicitly declared constraints only,
each with an executable read-only `check` command (exit 0 = satisfied); a rule without an
executable command is not recorded. The `check` command follows the form fixed in the contract
(allowed prefix, single command, no shell metacharacters) — a command outside that form will be
refused unexecuted by `arch-review` and `arch-health`, so a rule carrying one is a rule that is
never checked. There may be no rules at all — an empty block is valid. The structural form is a key
decision (it gets an ADR in Step 9).

### Step 5 — Version verification

Status: `[arch-new] 🔎 Verifying versions: <n> technologies`

A hard gate: no version is stated without verification against a primary source. For each
candidate, follow the procedure in `architecture-contract.md` (section "Version verification
against primary sources"):

1. Find the technology's entry in `sources.md`; only sources from there count as proof.
2. No entry → find the official source via `WebSearch`, show the URL to the user, gate `❓`: use it
   and add it to `sources.md`, or reject the technology. `sources.md` is edited only after
   confirmation.
3. `WebFetch` the release page / official registry from the entry.
4. Record: name, current stable version, verification date (today), source URL.
5. Fetch failed or the version is not on the page → `🛑` for that technology: the version is not
   stated, plausible generation is forbidden; the technology stays without a version until
   resolved.

Check: every technology has a version, a date, and a URL — or the status `🛑`. The model's own
knowledge of versions is used only to narrow the candidates; it is not proof.

### Step 6 — Self-check

Status: `[arch-new] ⚖️ Running the self-check [arch-critic-runner]`

Launch the `arch-critic-runner` subagent (the `Agent` tool) with a fresh context: pass the goals
from Step 1, the candidate table with verified versions, and the structural form with the rules
from Step 4; do not pass the reasoning that produced the stack. Collect the verdict report.

Validate the form of what came back before using it: the heading `Decision review: <name>`, the
verdict line, and the table of four axes. A reply without them means the subagent ran without the
`arch-critic` procedure — its `skills:` preload did not resolve, usually because the skill itself is
not installed. Then the run is "self-check not performed": take the second route of Step 2, and never
accept the improvised text as a verdict.

Blocking objections (an unverified version, a dependency without a named need) → return to
Steps 3–5. Debatable ones are kept: unresolved objections go into the Step 8 gate, the summary, and
the corresponding ADR; dropping them silently is forbidden.

### Step 7 — Candidate resolve

Status: `[arch-new] 🔗 Candidate resolve: <ecosystems>`

Follow the "Resolve mechanics" section for the complete set — inside a system temp directory
(`mktemp -d`); project files are not touched. Check: a lockfile was produced with no conflict for
each ecosystem.

### Step 8 — Confirmation gate

Status: `[arch-new] ⏸️ Waiting for confirmation of the package: <n> decisions`

The single point of consent to write. Print the full package and end the turn:

- for each key decision — a structured block following the `adr-write` package fields: decision
  title; context (product goal and circumstances); decision (each version marked
  "verified: YYYY-MM-DD"); alternatives with reasons for rejection (at least one); consequences
  (including the limit of the compatibility proof); affected entries (`target: stack | rules`,
  `name`, `action: add | change | remove | reject`);
- the structural form and the rules;
- the critic's unresolved objections;
- the list of files to be created: `docs/architecture/*`, the ADRs, and the dependency manifest of
  each ecosystem (`package.json`, `pyproject.toml`, `Cargo.toml`, `go.mod`) with its lockfile.

A key decision is: the choice of a technology for a layer (language, framework, storage, a major
infrastructure library), the high-level structural form, and every deviation from the house stack.
Supporting libraries do not get their own ADR — they are covered by the summary ADR "initial stack
composition" (the full table of `stack` entries).

The manifest is in the package because it is the same decision as `tech-stack.md`, written in the
form the ecosystem's resolver reads — the machine-readable carrier of the versions just verified.
Offer two routes in the gate and take the answer together with the confirmation:

- **the skill writes it** — dependencies and their recorded versions only: no scripts, no build
  configuration, no entry points, no code;
- **the user scaffolds it** (`npm create vite`, `uv init`, `cargo new`, `go mod init`) and the run
  resumes at Step 9 once the manifest exists.

Neither route is picked for the user by default. Declining both is the "installability unproven"
route of "Resolve mechanics": the documents and the ADRs are still written, the resolve is not.

The user confirms once, for the whole package. Rejecting the package or part of it → return to
Steps 3–6. Before confirmation, not one project file is created or changed.

### Step 9 — Write

Status: `[arch-new] ✍️ Writing: manifest, docs/architecture/`

The order is fixed:

1. The dependency manifest of each ecosystem, by the route confirmed at the gate: created when the
   project has none, and updated pointwise when one already exists (Step 1.2) — never overwritten
   wholesale. Only dependencies and their recorded versions go in; scripts, build configuration, and
   entry points belong to the product, not to the architecture. On the user's scaffold route the
   manifest is already there, and this substep is the check that it carries the confirmed versions.
2. A final lock-only resolve in the project directory ("Resolve mechanics"); the lockfile stays as
   a deliverable. From its result — `verified` (date) in the `stack` entries and an ecosystem entry
   in the `resolve` block: `lockfile`, `proof` (`lockfile:sha256-<first 12 characters>`, hashed
   with `shasum -a 256`), `resolved` (date).
3. For each confirmed key decision, invoke the `adr-write` skill (the `Skill` tool); last of all,
   the summary ADR "initial stack composition". Take the package fields verbatim from the confirmed
   blocks of the Step 8 gate.
4. Create `docs/architecture/architecture.md` and `docs/architecture/tech-stack.md` from the
   templates `${CLAUDE_SKILL_DIR}/architecture-template.md` and
   `${CLAUDE_SKILL_DIR}/tech-stack-template.md`; in the `stack` and `rules` entries fill the `adr`
   field with the number of the ADR that justifies it.

### Step 10 — Summary

Status: `[arch-new] ✅ Done: <files>, ADRs: <n>`

A short recap: the list of created files, the ADR count; deviations — one line each: the critic's
unresolved objections (by ADR reference), technologies with the status `🛑`, multi-ecosystem
resolve boundaries, the limit of the proof (installability on the date, not semantic
compatibility). Do not repeat the stack and the substance of the decisions in chat — they are in
the files.

The last block of the summary is the hand-off the skill cannot perform itself: the conventions file
of the project (`CLAUDE.md` or `AGENTS.md`) is not edited here, so print the pointer line ready to
paste, followed by the reason it matters —

```
Architecture: `docs/architecture/architecture.md` — the hub. Read it before changing the stack or
the recorded structure.
```

— and state plainly that until that line exists, working sessions will not find the architecture on
their own and `/arch-health` will report it as a "navigation" finding.

## Resolve mechanics

The one mandatory tier of installability proof is a resolve by the ecosystem's toolchain; builds
and smoke runs are not used.

- The whole set of versions for one ecosystem goes into one manifest and one resolver pass;
  checking libraries one at a time proves nothing.
- Lock-only commands, with no installation into the project:

| Ecosystem | Command |
|---|---|
| node | `npm install --package-lock-only` |
| python | `uv lock` |
| rust | `cargo generate-lockfile` |
| go | `go mod download`; run the candidate pass with a temporary `GOMODCACHE`; do not use `go mod tidy` (it scans code imports) |

- Before the first pass of an ecosystem, check that its resolver is on the machine (`command -v npm`
  / `uv` / `cargo` / `go`). On a greenfield project a missing toolchain is the normal case, not an
  anomaly. Absent → `❓` gate naming the ecosystem and the missing command, with three routes: the
  user installs it and the run continues; the candidate is replaced by one from an ecosystem that is
  available; or the run proceeds with **installability unproven** for that ecosystem. Never install
  the toolchain, and never treat a missing resolver as a passed resolve.
- The "unproven" route is recorded, not swallowed: no `resolve` entry is written for that ecosystem,
  the `stack` entries keep their verified versions, and the wording "installability not proven:
  <ecosystem> resolver unavailable on <date>" goes into the "Consequences" of the corresponding ADR
  and into the Step 10 summary. `arch-review` and `arch-health` will report the missing `resolve`
  entry as a finding — that is the intended outcome, and the way back is to run the resolve once the
  toolchain is in place.
- The candidate pass (Step 7) runs in a `mktemp -d` directory; do not clean that directory up by
  hand (no `rm` — the OS handles it). The final pass (Step 9) runs in the project directory, against
  the manifest written in Step 9.1 — there is nothing to resolve before that substep.
- Conflict → read the resolver output, adjust the version, retry. The limit is 3 iterations, then a
  `❓` gate with the conflict text and the options; entries for the conflicting ecosystem are not
  written into `tech-stack.md`.
- Several ecosystems in the stack → a separate resolve for each; compatibility between ecosystems
  is not proven by a resolve — state that honestly in the summary and the ADR.
- The limit of the method: a resolve proves installability on the date of the check, not semantic
  API compatibility. The limit is recorded in the "Consequences" section of the relevant ADRs and
  in the summary.

## Boundaries (what the skill does not do)

- Invocation is manual only, via `/arch-new`. Point decisions for a feature inside a live product
  belong to `/arch-change`, not here.
- No version, API, or compatibility fact is stated without the verification in Step 5; with no
  confirmation the answer is "not verified". Compatibility is not declared proven without a
  resolve.
- Product code is not written. The dependency manifest is the single exception — it is the
  machine-readable form of the recorded versions, it is confirmed at the Step 8 gate like every other
  file, and it holds dependencies and nothing else: no scripts, no build configuration, no entry
  points. An existing manifest is updated pointwise, never overwritten wholesale.
- Dependencies are not installed (`node_modules`, venv) — lock-only resolve only; nothing is
  deployed. Toolchains and resolvers are not installed either: a missing one is a gate for the user,
  never a `brew install` on the skill's initiative.
- ADRs are written only by invoking `adr-write`; writing one directly is forbidden; existing ADRs
  are not edited.
- Files are not deleted; the temp directories of the candidate resolve are not removed (`rm` is
  forbidden). Existing `docs/architecture/*` files are not overwritten without a `⏸️` gate.
- Before the confirmation at the Step 8 gate, project files are not created or changed; the
  candidate resolve happens only in system temp.
- `house-stack.md` and `sources.md` are not edited without the user's confirmation.
- The target project's `CLAUDE.md` (or `AGENTS.md`) is not edited; the user adds the pointer line to
  `docs/architecture/architecture.md` by hand — the skill hands it over ready to paste in Step 10
  instead of leaving the user to discover the gap through an `/arch-health` finding.
