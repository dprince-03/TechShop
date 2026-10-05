-- Logistics: zones, rates, riders, devices, shifts, delivery jobs, events, locations (docs/mobile.md §4).

-- name: LogisticsListZones :many
select * from logistics.delivery_zones order by state_code, name;

-- name: LogisticsCreateZone :one
insert into logistics.delivery_zones (name, state_code, lgas, is_active) values ($1, $2, $3, $4) returning *;

-- name: LogisticsUpdateZone :one
update logistics.delivery_zones set name = $2, lgas = $3, is_active = $4 where id = $1 returning *;

-- name: LogisticsResolveZone :one
-- The zone for a state and LGA: an LGA-specific zone first, then the state-wide zone (no LGAs listed).
select * from logistics.delivery_zones
where is_active and state_code = $1 and (cardinality(lgas) = 0 or @lga::text = any(lgas))
order by cardinality(lgas) > 0 desc, name limit 1;

-- name: LogisticsCreateRate :one
insert into logistics.delivery_rates (zone_id, fulfilled_by, max_weight_grams, fee_kobo, eta_min_days, eta_max_days, valid_from)
values ($1, $2, $3, $4, $5, $6, $7) returning *;

-- name: LogisticsZoneRates :many
select * from logistics.delivery_rates where zone_id = $1 order by fulfilled_by, max_weight_grams, valid_from desc;

-- name: LogisticsRate :one
-- The current rate band for a parcel: the smallest weight band that fits, newest valid version.
select * from logistics.delivery_rates
where zone_id = $1 and fulfilled_by = $2 and max_weight_grams >= @weight::int and valid_from <= current_date
order by max_weight_grams, valid_from desc limit 1;

-- name: LogisticsCreateRider :one
insert into logistics.riders (user_id, vehicle_type, plate_number, home_warehouse_id) values ($1, $2, $3, $4) returning *;

-- name: LogisticsGetRider :one
select * from logistics.riders where user_id = $1;

-- name: LogisticsSetRiderStatus :exec
update logistics.riders set status = $2 where user_id = $1;

-- name: LogisticsListRiders :many
-- Riders with their latest position, open shift and active job count (for the assign sheet and map).
select r.user_id, r.vehicle_type, r.plate_number, r.home_warehouse_id, r.status, u.first_name, u.last_name, u.phone,
       l.latitude, l.longitude, l.recorded_at as last_seen_at,
       (select s.id from logistics.rider_shifts s where s.rider_id = r.user_id and s.ended_at is null) as open_shift_id,
       (select count(*) from logistics.delivery_jobs j where j.rider_id = r.user_id and j.status in ('assigned', 'picked_up', 'en_route'))::int as active_jobs,
       coalesce((select array_agg(z.zone_id) from logistics.rider_zones z where z.rider_id = r.user_id), '{}')::uuid[] as zone_ids
from logistics.riders r join identity.users u on u.id = r.user_id
left join logistics.rider_locations l on l.rider_id = r.user_id
order by r.status, u.first_name;

-- name: LogisticsSetRiderZones :exec
delete from logistics.rider_zones where rider_id = $1;

-- name: LogisticsAddRiderZone :exec
insert into logistics.rider_zones (rider_id, zone_id) values ($1, $2) on conflict do nothing;

-- name: LogisticsRegisterDevice :one
insert into logistics.rider_devices (rider_id, device_id, platform) values ($1, $2, $3)
on conflict (rider_id, device_id) do update set platform = excluded.platform
returning *;

-- name: LogisticsApproveDevice :one
update logistics.rider_devices set approved_by = $2, approved_at = now(), revoked_at = null where id = $1 returning *;

-- name: LogisticsRevokeDevice :one
update logistics.rider_devices set revoked_at = now() where id = $1 returning *;

-- name: LogisticsListDevices :many
select d.*, u.first_name, u.last_name from logistics.rider_devices d join identity.users u on u.id = d.rider_id
where (sqlc.narg('pending')::bool is not true or (d.approved_at is null and d.revoked_at is null))
order by d.created_at desc limit 200;

-- name: LogisticsDeviceApproved :one
select exists (select 1 from logistics.rider_devices where rider_id = $1 and device_id = $2 and approved_at is not null and revoked_at is null)::bool;

-- name: LogisticsOpenShift :one
select * from logistics.rider_shifts where rider_id = $1 and ended_at is null;

-- name: LogisticsStartShift :one
insert into logistics.rider_shifts (rider_id, device_session_id, start_latitude, start_longitude, location_consent_at)
values ($1, $2, $3, $4, now()) returning *;

-- name: LogisticsEndShift :one
update logistics.rider_shifts set ended_at = now() where rider_id = $1 and ended_at is null returning *;

-- name: LogisticsStaleShifts :many
-- Shifts open for more than 12 hours are closed automatically (location tracking stops).
select * from logistics.rider_shifts where ended_at is null and started_at < now() - interval '12 hours';

-- name: LogisticsCreateJob :one
insert into logistics.delivery_jobs (kind, fulfilment_id, return_request_id, zone_id, pickup_warehouse_id, pickup_address,
  dropoff_latitude, dropoff_longitude, window_start, window_end, recipient_name)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning *;

-- name: LogisticsGetJob :one
select * from logistics.delivery_jobs where id = $1;

-- name: LogisticsGetJobForUpdate :one
select * from logistics.delivery_jobs where id = $1 for update;

-- name: LogisticsJobForFulfilment :one
select * from logistics.delivery_jobs where fulfilment_id = $1 and kind = 'delivery' and status <> 'cancelled' order by created_at desc limit 1;

-- name: LogisticsAssignJob :one
update logistics.delivery_jobs set rider_id = $2, status = 'assigned', assigned_by = $3, assigned_at = now(), route_sequence = $4
where id = $1 and status in ('unassigned', 'assigned', 'failed') returning *;

-- name: LogisticsUnassignJob :one
update logistics.delivery_jobs set rider_id = null, status = 'unassigned', assigned_by = null, assigned_at = null
where id = $1 and status = 'assigned' returning *;

-- name: LogisticsSetJobStatus :one
update logistics.delivery_jobs
set status = $2,
    attempts = case when $2 = 'failed' then attempts + 1 else attempts end,
    delivered_at = case when $2 = 'delivered' then now() else delivered_at end,
    proof_method = coalesce(sqlc.narg('proof_method'), proof_method),
    proof_file_id = coalesce(sqlc.narg('proof_file_id'), proof_file_id),
    recipient_name = coalesce(sqlc.narg('recipient_name'), recipient_name)
where id = $1 returning *;

-- name: LogisticsSetJobOTP :exec
update logistics.delivery_jobs set otp_hash = $2, otp_expires_at = $3, otp_attempts = 0, otp_verified_at = null where id = $1;

-- name: LogisticsOTPFailed :one
update logistics.delivery_jobs set otp_attempts = otp_attempts + 1 where id = $1 returning otp_attempts;

-- name: LogisticsOTPVerified :exec
update logistics.delivery_jobs set otp_verified_at = now() where id = $1;

-- name: LogisticsRiderJobs :many
-- A rider's work: active jobs plus today's finished ones, with what the job screen needs.
select j.*, f.order_id, o.order_number, o.ship_to, o.contact_phone, w.name as pickup_warehouse_name, w.address as pickup_warehouse_address
from logistics.delivery_jobs j
left join sales.fulfilments f on f.id = j.fulfilment_id
left join sales.orders o on o.id = f.order_id
left join inventory.warehouses w on w.id = j.pickup_warehouse_id
where j.rider_id = $1 and (j.status in ('assigned', 'picked_up', 'en_route', 'failed') or j.updated_at > now() - interval '24 hours')
order by j.route_sequence nulls last, j.window_start nulls last, j.created_at;

-- name: LogisticsJobView :one
select j.*, f.order_id, o.order_number, o.ship_to, o.contact_phone, o.customer_user_id, w.name as pickup_warehouse_name, w.address as pickup_warehouse_address
from logistics.delivery_jobs j
left join sales.fulfilments f on f.id = j.fulfilment_id
left join sales.orders o on o.id = f.order_id
left join inventory.warehouses w on w.id = j.pickup_warehouse_id
where j.id = $1;

-- name: LogisticsBoard :many
select j.*, z.name as zone_name, o.order_number, u.first_name as rider_first_name
from logistics.delivery_jobs j
join logistics.delivery_zones z on z.id = j.zone_id
left join sales.fulfilments f on f.id = j.fulfilment_id
left join sales.orders o on o.id = f.order_id
left join identity.users u on u.id = j.rider_id
where (sqlc.narg('status')::text is null or j.status = sqlc.narg('status'))
  and (sqlc.narg('zone_id')::uuid is null or j.zone_id = sqlc.narg('zone_id'))
  and (j.status not in ('delivered', 'returned', 'cancelled') or j.updated_at > now() - interval '24 hours')
order by j.window_start nulls last, j.created_at limit 500;

-- name: LogisticsBoardCounts :one
select count(*) filter (where status = 'unassigned')::int as unassigned,
       count(*) filter (where status = 'assigned')::int as assigned,
       count(*) filter (where status in ('picked_up', 'en_route'))::int as in_transit,
       count(*) filter (where status = 'failed')::int as failed,
       count(*) filter (where status = 'delivered' and delivered_at > date_trunc('day', now()))::int as delivered_today
from logistics.delivery_jobs;

-- name: LogisticsInsertEvent :one
-- Idempotent per client_event_id (offline sync): a replay returns no row.
insert into logistics.delivery_events (delivery_job_id, client_event_id, kind, reason, latitude, longitude, accuracy_m, note, file_id, created_by, occurred_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
on conflict (client_event_id) do nothing
returning *;

-- name: LogisticsJobEvents :many
select * from logistics.delivery_events where delivery_job_id = $1 order by occurred_at, created_at;

-- name: LogisticsUpsertLocation :exec
insert into logistics.rider_locations (rider_id, latitude, longitude, accuracy_m, job_id, recorded_at) values ($1, $2, $3, $4, $5, $6)
on conflict (rider_id) do update set latitude = excluded.latitude, longitude = excluded.longitude, accuracy_m = excluded.accuracy_m,
  job_id = excluded.job_id, recorded_at = excluded.recorded_at
where logistics.rider_locations.recorded_at < excluded.recorded_at;

-- name: LogisticsInsertPing :exec
insert into logistics.rider_location_pings (rider_id, latitude, longitude, accuracy_m, speed, heading, recorded_at) values ($1, $2, $3, $4, $5, $6, $7)
on conflict do nothing;

-- name: LogisticsEnRouteJobs :many
select j.id, f.order_id from logistics.delivery_jobs j join sales.fulfilments f on f.id = j.fulfilment_id
where j.rider_id = $1 and j.status = 'en_route';

-- name: LogisticsFulfilmentDeliveryInfo :one
-- Everything needed to create a delivery job for a packed fulfilment.
select f.id, f.order_id, f.seller_id, f.fulfilled_by, f.method, f.warehouse_id, o.ship_to, o.order_number,
  (select r.warehouse_id from inventory.stock_reservations r where r.order_id = f.order_id and r.owner_seller_id = f.seller_id
     and r.warehouse_id is not null limit 1) as reserved_warehouse_id
from sales.fulfilments f join sales.orders o on o.id = f.order_id where f.id = $1;

-- name: LogisticsPurgePings :execrows
-- Raw location points are kept about 90 days (docs/mobile.md §4.4).
delete from logistics.rider_location_pings where recorded_at < now() - interval '90 days';
