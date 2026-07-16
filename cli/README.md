# CLI tools — small, deterministic command-line helpers

Personal command-line tools: single-purpose, deterministic, no model in the loop. Each lives in its
own folder with its entry point, tests, and a README. Part of the [OFFICINA](../) repository.

## Tools

| Tool | Command | What it does |
|------|---------|--------------|
| [`auto-commit`](auto-commit/) | `auto-commit [--dry-run] [-y]` | Deterministic quick commit for small edits (1–3 files, one directory). Builds a mechanical `scope: verb files` message from git status, scans added lines for secrets, and refuses anything wider — redirecting to the [`committing-changes`](../skills/committing-changes/SKILL.md) skill. Never stages until checks pass, never pushes. See [its README](auto-commit/README.md). |
| [`jig`](jig/) | `jig [flags] <name>` | Scaffolds a ready-to-work local repository in one command: an opinionated doc structure, a Claude Code setup (`CLAUDE.md`, `.claude/` rules), an initialized language environment (`go`/`uv`/`npm`), `git init`, and a first commit. A deterministic Go scaffolder — no model in the loop. See [its README](jig/README.md). |

## Delivery

Script tools run directly from their folder — no build step. To put a command on your PATH, symlink
its entry script into a directory that's already there:

```
ln -s "$PWD/auto-commit/auto-commit" ~/.local/bin/auto-commit
```

Compiled tools build a binary instead. `jig` is a Go program — build it straight onto your PATH:

```
(cd jig && go build -o "$HOME/.local/bin/jig" .)
```

Python tools carry their own tests; run them from the tool's folder:

```
uv run --with pytest pytest -q
```

Go tools test the same way, from their folder:

```
go test ./...
```

---

**Keywords:** CLI tools, command-line helpers, git automation, auto-commit, project scaffolding, repository generator, Claude Code setup, deterministic tooling, developer tools, Python CLI, Go CLI, shell.
**Topics:** `cli` · `git` · `developer-tools` · `automation`
