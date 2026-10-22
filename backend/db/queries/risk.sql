-- Risk engine, cases and the device blocklist. Plan: trust-safety.md §3–5.

-- name: RiskActiveRuleSet :one
select * from risk.risk_rule_sets where checkpoint = $1 and status = 'active';

-- name: RiskRuleSet :one
select * from risk.risk_rule_sets where id = $1;

-- name: RiskRuleSetForUpdate :one
select * from risk.risk_rule_sets where id = $1 for update;

-- name: RiskRuleSets :many
-- The active rule set and any drafts per checkpoint, plus the last few retired ones.
select * from risk.risk_rule_sets
where status in ('active', 'draft') or (status = 'retired' and created_at > now() - interval '180 days')
order by checkpoint, version desc;

-- name: RiskRules :many
select * from risk.risk_rules where rule_set_id = $1 order by weight desc, key;

-- name: RiskNextVersion :one
select (coalesce(max(version), 0) + 1)::int from risk.risk_rule_sets where checkpoint = $1;

-- name: RiskDraftFor :one
select count(*)::int from risk.risk_rule_sets where checkpoint = $1 and status = 'draft';

-- name: RiskCreateRuleSet :one
insert into risk.risk_rule_sets (checkpoint, version, thresholds, created_by, notes)
values ($1, $2, $3, $4, $5)
returning *;

-- name: RiskCopyRules :exec
insert into risk.risk_rules (rule_set_id, key, label, condition, weight, action, enabled)
select @target::uuid, key, label, condition, weight, action, enabled from risk.risk_rules where rule_set_id = @source::uuid;

-- name: RiskUpdateRule :one
update risk.risk_rules set weight = $3, enabled = $4 where id = $1 and rule_set_id = $2 returning *;

-- name: RiskSetThresholds :exec
update risk.risk_rule_sets set thresholds = $2 where id = $1 and status = 'draft';

-- name: RiskRetireActive :exec
update risk.risk_rule_sets set status = 'retired' where checkpoint = $1 and status = 'active';

-- name: RiskPublish :exec
update risk.risk_rule_sets set status = 'active', published_by = $2, published_at = now() where id = $1 and status = 'draft';

-- name: RiskDeleteDraft :exec
delete from risk.risk_rule_sets where id = $1 and status = 'draft';

-- name: RiskInsertDecision :one
insert into risk.risk_decisions (checkpoint, subject_type, subject_id, user_id, score, outcome, reasons, rule_set_version, latency_ms)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
returning id;

-- name: RiskRecentDecisions :many
-- Decisions of one checkpoint in the window, with the label of any case they opened (test bench).
select d.id, d.score, d.outcome, d.reasons, c.label
from risk.risk_decisions d
left join risk.risk_cases c on c.decision_id = d.id
where d.checkpoint = $1 and d.created_at > $2
order by d.created_at desc
limit 5000;

-- name: RiskDecision :one
select * from risk.risk_decisions where id = $1;

-- name: RiskOpenCase :one
insert into risk.risk_cases (kind, subject_type, subject_id, reason, score, money_at_risk_kobo, decision_id, sla_due_at)
values ($1, $2, $3, $4, $5, $6, $7, $8)
returning *;

-- name: RiskCaseEvent :exec
insert into risk.case_events (case_id, actor_id, action, note) values ($1, $2, $3, $4);

-- name: RiskCases :many
-- The queue: open work first by score × money at risk, then everything else newest first.
select c.*, u.first_name as assignee_first_name, u.last_name as assignee_last_name
from risk.risk_cases c
left join identity.users u on u.id = c.assigned_to
where (sqlc.narg('status')::text is null or c.status = sqlc.narg('status')
       or (sqlc.narg('status') = 'active' and c.status in ('open', 'in_review', 'escalated')))
  and (sqlc.narg('kind')::text is null or c.kind = sqlc.narg('kind'))
order by (c.status in ('open', 'in_review', 'escalated')) desc,
         coalesce(c.score, 0)::bigint * greatest(coalesce(c.money_at_risk_kobo, 0), 100000) desc,
         c.created_at desc
limit 200;

-- name: RiskCaseByRef :one
select c.*, u.first_name as assignee_first_name, u.last_name as assignee_last_name
from risk.risk_cases c left join identity.users u on u.id = c.assigned_to
where c.reference = $1;

-- name: RiskCaseForUpdate :one
select * from risk.risk_cases where reference = $1 for update;

-- name: RiskCaseEvents :many
select e.*, u.first_name, u.last_name from risk.case_events e left join identity.users u on u.id = e.actor_id
where e.case_id = $1 order by e.created_at;

-- name: RiskClaimCase :one
update risk.risk_cases set status = 'in_review', assigned_to = $2 where id = $1 returning *;

-- name: RiskEscalateCase :one
update risk.risk_cases set status = 'escalated' where id = $1 returning *;

-- name: RiskResolveCase :one
update risk.risk_cases set status = $2, label = $3, resolved_by = $4, resolved_at = now(), resolution_note = $5
where id = $1 returning *;

-- name: RiskCaseCounts :one
select
  count(*) filter (where status in ('open', 'in_review', 'escalated'))::int as active,
  count(*) filter (where status in ('open', 'in_review', 'escalated') and kind = 'order')::int as orders,
  count(*) filter (where status in ('open', 'in_review', 'escalated') and sla_due_at < now())::int as overdue,
  count(*) filter (where resolved_at > now() - interval '7 days')::int as resolved_7d
from risk.risk_cases;

-- name: RiskOpenCasesForSubject :one
select count(*)::int from risk.risk_cases where subject_type = $1 and subject_id = $2 and status in ('open', 'in_review', 'escalated');

-- name: RiskBlocklist :many
select b.*, u.first_name as created_by_first_name from risk.device_blocklist b
left join identity.users u on u.id = b.created_by
where (sqlc.narg('imei')::text is null or b.imei = sqlc.narg('imei'))
order by b.created_at desc limit 200;

-- name: RiskBlockDevice :one
insert into risk.device_blocklist (imei, reason, source, reference, created_by) values ($1, $2, $3, $4, $5)
on conflict (imei) do nothing
returning *;

-- name: RiskUnblockDevice :one
delete from risk.device_blocklist where imei = $1 returning *;

-- name: RiskDeviceBlocked :one
select b.reason from risk.device_blocklist b where b.imei = $1;

-- name: RiskBlocklistCount :one
select count(*)::int from risk.device_blocklist;

-- name: RiskSubjectLabel :one
-- A readable name for a case's subject: order or return number, trade-in reference, shop, or a
-- masked person (first name and last initial).
select coalesce(
  (select o.order_number from sales.orders o where o.id = @id::uuid),
  (select r.rma_number from aftersales.return_requests r where r.id = @id::uuid),
  (select t.reference from aftersales.trade_ins t where t.id = @id::uuid),
  (select s.display_name from sellers.sellers s where s.id = @id::uuid),
  (select u.first_name || ' ' || left(u.last_name, 1) || '.' from identity.users u where u.id = @id::uuid),
  '')::text as label;
