-- Seller payouts (docs/payments-finance.md §9). Amounts come from order lines and are capped by
-- the seller's live ledger balance, so refunds and chargebacks already booked are netted off.

-- name: PayoutsFulfilmentEarnings :many
-- Every unpaid-out fulfilment of a seller whose order was captured (a sale journal exists), with
-- what it earns (line earnings minus commission) and whether it can be paid out yet: delivered,
-- outside the return window and with no return still open on its items.
select f.id, f.status, f.delivered_at, o.order_number,
       (coalesce(sum(coalesce(ol.seller_earning_kobo, ol.line_total_kobo)) filter (where ol.cancelled_at is null), 0)
        - coalesce(sum((ol.line_total_kobo * ol.commission_bps + 5000) / 10000) filter (where ol.cancelled_at is null), 0))::bigint as earning_kobo,
       (f.status = 'delivered' and f.delivered_at < @eligible_before::timestamptz
        and not exists (select 1 from aftersales.return_items ri join aftersales.return_requests rr on rr.id = ri.return_request_id
                        join sales.order_lines x on x.id = ri.order_line_id
                        where x.fulfilment_id = f.id and rr.status not in ('rejected', 'refunded', 'replaced', 'closed')))::boolean as eligible
from sales.fulfilments f
join sales.orders o on o.id = f.order_id
join sales.order_lines ol on ol.fulfilment_id = f.id
where f.seller_id = @seller_id
group by f.id, o.order_number
having exists (select 1 from finance.ledger_journals j where j.reference_type = 'order' and j.reference_id = f.order_id and j.kind = 'sale')
   and not exists (select 1 from finance.payout_items pi where pi.fulfilment_id = f.id)
order by f.delivered_at nulls last, f.created_at;

-- name: PayoutsPayableBalance :one
-- What TechShop owes the seller right now per the ledger (credits are negative, so negate).
select coalesce(-sum(e.amount_kobo), 0)::bigint from finance.ledger_entries e
join finance.ledger_accounts a on a.id = e.account_id
where a.seller_id = $1;

-- name: PayoutsInFlight :one
-- Money already committed to payouts that haven't been paid (or failed) yet.
select coalesce(sum(amount_kobo), 0)::bigint from finance.payouts
where seller_id = $1 and status in ('scheduled', 'processing');

-- name: PayoutsActiveHolds :many
select * from finance.payout_holds
where seller_id = $1 and released_at is null and (until is null or until > now())
order by created_at;

-- name: PayoutsCredentialChanges :one
-- Security changes on the shop owner's account in the last 24 hours (account-takeover guard).
select count(*)::int from identity.auth_events e join sellers.sellers s on s.owner_user_id = e.user_id
where s.id = $1 and e.kind in ('password_changed', 'phone_changed', 'email_changed', 'mfa_removed')
  and e.created_at > now() - interval '24 hours';

-- name: PayoutsOpenRiskCases :one
select count(*)::int from risk.risk_cases
where subject_type = 'seller' and subject_id = $1 and status in ('open', 'in_review', 'escalated');

-- name: PayoutsCandidateSellers :many
-- Marketplace sellers with an unpaid delivered fulfilment (the batch builder checks the rest).
select distinct s.id, s.display_name, s.status from sellers.sellers s
join sales.fulfilments f on f.seller_id = s.id
where s.type <> 'first_party' and f.status = 'delivered'
  and not exists (select 1 from finance.payout_items pi where pi.fulfilment_id = f.id)
order by s.display_name;

-- name: PayoutsDefaultBank :one
select * from sellers.seller_bank_accounts where seller_id = $1 and is_default;

-- name: PayoutsCreateBatch :one
insert into finance.payout_batches (prepared_by) values ($1) returning *;

-- name: PayoutsCreatePayout :one
insert into finance.payouts (batch_id, seller_id, bank_account_id, amount_kobo, scheduled_for)
values ($1, $2, $3, $4, $5) returning *;

-- name: PayoutsAddItem :exec
insert into finance.payout_items (payout_id, fulfilment_id, amount_kobo) values ($1, $2, $3);

-- name: PayoutsRefreshBatchTotals :one
update finance.payout_batches b set
  total_kobo = (select coalesce(sum(amount_kobo), 0) from finance.payouts p where p.batch_id = b.id),
  payout_count = (select count(*) from finance.payouts p where p.batch_id = b.id)
where b.id = $1 returning *;

-- name: PayoutsGetBatch :one
select * from finance.payout_batches where id = $1;

-- name: PayoutsGetBatchForUpdate :one
select * from finance.payout_batches where id = $1 for update;

-- name: PayoutsListBatches :many
select * from finance.payout_batches
where (sqlc.narg('status')::text is null or status = sqlc.narg('status'))
  and (sqlc.narg('before')::timestamptz is null or created_at < sqlc.narg('before'))
order by created_at desc limit @lim;

-- name: PayoutsSubmitBatch :one
update finance.payout_batches set status = 'submitted', submitted_at = now() where id = $1 and status = 'draft' returning *;

-- name: PayoutsApproveBatch :one
update finance.payout_batches set status = 'approved', approved_by = @approver, approved_at = now()
where id = @id and status = 'submitted' returning *;

-- name: PayoutsSecondApprove :one
update finance.payout_batches set second_approved_by = @approver, second_approved_at = now()
where id = @id and status = 'submitted' returning *;

-- name: PayoutsSetBatchStatus :one
update finance.payout_batches set status = $2 where id = $1 returning *;

-- name: PayoutsBatchPayouts :many
select p.*, s.display_name, b.account_number_last4, b.bank_code, b.account_name
from finance.payouts p join sellers.sellers s on s.id = p.seller_id
join sellers.seller_bank_accounts b on b.id = p.bank_account_id
where p.batch_id = $1 order by s.display_name;

-- name: PayoutsGet :one
select * from finance.payouts where id = $1;

-- name: PayoutsGetForUpdate :one
select * from finance.payouts where id = $1 for update;

-- name: PayoutsDelete :exec
delete from finance.payouts where id = $1;

-- name: PayoutsSetProcessing :one
update finance.payouts set status = 'processing', provider_reference = @reference, transfer_code = coalesce(sqlc.narg('transfer_code'), transfer_code),
  attempts = attempts + 1
where id = @id and status in ('scheduled', 'processing') returning *;

-- name: PayoutsMarkPaid :one
update finance.payouts set status = 'paid', paid_at = now(), transfer_code = coalesce(sqlc.narg('transfer_code'), transfer_code)
where id = @id and status = 'processing' returning *;

-- name: PayoutsMarkFailed :one
update finance.payouts set status = 'failed', failure_reason = @reason where id = @id and status in ('scheduled', 'processing', 'paid') returning *;

-- name: PayoutsReleaseItems :exec
-- A failed payout gives its fulfilments back, so the next batch can pay them.
delete from finance.payout_items where payout_id = $1;

-- name: PayoutsStaleProcessing :many
select * from finance.payouts where status = 'processing' and updated_at < now() - interval '10 minutes' limit 100;

-- name: PayoutsBatchOpenCount :one
select count(*)::int from finance.payouts where batch_id = $1 and status in ('scheduled', 'processing');

-- name: PayoutsSetRecipientCode :exec
update sellers.seller_bank_accounts set provider = 'paystack', provider_recipient_code = $2 where id = $1;

-- name: PayoutsForSeller :many
select p.*, b.account_number_last4 from finance.payouts p join sellers.seller_bank_accounts b on b.id = p.bank_account_id
where p.seller_id = @seller_id and (sqlc.narg('before')::timestamptz is null or p.created_at < sqlc.narg('before'))
order by p.created_at desc limit @lim;

-- name: PayoutsItems :many
select pi.fulfilment_id, pi.amount_kobo, o.order_number, f.delivered_at
from finance.payout_items pi join sales.fulfilments f on f.id = pi.fulfilment_id join sales.orders o on o.id = f.order_id
where pi.payout_id = $1 order by f.delivered_at;

-- name: PayoutsCreateHold :one
insert into finance.payout_holds (seller_id, reason, note, until, created_by) values ($1, $2, $3, $4, $5) returning *;

-- name: PayoutsReleaseHold :one
update finance.payout_holds set released_at = now() where id = $1 and released_at is null returning *;

-- name: PayoutsIsHoliday :one
select exists (select 1 from platform.public_holidays where date = $1)::boolean;

-- name: PayoutsBankAccount :one
select * from sellers.seller_bank_accounts where id = $1;

-- name: PayoutsDevAgeSeller :exec
-- Development only (dev route): moves a seller's deliveries past the return window and their
-- bank accounts past the 24-hour hold, so payouts can be tested without waiting a week.
with d as (
  update sales.fulfilments f set delivered_at = now() - interval '8 days'
  where f.seller_id = @seller_id::uuid and f.status = 'delivered'
)
update sellers.seller_bank_accounts b set created_at = now() - interval '2 days' where b.seller_id = @seller_id::uuid;

-- name: PayoutsRiskFeatures :one
-- What the payout risk rules look at: the 30-day average paid out, and the share of the shop's
-- paid orders in 90 days that drew a chargeback.
select coalesce((select avg(po.amount_kobo) from finance.payouts po where po.seller_id = @seller_id::uuid and po.status = 'paid'
                   and po.paid_at > now() - interval '30 days'), 0)::float8 as avg_paid_30d,
  coalesce((select count(distinct d.id) from payments.disputes d join payments.payments p on p.id = d.payment_id
     join sales.fulfilments f on f.order_id = p.order_id where f.seller_id = @seller_id::uuid and d.created_at > now() - interval '90 days'), 0)::int as chargebacks_90d,
  coalesce((select count(*) from sales.fulfilments f join sales.orders o on o.id = f.order_id
     where f.seller_id = @seller_id::uuid and o.placed_at > now() - interval '90 days' and o.status not in ('pending_payment', 'cancelled')), 0)::int as orders_90d;
