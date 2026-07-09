# CLI authoring guide

Methodology for building CLI tools in this repository: scope boundary, language choice, structure,
conventions, testing, templates. Read before any task that involves writing a CLI or utility.

Paired document — `skill-agent-authoring-guide.md` (skills and subagents).

## 1. When to use what: CLI vs skill vs loop

| Artifact | When to use | Who runs it |
|----------|-------------|-------------|
| **CLI** | A deterministic terminal action with no model in the loop at run time: shipping artifacts, calling an API, processing files, machine chores. | A human or another script |
| **Skill** | A model is needed: interpreting context, generating, deciding by situation. | Claude in a session |
| **Loop** | A recurring model-driven cycle on a schedule or condition. | A scheduler / condition |

Boundary rule: a CLI may call `claude` or a skill as a subprocess, but it never stands in for the
model's own logic. If a step requires "understand and decide", it is a skill, not an `if` in a
script.

## 2. Choosing the language for the task

Decision tree:

- **bash** — up to ~50 lines, gluing commands together, no data structures, no JSON parsing. The
  default for wrappers and delivery.
- **Python** — there is logic, branching, work with JSON/HTTP/data, tests are needed. The default for
  API integrations and automation.
- **Node** — off by default. Allowed only when directly reusing code from `plugin-dev`.

The bash → Python boundary: once arrays or dictionaries appear, command-output parsing is needed, or
there are more than three branches — move to Python. Three runtimes for personal tooling is one too
many.

**Target Python version — 3.14** (the current stable branch, released 2025-10-07; the default
`python3`, managed by `uv`, with shims in `~/.local/bin`). 3.13 is an acceptable conservative
minimum. `match`, union annotations `int | None`, and `tomllib` have been available since 3.10–3.11
and are guaranteed present in the target version. Do not touch the system `/usr/bin/python3` (3.9).
Verify with `python3 --version` — the machine's version is authoritative.

## 3. File structure

Model — colocation: a tool lives entirely in its own folder.

- Each tool is a separate folder `cli/<tool-name>/`, even a single-file one.
- Entry point — `cli/<tool-name>/<tool-name>` (executable, kebab-case, no extension) or
  `cli/<tool-name>/main.py` for Python.
- A tool's private logic — inside its folder: `cli/<tool-name>/lib/` or modules next to the entry
  point.
- The root `/lib` — shared code only, with **two or more** consumers.
- Names: folder and entry point — `kebab-case`; Python modules — `snake_case`.

Promotion rule for `/lib`: one consumer — the code stays in the tool's folder; a second appears — it
moves to `/lib`. Not before the second real consumer, so you don't breed one-off abstractions.

Every tool folder is self-contained: understandable and testable apart from the rest.

## 4. Anatomy of a CLI script

Mandatory elements of the entry point:

- **Shebang**: `#!/usr/bin/env bash` or `#!/usr/bin/env python3`.
- **bash** — `set -euo pipefail` as the first line after the shebang.
- **`--help` / `-h`** — mandatory. Describes purpose, usage, flags. Runs with no side effects.
- **Argument parsing**: bash — `case` or `getopts`; Python — `argparse` (stdlib).
- **Header comment** — a one-line statement of purpose at the top of the file.

Keep the entry point thin: argument parsing plus a call into the logic. The logic itself lives in
modules (`lib/`) so it can be tested without running the CLI.

Importing from the root `/lib` in Python — via `sys.path` computed from the script's location, or via
`PYTHONPATH`. Do not rely on the current working directory.

## 5. Behavioral conventions

- **Exit codes**: `0` — success; `1` — expected error (bad input, not found); `2` — usage error
  (bad flags). On failure — a non-zero code, no silent failures.
- **Streams**: result → `stdout`; logs, progress, errors → `stderr`. Output must survive a pipe.
- **Errors**: a message to `stderr` prefixed with the tool name and a non-zero code.
- **Destructive actions** (delete, overwrite, force): behind an explicit `--force` flag or a
  confirmation. By default — dry-run or refusal.
- **Convention flags** with one meaning across all tools: `--help`, `--dry-run`, `--force`,
  `--verbose`.

## 6. Configuration and secrets

- Baseline — read from environment variables. A local `.env` for development, always in
  `.gitignore`.
- A tool's persistent configuration — `~/.config/<tool-name>/config`, outside the repository.
- Secrets in git are forbidden. Only a `.env.example` with empty keys is allowed in the repository.
- Source precedence: flag → env → config file → default.
- A missing required key — a clear error on `stderr` stating which variable to set.

## 7. Dependencies

- **Zero dependencies by default**: bash — no external packages; Python — stdlib only.
- The default `python3` (3.14, managed by `uv`) is a clean interpreter. It carries no third-party
  packages; pandas, lxml, and the like are added at the tool level.
- An external dependency — only on clear need. Isolate it at the tool level, not globally.
- External utilities (`jq`, `gh`, and similar) — check for presence at the start of the script; on
  absence, raise a clear error.
- Do not pull in heavy frameworks for a simple tool.

### Wiring dependencies through uv

The default mechanism is `uv`. Two ways, both isolating dependencies from the global environment:

- **Inline metadata (PEP 723)** — for single-file tools. Dependencies are declared in the script
  header, run via `uv run <script>`; `uv` brings up an ephemeral environment itself. Manage the
  header with `uv init --script`, `uv add --script`, `uv lock --script` (the last pins a lock file
  next to the script). No separate file and no install step — this fits the "no install" rule from
  §8. PEP 723 is accepted; the canonical spec is maintained as the PyPA "Inline script metadata"
  specification.
- **Tool `venv`** — for multi-file tools: `uv venv` in the tool's folder, dependencies in
  `cli/<tool-name>/requirements.txt` (or `pyproject.toml`).

### Dependency level and the system Python

- Dependencies are pinned **at the tool level**, not the repository level. There is no global
  `requirements.txt` at the root.
- Do not install packages globally into the default `python3` — only into the tool's environment
  (inline or `venv`).
- Do not `pip install` into the system `/usr/bin/python3` (3.9): Apple ships a fixed, outdated version
  as part of the Command Line Tools — it is a system dependency, not a project environment. Keep your
  own Python under `uv` and work inside its environment.

### Probing the environment

At the start of a Python task, check against the machine: `python3 --version` (expect 3.14) and, when
in doubt, `python3 -m pip list`. The version and package set on the machine take precedence over what
is written here.

## 8. Delivery

- Run directly from the repository. The entry-point file is executable (`chmod +x`).
- Delivery — via PATH (symlink or adding the folder to PATH) or an alias. No install and no build.
- A tool does not depend on the current working directory: paths are absolute or relative to the
  script's location.
- Self-sufficiency: after `git pull` the tool works with no extra steps beyond the declared secrets or
  `venv`.

## 9. Testing

- **Unit tests are mandatory** for non-trivial logic: Python — `pytest` or `unittest` (stdlib); bash
  — `bats` or simple assert scripts.
- **Smoke test** of the entry point: `--help` works, the basic path passes with code `0`.
- Tests — next to the tool's code (`cli/<tool-name>/lib/<module>_test.py`), not in a shared
  `testing/`. The `testing/` folder is a proving ground for artifacts, not a home for unit tests.
- Separate logic from I/O so it can be tested without running the whole CLI.

## 10. Validation checklist before "done"

- [ ] Shebang in place; for bash — `set -euo pipefail`.
- [ ] `--help` works and describes purpose and flags.
- [ ] Exit codes by convention; errors on `stderr`.
- [ ] Result on `stdout`, pipe-friendly.
- [ ] Secrets not in the repository; a missing key gives a clear error.
- [ ] Dependencies minimal; external utilities are checked for.
- [ ] File is executable; independent of the current directory.
- [ ] Unit tests on the logic and the smoke test pass.
- [ ] Structure by convention: tool in `cli/<tool-name>/`, shared code in `/lib` only with ≥2
      consumers.
- [ ] Comments and prose — per `writing-style.md`; technical identifiers in English.

## 11. Templates (copy-ready)

### 11.1. bash wrapper

```bash
#!/usr/bin/env bash
set -euo pipefail
# <tool-name> — one-line purpose.

usage() {
  cat <<'EOF'
Usage: <tool-name> [options] <arg>

Tool purpose.

Options:
  -h, --help     show help
      --dry-run  show actions without executing
EOF
}

require() { command -v "$1" >/dev/null 2>&1 || { echo "<tool-name>: requires '$1'" >&2; exit 1; }; }

main() {
  local dry_run=0
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help) usage; exit 0 ;;
      --dry-run) dry_run=1; shift ;;
      -*) echo "<tool-name>: unknown flag '$1'" >&2; usage >&2; exit 2 ;;
      *) break ;;
    esac
  done

  require jq
  # logic; result → stdout, errors → stderr
}

main "$@"
```

### 11.2. Python CLI (thin entry point)

```python
#!/usr/bin/env python3
"""<tool-name> — one-line purpose."""
import argparse
import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from lib import core  # the tool's private logic


def main() -> int:
    parser = argparse.ArgumentParser(prog="<tool-name>", description="Purpose.")
    parser.add_argument("arg")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    token = os.environ.get("TOOL_TOKEN")
    if not token:
        print("<tool-name>: set TOOL_TOKEN", file=sys.stderr)
        return 1

    result = core.run(args.arg, token=token, dry_run=args.dry_run)
    print(result)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
```

### 11.3. lib module (logic + test)

```python
# cli/<tool-name>/lib/core.py
def run(arg: str, *, token: str, dry_run: bool = False) -> str:
    """Pure logic, no I/O — tested directly."""
    if dry_run:
        return f"dry-run: {arg}"
    return f"done: {arg}"
```

```python
# cli/<tool-name>/lib/core_test.py
from lib import core


def test_dry_run():
    assert core.run("x", token="t", dry_run=True) == "dry-run: x"


def test_run():
    assert core.run("x", token="t") == "done: x"
```
