# jig

**Scaffold a repository your AI agent already understands.**

One command turns an empty folder into a working local repository: an opinionated
documentation structure (product, planning, architecture + ADRs, design), a ready
Claude Code setup (`CLAUDE.md`, `.claude/` rules), an initialized language environment,
`git init`, and a first commit. jig prepares the ground; your AI agent works on it from
line one. There is no model inside jig — it is a deterministic scaffolder.

## Why

Every new project starts with the same unpaid half-hour: copy the doc layout from the
last repo, hand-write a `CLAUDE.md`, wire up the language toolchain, `git init`, delete
the parts that don't apply. jig collapses that into a single command and gives every
repository the same shape — so the AI agent that works there already knows where things
live.

## Quick start

Build the binary into a directory on your `PATH`:

```
(cd cli/jig && go build -o "$HOME/.local/bin/jig" .)
```

Then scaffold a project:

```
jig my-tool
```

With no flags, jig runs an interview and asks only what it needs. Or drive it entirely
from flags:

```
jig --level full --type cli --lang go my-tool
```

That produces a ready Go CLI: the core doc tree, `go.mod`, a hello-world `main.go`, and a
Go block appended to the root `.gitignore` — `go build -o bin/my-tool .` runs immediately,
and the binary stays out of `git status`.

## What you get

- **A doc structure that's ready to fill, not invent** — `docs/product/` (description,
  architecture, ADR log), `docs/planning/` (backlog), and, in full projects, marketing,
  analytics, and design.
- **Claude Code, pre-wired** — a repository-map `CLAUDE.md` at the root, per-folder
  `CLAUDE.md` guidance, and a writing-style rule under `.claude/`.
- **A real language environment** — `go mod` / `uv` / `npm` run for you, per slot, with the
  right `.gitignore` blocks.
- **A clean git history from the start** — `git init` and one first commit, or none if you
  ask for none.
- **Bilingual interface** — the CLI speaks English and Russian; the scaffolded content is
  English.

## Usage

```
jig [flags] <name>
  --level standard|full           what to set up
  --description <text>            one line about the project
  --type cli|web|library          full only
  --lang <slot>=<lang>            repeatable: --lang backend=go --lang frontend=js
  --commit / --no-commit          whether to make the first commit
  --ui-lang ru|en                 interface language; overrides the environment
  --dry-run                       show the plan, write nothing
  --version                       print the version and exit
  -h, --help
```

Any flag you set removes the matching interview question. For a type with a single slot
(`cli`, `library`), `--lang` takes a bare language with no slot name: `--lang go`. The form
`--lang .=go` is rejected explicitly — exit code 2.

`--dry-run` prints the list of files and the list of commands that would run, and writes
nothing.

## Interview

The default entry point. jig asks for whatever the flags don't already answer: level, name,
description, type, language per slot, and whether to commit. The default name is the current
folder's name, if it passes validation.

If `stdin` is not a terminal and data is missing, jig refuses with exit code 2 and lists the
missing flags. It never launches a prompt into the void.

## Levels

| Level | What it lays down |
|---|---|
| `standard` | the template core minus full-only paths (structure + Claude setup) + `git init` |
| `full` | the whole core + type docs + language environments per slot |

The core tree is one tree; the paths available only at `full` (`docs/product/strategy`,
`docs/product/release-0.1.0`, `docs/planning/release-plan`, `docs/marketing`,
`docs/analytics`) are listed in `templates/full-only.json` and skipped at `standard`.

## Types and languages

The notion of "type" exists only at the `full` level.

| Type | Slot | Languages | Initializer | Language files |
|---|---|---|---|---|
| `cli` | `.` | `go` (default), `python` | `go mod init <name>` / `uv init --app` | `main.go` + Go `.gitignore` block / Python block |
| `library` | `.` | `go` (default), `python` | `go mod init <name>` / `uv init --lib` | `lib.go` + Go `.gitignore` block / Python block |
| `web` | `backend` | `python` (default), `go` | `uv init --app` / `go mod init <name>` | Go / Python `.gitignore` block |
| `web` | `frontend` | `ts` (default), `js` | `npm create vite@latest . -- --template vanilla-ts` / `vanilla` | — (`create-vite` brings its own `.gitignore`) |

Type `web` also lays down the `docs/design/` doc pack (design system: `DESIGN.md`,
`design-guide/`, `template/`). The Vite `vanilla` template means a frontend with no
framework; React, Vue, and Svelte are not opened in v1.

A pack never overwrites what exists: a pack's `.gitignore` is **appended** to the end of the
existing file after a blank line; any other existing file is skipped.

## What gets created

Result of `jig --level full --type web --lang backend=python --lang frontend=ts my-app`:

    my-app/
    ├── .git/              (git init)
    ├── .gitignore
    ├── .claude/           settings.json + rules/writing-style.md
    ├── CLAUDE.md          repository map and rules
    ├── docs/
    │   ├── product/       description, architecture (ARCHITECTURE.md, adr/), strategy, release
    │   ├── planning/      BACKLOG.md, release-plan/
    │   ├── marketing/     (full)
    │   ├── analytics/     (full)
    │   ├── design/        the web type's doc pack: DESIGN.md, design-guide/, template/
    │   └── archive/
    ├── backend/           (uv init)
    └── frontend/          (npm create vite)

At `standard`, the full-only paths drop out of `docs/` and there is no code at the root.
There is deliberately no root `README.md` in the core: a day-one project is local, the map
is carried by `CLAUDE.md`, and the README appears at publication.

Template variables: `.Name`, `.Description`, `.Date`, `.Author`, `.PackageName` (the name
without hyphens — the Go package name), `.Slots` (named sub-slots; web: `backend`,
`frontend`; each with `.Dir` and `.Lang`). At `standard` and for single-`.`-slot types,
`.Slots` is empty.

## Interface language

English and Russian. Resolution: `--ui-lang` if set; otherwise `LC_ALL`; otherwise `LANG`. A
value starting with `ru` means Russian, anything else English; empty or unknown means
English. Questions, answer options, error text, `--help`, and the report are bilingual. Level,
type, and language keys are never translated — they are flag values and manifest keys at the
same time.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | success |
| `1` | expected error: directory not empty, tool missing, initializer failed |
| `2` | usage error: invalid name, unknown flag, unknown type, slot without a language, missing data on a non-interactive run |

## Contract

**Name.** Lowercase latin letters, digits, and hyphen; the first character is a letter; a
hyphen only between two alphanumerics (`my-` and `my--app` are rejected). Checked before
anything else: the name goes into both the directory path and `go mod init`. One rule cuts
off path traversal (`../evil`, `/etc`), spaces, case, and dotted names. A Go-keyword name
(`type`, `func`, …) with a `go` slot is refused with code 2.

**Step order.** render core (level-filtered) → `git init` → type doc pack → slots
(initializer, then packs) → commit. `--dry-run` shows the same paths, doc pack included.

**Target directory.** `jig my-app` creates `./my-app`; the project name equals the folder
name. A non-empty directory is refused. There is no `--force`, no merge.

**Preflight.** Before creating the directory: `git` on `PATH` always; the chosen slots'
initializers for `full`; `user.name` and `user.email` if a commit is requested. Any missing
one is refused before the first file is written.

**Rollback.** None. A failed initializer leaves the directory as is, and jig reports which
slot it stopped on.

**Commit.** The last step, or the initializers' output would not be in it. One commit for the
whole directory. The message is `chore: initial scaffold from jig`, always in English.

## Build and install

jig is compiled. Build the binary into a directory on your `PATH`:

    (cd cli/jig && go build -o "$HOME/.local/bin/jig" .)

After `git pull`, rebuild with the same command. `jig --version` prints the release version
(`0.2.0`), baked into the source and bumped by hand at release.

## Dependencies

Only the Go standard library (1.26). No external packages. The template tree is embedded via
`//go:embed`.

At run time it needs: `git` always; `uv`, `go`, `npm` per the chosen slots. Presence is
checked before any file is written.

---

**Keywords:** repository scaffolding, project bootstrap, repository generator, Claude Code
setup, CLAUDE.md scaffold, AI-ready repository, agent-ready project, Go CLI, git init,
project structure, ADR template, developer tooling, deterministic scaffolder.
**Topics:** `cli` · `scaffolding` · `claude-code` · `developer-tools` · `go`
