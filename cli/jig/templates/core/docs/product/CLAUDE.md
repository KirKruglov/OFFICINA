# docs/product

Product documents — everything that is not the codebase.

## Purpose
Product documents; code lives at the repository root, not here.

- description/         — product descriptions: what it is, who it's for, what it consists of
- architecture/        — architecture: ARCHITECTURE.md + the ADR log in adr/
- strategy/            — strategy and vision (in full projects)
- release-<version>/   — release materials (in full projects):
    prd-release/ — PRD, spec/ — specifications, use-case/ — scenarios, mockups/ — mockups

## Local rules
- One document — one file; document names in kebab-case.
- Every significant architectural decision — a separate ADR from the template.
- A new release — a new release-<version>/ folder next to the previous one.

## How to keep it
Architecture: ARCHITECTURE.md is a living document, an ADR for each decision.
The ADR template is architecture/adr/adr-template.md.
