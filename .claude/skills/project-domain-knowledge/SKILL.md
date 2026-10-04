---
name: project-domain-knowledge
description: Business rules, terminology, user roles, workflows, tenants/organizations, integrations, and decisions for the user's specific projects and clients. Use this skill whenever working on features, data models, copy, reports, or bug fixes for a named project, or when the user refers to project-specific terms, organizations, roles, or workflows — check here before assuming how the business works. Projects are documented one file each in references/.
---

# Project Domain Knowledge

Code shows *how* a system works; this skill records *why* and *what the business expects*, which Claude can't infer from code.

## How to use

1. Identify which project the task belongs to.
2. Read that project's file in `references/` (one file per project, named `<project-slug>.md`).
3. Apply its rules, glossary, and constraints. If the task contradicts a documented rule, flag it rather than silently changing behaviour.
4. If you learn a new durable rule during the task (the user states it), suggest adding it to the project file.

## Project files

- `references/_TEMPLATE.md` — copy this to start a new project file
- `references/techshop.md` — TechShop: Nigerian tech retailer + multi-vendor marketplace + supplier (B2C and B2B), five web apps, two mobile apps, Go API

Add a line here for each project file you create, e.g.:
- `references/trp-platform-suite.md` — multi-tenant platform for three organizations
- `references/swan.md` — songwriter royalty/split app
- `references/bibliotheca.md` — authors, libraries, and readers platform

## Rules

- Never invent business rules. If a rule isn't documented and matters for correctness (pricing, permissions, compliance), ask.
- Keep client-confidential details out of public repos; if this skill lives in a public repo, store only non-sensitive rules.
