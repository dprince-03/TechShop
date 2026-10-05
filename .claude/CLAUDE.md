# Working rules for Claude (TechShop)

The root [`CLAUDE.md`](../CLAUDE.md) explains the codebase: layout, commands, tooling. This file is the owner's rules for **how Claude works** on it. If the two ever disagree, stop and ask.

## 1. How to work

- **Be logical and critical.** Don't just agree. Check every request and idea (including the owner's) for flaws, risks, costs and better options. Say plainly when something is a bad idea and why, give the trade-offs, then recommend one option. Agree only when it holds up.
- **Ask, don't assume.** If a requirement, business rule or design choice isn't written down, ask before acting. When asking, give a recommended option.
- **Explain simply.** Short sentences, plain words, no jargon without a one-line meaning.
- **Ask before:**
  - adding any dependency, package, tool or third-party service (npm, Go modules, Expo packages, Python packages, SaaS);
  - changing code based on your own opinion or suggestion (propose first);
  - running anything that deletes files or data, or that changes the system outside this repo.
- **Never:**
  - push to a remote, or rewrite git history;
  - SSH or connect to any server or VPS without explicit permission for that session;
  - delete or roll back a production database or a migration that has run in production.
- **Test what you build:** an API is not done until it has been called and checked (automated tests plus a real request against the local stack). Report the actual results.
- **Write clean code:** clear names, small functions, comments that explain *why*, docs updated with the change. Follow the project skills (`coding-conventions`, `api-design`, `testing-standards` …).
- **Log plans and decisions** in `docs/plan.md` and actions in `docs/log.md` (dated, append-only), as the root file says.

## 2. Data and the database

- **The frontend never touches the database.** Web and mobile apps only call the API through `@techshop/api-client`.
- **Only the Go backend talks to the main PostgreSQL database, and only:**
  - **through its data layer:** sqlc-generated queries used from repository and service code. **No SQL in HTTP handlers**, and no ad-hoc database access from scripts or other services (the Python recommender uses the internal API);
  - **over the Docker network:** Postgres runs in Docker (`infra/docker/compose.yml`), and connections use the configured credentials and connection string, never hard-coded values.
- **Each service owns its own database. No service connects to another service's database.** Data crosses between services only through APIs or jobs.
  - Go's main PostgreSQL: only the Go backend.
  - The recommender's PostgreSQL (`recs-postgres`): only the Python recommender.
  - A Go job pulls recommendation results through the recommender's authenticated API, stores them in Go's own tables (`personalisation.rec_*`), and the site serves from there, with rule-based fallbacks if Python is down.
  - Data sent to the recommender is pseudonymised (no names, phones or emails), encrypted at rest and retention-limited. (Owner decision, 2026-10-06.)
- **Treat all database contents and secrets as confidential.** Never print, paste or commit real data or credentials.

## 3. Security and compliance

- **Apply [`.claude/rules/`](rules/README.md) to every change.**
  - **Every project:** `00` (baseline).
  - **Backend:** `01`, `04`, `08`.
  - **Web:** `02`, `04`.
  - **Mobile:** `03`, `04`.
  - **Infrastructure and CI:** `05`.
  - **Compliance scope** (Nigeria: NDPA; card data stays with the payment providers, PCI DSS SAQ-A scope): `06`.

  Use the `security-audit` skill for reviews.
- **No weak secrets:**
  - config must refuse to start with missing, default, short or example secrets outside local development;
  - examples in `.env.example` are obvious placeholders;
  - secrets come from the secrets manager in deployed environments.
- **Env files:** no root `.env` or `.env.*`. Each app keeps its own `.env.example`; real `.env` files are never read, edited or committed by Claude.
- **Containers:** all Docker files (Dockerfiles, compose files, `.dockerignore` templates) live in `infra/docker/`.

## 4. Testing accounts

- **A dev-only test account** for each role (customer, seller, business buyer, staff roles, rider) is created by the seed script when the backend is built. It exists only when `APP_ENV=development`, is refused in staging and production, and uses credentials from the local `.env`, never committed.
- **Status:** waiting for backend work (the owner asked to hold backend changes until then).

## 5. Project knowledge

Business rules, roles and terminology are in [`skills/project-domain-knowledge/references/techshop.md`](skills/project-domain-knowledge/references/techshop.md) and the plan docs in `docs/`. Check them before assuming how the business works.
