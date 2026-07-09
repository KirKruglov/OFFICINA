# Settings — sanitized global Claude Code configuration

The global layer of my [Claude Code](https://claude.com/claude-code) setup: personal preferences,
permissions, a custom status line, and two `CLAUDE.md` examples. Part of the
[OFFICINA](../../) repository.

These files are examples, not a drop-in profile — copy what fits and adapt it. Paths are
genericized (`~/.claude/…`) and no secrets are included.

## Files

| File | Layer | What it is |
|------|-------|------------|
| [`CLAUDE.global.md`](CLAUDE.global.md) | Global (`~/.claude/`) | Personal preferences applied to **every** project — communication style, default behavior, git conventions. Short and stable. |
| [`CLAUDE.project.md`](CLAUDE.project.md) | Project (repo root) | A reusable **template** for a project's root `CLAUDE.md`. Placeholders (`<PROJECT_NAME>`, `<AREA_DIR>`, …) are filled per project; the methodological core carries over unchanged. |
| [`settings.json`](settings.json) | Global (`~/.claude/`) | Global Claude Code settings: permission allowlist, status line, enabled plugins, effort level, TUI. |
| [`settings.local.json`](settings.local.json) | Project-local | Example of machine-/repo-local overrides (normally git-ignored): a few extra git permissions and `.env` denials. Shown as a structure example. |
| [`statusline-command.sh`](statusline-command.sh) | Global (`~/.claude/`) | Custom status line script — more informative than the default. |

## Two `CLAUDE.md` files — why

Claude Code reads `CLAUDE.md` at two levels, and they do different jobs:

- **`CLAUDE.global.md`** → `~/.claude/CLAUDE.md`. Cross-project *preferences* (how I want Claude to
  behave everywhere). Loaded in every session, regardless of project.
- **`CLAUDE.project.md`** → a project's root `CLAUDE.md`. Project-specific *navigation and rules*
  (repo layout, where to look, per-area conventions). Loaded only in that project.

They compose: global sets the baseline, project layers specifics on top. Kept as separate files here
so each is readable on its own.

## `settings.json` — annotated

Non-obvious keys and why they're set the way they are:

| Key | Value | What it does |
|-----|-------|--------------|
| `permissions.allow` | read-only + safe git/curl | Auto-approves a curated allowlist (Read, search, `git` read commands, `curl -s`) so routine calls don't prompt. |
| `permissions.defaultMode` | `auto` | Auto-approves allowlisted calls; still prompts for anything outside the list. |
| `effortLevel` | `high` | Reasoning depth per turn. Higher = more thorough, slower. |
| `tui` | `fullscreen` | Full-screen terminal UI instead of inline. |
| `editorMode` | `normal` | Standard editing (vs `vim` keybindings). |
| `autoCompactEnabled` | `true` | Auto-summarizes context when the window fills, so long sessions continue. |
| `enabledPlugins` | code-review, frontend-design, superpowers | Plugins loaded from the official marketplace (`extraKnownMarketplaces`). |
| `agentPushNotifEnabled` | `true` | Push notifications on background-agent events. |

⚠️ **Safety-sensitive flags — review before copying:**

| Key | Effect |
|-----|--------|
| `skipDangerousModePermissionPrompt` | Suppresses the confirmation before entering dangerous (bypass-permissions) mode. |
| `skipAutoPermissionPrompt` | Suppresses the auto-approval prompt. |
| `skipWorkflowUsageWarning` | Hides the multi-agent workflow token-cost warning. |

These lower friction at the cost of guardrails. They fit a solo, trusted-environment workflow — copy
them only if that matches yours.

## Status line

`statusLine.command` points at `~/.claude/statusline-command.sh` (the sanitized form of the original
absolute path — drop the script into `~/.claude/` for it to resolve). It requires
[`jq`](https://jqlang.github.io/jq/), and the branch glyph needs a
[Nerd Font](https://www.nerdfonts.com/). What it looks like in the terminal:

```
Opus 4.8 (1M context) (high) | █░░░░░░░░░ 9% (89k/1000k) | 5h:6% ~4h11m | wk:48% ~7h31m | ⎇ release/0.1.0
```

Segment by segment (`|`-separated), each color-coded in the terminal:

| Segment | Meaning | Color |
|---------|---------|-------|
| `Opus 4.8 (1M context) (high)` | Active model + effort level | cyan + dim gray |
| `█░░░░░░░░░ 9% (89k/1000k)` | Context-window usage: bar, percent, used/total tokens | green <50% → yellow ≥50% → red ≥80% |
| `5h:6% ~4h11m` | 5-hour rate-limit usage + time to reset | magenta → yellow ≥50% → red ≥80% |
| `wk:48% ~7h31m` | Weekly rate-limit usage + time to reset | blue → yellow ≥50% → red ≥80% |
| `⎇ release/0.1.0` | Current git branch | gray |

## How to use

Copy individual files into `~/.claude/` (global layer) or a project's `.claude/` (project layer)
and adapt. `CLAUDE.project.md` is a template — start from it, replace the `<...>` placeholders, and
delete the optional blocks that don't apply.
