---
name: merging-branches
description: Use BEFORE any git merge, rebase, or branch integration. Reads project CLAUDE.md to determine merge strategy, runs pre-merge safety checks (diff name-status both sides, clean working tree, up-to-date with remote, conflict scan), requires explicit user confirmation for merge to main/master/release branches. Trigger phrases — "merge X into Y", "rebase onto", or any request implying branch integration.
---

# Merging Branches

When invoked, follow this procedure step by step. Create a TodoWrite task per step.

## Rule Hierarchy

1. Explicit user instruction in the current message — wins.
2. Project `CLAUDE.md` — sections `## Git workflow`, `## Branching`, `## Release process`.
3. Global `~/.claude/CLAUDE.md`.
4. This SKILL.md.

## Step 1: Read conventions and determine strategy

- Read project `CLAUDE.md`. Look for sections: `## Git workflow`, `## Branching`, `## Release process`.
- Read `~/.claude/CLAUDE.md` as fallback.
- If you find merge-strategy guidance, surface the relevant quote to the user and ask to confirm interpretation:

```
From project CLAUDE.md:
> On release push — squash-merge the branch's commits.

Interpreted strategy: squash. Confirm? (y/n/other)
```

If strategy is NOT described in any CLAUDE.md — ask user directly: squash / merge commit / rebase / ff-only.

## Step 2: Identify source and target

- Both must come from user request. If either is ambiguous — ask. Do NOT guess based on current branch.

## Step 3: Classify target

**Protected branches** (require extra confirmation):
- `main`, `master`
- `release/*`
- `production`, `prod`, `staging`

For protected target, add an explicit confirmation gate:

```
⚠ Target = main (protected branch).
This will change commit history on `main` permanently. Confirm explicitly by typing the exact phrase: "yes merge to main"
```

Only proceed if user types the exact confirmation phrase.

Note: the phrase must include the full branch name including any `/` characters. For `release/0.1.0` the confirmation is `"yes merge to release/0.1.0"`.

## Step 4: Pre-merge safety checks

Run all four in order. Any failure → STOP and report.

### 4a. Clean working tree

```bash
git status --porcelain
```

If non-empty → STOP. List dirty files. Ask user: stash / commit (invoke committing-changes) / cancel.

After `committing-changes` completes and reports success, resume this skill from Step 4b. The merge has NOT been performed yet — only the working tree was cleaned.

### 4b. Up-to-date with remote

Detect target's configured upstream (avoid hardcoding `origin`):
```bash
git rev-parse --abbrev-ref <target>@{upstream}
```

If this fails (no upstream configured) → skip the rest of 4b with a note in the plan: "step 4b skipped — no upstream for <target>".

Otherwise:
```bash
git fetch
git rev-list --count <target>..<target>@{upstream}
```

If count > 0 (local target is behind upstream) → STOP. Offer `git pull --ff-only` on target before proceeding.

Also check source the same way:
```bash
git rev-parse --abbrev-ref <source>@{upstream} 2>/dev/null && \
  git rev-list --count <source>..<source>@{upstream}
```

If source is behind its upstream — warn user but allow continuing.

### 4c. Diff name-status both sides

```bash
git diff --name-status <target>...<source>
git diff --name-status <source>...<target>
```

Show both lists. Highlight files that appear in BOTH (potential conflicts).

### 4d. Pre-conflict scan (cascading fallback)

**Phase 1: detect Git version**
```bash
git --version
```

**Phase 2: pick strategy by version**

- **Git ≥ 2.38** — modern `merge-tree`:
  ```bash
  git merge-tree --write-tree <target> <source>
  ```
  Parse output for conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`). Git auto-computes the merge base.

- **Git ≥ 2.0, < 2.38** — old `merge-tree`:
  ```bash
  base=$(git merge-base <target> <source>)
  git merge-tree "$base" <target> <source>
  ```
  Old format: blocks "changed in both" + conflict markers.

- **Any version — fallback via temp worktree** (if `merge-tree` failed or gave unreadable output):
  ```bash
  WT=/tmp/claude-conflict-check-$$
  git worktree add --detach "$WT" <target>
  git -C "$WT" merge --no-commit --no-ff <source> || true
  git -C "$WT" diff --name-only --diff-filter=U
  git -C "$WT" merge --abort 2>/dev/null || true
  git worktree remove --force "$WT"
  ```
  
  Uses `git -C <path>` to avoid `cd` (each Bash call resets working directory in Claude's model). `|| true` ensures cleanup runs even if merge fails.

**Results:**
- No conflicts → note "predicted conflicts: none", continue.
- Conflicts → show file list (+ blocks if available), ask user: proceed and resolve manually / cancel.
- Scan failed (worktree fallback also failed, e.g., shallow clone) → say "pre-conflict scan not performed", ask user to OK skipping.

**Edge case:** Git < 2.0 → say "Git version too old for safe pre-merge scan", ask user to confirm skipping the step.

## Step 5: Present consolidated plan

```
Merge plan:
  source: <branch> (<N commits>)
  target: <branch> (protected: true/false)
  strategy: <squash|merge|rebase|ff-only> (source: CLAUDE.md | user)

  Files changed: <N>
  Overlaps with target: <N>
  Predicted conflicts: none | <list>

  Commands:
    git checkout <target>
    git merge ... <source>

Confirm? (y/n)
```

Without `y`, do NOT execute.

## Step 6: Execute merge

| Strategy | Commands |
|---|---|
| squash | `git checkout <target> && git merge --squash <source>` then invoke `committing-changes` with context: "this is a squash merge — skip atomic-split check (Step 4), produce a single commit". |
| merge commit | `git checkout <target> && git merge --no-ff <source> -m "merge: <source> into <target>"` |
| rebase | `git checkout <source> && git rebase <target> && git checkout <target> && git merge --ff-only <source>` |
| ff-only | `git checkout <target> && git merge --ff-only <source>` (fails if not fast-forward — feature, not bug) |

## Step 7: Conflict during actual merge

- STOP. Do NOT run `git merge --abort` without confirmation.
- Show conflicts: `git diff --name-only --diff-filter=U`.
- Ask user: resolve manually / abort.
- Do NOT auto-edit conflict files.

## Step 8: Final report

- SHA of merge commit (or new HEAD for ff/rebase).
- Run `git status` and confirm clean.
- Brief: "Merged N commits into <target>".
- **Do NOT push.** Push is always separate, explicit.
- **Do NOT delete source branch** without explicit user command.

## Step 9: Forbidden without explicit user command

These require a separate, explicit user instruction:
- `git push` (any form)
- `--force` / `--force-with-lease`
- `git branch -D <source>`
- `git push origin --delete <source>`
- `git reset --hard`

If Claude is tempted to use any of these — STOP, report to user, wait for explicit go-ahead.

## Override phrases

| Phrase | Effect |
|---|---|
| "force squash/merge/rebase" | Set strategy manually, skip CLAUDE.md read |
| "skip checks" | Skip 4b (remote sync) and 4d (conflict pre-scan) only. 4a (clean working tree) and 4c (diff display) always run. |
| "force" | Does NOT enable `--force` push. Force-push still requires "yes force push to <branch>" |

## Absolute prohibitions

- Any `git config` modifications.
- `git push --force` to protected branch without typed phrase "yes force push to <branch>".
- `git reset --hard` if untracked files exist — first offer stash, then separate confirmation.

## Edge cases (handle and report)

- **Not a git repository**: report and stop. Never run `git init`.
- **Detached HEAD on target**: report and stop — cannot meaningfully merge into detached HEAD without explicit user intent.
- **No `origin` remote**: skip step 4b (up-to-date check) with a note in the plan. Local-only merge still proceeds.
- **Shallow clone** (`.git/shallow` exists): pre-conflict scan may be inaccurate — warn user before step 4d.
- **Submodules** in either branch: warn separately — submodule pointer changes can surprise users.
- **Existing `.git/MERGE_HEAD`**: STOP. A previous merge is unfinished. Ask user to resolve or abort it first.
- **Existing `.git/rebase-merge/` or `.git/rebase-apply/`**: STOP. A previous rebase is unfinished. Ask user to resolve (`git rebase --continue`) or abort (`git rebase --abort`) first.
- **Pre-conflict scan fails entirely** (worktree fallback also fails): say "pre-merge conflict prediction unavailable", require explicit user OK to proceed without it.
