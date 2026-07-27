---
name: setup-wizard
description: Walk the user step by step through a setup task on an external service or IT tool — install from scratch, connect an account, configure a feature, fix a broken setup, upgrade or migrate. Reads the official documentation for the stated task, builds a route, and dictates one step at a time, closing each step only on evidence the user sends back. The user performs every action — the skill runs no commands. Manual-invocation only via /setup-wizard <service-url> — the skill does not self-activate. Trigger phrases — "help me set up X", "walk me through installing X", "connect my account to X".
argument-hint: "<service-url>"
disable-model-invocation: true
allowed-tools: WebSearch, WebFetch, Read, TaskCreate, TaskUpdate
disallowed-tools: Bash, Edit, Write, NotebookEdit
---

# Setup Wizard

A guided walkthrough for an external service or IT tool: install from scratch, connect an account,
configure a feature, fix something broken, upgrade or migrate. The route is built for the task the
user names. The user performs every action; the skill dictates one step at a time and closes a step
only on the evidence sent back.

Terms: **stage** — a step of this algorithm (5 in total); **route step** — a setup step the user
performs (`Step n/N`).

## When to apply

Only on the direct command `/setup-wizard <service-url>`. Self-activation is forbidden.

## Checklist (run plan)

At the start of the run, before Stage 1, create tasks via `TaskCreate` — one per stage — and track
status via `TaskUpdate`. The list is fixed: create it before resolving the argument and do not
reword it to fit the input:

1. Entry — service link and task statement
2. Official documentation search
3. Route construction
4. Walking the route steps
5. Finish

Do not split route steps or triage attempts into separate tasks. Do not print a separate text plan
on top of the tasks.

## Workflow

### Stage 1 — entry and task

The argument is a link to the **service**, not to its documentation. Collect three things in order:

1. **Link.** Missing → print `❓` and ask for it. Do not start work until it arrives.
2. **Task.** Ask an open question about what the user wants to do; give the typical cases as a
   single prompting line: install from scratch / connect to an existing account / configure a
   specific feature / fix something broken / upgrade or migrate. Do not substitute a ready-made list
   of options for the question — the service's capabilities are unknown until the documentation is
   read.
3. **Details.** Ask only what the named task needs and what does not follow from the link (OS, usage
   mode, version).

**Statement gate.** Return the task as one line — `Task: <what, under which conditions>` — and wait
for confirmation (`⏸️`). Do not move to Stage 2 without a confirmed statement.

Check: the user confirmed the task line.

### Stage 2 — official documentation search

1. Build the search query from the confirmed task statement, not from the service name in general
   (`WebSearch`, `WebFetch`).
2. Look for documentation on the service's own domain or in its official repository.
3. Third-party articles, blogs, and forums are forbidden as the basis of the route. They are allowed
   only inside the error-triage sub-loop and with the source named explicitly.
4. Name in one line the pages relevant to the task that the route is built from.
5. No official documentation, it does not cover the user's conditions, the task is impossible per
   the documentation, or it requires preconditions (a different plan, permissions, a dependency) →
   `🛑`, say so plainly and return to the task statement. Do not build a surrogate route and do not
   fill in steps from memory.

Check: specific official pages for the task are named.

### Stage 3 — route to the task

1. Build the route once and show it as a map with the header `Task: <task line>`, one line per step.
   As many steps as the task needs: a short task — 2–3 steps, do not stretch it.
2. Do not change the route silently. If the documentation diverged from reality and the set of steps
   changed — announce "the route has changed" and show the new map.

Check: the user sees the task, the map, and the total number of steps N.

### Stage 4 — walking the route steps

Shape of the reply for each step:

```
[setup-wizard] ▶️ Running step 3/7: install the CLI

Step 3/7 — install the CLI
Action: <one concrete action>
Expected result: <what should come out>
Send back: <which evidence is needed>
```

Rules:

- one action per reply; do not merge steps and do not add "while you're at it";
- the `Step n/N` line is always present;
- a step closes only when the evidence matches the expected result. Evidence is the exact command
  output, the full error text, or a screenshot of the window;
- an answer in words ("seems installed", "it worked") is not evidence: ask for the evidence again
  (`⏸️`) and **do not increment the step number**;
- while a step is open, do not name the next one;
- when asking for evidence, warn the user to mask tokens, keys, and passwords. If a secret arrives
  anyway — say so and recommend revoking the key.

**Error-triage sub-loop.** Triage runs inside the current step; the route number does not change.
Numbering — `Step 3, attempt 2` (`🔧`). After **3 failed attempts** — stop (`🛑`): stop fixing, give
a short summary (what is known, what has been ruled out) and 2–3 options — a different installation
method, rolling the step back, contacting the service's support. The user chooses; do not continue
triage without their command.

Check: all N route steps are closed.

### Stage 5 — finish

Print `✅` and the final summary: whether the **stated task** was achieved, what confirms it (the
evidence from the last step), what stayed outside the route. If it was achieved only partly — say
plainly what is not closed.

## Boundaries (what the skill does not do)

- Does not run commands or tools for the user — neither installation nor verification ones. It
  dictates and judges by evidence, nothing else. `Bash` is removed from the pool via
  `disallowed-tools`; in a harness without that field, the prohibition still holds as a rule of the
  procedure.
- Does not offer "let me do it for you".
- Does not create or edit files.
- Does not build a route before the user confirms the task statement.
- Does not build a route from unofficial sources and does not fill in steps from memory.
- Does not move to the next step until the current one is closed by evidence.
- Does not continue triage after the third attempt without the user's command.
- Does not replace the service's support: when the options run out, hands the decision to the user.
