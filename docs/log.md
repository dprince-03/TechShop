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
- Mobile, backend and recommendation plans (2026-10-03), documentation only:
  - Three planning agents ran in parallel (mobile, backend, recommendations). For recommendations, the owner reviewed the integration options and project impact first, then asked for every clarification to be answered, all industry approaches to be adopted, all models to be combined, and everything to be documented.
  - Written:
    - `docs/recommendations.md` (industry approaches; 11-model catalogue with pros, cons, trade-offs and data gates; hybrid ensemble; combined integration; tracking/NDPA; fairness policy; evaluation; `recommender/` structure; project impact; cost; roadmap; risks; 5 diagrams).
    - `docs/backend.md` (scaffold fixes, modular monolith, API/auth/RBAC conventions, all domain modules, revised migration order, data layer, config, observability, OWASP, realtime, testing, deployment readiness, milestones M0–M7; 8 diagrams).
    - `docs/mobile.md` (decisions, shared architecture, customer and logistics apps with route trees, payment/offline/location flows, shared packages, endpoints, release/CI/CD, compliance, milestones; 5 diagrams).
    - `docs/architecture-decisions.md` (15 reconciliations).
    - `docs/schema-changes.md` (58 proposed items, not applied).
  - Links added to `docs/database.md` and `README.md`.
  - Verification:
    - All 18 new Mermaid diagrams render to `docs/diagrams/`. Fixed 2 sequence diagrams that failed on `;` in message text.
    - Every snake_case name in the new docs checked against the live schema introspection and the proposed tables. Only enum values, config names, proposed columns and deliberately rejected alternative names remain.
    - 75 relative links resolve.
    - No files changed under `backend/`, `frontend/`, `mobile/`, `shared/`, `CLAUDE.md` or `schema.sql`.
    - No containers or test ports left running.
- Interactive mockups (2026-10-03), outside the repo (claude.ai artifacts; sources in the session scratchpad):
  - Prototype: https://claude.ai/artifact/R9tuhiEChYWm1vvVk9NSDK. System simulator: https://claude.ai/artifact/LDLEc8UZvPVinrgcYpDv8N.
  - Verification (headless Chrome scripts, no servers started):
    - Prototype: checkout with phone validation, then simulated payment, then a balanced journal in Staff › Finance. Then dispatch pack and assign, and Logistics delivery with the customer's OTP.
    - Prototype: seller registration, then staff KYC approval, then a published listing. Also wholesale tier pricing and quote, and the customer-app cart.
    - Simulator: every scenario stepped to the end with all failure toggles on.
    - Zero page errors; no horizontal overflow at 400px; light and dark checked.
    - Fixed during testing: the toast blocked clicks; the app switcher overflowed on phones; the stepper stretched in the app cart; numeric table headers were misaligned; the generator card ignored `hidden`; the canvas stretched vertically.
- System plans, mockups and simulations (2026-10-03), documentation only:
  - Written: `docs/messaging-marketing.md`, `docs/identity-access.md`, `docs/payments-finance.md`, `docs/search-catalogue.md`, `docs/orders-fulfilment.md`, `docs/trust-safety.md`; `docs/mockups/` (8 pages + README); `schema-changes.md` §4–§9 (#59–#100); `architecture-decisions.md` #16–#22; links added to `README.md` and `docs/database.md`.
  - Published (private claude.ai artifacts):
    - Messaging & marketing: https://claude.ai/artifact/JqjVaSjo3AQ5UW54tD5MRL
    - Identity & access: https://claude.ai/artifact/8m29HNrhtdrxfF6v7sVscz
    - Payments & finance: https://claude.ai/artifact/66AxZDYn9Et9dCHCifkyUZ
    - Search & catalogue: https://claude.ai/artifact/87ugTiEZ65XKffZRw6JVEj
    - Orders & fulfilment: https://claude.ai/artifact/Ap4dHGyjFv8y5QRh6Fb6qa
    - Trust & safety: https://claude.ai/artifact/GdkcuahVsdkYZsn4i7tsTS
    - (Earlier today) Prototype: https://claude.ai/artifact/R9tuhiEChYWm1vvVk9NSDK · System simulator: https://claude.ai/artifact/LDLEc8UZvPVinrgcYpDv8N
  - Verification:
    - **Headless Chrome scripts per page, for each UI:**
      - messaging: campaign approval and send;
      - identity: OTP limits, TOTP, step-up, reuse detection;
      - payments: the trial balance stayed ₦0 through 22 journals;
      - search: 9 query types;
      - orders: Σ stock movements = on hand after every action;
      - trust: risk presets and IMEI Luhn and blocklist checks.
    - **Simulations:** every scenario stepped to the end with all toggles on.
    - **Display:** zero page errors on all six; no horizontal overflow at 400px; light and dark checked.
    - **Diagrams:** all 22 new Mermaid diagrams render.
    - **Links:** 195 relative links resolve.
    - **No code changes:** nothing changed under `backend/`, `frontend/`, `mobile/`, `shared/`.
    - **No servers:** none started; only headless Chrome, which exits after each run.
  - Fixed during testing:
    - a modal backdrop swallowed form submits;
    - the lock-screen icon colour;
    - a duplicate synonym expansion ("used" applied twice);
    - an invalid sample IMEI;
    - the buy-box score column overflowing;
    - long-button wrapping.
- Project `.claude/` reconciliation (2026-10-05):
  - Reviewed the owner's new `.claude/` files. Flagged conflicts and asked the owner:
    - the DB-access wording;
    - Tailwind/shadcn and multi-tenant defaults in the skills;
    - free-text permission entries that Claude Code ignores.
  - Edited (only under `.claude/`):
    - `CLAUDE.md`, `settings.json` (validated as JSON);
    - `skills/frontend-design-system/SKILL.md`, `skills/coding-conventions/{SKILL.md,references/go-gin.md,references/react-nextjs.md}`;
    - `skills/module-scaffolding/{SKILL.md,references/go-gin-module.md}`, `skills/database-migrations/SKILL.md`, `skills/devops-deployment/SKILL.md`;
    - `skills/project-domain-knowledge/{SKILL.md,references/techshop.md}` (new).
  - Not touched: `backend/`, `frontend/`, `mobile/`, `shared/`, root `CLAUDE.md`. No commands that change the system were run.
- Database build (2026-10-05):
  - Wrote `backend/db/migrations/00002_foundation` … `00022_reference_data`. **Result:** 195 tables (plus 6 default partitions) in 22 schemas, 378 foreign keys, 9 named reference sequences.
  - Added explicit FK indexes and `updated_at` triggers with a scratchpad helper that introspects a throwaway database and appends statements to the migration that created each table: 283 FK indexes, 66 triggers.
  - **Verification** (throwaway `postgres:17-alpine` on a private Docker network; goose built from the backend module and run inside a container on that network):
    - **Goose cycle:** up 0 → 22 succeeded; `down-to 0` left 0 objects and 0 schemas; up to 22 again succeeded.
    - **Coverage:** 0 unindexed FKs, 0 tables missing the `updated_at` trigger, every table has a `[schema] purpose` comment.
    - **Constraint tests:** `docs/database/tools/constraint-tests.sql` rewritten for the schemas. **49 PASS, 0 FAIL.** New rules covered:
      - refund approver ≠ requester;
      - payout batch approver ≠ preparer;
      - campaign approver ≠ creator;
      - flash claims ≤ stock limit;
      - unique posting key;
      - closed period rejects postings;
      - unbalanced journal;
      - append-only entries and journals;
      - delivered job needs a verified code or photo;
      - no self-granted roles;
      - tracked links only to techshop.ng;
      - `delivery_otp` purpose removed.
    - **sqlc:** `go tool sqlc generate` in a scratch copy (sample queries across 6 schemas) succeeded, with 201 model types. Nothing was added to `backend/db/queries` or `internal/store`.
    - **Risk checker:** `check_migration.py` flagged only `drop`/`cascade` in Down sections and plain `create index` on new, empty tables. This is expected for a first build.
  - **Applied to the local project database** (`techshop-postgres`) over the compose network, from a runner container (`postgres` host, compose credentials):
    - before: only the empty `00001` baseline, no other tables or data;
    - after: version 22, 195 tables, 37 states, 18 ledger accounts;
    - the container was stopped afterwards, and the volume kept.
  - **Docs:**
    - `database.md` updated (source of truth, migration list, schema map);
    - `schema-changes.md` marked applied, with deviations (§10);
    - `database/schema.sql` regenerated as a `pg_dump --schema-only`;
    - `database/tools/{generate.sh,introspect.sql,generate.mjs}` now load the migrations and handle multiple schemas;
    - `erd.md` regenerated: 195 tables, 376 relationships drawn (2 composite FKs aren't drawn), 23 diagrams;
    - 48 diagrams rendered;
    - 4 stale ERD images from the old domain names removed;
    - 189 doc links resolve.
  - **Not touched:** backend Go code, Makefile, compose, `.env` files. No dependencies added. No git commands run.
  - **Ports:** no containers left running; 5432 closed.
- Backend M0 + M1 (2026-10-05):
  - **Packages added:** huma/v2, riverqueue/river (+ riverpgxv5), golang-jwt/jwt/v5, golang.org/x/crypto, pquerna/otp, aws-sdk-go-v2 (s3, credentials), johnfercher/maroto/v2; google/uuid promoted to a direct dependency.
  - **Migration:** `00023_river_jobs.sql` (River 001–008 in schema `river`).
  - **M0 platform:**
    - config with secret checks; problem+json errors; request IDs and security headers;
    - route guards (audience, permission, step-up); signed cursors; unit of work; outbox relay; River jobs;
    - realtime hub (SSE); Postgres rate limits; idempotency keys;
    - envelope encryption (AES-GCM with associated data); presigned storage (fake by default, S3 adapter);
    - RBAC checker with NOTIFY invalidation;
    - `cmd/{api,worker,migrate,seed,openapi}`; dev seed.
  - **M1 modules:**
    - **identity:** phone OTP (limits, single use, 5 attempts), sign-up, password (argon2id, delay, no enumeration), staff TOTP, recovery codes, seller new-device OTP, step-up, refresh rotation with reuse detection, sessions, addresses, consents, privacy requests, staff invites, roles, exits, MFA reset, audit and security logs, service tokens;
    - **notify:** one send pipeline with critical, transactional and marketing queues, consent, caps and quiet hours, inbox plus SSE, preferences, push tokens, delivery webhooks;
    - **catalog:** categories, attributes, products, listings with auto-checks, FCCPA guard, moderation, read-model rebuild, search with synonyms, price intents, did-you-mean and facets, buy box, saved items, stock alerts, search merchandising;
    - **inventory:** single stock path, race-free reservations, device units with IMEI Luhn and blocklist checks, transfers, blind counts with a second-person approval;
    - **files, content, tracking** (consent-gated).
  - **Verification:**
    - unit tests: money, validate, crypto, auth, notify;
    - end-to-end in containers, all passing: health, OTP sign-up and refresh-reuse revocation, addresses and TOTP step-up, staff admin (immediate role revocation, invite, exit), catalogue and search (synonyms, price intent, typo, facets, redirect, zero-result insight fixed by a synonym), buy box and FCCPA guard, saved, content, files, tracking consent, inventory (negative stock, IMEI, count separation of duties);
    - OpenAPI spec: 99 paths.
- Backend M2 — commerce (2026-10-05):
  - **New modules:** `ledger`, `marketing`, `sales`, `payments`; queries in `db/queries/{sales,payments,finance,marketing}.sql` (sqlc). No migrations added.
  - **Platform fixes found by M2 tests:**
    - the outbox relay failed whole batches when one commit queued the same unique River job twice ("ON CONFLICT DO UPDATE command cannot affect row a second time"), so no events were delivered — `uow` now drops exact-duplicate jobs before the bulk insert (unit test added);
    - `uow.Tx.Savepoint` added (nested transaction that keeps the outer one alive);
    - Huma fields with a `default` tag needed `omitempty` to be optional;
    - 11 golangci-lint findings fixed, 5 of them in M0/M1 code (unchecked writes in the SSE hub, a deprecated `strings.Title`, an overwritten logger parameter, …).
  - **Endpoints:**
    - cart (guest token, merge), payment methods, checkout preview, idempotent `POST /orders`, my orders, cancel order or line, order SSE stream, guest tracking (phone check, rate limit);
    - payment retry and status, my refunds, store credit, signed webhooks per provider;
    - staff orders (masked contact, cancel);
    - finance (dashboard, trial balance, journals, account ledger, reversal, adjustment, periods and close, refunds request/approve/reject, payments re-verify, provider health, reconciliation exceptions);
    - marketing (promotions, flash prices, coupons and unique codes, segments and preview, campaigns with submit/approve/pause/resume/cancel, journeys);
    - tracked link redirect and one-click unsubscribe.
    - The OpenAPI spec is now 156 paths / 172 operations.
  - **Jobs:** webhook processing, refund execution, order expiry (every minute), payment sweep (5 min), marketing tick (promotion schedule, due campaigns, journeys; every minute), abandoned carts (hourly), nightly ledger checks.
  - **Seed:** default Lagos addresses for both test customers, marketing consent for one, coupon `WELCOME10`, a live flash deal (5 units, 1 per person), a scheduled December sale, a "Lagos shoppers" audience, an active abandoned-cart journey; the seed now exports every fixture product's ids.
  - **Config:** payment settings validated outside development (Paystack key strength; Monnify/OPay refused until their live adapters exist). `backend/.env.example` now documents all 64 settings (it still had only the original scaffold's entries).
  - **Verification:**
    - unit tests (all passing, also with `-race`): pricing discount allocation and price hash, segment SQL binds every value, holdout/A-B bucketing, ledger line merge and account codes, fake webhook signatures, naira formatting, job de-duplication, config secret refusal (11 cases);
    - `gofmt`, `go vet`, golangci-lint (0 issues), `go build` — the backend part of `make check` and CI;
    - end-to-end against the containerised stack (fresh migrate + seed), **15/15 passing**, zero ERROR log lines:
      - guest cart → merge → preview → stale hash 409 → order → idempotent replay (same order, `Idempotent-Replayed: true`) → forged webhook 401 → signed webhook → paid → duplicate webhook acknowledged once → exactly one sale journal → trial balance zero → staff view masks the phone → guest tracking needs the right phone;
      - last-unit race: two concurrent checkouts → one 201, one `409 out_of_stock`; cancelling the unpaid winner releases the unit;
      - coupons (unknown, once per customer, released on cancellation) and flash prices (guest list price, signed-in flash price);
      - refunds: cancellation raises a system refund; officer refused above ₦500k; manager approves; order `refunded`; a self-requested refund can't be self-approved; over-refund 422; support can't approve; ledger balanced;
      - transfer: half paid keeps the order pending; the rest pays it;
      - expiry: an unpaid order is cancelled after the provider check (stack runs with `ORDER_PAYMENT_TTL=100s`);
      - campaigns: creator can't approve (separation of duties), exec lacks the permission, admin approves, recipients snapshotted, campaign sent; forged unsubscribe refused.
  - **Provider docs:** the live Paystack adapter follows Paystack's public API (initialize, verify, refund, `x-paystack-signature` HMAC-SHA512). Not exercised against the sandbox (no keys, no live calls); confirm before go-live. Monnify and OPay signature schemes are not yet confirmed.
  - **Test harness note:** earlier runs of the throwaway test Postgres created anonymous Docker volumes (the image declares a `VOLUME`). The harness now uses tmpfs, so no new volumes are created; the old ones were left in place (never delete volumes) and are listed in the final report.
- Backend M3 — logistics and fulfilment (2026-10-05):
  - **New modules:** `logistics`, `wms`; `sales.Move` (fulfilment state machine with delivery side effects) and the `DeliveryQuoter` port; queries in `db/queries/{logistics,fulfilment}.sql` plus additions to `sales.sql`. No migrations added; no new packages.
  - **Endpoints (spec now 189 paths):**
    - public delivery estimate;
    - rider: register phone, start/end shift, my jobs, job actions, offline batch sync, location upload, SSE feed;
    - dispatch (staff web and the app's dispatcher mode): board with counts, job trail, assign/unassign/return to hub, riders (create, zones, suspend/reinstate), phone approval/revocation, zones and rate bands, SSE feed;
    - staff warehouse: waves, pick lists (claim, scan), pack, manifests (create, sign), carriers, shipping view;
    - signed carrier webhook; development-only fake carrier trigger.
  - **Jobs:** logistics housekeeping every 15 minutes (close 12-hour shifts, purge location points older than 90 days).
  - **Seed:** Lagos and Abuja zones with three weight bands each (TechShop and seller rates), the fake carrier, and a rider test account (`rider@dev.techshop.ng`, phone OTP sign-in, phone `dev-rider-phone-0001` pre-approved by the seeded dispatcher).
  - **Bugs found and fixed while testing:**
    - a wrong delivery code would have deadlocked: the attempt counter was updated on a second connection while the job row was locked. It is now recorded in the same transaction after the savepoint rolls back;
    - parcel labels could contain `-`/`_` (base64): now an unambiguous uppercase alphabet;
    - the single-action route required `jobId` in the body although it's in the path;
    - the e2e helper could pick up an older sign-in code; it now waits for a new one.
  - **Verification:**
    - unit tests (all passing, also with `-race`): fulfilment state machine, rider job view never contains the customer's phone or surname, carrier webhook signature;
    - `gofmt`, `go vet`, golangci-lint (0 issues);
    - end-to-end against the containerised stack (fresh migrate + seed): **18/18 passing**, zero ERROR log lines:
      - doorstep: receive an IMEI unit → order paid → wave → wrong/unknown IMEI refused → pick → over-pick refused → pack (rider job) → assign refused off shift → shift refused without consent → shift → assign → customer token refused on rider routes → rider can't dispatch → rider job list has no phone number → wrong label refused → pickup → delivery code by SMS → offline batch applied once, replay reported as duplicate → location visible to dispatch → no proof refused → wrong code refused → delivered with code → order delivered → no tracking after shift ends → trial balance zero;
      - carrier: Kano address → pack books the fake carrier → manifest (double manifest refused) → signed hand-over → forged webhook 401 → in transit (order shipped) → delivered → replayed event acknowledged as duplicate → trial balance zero;
      - delivery estimate: Lagos ≤2 kg ₦2,500, ≤10 kg seller band ₦4,500, Kano not covered;
      - all M0–M2 tests still pass.
    - **Real call:** the order SSE stream returned the snapshot for a delivered order; the dispatch stream refused an unauthenticated call with 401.
  - **Not covered end to end yet:** commission on delivery of a marketplace (seller) fulfilment. Every seeded vendor listing ships itself, which needs seller fulfilment (M4). The code path is the same `sales.Move` → `ledger.Commission`.
  - **Build paused after M3 at the owner's request.**
- Recommendation scaffold (2026-10-06):
  - **Created:** `recommendation/` (61 files): pyproject, README, `.env.example` (no DATABASE_URL), configs (surfaces, gates, 8 model configs), the `src/techshop_recs` package (stubs per the plan), tests (including `test_no_db_access.py`), plus `infra/docker/recommendation.Dockerfile`.
  - **Verification:** all Python files compile (`py_compile`).
  - **Not done:** nothing installed or run.
- `backend/pkg/` (2026-10-06):
  - **Moved** `money`, `validate` and `crypto` from `internal/` to `pkg/` and updated imports.
  - **New packages:**
    - `pkg/pagination`: signed cursors. `internal/httpx` keeps thin wrappers, so handlers didn't change and a bad cursor is still a 422.
    - `pkg/database/postgres`: a reusable pgx pool with plain Options. `internal/database.NewPool` wires the config into it. `pkg/database/doc.go` reserves `redis/` and `mongo/`.
    - `pkg/logging`.
  - **Stay internal:** the problem+json helpers (tied to Huma and TechShop's database rules) and migrations.
  - **Verification:**
    - `go build`, `go vet` and `go test ./...` pass, including a new pagination test;
    - golangci-lint: 0 issues;
    - nothing in `pkg/` imports `internal/`.
    - End-to-end tests were not re-run, because behaviour is unchanged.
- Recommender database (2026-10-06):
  - **Rules:** updated the data rules in `CLAUDE.md` and `.claude/CLAUDE.md`, and added the critical-thinking rule.
  - **Docker:** added the `recs-postgres` service in `infra/docker/compose.recs.yml` (separate file so the main stack still starts without its password). `docker compose config` validates it, and it refuses to start without `RECS_POSTGRES_PASSWORD`.
  - **Recommender:** added `RECS_DATABASE_URL` to `recommendation/` (`.env.example`, settings), the psycopg driver to `pyproject.toml` (declared, not installed), and rewrote `test_no_db_access.py` (it now blocks Go's database instead of every database).
  - **Not built yet:** the Go pull job and the recommender's results API.
- No-Redis rule dropped (2026-10-06): decision #30 recorded; docs, skills and the `pkg/database` doc comment updated. No code or infrastructure changed.

### 2026-10-20

- Typed API client (shared/api-client):
  - **Local stack for testing:** throwaway Postgres (tmpfs, no host port), migrate and seed in containers on a private network; only the API is published on `127.0.0.1:8080`. Dev secrets are generated per run in the scratchpad and never committed. The committed spec matched a fresh `go run ./cmd/openapi` (189 paths).
  - **Added:** `openapi-typescript@7.13.0` (dev), `npm run api:types`, `make api-types`, `make test-api-client`, a CI step that fails on stale types.
  - **Layout:** `src/http/` (client, errors, request typing), `src/generated/` (contract), `src/contract.ts` (`Schema<"…">` alias), `src/format/` (money, phone), `src/domain/` (UI view models). Removed the old untyped client and `HealthResponse`.
  - **Verification:**
    - unit tests 5/5 (path and query building, problem+json → `ApiError`, headers and body, network failure);
    - type-level checks: missing path, unknown path, missing body and missing required query are all compile errors;
    - contract tests against the live seeded API: **15/15** — categories, category products, search (sort, cursor paging, forged cursor 422), suggest, product detail (buy box, 404 code), deals, best sellers, brands, states (37), delivery estimate (Lagos covered, Kano not), help list and article, banners, unpublished page 404;
    - `tsc` clean for the package; `npm run typecheck` passes for all five web apps.
  - **Backend observations (not changed; backend is off-limits during frontend work):** product detail returns `category` as `{"Slug","Name"}` (a Go struct without JSON tags); no CMS pages or banners are seeded, so legal and help pages keep their written placeholder content when the API returns 404.
- Web session kit (`frontend/packages/bff`):
  - **Files:** `crypto.ts` (seal/unseal), `config.ts` (env + secret rules), `session.ts` (session shape, refresh outcome), `proxy.ts` (refresh in proxy), `server.ts` (`createBff`: `getSession`, `requireSession`, `api`, `publicApi`, `signIn`, `signOut`; `assertSameOrigin`), `paths.ts` (safe `?next=`).
  - **Verification:** unit tests 6/6 (round-trip, tamper and wrong-key rejection, randomised sealing, production secret refusal, expiry and refresh skew, 401 → revoked vs 5xx/network → unavailable, open-redirect payloads). `tsc` clean.
  - **Identity contract tests** against the live API: **12/12** — OTP sign-in, wrong code refused, new-phone sign-up (`needs_profile` → complete → `/me`), phone validation 422, generic 401 for a wrong password and for an unknown user, refresh rotation, logout revokes refresh, `/me` needs a token, sessions list, profile update, addresses CRUD with cross-user 404, saved toggle, notifications. Running total: **27/27**.
  - **Observed:** a refresh returns a new `sessionId` (the API issues a new session row per rotation in the same family). The BFF stores whatever the API returns, so nothing depends on it staying stable.
- Market (`frontend/apps/market`) wired to the API:
  - **Moved** components into `features/*`; removed `lib/api.ts` and `lib/env.ts` (replaced by `lib/bff.ts` and `proxy.ts`).
  - **Pages now live:** home, category, sub-category, search (filters: brand facet, condition, seller type, price, in stock; sort; cursor pages; did-you-mean), deals, best sellers, product (buy box per variant, all offers, specs, accessories, related), cart, sign-in, account, saved, checkout, payment return, orders, order detail, guest tracking, delivery checker, help articles. 404s are real.
  - **Still sample:** cars (backend M7), and the written help topics and legal outlines (business text not yet written).
  - **Verification (live stack, `next dev`):**
    - typecheck, ESLint and `next build` (no env set; all routes dynamic) pass;
    - market end-to-end tests **14/14**:
      - pages show API products and prices;
      - guest cart add, quantity change (invalid quantity refused) and remove;
      - checkout redirects a guest to sign-in;
      - password sign-in merges the guest cart;
      - profile update; address validation errors, then save;
      - checkout quote, bad coupon refused, order placed, redirected to the fake Paystack page, paid, order page shows "Payment confirmed" and "Card (Paystack): Paid";
      - guest tracking finds the order with the right phone and refuses a wrong one;
      - another customer gets 404 on the order;
      - sign-out revokes the session and `/account` then asks to sign in;
      - new-number OTP → profile → signed in; wrong password refused with a generic message;
      - `?next=//evil.example` lands on `/`;
      - a cross-site `Origin` can't run an action;
      - a tampered cookie is ignored;
      - the session cookie is `HttpOnly`, `SameSite=Lax`, `Path=/` and has no `Domain`.
  - **Test kit fixes found on the way:** the Flight response's action value is the row that row 0's `a` points to, and dev debug rows share IDs. Fixed in `@techshop/testing`.
  - **Mistake caught:** a scripted edit to `bff/src/server.ts` silently didn't apply (an earlier Prettier run had changed the text). The e2e test found it (`clientApi is not a function`). The typecheck is now re-run after every scripted edit.
- Staff portal part 1 (`frontend/apps/staff`):
  - **Live:** sign-in (password + TOTP/recovery), dashboard (role-filtered live figures), orders list (status tabs, search, paging) and detail (lines, payments with re-verify, history, cancel with reason), finance (overview with provider health, refunds approve/reject with step-up, reconciliation resolve/escalate, trial balance, journals, period close), catalogue (search, moderation decisions, seller suggestions, reindex, create product + listing), account (devices with remote sign-out, permissions), notifications, global search (orders, products, people by permission).
  - **Preview pages:** modules for M4–M7 show sample data with a notice naming the milestone.
  - **Contract tests:** staff endpoints **8/8** — TOTP second step and wrong code; customer token refused; `/me` permissions; all 36 list endpoints 200 for admin; per-role 403s; masked phone on staff orders; refund requester can't approve, manager can; step-up rejects a wrong code. While writing them: 8 assumed list paths turned out to be POST-only or to need `warehouseId`. Corrected from the spec, not guessed.
  - **Staff e2e:** **8/8** — sign-in redirect keeps the return path; wrong password and wrong code refused with no cookie set; admin sees live dashboard, orders, order detail and all finance tabs (trial balance balanced); support agent's sidebar hides finance and `/finance` is refused; sign-out ends the session; catalogue tabs; product validation errors; product + listing created and found by search; global search; preview module notice; 404 for unknown modules; support agent can't open the create form.
  - **Test kit:** actions behind route groups are found by matching worker names without `(group)` segments. Seeded TOTP codes used by an earlier run are retried on the next 30-second step.
- Staff portal part 2:
  - **Modules:** dispatch, warehouse (+ pick list and manifest pages), inventory, marketing, content, search tuning, admin and HR, all wired to the API.
  - **Verification:** typecheck and ESLint pass; staff e2e **13/13**:
    - all 23 module pages render for admin with live data (Lagos zone, `WELCOME10` coupon visible);
    - full warehouse flow: a fresh paid order → serialised unit received with a valid IMEI → wave released → list claimed → every line scanned → parcel packed with a label;
    - a marketing exec sees no Approve button; a support agent is refused admin and HR;
    - the content editor saves a page, admin publishes it, and the public API serves it;
    - add a search synonym.
  - **Found while testing:**
    - publish is refused for the page's last editor (backend separation of duties). The UI now hides the button and the error now shows the API's reason. My first test used one person for both, which was the wrong expectation.
    - an e2e run that ended halfway left lines on a pick list, so the next "release wave" had nothing to pick. The test now places its own paid order first, so the full path runs every time.
- Wholesale, corporate and seller:
  - **Legal pages in all four storefront apps** now show the CMS page once it's published, and the draft outline until then (`publishedPage` in `@techshop/api-client`; `LegalPageView` and `PublishedDocument` in `@techshop/ui`).
  - **Wholesale** gets the session kit and proxy. Its notice now names what is sample (prices, quantity tiers, business accounts) and why (B2B is backend M6).
  - **Corporate and seller:** the public API client is created per call, so `next build` never needs `API_URL`.
  - **Seller Centre** stays sample: its backend is M4.
  - **Verification:**
    - market e2e adds a CMS test (**15/15** in total): the content editor saves the cookie policy; the public page still shows the outline (drafts don't leak); admin publishes; the market shows the published text;
    - smoke checks: wholesale, corporate and seller routes answer 200/404 as expected (seller on port 3006 for testing, because 3003 is in use by another project on this machine);
    - CI-equivalent run for `frontend/`: Prettier check, ESLint (7 packages), typecheck (7) and `next build` of all 5 apps with no env set all pass.
- Logistics app (`mobile/apps/logistics`), first version (stopped early when the session's usage limit was reached):
  - **Added** with `npx expo install`: expo-secure-store, expo-location, expo-camera, expo-crypto, expo-sqlite; camera and location permission strings in `app.json`.
  - **Built:**
    - phone OTP sign-in (logistics audience, per-install device ID that dispatch approves); refresh token in secure storage, access token in memory; one shared refresh; one retry on 401;
    - Today: shift start/end with location consent, foreground location every 60 s, jobs grouped pickup / deliver / done, cached for offline;
    - job screen: label scan (QR/Code 128) or typed label, start delivery, delivery code, couldn't-deliver reasons;
    - offline queue in SQLite with backoff, conflicts kept for the rider;
    - sync screen; profile with sign-out that clears local job data.
  - **Verification:**
    - rider contract tests **2/2**: a full delivery through the typed client, plus a customer refused on rider routes. The contract suite is **45/45** in total;
    - app unit tests **9/9** (queue and session);
    - `tsc` and ESLint clean.
  - **Not done:**
    - running it on a device or emulator;
    - `expo-doctor` and an Android bundle export;
    - background location;
    - photo proof;
    - the dispatcher view;
    - the customer app.
- Logistics app verification (continued, 06:32):
  - **`expo-doctor`:** 5 Expo packages in the customer app were a patch version behind SDK 57 (from the original scaffold). Aligned them with `npx expo install --fix`; both apps now pass **21/21**.
  - **Android bundle:** `npx expo export --platform android` builds the logistics app (3 MB Hermes bundle), so the app and the shared API client compile through Metro. Not yet run on a device or emulator.
  - **CI and `make check`:** the mobile job now runs the unit tests (`npm test` across workspaces; logistics 9/9). Mobile lint and typecheck pass.
- Mobile shared UI + customer app slice:
  - **Moved** `apps/logistics/src/components/ui.tsx` → `packages/ui/src/index.tsx`. Both apps' `theme/` now re-export tokens and `useColors` from it.
  - **Customer app:** `lib/api.ts`, `features/catalogue` (`useLoad`, `ProductList`, `ProductRow` with prices read out in full: "₦X, was ₦Y, Z% off"), and screens `index`, `search`, `c/[slug]`, `p/[slug]`.
  - **Verification:**
    - mobile lint, typecheck and unit tests (9/9) pass across all workspaces;
    - `expo-doctor` **21/21** for every workspace;
    - Android bundle export succeeds for both apps (customer 2.7 MB, logistics 3 MB);
    - the endpoints the customer screens use are covered by the catalogue contract tests (15/15);
    - not run on a device or emulator.
- Customer app sign-in + shared mobile session:
  - **Moved** the session class to `mobile/packages/auth` (`@techshop/mobile-auth`, renamed `MobileSession`) with its tests. Logistics now imports it.
  - **Added** expo-secure-store and expo-crypto to the customer app (`npx expo install`).
  - **Customer app:** `SessionProvider` (phone code; a new number goes to the profile step, then is signed in; restore on launch; sign-out), plus sign-in (modal) and account screens. The home screen shows Sign in or Account.
  - **Verification:**
    - new contract test for the `customer_app` audience (new number → profile → `/me` → refresh rotation → session listed → logout);
    - full contract suite **46/46** on a freshly seeded stack;
    - mobile lint, typecheck and tests (queue 5/5, session 4/4) pass;
    - `expo-doctor` 21/21 for all four workspaces;
    - Android bundles export (customer 2.7 MB, logistics 3 MB);
    - not run on a device or emulator.
- Customer app cart (06:43):
  - **`CartProvider`:** the account's cart when signed in; a guest cart keyed by `X-Cart-Token` (token kept in secure storage), merged into the account on sign-in.
  - **Screens:** add to cart from each product offer (the best offer is the primary button); a cart screen with quantity −/+, remove, and API totals and problems. Checkout is clearly marked "coming soon in the app".
  - **Verification:**
    - typecheck and ESLint pass;
    - `expo-doctor` 21/21 for all workspaces;
    - Android bundle export (2.8 MB);
    - mobile tests pass;
    - the cart endpoints, including guest token and merge, are covered by the commerce contract tests (8/8);
    - not run on a device.
- Customer app orders (06:45): order list and order detail (parcels by seller, totals, history; awaiting-payment note), and guest tracking (order number + checkout phone) linked from home and account. Typecheck, ESLint, `expo-doctor` 21/21 and the Android bundle export pass. The endpoints are covered by the commerce contract tests. Not run on a device.
- Customer app checkout (06:48):
  - **Added** `expo-web-browser` (`npx expo install`).
  - **`app/checkout.tsx`:** address choose or add, payment method, API quote, pay (price hash + idempotency key), provider page in the in-app browser, a payment status check, then the order page; bank transfers show the account in place. The cart's button now goes to checkout (or to sign-in for guests).
  - **Verification:** typecheck and ESLint clean; `expo-doctor` 21/21; Android bundle export (2.8 MB). The endpoints (preview, place, idempotent replay, stale-hash 409, fake payment, payment status) are covered by the commerce contract tests. Not run on a device.
- Final regression run (06:49–06:54), on a freshly migrated and seeded stack with `next dev` for market and staff:
  - contract tests **46/46**;
  - market e2e **15/15**;
  - staff e2e **13/13**;
  - unit tests: api-client 5/5, bff 7/7, mobile queue 5/5, mobile session 4/4.
  - All servers and containers started for testing were stopped afterwards. Ports 3000 and 3003 belong to other projects and were left alone.
- Dev ports 3001–3005 → 4101–4105:
  - **Updated** 13 files: app scripts, e2e defaults, nginx, three READMEs, `backend/.env.example` and `internal/config/config.go`; also the project-domain note.
  - **Verification:**
    - no old port references remain;
    - backend `gofmt`/`go vet`/config tests pass and the API builds;
    - `nginx -t` passes on the new config;
    - market and staff started with their own `npm run dev` on 4102 and 4105: market e2e **15/15**, staff e2e **13/13**.
  - Servers and the test stack were stopped afterwards.
- Mobile apps tested on the Android emulator (Pixel 8 AVD, Android 17, Expo Go 57; API reached via `adb reverse` because the apps' `.env.local` points at `localhost:8080`):
  - **Customer journey, passing:** browse → category → product → add to cart as a guest → quantity change → sign up by phone code (new number → profile) → guest cart merged → checkout with a new address (state matched from the API list) → API quote (VAT ₦0 on a marketplace item, as designed) → provider page in the in-app browser → order placed → payment confirmed → order list and detail.
  - **Rider journey, passing:** sign in by phone code → the phone registers for approval → shift refused until dispatch approves → approved → shift started → assigned job appears (first name and address only, no phone) → wrong label refused → pickup by label → offline "start delivery" queued → survives an app restart offline ("1 waiting to sync") → synced on reconnect → wrong delivery code refused → delivered. The customer app then shows the order as Delivered. End shift; sign-out revokes the session (logout 204).
  - **Stale login after a backend reset:** both apps sign out cleanly.
  - **14 issues found and fixed:**
    1. customer: "Cars" category shown though cars aren't sold in the app → vehicle categories hidden;
    2. both apps: double top gap under navigation headers → `Screen underHeader`;
    3. customer: sign-in from the cart didn't return to the cart → returns to the previous screen;
    4. customer: order screen couldn't refresh → pull-to-refresh (`Screen onRefresh`);
    5. customer: stale "checkout coming soon" text;
    6. both: buttons could keep a stale "busy" state for screen readers → explicit booleans;
    7. logistics: "Back to today" crashed with no history (`GO_BACK not handled`) → back or replace;
    8. logistics: no in-app location disclosure before the OS prompt (a Google Play requirement) → a disclosure card before the first request;
    9. logistics: shift start waited ~25 s for a GPS fix → 8 s limit, then the last known position (now ~4 s);
    10. logistics: the phone was never registered, so dispatch could never approve a new rider → registered at sign-in and again when a shift is refused;
    11. logistics: Today didn't refresh after a pickup;
    12. logistics: Today didn't show the "waiting to sync" count (11 and 12: reload on focus);
    13. logistics: the job screen said "not in your list" while still loading → spinner;
    14. logistics: sign-out stayed on Profile → back to sign-in.
  - **After the fixes:** lint, typecheck, unit tests, `expo-doctor` 21/21 and both Android bundle exports pass. Fixes 1–5, 7, 9, 10, 11–14 were re-checked on the emulator. Fix 8 was not seen: location permission was already granted from the first run, and the card only shows before the first system prompt. Fix 6 is checked in code only.
  - **Notes:** the machine restarted mid-test (the emulator, test database and scratch scripts were rebuilt). While checking a payment, one status read went straight to the throwaway test database; everything else used the API.

## 2026-10-20 — CI fix: generated API types check (PR #1)

- **Failure:** CI job "Frontend (web apps)", step "Generated API types are up to date". `npm run api:types` writes raw openapi-typescript output (4-space indent); the committed `schema.ts` had been Prettier-formatted, so `git diff --exit-code` saw the whole file as changed.
- **Check:** regenerated locally, then ran Prettier on it: no diff against the committed file, so the contract itself was in sync; only formatting differed.
- **Fix:** `api:types` in `frontend/package.json` now runs Prettier on the generated file, so output is the same every run.
- **Verification (local, same steps as the CI job):** generated-types check passes; api-client typecheck + unit tests 5/5; format:check, lint, typecheck, build pass.

## 2026-10-21 — Fix: dispatch responses no longer carry the delivery-code hash

- **Issue (found during frontend work):** `GET /dispatch/board` and `POST /dispatch/jobs/{id}/assign|unassign` returned sqlc rows with `otpHash` and `otpAttempts`. Handlers nulled the hash at runtime, but the field stayed in the contract, so any handler that forgot would leak it.
- **Fix:** response types `DispatchJob` and `BoardJob` in `internal/logistics/http.go`, built field by field from the row with no OTP hash or attempt count. OpenAPI spec and generated TS types regenerated; no web or mobile code used those fields.
- **Tests:** unit `TestDispatchJobHidesDeliveryCode`; e2e board and assign assertions now fail if either field appears.

## 2026-10-21 — Backend restructure (move, not rewrite)

- **Moves (`git mv`, history kept):** 13 feature packages → `internal/modules/<feature>`; 15 plumbing packages → `internal/platform/<pkg>`. Import paths rewritten by `sed` over exact module paths; `gofmt` re-sorted imports. `internal/store`, `internal/seed`, `pkg/` unchanged.
- **Splits (cut, not retyped):** each module's `http.go` → `routes.go` (renamed with `git mv`) + `dto.go` (response/request types and their builder functions), done with a small `go/parser` script; imports pruned with the already-installed `gopls imports`. Ledger's nightly integrity check moved to `ledger/jobs.go`. `ledger.go`, `wms.go`, `notify.go` → `service.go`. Small single-file modules (content, files, tracking) stay one file until they grow.
- **Proof nothing changed:** a line-multiset check per split file (every original line present once, nothing added); `go build`, `go vet`, unit tests pass; regenerated `openapi.yaml` is byte-identical; full e2e suite passes.
- **Docs updated:** root `CLAUDE.md` config paths, `infra/README.md`, `docs/backend.md` diagram, `docs/messaging-marketing.md` template path, coding-conventions Go note (old "takes precedence" note replaced by the new layout).
- **E2E note:** `TestUnpaidOrderExpires` needs `WORKER_IN_PROCESS=true` and `ORDER_PAYMENT_TTL=100s` on the test stack; without them it times out (test-setup gap, not a code bug). With them: pass.

## 2026-10-21 — Database docs added (owner request)

- New in `docs/database/`: `DATA-MODEL.md`, `SCHEMA.md`, `MIGRATIONS.md`, `RLS-POLICIES.md`, `SEED-DATA.md`. Short, linking to the existing `erd.md`, `states.md`, `access.md` instead of duplicating them.
- Facts checked against the code: 205 tables in 22 schemas + `river` (counted from migrations); token audiences from `platform/auth/tokens.go`; CI runs migrations up only; seed roles from the seed output; vendors and warehouses from `seed/commerce.go`.
- `RLS-POLICIES.md` states the truth: **no RLS today**; access is enforced in the API. It flags a real gap (one DB user owns everything, against rule 01 §6) and recommends a role split, then RLS on seller-scoped tables. Owner decision needed.

## 2026-10-21 — M4 step 0: least-privilege database users (owner approved RLS in M4)

- Migration `00024_database_roles`: group roles `techshop_rw` (DML on all app schemas incl. future tables via default privileges; update/delete revoked on audit log, ledger, stock movements, delivery events) and `techshop_ro` (select on non-personal schemas). `ensure_monthly_partitions` made `security definer` with a pinned `search_path`.
- Config `DATABASE_MIGRATE_URL` (owner, migrations only); `cmd/migrate`, startup migrations and `make migrate-*` use it. `CheckAppRole` refuses a superuser/owner login outside development, warns locally. `make db-users` + `infra/docker/postgres/app-users.sql` create `techshop_app` (password via env, never stored).
- **Verified** on a throwaway stack (migrate as owner, seed + API as `techshop_app`): app login cannot create/drop tables or change history (permission denied), can read and insert; full backend e2e suite passes as `techshop_app`; no permission errors in API logs.
- Owner action for local setup: run `make db-users APP_DB_PASSWORD=…`, then set `DATABASE_URL` to `techshop_app` and `DATABASE_MIGRATE_URL` to the owner (see `backend/.env.example`). Until then it still works with a warning.

## 2026-10-21 — M4 slice 1: seller onboarding (shops, KYC, bank accounts, KYC review) + RLS

- New module `internal/modules/sellers` (`service.go`, `routes.go`, `dto.go`, `rules.go` + tests). 11 endpoints: Seller Centre `POST/GET /seller/shops`, `GET …/onboarding`, `POST …/kyc`, `GET /seller/banks`, `POST/DELETE …/bank-accounts`, `POST …/default`; staff `GET /staff/vendors`, `GET /staff/vendors/{id}`, `POST /staff/vendors/kyc/{id}/review`, `POST …/reveal`.
- `payments/banks.go`: name enquiry (`ResolveAccount`) and bank list via Paystack (`/bank/resolve`, `/bank` with cursor paging; checked against Paystack API references found by web search, official page returned 403); dev fake documented in code.
- Migration `00025_seller_rls`: RLS on `sellers.sellers`, `seller_members`, `seller_kyc_submissions`, `kyc_documents`, `seller_bank_accounts`.
- **Tests:** unit (ID formats, slugs, name matching); e2e `TestSellerOnboardingAndKycReview`, `TestSellerKycRejectAndResubmit` (validation, file ownership, fake name enquiry match/mismatch/not-found, defaults and deletes, 404 for another seller, audience separation, permission refusal, approve/reject/resubmit, reveal with reason, no full numbers or hmac/encrypted fields in responses). Full backend e2e suite passes as `techshop_app` (20 tests). `golangci-lint`: 0 issues. API: 200 paths; TS types regenerated; api-client typecheck passes.
- **RLS proved in Postgres** as `techshop_app` with `app.seller_id` = shop B: shop A's bank and KYC rows invisible (0), only 1 shop row visible, `UPDATE` of shop A affects 0 rows, `INSERT` into shop A refused by policy.

## 2026-10-21 — M4 slices 2–3: seller listings and seller-shipped orders

- **Listings:** `sellers/listings.go` (create, list, detail with moderation history, price, stock, pause/resume, edit and resubmit); `catalog/listings.go` (reprice with bait-price guard, pause, edit-and-resubmit); `validateListing` gives clear errors for missing condition notes or stock. Migration `00026_listing_moderation` (`auto_flagged`, listings RLS).
- **Orders:** `sales/seller.go` (accept with handling-time deadline, cancel with refund and restock, SLA job every 5 min, 3-strike enforcement, new-order message, dev-only `/dev/sla/{id}/accept|pack` trigger); `sellers/fulfilment.go` (queue, detail, accept, pack, cancel). Migration `00027_fulfilment_rls`.
- **Bugs found by tests and fixed:** Huma ignores path params in embedded structs (inputs now declare them explicitly); the bait-price guard included the listing's own new price in the median it compared against (now `CatalogMedianPriceOthers`); seller listing input no longer requires `fulfilledBy`.
- **Tests:** e2e `TestSellerListingsAndModeration`, `TestSellerOrdersAcceptPackCancelAndSLA` (queue, privacy of the address, accept/pack/cancel transitions, refund request, stock release, late accept, late pack, 3-strike review, ledger balanced). RLS checked in Postgres as `techshop_app`: another shop's active listing visible, in-review hidden, update/delete 0 rows, insert refused; another shop's fulfilment readable but `UPDATE` 0 rows. `golangci-lint` 0 issues; API 211 paths; TS types regenerated.
- **Blocked:** hand-over/delivery of seller-shipped orders (owner questions in `plan.md`).

## 2026-10-21 — M4 slice 4: reviews, reports, seller ratings

- `internal/modules/reviews` (`service.go`, `routes.go`, `dto.go`, `rules.go` + tests); `sellers/reviews.go` (Seller Centre list and answer, in the shop-scoped transaction); `db/queries/reviews.sql`; migration `00028_review_responses`; search read model weights unverified reviews at 0.5.
- **Bugs found and fixed:**
  - existing catalogue refresh job deduplicated by args within 2 s **including finished and running jobs**, so a second change to a product inside that window (review then publication; price then stock) was dropped until the nightly rebuild. Uniqueness removed (a refresh rebuilds one row; River requires running jobs in any unique set, so a narrower set couldn't fix it).
  - two request types named `ReviewInput` crashed schema registration at boot; the KYC one is now `KycDecision`.
- **Tests:** unit (hold rules incl. Nigerian and +234 phone formats, links, incentive phrases, normalisation); e2e `TestReviewsRatingsAndModeration` (verified purchase via a real carrier delivery, one per product, claim of another's purchase refused, held review hidden publicly with reason visible to staff, publish, weighted average checked against the previous summary, verified filter, read model updated, reports once/not own/settled, seller rating once and only own delivery, seller can't answer reviews of items it didn't ship). Test made re-runnable on a used database.
- Web and mobile typecheck pass against the regenerated types; API 218 paths.
- **Not covered yet:** a seller answering a review end to end needs a delivered seller-shipped order, which waits for the hand-over decision.

## 2026-10-21 — M4 slice 7a: Seller Centre web app wired to the API

- **Seller app** (`frontend/apps/seller`), grouped by feature: `features/{auth,session,shops,kyc,payouts,listings,orders,reviews}`, `lib/{bff,actions,format}.ts`, `components/RowAction.tsx`, `proxy.ts`. Pages under `/centre`: shop chooser, open a shop, overview checklist (next step from the API), verification (KYC with document upload), payout accounts, listings (filter, new = search → choose product → offer, detail with price/stock/pause/resubmit and moderation history), orders (queue with deadlines, accept, pack, cancel with reason), reviews (answer once). Real sign-in (phone OTP + new-device code) replaces the preview form; `/register` explains the steps and starts sign-in. Header shows "Your shop / Sign out" when signed in.
- **KYC uploads:** files post to a Server Action, which uploads them server-side to the presigned URL (with the signed Content-Type) and completes them; the browser never talks to the API or storage. `serverActions.bodySizeLimit` = 41 MB (four 10 MB documents; checked in the bundled Next 16 docs). **Deployment note:** nginx `client_max_body_size` for the seller host must allow this, and is the place to cap it.
- **Shared code moved, not copied** (reusables in one place): `SignInForm` (market → `@techshop/ui`, takes the action as a prop), `Table`/`Status` (staff → `@techshop/ui`), new `ActionForm`/`Field` in `@techshop/ui`, `failure()`/`ActionState` and `createAuthFlow().handle(form)` in `@techshop/bff` (market's sign-in action now uses it). Follow-up: market and staff still have their own form-error helpers; fold them into `failure()`.
- **Tests:** new `apps/seller/test/seller.test.ts` (2/2): sign-up via OTP, open shop, KYC validation then real upload, ID shown only as last 4, bank validation and name match, staff approval via API, product search/choose/offer with validation, moderation via API, price and stock, other seller gets 404, pages need a session. Market e2e 15/15 and staff e2e 13/13 re-run after the moves. Frontend format, lint, typecheck and build pass.
- **Correction to the slice 2–3 entry:** "Huma ignores path params in embedded structs" was too strong — other routes embed a query-param struct and it works. Verified only: declaring the path fields explicitly in those handlers fixed the requests; the cause was not confirmed.

## 2026-10-21 — M4 slice 7b: staff screens for vendors and review moderation

- **Vendors module is live** (`/vendors`, `/vendors/[id]`): to-review / all / active / suspended tabs with search and paging; detail shows verification (ID as last 4 only), duplicate-ID and shared-bank-account warnings, payout accounts with the bank-name check against the legal holder, and members. Actions: approve / ask for more information / reject (note required unless approving), and reveal the full ID with a reason (the API demands a fresh authenticator check; the existing step-up button handles it; every reveal audited).
- **KYC documents** open through `/vendors/[id]/documents/[fileId]`, which asks the API for the 5-minute link only on click, so the API's KYC-view audit records real openings, not page loads (`no-store`, `no-referrer`).
- **Review moderation** is a "Reviews" tab in Catalogue (catalog.moderate): held and reported reviews with the reasons, publish / hide / reject.
- Staff test helpers moved to `apps/staff/test/helpers.ts` (shared by both test files). New `test/vendors.test.ts` (3/3): list → detail (no full ID in the page) → document opens via link → reveal refused without a reason, then shown → decision needs a note unless approving → approve; dispatcher turned away; held review appears with its reason and is hidden. The old preview-module assertion now uses `/b2b` (Vendors went live). Staff e2e 16/16; frontend format, lint, typecheck, build pass.

## 2026-10-21 — M4 slice 7c: reviews on the market

- **API:** `POST /api/v1/me/order-lines/{lineId}/review` reviews the product of the caller's delivered order line (verified purchase), so order pages don't need product slugs. Backend e2e reviews test now uses it for the verified review, plus 409 for a second review through it and 404 for someone else's line.
- **Market** (`features/reviews`): product pages show the rating summary (weighted, with the verified count), up to 10 reviews (verified first, filter to verified only), seller answers, and a "Report this review" form for signed-in shoppers. Delivered parcels on the order page get "Review" per item (→ `/orders/[number]/review/[lineId]`) and "Rate {seller}". All forms use the shared `ActionForm`.
- **Test kit:** `tokenForPhone` and `deliveredOrder` (Kano address → fake carrier → delivered) in `shared/api-client/test/kit.ts`, exported by `@techshop/testing`.
- **Tests:** market e2e "reviews" (fresh customer with a delivered order; the web session signs in with a password set through the API because the API allows one code per number per minute; star rating required; review goes live; seller rated once, second refused; review shows on the product page as verified; another shopper reports it once, second refused). Market e2e 16/16; frontend format, lint, typecheck, build; api-client unit 5/5; backend unit and lint pass.

## 2026-10-21 — M4 slice 7d: reviews in the customer app

- `@techshop/mobile-ui`: new `StarPicker` (five 44 px buttons, announced as a radio group, `rating` token for the stars).
- Customer app `features/reviews`: product screen shows the weighted summary and latest reviews (verified first, seller answers) with "Report this review" (reasons) for signed-in shoppers; delivered parcels on the order screen get "Review {item}" (→ `/review/[lineId]`, uses the order-line endpoint) and "Rate {seller}".
- **Verified on the Android emulator (Pixel_8, Expo Go)** against the local stack: product screen shows the reviews section; signed in by phone code as a fresh customer with a delivered order (prepared with the kit's `deliveredOrder`); order screen shows the review and rating controls; seller rated ("Thanks for rating the seller."); posting with no stars says "Choose 1 to 5 stars."; review posted live; another shopper's review reported ("A moderator will look at it."). Screenshot checked.
- Mobile lint, typecheck (after regenerating Expo Router's typed routes), unit tests (5 + 4), Expo doctor 21/21, Android bundle export pass. Metro and the emulator stopped afterwards.

## 2026-10-21 — Cleanup: one form-error helper

- Market account actions and staff `runStaffAction` now use `failure()` from `@techshop/bff` (staff keeps its own wording for refusal codes such as `separation_of_duties` on top). Step-up, session-ended and field-error handling behave as before; field errors now come with "Check the highlighted fields." in staff too. Market e2e 16/16, staff e2e 16/16 after the change.

## 2026-10-21 — Payments: cached bank list, Paystack adapter tests

- The provider bank list (live mode) is cached in memory for 24 hours, and a stale list is served if the provider is down; before, each Seller Centre page fetched every page of Paystack's list.
- Unit tests for the live Paystack calls against a local stand-in with the documented shapes: the bank list follows `meta.next` and skips inactive banks; name enquiry returns the account name and maps a failed lookup to `ErrAccountNotFound`.

## 2026-10-21 — Full regression on fresh stacks (after M4 work)

- **Run 1** (fresh stack): backend e2e ok (23 tests); api-client contract 46/46; market 15/16 (**1 failure**, test name not captured); staff 16/16; seller 2/2. Market re-run on the same stack: 16/16.
- **Run 2** (fresh stack, same order, full market log kept): backend e2e ok; contract 46/46; market 16/16.
- **Open:** one intermittent market e2e failure, cause unknown (not reproduced in two further runs). If it comes back, the full log in the run will name it.

## 2026-10-21 — M5 slice 1: returns (RMA)

- New module `internal/modules/aftersales` (`service.go`, `routes.go`, `dto.go`, `rules.go` + tests), `db/queries/aftersales.sql`. 9 endpoints: customer `POST /me/orders/{number}/returns`, `GET /me/returns`, `GET /me/returns/{rma}`; staff (`returns.manage`) list, detail, decision, receive, inspect an item, raise the refund. Payments gained `RequestReturnRefund` (refund linked to the return, same dual control).
- **Tests:** unit (7-day window, good standing, refund share, serial match, restock condition); e2e `TestReturnInspectRestockRefund` (change-of-mind refused, quantity limits, auto-approval, only your own return, no re-decision, refund only after inspection, receive → inspect → open-box stock +1, no double inspection, refund request reaches finance, ledger balanced) and `TestReturnSerialMismatchOpensRiskCase` (phone with a real IMEI delivered by carrier; pickup return; wrong IMEI → 409, return rejected and visible to the customer).
- **Test harness fixes found on the way:** each Go e2e test now uses its own client address (`X-Forwarded-For` from the test name), because the suite outgrew the API's 20-codes-an-hour per-network limit; tests that pick in the warehouse take turns (a wave takes every waiting paid line, so parallel tests could steal each other's orders); the delivery helper accepts IMEIs (`"*"` for any serialised line).
- Full backend e2e on a fresh stack: 25 passed. Lint 0 issues; API 228 paths; web and mobile typecheck against the new types.

## 2026-10-21 — Returns policy from the researched decisions (plan 2026-10-21, items 2–3)

- Change-of-mind returns now allowed within 7 days: the customer confirms the item is unopened, brings it to a store (`dropoff_only` otherwise), hygiene categories refused (`audio`; `underwear`/`cosmetics` if added). Never auto-approved; support can reject at receipt if it arrives opened.
- Our-fault returns (faulty, damaged, wrong, not as described) refund the delivery fee when the whole parcel comes back; partial returns and change of mind don't.
- Tests: unit (change-of-mind eligibility, delivery refund); e2e `TestReturnChangeOfMind` (waits for support, item-only refund, approve → receive → reject without reason refused → reject with reason) and the existing returns test now checks delivery is included and the new change-of-mind refusals. Full backend e2e on a fresh stack: 26 passed; lint 0 issues; web and mobile typecheck.

## 2026-10-21 — Seller hand-over at a TechShop hub (branch feat/m4-returns-handover, PR #5)

- Moved work off `feat/m4-marketplace` after finding PR #4 already merged. Two stranded commits (84896ab, 0b729a9) are now on `feat/m4-returns-handover`.
- **Backend:**
  - Migration 00029 adds `sales.fulfilments.handover_by`.
  - The seller pack action now creates a `TSS-` parcel label and sets the deadline (`sales.PackSellerFulfilment`, `wms.SellerParcel`).
  - New `POST /api/v1/staff/wms/dropoffs` (`wms.ReceiveDropoff`). It takes the label and moves the parcel to handed_over, then raises a rider job or carrier booking through a shared `handoff` that pack also uses now.
  - New `GET /api/v1/seller/dropoff-points`.
  - The SLA job now enforces the hand-over deadline, and the dev route gains `/dev/sla/{id}/handover`.
  - `packed → cancelled` is now allowed in the flow, but only `CancelFulfilment` uses it, for seller late hand-over.
- **Web:**
  - The Seller Centre order page shows the drop-off panel (code, deadline, hubs) and plain status labels.
  - The staff Warehouse page gains a "Seller drop-offs" desk at warehouses and hubs.
  - `@techshop/testing` now exports `placePaidOrder`.
- **Verification (fresh throwaway stack):**
  - Backend e2e: 27/27 pass, including the new `TestSellerDropoffToDelivery`. It covers the Lagos rider job, Kano carrier delivery, late hand-over cancel and refund, store refusal, unknown label, double scan, and the seller token refused. `go test ./internal/modules/sales` passes, including the deadline unit test. Lint: 0 issues.
  - Seller web e2e: 2/2. Staff web e2e: 16/16. Seller and staff typecheck and lint pass.
  - Screenshots of the new panels were not taken this session.

## 2026-10-21 — Seller payouts (branch feat/m4-seller-payouts)

- **Backend (commit 35f6c4d):**
  - Migration 00030: `order_lines.seller_earning_kobo`, batch `submitted_at` and second approver, payout `transfer_code` and `attempts`.
  - New `internal/modules/payouts` (rules, service, jobs, routes, dto, seller views).
  - Paystack transfer adapter plus dev fake in `payments/transfers.go`. Transfer webhooks are relayed as `transfer.updated`.
  - Ledger gains `PayoutPaid` and `PayoutReversed`.
  - Seller routes: `/seller/shops/{id}/balance`, `/payouts`, `/payouts/{id}`.
  - Staff routes: `/staff/finance/payout-batches` (prepare, list, get, exclude, submit, approve, cancel) and payout holds.
  - Dev routes: `/dev/payouts/age/{seller}` and `/dev/transfers/{ref}/{event}`.
  - API types renamed (`PayoutBatch`, `BatchPayout`, `SellerBalance`, `PayoutItem`, `SkippedSeller`) after a Huma schema name clash with an existing `Item`.
- **Web:**
  - Seller Centre › Payouts shows the balance (available, pending, holds in plain words), payout history, a statement page, and payout accounts. The nav label is now "Payouts".
  - Staff Finance gains a Payouts tab (prepare, list) and a batch page (submit, approve with step-up, leave a seller out, optionally with a hold, cancel).
- **Test stack:** the scratchpad was cleared between sessions, so `stack.sh` was rebuilt from the transcript. It now migrates as the owner, creates `techshop_app`, and runs the worker in-process. The test stack sets `PAYOUT_APPROVAL_THRESHOLD_KOBO=50000000` so the two-approver path runs.
- **Verification:**
  - Backend e2e on a fresh stack: 28/28.
    - New `TestSellerPayoutBatch` covers: pending until the window passes, the new-account hold, a manual hold skipping the seller (officers can't release it), prepare → submit → two approvals (same approver refused), one paid and one bank-refused payout, statement, replayed transfer webhooks changing nothing, and no double payment in the next batch.
  - Database checks on the test stack: one payout journal per paid payout, none for the failed one; the two-approver batch recorded both approvers.
  - Unit tests: payout rules, plus the Paystack transfer adapter against a local stand-in (queued, duplicate-reference retry resolved by verify, refused, status mapping).
  - Lint: 0 issues.
  - Web: Seller Centre 2/2, staff 17/17. Typecheck and lint clean.
  - A back-to-back third e2e run hit the checkout rate limit (429) for the shared seeded customer. That is expected behaviour, not a regression.
  - Screenshots not taken.
