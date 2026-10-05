-- Cars: listings (never added to the cart), photos, inspections, documents, viewing bookings
-- and financing enquiries.

-- +goose Up
create sequence cars.car_ref_seq start 1001;

create table cars.car_listings (
  id              uuid primary key default gen_random_uuid(),
  reference       text not null unique default platform.next_ref('CAR-', 'cars.car_ref_seq'),
  slug            text not null unique check (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
  seller_id       uuid not null references sellers.sellers (id),
  make            text not null,
  model           text not null,
  year            smallint not null check (year between 1950 and 2100),
  trim            text,
  body_type       text,
  transmission    text not null check (transmission in ('automatic', 'manual')),
  fuel_type       text check (fuel_type in ('petrol', 'diesel', 'hybrid', 'electric', 'cng')),
  mileage_km      integer not null check (mileage_km >= 0),
  condition       text not null check (condition in ('brand_new', 'foreign_used', 'nigerian_used')),
  vin             text unique check (vin ~ '^[A-HJ-NPR-Z0-9]{17}$'),
  colour          text,
  state_code      text not null references identity.nigerian_states (code),
  city            text not null,
  price_kobo      bigint not null check (price_kobo > 0),
  currency        char(3) not null default 'NGN',
  status          text not null default 'draft' check (status in ('draft', 'in_review', 'active', 'reserved', 'sold', 'withdrawn')),
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now(),
  check (condition <> 'brand_new' or mileage_km < 500)
);
comment on table cars.car_listings is '[cars] Cars for sale (separate from products): book a viewing, inspect, then pay. Never add-to-cart.';

create table cars.car_images (
  id              uuid primary key default gen_random_uuid(),
  car_listing_id  uuid not null references cars.car_listings (id) on delete cascade,
  file_id         uuid not null references platform.files (id),
  alt             text not null,
  position        integer not null default 0
);
comment on table cars.car_images is '[cars] Photos of a car listing.';

create table cars.car_inspections (
  id              uuid primary key default gen_random_uuid(),
  car_listing_id  uuid not null references cars.car_listings (id) on delete cascade,
  inspector_id    uuid references identity.staff_members (user_id),
  status          text not null default 'scheduled' check (status in ('scheduled', 'passed', 'passed_with_notes', 'failed')),
  checklist       jsonb,
  report_file_id  uuid references platform.files (id),
  inspected_at    timestamptz,
  created_at      timestamptz not null default now(),
  check (status = 'scheduled' or inspected_at is not null)
);
comment on table cars.car_inspections is '[cars] Inspection reports (engine, brakes, body, electricals, documents/VIN).';

create table cars.car_documents (
  id              uuid primary key default gen_random_uuid(),
  car_listing_id  uuid not null references cars.car_listings (id) on delete cascade,
  kind            text not null check (kind in ('customs_papers', 'proof_of_ownership', 'vehicle_licence', 'roadworthiness', 'insurance')),
  file_id         uuid not null references platform.files (id),
  verified_by     uuid references identity.users (id),
  verified_at     timestamptz,
  check ((verified_by is null) = (verified_at is null))
);
comment on table cars.car_documents is '[cars] Ownership and import documents, verified by staff. A listing goes active only with verified documents.';

create table cars.viewing_bookings (
  id                uuid primary key default gen_random_uuid(),
  car_listing_id    uuid not null references cars.car_listings (id),
  customer_user_id  uuid references identity.users (id),
  name              text not null,
  phone             text not null check (phone ~ '^\+234[0-9]{10}$'),
  preferred_date    date not null,
  slot              text not null check (slot in ('morning', 'afternoon', 'late_afternoon')),
  notes             text,
  status            text not null default 'requested' check (status in ('requested', 'confirmed', 'completed', 'no_show', 'cancelled')),
  created_at        timestamptz not null default now(),
  updated_at        timestamptz not null default now()
);
comment on table cars.viewing_bookings is '[cars] Requests to see a car (the Book a viewing form).';

create table cars.financing_enquiries (
  id                   uuid primary key default gen_random_uuid(),
  car_listing_id       uuid not null references cars.car_listings (id),
  customer_user_id     uuid references identity.users (id),
  name                 text not null,
  phone                text not null check (phone ~ '^\+234[0-9]{10}$'),
  email                citext,
  monthly_income_band  text,
  down_payment_kobo    bigint check (down_payment_kobo >= 0),
  status               text not null default 'new' check (status in ('new', 'contacted', 'referred', 'closed')),
  created_at           timestamptz not null default now(),
  updated_at           timestamptz not null default now()
);
comment on table cars.financing_enquiries is '[cars] Car financing enquiries, referred to finance partners.';

-- Indexes on foreign-key columns (every FK is indexed).
create index car_documents_car_listing_id_idx on cars.car_documents (car_listing_id);
create index car_documents_file_id_idx on cars.car_documents (file_id);
create index car_documents_verified_by_idx on cars.car_documents (verified_by);
create index car_images_car_listing_id_idx on cars.car_images (car_listing_id);
create index car_images_file_id_idx on cars.car_images (file_id);
create index car_inspections_car_listing_id_idx on cars.car_inspections (car_listing_id);
create index car_inspections_inspector_id_idx on cars.car_inspections (inspector_id);
create index car_inspections_report_file_id_idx on cars.car_inspections (report_file_id);
create index car_listings_seller_id_idx on cars.car_listings (seller_id);
create index car_listings_state_code_idx on cars.car_listings (state_code);
create index financing_enquiries_car_listing_id_idx on cars.financing_enquiries (car_listing_id);
create index financing_enquiries_customer_user_id_idx on cars.financing_enquiries (customer_user_id);
create index viewing_bookings_car_listing_id_idx on cars.viewing_bookings (car_listing_id);
create index viewing_bookings_customer_user_id_idx on cars.viewing_bookings (customer_user_id);

-- Keep updated_at current.
create trigger car_listings_set_updated_at before update on cars.car_listings
  for each row execute function platform.set_updated_at();
create trigger financing_enquiries_set_updated_at before update on cars.financing_enquiries
  for each row execute function platform.set_updated_at();
create trigger viewing_bookings_set_updated_at before update on cars.viewing_bookings
  for each row execute function platform.set_updated_at();

-- +goose Down
drop table if exists cars.financing_enquiries, cars.viewing_bookings, cars.car_documents,
  cars.car_inspections, cars.car_images, cars.car_listings cascade;
drop sequence if exists cars.car_ref_seq;
