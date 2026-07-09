<h1 align="center">OFFICINA</h1>

<p align="center">
  <em>Latin</em> <strong>officina</strong> — a workshop, a forge where tools are made.
</p>

<p align="center">
  <img alt="Platform: macOS" src="https://img.shields.io/badge/platform-macOS-000000?logo=apple&logoColor=white">
  <img alt="Focus: Claude Code" src="https://img.shields.io/badge/focus-Claude%20Code-da7756">
  <img alt="Status: work in progress" src="https://img.shields.io/badge/status-work%20in%20progress-f5a623">
  <img alt="License: MIT" src="https://img.shields.io/badge/license-MIT-green">
</p>

<p align="center">
  <strong>A developer's workshop for building with Claude Code</strong> — skills, subagents, CLI tools,
  and a working <strong>methodology</strong>, plus curated VS Code and shell dotfiles for macOS.
</p>

---

> [!NOTE]
> **Work in progress.** The structure is laid out; sections are filled by deliberate curation —
> not bulk export — and grow as new tools, skills, and rules are added.

## Why OFFICINA

Most public **dotfiles** and **setup** repositories are configuration dumps — a snapshot of *which
toggles are flipped*. OFFICINA leads with **methodology**: *how to actually work* with Claude Code and
your tooling. Configuration lives here too, but it orbits the method — not the other way around.

- **Method first.** Reusable rules and guides — writing style, CLI authoring, skill & subagent
  authoring, loop engineering — kept readable on their own.
- **Claude Code at the core.** Curated skills, subagents, and settings, selected on purpose rather
  than mirrored wholesale.
- **Environment around it.** VS Code and shell as applied sections that support the core, not the
  headline.
- **macOS, one command.** Portable setup: bring a fresh machine up to the same environment with a
  single install script.

## What's inside

| Section | What it holds | Status |
|---|---|---|
| [`claude/`](claude/) | Core: skills, subagents, Claude Code settings | 🚧 Growing |
| [`methodology/`](methodology/) | Rules and guides — *how to work* (the main differentiator) | 🚧 In progress |
| [`cli/`](cli/) | Personal CLI tools and shared libraries | 🚧 In progress |
| [`vscode/`](vscode/) | VS Code settings, keybindings, curated extensions | ✅ Ready |
| [`shell/`](shell/) | zsh, aliases, git config, fonts | 🚧 In progress |
| [`install/`](install/) | Install and layout scripts | 🚧 `vscode.sh` ready |

## Quick start

Requires **macOS**. Clone the repo and explore the section that fits your need — each ships its own
README and, where relevant, a one-command installer under [`install/`](install/):

```bash
git clone https://github.com/KirKruglov/OFFICINA.git
cd OFFICINA
```

The first ready installer sets up the VS Code environment:

```bash
./install/vscode.sh   # settings, keybindings, MesloLGS Nerd Font, curated extensions
```

More installers and sections land as the repository grows — see the [section map](#whats-inside)
for what's available now.

## Contributing

Issues, questions, and small fixes are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md). OFFICINA is a
curated showcase exported from a private source, so larger changes are best raised as an issue first.

## License

Released under the [MIT License](LICENSE) — © 2026 Kir Kruglov. Free to use, modify, and distribute.

---

<sub>
<strong>Keywords:</strong> Claude Code, Claude Code skills, subagents, AI coding workflow, developer
methodology, dotfiles, VS Code settings, zsh config, macOS developer setup, CLI tools.<br>
<strong>Topics:</strong> <code>claude-code</code> · <code>claude-code-skills</code> ·
<code>subagents</code> · <code>dotfiles</code> · <code>vscode</code> · <code>macos</code> ·
<code>developer-tools</code> · <code>methodology</code>
</sub>
