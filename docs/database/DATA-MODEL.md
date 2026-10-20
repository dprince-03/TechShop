# Data model

How TechShop's data fits together, in plain words. For every table and column see [`SCHEMA.md`](SCHEMA.md); for the diagrams see [`erd.md`](erd.md); for status lifecycles see [`states.md`](states.md).

## The big ideas

1. **One marketplace model.** TechShop itself is a row in `sellers.sellers` (`type = 'first_party'`). Products are catalogue entries; a **listing** is one seller's offer (price, stock, condition) on one product variant. Own stock and vendor stock use the same code path.
2. **One order, many fulfilments.** A checkout creates one `sales.orders` row and **one `sales.fulfilments` row per seller**. `sales.order_lines` copy the name and price at the time of sale, so history never changes when a product does.
3. **Money is kobo, and every movement is double-entry.** Amounts are `bigint` kobo (₦1 = 100 kobo), never floats. Payments, commission, seller payables, payouts and refunds post balanced journals to `finance.ledger_journals` / `ledger_entries`; a trigger rejects any journal that doesn't sum to zero.
4. **Physical devices are tracked one by one.** `inventory.device_units` holds each IMEI or serial number, so warranty, returns, trade-ins and theft checks refer to the actual unit.
5. **Cars are separate.** `cars.*` tables follow a viewing → inspection → financing flow and never go in a cart.
6. **Personal data is protected in the application.** NIN, ID numbers and bank account numbers are encrypted by the API (`bytea`, envelope encryption); only `*_last4` is shown. Passwords, tokens and OTPs are stored as hashes only. Card data is never stored.

## Domains

Each domain is its own Postgres schema (22 in total, plus `river` for the job queue).

| Schema | Holds | Key tables |
|---|---|---|
| `identity` | People and access | `users`, `sessions`, `roles`, `staff_members`, `addresses`, `consents` |
| `platform` | Shared plumbing | `files`, `audit_log`, `outbox_events`, `idempotency_keys`, `encryption_keys` |
| `sellers` | Marketplace shops | `sellers`, `seller_members`, `seller_kyc_submissions`, `seller_bank_accounts` |
| `b2b` | Business buyers | `businesses`, `business_members`, `credit_accounts` |
| `catalog` | What we sell | `products`, `product_variants`, `listings`, `product_reviews`, `seller_ratings` |
| `search` | Search tuning and analytics | `search_synonyms`, `search_queries` |
| `inventory` | Stock | `warehouses`, `inventory_levels`, `device_units`, `stock_movements`, `stock_reservations` |
| `purchasing` | Buying from suppliers | `suppliers`, `purchase_orders` |
| `pos` | Physical stores | tills, cashier shifts |
| `sales` | Carts and orders | `carts`, `orders`, `fulfilments`, `order_lines`, `quotes` |
| `fulfilment` | Warehouse work and carriers | `pick_lists`, parcels, `shipments`, `seller_sla_events` |
| `logistics` | Own delivery | zones and rates, `riders`, `delivery_jobs`, `delivery_events` |
| `payments` | Money in | `payments`, webhooks, `refunds`, `disputes`, `virtual_accounts` |
| `finance` | Books | ledger, periods, `invoices`, commission, `payout_batches`, `payouts`, `payout_holds` |
| `aftersales` | After the sale | `return_requests`, `warranty_claims`, `repair_jobs`, `trade_ins` |
| `messaging` | SMS, email, push, inbox | message log, `notification_preferences`, `message_templates` |
| `marketing` | Promotions and campaigns | `promotions`, coupons, `segments`, `campaigns`, `journeys` |
| `content`, `support` | CMS and tickets | `cms_pages`, `banners`, `help_articles`, `support_tickets` |
| `risk` | Trust and safety | rules, decisions, `risk_cases`, `kyc_checks`, link graph, block lists |
| `cars` | Car sales | `car_listings`, inspections, `viewing_bookings` |
| `personalisation` | Behaviour and recommendations | events (with consent), `rec_*` results pulled from the recommender |

## How the main things connect

- A **user** can be a customer, a member of one or more **sellers**, a member of a **business**, and/or a **staff member**. Roles decide what they may do ([`access.md`](access.md)).
- A **product** has **variants** (e.g. 128 GB / Black); each variant has **listings** from one or more sellers. The cheapest eligible listing wins the buy box (`catalog.product_offer_summary`).
- A **cart** holds listings. **Checkout** reserves stock (`inventory.stock_reservations`), creates the **order**, its **fulfilments** and **lines**, and a **payment**.
- When the payment succeeds, the ledger records the sale; each fulfilment is picked, packed and handed to a **rider** (`logistics.delivery_jobs`) or a **carrier** (`fulfilment.shipments`).
- On delivery, commission is booked and the seller's share becomes payable. **Payouts** pay it out in approved batches.
- **Returns, warranty and trade-ins** point back to the order line and, for devices, the exact `device_unit`.
- Important actions write `platform.audit_log`; side effects (SMS, email, push, webhooks) go through `platform.outbox_events` in the same transaction, so they are never lost or sent twice.

## Rules the database enforces itself

- Foreign keys, `NOT NULL`, `UNIQUE` and `CHECK` constraints on every table; statuses are `text` with `CHECK (… IN (…))`.
- Balanced ledger journals and closed accounting periods (`finance.assert_journal_balanced`, `finance.assert_period_open`).
- Append-only history: ledger, stock movements, delivery events (`platform.forbid_change`).
- `updated_at` maintained by a trigger on every table that has it.
- Separation of duties where it matters (e.g. nobody grants a role to themselves).
