---
name: technical-docs
description: Write and maintain technical documentation — READMEs, architecture overviews, Architecture Decision Records (ADRs), onboarding guides, environment setup docs, API usage guides, operational runbooks, incident postmortems, and inline code documentation. Use this skill whenever the user asks to document a project, feature, system, decision, or process, write or update a README, record an architecture decision, create onboarding docs, or write a postmortem.
---

# Technical Documentation

## Principles

- **Audience first**: new developer? ops on-call? client's technical team? Write for that reader.
- **Task-oriented**: lead with how to do the thing; put background after.
- **Runnable**: commands copy-pasteable and tested; specify versions.
- **Close to the code**: docs live in the repo (`README.md`, `docs/`), updated in the same PR as the change.
- **No secrets** in docs — reference where secrets live (vault, env group), never the values.
- **Diagrams as code** when possible (Mermaid) so they stay editable.

## Doc types & templates (in `assets/`)

| Need | Template | Location |
|---|---|---|
| Project entry point | `README.template.md` | `/README.md` |
| Significant technical decision | `ADR.template.md` | `docs/adr/NNNN-title.md` |
| System overview | `ARCHITECTURE.template.md` | `docs/architecture.md` |
| Operational procedure | `RUNBOOK.template.md` | `docs/runbooks/<topic>.md` |
| Incident review | `POSTMORTEM.template.md` | `docs/postmortems/YYYY-MM-DD-<title>.md` |
| New team member | `ONBOARDING.template.md` | `docs/onboarding.md` |

## When to write an ADR

Any decision that is costly to reverse or that someone will later ask "why did we do this?": choice of database, auth approach, multi-tenancy model, hosting provider, framework, major library, API style. ADRs are immutable once accepted; supersede with a new ADR instead of editing.

## Writing style

- Short sentences, active voice, present tense.
- Headings that answer questions ("How to run migrations").
- Code blocks with language tags; expected output where helpful.
- Explain *why* for non-obvious steps.
- Date and owner on docs that can go stale.

## Code-level docs

- Go: doc comments on exported identifiers, starting with the name.
- TS/JS: JSDoc/TSDoc on exported functions and complex types.
- Java: Javadoc on public APIs.
- Comments explain intent and constraints, not what the code literally does.
