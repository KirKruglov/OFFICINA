---
name: release-finalize
description: Finalize a product release and prepare the repo for a public push or deploy — updates the changelog and root README, scans for secrets and version consistency, creates an annotated git tag. Does NOT merge, push, or deploy. Universal — auto-detects release artifacts, optional .release-finalize.yml overrides. Manual-invocation only via /release-finalize — the skill does not self-activate.
disable-model-invocation: true
argument-hint: "[version]"
allowed-tools: Read, Grep, Edit, Bash(git:*), TodoWrite
---

# Release Finalize

Gathers the release artifacts of any product and prepares the repository for a push to a public
repository or a deploy. Runs autonomously: a single checkpoint before applying edits, then a
straight run to the tag with no intermediate confirmations.

Invocation is manual only, via `/release-finalize` — the skill never self-activates.

## Scope of responsibility

- ✅ Changelog — prepend a new section (Keep a Changelog / SemVer).
- ✅ Root `README.md` — sync to the release context.
- ✅ Pre-publish scan — secrets, version consistency, clean tree.
- ✅ Annotated git tag on HEAD.
- ❌ merge — forbidden.
- ❌ `git push`, deploy — only suggest the command, never run it.

Artifacts are auto-detected; overrides come from a `.release-finalize.yml` file at the root of the
target repo (format — this skill's `templates/release-config.yml`, e.g.
`${CLAUDE_SKILL_DIR}/templates/release-config.yml` in Claude Code).

## Process indication

At the start of each phase, print a status line:

`[release-finalize] <emoji> <action in present tense>: <progress> [<context>]`

Reserved emoji: `🛑` stop/block · `❓` clarification · `⏸️` awaiting approval · `✅` finish.
Working actions use a per-phase topical emoji (`🧭 🔎 ✍️ 📝 🏷️` and the like).

## Checklist (run plan)

Before Phase 0, print the plan and create one TodoWrite task per phase; track status via TodoWrite.
The list is fixed and executed in order:

1. Detect — version, branch, artifacts, commits
2. Scan — secrets, version consistency, cleanliness
3. Draft — changelog draft + README edits
4. Apply — apply changelog and README (after approval)
5. Commit — delegate to `committing-changes`
6. Tag — annotated tag
7. Hand-off — summary and commands

## Workflow

Phases 0–2 run autonomously. Between Phase 2 and Phase 3 there is a single `⏸️` checkpoint.
Phases 3–6 run autonomously, with no further confirmations.

### Phase 0 — Detect

Status: `[release-finalize] 🧭 Gathering release context: <version|?>`

1. Release version `X.Y.Z` — **source of truth: the command argument (`$1`) or the user's message**.
   If not given — try to extract it from version files; still missing → `❓` "Specify the release
   version (e.g. 0.2.0)". Version files serve only for the consistency check (Phase 1); they do not
   determine the release version.
2. Read `.release-finalize.yml` at the repo root, if present, and take its overrides (changelog,
   readme, version_files, branch_pattern, public_prep). No file — work off auto-detection.
3. Determine the artifacts (honoring overrides):
   - changelog: `CHANGELOG.md`, otherwise `HISTORY.md`; if neither exists → mark a new
     `CHANGELOG.md` (heading `# Changelog`) for creation in Phase 3;
   - readme: the root `README.md`;
   - version files: `pyproject.toml`, `package.json`, `VERSION`, `Cargo.toml` (whichever are present).
4. Branch: `git rev-parse --abbrev-ref HEAD`. If `branch_pattern` is set and the branch does not
   match — warn, but do not stop.
5. Latest tag: `git describe --tags --abbrev=0`. If a tag exists →
   `git log <tag>..HEAD --oneline`; no tags → `git log --oneline`. Keep the list as reference for
   Phase 2.

### Phase 1 — Scan

Status: `[release-finalize] 🔎 Checking readiness to publish`

1. Clean tree: `git status --porcelain`. Non-empty → `🛑`, list the files, offer stash /
   `committing-changes` / cancel.
2. Secrets: `git grep -nIE` on assignment patterns, not bare words —
   `(AKIA[0-9A-Z]{16}|BEGIN [A-Z ]*PRIVATE KEY|(api[_-]?key|secret|token|password)[[:space:]]*[=:][[:space:]]*['"])`
   and on the paths from `public_prep.exclude` (e.g. `.env`). On a hit → `🛑`, show file:line (do
   not print the full value), stop.
3. Version consistency: read the version in each version file, compare with `X.Y.Z`. On a
   mismatch → `🛑`, show a file→version table.
4. `.gitignore` / stray files: from `git status` and any un-ignored private paths — `⚠️`, list them
   for review (not a stop).

### Phase 2 — Draft

Status: `[release-finalize] ✍️ Preparing changelog and README edits`

1. Group the commits from Phase 0 by conventional-commit type (`feat` → Added, `fix` → Fixed,
   `docs`/`refactor`/`chore` → Changed, `BREAKING` → its own block). Read the template
   `templates/changelog-entry.md` (e.g. `${CLAUDE_SKILL_DIR}/templates/changelog-entry.md`).
2. Assemble the section: `{VERSION}`→X.Y.Z, `{DATE}`→today (the system `Today's date`),
   `{NARRATIVE}`→2–3 points on the essence of the release, `{CHANGES_LIST}`→a minimal list of
   `- module: what changed`.
3. README: prepare a generic sync to the release context (version / component status, if such
   sections exist; no hardcoding of specific sections). If there are no obvious edits — leave the
   README unchanged.

### Checkpoint `⏸️`

Status: `[release-finalize] ⏸️ Awaiting approval: vX.Y.Z`

Show in one block: the version, branch, detected artifacts, the Phase 1 scan report, the prepared
changelog section, and the proposed README edits (before/after). Ask (via `AskUserQuestion`):
"Apply? (ok / edits)". Do not touch files until `ok`.

### Phase 3 — Apply

Status: `[release-finalize] 📝 Applying edits`

Changelog exists → prepend the section (insert before the first `##`). Does not exist → create
`CHANGELOG.md` with heading `# Changelog` and the first release section. Apply the README edits, if
any.

### Phase 4 — Commit

Status: `[release-finalize] 📦 Committing the finalization`

Delegate to the `committing-changes` skill with the context "release vX.Y.Z: changelog + README".
Do not proceed to Phase 5 until it completes.

### Phase 5 — Tag

Status: `[release-finalize] 🏷️ Creating the tag`

```
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git rev-parse vX.Y.Z
```

### Phase 6 — Hand-off

Status: `[release-finalize] ✅ Release vX.Y.Z finalized`

Print the summary and next steps (do not run the commands):

```
Release vX.Y.Z finalized:
  ✅ changelog — section vX.Y.Z added
  ✅ README — synced
  ✅ git tag vX.Y.Z on <SHA>

Next:
  • merging-branches — merge the release branch → main
  • git push origin vX.Y.Z — publish the tag
  • deploy — per the product's procedure
```

## Boundaries (what the skill does not do)

- Does not run `git merge`, `git push`, `git branch -D`, or a deploy — only suggests the commands.
- Does not touch files outside the declared artifact set (changelog, root README).
- On a secret hit or a version mismatch — `🛑`, does not silently continue.
- Does not apply edits before approval at the `⏸️` checkpoint.
- Version bump in version files (`pyproject.toml` / `package.json` / …) is out of scope and must be
  done before the skill runs; Phase 1 only checks consistency.
- Bash: do not use `git -C <path>`; the working directory is already set; avoid compound `&&`
  commands for read-only calls.
