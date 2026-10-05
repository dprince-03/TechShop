-- Logistics: delivery zones and rates, riders (devices, zones, shifts, locations), delivery and
-- pickup jobs with delivery codes, and the tracking trail. Plan: docs/mobile.md (logistics app).

-- +goose Up
create table logistics.delivery_zones (
  id          uuid primary key default gen_random_uuid(),
  name        text not null unique,
  state_code  text not null references identity.nigerian_states (code),
  lgas        text[] not null default '{}',
  is_active   boolean not null default true,
  created_at  timestamptz not null default now()
);
comment on table logistics.delivery_zones is '[logistics] Delivery areas (state + LGAs) used for fees, ETAs and rider assignment.';

create table logistics.delivery_rates (
  id                uuid primary key default gen_random_uuid(),
  zone_id           uuid not null references logistics.delivery_zones (id) on delete cascade,
  fulfilled_by      text not null check (fulfilled_by in ('techshop', 'seller')),
  max_weight_grams  integer not null check (max_weight_grams > 0),
  fee_kobo          bigint not null check (fee_kobo >= 0),
  eta_min_days      smallint not null check (eta_min_days >= 0),
  eta_max_days      smallint not null,
  valid_from        date not null default current_date,
  unique (zone_id, fulfilled_by, max_weight_grams, valid_from),
  check (eta_max_days >= eta_min_days)
);
comment on table logistics.delivery_rates is '[logistics] Delivery fee and ETA per zone and weight band.';

create table logistics.riders (
  user_id            uuid primary key references identity.staff_members (user_id),
  vehicle_type       text not null check (vehicle_type in ('bike', 'car', 'van')),
  plate_number       text unique,
  home_warehouse_id  uuid not null references inventory.warehouses (id),
  status             text not null default 'off_shift' check (status in ('off_shift', 'available', 'on_delivery', 'suspended')),
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now()
);
comment on table logistics.riders is '[logistics] Dispatch riders (staff using the logistics app).';

create table logistics.rider_devices (
  id           uuid primary key default gen_random_uuid(),
  rider_id     uuid not null references logistics.riders (user_id) on delete cascade,
  device_id    text not null,
  platform     text not null check (platform in ('ios', 'android')),
  approved_by  uuid references identity.users (id),
  approved_at  timestamptz,
  revoked_at   timestamptz,
  created_at   timestamptz not null default now(),
  unique (rider_id, device_id),
  check (approved_by is distinct from rider_id)
);
comment on table logistics.rider_devices is '[logistics] Phones bound to a rider; a dispatcher approves each one and can revoke a lost phone.';

create table logistics.rider_zones (
  rider_id  uuid not null references logistics.riders (user_id) on delete cascade,
  zone_id   uuid not null references logistics.delivery_zones (id) on delete cascade,
  primary key (rider_id, zone_id)
);
comment on table logistics.rider_zones is '[logistics] Zones each rider works in (assignment by zone).';

create table logistics.rider_shifts (
  id                   uuid primary key default gen_random_uuid(),
  rider_id             uuid not null references logistics.riders (user_id),
  device_session_id    uuid references identity.user_sessions (id),
  started_at           timestamptz not null default now(),
  ended_at             timestamptz,
  start_latitude       numeric(9, 6),
  start_longitude      numeric(9, 6),
  location_consent_at  timestamptz not null,
  check (ended_at is null or ended_at > started_at)
);
comment on table logistics.rider_shifts is '[logistics] Shift history; location is only shared while a shift is open and with consent.';
create unique index rider_shifts_one_open on logistics.rider_shifts (rider_id) where ended_at is null;

create table logistics.delivery_jobs (
  id                   uuid primary key default gen_random_uuid(),
  kind                 text not null default 'delivery' check (kind in ('delivery', 'return_pickup', 'seller_pickup')),
  fulfilment_id        uuid references sales.fulfilments (id),
  return_request_id    uuid,
  rider_id             uuid references logistics.riders (user_id),
  zone_id              uuid not null references logistics.delivery_zones (id),
  status               text not null default 'unassigned' check (status in ('unassigned', 'assigned', 'picked_up', 'en_route',
                         'delivered', 'failed', 'returned', 'cancelled')),
  pickup_warehouse_id  uuid references inventory.warehouses (id),
  pickup_address       jsonb,
  dropoff_latitude     numeric(9, 6),
  dropoff_longitude    numeric(9, 6),
  route_sequence       smallint,
  window_start         timestamptz,
  window_end           timestamptz,
  attempts             smallint not null default 0 check (attempts between 0 and 5),
  assigned_by          uuid references identity.users (id),
  assigned_at          timestamptz,
  delivered_at         timestamptz,
  recipient_name       text,
  otp_hash             bytea,
  otp_expires_at       timestamptz,
  otp_attempts         smallint not null default 0 check (otp_attempts between 0 and 5),
  otp_verified_at      timestamptz,
  proof_method         text check (proof_method in ('otp', 'photo', 'photo_offline')),
  proof_file_id        uuid references platform.files (id),
  created_at           timestamptz not null default now(),
  updated_at           timestamptz not null default now(),
  check (window_end is null or window_end > window_start),
  check (status in ('unassigned', 'cancelled') or rider_id is not null),
  check (kind <> 'delivery' or fulfilment_id is not null),
  check (kind <> 'return_pickup' or return_request_id is not null),
  check (pickup_warehouse_id is not null or pickup_address is not null),
  -- Delivered only with a verified code or a photo proof (fixes "delivered without checking the OTP").
  check (status <> 'delivered' or (delivered_at is not null and (
           (proof_method = 'otp' and otp_verified_at is not null) or
           (proof_method in ('photo', 'photo_offline') and proof_file_id is not null))))
);
comment on table logistics.delivery_jobs is '[logistics] A delivery or pickup run. Delivered requires a verified delivery code or photo proof. The delivery code lives only here.';
create index delivery_jobs_rider_status_idx on logistics.delivery_jobs (rider_id, status);
create index delivery_jobs_open_idx on logistics.delivery_jobs (zone_id, status) where status in ('unassigned', 'assigned');

create table logistics.delivery_events (
  id               uuid primary key default gen_random_uuid(),
  delivery_job_id  uuid not null references logistics.delivery_jobs (id) on delete cascade,
  client_event_id  uuid unique,
  kind             text not null check (kind in ('assigned', 'picked_up', 'en_route', 'location', 'attempt_failed',
                     'delivered', 'returned', 'note')),
  reason           text check (reason in ('customer_unreachable', 'customer_refused', 'wrong_address', 'rescheduled', 'unsafe', 'other')),
  latitude         numeric(9, 6),
  longitude        numeric(9, 6),
  accuracy_m       real,
  note             text,
  file_id          uuid references platform.files (id),
  created_by       uuid references identity.users (id),
  occurred_at      timestamptz not null default now(),
  created_at       timestamptz not null default now(),
  check (kind <> 'attempt_failed' or reason is not null)
);
comment on table logistics.delivery_events is '[logistics] Append-only trail for a job. client_event_id makes offline syncs from the app idempotent.';
create trigger delivery_events_append_only before update or delete on logistics.delivery_events
  for each row execute function platform.forbid_change();

create table logistics.rider_locations (
  rider_id     uuid primary key references logistics.riders (user_id) on delete cascade,
  latitude     numeric(9, 6) not null,
  longitude    numeric(9, 6) not null,
  accuracy_m   real,
  job_id       uuid references logistics.delivery_jobs (id) on delete set null,
  recorded_at  timestamptz not null
);
comment on table logistics.rider_locations is '[logistics] Latest known position per rider (live tracking and the dispatch map).';

create table logistics.rider_location_pings (
  rider_id     uuid not null references logistics.riders (user_id),
  latitude     numeric(9, 6) not null,
  longitude    numeric(9, 6) not null,
  accuracy_m   real,
  speed        real,
  heading      real,
  recorded_at  timestamptz not null,
  received_at  timestamptz not null default now(),
  primary key (rider_id, recorded_at)
) partition by range (recorded_at);
comment on table logistics.rider_location_pings is '[logistics] Full location trail while on shift (optional). Partitioned monthly; about 90 days retention.';
create table logistics.rider_location_pings_default partition of logistics.rider_location_pings default;

-- Indexes on foreign-key columns (every FK is indexed).
create index delivery_events_created_by_idx on logistics.delivery_events (created_by);
create index delivery_events_delivery_job_id_idx on logistics.delivery_events (delivery_job_id);
create index delivery_events_file_id_idx on logistics.delivery_events (file_id);
create index delivery_jobs_assigned_by_idx on logistics.delivery_jobs (assigned_by);
create index delivery_jobs_fulfilment_id_idx on logistics.delivery_jobs (fulfilment_id);
create index delivery_jobs_pickup_warehouse_id_idx on logistics.delivery_jobs (pickup_warehouse_id);
create index delivery_jobs_proof_file_id_idx on logistics.delivery_jobs (proof_file_id);
create index delivery_jobs_return_request_id_idx on logistics.delivery_jobs (return_request_id);
create index delivery_zones_state_code_idx on logistics.delivery_zones (state_code);
create index rider_devices_approved_by_idx on logistics.rider_devices (approved_by);
create index rider_locations_job_id_idx on logistics.rider_locations (job_id);
create index rider_shifts_device_session_id_idx on logistics.rider_shifts (device_session_id);
create index rider_zones_zone_id_idx on logistics.rider_zones (zone_id);
create index riders_home_warehouse_id_idx on logistics.riders (home_warehouse_id);

-- Keep updated_at current.
create trigger delivery_jobs_set_updated_at before update on logistics.delivery_jobs
  for each row execute function platform.set_updated_at();
create trigger riders_set_updated_at before update on logistics.riders
  for each row execute function platform.set_updated_at();

-- +goose Down
drop table if exists logistics.rider_location_pings, logistics.rider_locations, logistics.delivery_events,
  logistics.delivery_jobs, logistics.rider_shifts, logistics.rider_zones, logistics.rider_devices,
  logistics.riders, logistics.delivery_rates, logistics.delivery_zones cascade;
