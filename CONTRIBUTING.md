# Contributing to OFFICINA

Thanks for your interest! A quick note on how this repository works, so contributions land smoothly.

## How this repo works

OFFICINA is a **curated showcase** of a personal development setup — skills, subagents, CLI tools,
methodology, and dotfiles. It is a **derivative, not the source of truth**: artifacts are maintained
in a private source and exported here after sanitization. The flow is **one-directional** (private
source → sanitize → showcase), so this repo never becomes a second source of truth.

That shapes what contributions look like — see below.

## What's welcome

- **Issues** — bug reports (a broken script, a wrong path, a dead link), questions, ideas, and
  suggestions. This is the best channel for anything substantive.
- **Small pull requests** — typo fixes, broken links, documentation clarity, and script bugs
  (macOS). These are reviewed and merged directly.

## What to expect for larger changes

Substantive changes to an artifact (rewriting a skill, reworking a config, adding a feature) are
applied in the **private source of truth** and then re-synced here. So a large artifact PR may be
closed with the change applied upstream instead of merged as-is — the outcome ships, just through a
different path. If you have such a change in mind, **open an issue first** to discuss it.

## Reporting an issue

Please include:

- **What you ran** and what you expected vs. what happened.
- **Environment** — macOS version, and the relevant tool version (e.g. VS Code, `zsh`).
- Steps to reproduce, and any error output.

> [!NOTE]
> This setup targets **macOS only**. Reports for other platforms are appreciated but may be
> out of scope.

## Pull request guidelines

- Keep PRs **small and atomic** — one concern per PR.
- Write in **English**, and match the surrounding style (comment density, naming, formatting).
- **No machine-specific paths** (`/Users/<name>/...`) and **no secrets or tokens** — the same
  sanitization rule applies to every file here.
- For shell scripts, keep them POSIX/`bash` clean; existing scripts use `set -euo pipefail`.
- Commit messages in English, imperative mood (e.g. `fix: correct extension count in README`).

## License

By contributing, you agree that your contributions are licensed under the
[MIT License](LICENSE), the same license that covers this repository.

---

New here? Start with the [README](README.md) for the project map and philosophy.
