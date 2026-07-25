# CLI tools — small, deterministic command-line helpers with no model in the loop

Personal **command-line tools** for the work an AI agent shouldn't be doing: a quick git commit built
from `git status` rather than from a model, and a repository scaffolder that lays down a doc tree,
a Claude Code setup, and an initialized language environment in one command. Each tool lives in its
own folder with its entry point, its tests, and its own README. Part of the [OFFICINA](../)
repository.

Nothing here calls a model at run time. That's the point — see below.

## Why a CLI and not an agent

The first question in the OFFICINA method is what to solve a task *with*, and the cheapest answer is
usually a script. A model in the loop buys you judgment; you pay for it in tokens, latency, and
variance. When there's no judgment to be had — when the same inputs must always produce the same
output — that trade is pure loss.

The boundary is sharp: **if a step needs "understand and decide", it's a skill; if it can be an `if`,
it's a script.** A CLI may call an agent as a subprocess, but it never stands in for the model's own
reasoning, and the reverse holds too — asking a model to run a fixed procedure produces confident
variation where you wanted a guarantee.

Both tools here sit on that boundary deliberately. `auto-commit` refuses the cases that need judgment
and hands them to a skill instead of guessing. `jig` writes files from templates and runs
initializers; it has no model inside it at all. The full reasoning, including the language decision
tree and the conventions every tool follows, is in
[`methodology/cli-authoring-guide.md`](../methodology/cli-authoring-guide.md).

## The tools

| Tool | Command | Language | What it does |
|------|---------|----------|--------------|
| [`auto-commit`](auto-commit/) | `auto-commit [--dry-run] [-y]` | Python | Deterministic quick commit for small edits — message built from git status, secret scan, refuses anything wider |
| [`jig`](jig/) | `jig [flags] <name>` | Go | Scaffolds a ready-to-work repository: doc tree, Claude Code setup, language environment, `git init`, first commit |

### `auto-commit` — a commit message with no model in it

Most commits in a working day are one file, one directory, one obvious change. Routing those through
a model means waiting on a round trip to be told what `git status` already knew. `auto-commit`
handles exactly that niche and nothing else.

The message is **mechanical** by construction: `scope: <verb> <file names>` — `parser: add parser.py`,
`config: update config.py`. The verb comes from the git operation (`A`→`add`, `M`→`update`,
`D`→`remove`, `R`→`rename`; a mixed set becomes `update`), the scope from the last segment of the
common directory prefix. Intent — `feat` versus `fix` — is deliberately **not** inferred: with zero
input beyond file names, it isn't derivable, and a guessed type is worse than no type.

The thresholds are the interesting part. More than 3 files, or files spanning more than one parent
directory, and the tool **refuses** and points you at the
[`committing-changes`](../skills/committing-changes/SKILL.md) skill — the case now needs judgment, so
it goes to the artifact that has some. A secret in the added lines is a hard refusal with no override.
Files over 1 MB warn. Staging (`git add -u`) happens only after every check passes and you confirm; a
refusal, a `--dry-run`, or a cancel all leave the index exactly as it was. It never pushes.

Streams follow the convention: the result — SHA and subject — goes to `stdout`, everything else (plan,
prompt, warnings, errors) to `stderr`, so the output survives a pipe. Exit codes are `0` success or
dry-run, `1` expected refusal, `2` usage error. With no TTY and no `-y`, it refuses rather than
hanging. Full behavior in [its README](auto-commit/README.md).

### `jig` — scaffold a repository your AI agent already understands

Every new project starts with the same unpaid half-hour: copy the doc layout from the last repo,
hand-write a `CLAUDE.md`, wire up the language toolchain, `git init`, delete the parts that don't
apply. `jig` collapses that into one command, and gives every repository the same shape — so the agent
working there already knows where things live.

```bash
jig --level full --type cli --lang go my-tool
```

That produces a working Go CLI: the core doc tree (`docs/product/` with architecture and an ADR log,
`docs/planning/` with a backlog, and at `full` also strategy, marketing, analytics, and design), a
repository-map `CLAUDE.md` at the root plus per-folder guidance and a writing-style rule under
`.claude/`, a `go.mod` and hello-world `main.go` with the right `.gitignore` block appended, `git init`,
and one first commit. Run bare, `jig my-tool` asks only for what the flags don't already answer.

It's a **deterministic scaffolder**: a Go binary, standard library only, templates embedded via
`//go:embed`. `--dry-run` prints every path it would write and every command it would run, and writes
nothing. A pack never overwrites what exists — `.gitignore` is appended after a blank line, any other
existing file is skipped. Preflight checks run before the first file is written: `git` on `PATH`
always, the chosen slots' initializers (`go`, `uv`, `npm`) at `full`, and `user.name` / `user.email`
if a commit is requested. There's no rollback and no `--force`, because there's nothing to undo — it
refuses a non-empty directory outright.

The interface is bilingual (English and Russian, resolved from `--ui-lang`, `LC_ALL`, or `LANG`),
while everything it scaffolds is English. Flags, levels, types, exit codes, and the full contract are
in [its README](jig/README.md).

### How they hand off

Neither tool tries to be complete on its own. `auto-commit` covers the mechanical bottom end of
committing and escalates everything else to the [`committing-changes`](../skills/) skill — one
threshold, two artifacts, no overlap. `jig` scaffolds the exact doc structure and Claude setup that
the guides in [`methodology/`](../methodology/) describe: it's the method turned into a binary, so the
conventions arrive with the repository instead of being reconstructed per project.

## Install

**Script tools** run directly from their folder — no build step. To put the command on your `PATH`,
symlink the entry point into a directory that's already there:

```bash
ln -s "$PWD/auto-commit/auto-commit" ~/.local/bin/auto-commit
```

**Compiled tools** build a binary instead. `jig` is a Go program — build it straight onto your `PATH`:

```bash
(cd jig && go build -o "$HOME/.local/bin/jig" .)
```

After a `git pull`, rebuild with the same command; script tools pick up changes with no action. If
`~/.local/bin` isn't on your `PATH` yet, add it in your shell profile:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## How a tool here is built

Every tool follows the same conventions, which is what makes them readable in isolation and safe to
copy out one at a time:

| Convention | Rule |
|---|---|
| **Layout** | One folder per tool — entry point, private logic, tests, README. Self-contained |
| **Entry point** | `<tool-name>/<tool-name>`, executable, kebab-case, no extension (`main.go` for Go) |
| **Dependencies** | Zero by default — bash with no external packages, Python stdlib only, Go standard library only |
| **Exit codes** | `0` success · `1` expected error · `2` usage error. Never a silent failure |
| **Streams** | Result → `stdout`; logs, progress, prompts, errors → `stderr`. Output survives a pipe |
| **`--help`** | Mandatory, side-effect free, describes purpose and every flag |
| **Destructive actions** | Behind an explicit flag or a confirmation. Default is dry-run or refusal |
| **Shared code** | Moves to a shared library on the *second* real consumer, never the first |

Language choice is a decision tree rather than a preference: **bash** up to about 50 lines of gluing
commands, **Python** once there's branching, JSON, or a need for tests, **Go** when the tool must ship
as a single static binary or embed its own assets. The full guide, including configuration, secrets
handling, and dependency wiring through `uv`, is
[`methodology/cli-authoring-guide.md`](../methodology/cli-authoring-guide.md).

## Tests

Each tool carries its own tests and runs them from its own folder. Python:

```bash
uv run --with pytest pytest -q
```

Go:

```bash
go test ./...
```

## FAQ

### Why generate a commit message with a script instead of an AI?

For small, obvious changes there's nothing to interpret — the operation and the file names are already
in `git status`, and a round trip to a model adds latency and variance for no new information. The
trade-off is that the subject describes the *operation*, not the *intent*: `update config.py`, never
`fix: handle empty config`. When intent matters, the case is above `auto-commit`'s threshold anyway
and belongs to the [`committing-changes`](../skills/committing-changes/SKILL.md) skill.

### When should I write a CLI instead of a skill or a subagent?

Write a CLI when the same inputs must always produce the same output and no step requires judgment.
Write a skill when the procedure is fixed but the decisions inside it depend on context. Write a
subagent when you need a whole role with its own context window. The dividing question is whether any
step needs "understand and decide" — if yes, a script will fake it badly.

### Can I take one tool without cloning the whole repository?

Yes. Each folder is self-contained, with zero external dependencies by design. Copy
`cli/auto-commit/` anywhere and symlink the entry point; copy `cli/jig/` and run `go build`. Nothing
reaches back into the rest of the repo at run time.

### Does `auto-commit` ever commit something I didn't expect?

It stages with `git add -u`, which touches only tracked files that already changed — untracked files
are never committed, only flagged in the plan. Staging happens after the checks pass and after you
confirm, and every refusal path leaves the index untouched. `--dry-run` shows the exact plan without
changing anything.

### Is `jig` safe to run in an existing project?

No, and it declines to try: a non-empty target directory is refused outright, and there is no
`--force`. It's a scaffolder for new repositories. To see what it would produce anywhere, run it with
`--dry-run`.

### What does `jig` need installed?

`git` on your `PATH` always, plus the initializers for whatever you scaffold — `go`, `uv`, or `npm`
per language slot at the `full` level. All of them are checked in preflight, before the first file is
written, so a missing tool is an error and never a half-built directory.

## Related

- [OFFICINA](../) — the method and the rest of the toolkit
- [`methodology/cli-authoring-guide.md`](../methodology/cli-authoring-guide.md) — the conventions above, in full
- [`skills/`](../skills/) — the artifacts these tools hand off to when judgment is required
- [`claude/`](../claude/) — Claude Code subagents and settings
