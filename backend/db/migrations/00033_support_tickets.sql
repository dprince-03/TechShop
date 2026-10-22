-- Support tickets (M5): subject, reply SLA and the timestamps the auto-close job needs.
-- Plan: docs/backend.md §5 (support), docs/plan.md 2026-10-23.

-- +goose Up
alter table support.support_tickets
  add column subject           text check (length(subject) <= 140),
  add column sla_due_at        timestamptz,
  add column pending_since     timestamptz,
  add column closed_at         timestamptz,
  add column last_message_at   timestamptz not null default now(),
  add column customer_unread   boolean not null default false,
  add constraint support_tickets_closed_check check (status <> 'closed' or closed_at is not null),
  add constraint support_tickets_topic_check check (topic in ('order', 'delivery', 'payment', 'refund', 'return', 'warranty',
    'trade_in', 'account', 'selling', 'business', 'other'));
comment on column support.support_tickets.sla_due_at is 'When the first staff reply is due (24 h normal, 4 h high or urgent).';
comment on column support.support_tickets.pending_since is 'When the ticket started waiting on the customer (auto-resolved after 7 days).';

create index support_tickets_pending_idx on support.support_tickets (pending_since) where status = 'pending_customer';
create index support_tickets_resolved_idx on support.support_tickets (resolved_at) where status = 'resolved';
create index support_tickets_contact_email_idx on support.support_tickets (contact_email) where contact_email is not null;

-- +goose Down
drop index if exists support.support_tickets_contact_email_idx;
drop index if exists support.support_tickets_resolved_idx;
drop index if exists support.support_tickets_pending_idx;
alter table support.support_tickets
  drop constraint if exists support_tickets_topic_check,
  drop constraint if exists support_tickets_closed_check,
  drop column if exists customer_unread,
  drop column if exists last_message_at,
  drop column if exists closed_at,
  drop column if exists pending_since,
  drop column if exists sla_due_at,
  drop column if exists subject;
