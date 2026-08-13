---
name: handoff
description: Composes a structured handoff of the current coding session — decisions, in-progress work, open questions, touched files, and next steps — for paste into a new session. Manual-invocation only via /handoff — the skill does not self-activate. Trigger phrases — "handoff this session", "session brief", "continue in a new session", "передача сессии", "сделай handoff".
disable-model-invocation: true
argument-hint: "[topic]"
allowed-tools: Read, Grep, Glob, TaskList, TaskGet
disallowed-tools: Edit, Write, NotebookEdit
---

# Handoff

Composes a handoff block of the current session: decisions, in-progress work, open questions,
touched files, and next steps. The output is markdown in chat, ready to paste as the first
message of a new session. Invocation is manual only, via `/handoff`.

Terms: **handoff block** — the output markdown; **source** — a conversation turn, a `TaskList`
entry, or a line from the git snapshot; **topic** — `$ARGUMENTS` or the main line of work.

## When to apply

Only on the direct command `/handoff`. Self-activation is forbidden, including when the
session is full and after `/compact`.

## References

- Block layout: `${CLAUDE_SKILL_DIR}/templates/handoff-template.md` (or
  `templates/handoff-template.md` next to this file).
- Phrasing bar: `${CLAUDE_SKILL_DIR}/references/examples.md` (or
  `references/examples.md` next to this file).

## Process indication

Print a status line at the start of every step:

`[handoff] <emoji> <present-tense action>: <value> [<context>]`

Reserved emoji: `🛑` stop / blocking error · `❓` clarification from the user · `⏸️` waiting
for confirmation (gate) · `✅` successful finish. Working actions take a themed emoji, one
per step.

A gate (`❓`) is executed as text: print the status line and one question, then end the
turn. Do not use `AskUserQuestion`.

## Load-time snapshot

Substituted on invocation. Empty output, an error, the literal command text still visible,
or `not a git repository` — there is no git source; do not re-run the commands.
`${CLAUDE_PROJECT_DIR}` unsubstituted → the project root from context.
`${CLAUDE_SESSION_ID}` unsubstituted → `—`. Date line empty or still a command → the
conversation's current date. `git log` is reference only: a commit enters **Done** only
when the conversation or a closed task confirms it.

Directory: `${CLAUDE_PROJECT_DIR}`
Session: `${CLAUDE_SESSION_ID}`
Date: !`date +%Y-%m-%d`

Branch:

!`git branch --show-current`

Status:

!`git status --short`

Diff:

!`git diff --stat`

Recent commits:

!`git log -5 --oneline`

## Workflow

### Step 1 — Entry check and snapshot

Status: `[handoff] 📥 Snapshotting session: <topic|empty>`

1. **Read `TaskList`.** Do this first. The tool is missing or the list is empty — there
   are no tasks; do not stop the step.
2. **Read the snapshot** above and `$ARGUMENTS`.
3. **Stop the run** if the session has neither a goal nor work: no substantive turns,
   decisions, tasks, or edits. Status `🛑`, text: `nothing to hand off — the session has
   only just started`.
4. **Fix minimal mode** if turns are few, a goal is named, and there are not yet
   decisions, tasks, or edits: only the topic, the status, one next step, and the note
   `short session — only the intent was recorded`. An edit or a decision is present —
   full block.
5. **Narrow the scope** if `$ARGUMENTS` is set: keep only that topic, include nothing
   else.

→ check: `🛑` was printed, minimal mode was fixed, or a source for the full block exists.

### Step 2 — Extract the items

Status: `[handoff] 🔎 Collecting handoff items: <topic>`

1. **Lean on the git snapshot and the tasks** if history was compacted with `/compact`.
   A gap they do not close → a question in item 5.
2. **Collect the six fields** from the table below. An item with no source is not
   written: drop it or turn it into a question in item 5.
3. **Split parallel topics** if no filter is set and the session has more than one
   line: subsections inside **In progress** and **Next steps**; mark the main one.
   Main — the last active line, or the line with dirty files.
4. **Set the block status:** a blocker with no workaround → `blocked`; the current
   turn says the work is stopping → `paused`; otherwise → `in progress`.
5. **Ask one `❓` question** if a section is ambiguous (a task is named, the outcome
   is not). Do not ask more than one question before emitting the block. Unambiguous —
   skip this item.
6. **Repeat items 2–5** until the field list is closed.

Field table — where to take each item and what to write:

| Field | Where | Rule |
|---|---|---|
| Done | conversation, closed tasks, commits confirmed in the session | confirmed only |
| In progress | conversation, open tasks, uncommitted files | each item has its current state |
| Open questions | conversation | unresolved, blocker, or fork |
| Files | `git status` / `git diff`; plus a file from the conversation if `Glob` finds it | `path — what was done`; no files — omit the section |
| Next steps | unfinished work and the explicit next move | numbered list, nearest first |
| Context | what the lists above do not say | 1–2 sentences |

→ check: every non-empty item has a source; empty **Done**, **In progress**, and
**Open questions** equal `— none —`.
→ failed: return to item 2 and fill or drop the item; after 2 repeats — `🛑` naming
the source that is missing.

### Step 3 — Assemble the block

Status: `[handoff] ✍️ Assembling block: <topic>`

1. **Fill** the template from References. Keep the template headers. Write field
   values and the opener line in the user's language. Check **Done**, **In progress**,
   and **Context** against the phrasing bar in References.
2. **Substitute** the date, directory, branch, and session id from the snapshot. No
   git source — `Branch: — no git —`.
3. **Drop the Files section** if the list is empty. Leave no `<…>` placeholders.
4. **Print** the whole block in one code fence; directly under it — the opener line:

   > **New session:** *"Continue: <topic>. Context below."* — then paste the block.

5. **Revise the block and print it again** if the user fills a gap in the next turn.
   Do not ask new questions.

Status at finish: `[handoff] ✅ Block ready: <topic>`

→ check: the output has a code fence, the opener line, and every required template
section; the Files section is either filled or absent.
→ failed: return to item 1 and finish the assembly; after 2 repeats — `🛑` naming
the section that is missing.

## Boundaries (what the skill does not do)

- Does not load or run without the `/handoff` command, including when the session is
  full and after `/compact`.
- Does not create or edit files.
- Does not read `~/.claude/projects/**/*.jsonl`.
- Does not invent decisions, file statuses, or check results.
- Does not copy secrets, keys, or `.env` contents.
- Does not commit, push, or continue the session's task.
- Does not run Bash. Git state comes only from the load-time snapshot.
- Does not treat `claude --resume` as the goal of the run: the output is a briefing
  for a new session.
