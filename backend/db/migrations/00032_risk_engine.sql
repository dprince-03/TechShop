-- Risk engine (M5): launch rule sets, case references and a case timeline.
-- Plan: docs/trust-safety.md §3–4; docs/plan.md 2026-10-23.

-- +goose Up
-- The launch rule sets are written by this migration, not by a person, so a rule set may have no
-- creator or publisher (null = system). Anything a person publishes still records who.
alter table risk.risk_rule_sets alter column created_by drop not null;
alter table risk.risk_rule_sets drop constraint risk_rule_sets_check;
alter table risk.risk_rule_sets add constraint risk_rule_sets_published_check
  check (status = 'draft' or published_at is not null);
alter table risk.risk_rule_sets add column notes text;

create sequence risk.case_ref_seq start 301;
alter table risk.risk_cases add column reference text unique default platform.next_ref('RK-', 'risk.case_ref_seq');
update risk.risk_cases set reference = platform.next_ref('RK-', 'risk.case_ref_seq') where reference is null;
alter table risk.risk_cases alter column reference set not null;
alter table risk.risk_cases add column resolution_note text;
create index risk_cases_subject_idx on risk.risk_cases (subject_type, subject_id);

-- What happened on a case, in order: opened, claimed, notes, escalated, cleared or blocked.
create table risk.case_events (
  id          uuid primary key default gen_random_uuid(),
  case_id     uuid not null references risk.risk_cases (id) on delete cascade,
  actor_id    uuid references identity.users (id),
  action      text not null check (action in ('opened', 'claimed', 'note', 'escalated', 'cleared', 'blocked')),
  note        text,
  created_at  timestamptz not null default now()
);
comment on table risk.case_events is '[risk] Timeline of a risk case: who claimed, noted, escalated, cleared or blocked it.';
create index case_events_case_id_idx on risk.case_events (case_id, created_at);
create index case_events_actor_id_idx on risk.case_events (actor_id);
create index risk_decisions_checkpoint_idx on risk.risk_decisions (checkpoint, created_at);

-- Launch rule sets (trust-safety.md §3 table). Conditions: {"all": [{"f": feature, "op": lt|lte|gt|gte|eq, "v": number}]}.
-- +goose StatementBegin
do $$
declare
  rs uuid;
begin
  insert into risk.risk_rule_sets (checkpoint, version, status, published_at, notes)
  values ('checkout', 1, 'active', now(), 'Launch rules') returning id into rs;
  insert into risk.risk_rules (rule_set_id, key, label, condition, weight, action) values
    (rs, 'new_account_high_value', 'New account (under 24 h) and order over ₦500,000',
       '{"all": [{"f": "account_age_hours", "op": "lt", "v": 24}, {"f": "order_total_kobo", "op": "gt", "v": 50000000}]}', 30, 'score'),
    (rs, 'first_order_over_1m', 'First order over ₦1,000,000',
       '{"all": [{"f": "delivered_orders", "op": "eq", "v": 0}, {"f": "order_total_kobo", "op": "gt", "v": 100000000}]}', 25, 'score'),
    (rs, 'failed_payments_10m', '3 or more failed payments in 10 minutes',
       '{"all": [{"f": "failed_payments_10m", "op": "gte", "v": 3}]}', 35, 'score'),
    (rs, 'multiple_same_phone', '2 or more units of the same phone',
       '{"all": [{"f": "max_phone_units", "op": "gte", "v": 2}]}', 15, 'score'),
    (rs, 'open_risk_case', 'Customer has an open risk case',
       '{"all": [{"f": "open_cases", "op": "gte", "v": 1}]}', 40, 'score'),
    (rs, 'repeat_customer', 'Repeat customer with delivered orders',
       '{"all": [{"f": "delivered_orders", "op": "gte", "v": 1}]}', -20, 'score');

  insert into risk.risk_rule_sets (checkpoint, version, status, published_at, notes)
  values ('payout', 1, 'active', now(), 'Launch rules') returning id into rs;
  insert into risk.risk_rules (rule_set_id, key, label, condition, weight, action) values
    (rs, 'payout_over_3x_average', 'Payout over 3× the 30-day average',
       '{"all": [{"f": "payout_vs_avg", "op": "gt", "v": 3}]}', 30, 'score'),
    (rs, 'chargeback_rate', 'Chargeback rate over 1 %',
       '{"all": [{"f": "chargeback_rate_pct", "op": "gt", "v": 1}]}', 25, 'score'),
    (rs, 'open_seller_case', 'Shop has an open risk case',
       '{"all": [{"f": "open_cases", "op": "gte", "v": 1}]}', 60, 'score');

  insert into risk.risk_rule_sets (checkpoint, version, status, published_at, notes)
  values ('trade_in', 1, 'active', now(), 'Launch rules') returning id into rs;
  insert into risk.risk_rules (rule_set_id, key, label, condition, weight, action) values
    (rs, 'imei_fails_luhn', 'IMEI fails the Luhn check',
       '{"all": [{"f": "imei_luhn_ok", "op": "eq", "v": 0}]}', 0, 'block'),
    (rs, 'imei_blocklisted', 'IMEI is on the stolen-device blocklist',
       '{"all": [{"f": "imei_blocked", "op": "eq", "v": 1}]}', 0, 'block'),
    (rs, 'new_account_many_trade_ins', 'Account under 30 days with 3+ trade-ins this week',
       '{"all": [{"f": "account_age_days", "op": "lt", "v": 30}, {"f": "trade_ins_7d", "op": "gte", "v": 3}]}', 40, 'score');

  insert into risk.risk_rule_sets (checkpoint, version, status, published_at, notes)
  values ('return', 1, 'active', now(), 'Launch rules') returning id into rs;
  insert into risk.risk_rules (rule_set_id, key, label, condition, weight, action) values
    (rs, 'high_return_rate', 'Returns over 30 % of delivered orders in 90 days',
       '{"all": [{"f": "return_rate_90d", "op": "gt", "v": 0.30}]}', 40, 'score'),
    (rs, 'open_risk_case', 'Customer has an open risk case',
       '{"all": [{"f": "open_cases", "op": "gte", "v": 1}]}', 40, 'score');
end $$;
-- +goose StatementEnd

-- +goose Down
delete from risk.risk_rule_sets where created_by is null;
drop index if exists risk.risk_decisions_checkpoint_idx;
drop table if exists risk.case_events;
drop index if exists risk.risk_cases_subject_idx;
alter table risk.risk_cases drop column if exists resolution_note;
alter table risk.risk_cases drop column if exists reference;
drop sequence if exists risk.case_ref_seq;
alter table risk.risk_rule_sets drop column if exists notes;
alter table risk.risk_rule_sets drop constraint risk_rule_sets_published_check;
alter table risk.risk_rule_sets add constraint risk_rule_sets_check
  check (status = 'draft' or (published_by is not null and published_at is not null));
alter table risk.risk_rule_sets alter column created_by set not null;
