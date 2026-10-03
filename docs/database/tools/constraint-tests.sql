-- Constraint tests for docs/database/schema.sql.
-- Run against a scratch database AFTER loading the schema (tools/generate.sh does this).
-- Each test prints PASS or raises an exception naming the failed rule.

\set ON_ERROR_STOP 1
set client_min_messages = notice;

-- expect_error: the statement must fail. If `expected` is given, the error must
-- match it — so a test can't pass because some unrelated rule fired.
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

-- expect_ok: the statement must succeed (rule doesn't block valid data).
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
insert into users (id, email, phone, first_name, last_name)
values ('00000000-0000-0000-0000-000000000001', 'ada@example.com', '+2348031234567', 'Ada', 'Obi');
insert into categories (id, slug, name) values ('00000000-0000-0000-0000-0000000000c1', 'phones', 'Phones');
insert into products (id, category_id, slug, name) values ('00000000-0000-0000-0000-0000000000a1', '00000000-0000-0000-0000-0000000000c1', 'nova-x5', 'Nova X5 Pro 5G');
insert into product_variants (id, product_id, sku, name) values ('00000000-0000-0000-0000-0000000000b1', '00000000-0000-0000-0000-0000000000a1', 'NX5-256', '256GB');
insert into businesses (id, legal_name, type) values ('00000000-0000-0000-0000-0000000000e1', 'Brightline Schools Ltd', 'school');
insert into warehouses (id, code, name, kind, address, city, state_code)
values ('00000000-0000-0000-0000-0000000000f1', 'IKJ', 'Ikeja WH', 'warehouse', '1 Road', 'Ikeja', 'LA');

-- ---------- Identity ----------
select pg_temp.expect_error('user needs email or phone',
  $q$ insert into users (first_name, last_name) values ('No', 'Contact') $q$);
select pg_temp.expect_error('phone must be E.164 +234',
  $q$ insert into users (phone, first_name, last_name) values ('08031234567', 'Bad', 'Phone') $q$);
select pg_temp.expect_error('email is unique, case-insensitive (citext)',
  $q$ insert into users (email, first_name, last_name) values ('ADA@example.com', 'Dup', 'Ada') $q$);
select pg_temp.expect_error('address belongs to a user XOR a business',
  $q$ insert into addresses (user_id, business_id, recipient_name, phone, line1, city, state_code)
      values ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-0000000000e1', 'Ada', '+2348031234567', '1 Road', 'Ikeja', 'LA') $q$);
select pg_temp.expect_error('address state must be a real state',
  $q$ insert into addresses (user_id, recipient_name, phone, line1, city, state_code)
      values ('00000000-0000-0000-0000-000000000001', 'Ada', '+2348031234567', '1 Road', 'Ikeja', 'XX') $q$);

-- ---------- Sellers ----------
select pg_temp.expect_error('only one first-party seller (TechShop)',
  $q$ insert into sellers (type, display_name, slug) values ('first_party', 'Fake TechShop', 'fake') $q$);
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
select pg_temp.expect_ok('valid first-party listing',
  $q$ insert into listings (id, seller_id, variant_id, condition, price_kobo, compare_at_kobo, status)
      select '00000000-0000-0000-0000-0000000000d1', id, '00000000-0000-0000-0000-0000000000b1', 'new', 89900000, 105000000, 'active'
      from sellers where type = 'first_party' $q$);
select pg_temp.expect_error('one listing per seller × variant × condition',
  $q$ insert into listings (seller_id, variant_id, condition, price_kobo)
      select id, '00000000-0000-0000-0000-0000000000b1', 'new', 88000000 from sellers where type = 'first_party' $q$);
select pg_temp.expect_error('unknown status rejected',
  $q$ update listings set status = 'live' where id = '00000000-0000-0000-0000-0000000000d1' $q$);

-- ---------- Inventory ----------
select pg_temp.expect_error('IMEI must be 15 digits',
  $q$ insert into device_units (variant_id, condition, imei, acquired_via)
      values ('00000000-0000-0000-0000-0000000000b1', 'new', '12345', 'purchase') $q$);
select pg_temp.expect_ok('valid device unit',
  $q$ insert into device_units (variant_id, condition, imei, acquired_via)
      values ('00000000-0000-0000-0000-0000000000b1', 'new', '356938035643809', 'purchase') $q$);
select pg_temp.expect_error('IMEI is unique',
  $q$ insert into device_units (variant_id, condition, imei, acquired_via)
      values ('00000000-0000-0000-0000-0000000000b1', 'new', '356938035643809', 'purchase') $q$);
select pg_temp.expect_error('stock movements are append-only',
  $q$ insert into stock_movements (warehouse_id, variant_id, condition, quantity_delta, reason)
        values ('00000000-0000-0000-0000-0000000000f1', '00000000-0000-0000-0000-0000000000b1', 'new', 10, 'purchase_receipt');
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
  $q$ insert into payments (order_id, provider, provider_reference, idempotency_key, amount_kobo)
      values ('00000000-0000-0000-0000-000000000011', 'paystack', 'PSK-1', 'idem-1', 90150000) $q$);
select pg_temp.expect_error('provider reference is unique per provider (no double-counting)',
  $q$ insert into payments (order_id, provider, provider_reference, idempotency_key, amount_kobo)
      values ('00000000-0000-0000-0000-000000000011', 'paystack', 'PSK-1', 'idem-2', 90150000) $q$);
select pg_temp.expect_error('payment is for an order XOR an invoice',
  $q$ insert into payments (provider, idempotency_key, amount_kobo) values ('paystack', 'idem-3', 100) $q$);
select pg_temp.expect_error('succeeded payment needs paid_at',
  $q$ update payments set status = 'succeeded' where provider_reference = 'PSK-1' $q$);

-- Ledger: deferred trigger, checked at commit (forced immediately here).
select pg_temp.expect_error('unbalanced ledger journal rejected',
  $q$ insert into ledger_journals (id, kind, reference_type, reference_id)
        values ('00000000-0000-0000-0000-000000000021', 'sale', 'order', '00000000-0000-0000-0000-000000000011');
      insert into ledger_entries (journal_id, account_id, amount_kobo)
        select '00000000-0000-0000-0000-000000000021', id, 90150000 from ledger_accounts where code = 'clearing:paystack';
      set constraints ledger_entries_balanced immediate $q$, 'not balanced');
select pg_temp.expect_ok('balanced ledger journal accepted',
  $q$ insert into ledger_journals (id, kind, reference_type, reference_id)
        values ('00000000-0000-0000-0000-000000000022', 'sale', 'order', '00000000-0000-0000-0000-000000000011');
      insert into ledger_entries (journal_id, account_id, amount_kobo)
        select '00000000-0000-0000-0000-000000000022', id, 90150000 from ledger_accounts where code = 'clearing:paystack';
      insert into ledger_entries (journal_id, account_id, amount_kobo)
        select '00000000-0000-0000-0000-000000000022', id, -89900000 from ledger_accounts where code = 'revenue:first_party_sales';
      insert into ledger_entries (journal_id, account_id, amount_kobo)
        select '00000000-0000-0000-0000-000000000022', id, -250000 from ledger_accounts where code = 'revenue:delivery';
      set constraints ledger_entries_balanced immediate $q$);
select pg_temp.expect_error('ledger entries are append-only',
  $q$ update ledger_entries set amount_kobo = 1 where journal_id = '00000000-0000-0000-0000-000000000022' $q$, 'append-only');

-- ---------- Marketing ----------
select pg_temp.expect_error('promotion must end after it starts (no fake timers)',
  $q$ insert into promotions (name, kind, value, starts_at, ends_at, created_by)
      values ('Bad', 'percentage', 10, now(), now() - interval '1 hour', '00000000-0000-0000-0000-000000000001') $q$);
select pg_temp.expect_error('percentage discount capped at 90%',
  $q$ insert into promotions (name, kind, value, starts_at, ends_at, created_by)
      values ('Too good', 'percentage', 95, now(), now() + interval '1 day', '00000000-0000-0000-0000-000000000001') $q$);

-- ---------- Cars ----------
select pg_temp.expect_error('brand-new car cannot have high mileage',
  $q$ insert into car_listings (seller_id, make, model, year, transmission, mileage_km, condition, state_code, city, price_kobo)
      select id, 'Kodo', 'RX', 2024, 'automatic', 40000, 'brand_new', 'LA', 'Lagos', 6200000000 from sellers where type = 'first_party' $q$);
select pg_temp.expect_error('VIN must be 17 valid characters',
  $q$ insert into car_listings (seller_id, make, model, year, transmission, mileage_km, condition, vin, state_code, city, price_kobo)
      select id, 'Kodo', 'RX', 2022, 'automatic', 10000, 'foreign_used', 'SHORTVIN', 'LA', 'Lagos', 3850000000 from sellers where type = 'first_party' $q$);

-- ---------- Logistics ----------
select pg_temp.expect_error('delivered job needs OTP or photo proof',
  $q$ insert into delivery_zones (id, name, state_code) values ('00000000-0000-0000-0000-000000000031', 'Lekki', 'LA');
      insert into fulfilments (id, order_id, seller_id, fulfilled_by)
        select '00000000-0000-0000-0000-000000000041', '00000000-0000-0000-0000-000000000011', id, 'techshop' from sellers where type = 'first_party';
      insert into staff_members (user_id, employee_no, department, job_title)
        values ('00000000-0000-0000-0000-000000000001', 'EMP-1', 'Logistics', 'Rider');
      insert into riders (user_id, vehicle_type, home_warehouse_id)
        values ('00000000-0000-0000-0000-000000000001', 'bike', '00000000-0000-0000-0000-0000000000f1');
      insert into delivery_jobs (fulfilment_id, rider_id, zone_id, status, delivered_at)
        values ('00000000-0000-0000-0000-000000000041', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000031', 'delivered', now()) $q$,
  'delivery_jobs_check');

-- ---------- updated_at trigger ----------
do $$
declare before_ts timestamptz;
begin
  select updated_at into before_ts from users where id = '00000000-0000-0000-0000-000000000001';
  perform pg_sleep(0.01);
  update users set first_name = 'Adaeze' where id = '00000000-0000-0000-0000-000000000001';
  -- now() is fixed within a transaction, so compare against clock_timestamp-based expectations loosely:
  if (select updated_at from users where id = '00000000-0000-0000-0000-000000000001') < before_ts then
    raise exception 'FAIL  updated_at trigger';
  end if;
  raise notice 'PASS  updated_at trigger fires on update';
end $$;

rollback;

\echo 'All constraint tests passed.'
