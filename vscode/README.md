# VS Code setup for Claude Code — terminal settings, `shift+enter`, Nerd Font, and 20 extensions (macOS)

A ready-made **VS Code configuration** tuned for running an AI coding agent in the integrated
terminal: a `terminal.integrated` block built for long-running TUI sessions, `shift+enter` bound to
insert a newline instead of sending, the **MesloLGS Nerd Font** for correct glyph rendering, and a
curated list of 20 extensions. One script, and a fresh macOS machine has the same environment. Part of
the [OFFICINA](../) repository.

Only **deviations from the defaults** are published here — every key in `settings.json` is there
because the default was wrong for this workflow, not because it was exported wholesale.

## Quick start

```bash
git clone https://github.com/KirKruglov/OFFICINA.git && cd OFFICINA
./install/vscode.sh
```

The script checks for the `code` CLI, installs the font via Homebrew, copies `settings.json` and
`keybindings.json` into VS Code's user directory (backing up whatever was there), and installs all 20
extensions. Then **fully restart VS Code** — quit and reopen, not just reload the window; the terminal
font and profiles are read at startup.

Prerequisites: macOS, VS Code with the `code` CLI on your `PATH` (`Cmd+Shift+P` →
`Shell Command: Install 'code' command in PATH`), and Homebrew for the font.

## Why VS Code needs tuning for a terminal agent

Running Claude Code in the integrated terminal isn't the use case VS Code's terminal defaults were
designed for. The defaults assume short-lived shells: you run a command, read a few lines, and move
on. An agent session is the opposite — it runs for hours, prints thousands of lines, redraws a
full-screen TUI, and expects multi-line input.

Four defaults actively get in the way. `Enter` submits, so there's no way to type a paragraph.
Scrollback truncates long output. Contrast correction repaints the TUI's own colors. And a terminal
font without box-drawing and icon glyphs turns diffs and status lines into mojibake. Each key below
fixes exactly one of those.

## The terminal settings that matter

### `shift+enter` — a newline instead of sending

The single most useful binding in the whole setup. Without it, every `Enter` submits the message and
long prompts have to be written elsewhere and pasted in.

```json
{
  "key": "shift+enter",
  "command": "workbench.action.terminal.sendSequence",
  "args": { "text": "\u001b\r" },
  "when": "terminalFocus"
}
```

`\u001b\r` is `ESC` followed by carriage return — the sequence the Claude Code TUI reads as "insert a
line break, don't submit". The `terminalFocus` condition keeps the binding out of the editor, where
`shift+enter` still does what it normally does.

### Keeping the TUI's own colors

```json
"terminal.integrated.minimumContrastRatio": 1
```

VS Code raises foreground contrast against the background by default, which is helpful for ordinary
shell output and destructive for a TUI that paints its own palette — diff highlighting and syntax
colors come out washed out or wrong. Setting the ratio to `1` disables the correction and hands the
colors back to the application.

### Scrollback, persistent sessions, and sticky scroll

```json
"terminal.integrated.scrollback": 10000,
"terminal.integrated.persistentSessionScrollback": 2000,
"terminal.integrated.enablePersistentSessions": true,
"terminal.integrated.persistentSessionReviveProcess": "never",
"terminal.integrated.stickyScroll.enabled": true,
"terminal.integrated.shellIntegration.decorationsEnabled": "both"
```

10 000 lines of scrollback instead of the default 1 000 — an agent session generates output fast
enough that the default loses the beginning of the run. Persistent sessions keep the buffer across
window reloads, while `persistentSessionReviveProcess: "never"` deliberately does *not* restart the
process: you keep the transcript to read, without an agent silently resuming on its own. Sticky scroll
pins the current command to the top, and shell-integration decorations mark where each command started
and whether it succeeded.

### Three named session profiles

```json
"terminal.integrated.profiles.osx": {
  "Session 1": { "path": "zsh", "color": "terminal.ansiBlue",    "icon": "terminal" },
  "Session 2": { "path": "zsh", "color": "terminal.ansiGreen",   "icon": "terminal-tmux" },
  "Session 3": { "path": "zsh", "color": "terminal.ansiMagenta", "icon": "terminal-bash" }
},
"terminal.integrated.tabs.title": "${process}${separator}${cwdFolder}",
"terminal.integrated.tabs.description": "${cwdFolder}"
```

Parallel agent sessions are the normal case, and identical tabs labelled `zsh` make them unusable.
Three profiles in distinct colors, plus a tab title carrying the process and the working folder, make
the right tab findable at a glance. `splitCwd: "workspaceRoot"` means a split opens at the repository
root instead of inheriting wherever the previous shell wandered off to, and
`confirmOnExit: "hasChildProcesses"` catches the accidental close of a window with a session still
running.

### An audible bell when the agent finishes

```json
"accessibility.signals.terminalBell": { "sound": "on" }
```

Claude Code rings the terminal bell when it finishes a response. With the signal enabled, you can look
away during a long run and get called back — the difference between watching a progress indicator and
doing something else.

### The rest

| Key | Value | Why |
|---|---|---|
| `terminal.integrated.fontFamily` | `MesloLGS Nerd Font, Menlo, monospace` | Glyph coverage, with fallbacks |
| `terminal.integrated.gpuAcceleration` | `auto` | Keeps long output scrolling smoothly |
| `terminal.integrated.mouseWheelScrollSensitivity` | `3` | Faster travel through a 10 000-line buffer |
| `files.autoSave` / `files.autoSaveDelay` | `afterDelay` / `1000` | Fewer conflicts when the agent edits files you also have open |
| `explorer.fileNesting.patterns` | lockfiles, maps, `.d.ts`, SQLite sidecars | Collapses generated siblings under their source file |
| `extensions.ignoreRecommendations` | `true` | No per-repository extension nag |
| `extensions.autoUpdate` / `autoCheckUpdates` | `on` / `false` | Updates apply, but no background checking mid-session |
| `workbench.startupEditor` | `none` | Opens into the workspace, not the welcome page |
| `workbench.colorTheme` | `Dark Modern` | Built-in theme — no extension dependency |

## MesloLGS Nerd Font

The terminal font isn't cosmetic here. Claude Code's TUI, a status line, and any git prompt draw with
box-drawing characters and icon glyphs that a standard font doesn't carry — without them you get
replacement boxes through the whole interface. MesloLGS Nerd Font is Menlo patched with the Nerd Font
glyph set, so it stays legible at small sizes while covering everything the TUI draws.

The installer handles it:

```bash
brew install --cask font-meslo-lg-nerd-font
```

Without Homebrew the script warns and continues — install the font by hand, or the terminal will
render incorrectly. `Menlo, monospace` is listed as a fallback in `fontFamily` so the setup degrades
to plain text rather than breaking outright.

## The 20 extensions

One ID per line in [`extensions.txt`](extensions.txt), installed by the script with
`code --install-extension`. Curated, not exported wholesale — nothing here is installed "just in
case".

| Purpose | Extensions |
|---|---|
| **Agent** | `anthropic.claude-code` |
| **Git** | `eamodio.gitlens`, `github.vscode-github-actions` |
| **JS / TS** | `dbaeumer.vscode-eslint`, `esbenp.prettier-vscode`, `ms-vscode.vscode-typescript-next` |
| **Python** | `ms-python.python`, `ms-python.vscode-pylance`, `ms-python.debugpy`, `ms-python.vscode-python-envs` |
| **Jupyter** | `ms-toolsai.jupyter`, `ms-toolsai.jupyter-keymap`, `ms-toolsai.jupyter-renderers`, `ms-toolsai.vscode-jupyter-cell-tags`, `ms-toolsai.vscode-jupyter-slideshow` |
| **Containers** | `ms-azuretools.vscode-docker`, `ms-azuretools.vscode-containers`, `ms-vscode-remote.remote-containers` |
| **Data & docs** | `qwtel.sqlite-viewer`, `yzhang.markdown-all-in-one` |

Install them without touching your settings:

```bash
while read -r ext; do code --install-extension "$ext" --force; done < vscode/extensions.txt
```

## Files and where they go

| File | Installs to | What it is |
|------|-------------|------------|
| [`settings.json`](settings.json) | `~/Library/Application Support/Code/User/settings.json` | Curated user settings (JSONC, comments preserved) |
| [`keybindings.json`](keybindings.json) | `~/Library/Application Support/Code/User/keybindings.json` | Key bindings — currently just `shift+enter` |
| [`extensions.txt`](extensions.txt) | — | Extension IDs, one per line, from `code --list-extensions` |

Install script — [`../install/vscode.sh`](../install/vscode.sh). Platform — **macOS only**: the
destination path, the profile key (`profiles.osx`), and the Homebrew font cask are all
macOS-specific. On Linux or Windows the settings themselves transfer; the paths and the font
installation don't.

## What the installer does

[`install/vscode.sh`](../install/vscode.sh) runs under `set -euo pipefail` and refuses to do anything
half-way:

1. **Checks the environment** — bails out if the platform isn't macOS, if the `code` CLI is missing,
   or if any of the three source files isn't where it expects.
2. **Installs the font** — skips if the cask is already present; warns and continues if Homebrew
   isn't installed.
3. **Backs up, then copies** — an existing `settings.json` or `keybindings.json` is copied to
   `*.bak-<timestamp>` before being replaced. Nothing is overwritten silently, and the backup is
   timestamped so repeated runs don't clobber each other.
4. **Installs the extensions** — one `code --install-extension … --force` per line.

The model is **copy, not symlink**: this is a configuration for setting up another machine, not a
live-linked dotfiles repo. That choice is deliberate and recorded — see
[`docs/officina-public-repo-decisions.md`](../docs/officina-public-repo-decisions.md).

## Updating from the live config

When the settings change on the machine, the flow is one-directional — edit on the machine, then
re-curate into the repository:

1. Copy `~/Library/Application Support/Code/User/settings.json` and re-apply the curation: drop keys
   owned by extensions, machine-specific paths, and anything questionable to publish. Only deviations
   from the defaults stay.
2. Run `code --list-extensions`, rebuild `extensions.txt`, and remove the excluded ones.

`*.bak` files, an empty `snippets/`, `tasks.json`, and `mcp.json` are never published.

## FAQ

### How do I get a newline in Claude Code instead of sending the message?

Bind `shift+enter` to `workbench.action.terminal.sendSequence` with the text `\u001b\r` and the
condition `terminalFocus` — the full snippet is in [`keybindings.json`](keybindings.json). If it still
sends, check that the binding isn't shadowed by another `shift+enter` entry and that the terminal, not
the editor, has focus.

### Why does the Claude Code TUI look washed out in the VS Code terminal?

`terminal.integrated.minimumContrastRatio` defaults to a value that repaints foreground colors for
readability, which overrides the palette the TUI draws with. Set it to `1` and the application's own
colors come through.

### Do I need a Nerd Font?

For correct rendering, yes. Box-drawing and icon glyphs used by the TUI, the status line, and most git
prompts aren't in standard monospace fonts, and they render as replacement boxes without one. If you
can't install a font, everything still works — it just looks broken.

### Will the installer overwrite my existing VS Code settings?

It replaces the two files, but copies each one to `settings.json.bak-<timestamp>` first. To merge
instead of replace, open [`settings.json`](settings.json) and copy across only the keys you want —
every one of them is a deliberate deviation from a default, so they're readable in isolation.

### Does this work on Linux or Windows?

The settings do; the installer doesn't. The destination path, `terminal.integrated.profiles.osx`, and
the Homebrew font cask are macOS-specific. Copy the relevant keys by hand and install MesloLGS Nerd
Font through your own package manager.

### Why not use VS Code Settings Sync?

Settings Sync mirrors a machine, including everything installed by accident. This is a curated
configuration meant to be read and understood before it's applied — 20 chosen extensions, and only
the settings that differ from the defaults on purpose.

## Related

- [OFFICINA](../) — the method and the rest of the toolkit
- [`claude/`](../claude/) — Claude Code subagents, `CLAUDE.md` layers, `settings.json`, status line
- [`skills/`](../skills/) — portable `SKILL.md` procedures for the agent running in this terminal
- [`install/`](../install/) — the installer scripts
