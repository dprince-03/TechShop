-- Least-privilege database roles (docs/database/RLS-POLICIES.md, .claude/rules/01 §6).
-- The user that runs migrations owns every table. The API and worker log in as a separate user
-- that is a member of techshop_rw: it reads and writes rows but cannot create, alter or drop
-- tables, and cannot change or delete append-only history (audit log, ledger, stock movements,
-- delivery events). techshop_ro is for reporting: read-only, and only on schemas without
-- personal data. Login users and their passwords are secrets, so they are created outside
-- migrations (locally: `make db-users`; deployed: by ops), as members of these group roles.

-- +goose Up
-- +goose StatementBegin
do $$
declare
  s text;
  all_schemas text[] := array['platform', 'identity', 'sellers', 'b2b', 'catalog', 'search', 'inventory',
    'purchasing', 'pos', 'sales', 'logistics', 'fulfilment', 'payments', 'finance', 'aftersales',
    'messaging', 'marketing', 'content', 'support', 'risk', 'cars', 'personalisation', 'river'];
  -- Reporting may read these: no names, phones, addresses or identity documents in them.
  readonly_schemas text[] := array['catalog', 'search', 'inventory', 'purchasing', 'content'];
begin
  if not exists (select 1 from pg_roles where rolname = 'techshop_rw') then
    create role techshop_rw nologin;
  end if;
  if not exists (select 1 from pg_roles where rolname = 'techshop_ro') then
    create role techshop_ro nologin;
  end if;

  foreach s in array all_schemas loop
    execute format('grant usage on schema %I to techshop_rw', s);
    execute format('grant select, insert, update, delete on all tables in schema %I to techshop_rw', s);
    execute format('grant usage, select, update on all sequences in schema %I to techshop_rw', s);
    execute format('grant execute on all functions in schema %I to techshop_rw', s);
    -- Tables, sequences and functions added by later migrations get the same grants.
    execute format('alter default privileges in schema %I grant select, insert, update, delete on tables to techshop_rw', s);
    execute format('alter default privileges in schema %I grant usage, select, update on sequences to techshop_rw', s);
    execute format('alter default privileges in schema %I grant execute on functions to techshop_rw', s);
  end loop;

  foreach s in array readonly_schemas loop
    execute format('grant usage on schema %I to techshop_ro', s);
    execute format('grant select on all tables in schema %I to techshop_ro', s);
    execute format('alter default privileges in schema %I grant select on tables to techshop_ro', s);
  end loop;
end $$;
-- +goose StatementEnd

-- History is append-only: the app may add rows but never change or remove them. (The ledger,
-- stock movements and delivery events also have forbid_change triggers; the audit log relies on
-- this revoke.)
revoke update, delete on platform.audit_log, finance.ledger_journals, finance.ledger_entries,
  inventory.stock_movements, logistics.delivery_events from techshop_rw;

-- The worker creates next months' log partitions; that is DDL, so the function runs with its
-- owner's rights instead of the caller's. search_path is pinned so it can't be hijacked.
alter function platform.ensure_monthly_partitions(regclass, date, int)
  security definer set search_path = pg_catalog, pg_temp;
revoke execute on function platform.ensure_monthly_partitions(regclass, date, int) from public;
grant execute on function platform.ensure_monthly_partitions(regclass, date, int) to techshop_rw;

-- +goose Down
alter function platform.ensure_monthly_partitions(regclass, date, int) security invoker reset search_path;
grant execute on function platform.ensure_monthly_partitions(regclass, date, int) to public;
-- +goose StatementBegin
do $$
declare
  s text;
  all_schemas text[] := array['platform', 'identity', 'sellers', 'b2b', 'catalog', 'search', 'inventory',
    'purchasing', 'pos', 'sales', 'logistics', 'fulfilment', 'payments', 'finance', 'aftersales',
    'messaging', 'marketing', 'content', 'support', 'risk', 'cars', 'personalisation', 'river'];
begin
  foreach s in array all_schemas loop
    execute format('alter default privileges in schema %I revoke all on tables from techshop_rw, techshop_ro', s);
    execute format('alter default privileges in schema %I revoke all on sequences from techshop_rw', s);
    execute format('alter default privileges in schema %I revoke all on functions from techshop_rw', s);
  end loop;
end $$;
-- +goose StatementEnd
drop owned by techshop_rw, techshop_ro;
drop role if exists techshop_rw;
drop role if exists techshop_ro;
