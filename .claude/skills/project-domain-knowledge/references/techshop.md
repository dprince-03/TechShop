# TechShop

**Client / owner:** the repo owner (a real Nigerian business)
**Status:** active. Planning is done and the scaffold exists; the web apps are built as interfaces only, the API is not built yet.
**Repo(s):** this monorepo: `frontend/`, `mobile/`, `backend/`, `shared/`, `infra/`, `docs/`
**Environments:** local only so far. Planned hostnames: `techshop.ng` (corporate), `market.`, `wholesale.`, `seller.`, `staff.`, `api.`. Hosting is an open decision.

## 1. Purpose

TechShop sells technology in Nigeria three ways:
1. **its own stock** (retailer);
2. **other sellers' stock** (a multi-vendor marketplace);
3. **to businesses in bulk** (supplier / wholesale).

It serves both consumers (B2C) and businesses (B2B). Catalogue: phones (all kinds), laptops (business, gaming, workstation), accessories, gaming, smart home, office, workstations, and cars.

## 2. Tenants / organizations

**Not multi-tenant.** One company runs everything. Data is separated by **ownership**:

| Owner | Scope rule |
|---|---|
| Customer | Their orders, addresses, carts, returns (`customer_user_id`) |
| Seller (vendor shop) | Their listings, fulfilments, payouts (seller membership: owner, manager, staff, finance) |
| Business (B2B buyer) | Their quotes, orders, invoices, credit (business membership: admin, buyer, approver) |
| Staff | Permissions per role, plus optional data scopes (store, warehouse, delivery zone) |
| Rider | Their assigned delivery and pickup jobs (bound device) |

## 3. User roles & permissions

| Role | Can | Cannot |
|---|---|---|
| Customer | Shop, pay, track, return, review delivered items | See other customers' data |
| Seller member | Manage their shop's listings, stock, orders and payouts (by seller role) | Edit product content owned by TechShop; see other sellers' data |
| Business member | Request quotes, order on credit up to limits, approve (approver role) | Exceed approval limits |
| Staff (one portal, 20 modules) | Only what their roles grant (RBAC, `module.action` permissions) | Approve their own requests (refunds, payouts, campaigns, adjustments); grant themselves roles |
| Rider / dispatcher | Riders: deliver assigned jobs with the customer's code. Dispatchers: assign jobs in their zones | Mark delivered without a verified code or photo |

The staff role list is **proposed, not approved** (`docs/identity-access.md` §7.2).

## 4. Glossary

| Term | Meaning in this project |
|---|---|
| Listing | One seller's offer of a product variant in a condition (price, stock, warranty) |
| Buy box | The listing shown first when several sellers offer the same product |
| Fulfilment | The part of an order handled by one seller; orders are split by seller |
| UK-used / "tokunbo" | Imported used device (`uk_used`) |
| Delivery code | The 6-digit OTP the customer gives the rider on delivery |
| Kobo | 1/100 naira; all money is stored as integer kobo |
| Moniepoint | Online payments go through **Monnify** (Moniepoint's gateway); provider value `moniepoint` |
| Step-up | Re-entering the authenticator code before a sensitive action |
| Posting key | Unique key that stops a money event being posted to the ledger twice |

## 5. Core workflows

1. **Buy:** cart (grouped by seller) → checkout → stock reserved → pay (Paystack, OPay or Monnify) → webhook verified with the provider → paid → each fulfilment picked or accepted → handed over → delivered with the delivery code.
2. **Sell:** seller registers → KYC (NIN, selfie, bank name match) → approved → lists against catalogue products → moderation → live → orders → paid after delivery plus the return window, minus commission.
3. **B2B:** quote → accept → invoice on credit or prepay → partial deliveries → payment to the business's reserved account.
4. **Return:** request → approve → pickup → inspect (serial must match) → refund or replace.

Full detail: `docs/database/flows.md` and the system docs.

## 6. Business rules

- **Only the Go backend talks to PostgreSQL**, through its data layer, over the Docker network. Web, mobile and the Python recommender never touch the database. (Owner, confirmed 2026-10-05.)
- Money is always integer kobo; prices are VAT-inclusive for consumers. VAT treatment is a placeholder until finance confirms.
- Cars are never added to the cart: customers book a viewing.
- "Was" prices must be real (no higher than the 30-day maximum; FCCPA).
- Marketing messages only go out with per-channel consent (NDPA). Sign-in and delivery codes are always sent.
- Sensitive staff actions need an MFA step-up and a second person.
- Seller and staff web apps are never indexed by search engines.

## 7. Data & compliance

- **Personal data collected:** names, phones, emails, addresses, NIN and bank details (sellers, encrypted), device and location (riders on shift), behaviour events (with consent).
- **Applicable regulations:** NDPA 2023, FCCPA (consumer protection), NCC Do-Not-Disturb rules for SMS, card schemes via providers (PCI DSS scope kept minimal: hosted checkout). Legal review is pending for several items.
- **Retention:** see each system doc (for example, message bodies 90 days, auth events 12 months).

## 8. Integrations

| Service | Purpose | Notes |
|---|---|---|
| Paystack, OPay, Monnify | Payments, transfers, refunds | Webhooks verified, then re-verified with the API |
| Termii (+ backup) | SMS: codes, delivery codes, receipts | Critical lane |
| Email provider(s) | Transactional + marketing (separate subdomains) | Owner to choose |
| Expo push | App notifications | |
| S3 / MinIO | Files: images, KYC documents, delivery proof | Presigned uploads |
| KYC provider | NIN, selfie, CAC checks | Owner to choose; manual review at launch |
| Interstate carrier | Deliveries outside rider cities | Owner to choose |

## 9. Tech stack & architecture notes

- **Web:** Next.js 16 apps (corporate, market, wholesale, seller, staff), custom CSS + design tokens, no Tailwind.
- **Mobile:** Expo SDK 57 apps (customer, logistics).
- **API:** Go 1.26 + Gin, pgx, goose, sqlc, slog, River jobs, SSE. Redis allowed for caching and fast counters (no-Redis rule dropped 2026-10-06); Postgres stays the source of truth.
- **Recommender:** Python, through internal API endpoints only.
- **Plans and decisions:** `docs/` (start at `README.md` › Planning docs) and `docs/architecture-decisions.md`. Interactive mockups are in `docs/mockups/`.

## 10. Known constraints & gotchas

- Port 3000 is taken by AdGuard on the owner's machine; the corporate site uses 3005.
- `frontend/` and `mobile/` have separate installs (different React versions).
- Backend work is **on hold** until the owner says it's time. That includes the dev test accounts and the weak-secret checks.

## 11. Decision log

The full, dated log is `docs/plan.md`. Key items:

| Date | Decision | Reason | Decided by |
|---|---|---|---|
| 2026-10-03 | Hybrid, gated recommendation system; Python never touches the DB | Quality plus the data rule | Owner |
| 2026-10-03 | System plans for messaging, identity, payments, search, orders, trust | Complete planning before build | Owner |
| 2026-10-05 | "Backend never hits the DB directly" = data layer only + Docker network only | Clarifies the new `.claude/CLAUDE.md` rule | Owner |

## 12. Open questions

- Staff role list, commission rates and VAT treatment.
- Hosting, and the email, KYC and carrier providers.
- Return policy, the late-payment rule, approval thresholds.
- Approval of the proposed schema changes (`docs/schema-changes.md`).
