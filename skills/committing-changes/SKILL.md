---
name: committing-changes
description: Use BEFORE creating any git commit. Enforces Conventional Commits format, analyzes staged changes to propose atomic splits, validates safety (no secrets, no --amend, no --no-verify unless explicitly requested), and removes Co-Authored-By trailer. Trigger phrases — "commit this", "make a commit", or any user request that implies creating a commit.
---

# Committing Changes

When invoked, follow this procedure step by step. Create a TodoWrite task for each step and complete them in order.

## Rule Hierarchy

Apply rules in this priority order (highest first):

1. Explicit user instruction in the current message — always wins.
2. Project conventions file — `CLAUDE.md` or `AGENTS.md` — sections `## Git workflow`, `## Git conventions`, `## Commit rules`.
3. Global conventions — `~/.claude/CLAUDE.md` (or your harness's global agents file, e.g. `~/.codex/AGENTS.md`).
4. This SKILL.md — process defaults.

## Step 1: Read conventions

- Read the project conventions file if it exists — `CLAUDE.md` or `AGENTS.md`. Look for sections: `## Git workflow`, `## Git conventions`, `## Commit rules`.
- Read the global conventions file — `~/.claude/CLAUDE.md` (or the equivalent for your harness).
- Project rules override global rules.

## Step 2: Gather state (parallel)

Run these Bash commands in parallel:
- `git status` (do NOT use `-uall` flag)
- `git diff --staged`
- `git diff` (unstaged)
- `git log -n 10 --oneline`

If `git status` shows the working tree is not a git repo — STOP. Report to user. Do not run `git init`.

## Step 3: Safety scan on staged files

Block the commit and report to user if any of these are in staged files:

**Suspicious filenames:**
- `.env*`, `*credentials*`, `*secret*`, `*.pem`, `*.key`, `id_rsa*`

**Suspicious content patterns (grep staged diff):**
- `AWS_[A-Z_]+ ?= ?['\"][A-Za-z0-9/+=]{20,}`
- `sk-[a-zA-Z0-9]{20,}` (OpenAI/Anthropic)
- `ghp_[A-Za-z0-9]{20,}` (GitHub PAT)
- `xox[baprs]-[A-Za-z0-9-]{10,}` (Slack)
- `-----BEGIN.*PRIVATE KEY-----`

**Garbage files:**
- `.DS_Store`, `Thumbs.db`, anything inside `node_modules/`, `__pycache__/`, `.next/`, `dist/`, `build/`

**Large binaries:** any staged file >1MB (check with `git diff --staged --stat`).

**Tiered response:**
- **Secrets and credentials (suspicious filenames + content patterns above)** — unconditional STOP. NO bypass, even on user request. List offending files/lines and refuse to commit until they are removed from staging.
- **Garbage files and large binaries** — STOP, list them, ask user whether to proceed despite the warning.

## Step 4: Analyze for atomic split

**Cross-skill context:** If this skill was invoked by another skill (e.g., `merging-branches`) with context containing "skip atomic-split check" or "squash merge" — skip this entire step and proceed directly to Step 5. The caller has determined a single commit is correct.

Inspect staged changes for "should this be multiple commits" heuristics:

- Changes touch ≥2 unrelated top-level directories (e.g., `src/auth/*` AND `src/billing/*`)
- Mixed categories: code + independent docs + independent config
- Multiple independent fixes in different files

If a split is warranted, present the logical groups to the user:

```
Found N logical changes:
1. feat(auth): add OAuth provider — src/auth/oauth.ts, src/auth/oauth.test.ts
2. docs(api): update API documentation — docs/api/*.md
3. chore(deps): bump dependencies — package.json, pnpm-lock.yaml

Split into N commits? (yes / no / custom)
```

Wait for user response.
- "yes" → proceed with N separate commits, each through Steps 5–10
- "no" → proceed as a SINGLE commit with all staged changes (continue to Step 5)
- "custom" → ask user to specify the grouping
- No response → do not commit until user clarifies

If no split was warranted in the first place (single logical group or small changes) — proceed as one commit, no question asked.

## Step 5: Compose the message

Format: `type(scope): subject`

**Allowed types (10 common Conventional Commits types):**

| type | when |
|---|---|
| `feat` | new functionality |
| `fix` | bug fix |
| `docs` | documentation only |
| `style` | formatting, whitespace, code style (NOT logic changes) |
| `refactor` | refactor without feat/fix |
| `test` | tests only |
| `chore` | misc (deps, gitignore, build config) |
| `perf` | performance |
| `build` | build system |
| `ci` | CI/CD |

**Subject rules:**
- imperative mood ("add" not "added")
- lowercase
- no trailing period
- ≤72 chars
- English

**Body is REQUIRED when:**
- Commit touches files in ≥2 different code areas
- The decision is non-obvious (workaround, unusual approach)
- Breaking change (then include `BREAKING CHANGE:` footer)

Body answers "what AND why" — not "how" (the diff shows how).

## Step 6: Pre-action validation

Strip / refuse the following:

- ❌ `Co-Authored-By` trailer — REMOVE if present in any draft message
- ❌ `--amend` — only allowed if user explicitly said "amend"
- ❌ `--no-verify` — only allowed if user explicitly asked to skip hooks
- ❌ `git add -A` / `git add .` — stage only explicit files by path. If the user's request implies "add everything," enumerate files from `git status` output and stage each explicitly: `git add file1 file2 file3`.

## Step 7: Show the plan, wait for approval

Format:
```
Plan (single commit):
- git add <files>
- git commit -m "<subject>"
  <optional body preview>

Plan (split into N commits — if user confirmed split in Step 4):
1. git add <files-group-1>
   git commit -m "<subject-1>"
2. git add <files-group-2>
   git commit -m "<subject-2>"
... (one block per group)

Confirm? (y/n)
```

Without `y` (or equivalent affirmation), do NOT commit.

## Step 8: Execute via HEREDOC

For multi-line messages, use HEREDOC to preserve formatting:

```bash
git commit -m "$(cat <<'EOF'
<subject>

<body>
EOF
)"
```

For one-line subject, `-m "subject"` is fine.

## Step 9: Pre-commit hook failure

If the pre-commit hook fails — the commit did NOT happen.

- Do NOT run `--amend` (it would modify the previous, unrelated commit).
- Fix the underlying problem.
- Re-stage the relevant files.
- Create a NEW commit with the same intended message.

If the hook fails a second time — STOP. Report to user. Do NOT try a third time.

## Step 10: Final report

- Run `git status` and confirm it's clean.
- Report SHA + subject on one line: `c726baa docs(specs): add git skills design spec`.
- If multiple commits (from split), list each one.
- **Do NOT push.** Push is always a separate, explicit user command.

## Override phrases

If the user explicitly uses one of these phrases, follow the override without re-asking:

| Phrase | Effect |
|---|---|
| "skip checks" | Skip safety scan, but secret detection still blocks |
| "amend" | Allow `--amend` |
| "no-verify" | Allow `--no-verify` |

## Absolute prohibitions

Never, even on explicit request:
- Any `git config` modifications.

Requires special confirmation beyond regular y/n:
- Committing files matched by `.gitignore` as explicit paths → warning + confirmation.

## Edge cases (handle and report)

- **Detached HEAD** (`git symbolic-ref HEAD` fails): report state to user, ask whether to proceed (commit will be unreachable without a branch).
- **Submodules in staged** (`git diff --staged --submodule`): warn separately — submodule commits behave differently, confirm intent.
- **Existing `.git/MERGE_HEAD`** (unfinished merge): STOP. Tell user a previous merge is in progress, ask how to handle.
- **Pre-commit hook failing in a loop**: if hook fails twice in a row, STOP after the 2nd attempt. Do not retry. Ask user.
- **Not a git repository**: report and stop. Never run `git init`.
