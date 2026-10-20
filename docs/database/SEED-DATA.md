# Seed data

There are two kinds of starting data, and they are kept strictly apart.

## 1. Reference data (every environment, via migration)

`00022_reference_data` inserts what the app needs to work anywhere, including production:

- the 36 states and the FCT (`identity.nigerian_states`);
- **TechShop as the first-party seller** (`sellers.sellers`, `type = 'first_party'`);
- the chart of accounts for the ledger;
- payment provider health rows;
- open accounting periods;
- the admin role and default feature flags.

Not included yet (on purpose): LGAs, public holidays, and the full staff role set (owner decision).

## 2. Development seed (local only, via `cmd/seed`)

```bash
cd backend
go run ./cmd/seed > seed.json   # needs APP_ENV=development and SEED_PASSWORD
```

- **Refuses to run unless `APP_ENV=development`.** It can never create test accounts in staging or production.
- **Idempotent:** running it again updates rather than duplicates.
- **Credentials:** every account uses `SEED_PASSWORD` from your local `backend/.env`. Nothing is committed. The output JSON lists the accounts and IDs; treat it as a secret file (`chmod 600`) and don't commit it.
- Staff accounts are created with TOTP; the output includes what the test kits need to sign in.

### What it creates

| Area | Data | Source |
|---|---|---|
| **Accounts (one per role)** | Customers `customer`, `customer_b`; staff `admin`, `ops_manager`, `store_manager`, `warehouse_lead`, `dispatcher`, `cashier`, `catalog_editor`, `content_editor`, `marketing_manager`, `marketing_exec`, `finance_officer`, `finance_manager`, `finance_manager_b`, `auditor`, `support_agent`, `support_lead`, `aftersales`, `kyc_reviewer`, `risk_analyst`, `b2b_manager`, `car_sales`, `hr_officer`; `seller_owner`; `rider` | `seed.go` |
| **Catalogue** | 8 categories, 27 products with variants and listings, 4 cars, 6 wholesale products | `fixtures.json`, `commerce.go` |
| **Stock** | Ikeja warehouse and Wuse hub, with stock levels for every listing | `commerce.go` |
| **Sellers** | Sample marketplace vendors (Ikeja Gadget Hub, Wuse Tech Store, Sabon Gari Mobile and others) with their own listings; `seller_owner` owns one | `commerce.go` |
| **Checkout and marketing (M2)** | Delivery addresses, marketing consent, a welcome coupon, a live flash deal, a scheduled promotion, a Lagos audience, an abandoned-cart journey | `m2.go` |
| **Logistics (M3)** | Lagos and Abuja delivery zones with rate bands, the fake carrier, a rider with an approved phone | `m3.go` |

All product names, prices and people are **sample data**; the web apps label it as such.

### Adding seed data in a new milestone

Add a file `internal/seed/m<N>.go` that registers a hook:

```go
func init() { Hooks = append(Hooks, seedMarketplace) }
```

Keep it idempotent (upsert by a natural key), dev-only, and with no real personal data.

## Test stacks

The e2e tests (`backend/test/e2e`, `shared/api-client/test`) run against a throwaway Postgres that is migrated and seeded fresh each run. They read the seed output through `SEED_FILE`. The expiry test also needs `WORKER_IN_PROCESS=true` and `ORDER_PAYMENT_TTL=100s` on the API.
