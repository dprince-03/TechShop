---
name: coding-conventions
description: House coding standards for Go (Gin), Node.js (Express, TypeScript/JavaScript), React/Next.js and Java projects — folder structure, naming, error handling, logging, config, and code style. Use this skill whenever writing, editing, refactoring or reviewing application code in any of these stacks, even for small changes, so new code matches the project's conventions instead of generic patterns. Also use when the user asks "how should I structure this", "clean this up", or "make this consistent".
---

# Coding Conventions

## Golden rule

**The existing codebase wins.** Before writing code, inspect 2–3 neighbouring files and match their structure, naming, error style, and libraries. The defaults in this skill apply only where the project has no established pattern. If the codebase conflicts with a security rule, follow the security rule and mention the conflict.

## Workflow

1. Identify the stack (look at `go.mod`, `package.json`, `pom.xml`, `next.config.*`).
2. Read the matching reference file:
   - Go / Gin → `references/go-gin.md`
   - Node.js / Express → `references/node-express.md`
   - React / Next.js → `references/react-nextjs.md`
   - Java / Spring → `references/java-spring.md`
3. Scan nearby files for local conventions.
4. Write the code, then self-check against the universal rules below.

## Universal rules (all stacks)

**Structure**
- Layered: transport (handlers/controllers/routes) → service (business logic) → repository (data access). Handlers never talk to the DB directly; repositories never contain business rules.
- One responsibility per file; keep files under ~300 lines where practical.
- Group by feature/module (`users/`, `billing/`) rather than by technical type for anything beyond a tiny app.

**Naming**
- Descriptive over short: `activeSubscriptionCount`, not `cnt`.
- Booleans read as questions: `isActive`, `hasAccess`, `canEdit`.
- Functions are verbs: `createInvoice`, `findUserByEmail`.
- Avoid abbreviations except universally known ones (`id`, `url`, `db`, `ctx`).

**Errors**
- Never swallow errors. Handle, wrap with context, or return.
- Distinguish expected domain errors (not found, validation, conflict, forbidden) from unexpected ones. Map domain errors to HTTP status at the transport layer only.
- Never leak internals (stack traces, SQL, file paths) to API responses.

**Config & secrets**
- All config from environment variables, validated at startup; fail fast if required values are missing.
- No secrets, URLs, or credentials hardcoded. Provide `.env.example` with dummy values.

**Logging**
- Structured JSON logs in production with `level`, `msg`, `request_id`, `user_id` (and `tenant_id` in multi-tenant apps) where available.
- Never log passwords, tokens, full card numbers, or raw PII.

**Multi-tenancy (when the app is multi-tenant)**
- Tenant ID is derived from the authenticated identity on the server, passed explicitly through context, and applied to every query. Never trust a tenant ID from the request body.

**Ownership scoping (single-company apps such as TechShop)**
- TechShop is **not** multi-tenant (one company). Scope data by **ownership** instead: `customer_user_id` for customers, seller membership for Seller Centre, business membership for B2B, staff permissions plus data scopes (store, warehouse, zone) for staff. Put the scope in the SQL (`WHERE … AND customer_user_id = $2`) — see `docs/identity-access.md` §7 and `project-domain-knowledge/references/techshop.md`. Never trust an owner ID from the request body.

**Money & time**
- Money as integer minor units (kobo, cents) or decimal types — never floats.
- Store timestamps in UTC; convert at the edges.

**Comments**
- Explain *why*, not *what*. Delete commented-out code.
- Exported/public functions get a short doc comment.

**Dependencies**
- Prefer the standard library and well-maintained packages. Justify new dependencies; don't add a library for a 10-line function.

## Self-check before finishing

- [ ] Matches neighbouring file style
- [ ] Layers respected (no DB calls in handlers)
- [ ] Inputs validated at the boundary
- [ ] Errors wrapped/handled, no internals leaked
- [ ] No hardcoded secrets or config
- [ ] Data scoping applied (tenant or ownership)
- [ ] Tests added or updated (see `testing-standards` skill if available)
