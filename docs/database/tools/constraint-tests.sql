-- Constraint tests for the migrated database (backend/db/migrations, one Postgres schema per domain).
-- Run against a scratch database AFTER `goose up` (never the project's own database).
-- Each test prints PASS or raises an exception naming the failed rule. Everything is rolled back.

\set ON_ERROR_STOP 1
set client_min_messages = notice;
set search_path = identity, platform, sellers, b2b, catalog, search, inventory, purchasing, pos, sales,
  logistics, fulfilment, payments, finance, aftersales, messaging, marketing, content, support, risk, cars,
  personalisation, public;

-- expect_error: the statement must fail. If `expected` is given, the error must match it, so a
-- test can't pass because some unrelated rule fired.
create function pg_temp.expect_error(label text, stmt text, expected text default null) returns void
language plpgsql as $$
declare msg text;
begin
  begin
    execute stmt;
  exception when others then
    msg := sqlerrm;
    if expected is not null and msg !~* expected then
      raise exception 'FAIL  % — rejected for the wrong reason: %', label, msg;
    end if;
    raise notice 'PASS  %  (rejected: %)', label, msg;
    return;
  end;
  raise exception 'FAIL  % — statement was accepted but should be rejected', label;
end $$;

-- expect_ok: the statement must succeed (the rule doesn't block valid data).
create function pg_temp.expect_ok(label text, stmt text) returns void
language plpgsql as $$
begin
  execute stmt;
  raise notice 'PASS  %', label;
exception when others then
  raise exception 'FAIL  % — valid statement rejected: %', label, sqlerrm;
end $$;

begin;

-- ---------- Fixtures ----------
insert into users (id, email, phone, first_name, last_name) values
  ('00000000-0000-0000-0000-000000000001', 'ada@example.com', '+2348031234567', 'Ada', 'Obi'),
  ('00000000-0000-0000-0000-000000000002', 'tolu@example.com', '+2348031234568', 'Tolu', 'Ade'),
  ('00000000-0000-0000-0000-000000000003', 'ngozi@example.com', '+2348031234569', 'Ngozi', 'Eze');
insert into categories (id, slug, name) values ('00000000-0000-0000-0000-0000000000c1', 'phones', 'Phones');
insert into products (id, category_id, slug, name) values ('00000000-0000-0000-0000-0000000000a1', '00000000-0000-0000-0000-0000000000c1', 'nova-x5', 'Nova X5 Pro 5G');
insert into product_variants (id, product_id, sku, name) values ('00000000-0000-0000-0000-0000000000b1', '00000000-0000-0000-0000-0000000000a1', 'NX5-256', '256GB');
insert into businesses (id, legal_name, type) values ('00000000-0000-0000-0000-0000000000e1', 'Brightline Schools Ltd', 'school');
insert into warehouses (id, code, name, kind, address, city, state_code)
values ('00000000-0000-0000-0000-0000000000f1', 'IKJ', 'Ikeja WH', 'warehouse', '1 Road', 'Ikeja', 'LA');

-- ---------- Identity ----------
select pg_temp.expect_error('user needs email or phone',
  $q$ insert into users (first_name, last_name) values ('No', 'Contact') $q$);
select pg_temp.expect_ok('a deleted (anonymised) user may have no email or phone',
  $q$ insert into users (first_name, last_name, status) values ('Deleted', 'user', 'deleted') $q$);
select pg_temp.expect_error('phone must be E.164 +234',
  $q$ insert into users (phone, first_name, last_name) values ('08031234567', 'Bad', 'Phone') $q$);
select pg_temp.expect_error('email is unique, case-insensitive (citext)',
  $q$ insert into users (email, first_name, last_name) values ('ADA@example.com', 'Dup', 'Ada') $q$);
select pg_temp.expect_error('delivery codes no longer live in verification_codes',
  $q$ insert into verification_codes (channel, destination, purpose, code_hash, expires_at)
      values ('sms', '+2348031234567', 'delivery_otp', '\x00', now() + interval '10 minutes') $q$, 'verification_codes_purpose_check');
select pg_temp.expect_error('address belongs to a user XOR a business',
  $q$ insert into addresses (user_id, business_id, recipient_name, phone, line1, city, state_code)
      values ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-0000000000e1', 'Ada', '+2348031234567', '1 Road', 'Ikeja', 'LA') $q$);
select pg_temp.expect_error('address state must be a real state',
  $q$ insert into addresses (user_id, recipient_name, phone, line1, city, state_code)
      values ('00000000-0000-0000-0000-000000000001', 'Ada', '+2348031234567', '1 Road', 'Ikeja', 'XX') $q$);
select pg_temp.expect_error('nobody grants a role to themselves',
  $q$ insert into staff_members (user_id, employee_no, department, job_title)
        values ('00000000-0000-0000-0000-000000000002', 'EMP-2', 'Finance', 'Officer');
      insert into staff_roles (user_id, role_id, granted_by)
        select '00000000-0000-0000-0000-000000000002', id, '00000000-0000-0000-0000-000000000002' from roles where key = 'admin' $q$,
  'staff_roles_check');

-- ---------- Sellers ----------
select pg_temp.expect_error('only one first-party seller (TechShop)',
  $q$ insert into sellers (type, display_name, slug) values ('first_party', 'Fake TechShop', 'fake') $q$, 'sellers_one_first_party');
select pg_temp.expect_error('marketplace sellers need an owner',
  $q$ insert into sellers (type, display_name, slug) values ('business', 'No Owner Ltd', 'no-owner') $q$);

-- ---------- Catalogue ----------
select pg_temp.expect_error('listing price must be positive',
  $q$ insert into listings (seller_id, variant_id, condition, price_kobo)
      select id, '00000000-0000-0000-0000-0000000000b1', 'new', 0 from sellers where type = 'first_party' $q$);
select pg_temp.expect_error('compare-at price must be higher than price (no fake discounts)',
  $q$ insert into listings (seller_id, variant_id, condition, price_kobo, compare_at_kobo)
      select id, '00000000-0000-0000-0000-0000000000b1', 'new', 89900000, 89900000 from sellers where type = 'first_party' $q$);
select pg_temp.expect_error('used listings must describe their condition',
  $q$ insert into listings (seller_id, variant_id, condition, price_kobo)
      select id, '00000000-0000-0000-0000-0000000000b1', 'uk_used', 64000000 from sellers where type = 'first_party' $q$);
select pg_temp.expect_error('an active listing needs published_at',
  $q$ insert into listings (seller_id, variant_id, condition, price_kobo, status)
      select id, '00000000-0000-0000-0000-0000000000b1', 'new', 89900000, 'active' from sellers where type = 'first_party' $q$);
select pg_temp.expect_ok('valid first-party listing',
  $q$ insert into listings (id, seller_id, variant_id, condition, price_kobo, compare_at_kobo, status, published_at)
      select '00000000-0000-0000-0000-0000000000d1', id, '00000000-0000-0000-0000-0000000000b1', 'new', 89900000, 105000000, 'active', now()
      from sellers where type = 'first_party' $q$);
select pg_temp.expect_error('one listing per seller × variant × condition',
  $q$ insert into listings (seller_id, variant_id, condition, price_kobo)
      select id, '00000000-0000-0000-0000-0000000000b1', 'new', 88000000 from sellers where type = 'first_party' $q$);
select pg_temp.expect_error('unknown status rejected',
  $q$ update listings set status = 'live' where id = '00000000-0000-0000-0000-0000000000d1' $q$);

-- ---------- Inventory ----------
select pg_temp.expect_error('IMEI must be 15 digits',
  $q$ insert into device_units (variant_id, condition, imei, acquired_via, owner_seller_id)
      select '00000000-0000-0000-0000-0000000000b1', 'new', '12345', 'purchase', id from sellers where type = 'first_party' $q$);
select pg_temp.expect_ok('valid device unit',
  $q$ insert into device_units (variant_id, condition, imei, acquired_via, owner_seller_id)
      select '00000000-0000-0000-0000-0000000000b1', 'new', '356938035643809', 'purchase', id from sellers where type = 'first_party' $q$);
select pg_temp.expect_error('IMEI is unique',
  $q$ insert into device_units (variant_id, condition, imei, acquired_via, owner_seller_id)
      select '00000000-0000-0000-0000-0000000000b1', 'new', '356938035643809', 'purchase', id from sellers where type = 'first_party' $q$);
select pg_temp.expect_error('reserved stock can never exceed on hand',
  $q$ insert into inventory_levels (warehouse_id, variant_id, condition, owner_seller_id, on_hand, reserved)
      select '00000000-0000-0000-0000-0000000000f1', '00000000-0000-0000-0000-0000000000b1', 'new', id, 1, 2 from sellers where type = 'first_party' $q$);
select pg_temp.expect_error('stock movements are append-only',
  $q$ insert into stock_movements (warehouse_id, variant_id, condition, owner_seller_id, quantity_delta, reason)
        select '00000000-0000-0000-0000-0000000000f1', '00000000-0000-0000-0000-0000000000b1', 'new', id, 10, 'purchase_receipt'
        from sellers where type = 'first_party';
      delete from stock_movements $q$, 'append-only');

-- ---------- Sales ----------
select pg_temp.expect_error('order total must equal subtotal + delivery - discount + VAT',
  $q$ insert into orders (channel, customer_user_id, subtotal_kobo, delivery_fee_kobo, total_kobo, ship_to)
      values ('market', '00000000-0000-0000-0000-000000000001', 89900000, 250000, 89900000, '{}') $q$, 'orders_check');
select pg_temp.expect_ok('valid order gets a TS- number',
  $q$ insert into orders (id, channel, customer_user_id, subtotal_kobo, delivery_fee_kobo, total_kobo, ship_to)
      values ('00000000-0000-0000-0000-000000000011', 'market', '00000000-0000-0000-0000-000000000001', 89900000, 250000, 90150000, '{}') $q$);
select pg_temp.expect_error('online orders need a delivery address snapshot',
  $q$ insert into orders (channel, customer_user_id, subtotal_kobo, total_kobo)
      values ('market', '00000000-0000-0000-0000-000000000001', 100, 100) $q$);

-- ---------- Payments ----------
select pg_temp.expect_ok('valid payment',
  $q$ insert into payments (id, order_id, provider, provider_reference, idempotency_key, amount_kobo)
      values ('00000000-0000-0000-0000-000000000051', '00000000-0000-0000-0000-000000000011', 'paystack', 'PSK-1', 'idem-1', 90150000) $q$);
select pg_temp.expect_error('provider reference is unique per provider (no double-counting)',
  $q$ insert into payments (order_id, provider, provider_reference, idempotency_key, amount_kobo)
      values ('00000000-0000-0000-0000-000000000011', 'paystack', 'PSK-1', 'idem-2', 90150000) $q$);
select pg_temp.expect_error('payment is for an order XOR an invoice',
  $q$ insert into payments (provider, idempotency_key, amount_kobo) values ('paystack', 'idem-3', 100) $q$);
select pg_temp.expect_error('succeeded payment needs paid_at',
  $q$ update payments set status = 'succeeded' where provider_reference = 'PSK-1' $q$);
select pg_temp.expect_error('refund requester cannot approve it (separation of duties)',
  $q$ insert into refunds (payment_id, order_id, amount_kobo, reason, status, requested_by, approved_by)
      values ('00000000-0000-0000-0000-000000000051', '00000000-0000-0000-0000-000000000011', 1000, 'Faulty', 'approved',
              '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000002') $q$, 'refunds_check');
select pg_temp.expect_ok('refund approved by a different person',
  $q$ insert into refunds (payment_id, order_id, amount_kobo, reason, status, requested_by, approved_by)
      values ('00000000-0000-0000-0000-000000000051', '00000000-0000-0000-0000-000000000011', 1000, 'Faulty', 'approved',
              '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000003') $q$);

-- ---------- Ledger ----------
select pg_temp.expect_error('unbalanced ledger journal rejected',
  $q$ insert into ledger_journals (id, posting_key, kind, reference_type, reference_id)
        values ('00000000-0000-0000-0000-000000000021', 'test:unbalanced', 'sale', 'order', '00000000-0000-0000-0000-000000000011');
      insert into ledger_entries (journal_id, account_id, amount_kobo)
        select '00000000-0000-0000-0000-000000000021', id, 90150000 from ledger_accounts where code = 'clearing:paystack';
      set constraints finance.ledger_entries_balanced immediate $q$, 'not balanced');
select pg_temp.expect_ok('balanced ledger journal accepted',
  $q$ insert into ledger_journals (id, posting_key, kind, reference_type, reference_id)
        values ('00000000-0000-0000-0000-000000000022', 'payment:PSK-1:captured', 'sale', 'order', '00000000-0000-0000-0000-000000000011');
      insert into ledger_entries (journal_id, account_id, amount_kobo)
        select '00000000-0000-0000-0000-000000000022', id, 90150000 from ledger_accounts where code = 'clearing:paystack';
      insert into ledger_entries (journal_id, account_id, amount_kobo)
        select '00000000-0000-0000-0000-000000000022', id, -89900000 from ledger_accounts where code = 'revenue:first_party_sales';
      insert into ledger_entries (journal_id, account_id, amount_kobo)
        select '00000000-0000-0000-0000-000000000022', id, -250000 from ledger_accounts where code = 'revenue:delivery';
      set constraints finance.ledger_entries_balanced immediate $q$);
select pg_temp.expect_error('a money event can never be posted twice (unique posting_key)',
  $q$ insert into ledger_journals (posting_key, kind, reference_type, reference_id)
      values ('payment:PSK-1:captured', 'sale', 'order', '00000000-0000-0000-0000-000000000011') $q$, 'posting_key');
select pg_temp.expect_error('ledger entries are append-only',
  $q$ update ledger_entries set amount_kobo = 1 where journal_id = '00000000-0000-0000-0000-000000000022' $q$, 'append-only');
select pg_temp.expect_error('ledger journals are append-only (reverse, never edit)',
  $q$ update ledger_journals set memo = 'edited' where id = '00000000-0000-0000-0000-000000000022' $q$, 'append-only');
select pg_temp.expect_error('closed accounting period rejects postings',
  $q$ insert into accounting_periods (month, status, closed_by, closed_at)
        values ('2026-01-01', 'closed', '00000000-0000-0000-0000-000000000003', now());
      insert into ledger_journals (posting_key, kind, reference_type, reference_id, period)
        values ('test:closed', 'adjustment', 'order', '00000000-0000-0000-0000-000000000011', '2026-01-01') $q$, 'is closed');

-- ---------- Payouts ----------
select pg_temp.expect_error('payout batch preparer cannot approve it (dual control)',
  $q$ insert into payout_batches (status, prepared_by, approved_by, approved_at)
      values ('approved', '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000002', now()) $q$, 'payout_batches_check');
select pg_temp.expect_ok('payout batch approved by a different person',
  $q$ insert into payout_batches (status, prepared_by, approved_by, approved_at)
      values ('approved', '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000003', now()) $q$);

-- ---------- Marketing ----------
select pg_temp.expect_error('promotion must end after it starts (no fake timers)',
  $q$ insert into promotions (name, kind, value, starts_at, ends_at, created_by)
      values ('Bad', 'percentage', 10, now(), now() - interval '1 hour', '00000000-0000-0000-0000-000000000001') $q$);
select pg_temp.expect_error('percentage discount capped at 90%',
  $q$ insert into promotions (name, kind, value, starts_at, ends_at, created_by)
      values ('Too good', 'percentage', 95, now(), now() + interval '1 day', '00000000-0000-0000-0000-000000000001') $q$);
select pg_temp.expect_error('flash deal claims can never exceed the stock limit',
  $q$ insert into promotions (id, name, kind, starts_at, ends_at, created_by)
        values ('00000000-0000-0000-0000-000000000061', 'Flash', 'flash_price', now(), now() + interval '1 day', '00000000-0000-0000-0000-000000000001');
      insert into promotion_listing_prices (promotion_id, listing_id, price_kobo, stock_limit)
        values ('00000000-0000-0000-0000-000000000061', '00000000-0000-0000-0000-0000000000d1', 79900000, 2);
      update promotion_listing_prices set claimed = 3 where promotion_id = '00000000-0000-0000-0000-000000000061' $q$,
  'promotion_listing_prices_check');
select pg_temp.expect_error('campaign creator cannot approve it',
  $q$ insert into segments (id, name, rules, created_by)
        values ('00000000-0000-0000-0000-000000000071', 'Everyone', '{}', '00000000-0000-0000-0000-000000000002');
      insert into campaigns (name, segment_id, channels, created_by, approved_by, approved_at, status)
        values ('Gaming week', '00000000-0000-0000-0000-000000000071', '{email}', '00000000-0000-0000-0000-000000000002',
                '00000000-0000-0000-0000-000000000002', now(), 'sending') $q$, 'campaigns_check');
select pg_temp.expect_error('tracked links only point to techshop.ng (no open redirect)',
  $q$ insert into tracked_links (url) values ('https://evil.example.com/') $q$, 'tracked_links_url_check');

-- ---------- Cars ----------
select pg_temp.expect_error('brand-new car cannot have high mileage',
  $q$ insert into car_listings (slug, seller_id, make, model, year, transmission, mileage_km, condition, state_code, city, price_kobo)
      select 'kodo-rx', id, 'Kodo', 'RX', 2024, 'automatic', 40000, 'brand_new', 'LA', 'Lagos', 6200000000 from sellers where type = 'first_party' $q$);
select pg_temp.expect_error('VIN must be 17 valid characters',
  $q$ insert into car_listings (slug, seller_id, make, model, year, transmission, mileage_km, condition, vin, state_code, city, price_kobo)
      select 'kodo-rx-2', id, 'Kodo', 'RX', 2022, 'automatic', 10000, 'foreign_used', 'SHORTVIN', 'LA', 'Lagos', 3850000000 from sellers where type = 'first_party' $q$);

-- ---------- Logistics ----------
insert into delivery_zones (id, name, state_code) values ('00000000-0000-0000-0000-000000000031', 'Lekki', 'LA');
insert into fulfilments (id, order_id, seller_id, fulfilled_by)
  select '00000000-0000-0000-0000-000000000041', '00000000-0000-0000-0000-000000000011', id, 'techshop' from sellers where type = 'first_party';
insert into staff_members (user_id, employee_no, department, job_title)
  values ('00000000-0000-0000-0000-000000000001', 'EMP-1', 'Logistics', 'Rider');
insert into riders (user_id, vehicle_type, home_warehouse_id)
  values ('00000000-0000-0000-0000-000000000001', 'bike', '00000000-0000-0000-0000-0000000000f1');
select pg_temp.expect_error('delivered job needs a verified code or photo proof (an OTP hash alone is not enough)',
  $q$ insert into delivery_jobs (fulfilment_id, rider_id, zone_id, pickup_warehouse_id, status, delivered_at, otp_hash, proof_method)
      values ('00000000-0000-0000-0000-000000000041', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000031',
              '00000000-0000-0000-0000-0000000000f1', 'delivered', now(), '\x01', 'otp') $q$, 'delivery_jobs_check');
select pg_temp.expect_ok('delivered job with a verified code is accepted',
  $q$ insert into delivery_jobs (fulfilment_id, rider_id, zone_id, pickup_warehouse_id, status, delivered_at, otp_hash, proof_method, otp_verified_at)
      values ('00000000-0000-0000-0000-000000000041', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000031',
              '00000000-0000-0000-0000-0000000000f1', 'delivered', now(), '\x01', 'otp', now()) $q$);

-- ---------- updated_at trigger ----------
do $$
declare before_ts timestamptz;
begin
  select updated_at into before_ts from users where id = '00000000-0000-0000-0000-000000000001';
  update users set first_name = 'Adaeze' where id = '00000000-0000-0000-0000-000000000001';
  if (select updated_at from users where id = '00000000-0000-0000-0000-000000000001') < before_ts then
    raise exception 'FAIL  updated_at trigger';
  end if;
  raise notice 'PASS  updated_at trigger fires on update';
end $$;

rollback;

\echo 'All constraint tests passed.'
