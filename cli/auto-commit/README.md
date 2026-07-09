# auto-commit

A deterministic quick commit for small edits — no model, no hand-written message.
The counterpart to the `committing-changes` skill for its niche: edits within a single directory, 1–3 files.
A set wider than the threshold is refused with a redirect to the skill.

## Scope of use

| Scenario | Tool |
|---|---|
| Editing 1–3 files in one directory | `auto-commit` |
| Several unrelated changes needing a meaningful message or a split | `committing-changes` skill |

The message is assembled from the git status of the files, so the subject is **mechanical** — it reflects the
operation, not the intent: `<verb> <file names>` (`add parser.py`, `update config.py`, `remove legacy.py`).
The intent type (`feat`/`fix`) is not inferred — a deliberate ceiling of the approach: with zero input it is not derivable.

## Usage

```
auto-commit [--dry-run] [-y|--yes] [-h|--help]
```

| Mode | Behavior |
|---|---|
| bare run | checks → build message → plan to stderr → `y/n` confirmation → commit |
| `--dry-run` | prints the plan and exits (code 0), changes nothing |
| `-y` / `--yes` | skips confirmation (scripted / non-interactive call) |

No TTY and no `-y` — refusal (not a hang).
The result (SHA + message) → `stdout`; the plan, prompt, warnings, and errors → `stderr`.

Exit codes: `0` success/dry-run; `1` expected refusal (threshold, secret, no changes, failed hook); `2` usage error.

## What it does

1. preflight: git repo, not detached HEAD, no merge in progress
2. edit set — `git diff HEAD` over tracked files (leaves untracked alone), **without mutating the index**
3. threshold: >3 files **or** files across more than one parent directory → refusal
4. safety scan: secrets (file names + patterns in added lines) → hard refusal with no override; file >1MB → warning
5. message `scope: <verb> <file names>`:
   - **verb** ← the git operation (`A`→`add`, `M`→`update`, `D`→`remove`, `R`→`rename`); mixed set → `update`
   - **scope** ← last segment of the common directory prefix; no common prefix → the prefix is dropped
   - no body
6. plan → confirmation → `git add -u` → `git commit -m`
7. report: SHA + subject to stdout. **Does not push.**

Staging (`git add -u`) happens only after all checks pass and confirmation.
A refusal (threshold, secret, failed hook), `--dry-run`, and cancel all leave the index untouched.

## Limitations (deliberate)

- the subject comes from file names and git status, not from code symbols;
- intent (`feat`/`fix`/`refactor`) is not inferred — the subject reflects the operation (`add`/`update`/`remove`), not the semantics; the message format is not Conventional Commits;
- no split is performed — only a threshold refusal;
- untracked files are not committed: with a non-empty tracked set → a warning in the plan; if untracked files are the only change (tracked empty), instead of "no changes" it prints a hint to use `git add` + the `committing-changes` skill;
- secret names (`.env`, `*.pem`, `*credentials*`, …) block by name, **except** templates (`.example`, `.sample`, `.template`, `.dist`) and documentation (`.md`, `.txt`, `.rst`); the content of such files is still checked by the content patterns.

## Delivery

Run directly from the repository. To put the command on PATH — via a symlink:

```
ln -s "$PWD/auto-commit" ~/.local/bin/auto-commit
```

## Tests

```
uv run --with pytest pytest test_auto_commit.py -q
```
