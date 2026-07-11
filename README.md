<p align="center">
  <strong>English</strong> ·
  <a href="README.de.md">Deutsch</a> ·
  <a href="README.es.md">Español</a> ·
  <a href="README.fr.md">Français</a> ·
  <a href="README.ru.md">Русский</a>
</p>

<h1 align="center">OFFICINA</h1>

<p align="center">
  <em>Latin</em> <strong>officina</strong> — a workshop, a forge where tools are made.
</p>

<p align="center">
  <img alt="Platform: macOS" src="https://img.shields.io/badge/platform-macOS-000000?logo=apple&logoColor=white">
  <img alt="Core: Claude Code" src="https://img.shields.io/badge/core-Claude%20Code-da7756">
  <img alt="License: MIT" src="https://img.shields.io/badge/license-MIT-green">
  <a href="https://github.com/KirKruglov/OFFICINA/commits/main"><img alt="Last commit" src="https://img.shields.io/github/last-commit/KirKruglov/OFFICINA?logo=github&label=last%20commit"></a>
</p>

<p align="center">
  <strong>Method and tools for systematic work with an AI development assistant.</strong> Skills,
  subagents, CLI, and methodology — assembled and battle-tested on real projects.
</p>

---

## Why OFFICINA

OFFICINA is a workshop for working with AI tools in product development. At its core is the
**method**: how to structure work with an AI assistant so it stays predictable and reproducible.
Around it — skills, subagents, CLI, and configs, ready to reuse.

- **Method first.** Reusable rules and guides — writing style, CLI authoring, skill & subagent
  authoring, loop engineering — each readable on its own.
- **Portable skills, Claude Code at the core.** Skills aren't tied to one tool — they run in Claude
  Code and compatible environments. Subagents and settings are tuned for Claude Code. Both are
  hand-picked rather than dumped wholesale.
- **Environment around the core.** VS Code as an applied section that supports the core rather than
  being the headline.

## Who it's for

- **You work with an AI assistant and want a system** — a method, not one-off prompts.
- **You write your own skills, subagents, CLI** — take battle-tested patterns and guides.
- **You set up your working environment** for AI development on macOS.

If you're after a plug-and-play, install-and-forget distribution — this isn't it: OFFICINA is about
method and deliberate selection, not a ready-made turnkey build.

## What's inside

| Section | What it holds |
|---|---|
| [`skills/`](skills/) | Portable skills — reusable modes and procedures (Claude Code and compatible environments) |
| [`claude/`](claude/) | Claude Code layer — subagents and settings |
| [`methodology/`](methodology/) | Rules and guides — *how to work* |
| [`cli/`](cli/) | Personal CLI tools and shared libraries |
| [`vscode/`](vscode/) | VS Code settings, keybindings, curated extensions |
| [`install/`](install/) | Install and layout scripts |

## How the method works

The starting point is one question — what to solve the task with:

- A deterministic action, no model needed — **CLI**
- Need to read the context and decide by situation — **skill**
- An isolated subtask or a persona role — **subagent**
- A recurring cycle on a schedule — **loop**

Then comes the matching guide in [`methodology/`](methodology/): structure, conventions, a pre-deploy
checklist. The method is shared; the artifact carries over between projects.

## Quick start

Requires **macOS**. Clone the repo and open the section that fits your need — each ships its own
README and, where relevant, a one-command installer under [`install/`](install/):

```bash
git clone https://github.com/KirKruglov/OFFICINA.git
cd OFFICINA
```

The first ready installer sets up the VS Code environment:

```bash
./install/vscode.sh   # settings, keybindings, MesloLGS Nerd Font, curated extensions
```

## Philosophy

- **The right tool for the task.** Deterministic work goes into a script, context-driven decisions to
  the model. An extra model in the loop adds cost and uncertainty.
- **Reliability rests on constraints.** Boundaries, budgets, and checks hold the system together. An
  autonomous loop needs something able to say "no".
- **Knowledge lives in the artifact.** A guide, skill, or subagent gets reused; what's worked out once
  isn't worked out again every time.
- **Hand-picked.** What lands in the repo is proven in practice and deliberately chosen.

## Contributing

Issues, questions, and small fixes are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md). OFFICINA is
an open project.

Useful? Star it — that's how the next developer finds the project. Fork it, take what fits, and adapt
it to your own setup.

## License

Released under the [MIT License](LICENSE) — © 2026 Kir Kruglov. Free to use, modify, and distribute.
