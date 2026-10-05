-- Content (CMS pages, banners, help articles) and support (tickets, messages, attachments).

-- +goose Up
create table content.cms_pages (
  id            uuid primary key default gen_random_uuid(),
  site          text not null check (site in ('corporate', 'market', 'wholesale', 'seller')),
  slug          text not null,
  title         text not null,
  body          jsonb not null default '[]'::jsonb,
  status        text not null default 'draft' check (status in ('draft', 'in_review', 'published')),
  published_at  timestamptz,
  updated_by    uuid references identity.users (id),
  approved_by   uuid references identity.users (id),
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  unique (site, slug),
  check (status <> 'published' or published_at is not null)
);
comment on table content.cms_pages is '[content] Editable pages per site (about, legal, credit terms…) as content blocks; publishing needs approval.';

create table content.banners (
  id             uuid primary key default gen_random_uuid(),
  site           text not null check (site in ('corporate', 'market', 'wholesale', 'seller')),
  placement      text not null,
  title          text not null,
  image_file_id  uuid references platform.files (id),
  link_url       text not null check (link_url ~ '^(/|https://)'),
  starts_at      timestamptz,
  ends_at        timestamptz,
  position       integer not null default 0,
  status         text not null default 'draft' check (status in ('draft', 'live', 'ended')),
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now(),
  check (ends_at is null or starts_at is null or ends_at > starts_at)
);
comment on table content.banners is '[content] Hero and promo banners per site and placement (links: internal paths or https only).';

create table content.help_articles (
  id            uuid primary key default gen_random_uuid(),
  site          text not null check (site in ('market', 'wholesale', 'seller')),
  topic         text not null,
  slug          text not null,
  title         text not null,
  body          jsonb not null default '[]'::jsonb,
  status        text not null default 'draft' check (status in ('draft', 'in_review', 'published')),
  published_at  timestamptz,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  unique (site, slug),
  check (status <> 'published' or published_at is not null)
);
comment on table content.help_articles is '[content] Help centre articles and FAQs per site.';

create sequence support.ticket_number_seq start 5001;

create table support.support_tickets (
  id                 uuid primary key default gen_random_uuid(),
  ticket_number      text not null unique default platform.next_ref('TKT-', 'support.ticket_number_seq'),
  customer_user_id   uuid references identity.users (id),
  seller_id          uuid references sellers.sellers (id),
  order_id           uuid references sales.orders (id),
  site               text check (site in ('corporate', 'market', 'wholesale', 'seller', 'app')),
  contact_name       text,
  contact_email      citext,
  contact_phone      text check (contact_phone ~ '^\+234[0-9]{10}$'),
  channel            text not null check (channel in ('web', 'app', 'whatsapp', 'email', 'phone')),
  topic              text not null,
  priority           text not null default 'normal' check (priority in ('low', 'normal', 'high', 'urgent')),
  status             text not null default 'open' check (status in ('open', 'pending_customer', 'resolved', 'closed')),
  assigned_to        uuid references identity.staff_members (user_id),
  first_response_at  timestamptz,
  resolved_at        timestamptz,
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now(),
  check (customer_user_id is not null or seller_id is not null or contact_email is not null or contact_phone is not null)
);
comment on table support.support_tickets is '[support] Support conversations with customers, sellers or public contact-form visitors.';
create index support_tickets_queue_idx on support.support_tickets (status, priority, created_at);

create table support.ticket_messages (
  id              uuid primary key default gen_random_uuid(),
  ticket_id       uuid not null references support.support_tickets (id) on delete cascade,
  author_user_id  uuid references identity.users (id),
  author_kind     text not null check (author_kind in ('customer', 'seller', 'staff', 'system')),
  body            text not null,
  is_internal     boolean not null default false,
  created_at      timestamptz not null default now(),
  check (not is_internal or author_kind in ('staff', 'system'))
);
comment on table support.ticket_messages is '[support] Messages in a ticket; internal notes are never returned to customers.';

create table support.message_attachments (
  message_id  uuid not null references support.ticket_messages (id) on delete cascade,
  file_id     uuid not null references platform.files (id),
  primary key (message_id, file_id)
);
comment on table support.message_attachments is '[support] Files attached to ticket messages.';

-- Indexes on foreign-key columns (every FK is indexed).
create index banners_image_file_id_idx on content.banners (image_file_id);
create index cms_pages_approved_by_idx on content.cms_pages (approved_by);
create index cms_pages_updated_by_idx on content.cms_pages (updated_by);
create index message_attachments_file_id_idx on support.message_attachments (file_id);
create index support_tickets_assigned_to_idx on support.support_tickets (assigned_to);
create index support_tickets_customer_user_id_idx on support.support_tickets (customer_user_id);
create index support_tickets_order_id_idx on support.support_tickets (order_id);
create index support_tickets_seller_id_idx on support.support_tickets (seller_id);
create index ticket_messages_author_user_id_idx on support.ticket_messages (author_user_id);
create index ticket_messages_ticket_id_idx on support.ticket_messages (ticket_id);

-- Keep updated_at current.
create trigger banners_set_updated_at before update on content.banners
  for each row execute function platform.set_updated_at();
create trigger cms_pages_set_updated_at before update on content.cms_pages
  for each row execute function platform.set_updated_at();
create trigger help_articles_set_updated_at before update on content.help_articles
  for each row execute function platform.set_updated_at();
create trigger support_tickets_set_updated_at before update on support.support_tickets
  for each row execute function platform.set_updated_at();

-- +goose Down
drop table if exists support.message_attachments, support.ticket_messages, support.support_tickets,
  content.help_articles, content.banners, content.cms_pages cascade;
drop sequence if exists support.ticket_number_seq;
