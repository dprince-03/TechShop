# TechShop data flows

How the key processes move through the tables in [`schema.sql`](schema.sql). Every flow runs inside the Go API, the only service that touches the database. Status lifecycles are in [`states.md`](states.md).

Rules that apply to every flow:
- **One transaction per step.** The data change and its `outbox_events` row are written together, so SMS, email and webhooks are sent if and only if the change committed.
- **Retries are safe.** Client POSTs carry an idempotency key (`idempotency_keys`), and provider webhooks are unique per `(provider, provider_event_id)` (`payment_events`).
- **Money is posted to the ledger** as balanced journals ([money flow](#money-flow)).
- **Staff and seller actions are written to `audit_log`.**

## 1. Checkout, payment and webhook

```mermaid
sequenceDiagram
  autonumber
  actor C as Customer
  participant W as Market / app
  participant API as Go API
  participant DB as PostgreSQL
  participant P as Paystack / OPay / Moniepoint

  C->>W: Checkout (address, delivery option)
  W->>API: POST /orders (Idempotency-Key)
  API->>DB: BEGIN
  API->>DB: insert orders (pending_payment) + fulfilments (one per seller) + order_lines (snapshots, commission_bps)
  API->>DB: reserve stock: inventory_levels.reserved += qty, device_units → reserved
  API->>DB: insert payments (initiated, idempotency_key)
  API->>DB: COMMIT
  API->>P: initialise transaction (amount_kobo, reference)
  API-->>W: payment URL
  W->>P: Customer pays
  P-->>API: webhook (signed)
  API->>DB: insert payment_events (unique provider_event_id) — duplicates ignored
  API->>P: verify transaction (never trust the webhook alone)
  API->>DB: BEGIN
  API->>DB: payments → succeeded (paid_at), orders → paid, order_status_history
  API->>DB: ledger journal "sale" (balanced, see templates)
  API->>DB: stock_movements (sale), inventory_levels on_hand −= qty, reserved −= qty
  API->>DB: outbox_events (order.paid → SMS + email)
  API->>DB: COMMIT
  API-->>W: order confirmed (TS-10482)
```

If payment fails or expires: `payments → failed/cancelled`, `orders → cancelled`, reservations released. No ledger entries are written, because no money moved.

## 2. Multi-seller order: fulfilment and delivery

```mermaid
sequenceDiagram
  autonumber
  participant API as Go API
  participant DB as PostgreSQL
  actor S as Seller / warehouse
  actor D as Dispatch (staff)
  actor R as Rider (logistics app)
  actor C as Customer

  API->>DB: order paid → each fulfilment pending
  S->>API: mark packed (Seller Centre / warehouse)
  API->>DB: fulfilments → packed
  D->>API: assign rider for zone
  API->>DB: delivery_jobs (assigned, rider_id, window) + delivery_events (assigned)
  API->>DB: verification_codes (purpose delivery_otp) → outbox: SMS OTP to customer
  R->>API: picked up / en route (with location)
  API->>DB: delivery_jobs → picked_up → en_route, delivery_events (location)
  R->>C: hand over parcel
  C-->>R: reads out OTP
  R->>API: confirm delivery (OTP or photo)
  API->>DB: check OTP hash, delivery_jobs → delivered (proof required by CHECK)
  API->>DB: fulfilments → delivered (delivered_at), orders → delivered when all fulfilments delivered
  API->>DB: ledger journal "commission" for marketplace lines
```

## 3. Seller earnings and payout

```mermaid
sequenceDiagram
  autonumber
  participant J as Payout job (daily)
  participant DB as PostgreSQL
  participant B as Bank / transfer provider
  actor S as Seller

  J->>DB: find fulfilments delivered, past return window, not in payout_items, no open returns/risk cases
  J->>DB: sum seller_payable per seller from ledger
  J->>DB: insert payouts (scheduled, default seller_bank_accounts) + payout_items (one per fulfilment, unique)
  J->>B: transfer amount_kobo
  J->>DB: payouts → processing (provider_reference)
  B-->>J: transfer succeeded
  J->>DB: payouts → paid (paid_at), ledger journal "payout"
  J->>DB: outbox: payout SMS + email to seller
  B-->>J: transfer failed
  J->>DB: payouts → failed, outbox: ask seller to fix bank details
```

## 4. Return and refund

```mermaid
sequenceDiagram
  autonumber
  actor C as Customer
  participant API as Go API
  participant DB as PostgreSQL
  actor Ops as Returns staff
  participant P as Payment provider

  C->>API: start return (order, lines, reason)
  API->>DB: return_requests (requested, RMA-…) + return_items
  Ops->>API: approve
  API->>DB: return_requests → approved → awaiting_pickup (delivery job for pickup)
  Ops->>API: item received and inspected
  API->>DB: return_requests → received → inspected, device_units → returned
  API->>DB: stock_movements (return) if resaleable
  Ops->>API: resolve as refund
  API->>DB: refunds (requested → approved, approved_by)
  API->>DB: ledger journal "refund" (reverse sale / seller payable)
  API->>P: refund payment (provider reference)
  P-->>API: refund confirmed
  API->>DB: refunds → succeeded, return_requests → refunded → closed, orders → partially_refunded / refunded
```

## 5. B2B: quote, order, invoice on credit, payment

```mermaid
sequenceDiagram
  autonumber
  actor B as Business buyer
  participant API as Go API
  participant DB as PostgreSQL
  actor AM as Account manager

  B->>API: request quote (items, delivery state, VAT invoice, credit)
  API->>DB: quotes (requested, QT-…) + quote_lines
  AM->>API: price the quote
  API->>DB: quote_lines.unit_price_kobo, quotes → sent (total, valid_until)
  B->>API: accept
  API->>DB: orders (channel wholesale, quote_id unique) + fulfilments + order_lines, quotes → accepted
  API->>DB: check credit_accounts (approved, outstanding + total ≤ limit)
  API->>DB: invoices (issued, INV-…, due_at = issued_at + term_days)
  API->>DB: ledger journal: receivable:b2b ↔ revenue + VAT
  B->>API: pays by bank transfer before due date
  API->>DB: payments (provider bank_transfer, invoice_id), invoices → paid
  API->>DB: ledger journal: bank ↔ receivable:b2b
```

## 6. Seller onboarding and KYC

```mermaid
sequenceDiagram
  autonumber
  actor S as New seller
  participant API as Go API
  participant DB as PostgreSQL
  actor K as KYC reviewer

  S->>API: register (Seller Centre form)
  API->>DB: users (+ verification_codes for phone), sellers (pending_kyc), seller_members (owner)
  API->>DB: files (ID, CAC), seller_kyc_submissions (NIN encrypted, last4), kyc_documents
  API->>DB: seller_bank_accounts (NUBAN encrypted), ledger_accounts (seller payable)
  K->>API: review
  API->>DB: seller_kyc_submissions → in_review → approved (reviewed_by, reviewed_at), audit_log
  API->>DB: sellers → active, outbox: welcome email
  K->>API: or request more info
  API->>DB: → needs_more_info, outbox: tell seller what's missing
```

## 7. Trade-in

```mermaid
sequenceDiagram
  autonumber
  actor C as Customer
  participant API as Go API
  participant DB as PostgreSQL
  actor I as Inspector

  C->>API: trade-in form (device, condition, IMEI)
  API->>DB: device_blocklist check (IMEI), trade_ins (submitted, TI-…)
  API->>DB: trade_ins → estimated (estimate_kobo)
  C->>API: proceed, drop off device
  API->>DB: trade_ins → awaiting_inspection
  I->>API: inspection result
  API->>DB: trade_ins → inspected (inspected_condition) → offer_sent (offer_kobo)
  C->>API: accept offer
  API->>DB: device_units (acquired_via trade_in), trade_ins.device_unit_id, stock_movements (trade_in)
  API->>DB: pay to bank or store credit, trade_ins → paid
```

## Money flow

Accounts are rows in `ledger_accounts`. Every journal's entries sum to zero; positive amounts are debits, negative amounts are credits.

```mermaid
flowchart LR
  CUST([Customer]) -->|pays| CLR[clearing:paystack / opay / moniepoint]
  BIZ([Business on credit]) -->|invoice| REC[receivable:b2b]
  CLR -->|provider settles| BANK[bank:operating]
  REC -->|transfer| BANK
  CLR -->|vendor items| SP[seller_payable:‹seller›]
  CLR -->|own stock| R1[revenue:first_party_sales]
  CLR -->|delivery fees| R2[revenue:delivery]
  CLR -->|VAT on own sales| VAT[liability:vat]
  SP -->|commission on delivery| R3[revenue:commission]
  SP -->|payout| BANK
  CLR -->|provider fees| FEE[expense:payment_fees]
  R1 & SP -->|refund approved| REF[liability:customer_refunds]
  REF -->|refund paid| CLR
```

### Journal templates

Worked examples in naira. Amounts are stored in kobo. **The VAT treatment is a placeholder until finance confirms it.**

| Journal (`kind`) | When | Debit (+) | Credit (−) |
|---|---|---|---|
| `sale` (marketplace item) | Payment succeeds: ₦100,000 vendor item + ₦2,500 delivery | clearing:paystack ₦102,500 | seller_payable:vendor ₦100,000 · revenue:delivery ₦2,500 |
| `sale` (TechShop stock, VAT-inclusive) | Payment succeeds: ₦107,500 incl. 7.5% VAT | clearing:paystack ₦107,500 | revenue:first_party_sales ₦100,000 · liability:vat ₦7,500 |
| `commission` | Fulfilment delivered: 10% on ₦100,000 | seller_payable:vendor ₦10,000 | revenue:commission ₦10,000 |
| `adjustment` (provider fee) | Provider settlement report | expense:payment_fees ₦1,537 | clearing:paystack ₦1,537 |
| `adjustment` (settlement) | Provider pays out to the bank | bank:operating ₦100,963 | clearing:paystack ₦100,963 |
| `payout` | Seller transfer succeeds | seller_payable:vendor ₦90,000 | bank:operating ₦90,000 |
| `refund` (approve) | Return approved, vendor item | seller_payable:vendor ₦100,000 | liability:customer_refunds ₦100,000 |
| `refund` (pay) | Provider refund confirmed | liability:customer_refunds ₦100,000 | clearing:paystack ₦100,000 |
| `sale` (B2B on credit) | Invoice issued: ₦8,480,000 + 7.5% VAT | receivable:b2b ₦9,116,000 | revenue:first_party_sales ₦8,480,000 · liability:vat ₦636,000 |
| `adjustment` (B2B paid) | Business pays invoice | bank:operating ₦9,116,000 | receivable:b2b ₦9,116,000 |
| `chargeback` | Customer disputes a card payment | revenue:first_party_sales or seller_payable ₦X | clearing:paystack ₦X |

**Balances are derived from the entries:**
- A seller's balance is the sum of their `seller_payable` entries.
- Provider clearing balances must match the providers' settlement reports. That's the Finance module's reconciliation tab.
