# Loop engineering rules

Rules and requirements for Claude when building a loop for a project or process.

---

## 0. What we are building

- A loop is a system that discovers, executes, verifies, persists, and reschedules work with no human
  in the inner cycle.
- The unit of design is a loop that launches itself, not a single run.
- Memory between conversations is what makes this a loop rather than a one-off task run many times.
- The real difficulty is not building the loop, but placing inside it something able to say "no" and
  stop it.
- The cost of a mistake scales with the number of turns it survives before detection; a loop by
  definition maximizes the number of turns. Every safety mechanism exists to shorten the distance
  between a mistake and its detection.
- A loop's reliability comes from the quality of its constraints, not the size of the model.

---

## 1. The five mandatory steps of one turn

Remove any of the five and the loop either will not turn or will spin in place. Install all five.

| Step | Requirement |
|---|---|
| Discovery | The loop finds its own work for the turn rather than receiving a ready list. The discovery logic lives in a skill, not in a wall of instructions in cron. Discovery sets the quality ceiling of the whole loop. |
| Handoff | Every finding worth acting on gets its own isolated git worktree. The cleaner the task is carved out, the easier the verification and merge. |
| Verification | After generation, a second subagent checks the result — with different instructions, sometimes a different model. This is "the thing that can say no". A loop with no real check is an agent nodding at itself. |
| Persistence | The result lands where it outlives the conversation: a PR plus an updated ticket, an inbox for the unresolved, a state file with progress. A loop's memory cannot live only in the context window. |
| Scheduling | A real trigger (timer/event) turns one turn into a loop; a state file carries the unfinished work into the next run. |

---

## 2. The six mandatory parts

The parts implement the steps. Mapping: discovery → skills, handoff → worktrees, verification →
subagents, persistence → memory, scheduling → automations, plus connectors.

| Part | Requirement |
|---|---|
| Automations | Launch the loop on a schedule/trigger. An automation calls a named skill, not a wall of instructions in cron. Kinds: local (machine on) and cloud (runs with the machine off). |
| Worktrees | The built-in git mechanism for several independent working directories. One isolated worktree per task — mandatory under parallelism. |
| Skills | Make project knowledge permanent in a SKILL.md file so context is not re-derived every turn. They repay the intent debt. A skill is reusable and maintainable; a wall of prompts is not. |
| Connectors | Over MCP. Connect the loop to the outside world (tracker, DB, staging API, Slack). They define the loop's field of view; a loop that sees only the filesystem is a tiny loop. |
| Subagents | Separate the writer from the judge. When one agent is both player and judge, the judge favors its own side. |
| Memory | Persistent state outside the conversation (a markdown file or a board), written to disk. Memory is not context: context resets on compaction, memory survives between rounds and days. The agent forgets, the repository does not. Auto memory saves the agent's notes on its own; the first ~200 lines of `MEMORY.md` are loaded into every session. Path-scoped rules (`.claude/rules/*.md` with a `paths:` field) load only for matching files — they save discovery context. |

The question when choosing tools: are all six capabilities present, not which brand of team
provides them.

### Deterministic gates vs LLM steps

- Anything deterministic logic can decide, never hand to a probabilistic model; where that line runs
  is where the loop's reliability is decided.
- Gather context with a deterministic orchestrator *before* waking the model (link lookup, tracker,
  docs, code search over MCP) — an LLM foraging for its own context is the least controllable.
- Alternate deterministic gates and creative LLM steps: the agent writes code → a hard-coded linter
  the agent cannot skip → the agent fixes → a hard-coded commit.
- Run environments as "cattle, not pets": each is disposable, so many agents work at once without
  getting in each other's way.

---

## 3. Generator and evaluator

- Do not try to make the generator self-critical — it works poorly. An agent judging its own output
  tends to praise it, because it sees the chain of self-persuasion rather than the result.
- Set up a separate evaluator as a skeptic: a different agent with entirely different instructions,
  looking at the code from scratch. This is a structural decision, not a matter of wording.
- The evaluator must act, not just read: the basis for judgment is "I clicked the button, the page
  navigated, here is a screenshot", not "the JSX looks fine". On the frontend — wire up a browser MCP
  (for example the third-party Playwright MCP — Anthropic does not ship it officially): open the page,
  click, screenshot, inspect the DOM.
- Change the evaluator's base model: the same model with new instructions often keeps the same blind
  spots.
- The evaluator's default stance is doubt, not trust: treat the code as broken until proven
  otherwise.
- The final completion decision is made by a fresh model, not the one that did the work. After each
  turn a small fast model checks the condition; if unmet — another turn runs. This is the
  maker–checker principle.
- The loop's floor is its evaluator: the generator's level sets what the loop can produce; the
  evaluator's level sets what it will not produce.

Evaluator agent template:

```
# Evaluator agent (.claude/agents/reviewer.md)
ROLE: Adversarial code reviewer.
ASSUMPTION: this code is BROKEN until proven otherwise.
DO NOT PRAISE. Look for what does not work.
CHECK in order:
1. Does it run? (run it, do not read it)
2. Tests: run them, paste the real output.
3. Edge cases the author missed.
4. Does the behavior match the ticket?
USE a browser MCP (Playwright): open the page, click,
screenshot, inspect the DOM. Judge behavior, not intent.
VERDICT: PASS only if all checks pass.
Otherwise REJECT + a list of each reason.

# Stop condition, evaluated by a fresh small model
/goal all tests in test/auth pass and the lint step is clean
```

Four steps to grow a loop's ability to say "no": separate generation from judgment structurally, set
up the evaluator as a skeptic, make it verify through action, hand the final decision to a fresh
model.

---

## 4. Five anti-patterns (forbidden)

Each failure = one missed or poorly executed step.

| Anti-pattern | Missed step | Symptom | Cure |
|---|---|---|---|
| Nodding loop | Verification | The loop never once said "no" to itself over hundreds of turns | Separate generator and evaluator |
| Amnesiac loop | Persistence | No cumulative progress; every morning starts from the same place | A state file on disk |
| Manual loop | Scheduling | The last run was on demo day | A real trigger: timer or event, not dependent on a human's memory |
| Blind loop | Discovery | A human still spends the morning deciding what the loop should do | Move discovery into a skill |
| Tangled loop | Handoff | Under parallelism the edits collide, the merge is a mess | One isolated worktree per task |

A disciplined loop installs all five steps; a rushed one installs only discovery and handoff (the
visible output) and skips the three that provide safety.

---

## 5. Choosing a scheduler

The rule is mechanical: is the loop's work tied to the local machine or can it leave it.

| | Cloud | Desktop | /loop |
|---|---|---|---|
| Where it runs | cloud | machine | machine |
| Machine on? | no | yes | yes |
| Session open? | no | no | yes |
| Min. interval | 1 h | 1 min | 1 min |
| Sees local files? | no | yes | yes |

- Frequent runs + access to local files → local `/loop`, at the cost of keeping the machine on.
- Work not tied to local state (e.g. a nightly task scan) → cloud (Cloud Routines or a GitHub Actions
  trigger), at the cost of an hour-plus interval and a fresh clone each time.
- Do not treat a local restart as embodying "work while you sleep": local = "several rounds while I am
  here", cloud = "runs even when I am away".
- A mature loop often uses both: local for tight inner checks, cloud for the nightly scan.
- Self-paced `/loop`: with no interval given, the model picks the delay between turns itself (in the
  range 1 min – 1 h).
- The cloud trigger via GitHub Actions is officially GA (`claude-code-action@v1`); install with
  `/install-github-app`. A separate class is Managed Agents (server-hosted agents with scheduled
  deployment via API); that is not Claude Code Desktop.

---

## 6. Operational discipline

- **Always read a sample.** Not everything, but a representative sample each day, forcing yourself to
  explain each chosen change: what it did and why. An inability to explain is a signal the map has
  fallen behind. The sample must be regular and genuinely examined.
- **Set a ceiling before shipping.** Hard ceilings before the first unattended run, not after the
  first bill: a per-run budget, a daily budget, a max retry count. These are circuit breakers that
  turn open-ended risk into bounded risk.
- **Keep one door open.** Build in at least one checkpoint where the loop stops and waits for a human.
  The mere existence of the pause keeps the human in a position to intervene. The checkpoint is a
  permanent feature, not temporary scaffolding.

---

## 7. Requirement for the discovery skill

The discovery step depends on a skill, not a wall of instructions. The SKILL.md structure — five
sections for the five steps plus a mandatory Stop section. Example layout:

```markdown
# .claude/skills/<name>/SKILL.md
---
name: <name>
trigger: invoked by a daily automation
---

## Read (DISCOVERY inputs)
- CI runs failed since the last run
- tasks opened in the last 24 hours
- commits merged since yesterday
- the previous ./state/triage.md

## Judge (the part that sets the ceiling)
For each candidate decide:
- actionable right now, or noise?
- does it block the release? → priority
- already tracked? → skip
Keep only what is worth a worktree today.

## Write (PERSISTENCE output)
Append to ./state/triage.md:
| finding | source | priority | status |
Commit the file so it can be read tomorrow.

## Hand off (HANDOFF preparation)
For each saved finding emit a task line:
worktree=fix/<slug> goal=<stop-condition>

## Stop (the boundary you keep for yourself)
Never merge. Never delete. Anything
you are unsure of goes to ./inbox/
for a human, not into a PR.
```

The Stop section is mandatory: the loop will do everything the skill says and nothing it does not.
This is the one place where the engineer's intent about where to keep control becomes permanent.
Without it, the loop will merge with a confidence it has not earned.

---

## 8. The minimal complete loop (all six elements)

```yaml
# 1. SCHEDULING -- a real trigger
# (.github/workflows/triage.yml)
on:
  schedule:
    - cron: '0 6 * * *'  # 06:00 daily, cloud

# 2. DISCOVERY -- a skill, not a wall of text
# the skill is invoked by a slash command in the headless-mode (-p) prompt,
# not by a separate CLI flag:
run: claude -p "/morning-triage"

# 3. PERSISTENCE -- state on disk
# the skill writes ./state/triage.md
# and commits it back to the repository

# 4. HANDOFF -- one worktree per finding
# --worktree (-w) -- a boolean flag: an isolated copy of the repo;
# the stop condition is set with the /goal command inside the prompt, not a flag
for finding in $(parse ./state/triage.md); do
  claude -w -p "draft a fix for $finding; /goal tests pass and lint is clean"
done

# 5. VERIFICATION -- a fresh model judges
# the /goal stop-check runs after each turn;
# a second reviewer agent hunts for flaws

# 6. HUMAN CHECK -- an open door
# PRs are opened but never merged automatically;
# anything uncertain lands in ./inbox/
```

A loop with all six elements is a real loop. A loop missing even one is one of the five anti-patterns
in disguise.

---

## Closing stance

Design a system that prompts the agent instead of the human — but as a human who intends to stay an
engineer, not merely the one who presses "go". The evaluator, the state file, the budget ceiling, the
open door — each keeps the human in a position to say "no" to a machine built to say "yes" at speed.
