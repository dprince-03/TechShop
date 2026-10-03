# TechShop — Work Log

Append-only, chronological record of work done, sources consulted, and verification results. Decisions and their rationale live in [`plan.md`](plan.md). **Never delete or rewrite past entries.**

Format: `### YYYY-MM-DD` heading, then one bullet per action.

---

### 2026-10-01 *(backfilled 2026-10-02)*

- Scaffolded `frontend/` with `create-next-app` (TS, App Router, `src/`, ESLint, no Tailwind).
- Scaffolded `backend/` (Gin, pgx, `caarlos0/env`, godotenv, gin-contrib/cors); `/health` endpoint; slog request logging; graceful shutdown.
- Added `docker-compose.yml` (Postgres 17), root `Makefile`, `.editorconfig`, READMEs.
- Added Drizzle to frontend → removed it from frontend → moved to `database/` → removed it entirely in favor of goose + sqlc in `backend/`. Verified migrate-up/status/create, auto-migrate on startup, and sqlc codegen with a temporary table/query.
- Removed root `.env`/`.env.example` and empty `pipeline/`.
- Added `.github/workflows/ci.yml`; ran every CI step locally — all pass. golangci-lint caught an unchecked `Close()` error in `internal/database/migrate.go`; fixed.
- Created `CLAUDE.md`; later trimmed to non-derivable rules.

### 2026-10-02

- Created `docs/plan.md` (decisions) and `docs/log.md` (this file); backfilled 2026-10-01 entries.
- **Design research — sources:**
  - https://www.apple.com/store — HTML + linked/inline CSS downloaded and measured (fonts, sizes, weights, tracking, colors, radius, breakpoints, shadows, transitions); page structure via fetch.
  - https://www.samsung.com/us/mobile/ — HTML + CSS measured; page structure via fetch.
  - https://us-store.msi.com/ — HTML + CSS measured (Bootstrap-based).
  - https://www.amazon.com/ — direct download blocked (HTTP 202, empty); markup and CSS variables partially analysed via fetch.
  - https://www.jumia.com.ng/ — direct download blocked (HTTP 403); page structure analysed via fetch.
  - https://www.gucci.com/us/en/ — blocked for all automated access (HTTP 403); notes based on known public design conventions, not measured.
- Computed WCAG contrast ratios for the candidate palette; all text pairings ≥ 4.5:1, control borders ≥ 3:1.
- Added `frontend/src/styles/{tokens,reset,base,layout,components,utilities}.css`; rewrote `frontend/src/app/globals.css` as the layered entry point.
- `npm run build` succeeded; confirmed compiled CSS keeps `@layer` order (~17 KB).
- Rendered a scratch preview (outside the repo) with headless Chrome at 390px/1440px, light/dark. Fixed: badge stretching in flex-column cards; white-on-light-red badge contrast in dark mode (new `-fill` tokens). Re-rendered — fixed.
- Created skill `.claude/skills/frontend-design/SKILL.md` with all research findings, synthesis table, token reference, and usage rules.
- Added the logging rule to `CLAUDE.md`.

- Recorded business context from the owner (model, payments, catalogue, planned sites and apps) in `plan.md`; listed architecture implications as pending proposals.
- Logged the multi-app monorepo restructure decisions in `plan.md`. Checked versions: Expo 57 pins its own React (differs from Next's 19.2), which is why web and mobile use separate install roots.
- Restructure **not started**: the first move command was denied by the permission prompt and the usage limit was reached. No files were moved; the repo is unchanged since the design-system entry.
- Restructure implemented:
  - Created `shared/design-tokens` (TS tokens + CSS generator) and `shared/api-client`. Generated `tokens.css` has exactly the same variable set as the previous hand-written file (diffed).
  - Created `frontend/` workspace (npm workspaces + Turborepo): `packages/ui` (moved CSS), apps `corporate` (3000), `market` (3001), `wholesale` (3002), `seller` (3003), `staff` (3004).
  - Created `mobile/` workspace: Expo 57 apps `customer` and `logistics` from the default template, trimmed to an empty shell; theme hooks built from shared tokens; ESLint configured (`eslint-config-expo`).
  - Moved `docker-compose.yml` → `infra/docker/compose.yml`; added nginx service and `infra/nginx/` routing config for `*.techshop.localhost`.
  - Updated backend CORS origins (`internal/config`, `.env.example`), root `Makefile`, CI workflow, READMEs, `CLAUDE.md`, and the `frontend-design` skill.
- Verification:
  - Web: typecheck, lint, Prettier, and production build pass for all 5 apps; each compiled CSS contains the generated tokens with correct layer order.
  - Mobile: typecheck and lint pass; `expo-doctor` 21/21 for both apps (after fixing duplicate React via `overrides` + clean reinstall); Android JS bundles exported for both apps, confirming Metro resolves `shared/`.
  - Infra: `docker compose config` valid; `nginx -t` passes.
  - End to end: started Postgres + nginx, API, market and staff dev servers; via nginx `api.techshop.localhost/health` → `{"status":"ok","database":"up"}`, market/staff served their pages (staff with `noindex`), unknown host → 404. Services stopped afterwards.
  - `make check` (everything CI runs) passes.
- Not done (blocked by permission rules): deleting the temporary `_web_old/` folder; updating the developer's local `backend/.env` CORS line.
- Homepages for all five web apps:
  - Created `.claude/rules/frontend-design.md` (UI/UX laws, heuristics, WCAG 2.2 AA, e-commerce, forms, motion, content, performance, definition of done), scoped to `frontend/**`, `mobile/**`, `shared/design-tokens/**`.
  - Added domain types and money helpers to `shared/api-client`; new `frontend/packages/fixtures` (sample data, fictional brands/sellers).
  - Added shared components to `@techshop/ui`: Logo, Icon, CategoryArt, Price, Rating, ProductCard, SectionHeader, SiteHeader, SiteFooter/PaymentMethods, Countdown, SampleDataNotice; `@techshop/ui/sites` for cross-app URLs. Added the `--color-on-fill` token.
  - Built homepages: market (search-first header, hero, trust strip, categories, deals with countdown, 7 category shelves, cars listings, best sellers, sell/business band), wholesale (trade-price hero, retail vs wholesale, benefits, tiered-price cards, categories, steps, audiences, quote CTA), corporate (mission hero, four businesses, categories, principles, careers, newsroom empty state, contact), seller (hero with inclusions, benefits, who can sell, steps, requirements, FAQ, CTA), staff (app shell with 20 grouped modules, top bar, KPIs, recent orders, needs-attention queue, quick actions).
  - Read the bundled Next 16 upgrade guide: applied `data-scroll-behavior="smooth"` and `connection()`.
- Verification:
  - Lint, typecheck and build pass for all apps and the ui package; `make check` passes.
  - Screenshots via headless Chrome at 1440px and 390px for all five apps, plus dark mode (market, staff). Fixed: market header not staying sticky (moved the sticky wrapper), staff KPI value wrapping on mobile, staff orders table clipping at 1440px, wholesale section misaligned (container width), a `!important` and raw hex values (replaced with selector specificity and a token).
- Moved corporate to port 3005 (3000 is used by AdGuard). Verified `npm run dev` serves on 3005, then stopped it. Documented how to change ports in `infra/README.md`.
- Port check after work: no servers or containers from this session left running. Port 3001 (`npm run dev` from the owner's VS Code terminal) and port 3000 (AdGuard) were left alone.
- Market shopping journey built: `/c/[category]` (cars render as listings), `/c/[category]/[sub]`, `/search`, `/deals`, `/best-sellers`, `/p/[slug]` (27 static pages), `/cars/[slug]` (4 static pages), `/cart`, `/saved`, market 404. New shared components: Breadcrumbs, EmptyState, QuantityStepper. New fixtures helpers: queryProducts, getProduct/getCar, relatedProducts, productSpecs, deliveryEstimate, sampleCart. Read the Next 16 docs for `PageProps`, `dynamicParams` and `generateStaticParams`.
- Verification:
  - Status codes: all pages 200; unknown product, category and car subcategory → 404.
  - Screenshots at 1440 and 390. Fixed: excess top spacing on inner pages (`.section--page`), stretched quantity stepper, vertical subcategory list on mobile, then page-wide horizontal overflow it caused (aside `min-width: 0`).
  - Overflow measured through the Chrome DevTools Protocol: no horizontal scroll on 13 pages at 390px and 10 at 360px.
  - Scripted interaction tests (8/8 pass): mobile filter toggle with `aria-expanded`, sort auto-submit and ordering, add-to-cart announcement, cart quantity and total, cart remove and announcement, viewing-form validation (focus to first error, invalid phone rejected, `+234` accepted, honest preview confirmation). Found and fixed: stepper `aria-label` lowercased product names.
  - `make check` passes.
- Port check: preview server on 3101 stopped. No containers running.
- Remaining pages built (backend untouched, as the owner instructed mid-task):
  - Shared templates in `@techshop/ui`: PageHeader, Prose, FaqList, LegalDocument (+ outlines), PreviewForm, NavDisclosure. `normaliseNigerianPhone` moved to `shared/api-client`.
  - Market: account, orders (tracking), delivery, help, help/[delivery|returns|warranty|payments|contact], trade-in, legal/[terms|privacy|cookies].
  - Wholesale: retail, wholesale (trade prices), categories, c/[category], p/[slug] with live price-break QuoteBox, quote, business/register, business/sign-in, business/credit, business/invoices, contact, help, legal, 404.
  - Corporate: about, partners, careers (empty state + talent pool), news (empty state), contact, legal, 404.
  - Seller: register (4-section KYC form), sign-in, fees, policies, policies/prohibited, help, help/contact, legal (incl. seller agreement), 404.
  - Staff: ModulePage + DataTable templates with sample config for all 19 modules, orders + orders/[id], catalogue + catalogue/new, notifications, account, search, 404.
- Verification:
  - Typecheck, lint and build pass for all apps; `make check` passes.
  - **Link crawl:** 239 pages across 5 apps, every internal and cross-app link followed, **0 broken links**.
  - **Overflow:** 167 distinct pages measured at 390px through the Chrome DevTools Protocol, **0 with horizontal scroll**.
  - **Order totals:** all 6 sample orders verified on the rendered pages (line sums = totals) after fixing inconsistent data.
  - **Interaction tests:** PreviewForm (16 required-field errors flagged, focus to first error, every error linked via `aria-describedby`, password mismatch, invalid and valid Nigerian phone, valid submit shows "nothing was sent"); QuoteBox price breaks (retail at 1, ₦632,000 trade at 10, total ₦6,320,000); mobile menu closes on navigation and on Escape with focus returned; sidebar `aria-current`. One initial phone-check failure was a test artifact (headless tabs don't fire blur on programmatic focus); re-run with real focusout events: pass.
  - Screenshots of each template at 1440/390 and in dark mode. Fixed: order detail data mismatch, invisible borderless cards on white.
- Port check: test servers on 3101–3105 and test browsers on 9333–9336 stopped; no containers running. Confirmed 0 backend files modified during the task.
- Database plan (documentation only, backend untouched):
  - Written: `docs/database.md` (overview: principles, domain map diagram, domains and tables, relationships, indexing, privacy, migration phasing, traceability, open decisions) and `docs/database/schema.sql` (full proposed DDL, ~95 tables, `[domain]` comments, constraints, ledger balance trigger, append-only triggers, auto FK indexes, seed data).
  - **Not yet validated:** loading `schema.sql` into a throwaway Postgres failed to start because Docker Desktop returned an HTTP 500 API error. Nothing was loaded.
  - **Still to create:** `docs/database/erd.md` (generated ER diagrams), `states.md`, `flows.md`, `access.md`, `tools/generate.sh` + `generate.mjs`, rendered SVGs in `diagrams/`. Stopped here because the usage limit was reached.
- Database plan completed (2026-10-03), documentation only; `backend/` untouched:
  - Files:
    - `docs/database.md` (overview).
    - `docs/database/schema.sql` (96 tables, 185 FKs, 229 CHECKs, ledger balance + append-only triggers, auto FK indexes, seeds).
    - `erd.md` (generated: 15 domain ER diagrams + all-tables overview + table index).
    - `states.md` (15 state-machine diagrams + table of simple lifecycles).
    - `flows.md` (7 sequence diagrams, money-flow diagram, 11 journal templates).
    - `access.md` (data-access diagram, app read/write table, proposed role × module matrix, module ownership, NDPA data classes).
    - `tools/` (`generate.sh`, `introspect.sql`, `generate.mjs`, `render.mjs`, `constraint-tests.sql`).
    - `diagrams/` (41 SVGs).
  - Verification:
    - `schema.sql` loads into Postgres 17 with zero errors (after Docker Desktop recovered from the earlier HTTP 500).
    - Coverage: every table has a `[domain]` comment, every FK is indexed, every `updated_at` has its trigger.
    - 33 constraint tests pass, each asserting the specific rule that rejected the row. Fixed one false pass where a test was rejected by a missing fixture instead of the rule under test.
    - All 17 state diagrams match their column's `CHECK` values exactly (verified against the live DB).
    - All 10 numeric journal templates posted to the ledger and balanced under the real deferred trigger.
    - Module ownership covers all 96 tables exactly once.
    - All 41 Mermaid diagrams render (mermaid-cli 12, system Chrome). Fixed 7 sequence diagrams that failed on `;` in message text.
    - Full `tools/generate.sh` run passes end to end; throwaway containers removed automatically.
