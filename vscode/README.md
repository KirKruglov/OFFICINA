# VS Code config for Claude Code — curated settings, keybindings, and extensions (macOS)

A ready-made VS Code configuration tuned for working with Claude Code in the terminal: a configured
`terminal.integrated`, `shift+enter` for inserting a newline in the TUI, the MesloLGS Nerd Font, and
a curated set of 20 extensions. One script, and a fresh macOS machine gets the same environment.
Part of the [OFFICINA](../) repository.

## Quick start

```bash
git clone https://github.com/KirKruglov/OFFICINA.git && cd OFFICINA
./install/vscode.sh
```

The script checks for the `code` CLI, installs the MesloLGS Nerd Font, copies the settings (existing
files are backed up to `*.bak-<timestamp>`), and installs the extensions. After installation, fully
restart VS Code (quit and reopen).

Requires an installed VS Code with the `code` CLI on your PATH (`Cmd+Shift+P` →
`Shell Command: Install 'code' command in PATH`) and Homebrew for the font.

## What's inside

- **Terminal tuned for Claude Code** — session profiles, persistent sessions, larger scrollback,
  meaningful tab titles.
- **`shift+enter` → newline** — works in the Claude Code TUI instead of sending.
- **MesloLGS Nerd Font** — correct rendering of icons and box-drawing in the terminal, installed
  automatically.
- **20 extensions** — a curated set with no clutter; installed with a single command.
- **File nesting and auto-save** — grouped files in the tree, `autoSave` with a delay.
- **Safe installation** — existing settings are not overwritten silently but backed up.

## Files

| File | Purpose | Source |
|------|---------|--------|
| `settings.json` | Curated user settings (JSONC) | `~/Library/Application Support/Code/User/settings.json` |
| `keybindings.json` | Key bindings; `shift+enter` → newline in the TUI | `~/Library/Application Support/Code/User/keybindings.json` |
| `extensions.txt` | Extension list, one ID per line (20 total) | `code --list-extensions` |

Install script — [`../install/vscode.sh`](../install/vscode.sh). Platform — macOS only.

## Updating from the live config

When settings change on the machine:

1. Copy `~/Library/Application Support/Code/User/settings.json`, re-apply curation (remove
   extension-owned keys and settings that are questionable to publish).
2. `code --list-extensions` → rebuild `extensions.txt`, drop the excluded ones.

---

**Keywords:** VS Code dotfiles, VS Code settings.json, Claude Code setup, terminal config, MesloLGS Nerd Font, macOS developer environment.
**Topics:** `claude-code` · `dotfiles` · `vscode-settings` · `macos` · `nerd-font` · `developer-tools`
