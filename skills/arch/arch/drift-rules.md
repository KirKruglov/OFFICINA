# Drift granularity

Rules for comparing recorded intent (`docs/architecture/tech-stack.md`, the `stack` block) against
project reality (manifests, lockfile), used by the `arch-review` and `arch-health` skills.

Current as of: 2026-07-19.

## Reference point

The comparison reference is the `version` field of a `stack` entry: an exact version (`5.1.2`) or a
range (`^5.1`, `>=5,<6`) in the notation of the entry's ecosystem.

## Comparison rule

A version in the manifest and lockfile that falls inside the recorded exact version or range is a
match. Bumps that Renovate carries out within the recorded decision (patches and minors inside the
range) do not count as drift.

## Direct vs transitive

Only direct dependencies are compared. The artifact contract does not record transitive ones, and
comparing them by name would produce noise with no recorded intent behind it; the integrity of the
transitive graph is covered by type D4.

## Drift types

The list is exhaustive:

| Code | Type | Intent | Reality | Severity |
|---|---|---|---|---|
| D1 | version outside the entry | `version` in `stack` | version in the manifest or lockfile outside the exact version/range | high |
| D2 | unrecorded dependency | no entry in `stack` | a direct dependency is present in the manifest | medium |
| D3 | dead entry | an entry exists in `stack` | no such dependency in the manifest | medium |
| D4 | manifest/lockfile desync | — | lockfile missing, or its versions contradict the manifest constraints | high |

The package comparison D1–D3 applies only to entries with `kind: package`. Entries with
`kind: runtime` or `kind: service` (language, runtime, DBMS) are checked against whatever reality is
available (version fields in the manifest, toolchain configs); when that is unavailable they are
marked "unverifiable" and stay out of drift.

## Informational layer (not drift)

- Entries with `verified` older than 90 days — stale verification (the default expiry).
- A recomputed lockfile hash that does not match `proof` while versions sit inside the recorded
  ranges — a stale `resolve` (the normal trace of Renovate bumps), not drift.

Both are printed as a separate report section with no severity.
