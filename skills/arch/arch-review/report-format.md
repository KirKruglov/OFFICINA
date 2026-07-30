# arch-review report format

A format reference for the output of the `arch-review` skill. What fills the sections varies with
the input; the fill-in logic lives in `SKILL.md`.

## Sections

| # | Section | Present |
|---|---------|---------|
| 1 | Summary | always |
| 2 | Drift table | when drift was found |
| 3 | Structural rules | when rules are declared |
| 4 | Stale verifications | always |
| 5 | Run coverage | always |
| 6 | Route to a fix | when drift was found |

## 1. Summary

- The verdict in one line: "no drift found", or the number of discrepancies by severity
  (high / medium).
- Marks of a degraded run: "no structural rules declared", a skipped tier.

## 2. Drift table

Columns: code / object / intent / reality / severity.

Code dictionary (semantics and comparison rules — `drift-rules.md`, see "References" in
`SKILL.md`):

| Code | Type | Severity |
|------|------|----------|
| D1 | version outside the entry | high |
| D2 | unrecorded dependency | medium |
| D3 | dead entry | medium |
| D4 | manifest/lockfile desync | high |

## 3. Structural rules

Columns: id / status / evidence.

- Statuses: "satisfied", "violated", "unverifiable".
- Evidence: for "violated" — the output of the `check` command; for "unverifiable" — the reason
  (an empty `check`, a prefix outside the allowlist, or a shell metacharacter in the command).
- When the `rules` block is empty or absent, the section is replaced by the note "no structural
  rules declared".

## 4. Stale verifications (informational layer)

Not drift, no severity. Contents:

- `stack` entries whose `verified` is older than 90 days;
- a recomputed lockfile hash that does not match `proof` while versions sit inside the recorded
  ranges.

An empty layer prints the line "no stale verifications".

## 5. Run coverage

- The date of the run.
- The manifests and lockfiles covered, by ecosystem.
- `stack` entries marked "unverifiable" (`kind: runtime` / `service` with no available reality).

## 6. Route to a fix

One line: `/arch-change` or an explicit user decision.

## Degenerate report (the architecture is not recorded)

When `docs/architecture/` is missing or its yaml blocks are broken, the report consists of:

- the verdict `🛑` "architecture not recorded / block broken";
- the list of what is missing (files, blocks);
- the recommendation to run `/arch-new`.

Sections 2–6 are absent.
