-- Payments: charges, provider events, refunds, provider health, reconciliation (docs/payments-finance.md).

-- name: PaymentsCreate :one
insert into payments.payments (order_id, invoice_id, provider, channel, idempotency_key, amount_kobo, expires_at)
values ($1, $2, $3, $4, $5, $6, $7)
returning *;

-- name: PaymentsGet :one
select * from payments.payments where id = $1;

-- name: PaymentsGetForUpdate :one
select * from payments.payments where id = $1 for update;

-- name: PaymentsGetByReference :one
select * from payments.payments where provider = $1 and provider_reference = $2;

-- name: PaymentsGetByAnyReference :one
select * from payments.payments where provider_reference = $1;

-- name: PaymentsSetOpened :one
update payments.payments set status = 'pending', provider_reference = $2, checkout_url = $3, channel = coalesce(sqlc.narg('channel'), channel), updated_at = now()
where id = $1 and status = 'initiated' returning *;

-- name: PaymentsSetStatus :one
update payments.payments
set status = $2, provider_status = coalesce(sqlc.narg('provider_status'), provider_status),
    paid_at = case when $2 = 'succeeded' then coalesce(paid_at, now()) else paid_at end,
    fees_kobo = coalesce(sqlc.narg('fees_kobo'), fees_kobo),
    failure_reason = coalesce(sqlc.narg('failure_reason'), failure_reason), updated_at = now()
where id = $1 returning *;

-- name: PaymentsForOrder :many
select * from payments.payments where order_id = $1 order by created_at;

-- name: PaymentsLatestForOrder :one
select * from payments.payments where order_id = $1 order by created_at desc limit 1;

-- name: PaymentsSucceededForOrder :one
select * from payments.payments where order_id = $1 and status in ('succeeded', 'reversed') order by paid_at limit 1;

-- name: PaymentsCountForOrder :one
select count(*)::int from payments.payments where order_id = $1;

-- name: PaymentsStalePending :many
-- The sweep re-verifies pending payments older than the threshold, so a missed webhook never strands an order.
select * from payments.payments where status in ('initiated', 'pending') and created_at < $1 order by created_at limit 100;

-- name: PaymentsInsertEvent :one
insert into payments.payment_events (provider, provider_event_id, event_type, payment_id, payload, signature_valid)
values ($1, $2, $3, $4, $5, $6)
on conflict (provider, provider_event_id) do nothing
returning id;

-- name: PaymentsGetEvent :one
select * from payments.payment_events where id = $1;

-- name: PaymentsMarkEventProcessed :exec
update payments.payment_events set processed_at = now(), payment_id = coalesce(payment_id, sqlc.narg('payment_id')) where id = $1;

-- name: PaymentsCreateRefund :one
insert into payments.refunds (payment_id, order_id, return_request_id, amount_kobo, method, reason, requested_by)
values ($1, $2, $3, $4, $5, $6, $7)
returning *;

-- name: PaymentsGetRefund :one
select * from payments.refunds where id = $1;

-- name: PaymentsGetRefundForUpdate :one
select * from payments.refunds where id = $1 for update;

-- name: PaymentsSetRefundStatus :one
update payments.refunds
set status = $2, approved_by = coalesce(sqlc.narg('approved_by'), approved_by),
    provider_reference = coalesce(sqlc.narg('provider_reference'), provider_reference),
    failure_reason = coalesce(sqlc.narg('failure_reason'), failure_reason), updated_at = now()
where id = $1 returning *;

-- name: PaymentsRefundedTotal :one
-- Money already committed to refunds of a payment's order total (anything not rejected or
-- failed). Excess-payment refunds are money on top of the total, so they don't count.
select coalesce(sum(amount_kobo), 0)::bigint from payments.refunds
where payment_id = $1 and status not in ('rejected', 'failed') and reason not like 'Excess payment: %';

-- name: PaymentsListRefunds :many
select r.*, o.order_number from payments.refunds r join sales.orders o on o.id = r.order_id
where (sqlc.narg('status')::text is null or r.status = sqlc.narg('status'))
order by r.created_at desc limit $1;

-- name: PaymentsCustomerRefunds :many
select r.id, r.amount_kobo, r.method, r.reason, r.status, r.created_at, r.updated_at, o.order_number
from payments.refunds r join sales.orders o on o.id = r.order_id
where o.customer_user_id = $1 order by r.created_at desc limit 100;

-- name: PaymentsOrderRefunds :many
select * from payments.refunds where order_id = $1 order by created_at;

-- name: PaymentsGetHealth :many
select * from payments.provider_health;

-- name: PaymentsRecordFailure :one
-- Circuit breaker: 5 failures in a row open the breaker for this provider.
insert into payments.provider_health (provider, failures, last_failure_at, breaker_state) values ($1, 1, now(), 'closed')
on conflict (provider) do update set failures = payments.provider_health.failures + 1, last_failure_at = now(),
  breaker_state = case when payments.provider_health.failures + 1 >= 5 then 'open' else payments.provider_health.breaker_state end, updated_at = now()
returning *;

-- name: PaymentsRecordSuccess :exec
insert into payments.provider_health (provider) values ($1)
on conflict (provider) do update set failures = 0, breaker_state = 'closed', updated_at = now();

-- name: PaymentsCreateException :one
insert into payments.reconciliation_exceptions (kind, provider, reference, amount_kobo, detail) values ($1, $2, $3, $4, $5) returning *;

-- name: PaymentsListExceptions :many
select * from payments.reconciliation_exceptions where (sqlc.narg('status')::text is null or status = sqlc.narg('status')) order by created_at desc limit $1;

-- name: PaymentsResolveException :one
update payments.reconciliation_exceptions set status = $2, resolution = $3, resolved_by = $4, resolved_at = now(), updated_at = now()
where id = $1 and status = 'open' returning *;

-- name: PaymentsCreateVirtualAccount :one
insert into payments.virtual_accounts (provider, account_number, bank_name, owner_type, order_id, business_id, expected_kobo, expires_at)
values ($1, $2, $3, $4, $5, $6, $7, $8) returning *;

-- name: PaymentsVirtualAccountForOrder :one
select * from payments.virtual_accounts where order_id = $1 and status = 'active' order by created_at desc limit 1;

-- name: PaymentsSetVirtualAccountStatus :exec
update payments.virtual_accounts set status = $2, received_kobo = $3, updated_at = now() where id = $1;

-- name: PaymentsFakeLatestCharge :one
-- Development fake provider: its "server state" is the last signed charge event it sent.
select payload from payments.payment_events
where provider = $1 and payload -> 'data' ->> 'reference' = @reference::text and signature_valid and event_type like 'charge.%'
order by received_at desc limit 1;

-- name: PaymentsOrderWithCustomer :one
select o.*, u.email as customer_email, u.phone as customer_phone, u.first_name as customer_first_name
from sales.orders o left join identity.users u on u.id = o.customer_user_id where o.id = $1;
