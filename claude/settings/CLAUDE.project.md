# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

This file acts as a navigation system for Claude when working on projects in the repository.

> **How to use this template.** This is a universal starter for a root `CLAUDE.md`.
> Replace every `<...>` placeholder with your project's data, and delete the optional
> blocks marked "delete if …" to match your layout. The methodological core (working
> rules, style, mandatory rules) carries over unchanged.

## Repo Overview

**<PROJECT_NAME>** — <one or two sentences about the product: what it is, who it is for, what problem it solves>.
This repository holds <what exactly: product documentation, planning, analytics — per the repository's makeup>.

<!-- Optional: delete if the project has no nested public code repository.
The product code lives in a separate repository inside the nested folder `<CODE_DIR>/` (its own git,
published as `<public-repo-name>`). The boundary between the private project workshop and the public
product runs along the repository, not `.gitignore`: `<CODE_DIR>/` is excluded from this workspace repo
and keeps its own history. Project documentation does not reach the public repository. -->

Source of the product vision — `<VISION_DOC>`.

## Working Rules

### Parallel agents
- Run independent tasks with parallel agents.
- Parallel agents are allowed only if each works with different files: creating new ones or editing files that no other agent in the wave touches.
- If two agents modify the same existing file, the tasks run sequentially.
- Subagents must not wait for commit confirmation. The orchestrator checks files, runs tests, and makes commits on its own.

### Analysis and evaluation tasks

If asked to perform a critical analysis, do your own critical analysis first. Do NOT copy or reuse old templates/mockups as a starting point unless explicitly asked to.

### Discussion session

If asked to discuss a topic or question:
1) Answer the user's questions briefly, as bullet points.
2) Do NOT dump all the information into the chat at once. Proceed step by step, question by question. Get the user's confirmation before moving on.
3) Make no edits to documents in this mode without the user's explicit confirmation or command.
4) Output all discussion results to the chat first. Create new documents only after receiving the user's command or permission.
5) During the discussion, hold your own critical stance; when needed, argue with the user and defend your point of view (with reasoning). Do not silently agree with everything the user proposes.
6) When idea generation or new solutions are needed: use creative thinking and methods, offering at least three alternative solutions in the chat.

### Bash commands
Strict requirements — each violation triggers an extra permission prompt.
- Do not use `git -C <path> ...` — run `git` directly: git refers to the current working directory.
- Avoid compound commands (`&&`, `;`, pipes) for read-only tasks when separate calls will do.

## Repository Navigation
Claude loads information from the relevant folder depending on the task context:
1. Load that folder's `CLAUDE.md`.
2. Determine the folder's contents and its fit to the task.
3. Find and load the needed file.

Loading the contents of all folders at once is **forbidden**.

**Before starting a task**
- Always show a plan of action and start implementation only after confirmation.

**Session initialization** (replace with your repository's areas)
- Task of type `<area-1>` → `<AREA_1_DIR>/CLAUDE.md`.
- Task of type `<area-2>` → `<AREA_2_DIR>/CLAUDE.md`.
- Task requires product context (vision, audience, economics, boundaries) → `<VISION_DOC>`.

<!-- Optional: delete if the project has no architecture decision log.
- Task touches an architectural/technical decision → `<ADR_DIR>/`: check existing ADRs and do not
  contradict what is accepted (or mark the old one `Superseded`). A significant decision with a long
  consequence is recorded as a separate ADR, but is **created only with the user's explicit
  permission** — Claude does not create ADRs on its own. -->

**Finding a file by description** (without an exact name): run `ls -la` or `find` over the repository folder and locate the matching file yourself.

<!-- Optional: delete if the project has no separate vision document.
**When to read `<VISION_DOC>`**
- The task requires understanding the product, audience, positioning, or strategy.
- The task touches product boundaries, competitors, or key assumptions.
- A decision is needed on tone, emphasis, or the structure of the material. -->

**Response style**
- The writing-style guide (chat and documents) — `<STYLE_GUIDE>` (e.g. `rules/writing-style.md`) — is mandatory; ignoring it is FORBIDDEN; read it when chatting and when creating text documents. Delete this line if there is no style guide.
- No filler phrases, faked emotion, or calls to action.
- Do not restate the task context in the response.
- Do not use rare or obscure words and terms.
- If the task is "check / analyze / evaluate" — do not edit files; return only a report and wait for a command.
- Give honest, candid assessments; do not soften conclusions. Do not try to please.
- Before stating facts about external tools/formats — verify against the project's documents or web sources; do not rely on training data.

**Files and operations**
- Do not create files without an explicit request. By default, output to the chat. Exception: per-folder `CLAUDE.md` files may deliberately permit creating and editing files without separate confirmation within their area — inside such folders their rule applies.
- Destructive operations (delete, force-push, mass-rename, migrations) — only after separate confirmation.
- Do not overwrite final files without explicit confirmation.
- When developing individual skills or agents within the project, always write all frontmatter text in English, except for necessary trigger phrases.

<!-- Optional: delete if the project has no architecture decision log.
- ADRs (`<ADR_DIR>/`) — an exception to the rule above: new ADRs are created only with the user's
  explicit permission. Claude does not create ADRs on its own. -->

**Conventions**
- Files and folders — `kebab-case`, Latin script.

## Repository Layout

```
├── .claude/            # project Claude Code settings (settings.json, agents/, skills/)
├── .gitignore          # git exclusions
├── <area-1>/           # <purpose of area 1>
├── <area-2>/           # <purpose of area 2>
├── <area-3>/           # <purpose of area 3>
└── rules/              # working rules and style guide
```

<!-- Optional: describe special folders — e.g. a nested public code repository,
archive folders (read-only unless the user says otherwise), etc. -->

## Mandatory Rules When Working

### Think before you start implementing a task
Make no assumptions. Do not hide confusion. Do not seek compromises.

Before implementing:
- State your assumptions clearly. If unsure — ask.
- If several interpretations exist, voice them — do not choose silently.
- If there is a simpler approach, say so. Insist on it when necessary.
- If something is unclear, stop. Name what is causing the difficulty. Ask.

### Simple first
- The minimal solution to the specific problem. No speculation.
- No abstractions for one-off use.
- No "flexibility" or "configurability" that was not requested.
- No error handling for impossible scenarios.
- Ask yourself: "Would an expert find this too complex?" If yes — simplify.

### Goal-oriented implementation
Define success criteria. Iterate until the goal is reached.

Turn tasks into measurable goals.
