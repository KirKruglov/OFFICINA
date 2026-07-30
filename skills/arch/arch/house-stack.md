# house-stack

A cross-project registry of allowed technologies. It bounds the choice space of the project skills
(`arch-new`, `arch-change`): by default they may only propose what is recorded here.

> **This is one author's stack, published as a working example — replace it before the first run.**
> The technologies below are personal choices, not a recommendation. What matters for the skills is
> the file's shape: one row per layer, a primary choice, an optional alternative, and the condition
> under which the alternative wins. Keep the shape, swap the content.
>
> Rewriting is not optional. `arch-new` and `arch-change` may only propose what is recorded here;
> everything else passes a `⏸️` gate per candidate. On a stack these rows do not cover, a run becomes
> a gate per layer rather than a design session.

Last updated: 2026-07-19. As of that date every entry was checked against its repository and
official documentation: open source with no restrictive licence, actively maintained.

## Rules of use

- A strong default, not an absolute: the skill proposes the primary choice for a layer.
- A layer's alternative is allowed only when its condition holds.
- A technology outside this registry passes a gate: an explicitly named product need plus user
  confirmation.
- Versions are not recorded here: a version is verified against a primary source on the date of the
  decision and lives in the project's `tech-stack.md`.

## Layers

| # | Layer | Primary | Alternative | Condition for the alternative |
|---|---|---|---|---|
| 1 | CLI and utilities | Go | Python, bash | selection criteria — table below |
| 2 | Data processing | Python + pandas | — | — |
| 3 | Web frontend | React + TypeScript + Vite; package manager pnpm | — | — |
| 4 | Web backend (API) | Python + FastAPI (uv environment) | — | — |
| 5 | Database / storage | SQLite | PostgreSQL | a multi-user service with concurrent writes, or a server deployment where the database lives separately from the application |
| 6 | Desktop | Wails v2 (Go logic, UI — React + TypeScript + Vite) | Electron | a mature ecosystem of native modules is needed, or identical rendering (Chromium) across all operating systems |
| 7 | Mobile | React Native + Expo (TypeScript) | — | — |

## Layer 1: how the tool is chosen

| Tool | When | CLI interface |
|---|---|---|
| Go — primary | a permanent tool: long-lived, run regularly, installed into `~/.local/bin`, distributed as a binary | stdlib `flag`; cobra — when there are subcommands, completion, or generated help |
| Python (uv, PEP 723 inline dependencies) | a library-bound task (format conversion, files, HTTP, parsing) or a one-off / exploratory script | typer — for a full CLI; otherwise just a `uv run` script |
| bash | glue for a few shell commands, no branching logic and no dependencies | — |

The Go/Python line is "tool or task": permanent and reusable — Go; one-off or living off
libraries — Python.

## Language perimeter

Go, Python, TypeScript; bash as a utility language. A technology that requires a new language (Dart,
Kotlin, Rust, and others) goes through the "outside the registry" gate.
