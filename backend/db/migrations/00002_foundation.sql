-- Foundation: extensions, one Postgres schema per domain, and shared helper functions.
-- Every later migration builds on this. Design notes: docs/database.md and docs/backend.md §4.1.

-- +goose Up
create extension if not exists citext;
create extension if not exists pg_trgm;

create schema platform;        comment on schema platform        is 'Shared building blocks: files, audit log, outbox, idempotency, feature flags, helper functions.';
create schema identity;        comment on schema identity        is 'People and access: users, sessions, codes, MFA, roles, staff, addresses, consents.';
create schema sellers;         comment on schema sellers         is 'Marketplace sellers: shops, members, KYC, payout bank accounts.';
create schema b2b;             comment on schema b2b             is 'Business buyers: organisations, members, credit, quotes.';
create schema catalog;         comment on schema catalog         is 'Products, variants, listings, prices, reviews and the search read model.';
create schema search;          comment on schema search          is 'Search merchandising and analytics: synonyms, redirects, pins, query logs.';
create schema inventory;       comment on schema inventory       is 'Stock: warehouses, levels, device units, movements, reservations, counts, costs.';
create schema purchasing;      comment on schema purchasing      is 'Buying stock from suppliers: purchase orders and goods receipts.';
create schema pos;             comment on schema pos             is 'Physical stores: tills and cashier shifts.';
create schema sales;           comment on schema sales           is 'Carts, orders, fulfilments and order lines.';
create schema logistics;       comment on schema logistics       is 'Delivery zones and rates, riders, delivery jobs and tracking.';
create schema fulfilment;      comment on schema fulfilment      is 'Warehouse work and shipping: pick lists, packing, parcels, manifests, carriers, seller SLAs.';
create schema payments;        comment on schema payments        is 'Taking money: payments, provider webhooks, refunds, transfer accounts, disputes, reconciliation.';
create schema finance;         comment on schema finance         is 'The books and paying out: double-entry ledger, periods, commission, invoices, seller payouts.';
create schema aftersales;      comment on schema aftersales      is 'Returns, inspections, warranty claims, repairs and trade-ins.';
create schema messaging;       comment on schema messaging       is 'Every message sent: notifications log, provider events, suppressions, preferences, templates, push tokens.';
create schema marketing;       comment on schema marketing       is 'Promotions, coupons, flash deals, segments, campaigns, journeys and tracked links.';
create schema content;         comment on schema content         is 'CMS pages, banners and help articles.';
create schema support;         comment on schema support         is 'Support tickets and messages.';
create schema risk;            comment on schema risk            is 'Trust and safety: KYC checks, risk rules and decisions, cases, link graph, blocklists, seller enforcement.';
create schema cars;            comment on schema cars            is 'Car listings, inspections, documents, viewings and financing enquiries.';
create schema personalisation; comment on schema personalisation is 'Behaviour events and recommendation model outputs.';

-- Keeps updated_at current. Attached explicitly to every table that has the column.
-- +goose StatementBegin
create function platform.set_updated_at() returns trigger
language plpgsql as $$
begin
  new.updated_at := now();
  return new;
end $$;
-- +goose StatementEnd

-- Human-readable references, e.g. platform.next_ref('TS-', 'sales.order_number_seq') → 'TS-10001'.
-- +goose StatementBegin
create function platform.next_ref(prefix text, seq regclass) returns text
language sql volatile as $$ select prefix || nextval(seq)::text $$;
-- +goose StatementEnd

-- Blocks UPDATE and DELETE on append-only tables (ledger, audit, stock movements, events).
-- +goose StatementBegin
create function platform.forbid_change() returns trigger
language plpgsql as $$
begin
  raise exception '%.% is append-only', tg_table_schema, tg_table_name;
end $$;
-- +goose StatementEnd

-- Creates monthly partitions for a range-partitioned table, from the month of `from_date`
-- for `months` months. Safe to call repeatedly; the worker calls it ahead of time.
-- +goose StatementBegin
create function platform.ensure_monthly_partitions(parent regclass, from_date date, months int)
returns void language plpgsql as $$
declare
  start_month date := date_trunc('month', from_date)::date;
  part_start date;
  part_name text;
  parent_schema text;
  parent_table text;
begin
  select n.nspname, c.relname into parent_schema, parent_table
  from pg_class c join pg_namespace n on n.oid = c.relnamespace
  where c.oid = parent;

  for i in 0 .. months - 1 loop
    part_start := (start_month + make_interval(months => i))::date;
    part_name := format('%s_%s', parent_table, to_char(part_start, 'YYYYMM'));
    execute format(
      'create table if not exists %I.%I partition of %s for values from (%L) to (%L)',
      parent_schema, part_name, parent, part_start, (part_start + interval '1 month')::date);
  end loop;
end $$;
-- +goose StatementEnd

-- +goose Down
drop function if exists platform.ensure_monthly_partitions(regclass, date, int);
drop function if exists platform.forbid_change();
drop function if exists platform.next_ref(text, regclass);
drop function if exists platform.set_updated_at();
drop schema if exists personalisation, cars, risk, support, content, marketing, messaging, aftersales,
  finance, payments, fulfilment, logistics, sales, pos, purchasing, inventory, search, catalog, b2b,
  sellers, identity, platform cascade;
-- Extensions are left installed: other database objects may rely on them.
