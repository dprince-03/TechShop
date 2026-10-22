-- Support tickets: customers, sellers and public contact forms; staff queue. Internal notes are
-- never returned to customers or sellers.

-- name: SupportCreateTicket :one
insert into support.support_tickets (customer_user_id, seller_id, order_id, site, contact_name, contact_email, contact_phone,
  channel, topic, subject, priority, sla_due_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
returning *;

-- name: SupportAddMessage :one
insert into support.ticket_messages (ticket_id, author_user_id, author_kind, body, is_internal)
values ($1, $2, $3, $4, $5)
returning *;

-- name: SupportTicket :one
select t.*, o.order_number, s.display_name as shop_name,
       cu.first_name as customer_first_name, cu.last_name as customer_last_name,
       au.first_name as assignee_first_name, au.last_name as assignee_last_name
from support.support_tickets t
left join sales.orders o on o.id = t.order_id
left join sellers.sellers s on s.id = t.seller_id
left join identity.users cu on cu.id = t.customer_user_id
left join identity.users au on au.id = t.assigned_to
where t.ticket_number = $1;

-- name: SupportTicketForUpdate :one
select * from support.support_tickets where ticket_number = $1 for update;

-- name: SupportMessages :many
-- Every message, internal notes included (staff only), oldest first.
select m.*, u.first_name, u.last_name from support.ticket_messages m left join identity.users u on u.id = m.author_user_id
where m.ticket_id = $1 order by m.created_at;

-- name: SupportPublicMessages :many
-- What the customer or seller may see: no internal notes.
select m.*, u.first_name, u.last_name from support.ticket_messages m left join identity.users u on u.id = m.author_user_id
where m.ticket_id = $1 and not m.is_internal order by m.created_at;

-- name: SupportTicketsForCustomer :many
select t.*, o.order_number from support.support_tickets t left join sales.orders o on o.id = t.order_id
where t.customer_user_id = $1 and t.seller_id is null
order by t.last_message_at desc limit 100;

-- name: SupportTicketsForSeller :many
select t.*, o.order_number from support.support_tickets t left join sales.orders o on o.id = t.order_id
where t.seller_id = $1
order by t.last_message_at desc limit 100;

-- name: SupportQueue :many
-- Staff queue. "active" = open or waiting on the customer; urgent and high first, then oldest
-- reply due; everything else newest first.
select t.*, o.order_number, s.display_name as shop_name,
       cu.first_name as customer_first_name, cu.last_name as customer_last_name,
       au.first_name as assignee_first_name, au.last_name as assignee_last_name
from support.support_tickets t
left join sales.orders o on o.id = t.order_id
left join sellers.sellers s on s.id = t.seller_id
left join identity.users cu on cu.id = t.customer_user_id
left join identity.users au on au.id = t.assigned_to
where (sqlc.narg('status')::text is null or t.status = sqlc.narg('status')
       or (sqlc.narg('status') = 'active' and t.status in ('open', 'pending_customer')))
  and (sqlc.narg('priority')::text is null or t.priority = sqlc.narg('priority')
       or (sqlc.narg('priority') = 'urgent_high' and t.priority in ('urgent', 'high')))
  and (sqlc.narg('assignee')::uuid is null or t.assigned_to = sqlc.narg('assignee'))
  and (sqlc.narg('q')::text is null or t.ticket_number ilike sqlc.narg('q') || '%' or o.order_number ilike sqlc.narg('q') || '%'
       or t.contact_email = sqlc.narg('q') or t.subject ilike '%' || sqlc.narg('q') || '%')
order by (t.status = 'open') desc,
         case t.priority when 'urgent' then 0 when 'high' then 1 when 'normal' then 2 else 3 end,
         t.sla_due_at nulls last, t.last_message_at desc
limit 200;

-- name: SupportSummary :one
select
  count(*) filter (where status = 'open')::int as open,
  count(*) filter (where status = 'open' and priority in ('urgent', 'high'))::int as urgent,
  count(*) filter (where status = 'open' and first_response_at is null and sla_due_at < now())::int as overdue,
  count(*) filter (where status = 'pending_customer')::int as waiting,
  count(*) filter (where resolved_at >= date_trunc('day', now() at time zone 'Africa/Lagos') at time zone 'Africa/Lagos')::int as resolved_today,
  coalesce(extract(epoch from avg(first_response_at - created_at) filter (where first_response_at > now() - interval '7 days')) / 60, 0)::int as avg_first_reply_minutes
from support.support_tickets;

-- name: SupportCustomerReplied :one
-- The customer (or seller) wrote: reopen if resolved, stop waiting, flag for staff.
update support.support_tickets
set status = 'open', pending_since = null, resolved_at = null, last_message_at = now(), customer_unread = false
where id = $1 and status <> 'closed'
returning *;

-- name: SupportStaffReplied :one
-- A public staff reply: the first one stops the SLA clock; the ticket waits on the customer
-- unless staff resolve it or keep it open.
update support.support_tickets
set first_response_at = coalesce(first_response_at, now()), last_message_at = now(), customer_unread = true,
    status = @status::text,
    pending_since = case when @status::text = 'pending_customer' then now() else null end,
    resolved_at = case when @status::text = 'resolved' then now() else null end
where id = @id
returning *;

-- name: SupportSetStatus :one
update support.support_tickets
set status = @status::text, priority = coalesce(sqlc.narg('priority'), priority),
    pending_since = case when @status::text = 'pending_customer' then coalesce(pending_since, now()) else null end,
    resolved_at = case when @status::text = 'resolved' then coalesce(resolved_at, now()) when @status::text = 'closed' then resolved_at else null end,
    closed_at = case when @status::text = 'closed' then now() else null end
where id = @id
returning *;

-- name: SupportAssign :one
update support.support_tickets set assigned_to = $2 where id = $1 returning *;

-- name: SupportMarkRead :exec
update support.support_tickets set customer_unread = false where id = $1;

-- name: SupportAutoResolve :many
-- Waiting on the customer for 7 days with no reply: resolved.
update support.support_tickets set status = 'resolved', resolved_at = now(), pending_since = null
where status = 'pending_customer' and pending_since < now() - interval '7 days'
returning id, ticket_number;

-- name: SupportAutoClose :exec
-- Resolved for 4 days: closed (a new message then starts a new ticket).
update support.support_tickets set status = 'closed', closed_at = now()
where status = 'resolved' and resolved_at < now() - interval '4 days';

-- name: SupportAgents :many
-- Active staff who can reply to tickets (for assignment).
select distinct u.id, u.first_name, u.last_name
from identity.users u
join identity.staff_members sm on sm.user_id = u.id and sm.status = 'active'
join identity.staff_roles sr on sr.user_id = u.id
join identity.role_permissions rp on rp.role_id = sr.role_id
join identity.permissions p on p.id = rp.permission_id
where p.key = 'support.reply' and u.status = 'active'
order by u.first_name, u.last_name;

-- name: SupportIsAgent :one
select exists (
  select 1 from identity.staff_roles sr
  join identity.staff_members sm on sm.user_id = sr.user_id and sm.status = 'active'
  join identity.role_permissions rp on rp.role_id = sr.role_id
  join identity.permissions p on p.id = rp.permission_id
  where sr.user_id = $1 and p.key = 'support.reply');

-- name: SupportCustomerOrder :one
select id from sales.orders where order_number = $1 and customer_user_id = $2;

-- name: SupportSellerOrder :one
select distinct o.id from sales.orders o join sales.fulfilments f on f.order_id = o.id
where o.order_number = $1 and f.seller_id = $2;

-- name: SupportUserContact :one
select first_name, last_name, email, phone from identity.users where id = $1;
