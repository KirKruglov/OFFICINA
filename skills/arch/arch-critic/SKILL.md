---
name: arch-critic
description: Critical review of a single architecture decision (proposed or already made) against product goal, verified versions, stack consistency, and cost of ownership. Read-only - returns a structured verdict in chat, never edits files. Use when arch-new or arch-change run a self-check of a candidate decision in a fresh-context subagent (self-check runs only as a subagent, never inline), or when the user asks to review an architecture choice. Triggers - "review this decision", "check this choice", "assess this architecture decision", "critique the architecture".
allowed-tools: Read, Grep, Glob
disallowed-tools: Edit, Write, NotebookEdit
---

# Arch Critic

A critical review of one architecture decision — proposed or already made. The skill judges the
decision and returns a verdict; writing, amending, and "fixing" decisions are outside its
boundaries.

## Modes

Two modes, one procedure:

- **Self-check** — invoked from the project skills (`arch-new`, `arch-change`) at the point of
  choice, before the decision is recorded in files. It runs **only** inside the
  `arch-critic-runner` subagent with a fresh context: the critic receives just the decision package
  (goal + candidates + structural part) and never the reasoning of whoever made the choice.
  Running the self-check inline from a project skill is not allowed — the independence of the
  review rests on context isolation.
- **Standalone review** — on the user's request or on the agent's own initiative during a
  discussion; runs inline, with the report going to chat.

## Process indication

Status-line format: `[arch-critic] <emoji> <action>: <value>`. `❓` appears only at the entry gate
(Step 1), `✅` only at the finish. A gate question is asked as a text status line in chat with the
turn ending there; do not use the `AskUserQuestion` tool.

## Workflow

### Step 1 — Recording the input and the context

Status: `[arch-critic] 📥 Recording the decision and its context: <decision name>`

1. Identify the decision under review: from the chat, a named file, or an ADR.
   - Self-check: the calling skill's package (the candidate table, the delta) is reviewed as one
     whole decision — consistency is checked across the set, and one verdict goes back to the
     caller.
   - Standalone review: the request sets the granularity. A set named as a whole ("review this
     combination", "assess the stack") is reviewed whole; independent decisions go one at a time,
     starting with the one named first. Ambiguous → `[arch-critic] ❓ Review these together or
     separately: <list of decisions>` and end the turn until an answer arrives.
2. Record the stated product goal the decision serves. The goal is not named and cannot be
   reconstructed from the input → `[arch-critic] ❓ Name the product goal this decision serves:
   <decision name>` and end the turn; the review does not start until it is answered. In self-check
   mode the question goes back to the calling skill as the subagent's final text instead of a
   report.
3. Read the context, where the files exist: `docs/architecture/tech-stack.md` (the `stack` yaml
   block), the relevant ADRs from `docs/architecture/adr/`, the project manifests
   (`package.json`, `pyproject.toml`, `Cargo.toml`, `go.mod`). Missing artifacts do not block the
   review — their absence is recorded and goes into the report.

### Step 2 — Review by axes

Status: `[arch-critic] 🔎 Reviewing by axes: <decision name>`

The list of axes is normative; the order is fixed and no axis may be skipped.

| Axis | Questions |
|---|---|
| **(a) Goal and minimal sufficiency** | Which named product goal does the decision serve. Does the existing stack, or a simpler option, serve the same goal. Every new dependency, layer, and abstraction needs a justification. |
| **(b) Verifiedness** | Are versions and compatibility confirmed: is there a `tech-stack.md` entry with the `verified` and `source` fields, and an ecosystem entry with `proof` in the `resolve` block. A `verified` older than 90 days makes the entry count as unverified. Version claims with no confirmation are marked "not verified"; they are never reconstructed from memory or "plausibly" filled in. |
| **(c) Consistency** | Conflicts with the current stack (`tech-stack.md`, the manifests) and with the accepted ADRs. A locally good decision that breaks the whole is grounds for an objection. |
| **(d) Cost of ownership** | What the decision adds to maintenance: new dependencies, infrastructure, expertise, migrations. A qualitative assessment, based on the facts in the input and the context. |

Rules for axis (b):

- The skill checks that confirmations exist and are fresh; it does not verify against primary
  sources itself — that is the project skills' job. A missing confirmation → an item in the report's
  "Not verified" block.
- Any claim about a version or about compatibility that is not confirmed in the supplied material →
  the "Not verified" block; never fill it in from memory.
- In self-check mode the evidence comes from the fields of the package that was passed (version,
  verification date, URL): the absence of `tech-stack.md` entries and of a `resolve` entry for a
  candidate decision is not a failure — the resolve and the write happen after the verdict; a
  candidate with no verification date in the package → "not verified".

### Step 3 — Verdict

Statuses: `[arch-critic] 📋 Forming the verdict: <verdict>`, at the finish
`[arch-critic] ✅ Review complete: <verdict>`

Fill in the report form (the "Report form" section). The addressee depends on the mode: a
standalone review sends the report to chat; a self-check makes the whole report the subagent's
final text, returned to the calling skill (carrying it into chat and into the ADR is the caller's
duty).

Verdict rules:

- **approved** — every axis is `ok`;
- **approved with conditions** — no failures, but there are questions or unconfirmed items; each
  one becomes a condition or a required justification;
- **objection** — at least one axis failed; the argument names either a simpler option (axis a) or
  a specific conflict (axis c). An objection is stated plainly, with no hedging.

## Report form

```markdown
# Decision review: <short decision name>

Verdict: <approved | approved with conditions | objection>

| Axis | Assessment | Argument |
|---|---|---|
| Goal and sufficiency | <ok / question / fail> | <…> |
| Verifiedness | <ok / not verified / stale> | <…> |
| Consistency | <ok / conflict> | <…> |
| Cost of ownership | <ok / question> | <…> |

Conditions / required justifications:
- <…>

Not verified (needs checking against a primary source):
- <…>
```

## Boundaries (what the skill does not do)

- Creates, edits, and deletes no files — the `docs/architecture/` artifacts included.
- States no version, API, or compatibility fact from memory as reliable; anything unconfirmed is
  marked "not verified".
- Does not verify against primary sources and does not resolve versions — it only assesses whether
  confirmations exist in the artifacts and how fresh they are.
- The verdict does not replace the user's confirmation at the calling skills' gates: "approved" is
  an argument for the decision, the decision itself stays with the user.
- Judges no code quality, style, or tests — only the architecture decision.
- Self-check mode runs only inside the `arch-critic-runner` subagent with a fresh context; an
  inline call from a project skill is not allowed — the critic must not see the reasoning of
  whoever made the decision.
