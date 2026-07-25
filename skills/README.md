# Claude Code Skills — portable `SKILL.md` modes and procedures for AI coding agents

Reusable **agent skills**: self-contained modes and procedures an AI coding agent invokes mid-session.
Each one is a folder with a `SKILL.md` — plain Markdown plus YAML frontmatter — so it isn't tied to
any single runtime. These run in **Claude Code** and in any harness that reads the `SKILL.md` format.
Four skills ship here: Conventional Commits, safe branch merging, a read-only discussion mode, and
release finalization. Part of the [OFFICINA](../) repository.

## What a skill is

A skill is a written procedure the agent loads on demand. It costs nothing until it's needed: the
agent reads only the `name` and `description` from the frontmatter, and pulls the body into context
when the situation matches. What lands in the body is the part you'd otherwise re-explain every
session — the order of steps, the checks that must not be skipped, the phrasing conventions, the
point where the agent has to stop and ask.

That makes a skill the right shape for **context-dependent work with a fixed procedure**. Work with
no decisions in it belongs in a script; a whole role belongs in a subagent. See
[Skill, subagent, CLI, or loop?](#skill-subagent-cli-or-loop) below.

## The skills in this folder

| Skill | Invoke | What it does |
|-------|--------|--------------|
| [`committing-changes`](committing-changes/SKILL.md) | auto — before any commit | Conventional Commits procedure with a secret scan and atomic-split proposal |
| [`discuss`](discuss/SKILL.md) | `/discuss` (manual only) | Read-only thinking-partner mode for Q&A, ideation, and design debate |
| [`merging-branches`](merging-branches/SKILL.md) | auto — before merge/rebase | Safe branch integration with pre-merge checks and protected-branch gates |
| [`release-finalize`](release-finalize/SKILL.md) | `/release-finalize [version]` | Changelog, README sync, pre-publish scan, annotated git tag |

### `committing-changes` — Conventional Commits, enforced

Agents write commit messages that describe the diff instead of the change, bundle three unrelated
edits into one commit, and occasionally stage a `.env` file. This skill fires before any commit and
runs the same procedure every time: read the project's conventions file (`CLAUDE.md` or `AGENTS.md`,
falling back to the global one), gather `git status` and both diffs in parallel, then scan the staged
files.

The scan is tiered. Credentials — `sk-…`, `ghp_…`, `xox…`, AWS keys, private-key blocks, `.env*`,
`*.pem` — are an unconditional stop with no bypass, even on request. Garbage and files over 1 MB stop
and ask. What survives is analyzed for an **atomic split**: if the staging area holds two unrelated
changes, the skill proposes two commits. The message is drafted as `type(scope): subject`, the
`Co-Authored-By` trailer is stripped, and nothing is committed until you approve it. It never pushes,
never amends, and never passes `--no-verify` unless you say so explicitly.

### `discuss` — a read-only thinking-partner mode

Not a procedure but a **standing mode**: once invoked with `/discuss`, every reply follows its rules
until you leave. Short answers, one topic per reply, one question at a time — asked as an explicit
`A / B / C` choice whenever the options can be enumerated. The agent holds its own position and
argues for it instead of agreeing to be agreeable, and when asked for ideas it puts at least three
distinct alternatives on the table before naming its pick.

The mode is strictly **read-only**: no files created, edited, or deleted, no commands with side
effects, until you exit with `exit discussion` or `/discuss off`. Reading, searching, and web access
stay open. Every reply carries the verbatim `💬 [discussion]` marker so the mode is never ambiguous.
It's manual-only — `disable-model-invocation: true` keeps the agent from wandering into it on its own.

### `merging-branches` — pre-merge safety gates

Runs before any merge, rebase, or branch integration. First it works out the **strategy** rather than
assuming one: it reads `## Git workflow`, `## Branching`, or `## Release process` from the project
conventions file, quotes what it found back to you, and asks you to confirm the interpretation. With
no guidance anywhere, it asks outright — squash, merge commit, rebase, or fast-forward only.

Source and target both come from your request; the skill never guesses from the current branch. A
target on `main`, `master`, `release/*`, `production`, `prod`, or `staging` is treated as protected
and needs an exact typed phrase (`yes merge to main`) — approval that can't be given by reflex. Then
four pre-merge checks run in order: clean working tree, up to date with the remote, a `name-status`
diff of both sides, and a version-aware conflict pre-scan. Any failure stops the run. It never pushes
and never deletes branches unasked.

### `release-finalize` — from working tree to annotated tag

Gathers the release artifacts of a product and gets the repository ready for a public push or a
deploy. It prepends a new changelog section in **Keep a Changelog** format with a SemVer version,
syncs the root `README.md` to the release context, runs a pre-publish scan for secrets, version
consistency, and a clean tree, and creates an annotated git tag.

It's universal by default — release artifacts are auto-detected, and an optional
`.release-finalize.yml` overrides the detection per repository. One checkpoint before any edits are
applied, then a straight run to the tag. It never merges, pushes, or deploys: the last thing it does
is print the commands you'd run next.

### How these skills chain

The git skills call each other rather than duplicating logic. On a squash merge, `merging-branches`
hands off to `committing-changes` to produce the single commit — one message convention, one safety
scan, no second copy of either. `release-finalize` reuses the same pair: it delegates its commit
phase to `committing-changes`, then points at `merging-branches` for the release-branch merge in its
hand-off block.

## Anatomy of a `SKILL.md`

Every skill is one folder with one `SKILL.md`. The file is Markdown with a YAML frontmatter block:

```yaml
---
name: release-finalize
description: Finalize a product release and prepare the repo for a public push or deploy — updates
  the changelog and root README, scans for secrets and version consistency, creates an annotated git
  tag. Does NOT merge, push, or deploy. Manual-invocation only via /release-finalize.
disable-model-invocation: true
argument-hint: "[version]"
allowed-tools: Read, Grep, Edit, Bash(git:*), TodoWrite
---

# Release Finalize

Everything below the frontmatter is the procedure the agent follows.
```

| Field | Required | What it controls |
|---|---|---|
| `name` | yes | The skill's identifier; matches the folder name and the `/name` command |
| `description` | yes | The only text always in context — it decides whether the skill gets loaded at all. Write it as trigger conditions, not as a summary |
| `disable-model-invocation` | no | `true` makes the skill manual-only: it fires on `/name` and never self-activates |
| `argument-hint` | no | Argument placeholder shown at the command prompt, e.g. `"[version]"` |
| `allowed-tools` | no | Restricts the tool pool while the skill runs, e.g. `Bash(git:*)` |

The body is the procedure itself: numbered steps, the exact commands to run, the conditions that
stop the run, and the wording of anything the agent prints. Anything a skill needs beyond that —
templates, checklists, reference tables — lives as a separate file in the same folder and is read on
demand.

## How to install a skill

Drop the skill's folder into your agent's skills directory:

```
~/.claude/skills/
└── committing-changes/
    └── SKILL.md
```

- **Claude Code** — `~/.claude/skills/<name>/` makes it available everywhere; a project's
  `.claude/skills/<name>/` scopes it to that repository.
- **Other harnesses** — the equivalent skills location for your runtime. The `SKILL.md` content is
  harness-neutral; only the directory path changes.

Copy one skill out of this repository without cloning anything else:

```bash
git clone --depth 1 https://github.com/KirKruglov/OFFICINA.git /tmp/officina
cp -R /tmp/officina/skills/committing-changes ~/.claude/skills/
```

Restart the session, or start a new one, for the agent to pick up the new folder.

## Auto-invoked vs manual skills

Two invocation modes, and the choice is a design decision rather than a preference.

**Auto-invoked** skills — `committing-changes`, `merging-branches` — carry no
`disable-model-invocation` field. The agent matches the situation against the `description` and loads
them on its own. This is right when the skill guards an operation that must never happen unguarded: a
commit, a merge. If it only ran when you remembered to ask for it, it wouldn't be a guardrail.

**Manual-only** skills — `discuss`, `release-finalize` — set `disable-model-invocation: true` and
fire only on their slash command. This is right when the skill changes how the whole session behaves,
or performs an irreversible act. A discussion mode the agent could enter by itself would be a mode
you never chose; a release that tagged itself would be a release nobody approved.

If an auto-invoked skill isn't triggering when you expect it to, the `description` is almost always
the reason. It's the only part of the file that's always in context — write it as the conditions
under which the skill applies, including the phrases a user would actually type, not as a neutral
summary of what it does.

## Skill, subagent, CLI, or loop?

Skills are one of four artifact types in the OFFICINA method. The choice starts from a single
question — what does the task actually need?

| Need | Artifact | Why |
|---|---|---|
| A deterministic action, no judgment | **CLI** | A script is cheaper, faster, and can't improvise. See [`jig`](../cli/jig/) |
| Read the context, then decide by situation | **Skill** | The procedure is fixed; the decisions inside it aren't |
| An isolated subtask or a persona role | **Subagent** | Its own context window and its own tool pool. See [`claude/`](../claude/) |
| A recurring cycle on a schedule | **Loop** | Needs budgets and a stop condition more than it needs a procedure |

Putting a model where a script would do adds cost and uncertainty; putting a script where judgment is
required produces confident nonsense. The full reasoning and the authoring guides live in
[`methodology/`](../methodology/).

## FAQ

### Do these skills work outside Claude Code?

Yes. A `SKILL.md` is Markdown with YAML frontmatter and carries no Claude Code-specific syntax in its
body. Any harness that reads the format can run them; only the installation path differs. The tool
names referenced in `allowed-tools` are the one place a harness may need a translation.

### Where do Claude Code skills live?

`~/.claude/skills/<name>/SKILL.md` for skills available in every project, or
`<project>/.claude/skills/<name>/SKILL.md` for skills scoped to one repository. The folder name
should match the `name` field in the frontmatter.

### Why isn't my skill triggering?

Check the `description` first — it's the only text the agent sees before deciding to load the skill,
so it has to state the trigger conditions explicitly. Then check for
`disable-model-invocation: true`, which makes the skill manual-only by design. Then confirm the
folder is in a skills directory the harness actually reads, and that the session was started after
the file appeared.

### Can a skill be restricted to certain tools?

Yes — `allowed-tools` in the frontmatter narrows the tool pool for as long as the skill runs.
`release-finalize` uses `Bash(git:*)` to permit git and nothing else. Note that `discuss` takes the
opposite approach: its read-only rule is behavioral, enforced by the procedure rather than by the
tool pool, because the mode still needs unrestricted reading and search.

### What's the difference between a skill and a slash command?

A slash command is an invocation mechanism; a skill is the content it invokes. A skill with
`disable-model-invocation: true` is reachable only as `/name` and behaves like a command. Without
that field, the same skill is also available to the agent on its own initiative — one file, two ways
in.

### Can I take one skill without adopting the rest of the repository?

Yes. Each folder is self-contained: copy it into your skills directory and it works. The only
coupling is between the git skills, which hand off to each other — `merging-branches` and
`release-finalize` both delegate their commit phase to `committing-changes`, so take that one along
if you use either.

## Related

- [OFFICINA](../) — the method and the rest of the toolkit
- [`methodology/`](../methodology/) — the guides behind these artifacts, skill authoring included
- [`claude/`](../claude/) — Claude Code subagents and settings
- [`cli/`](../cli/) — deterministic tools for the work that needs no model