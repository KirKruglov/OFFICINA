---
name: adr-write
description: Write an Architecture Decision Record into the ADR journal of the current project. Invoked by arch-new and arch-change at the moment an architecture decision has been confirmed by the user; takes a complete decision package (title, context, decision with source-verified versions, alternatives, consequences, affected records) and produces one numbered ADR file. Records a decision already made; makes no decisions itself. Never invoke outside arch-new or arch-change.
user-invocable: false
allowed-tools: Read, Glob, Write, Edit
---

# ADR Write

Writes a confirmed architecture decision into the `docs/architecture/adr/` journal of the current
project. It works only from inside the `arch-new` and `arch-change` procedures, after their
confirmation gate. It has no gate of its own: the decision was already confirmed by the user at the
calling skill. The skill records the package it was handed as-is — it makes no decisions and does
not amend the data.

## Input — the decision package

Before invoking, the calling skill prints the whole package as a structured block in chat; the
field contents are taken from that block verbatim.

| Field | Required |
|---|---|
| Decision title | required |
| Context: the product goal and the circumstances | required |
| Decision: technologies/approach; every version marked "verified: YYYY-MM-DD" | required |
| Alternatives with reasons for rejection (at least one) | required |
| Consequences | required |
| Affected entries: `stack`/`rules` entries with an action (`add`/`change`/`remove`/`reject`) | required |
| Superseded ADR (number) | optional |

## ADR format

- File: `docs/architecture/adr/NNNN-<slug>.md`.
- `NNNN` — a sequential number starting at `0001`, four digits, with no gaps and no reuse: the next
  number is the highest existing one + 1.
- `slug` — kebab-case, Latin script, derived from the decision title.
- Structure — per [adr-template.md](adr-template.md); every section is required.

## Process indication

Status-line format: `[adr-write] <emoji> <action>: <value> [<context>]`. `🛑` — a stop that returns
control to the calling skill, `✅` — a successful finish.

## Workflow

### Step 1 — Accepting the input and reading the journal

Status: `[adr-write] 📥 Accepting the decision: <title> [<calling skill>]`

1. Check the package for completeness against the "Input" table. Incomplete package → `🛑` return
   to the calling skill with the list of missing fields; no `❓` question goes to the user — the
   calling skill is the one holding the conversation.
2. Read the `docs/architecture/adr/` journal (Glob on `NNNN-*.md`): determine the next number, and
   when a superseded ADR is named, locate its file. The named superseded ADR is not in the journal
   → `🛑` return to the calling skill.
3. No `adr/` directory or an empty journal → the number is `0001`; the directory is created when
   writing in Step 3.

### Step 2 — Assembling the file

Status: `[adr-write] ✍️ Assembling the ADR: <NNNN>-<slug>`

1. Fill `${CLAUDE_SKILL_DIR}/adr-template.md` with the package data; the document date is today,
   the status is `accepted`.
2. Checks before writing:
   - the number is unique in the journal;
   - every required section is filled in;
   - every technology version in the "Decision" section carries the mark "verified: YYYY-MM-DD";
   - the `affects` block is non-empty and matches the template schema (`target: stack | rules`,
     `name`, `action: add | change | remove | reject`).

   Any failed check → `🛑` return to the calling skill with the reason, and the file is not
   written. A version with no verification date is a blocking reason with no exceptions: this is
   the last line of defence against unverified versions.

### Step 3 — Writing and returning

Status: `[adr-write] 💾 Writing the ADR: docs/architecture/adr/<NNNN>-<slug>.md`

1. Write the new file.
2. On a supersession — in the old ADR, update the status line only:
   `Status: superseded (ADR-NNNN)`, where `NNNN` is the number of the new file. This is the only
   edit an existing ADR may receive; the rest of its content is untouchable.
3. Return the path of the created file to the calling skill.
4. The final status comes after the write actually happened:
   `[adr-write] ✅ ADR written: docs/architecture/adr/<NNNN>-<slug>.md`.

## Boundaries (what the skill does not do)

- An invocation outside the `arch-new`/`arch-change` procedure (including a direct user request to
  write an ADR) → `🛑`, nothing is written; name the route — `/arch-new` and `/arch-change`.
- It has no confirmation gate of its own and asks the user nothing: the decision was confirmed at
  the calling skill's gate before the invocation.
- It makes and corrects no decisions: the package is written as-is; a doubt about its contents →
  `🛑` return to the calling skill, with no edit of the data.
- Existing ADRs are never rewritten or deleted; the one permitted edit is the status line on a
  supersession (Step 3.2).
- Versions without the mark "verified: YYYY-MM-DD" are never written, under any circumstances.
- No file outside `docs/architecture/adr/` is created or changed.
