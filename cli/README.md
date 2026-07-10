# CLI tools — small, deterministic command-line helpers

Personal command-line tools: single-purpose, deterministic, no model in the loop. Each lives in its
own folder with its entry script, tests, and a README. Part of the [OFFICINA](../) repository.

## Tools

| Tool | Command | What it does |
|------|---------|--------------|
| [`auto-commit`](auto-commit/) | `auto-commit [--dry-run] [-y]` | Deterministic quick commit for small edits (1–3 files, one directory). Builds a mechanical `scope: verb files` message from git status, scans added lines for secrets, and refuses anything wider — redirecting to the [`committing-changes`](../skills/committing-changes/SKILL.md) skill. Never stages until checks pass, never pushes. See [its README](auto-commit/README.md). |

## Delivery

Each tool runs directly from its folder — no build step. To put a command on your PATH, symlink its
entry script into a directory that's already there:

```
ln -s "$PWD/auto-commit/auto-commit" ~/.local/bin/auto-commit
```

Python tools carry their own tests; run them from the tool's folder:

```
uv run --with pytest pytest -q
```

---

**Keywords:** CLI tools, command-line helpers, git automation, auto-commit, deterministic tooling, developer tools, Python CLI, shell.
**Topics:** `cli` · `git` · `developer-tools` · `automation`
