-- Trust and safety: KYC checks, risk rules and decisions, cases, the link graph, ban lists,
-- velocity counters, stolen-device blocklist, seller performance and enforcement, reports.
-- Plan: docs/trust-safety.md.

-- +goose Up
create table risk.risk_rule_sets (
  id            uuid primary key default gen_random_uuid(),
  checkpoint    text not null check (checkpoint in ('signup', 'signin', 'checkout', 'payout', 'bank_change', 'trade_in', 'return', 'review')),
  version       integer not null check (version > 0),
  status        text not null default 'draft' check (status in ('draft', 'active', 'retired')),
  thresholds    jsonb not null default '{"step": 40, "hold": 70, "block": 90}'::jsonb,
  published_by  uuid references identity.users (id),
  published_at  timestamptz,
  created_by    uuid not null references identity.users (id),
  created_at    timestamptz not null default now(),
  unique (checkpoint, version),
  check (status = 'draft' or (published_by is not null and published_at is not null))
);
comment on table risk.risk_rule_sets is '[risk] Versioned rule sets per checkpoint; only one active per checkpoint. Publishing needs a step-up.';
create unique index risk_rule_sets_one_active on risk.risk_rule_sets (checkpoint) where status = 'active';

create table risk.risk_rules (
  id           uuid primary key default gen_random_uuid(),
  rule_set_id  uuid not null references risk.risk_rule_sets (id) on delete cascade,
  key          text not null check (key ~ '^[a-z][a-z0-9_]*$'),
  label        text not null,
  condition    jsonb not null,
  weight       integer not null default 0 check (weight between -100 and 100),
  action       text not null default 'score' check (action in ('score', 'hold', 'block')),
  enabled      boolean not null default true,
  unique (rule_set_id, key)
);
comment on table risk.risk_rules is '[risk] Rules in a rule set: a condition over features, a weight, or a deterministic hold/block.';

create table risk.risk_decisions (
  id                bigint generated always as identity,
  checkpoint        text not null,
  subject_type      text not null,
  subject_id        uuid not null,
  user_id           uuid references identity.users (id),
  score             smallint not null check (score between 0 and 100),
  outcome           text not null check (outcome in ('allow', 'step_up', 'hold', 'block', 'fallback_allow', 'fallback_hold')),
  reasons           jsonb not null default '[]'::jsonb,
  rule_set_version  integer,
  latency_ms        integer check (latency_ms >= 0),
  created_at        timestamptz not null default now(),
  primary key (id, created_at)
) partition by range (created_at);
comment on table risk.risk_decisions is '[risk] Every risk decision with score, reasons and rule-set version. Partitioned monthly; 24-month retention.';
create table risk.risk_decisions_default partition of risk.risk_decisions default;
create index risk_decisions_subject_idx on risk.risk_decisions (subject_type, subject_id);

-- risk_cases already existed in the reviewed design (as a support table); it lives in risk now.
create table risk.risk_cases (
  id                  uuid primary key default gen_random_uuid(),
  kind                text not null check (kind in ('order', 'payment', 'device', 'seller', 'account', 'trade_in', 'return', 'review')),
  subject_type        text not null,
  subject_id          uuid not null,
  reason              text not null,
  score               smallint check (score between 0 and 100),
  money_at_risk_kobo  bigint check (money_at_risk_kobo >= 0),
  decision_id         bigint,
  status              text not null default 'open' check (status in ('open', 'in_review', 'cleared', 'blocked', 'escalated')),
  label               text check (label in ('fraud', 'legit', 'unclear')),
  sla_due_at          timestamptz,
  assigned_to         uuid references identity.staff_members (user_id),
  resolved_by         uuid references identity.users (id),
  resolved_at         timestamptz,
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now(),
  check (status in ('open', 'in_review', 'escalated') or (resolved_by is not null and resolved_at is not null and label is not null))
);
comment on table risk.risk_cases is '[risk] Fraud and risk reviews. Analysts'' clear/block decisions become labels for rule tuning.';
create index risk_cases_queue_idx on risk.risk_cases (status, score desc, created_at) where status in ('open', 'in_review', 'escalated');

create table risk.device_blocklist (
  id          uuid primary key default gen_random_uuid(),
  imei        text not null unique check (imei ~ '^[0-9]{15}$'),
  reason      text not null check (reason in ('reported_stolen', 'fraud', 'counterfeit')),
  source      text not null,
  reference   text,
  created_by  uuid references identity.users (id),
  created_at  timestamptz not null default now()
);
comment on table risk.device_blocklist is '[risk] IMEIs that may not be sold, traded in, returned or repaired (stolen, fraud, counterfeit).';

create table risk.kyc_checks (
  id            uuid primary key default gen_random_uuid(),
  subject_type  text not null check (subject_type in ('seller', 'business', 'rider')),
  subject_id    uuid not null,
  kind          text not null check (kind in ('nin', 'selfie', 'cac', 'bank_name', 'director', 'guarantor', 'licence')),
  provider      text not null,
  result        text not null check (result in ('match', 'partial', 'no_match', 'error')),
  score         numeric(4, 3) check (score between 0 and 1),
  raw_ref       text,
  checked_at    timestamptz not null default now()
);
comment on table risk.kyc_checks is '[risk] Results of each automated KYC check (provider or manual), shown to the reviewer.';
create index kyc_checks_subject_idx on risk.kyc_checks (subject_type, subject_id);

create table risk.risk_entities (
  id          uuid primary key default gen_random_uuid(),
  kind        text not null check (kind in ('account', 'device', 'phone', 'email', 'address', 'card_fp', 'bank_hmac', 'imei', 'ip')),
  value_hash  bytea not null,
  first_seen  timestamptz not null default now(),
  last_seen   timestamptz not null default now(),
  unique (kind, value_hash)
);
comment on table risk.risk_entities is '[risk] Nodes of the link graph (accounts, devices, phones, cards…), stored as hashes.';

create table risk.risk_links (
  entity_a    uuid not null references risk.risk_entities (id) on delete cascade,
  entity_b    uuid not null references risk.risk_entities (id) on delete cascade,
  first_seen  timestamptz not null default now(),
  last_seen   timestamptz not null default now(),
  primary key (entity_a, entity_b),
  check (entity_a < entity_b)
);
comment on table risk.risk_links is '[risk] Edges of the link graph: two entities seen together (stored once, a < b).';

create table risk.risk_lists (
  id          uuid primary key default gen_random_uuid(),
  kind        text not null check (kind in ('phone', 'email', 'bank', 'address', 'email_domain', 'device')),
  value_hash  bytea not null,
  reason      text not null,
  expires_at  timestamptz,
  created_by  uuid references identity.users (id),
  created_at  timestamptz not null default now(),
  unique (kind, value_hash)
);
comment on table risk.risk_lists is '[risk] Ban lists (phones, emails, bank accounts, addresses, disposable email domains), hashed.';

create table risk.velocity_counters (
  key           text not null,
  window_start  timestamptz not null,
  count         integer not null default 0 check (count >= 0),
  primary key (key, window_start)
);
comment on table risk.velocity_counters is '[risk] Window counters for velocity features (orders per device, cards per account…). Same design as identity.auth_rate_limits.';

create table risk.seller_metrics_daily (
  seller_id         uuid not null references sellers.sellers (id) on delete cascade,
  day               date not null,
  orders            integer not null default 0 check (orders >= 0),
  defects           integer not null default 0 check (defects >= 0),
  late_shipments    integer not null default 0 check (late_shipments >= 0),
  seller_cancels    integer not null default 0 check (seller_cancels >= 0),
  returns           integer not null default 0 check (returns >= 0),
  counterfeit_strikes integer not null default 0 check (counterfeit_strikes >= 0),
  primary key (seller_id, day)
);
comment on table risk.seller_metrics_daily is '[risk] Per-seller daily counts used for the rolling 60-day performance rates.';

create table risk.enforcement_actions (
  id             uuid primary key default gen_random_uuid(),
  seller_id      uuid not null references sellers.sellers (id),
  step           text not null check (step in ('warning', 'ranking_demotion', 'listing_review', 'payouts_held', 'suspended', 'reinstated')),
  reason         text not null,
  created_by     uuid not null references identity.users (id),
  appeal_status  text check (appeal_status in ('submitted', 'upheld', 'rejected')),
  appealed_at    timestamptz,
  decided_by     uuid references identity.users (id),
  decided_at     timestamptz,
  created_at     timestamptz not null default now(),
  -- Appeals are decided by someone other than whoever took the action.
  check (decided_by is distinct from created_by or decided_by is null),
  check (appeal_status is null or appealed_at is not null),
  check (appeal_status not in ('upheld', 'rejected') or (decided_by is not null and decided_at is not null))
);
comment on table risk.enforcement_actions is '[risk] Seller enforcement ladder steps, each with reasons and an appeal decided by a different person.';

create table risk.listing_reports (
  id           uuid primary key default gen_random_uuid(),
  listing_id   uuid not null references catalog.listings (id) on delete cascade,
  reporter_id  uuid references identity.users (id),
  source       text not null check (source in ('buyer', 'brand_owner', 'staff', 'mystery_shop')),
  reason       text not null check (reason in ('counterfeit', 'wrong_info', 'prohibited', 'scam', 'other')),
  details      text,
  status       text not null default 'open' check (status in ('open', 'upheld', 'dismissed')),
  decided_by   uuid references identity.users (id),
  created_at   timestamptz not null default now(),
  updated_at   timestamptz not null default now(),
  check (status = 'open' or decided_by is not null)
);
comment on table risk.listing_reports is '[risk] Reports of counterfeit or misleading listings; upheld reports are seller strikes.';

create table risk.review_reports (
  id           uuid primary key default gen_random_uuid(),
  review_id    uuid not null references catalog.product_reviews (id) on delete cascade,
  reporter_id  uuid references identity.users (id),
  reason       text not null check (reason in ('abuse', 'personal_data', 'off_topic', 'fake', 'other')),
  status       text not null default 'open' check (status in ('open', 'upheld', 'dismissed')),
  decided_by   uuid references identity.users (id),
  created_at   timestamptz not null default now(),
  updated_at   timestamptz not null default now(),
  check (status = 'open' or decided_by is not null)
);
comment on table risk.review_reports is '[risk] Reports of reviews that break policy (sellers can report, never delete).';

-- Indexes on foreign-key columns (every FK is indexed).
create index device_blocklist_created_by_idx on risk.device_blocklist (created_by);
create index enforcement_actions_created_by_idx on risk.enforcement_actions (created_by);
create index enforcement_actions_decided_by_idx on risk.enforcement_actions (decided_by);
create index enforcement_actions_seller_id_idx on risk.enforcement_actions (seller_id);
create index listing_reports_decided_by_idx on risk.listing_reports (decided_by);
create index listing_reports_listing_id_idx on risk.listing_reports (listing_id);
create index listing_reports_reporter_id_idx on risk.listing_reports (reporter_id);
create index review_reports_decided_by_idx on risk.review_reports (decided_by);
create index review_reports_reporter_id_idx on risk.review_reports (reporter_id);
create index review_reports_review_id_idx on risk.review_reports (review_id);
create index risk_cases_assigned_to_idx on risk.risk_cases (assigned_to);
create index risk_cases_resolved_by_idx on risk.risk_cases (resolved_by);
create index risk_decisions_user_id_idx on risk.risk_decisions (user_id);
create index risk_links_entity_b_idx on risk.risk_links (entity_b);
create index risk_lists_created_by_idx on risk.risk_lists (created_by);
create index risk_rule_sets_created_by_idx on risk.risk_rule_sets (created_by);
create index risk_rule_sets_published_by_idx on risk.risk_rule_sets (published_by);

-- Keep updated_at current.
create trigger listing_reports_set_updated_at before update on risk.listing_reports
  for each row execute function platform.set_updated_at();
create trigger review_reports_set_updated_at before update on risk.review_reports
  for each row execute function platform.set_updated_at();
create trigger risk_cases_set_updated_at before update on risk.risk_cases
  for each row execute function platform.set_updated_at();

-- +goose Down
drop table if exists risk.review_reports, risk.listing_reports, risk.enforcement_actions,
  risk.seller_metrics_daily, risk.velocity_counters, risk.risk_lists, risk.risk_links,
  risk.risk_entities, risk.kyc_checks, risk.device_blocklist, risk.risk_cases, risk.risk_decisions,
  risk.risk_rules, risk.risk_rule_sets cascade;
