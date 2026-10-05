-- =============================================================================
-- TechShop database schema (PostgreSQL 17). GENERATED: do not edit by hand.
--
-- Source of truth: backend/db/migrations (goose). This file is a pg_dump --schema-only
-- of a scratch database after `goose up`, kept so the whole schema can be read in one
-- place. Regenerate it with docs/database/tools/generate.sh after changing migrations.
-- =============================================================================

--
-- PostgreSQL database dump
--




--
-- Name: aftersales; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA aftersales;


--
-- Name: SCHEMA aftersales; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA aftersales IS 'Returns, inspections, warranty claims, repairs and trade-ins.';


--
-- Name: b2b; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA b2b;


--
-- Name: SCHEMA b2b; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA b2b IS 'Business buyers: organisations, members, credit, quotes.';


--
-- Name: cars; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA cars;


--
-- Name: SCHEMA cars; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA cars IS 'Car listings, inspections, documents, viewings and financing enquiries.';


--
-- Name: catalog; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA catalog;


--
-- Name: SCHEMA catalog; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA catalog IS 'Products, variants, listings, prices, reviews and the search read model.';


--
-- Name: content; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA content;


--
-- Name: SCHEMA content; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA content IS 'CMS pages, banners and help articles.';


--
-- Name: finance; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA finance;


--
-- Name: SCHEMA finance; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA finance IS 'The books and paying out: double-entry ledger, periods, commission, invoices, seller payouts.';


--
-- Name: fulfilment; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA fulfilment;


--
-- Name: SCHEMA fulfilment; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA fulfilment IS 'Warehouse work and shipping: pick lists, packing, parcels, manifests, carriers, seller SLAs.';


--
-- Name: identity; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA identity;


--
-- Name: SCHEMA identity; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA identity IS 'People and access: users, sessions, codes, MFA, roles, staff, addresses, consents.';


--
-- Name: inventory; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA inventory;


--
-- Name: SCHEMA inventory; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA inventory IS 'Stock: warehouses, levels, device units, movements, reservations, counts, costs.';


--
-- Name: logistics; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA logistics;


--
-- Name: SCHEMA logistics; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA logistics IS 'Delivery zones and rates, riders, delivery jobs and tracking.';


--
-- Name: marketing; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA marketing;


--
-- Name: SCHEMA marketing; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA marketing IS 'Promotions, coupons, flash deals, segments, campaigns, journeys and tracked links.';


--
-- Name: messaging; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA messaging;


--
-- Name: SCHEMA messaging; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA messaging IS 'Every message sent: notifications log, provider events, suppressions, preferences, templates, push tokens.';


--
-- Name: payments; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA payments;


--
-- Name: SCHEMA payments; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA payments IS 'Taking money: payments, provider webhooks, refunds, transfer accounts, disputes, reconciliation.';


--
-- Name: personalisation; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA personalisation;


--
-- Name: SCHEMA personalisation; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA personalisation IS 'Behaviour events and recommendation model outputs.';


--
-- Name: platform; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA platform;


--
-- Name: SCHEMA platform; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA platform IS 'Shared building blocks: files, audit log, outbox, idempotency, feature flags, helper functions.';


--
-- Name: pos; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA pos;


--
-- Name: SCHEMA pos; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA pos IS 'Physical stores: tills and cashier shifts.';


--
-- Name: purchasing; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA purchasing;


--
-- Name: SCHEMA purchasing; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA purchasing IS 'Buying stock from suppliers: purchase orders and goods receipts.';


--
-- Name: risk; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA risk;


--
-- Name: SCHEMA risk; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA risk IS 'Trust and safety: KYC checks, risk rules and decisions, cases, link graph, blocklists, seller enforcement.';


--
-- Name: sales; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA sales;


--
-- Name: SCHEMA sales; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA sales IS 'Carts, orders, fulfilments and order lines.';


--
-- Name: search; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA search;


--
-- Name: SCHEMA search; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA search IS 'Search merchandising and analytics: synonyms, redirects, pins, query logs.';


--
-- Name: sellers; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA sellers;


--
-- Name: SCHEMA sellers; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA sellers IS 'Marketplace sellers: shops, members, KYC, payout bank accounts.';


--
-- Name: support; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA support;


--
-- Name: SCHEMA support; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA support IS 'Support tickets and messages.';


--
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;


--
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';


--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: assert_journal_balanced(); Type: FUNCTION; Schema: finance; Owner: -
--

CREATE FUNCTION finance.assert_journal_balanced() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
declare total bigint;
begin
  select coalesce(sum(amount_kobo), 0) into total from finance.ledger_entries where journal_id = new.journal_id;
  if total <> 0 then
    raise exception 'ledger journal % is not balanced (sum = %)', new.journal_id, total;
  end if;
  return null;
end $$;


--
-- Name: assert_period_open(); Type: FUNCTION; Schema: finance; Owner: -
--

CREATE FUNCTION finance.assert_period_open() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
  if exists (select 1 from finance.accounting_periods where month = new.period and status = 'closed') then
    raise exception 'accounting period % is closed', to_char(new.period, 'YYYY-MM');
  end if;
  return new;
end $$;


--
-- Name: ensure_monthly_partitions(regclass, date, integer); Type: FUNCTION; Schema: platform; Owner: -
--

CREATE FUNCTION platform.ensure_monthly_partitions(parent regclass, from_date date, months integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
declare
  start_month date := date_trunc('month', from_date)::date;
  part_start date;
  part_name text;
  parent_schema text;
  parent_table text;
begin
  select n.nspname, c.relname into parent_schema, parent_table
  from pg_class c join pg_namespace n on n.oid = c.relnamespace
  where c.oid = parent;

  for i in 0 .. months - 1 loop
    part_start := (start_month + make_interval(months => i))::date;
    part_name := format('%s_%s', parent_table, to_char(part_start, 'YYYYMM'));
    execute format(
      'create table if not exists %I.%I partition of %s for values from (%L) to (%L)',
      parent_schema, part_name, parent, part_start, (part_start + interval '1 month')::date);
  end loop;
end $$;


--
-- Name: forbid_change(); Type: FUNCTION; Schema: platform; Owner: -
--

CREATE FUNCTION platform.forbid_change() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
  raise exception '%.% is append-only', tg_table_schema, tg_table_name;
end $$;


--
-- Name: next_ref(text, regclass); Type: FUNCTION; Schema: platform; Owner: -
--

CREATE FUNCTION platform.next_ref(prefix text, seq regclass) RETURNS text
    LANGUAGE sql
    AS $$ select prefix || nextval(seq)::text $$;


--
-- Name: notify_outbox(); Type: FUNCTION; Schema: platform; Owner: -
--

CREATE FUNCTION platform.notify_outbox() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
  perform pg_notify('outbox', '');
  return null;
end $$;


--
-- Name: set_updated_at(); Type: FUNCTION; Schema: platform; Owner: -
--

CREATE FUNCTION platform.set_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
  new.updated_at := now();
  return new;
end $$;


--
-- Name: claim_number_seq; Type: SEQUENCE; Schema: aftersales; Owner: -
--

CREATE SEQUENCE aftersales.claim_number_seq
    START WITH 1001
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: repair_jobs; Type: TABLE; Schema: aftersales; Owner: -
--

CREATE TABLE aftersales.repair_jobs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    warranty_claim_id uuid,
    device_unit_id uuid,
    technician_id uuid,
    status text DEFAULT 'queued'::text NOT NULL,
    diagnosis text,
    chargeable boolean DEFAULT false NOT NULL,
    cost_kobo bigint,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT repair_jobs_check CHECK (((warranty_claim_id IS NOT NULL) OR (device_unit_id IS NOT NULL))),
    CONSTRAINT repair_jobs_check1 CHECK (((NOT chargeable) OR (cost_kobo IS NOT NULL))),
    CONSTRAINT repair_jobs_cost_kobo_check CHECK ((cost_kobo >= 0)),
    CONSTRAINT repair_jobs_status_check CHECK ((status = ANY (ARRAY['queued'::text, 'diagnosing'::text, 'awaiting_parts'::text, 'repairing'::text, 'done'::text, 'unrepairable'::text])))
);


--
-- Name: TABLE repair_jobs; Type: COMMENT; Schema: aftersales; Owner: -
--

COMMENT ON TABLE aftersales.repair_jobs IS '[aftersales] Workshop jobs for warranty claims, paid repairs and refurbishment of returns.';


--
-- Name: return_inspections; Type: TABLE; Schema: aftersales; Owner: -
--

CREATE TABLE aftersales.return_inspections (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    return_item_id uuid NOT NULL,
    device_unit_id uuid,
    scanned_serial text,
    imei_match boolean,
    grade text NOT NULL,
    disposition text NOT NULL,
    notes text,
    inspected_by uuid NOT NULL,
    inspected_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT return_inspections_check CHECK (((device_unit_id IS NULL) OR (imei_match IS NOT NULL))),
    CONSTRAINT return_inspections_disposition_check CHECK ((disposition = ANY (ARRAY['restock_new'::text, 'restock_open_box'::text, 'repair'::text, 'return_to_seller'::text, 'return_to_supplier'::text, 'write_off'::text]))),
    CONSTRAINT return_inspections_grade_check CHECK ((grade = ANY (ARRAY['A'::text, 'B'::text, 'C'::text, 'faulty'::text])))
);


--
-- Name: TABLE return_inspections; Type: COMMENT; Schema: aftersales; Owner: -
--

COMMENT ON TABLE aftersales.return_inspections IS '[aftersales] Inspection of returned items: serial must match the unit sold (swap fraud), grade and disposition.';


--
-- Name: return_items; Type: TABLE; Schema: aftersales; Owner: -
--

CREATE TABLE aftersales.return_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    return_request_id uuid NOT NULL,
    order_line_id uuid NOT NULL,
    quantity integer NOT NULL,
    CONSTRAINT return_items_quantity_check CHECK ((quantity > 0))
);


--
-- Name: TABLE return_items; Type: COMMENT; Schema: aftersales; Owner: -
--

COMMENT ON TABLE aftersales.return_items IS '[aftersales] Which order lines (and how many) are being returned.';


--
-- Name: rma_number_seq; Type: SEQUENCE; Schema: aftersales; Owner: -
--

CREATE SEQUENCE aftersales.rma_number_seq
    START WITH 1001
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: return_requests; Type: TABLE; Schema: aftersales; Owner: -
--

CREATE TABLE aftersales.return_requests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    rma_number text DEFAULT platform.next_ref('RMA-'::text, 'aftersales.rma_number_seq'::regclass) NOT NULL,
    order_id uuid NOT NULL,
    requested_by uuid NOT NULL,
    reason text NOT NULL,
    details text,
    return_method text,
    status text DEFAULT 'requested'::text NOT NULL,
    auto_approved boolean DEFAULT false NOT NULL,
    resolution text,
    decided_by uuid,
    rejection_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT return_requests_check CHECK (((status <> 'rejected'::text) OR (rejection_reason IS NOT NULL))),
    CONSTRAINT return_requests_reason_check CHECK ((reason = ANY (ARRAY['faulty'::text, 'damaged'::text, 'not_as_described'::text, 'wrong_item'::text, 'change_of_mind'::text]))),
    CONSTRAINT return_requests_resolution_check CHECK ((resolution = ANY (ARRAY['refund'::text, 'replace'::text, 'repair'::text]))),
    CONSTRAINT return_requests_return_method_check CHECK ((return_method = ANY (ARRAY['pickup'::text, 'store_dropoff'::text, 'seller_return'::text]))),
    CONSTRAINT return_requests_status_check CHECK ((status = ANY (ARRAY['requested'::text, 'approved'::text, 'rejected'::text, 'awaiting_pickup'::text, 'received'::text, 'inspected'::text, 'refunded'::text, 'replaced'::text, 'closed'::text])))
);


--
-- Name: TABLE return_requests; Type: COMMENT; Schema: aftersales; Owner: -
--

COMMENT ON TABLE aftersales.return_requests IS '[aftersales] Customer return requests (RMA) and their outcome. Faulty items within 7 days may be auto-approved.';


--
-- Name: trade_in_ref_seq; Type: SEQUENCE; Schema: aftersales; Owner: -
--

CREATE SEQUENCE aftersales.trade_in_ref_seq
    START WITH 1001
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: trade_ins; Type: TABLE; Schema: aftersales; Owner: -
--

CREATE TABLE aftersales.trade_ins (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    reference text DEFAULT platform.next_ref('TI-'::text, 'aftersales.trade_in_ref_seq'::regclass) NOT NULL,
    user_id uuid NOT NULL,
    device_type text NOT NULL,
    brand text NOT NULL,
    model text NOT NULL,
    storage text,
    declared_condition text NOT NULL,
    imei text,
    estimate_kobo bigint,
    inspected_condition text,
    offer_kobo bigint,
    payout_method text,
    bank_code text,
    account_number_encrypted bytea,
    account_number_last4 text,
    encryption_key_id uuid,
    status text DEFAULT 'submitted'::text NOT NULL,
    inspected_by uuid,
    device_unit_id uuid,
    paid_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT trade_ins_account_number_last4_check CHECK ((length(account_number_last4) = 4)),
    CONSTRAINT trade_ins_check CHECK (((status <> ALL (ARRAY['offer_sent'::text, 'accepted'::text, 'paid'::text])) OR (offer_kobo IS NOT NULL))),
    CONSTRAINT trade_ins_check1 CHECK (((payout_method IS DISTINCT FROM 'bank'::text) OR (status <> 'paid'::text) OR (account_number_encrypted IS NOT NULL))),
    CONSTRAINT trade_ins_check2 CHECK (((status <> 'paid'::text) OR (paid_at IS NOT NULL))),
    CONSTRAINT trade_ins_declared_condition_check CHECK ((declared_condition = ANY (ARRAY['like_new'::text, 'good'::text, 'fair'::text, 'faulty'::text]))),
    CONSTRAINT trade_ins_device_type_check CHECK ((device_type = ANY (ARRAY['phone'::text, 'laptop'::text, 'tablet'::text, 'console'::text, 'smartwatch'::text]))),
    CONSTRAINT trade_ins_estimate_kobo_check CHECK ((estimate_kobo >= 0)),
    CONSTRAINT trade_ins_imei_check CHECK ((imei ~ '^[0-9]{15}$'::text)),
    CONSTRAINT trade_ins_inspected_condition_check CHECK ((inspected_condition = ANY (ARRAY['like_new'::text, 'good'::text, 'fair'::text, 'faulty'::text]))),
    CONSTRAINT trade_ins_offer_kobo_check CHECK ((offer_kobo >= 0)),
    CONSTRAINT trade_ins_payout_method_check CHECK ((payout_method = ANY (ARRAY['bank'::text, 'store_credit'::text]))),
    CONSTRAINT trade_ins_status_check CHECK ((status = ANY (ARRAY['submitted'::text, 'estimated'::text, 'awaiting_inspection'::text, 'inspected'::text, 'offer_sent'::text, 'accepted'::text, 'declined'::text, 'paid'::text, 'cancelled'::text, 'blocked'::text])))
);


--
-- Name: TABLE trade_ins; Type: COMMENT; Schema: aftersales; Owner: -
--

COMMENT ON TABLE aftersales.trade_ins IS '[aftersales] Trade-ins from estimate to inspection, offer and payout (bank or store credit); accepted devices become device units.';


--
-- Name: warranty_claims; Type: TABLE; Schema: aftersales; Owner: -
--

CREATE TABLE aftersales.warranty_claims (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    claim_number text DEFAULT platform.next_ref('WR-'::text, 'aftersales.claim_number_seq'::regclass) NOT NULL,
    customer_user_id uuid NOT NULL,
    order_line_id uuid,
    device_unit_id uuid,
    imei_or_serial text NOT NULL,
    fault_description text NOT NULL,
    status text DEFAULT 'new'::text NOT NULL,
    resolution text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT warranty_claims_resolution_check CHECK ((resolution = ANY (ARRAY['repaired'::text, 'replaced'::text, 'refunded'::text, 'not_covered'::text]))),
    CONSTRAINT warranty_claims_status_check CHECK ((status = ANY (ARRAY['new'::text, 'received'::text, 'in_repair'::text, 'ready'::text, 'collected'::text, 'rejected'::text])))
);


--
-- Name: TABLE warranty_claims; Type: COMMENT; Schema: aftersales; Owner: -
--

COMMENT ON TABLE aftersales.warranty_claims IS '[aftersales] Warranty claims, identified by IMEI/serial.';


--
-- Name: business_members; Type: TABLE; Schema: b2b; Owner: -
--

CREATE TABLE b2b.business_members (
    business_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role text NOT NULL,
    approval_limit_kobo bigint,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT business_members_approval_limit_kobo_check CHECK ((approval_limit_kobo >= 0)),
    CONSTRAINT business_members_role_check CHECK ((role = ANY (ARRAY['admin'::text, 'buyer'::text, 'approver'::text])))
);


--
-- Name: TABLE business_members; Type: COMMENT; Schema: b2b; Owner: -
--

COMMENT ON TABLE b2b.business_members IS '[b2b] Users who buy or approve for a business. Buyers above approval_limit_kobo need an approver.';


--
-- Name: businesses; Type: TABLE; Schema: b2b; Owner: -
--

CREATE TABLE b2b.businesses (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    legal_name text NOT NULL,
    rc_number text,
    tin text,
    type text NOT NULL,
    size_band text,
    status text DEFAULT 'pending_verification'::text NOT NULL,
    account_manager_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT businesses_size_band_check CHECK ((size_band = ANY (ARRAY['1-10'::text, '11-50'::text, '51-200'::text, '201-1000'::text, '1000+'::text]))),
    CONSTRAINT businesses_status_check CHECK ((status = ANY (ARRAY['pending_verification'::text, 'active'::text, 'suspended'::text]))),
    CONSTRAINT businesses_type_check CHECK ((type = ANY (ARRAY['office_sme'::text, 'school'::text, 'reseller'::text, 'government'::text, 'ngo'::text, 'healthcare'::text, 'other'::text])))
);


--
-- Name: TABLE businesses; Type: COMMENT; Schema: b2b; Owner: -
--

COMMENT ON TABLE b2b.businesses IS '[b2b] Business buyers (wholesale site): trade pricing, VAT invoices, credit.';


--
-- Name: credit_accounts; Type: TABLE; Schema: b2b; Owner: -
--

CREATE TABLE b2b.credit_accounts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    business_id uuid NOT NULL,
    limit_kobo bigint DEFAULT 0 NOT NULL,
    currency character(3) DEFAULT 'NGN'::bpchar NOT NULL,
    term_days integer DEFAULT 30 NOT NULL,
    status text DEFAULT 'applied'::text NOT NULL,
    approved_by uuid,
    approved_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT credit_accounts_check CHECK (((status <> 'approved'::text) OR ((approved_by IS NOT NULL) AND (approved_at IS NOT NULL)))),
    CONSTRAINT credit_accounts_limit_kobo_check CHECK ((limit_kobo >= 0)),
    CONSTRAINT credit_accounts_status_check CHECK ((status = ANY (ARRAY['applied'::text, 'approved'::text, 'on_hold'::text, 'suspended'::text, 'closed'::text]))),
    CONSTRAINT credit_accounts_term_days_check CHECK (((term_days >= 1) AND (term_days <= 180)))
);


--
-- Name: TABLE credit_accounts; Type: COMMENT; Schema: b2b; Owner: -
--

COMMENT ON TABLE b2b.credit_accounts IS '[b2b] Pay-on-invoice credit for approved businesses (limit and payment term).';


--
-- Name: quote_lines; Type: TABLE; Schema: b2b; Owner: -
--

CREATE TABLE b2b.quote_lines (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    quote_id uuid NOT NULL,
    listing_id uuid,
    description text NOT NULL,
    quantity integer NOT NULL,
    unit_price_kobo bigint,
    CONSTRAINT quote_lines_quantity_check CHECK ((quantity > 0)),
    CONSTRAINT quote_lines_unit_price_kobo_check CHECK ((unit_price_kobo > 0))
);


--
-- Name: TABLE quote_lines; Type: COMMENT; Schema: b2b; Owner: -
--

COMMENT ON TABLE b2b.quote_lines IS '[b2b] Requested items (listed or free text) and their quoted unit prices.';


--
-- Name: quote_number_seq; Type: SEQUENCE; Schema: b2b; Owner: -
--

CREATE SEQUENCE b2b.quote_number_seq
    START WITH 1001
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: quotes; Type: TABLE; Schema: b2b; Owner: -
--

CREATE TABLE b2b.quotes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    quote_number text DEFAULT platform.next_ref('QT-'::text, 'b2b.quote_number_seq'::regclass) NOT NULL,
    business_id uuid,
    requested_by uuid,
    contact_name text,
    contact_email public.citext,
    contact_phone text,
    company_name text,
    assigned_to uuid,
    status text DEFAULT 'requested'::text NOT NULL,
    delivery_state text,
    needed_by date,
    needs_vat_invoice boolean DEFAULT false NOT NULL,
    wants_credit boolean DEFAULT false NOT NULL,
    notes text,
    total_kobo bigint,
    currency character(3) DEFAULT 'NGN'::bpchar NOT NULL,
    valid_until date,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT quotes_check CHECK (((status <> 'sent'::text) OR ((total_kobo IS NOT NULL) AND (valid_until IS NOT NULL)))),
    CONSTRAINT quotes_check1 CHECK (((requested_by IS NOT NULL) OR (contact_email IS NOT NULL) OR (contact_phone IS NOT NULL))),
    CONSTRAINT quotes_contact_phone_check CHECK ((contact_phone ~ '^\+234[0-9]{10}$'::text)),
    CONSTRAINT quotes_status_check CHECK ((status = ANY (ARRAY['requested'::text, 'drafting'::text, 'sent'::text, 'accepted'::text, 'declined'::text, 'expired'::text]))),
    CONSTRAINT quotes_total_kobo_check CHECK ((total_kobo >= 0))
);


--
-- Name: TABLE quotes; Type: COMMENT; Schema: b2b; Owner: -
--

COMMENT ON TABLE b2b.quotes IS '[b2b] B2B quote requests (from signed-in buyers or the public form) and the priced quotes sent back.';


--
-- Name: car_documents; Type: TABLE; Schema: cars; Owner: -
--

CREATE TABLE cars.car_documents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    car_listing_id uuid NOT NULL,
    kind text NOT NULL,
    file_id uuid NOT NULL,
    verified_by uuid,
    verified_at timestamp with time zone,
    CONSTRAINT car_documents_check CHECK (((verified_by IS NULL) = (verified_at IS NULL))),
    CONSTRAINT car_documents_kind_check CHECK ((kind = ANY (ARRAY['customs_papers'::text, 'proof_of_ownership'::text, 'vehicle_licence'::text, 'roadworthiness'::text, 'insurance'::text])))
);


--
-- Name: TABLE car_documents; Type: COMMENT; Schema: cars; Owner: -
--

COMMENT ON TABLE cars.car_documents IS '[cars] Ownership and import documents, verified by staff. A listing goes active only with verified documents.';


--
-- Name: car_images; Type: TABLE; Schema: cars; Owner: -
--

CREATE TABLE cars.car_images (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    car_listing_id uuid NOT NULL,
    file_id uuid NOT NULL,
    alt text NOT NULL,
    "position" integer DEFAULT 0 NOT NULL
);


--
-- Name: TABLE car_images; Type: COMMENT; Schema: cars; Owner: -
--

COMMENT ON TABLE cars.car_images IS '[cars] Photos of a car listing.';


--
-- Name: car_inspections; Type: TABLE; Schema: cars; Owner: -
--

CREATE TABLE cars.car_inspections (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    car_listing_id uuid NOT NULL,
    inspector_id uuid,
    status text DEFAULT 'scheduled'::text NOT NULL,
    checklist jsonb,
    report_file_id uuid,
    inspected_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT car_inspections_check CHECK (((status = 'scheduled'::text) OR (inspected_at IS NOT NULL))),
    CONSTRAINT car_inspections_status_check CHECK ((status = ANY (ARRAY['scheduled'::text, 'passed'::text, 'passed_with_notes'::text, 'failed'::text])))
);


--
-- Name: TABLE car_inspections; Type: COMMENT; Schema: cars; Owner: -
--

COMMENT ON TABLE cars.car_inspections IS '[cars] Inspection reports (engine, brakes, body, electricals, documents/VIN).';


--
-- Name: car_ref_seq; Type: SEQUENCE; Schema: cars; Owner: -
--

CREATE SEQUENCE cars.car_ref_seq
    START WITH 1001
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: car_listings; Type: TABLE; Schema: cars; Owner: -
--

CREATE TABLE cars.car_listings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    reference text DEFAULT platform.next_ref('CAR-'::text, 'cars.car_ref_seq'::regclass) NOT NULL,
    slug text NOT NULL,
    seller_id uuid NOT NULL,
    make text NOT NULL,
    model text NOT NULL,
    year smallint NOT NULL,
    "trim" text,
    body_type text,
    transmission text NOT NULL,
    fuel_type text,
    mileage_km integer NOT NULL,
    condition text NOT NULL,
    vin text,
    colour text,
    state_code text NOT NULL,
    city text NOT NULL,
    price_kobo bigint NOT NULL,
    currency character(3) DEFAULT 'NGN'::bpchar NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT car_listings_check CHECK (((condition <> 'brand_new'::text) OR (mileage_km < 500))),
    CONSTRAINT car_listings_condition_check CHECK ((condition = ANY (ARRAY['brand_new'::text, 'foreign_used'::text, 'nigerian_used'::text]))),
    CONSTRAINT car_listings_fuel_type_check CHECK ((fuel_type = ANY (ARRAY['petrol'::text, 'diesel'::text, 'hybrid'::text, 'electric'::text, 'cng'::text]))),
    CONSTRAINT car_listings_mileage_km_check CHECK ((mileage_km >= 0)),
    CONSTRAINT car_listings_price_kobo_check CHECK ((price_kobo > 0)),
    CONSTRAINT car_listings_slug_check CHECK ((slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'::text)),
    CONSTRAINT car_listings_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'in_review'::text, 'active'::text, 'reserved'::text, 'sold'::text, 'withdrawn'::text]))),
    CONSTRAINT car_listings_transmission_check CHECK ((transmission = ANY (ARRAY['automatic'::text, 'manual'::text]))),
    CONSTRAINT car_listings_vin_check CHECK ((vin ~ '^[A-HJ-NPR-Z0-9]{17}$'::text)),
    CONSTRAINT car_listings_year_check CHECK (((year >= 1950) AND (year <= 2100)))
);


--
-- Name: TABLE car_listings; Type: COMMENT; Schema: cars; Owner: -
--

COMMENT ON TABLE cars.car_listings IS '[cars] Cars for sale (separate from products): book a viewing, inspect, then pay. Never add-to-cart.';


--
-- Name: financing_enquiries; Type: TABLE; Schema: cars; Owner: -
--

CREATE TABLE cars.financing_enquiries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    car_listing_id uuid NOT NULL,
    customer_user_id uuid,
    name text NOT NULL,
    phone text NOT NULL,
    email public.citext,
    monthly_income_band text,
    down_payment_kobo bigint,
    status text DEFAULT 'new'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT financing_enquiries_down_payment_kobo_check CHECK ((down_payment_kobo >= 0)),
    CONSTRAINT financing_enquiries_phone_check CHECK ((phone ~ '^\+234[0-9]{10}$'::text)),
    CONSTRAINT financing_enquiries_status_check CHECK ((status = ANY (ARRAY['new'::text, 'contacted'::text, 'referred'::text, 'closed'::text])))
);


--
-- Name: TABLE financing_enquiries; Type: COMMENT; Schema: cars; Owner: -
--

COMMENT ON TABLE cars.financing_enquiries IS '[cars] Car financing enquiries, referred to finance partners.';


--
-- Name: viewing_bookings; Type: TABLE; Schema: cars; Owner: -
--

CREATE TABLE cars.viewing_bookings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    car_listing_id uuid NOT NULL,
    customer_user_id uuid,
    name text NOT NULL,
    phone text NOT NULL,
    preferred_date date NOT NULL,
    slot text NOT NULL,
    notes text,
    status text DEFAULT 'requested'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT viewing_bookings_phone_check CHECK ((phone ~ '^\+234[0-9]{10}$'::text)),
    CONSTRAINT viewing_bookings_slot_check CHECK ((slot = ANY (ARRAY['morning'::text, 'afternoon'::text, 'late_afternoon'::text]))),
    CONSTRAINT viewing_bookings_status_check CHECK ((status = ANY (ARRAY['requested'::text, 'confirmed'::text, 'completed'::text, 'no_show'::text, 'cancelled'::text])))
);


--
-- Name: TABLE viewing_bookings; Type: COMMENT; Schema: cars; Owner: -
--

COMMENT ON TABLE cars.viewing_bookings IS '[cars] Requests to see a car (the Book a viewing form).';


--
-- Name: attribute_definitions; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.attribute_definitions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    category_id uuid NOT NULL,
    key text NOT NULL,
    label text NOT NULL,
    data_type text NOT NULL,
    unit text,
    options jsonb,
    is_required boolean DEFAULT false NOT NULL,
    is_filterable boolean DEFAULT false NOT NULL,
    is_variant_axis boolean DEFAULT false NOT NULL,
    "position" integer DEFAULT 0 NOT NULL,
    CONSTRAINT attribute_definitions_check CHECK (((data_type <> 'enum'::text) OR (jsonb_typeof(options) = 'array'::text))),
    CONSTRAINT attribute_definitions_data_type_check CHECK ((data_type = ANY (ARRAY['text'::text, 'number'::text, 'boolean'::text, 'enum'::text]))),
    CONSTRAINT attribute_definitions_key_check CHECK ((key ~ '^[a-z][a-z0-9_]*$'::text))
);


--
-- Name: TABLE attribute_definitions; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.attribute_definitions IS '[catalog] Spec fields per category (RAM, storage, GPU…): drive spec tables, filters and variant axes.';


--
-- Name: brands; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.brands (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    slug text NOT NULL,
    name text NOT NULL,
    logo_file_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE brands; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.brands IS '[catalog] Manufacturers and brands.';


--
-- Name: categories; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    parent_id uuid,
    slug text NOT NULL,
    name text NOT NULL,
    description text,
    kind text DEFAULT 'product'::text NOT NULL,
    repurchase_days integer,
    "position" integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT categories_check CHECK ((parent_id IS DISTINCT FROM id)),
    CONSTRAINT categories_kind_check CHECK ((kind = ANY (ARRAY['product'::text, 'vehicle'::text]))),
    CONSTRAINT categories_repurchase_days_check CHECK ((repurchase_days > 0))
);


--
-- Name: TABLE categories; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.categories IS '[catalog] Category tree (Phones → Android…). kind = vehicle routes to cars; repurchase_days stops re-recommending phones.';


--
-- Name: image_hashes; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.image_hashes (
    product_image_id uuid NOT NULL,
    phash bigint NOT NULL
);


--
-- Name: TABLE image_hashes; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.image_hashes IS '[catalog] Perceptual hashes of product photos, to catch photos stolen from other sellers.';


--
-- Name: listing_price_history; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.listing_price_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    listing_id uuid NOT NULL,
    price_kobo bigint NOT NULL,
    compare_at_kobo bigint,
    changed_by uuid,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT listing_price_history_price_kobo_check CHECK ((price_kobo > 0))
);


--
-- Name: TABLE listing_price_history; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.listing_price_history IS '[catalog] Every price change. Proves discounts are genuine (FCCPA 30-day maximum) and feeds price-drop alerts.';


--
-- Name: listing_reviews; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.listing_reviews (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    listing_id uuid NOT NULL,
    reviewer_id uuid,
    decision text NOT NULL,
    reasons text[] DEFAULT '{}'::text[] NOT NULL,
    note text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT listing_reviews_check CHECK (((decision = 'auto_passed'::text) OR (reviewer_id IS NOT NULL))),
    CONSTRAINT listing_reviews_decision_check CHECK ((decision = ANY (ARRAY['auto_passed'::text, 'approved'::text, 'changes_requested'::text, 'rejected'::text, 'suspended'::text])))
);


--
-- Name: TABLE listing_reviews; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.listing_reviews IS '[catalog] Moderation decisions on listings (automatic checks and human reviewers), visible to the seller.';


--
-- Name: listings; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.listings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    seller_id uuid NOT NULL,
    variant_id uuid NOT NULL,
    condition text NOT NULL,
    grade text,
    price_kobo bigint NOT NULL,
    compare_at_kobo bigint,
    currency character(3) DEFAULT 'NGN'::bpchar NOT NULL,
    fulfilled_by text DEFAULT 'techshop'::text NOT NULL,
    seller_stock integer,
    min_order_quantity integer,
    handling_days smallint DEFAULT 1 NOT NULL,
    warranty_provider text DEFAULT 'manufacturer'::text NOT NULL,
    warranty_months smallint DEFAULT 0 NOT NULL,
    condition_notes text,
    battery_health_pct smallint,
    status text DEFAULT 'draft'::text NOT NULL,
    risk_score smallint,
    published_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT listings_battery_health_pct_check CHECK (((battery_health_pct >= 1) AND (battery_health_pct <= 100))),
    CONSTRAINT listings_check CHECK (((compare_at_kobo IS NULL) OR (compare_at_kobo > price_kobo))),
    CONSTRAINT listings_check1 CHECK (((condition = 'new'::text) OR (condition_notes IS NOT NULL))),
    CONSTRAINT listings_check2 CHECK (((fulfilled_by = 'techshop'::text) OR (seller_stock IS NOT NULL))),
    CONSTRAINT listings_check3 CHECK (((status <> 'active'::text) OR (published_at IS NOT NULL))),
    CONSTRAINT listings_condition_check CHECK ((condition = ANY (ARRAY['new'::text, 'uk_used'::text, 'refurbished'::text, 'open_box'::text]))),
    CONSTRAINT listings_fulfilled_by_check CHECK ((fulfilled_by = ANY (ARRAY['techshop'::text, 'seller'::text]))),
    CONSTRAINT listings_grade_check CHECK ((grade = ANY (ARRAY['A'::text, 'B'::text, 'C'::text]))),
    CONSTRAINT listings_handling_days_check CHECK (((handling_days >= 0) AND (handling_days <= 10))),
    CONSTRAINT listings_min_order_quantity_check CHECK ((min_order_quantity > 1)),
    CONSTRAINT listings_price_kobo_check CHECK ((price_kobo > 0)),
    CONSTRAINT listings_risk_score_check CHECK (((risk_score >= 0) AND (risk_score <= 100))),
    CONSTRAINT listings_seller_stock_check CHECK ((seller_stock >= 0)),
    CONSTRAINT listings_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'auto_check'::text, 'in_review'::text, 'changes_requested'::text, 'active'::text, 'paused'::text, 'suspended'::text, 'rejected'::text]))),
    CONSTRAINT listings_warranty_months_check CHECK (((warranty_months >= 0) AND (warranty_months <= 60))),
    CONSTRAINT listings_warranty_provider_check CHECK ((warranty_provider = ANY (ARRAY['manufacturer'::text, 'techshop'::text, 'seller'::text, 'none'::text])))
);


--
-- Name: TABLE listings; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.listings IS '[catalog] A seller''s offer on a variant: condition, price, stock, warranty. TechShop stock is a first_party listing.';


--
-- Name: price_tiers; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.price_tiers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    listing_id uuid NOT NULL,
    min_quantity integer NOT NULL,
    unit_price_kobo bigint NOT NULL,
    CONSTRAINT price_tiers_min_quantity_check CHECK ((min_quantity > 1)),
    CONSTRAINT price_tiers_unit_price_kobo_check CHECK ((unit_price_kobo > 0))
);


--
-- Name: TABLE price_tiers; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.price_tiers IS '[catalog] Wholesale quantity breaks: unit price from min_quantity upward (business accounts only).';


--
-- Name: product_attribute_values; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.product_attribute_values (
    product_id uuid NOT NULL,
    attribute_id uuid NOT NULL,
    value_text text,
    value_number numeric,
    value_bool boolean,
    CONSTRAINT product_attribute_values_check CHECK ((num_nonnulls(value_text, value_number, value_bool) = 1))
);


--
-- Name: TABLE product_attribute_values; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.product_attribute_values IS '[catalog] Spec values for a product (exactly one typed value per attribute).';


--
-- Name: product_compatibility; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.product_compatibility (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    accessory_product_id uuid NOT NULL,
    device_product_id uuid,
    rule jsonb,
    source text NOT NULL,
    verified boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT product_compatibility_check CHECK ((num_nonnulls(device_product_id, rule) = 1)),
    CONSTRAINT product_compatibility_check1 CHECK ((device_product_id IS DISTINCT FROM accessory_product_id)),
    CONSTRAINT product_compatibility_source_check CHECK ((source = ANY (ARRAY['manual'::text, 'attribute_rule'::text, 'inferred'::text])))
);


--
-- Name: TABLE product_compatibility; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.product_compatibility IS '[catalog] Which accessories fit which devices (explicit pair or attribute rule), for "fits your phone" shelves.';


--
-- Name: product_images; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.product_images (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid NOT NULL,
    variant_id uuid,
    file_id uuid NOT NULL,
    alt text NOT NULL,
    "position" integer DEFAULT 0 NOT NULL
);


--
-- Name: TABLE product_images; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.product_images IS '[catalog] Product photos (optionally per variant); alt text required.';


--
-- Name: product_offer_summary; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.product_offer_summary (
    product_id uuid NOT NULL,
    channel text NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    category_ids uuid[] NOT NULL,
    brand_id uuid,
    best_listing_id uuid,
    min_price_kobo bigint,
    max_price_kobo bigint,
    compare_at_kobo bigint,
    promo_price_kobo bigint,
    discount_pct smallint,
    offer_count integer DEFAULT 0 NOT NULL,
    conditions text[] DEFAULT '{}'::text[] NOT NULL,
    seller_types text[] DEFAULT '{}'::text[] NOT NULL,
    in_stock boolean DEFAULT false NOT NULL,
    stock_band text,
    rating_avg numeric(3,2),
    rating_count integer DEFAULT 0 NOT NULL,
    sales_30d integer DEFAULT 0 NOT NULL,
    return_rate_90d numeric(5,4),
    attrs jsonb DEFAULT '{}'::jsonb NOT NULL,
    facet_keys text[] DEFAULT '{}'::text[] NOT NULL,
    search_vector tsvector NOT NULL,
    refreshed_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT product_offer_summary_channel_check CHECK ((channel = ANY (ARRAY['retail'::text, 'wholesale'::text]))),
    CONSTRAINT product_offer_summary_check CHECK (((max_price_kobo IS NULL) OR (min_price_kobo IS NULL) OR (max_price_kobo >= min_price_kobo))),
    CONSTRAINT product_offer_summary_discount_pct_check CHECK (((discount_pct >= 0) AND (discount_pct <= 100))),
    CONSTRAINT product_offer_summary_max_price_kobo_check CHECK ((max_price_kobo > 0)),
    CONSTRAINT product_offer_summary_min_price_kobo_check CHECK ((min_price_kobo > 0)),
    CONSTRAINT product_offer_summary_offer_count_check CHECK ((offer_count >= 0)),
    CONSTRAINT product_offer_summary_stock_band_check CHECK ((stock_band = ANY (ARRAY['out'::text, 'low'::text, 'in'::text])))
);


--
-- Name: TABLE product_offer_summary; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.product_offer_summary IS '[catalog] Search read model: offers, buy box winner, facets and search vector per product and channel. Never used for pricing at checkout.';


--
-- Name: product_reviews; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.product_reviews (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid NOT NULL,
    order_line_id uuid,
    user_id uuid NOT NULL,
    rating smallint NOT NULL,
    title text,
    body text,
    is_verified boolean GENERATED ALWAYS AS ((order_line_id IS NOT NULL)) STORED,
    status text DEFAULT 'pending'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT product_reviews_rating_check CHECK (((rating >= 1) AND (rating <= 5))),
    CONSTRAINT product_reviews_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'published'::text, 'hidden'::text, 'rejected'::text])))
);


--
-- Name: TABLE product_reviews; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.product_reviews IS '[catalog] Product reviews. Verified purchase when tied to the order line bought; unverified ones are shown separately and weigh less.';


--
-- Name: product_suggestions; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.product_suggestions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    seller_id uuid NOT NULL,
    product_id uuid,
    kind text NOT NULL,
    payload jsonb NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    reviewed_by uuid,
    reviewed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT product_suggestions_check CHECK (((kind = 'new_product'::text) OR (product_id IS NOT NULL))),
    CONSTRAINT product_suggestions_check1 CHECK (((status = 'pending'::text) OR (reviewed_by IS NOT NULL))),
    CONSTRAINT product_suggestions_kind_check CHECK ((kind = ANY (ARRAY['new_product'::text, 'edit'::text]))),
    CONSTRAINT product_suggestions_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'accepted'::text, 'rejected'::text])))
);


--
-- Name: TABLE product_suggestions; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.product_suggestions IS '[catalog] Seller-proposed new products or edits, waiting for the catalogue team.';


--
-- Name: product_variants; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.product_variants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid NOT NULL,
    sku text NOT NULL,
    name text NOT NULL,
    axis_values jsonb DEFAULT '{}'::jsonb NOT NULL,
    gtin text,
    mpn text,
    weight_grams integer,
    length_mm integer,
    width_mm integer,
    height_mm integer,
    is_serialised boolean DEFAULT false NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT product_variants_height_mm_check CHECK ((height_mm > 0)),
    CONSTRAINT product_variants_length_mm_check CHECK ((length_mm > 0)),
    CONSTRAINT product_variants_weight_grams_check CHECK ((weight_grams > 0)),
    CONSTRAINT product_variants_width_mm_check CHECK ((width_mm > 0))
);


--
-- Name: TABLE product_variants; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.product_variants IS '[catalog] Purchasable configurations (256GB · Titanium). Serialised variants are tracked per unit (IMEI/serial).';


--
-- Name: products; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.products (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    category_id uuid NOT NULL,
    brand_id uuid,
    slug text NOT NULL,
    name text NOT NULL,
    model text,
    description text,
    status text DEFAULT 'draft'::text NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS ((setweight(to_tsvector('simple'::regconfig, ((COALESCE(name, ''::text) || ' '::text) || COALESCE(model, ''::text))), 'A'::"char") || setweight(to_tsvector('simple'::regconfig, COALESCE(description, ''::text)), 'C'::"char"))) STORED,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT products_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'in_review'::text, 'active'::text, 'archived'::text])))
);


--
-- Name: TABLE products; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.products IS '[catalog] Catalogue entries at model level. TechShop owns titles, images and specs; sellers sell through listings on variants.';


--
-- Name: review_images; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.review_images (
    review_id uuid NOT NULL,
    file_id uuid NOT NULL,
    "position" smallint DEFAULT 0 NOT NULL
);


--
-- Name: TABLE review_images; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.review_images IS '[catalog] Photos attached to a review.';


--
-- Name: saved_items; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.saved_items (
    user_id uuid NOT NULL,
    product_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE saved_items; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.saved_items IS '[catalog] Customers'' saved products (the heart button).';


--
-- Name: seller_ratings; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.seller_ratings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    seller_id uuid NOT NULL,
    fulfilment_id uuid NOT NULL,
    user_id uuid NOT NULL,
    rating smallint NOT NULL,
    comment text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT seller_ratings_rating_check CHECK (((rating >= 1) AND (rating <= 5)))
);


--
-- Name: TABLE seller_ratings; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.seller_ratings IS '[catalog] Customer rating of a seller per delivered fulfilment; aggregated onto sellers.rating_avg.';


--
-- Name: stock_alerts; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.stock_alerts (
    user_id uuid NOT NULL,
    product_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    notified_at timestamp with time zone
);


--
-- Name: TABLE stock_alerts; Type: COMMENT; Schema: catalog; Owner: -
--

COMMENT ON TABLE catalog.stock_alerts IS '[catalog] "Notify me when back in stock" requests.';


--
-- Name: banners; Type: TABLE; Schema: content; Owner: -
--

CREATE TABLE content.banners (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    site text NOT NULL,
    placement text NOT NULL,
    title text NOT NULL,
    image_file_id uuid,
    link_url text NOT NULL,
    starts_at timestamp with time zone,
    ends_at timestamp with time zone,
    "position" integer DEFAULT 0 NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT banners_check CHECK (((ends_at IS NULL) OR (starts_at IS NULL) OR (ends_at > starts_at))),
    CONSTRAINT banners_link_url_check CHECK ((link_url ~ '^(/|https://)'::text)),
    CONSTRAINT banners_site_check CHECK ((site = ANY (ARRAY['corporate'::text, 'market'::text, 'wholesale'::text, 'seller'::text]))),
    CONSTRAINT banners_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'live'::text, 'ended'::text])))
);


--
-- Name: TABLE banners; Type: COMMENT; Schema: content; Owner: -
--

COMMENT ON TABLE content.banners IS '[content] Hero and promo banners per site and placement (links: internal paths or https only).';


--
-- Name: cms_pages; Type: TABLE; Schema: content; Owner: -
--

CREATE TABLE content.cms_pages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    site text NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    body jsonb DEFAULT '[]'::jsonb NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    published_at timestamp with time zone,
    updated_by uuid,
    approved_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT cms_pages_check CHECK (((status <> 'published'::text) OR (published_at IS NOT NULL))),
    CONSTRAINT cms_pages_site_check CHECK ((site = ANY (ARRAY['corporate'::text, 'market'::text, 'wholesale'::text, 'seller'::text]))),
    CONSTRAINT cms_pages_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'in_review'::text, 'published'::text])))
);


--
-- Name: TABLE cms_pages; Type: COMMENT; Schema: content; Owner: -
--

COMMENT ON TABLE content.cms_pages IS '[content] Editable pages per site (about, legal, credit terms…) as content blocks; publishing needs approval.';


--
-- Name: help_articles; Type: TABLE; Schema: content; Owner: -
--

CREATE TABLE content.help_articles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    site text NOT NULL,
    topic text NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    body jsonb DEFAULT '[]'::jsonb NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    published_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT help_articles_check CHECK (((status <> 'published'::text) OR (published_at IS NOT NULL))),
    CONSTRAINT help_articles_site_check CHECK ((site = ANY (ARRAY['market'::text, 'wholesale'::text, 'seller'::text]))),
    CONSTRAINT help_articles_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'in_review'::text, 'published'::text])))
);


--
-- Name: TABLE help_articles; Type: COMMENT; Schema: content; Owner: -
--

COMMENT ON TABLE content.help_articles IS '[content] Help centre articles and FAQs per site.';


--
-- Name: accounting_periods; Type: TABLE; Schema: finance; Owner: -
--

CREATE TABLE finance.accounting_periods (
    month date NOT NULL,
    status text DEFAULT 'open'::text NOT NULL,
    closed_by uuid,
    closed_at timestamp with time zone,
    CONSTRAINT accounting_periods_check CHECK (((status <> 'closed'::text) OR ((closed_by IS NOT NULL) AND (closed_at IS NOT NULL)))),
    CONSTRAINT accounting_periods_month_check CHECK ((month = (date_trunc('month'::text, (month)::timestamp with time zone))::date)),
    CONSTRAINT accounting_periods_status_check CHECK ((status = ANY (ARRAY['open'::text, 'closing'::text, 'closed'::text])))
);


--
-- Name: TABLE accounting_periods; Type: COMMENT; Schema: finance; Owner: -
--

COMMENT ON TABLE finance.accounting_periods IS '[finance] Month-end close. Journals cannot be posted into a closed month.';


--
-- Name: commission_rules; Type: TABLE; Schema: finance; Owner: -
--

CREATE TABLE finance.commission_rules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    category_id uuid,
    seller_id uuid,
    rate_bps integer NOT NULL,
    valid_from date DEFAULT CURRENT_DATE NOT NULL,
    valid_to date,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT commission_rules_check CHECK (((valid_to IS NULL) OR (valid_to > valid_from))),
    CONSTRAINT commission_rules_rate_bps_check CHECK (((rate_bps >= 0) AND (rate_bps <= 10000)))
);


--
-- Name: TABLE commission_rules; Type: COMMENT; Schema: finance; Owner: -
--

COMMENT ON TABLE finance.commission_rules IS '[finance] Marketplace commission (basis points) by category and/or seller; the rate is snapshotted on order lines.';


--
-- Name: invoice_number_seq; Type: SEQUENCE; Schema: finance; Owner: -
--

CREATE SEQUENCE finance.invoice_number_seq
    START WITH 1001
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invoices; Type: TABLE; Schema: finance; Owner: -
--

CREATE TABLE finance.invoices (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    invoice_number text DEFAULT platform.next_ref('INV-'::text, 'finance.invoice_number_seq'::regclass) NOT NULL,
    business_id uuid NOT NULL,
    order_id uuid NOT NULL,
    kind text DEFAULT 'tax'::text NOT NULL,
    issued_at timestamp with time zone DEFAULT now() NOT NULL,
    due_at timestamp with time zone NOT NULL,
    subtotal_kobo bigint NOT NULL,
    vat_kobo bigint DEFAULT 0 NOT NULL,
    total_kobo bigint NOT NULL,
    amount_paid_kobo bigint DEFAULT 0 NOT NULL,
    currency character(3) DEFAULT 'NGN'::bpchar NOT NULL,
    status text DEFAULT 'issued'::text NOT NULL,
    pdf_file_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT invoices_amount_paid_kobo_check CHECK ((amount_paid_kobo >= 0)),
    CONSTRAINT invoices_check CHECK ((total_kobo = (subtotal_kobo + vat_kobo))),
    CONSTRAINT invoices_check1 CHECK ((amount_paid_kobo <= total_kobo)),
    CONSTRAINT invoices_check2 CHECK ((due_at >= issued_at)),
    CONSTRAINT invoices_kind_check CHECK ((kind = ANY (ARRAY['proforma'::text, 'tax'::text]))),
    CONSTRAINT invoices_status_check CHECK ((status = ANY (ARRAY['issued'::text, 'partially_paid'::text, 'paid'::text, 'overdue'::text, 'void'::text, 'written_off'::text]))),
    CONSTRAINT invoices_subtotal_kobo_check CHECK ((subtotal_kobo >= 0)),
    CONSTRAINT invoices_total_kobo_check CHECK ((total_kobo >= 0)),
    CONSTRAINT invoices_vat_kobo_check CHECK ((vat_kobo >= 0))
);


--
-- Name: TABLE invoices; Type: COMMENT; Schema: finance; Owner: -
--

COMMENT ON TABLE finance.invoices IS '[finance] B2B invoices (proforma or tax); on credit terms they are paid later against due_at.';


--
-- Name: ledger_accounts; Type: TABLE; Schema: finance; Owner: -
--

CREATE TABLE finance.ledger_accounts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    kind text NOT NULL,
    seller_id uuid,
    user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ledger_accounts_check CHECK (((seller_id IS NULL) OR (user_id IS NULL))),
    CONSTRAINT ledger_accounts_code_check CHECK ((code ~ '^[a-z_]+(:[a-z0-9_-]+)+$'::text)),
    CONSTRAINT ledger_accounts_kind_check CHECK ((kind = ANY (ARRAY['asset'::text, 'liability'::text, 'revenue'::text, 'expense'::text, 'equity'::text])))
);


--
-- Name: TABLE ledger_accounts; Type: COMMENT; Schema: finance; Owner: -
--

COMMENT ON TABLE finance.ledger_accounts IS '[finance] Chart of accounts: provider clearing, bank, revenue, VAT, refunds, one payable per seller, store credit per customer.';


--
-- Name: ledger_entries; Type: TABLE; Schema: finance; Owner: -
--

CREATE TABLE finance.ledger_entries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    journal_id uuid NOT NULL,
    account_id uuid NOT NULL,
    amount_kobo bigint NOT NULL,
    currency character(3) DEFAULT 'NGN'::bpchar NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ledger_entries_amount_kobo_check CHECK ((amount_kobo <> 0))
);


--
-- Name: TABLE ledger_entries; Type: COMMENT; Schema: finance; Owner: -
--

COMMENT ON TABLE finance.ledger_entries IS '[finance] Double-entry lines: positive = debit, negative = credit. Each journal sums to zero (deferred check).';


--
-- Name: ledger_journals; Type: TABLE; Schema: finance; Owner: -
--

CREATE TABLE finance.ledger_journals (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    posting_key text NOT NULL,
    kind text NOT NULL,
    reference_type text NOT NULL,
    reference_id uuid NOT NULL,
    period date DEFAULT (date_trunc('month'::text, now()))::date NOT NULL,
    reverses_journal_id uuid,
    memo text,
    posted_by uuid,
    posted_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ledger_journals_check CHECK (((kind <> 'reversal'::text) OR (reverses_journal_id IS NOT NULL))),
    CONSTRAINT ledger_journals_kind_check CHECK ((kind = ANY (ARRAY['sale'::text, 'delivery_fee'::text, 'commission'::text, 'payout'::text, 'refund'::text, 'chargeback'::text, 'adjustment'::text, 'reversal'::text, 'settlement'::text, 'fee'::text, 'cogs'::text, 'store_credit'::text, 'write_off'::text]))),
    CONSTRAINT ledger_journals_period_check CHECK ((period = (date_trunc('month'::text, (period)::timestamp with time zone))::date))
);


--
-- Name: TABLE ledger_journals; Type: COMMENT; Schema: finance; Owner: -
--

COMMENT ON TABLE finance.ledger_journals IS '[finance] One accounting event. posting_key makes posting idempotent; journals are never edited, only reversed.';


--
-- Name: payout_batches; Type: TABLE; Schema: finance; Owner: -
--

CREATE TABLE finance.payout_batches (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    prepared_by uuid NOT NULL,
    approved_by uuid,
    approved_at timestamp with time zone,
    total_kobo bigint DEFAULT 0 NOT NULL,
    payout_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT payout_batches_check CHECK (((approved_by IS DISTINCT FROM prepared_by) OR (approved_by IS NULL))),
    CONSTRAINT payout_batches_check1 CHECK (((status <> ALL (ARRAY['approved'::text, 'processing'::text, 'done'::text])) OR ((approved_by IS NOT NULL) AND (approved_at IS NOT NULL)))),
    CONSTRAINT payout_batches_payout_count_check CHECK ((payout_count >= 0)),
    CONSTRAINT payout_batches_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'submitted'::text, 'approved'::text, 'processing'::text, 'done'::text, 'cancelled'::text]))),
    CONSTRAINT payout_batches_total_kobo_check CHECK ((total_kobo >= 0))
);


--
-- Name: TABLE payout_batches; Type: COMMENT; Schema: finance; Owner: -
--

COMMENT ON TABLE finance.payout_batches IS '[finance] A payout run: prepared by one finance person, approved by a different one (with MFA step-up).';


--
-- Name: payout_holds; Type: TABLE; Schema: finance; Owner: -
--

CREATE TABLE finance.payout_holds (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    seller_id uuid NOT NULL,
    reason text NOT NULL,
    note text,
    until timestamp with time zone,
    created_by uuid,
    released_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT payout_holds_reason_check CHECK ((reason = ANY (ARRAY['kyc_incomplete'::text, 'risk_case'::text, 'bank_account_changed'::text, 'credentials_changed'::text, 'negative_balance'::text, 'enforcement'::text, 'manual'::text])))
);


--
-- Name: TABLE payout_holds; Type: COMMENT; Schema: finance; Owner: -
--

COMMENT ON TABLE finance.payout_holds IS '[finance] Reasons a seller is not paid out right now (e.g. 24 hours after a bank or credential change).';


--
-- Name: payout_items; Type: TABLE; Schema: finance; Owner: -
--

CREATE TABLE finance.payout_items (
    payout_id uuid NOT NULL,
    fulfilment_id uuid NOT NULL,
    amount_kobo bigint NOT NULL,
    CONSTRAINT payout_items_amount_kobo_check CHECK ((amount_kobo > 0))
);


--
-- Name: TABLE payout_items; Type: COMMENT; Schema: finance; Owner: -
--

COMMENT ON TABLE finance.payout_items IS '[finance] Which delivered fulfilments a payout covers; each fulfilment is paid out once.';


--
-- Name: payouts; Type: TABLE; Schema: finance; Owner: -
--

CREATE TABLE finance.payouts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    batch_id uuid,
    seller_id uuid NOT NULL,
    bank_account_id uuid NOT NULL,
    amount_kobo bigint NOT NULL,
    currency character(3) DEFAULT 'NGN'::bpchar NOT NULL,
    status text DEFAULT 'scheduled'::text NOT NULL,
    scheduled_for date NOT NULL,
    paid_at timestamp with time zone,
    provider_reference text,
    failure_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT payouts_amount_kobo_check CHECK ((amount_kobo > 0)),
    CONSTRAINT payouts_check CHECK (((status <> 'paid'::text) OR (paid_at IS NOT NULL))),
    CONSTRAINT payouts_check1 CHECK (((status <> 'failed'::text) OR (failure_reason IS NOT NULL))),
    CONSTRAINT payouts_status_check CHECK ((status = ANY (ARRAY['scheduled'::text, 'processing'::text, 'paid'::text, 'failed'::text, 'on_hold'::text])))
);


--
-- Name: TABLE payouts; Type: COMMENT; Schema: finance; Owner: -
--

COMMENT ON TABLE finance.payouts IS '[finance] Transfers of earnings to a seller''s bank account (journal posted only on success).';


--
-- Name: carriers; Type: TABLE; Schema: fulfilment; Owner: -
--

CREATE TABLE fulfilment.carriers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE carriers; Type: COMMENT; Schema: fulfilment; Owner: -
--

COMMENT ON TABLE fulfilment.carriers IS '[fulfilment] Third-party carriers for interstate delivery (behind carrier.Port).';


--
-- Name: manifest_parcels; Type: TABLE; Schema: fulfilment; Owner: -
--

CREATE TABLE fulfilment.manifest_parcels (
    manifest_id uuid NOT NULL,
    parcel_id uuid NOT NULL,
    scanned_at timestamp with time zone
);


--
-- Name: TABLE manifest_parcels; Type: COMMENT; Schema: fulfilment; Owner: -
--

COMMENT ON TABLE fulfilment.manifest_parcels IS '[fulfilment] Parcels on a manifest and when each was scanned.';


--
-- Name: manifests; Type: TABLE; Schema: fulfilment; Owner: -
--

CREATE TABLE fulfilment.manifests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    warehouse_id uuid NOT NULL,
    rider_id uuid,
    carrier_id uuid,
    signed_by_name text,
    signed_at timestamp with time zone,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT manifests_check CHECK (((rider_id IS NULL) <> (carrier_id IS NULL)))
);


--
-- Name: TABLE manifests; Type: COMMENT; Schema: fulfilment; Owner: -
--

COMMENT ON TABLE fulfilment.manifests IS '[fulfilment] Handover batches to a rider or carrier; counts must match before signing.';


--
-- Name: pack_records; Type: TABLE; Schema: fulfilment; Owner: -
--

CREATE TABLE fulfilment.pack_records (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    fulfilment_id uuid NOT NULL,
    parcel_id uuid NOT NULL,
    weight_grams integer NOT NULL,
    expected_weight_grams integer,
    flagged boolean GENERATED ALWAYS AS (((expected_weight_grams IS NOT NULL) AND ((abs((weight_grams - expected_weight_grams)) * 10) > expected_weight_grams))) STORED,
    packed_by uuid NOT NULL,
    packed_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT pack_records_expected_weight_grams_check CHECK ((expected_weight_grams > 0)),
    CONSTRAINT pack_records_weight_grams_check CHECK ((weight_grams > 0))
);


--
-- Name: TABLE pack_records; Type: COMMENT; Schema: fulfilment; Owner: -
--

COMMENT ON TABLE fulfilment.pack_records IS '[fulfilment] Packing: weighed parcel; more than 10% off the expected weight is flagged.';


--
-- Name: parcels; Type: TABLE; Schema: fulfilment; Owner: -
--

CREATE TABLE fulfilment.parcels (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    fulfilment_id uuid NOT NULL,
    label_code text NOT NULL,
    box_code text,
    weight_grams integer,
    scanned_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT parcels_weight_grams_check CHECK ((weight_grams > 0))
);


--
-- Name: TABLE parcels; Type: COMMENT; Schema: fulfilment; Owner: -
--

COMMENT ON TABLE fulfilment.parcels IS '[fulfilment] Physical parcels with a printed label code, scanned at handover and pickup.';


--
-- Name: pick_list_items; Type: TABLE; Schema: fulfilment; Owner: -
--

CREATE TABLE fulfilment.pick_list_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    pick_list_id uuid NOT NULL,
    fulfilment_id uuid NOT NULL,
    order_line_id uuid NOT NULL,
    bin_id uuid,
    quantity integer NOT NULL,
    picked_qty integer DEFAULT 0 NOT NULL,
    device_unit_id uuid,
    scanned_at timestamp with time zone,
    CONSTRAINT pick_list_items_check CHECK ((picked_qty <= quantity)),
    CONSTRAINT pick_list_items_picked_qty_check CHECK ((picked_qty >= 0)),
    CONSTRAINT pick_list_items_quantity_check CHECK ((quantity > 0))
);


--
-- Name: TABLE pick_list_items; Type: COMMENT; Schema: fulfilment; Owner: -
--

COMMENT ON TABLE fulfilment.pick_list_items IS '[fulfilment] Lines to pick; scanning the serial binds the exact device unit to the order line.';


--
-- Name: pick_lists; Type: TABLE; Schema: fulfilment; Owner: -
--

CREATE TABLE fulfilment.pick_lists (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    warehouse_id uuid NOT NULL,
    wave_at timestamp with time zone NOT NULL,
    zone text,
    status text DEFAULT 'open'::text NOT NULL,
    picker_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT pick_lists_status_check CHECK ((status = ANY (ARRAY['open'::text, 'picking'::text, 'done'::text, 'cancelled'::text])))
);


--
-- Name: TABLE pick_lists; Type: COMMENT; Schema: fulfilment; Owner: -
--

COMMENT ON TABLE fulfilment.pick_lists IS '[fulfilment] Wave pick lists (every 30 minutes) per warehouse and zone.';


--
-- Name: seller_sla_events; Type: TABLE; Schema: fulfilment; Owner: -
--

CREATE TABLE fulfilment.seller_sla_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    seller_id uuid NOT NULL,
    fulfilment_id uuid NOT NULL,
    kind text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT seller_sla_events_kind_check CHECK ((kind = ANY (ARRAY['late_accept'::text, 'late_pack'::text, 'seller_cancel'::text, 'late_handover'::text])))
);


--
-- Name: TABLE seller_sla_events; Type: COMMENT; Schema: fulfilment; Owner: -
--

COMMENT ON TABLE fulfilment.seller_sla_events IS '[fulfilment] Seller SLA breaches (feed seller performance and enforcement).';


--
-- Name: shipment_events; Type: TABLE; Schema: fulfilment; Owner: -
--

CREATE TABLE fulfilment.shipment_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    shipment_id uuid NOT NULL,
    carrier_event_id text,
    status text NOT NULL,
    location text,
    occurred_at timestamp with time zone NOT NULL,
    received_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE shipment_events; Type: COMMENT; Schema: fulfilment; Owner: -
--

COMMENT ON TABLE fulfilment.shipment_events IS '[fulfilment] Tracking events from carrier webhooks or polling (deduplicated).';


--
-- Name: shipments; Type: TABLE; Schema: fulfilment; Owner: -
--

CREATE TABLE fulfilment.shipments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    fulfilment_id uuid NOT NULL,
    carrier_id uuid,
    carrier_name text,
    tracking_number text NOT NULL,
    label_file_id uuid,
    status text DEFAULT 'booked'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT shipments_check CHECK (((carrier_id IS NOT NULL) OR (carrier_name IS NOT NULL))),
    CONSTRAINT shipments_status_check CHECK ((status = ANY (ARRAY['booked'::text, 'in_transit'::text, 'delivered'::text, 'exception'::text, 'lost'::text, 'returned'::text])))
);


--
-- Name: TABLE shipments; Type: COMMENT; Schema: fulfilment; Owner: -
--

COMMENT ON TABLE fulfilment.shipments IS '[fulfilment] Carrier shipments (TechShop-booked or seller''s own carrier with tracking).';


--
-- Name: addresses; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.addresses (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    business_id uuid,
    label text,
    recipient_name text NOT NULL,
    phone text NOT NULL,
    line1 text NOT NULL,
    line2 text,
    landmark text,
    city text NOT NULL,
    lga text,
    state_code text NOT NULL,
    latitude numeric(9,6),
    longitude numeric(9,6),
    is_default boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT addresses_check CHECK (((user_id IS NULL) <> (business_id IS NULL))),
    CONSTRAINT addresses_latitude_check CHECK (((latitude >= ('-90'::integer)::numeric) AND (latitude <= (90)::numeric))),
    CONSTRAINT addresses_longitude_check CHECK (((longitude >= ('-180'::integer)::numeric) AND (longitude <= (180)::numeric))),
    CONSTRAINT addresses_phone_check CHECK ((phone ~ '^\+234[0-9]{10}$'::text))
);


--
-- Name: TABLE addresses; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.addresses IS '[identity] Saved delivery and business addresses, owned by exactly one user or one business.';


--
-- Name: auth_events; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.auth_events (
    id bigint NOT NULL,
    user_id uuid,
    kind text NOT NULL,
    ip inet,
    user_agent text,
    device_id text,
    state_code text,
    meta jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT auth_events_kind_check CHECK ((kind = ANY (ARRAY['otp_requested'::text, 'otp_failed'::text, 'login_succeeded'::text, 'login_failed'::text, 'mfa_enrolled'::text, 'mfa_removed'::text, 'mfa_failed'::text, 'step_up_required'::text, 'step_up_succeeded'::text, 'refresh_reuse_detected'::text, 'session_revoked'::text, 'password_changed'::text, 'phone_changed'::text, 'email_changed'::text, 'role_granted'::text, 'role_revoked'::text, 'permission_changed'::text, 'reveal_id'::text, 'rate_limited'::text, 'access_denied'::text, 'staff_exited'::text])))
)
PARTITION BY RANGE (created_at);


--
-- Name: TABLE auth_events; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.auth_events IS '[identity] Append-only security events (sign-ins, failures, MFA, token reuse, role changes). Partitioned monthly.';


--
-- Name: auth_events_default; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.auth_events_default (
    id bigint NOT NULL,
    user_id uuid,
    kind text NOT NULL,
    ip inet,
    user_agent text,
    device_id text,
    state_code text,
    meta jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT auth_events_kind_check CHECK ((kind = ANY (ARRAY['otp_requested'::text, 'otp_failed'::text, 'login_succeeded'::text, 'login_failed'::text, 'mfa_enrolled'::text, 'mfa_removed'::text, 'mfa_failed'::text, 'step_up_required'::text, 'step_up_succeeded'::text, 'refresh_reuse_detected'::text, 'session_revoked'::text, 'password_changed'::text, 'phone_changed'::text, 'email_changed'::text, 'role_granted'::text, 'role_revoked'::text, 'permission_changed'::text, 'reveal_id'::text, 'rate_limited'::text, 'access_denied'::text, 'staff_exited'::text])))
);


--
-- Name: auth_events_id_seq; Type: SEQUENCE; Schema: identity; Owner: -
--

ALTER TABLE identity.auth_events ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME identity.auth_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: auth_rate_limits; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.auth_rate_limits (
    key text NOT NULL,
    window_start timestamp with time zone NOT NULL,
    count integer DEFAULT 0 NOT NULL,
    CONSTRAINT auth_rate_limits_count_check CHECK ((count >= 0))
);


--
-- Name: TABLE auth_rate_limits; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.auth_rate_limits IS '[identity] Window counters for OTP sends and sign-in attempts (per number, IP, device, global). No Redis needed.';


--
-- Name: consents; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.consents (
    id bigint NOT NULL,
    user_id uuid,
    anonymous_id uuid,
    purpose text NOT NULL,
    granted boolean NOT NULL,
    policy_version text NOT NULL,
    source text NOT NULL,
    recorded_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT consents_check CHECK (((user_id IS NOT NULL) OR (anonymous_id IS NOT NULL))),
    CONSTRAINT consents_purpose_check CHECK ((purpose = ANY (ARRAY['analytics'::text, 'personalisation'::text, 'marketing_sms'::text, 'marketing_email'::text, 'marketing_push'::text, 'marketing_whatsapp'::text, 'location_tracking'::text, 'terms'::text, 'privacy'::text]))),
    CONSTRAINT consents_source_check CHECK ((source = ANY (ARRAY['web'::text, 'app'::text, 'unsubscribe_link'::text, 'sms_stop'::text, 'staff'::text, 'import'::text])))
);


--
-- Name: TABLE consents; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.consents IS '[identity] Append-only record of consent given or withdrawn (NDPA). The latest row per subject and purpose wins.';


--
-- Name: consents_id_seq; Type: SEQUENCE; Schema: identity; Owner: -
--

ALTER TABLE identity.consents ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME identity.consents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: nigerian_lgas; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.nigerian_lgas (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    state_code text NOT NULL,
    name text NOT NULL
);


--
-- Name: TABLE nigerian_lgas; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.nigerian_lgas IS '[identity] Lookup: local government areas per state, so delivery fees match consistent spellings.';


--
-- Name: nigerian_states; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.nigerian_states (
    code text NOT NULL,
    name text NOT NULL
);


--
-- Name: TABLE nigerian_states; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.nigerian_states IS '[identity] Lookup: the 36 states and the FCT, used by addresses, zones and listings.';


--
-- Name: permissions; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.permissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    key text NOT NULL,
    module text NOT NULL,
    description text,
    is_sensitive boolean DEFAULT false NOT NULL,
    CONSTRAINT permissions_key_check CHECK ((key ~ '^[a-z_-]+\.[a-z_]+$'::text))
);


--
-- Name: TABLE permissions; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.permissions IS '[identity] Actions per staff module (module.action). Sensitive ones need a recent MFA step-up.';


--
-- Name: privacy_requests; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.privacy_requests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    kind text NOT NULL,
    status text DEFAULT 'requested'::text NOT NULL,
    cancel_until timestamp with time zone,
    file_id uuid,
    completed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT privacy_requests_check CHECK (((status <> 'completed'::text) OR (completed_at IS NOT NULL))),
    CONSTRAINT privacy_requests_kind_check CHECK ((kind = ANY (ARRAY['export'::text, 'delete'::text]))),
    CONSTRAINT privacy_requests_status_check CHECK ((status = ANY (ARRAY['requested'::text, 'cancelled'::text, 'processing'::text, 'completed'::text, 'failed'::text])))
);


--
-- Name: TABLE privacy_requests; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.privacy_requests IS '[identity] Data export and account deletion requests (deletion has a 7-day cancel window, then anonymisation).';


--
-- Name: role_permissions; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.role_permissions (
    role_id uuid NOT NULL,
    permission_id uuid NOT NULL
);


--
-- Name: TABLE role_permissions; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.role_permissions IS '[identity] Which permissions each role grants.';


--
-- Name: roles; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    key text NOT NULL,
    name text NOT NULL,
    description text,
    is_system boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT roles_key_check CHECK ((key ~ '^[a-z][a-z0-9_]*$'::text))
);


--
-- Name: TABLE roles; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.roles IS '[identity] Staff roles. Proposed templates in docs/identity-access.md §7.2; the owner approves the final list.';


--
-- Name: service_clients; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.service_clients (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    secret_hash bytea NOT NULL,
    scopes text[] DEFAULT '{}'::text[] NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    rotated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT service_clients_status_check CHECK ((status = ANY (ARRAY['active'::text, 'disabled'::text])))
);


--
-- Name: TABLE service_clients; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.service_clients IS '[identity] Machine clients (the Python recommender) using client credentials on the internal listener.';


--
-- Name: staff_invites; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.staff_invites (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    staff_user_id uuid NOT NULL,
    token_hash bytea NOT NULL,
    created_by uuid NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE staff_invites; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.staff_invites IS '[identity] Single-use invite links for new staff (set password, enrol TOTP).';


--
-- Name: staff_members; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.staff_members (
    user_id uuid NOT NULL,
    employee_no text NOT NULL,
    department text NOT NULL,
    job_title text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    hired_on date,
    exited_on date,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT staff_members_check CHECK (((exited_on IS NULL) OR (hired_on IS NULL) OR (exited_on >= hired_on))),
    CONSTRAINT staff_members_check1 CHECK (((status <> 'exited'::text) OR (exited_on IS NOT NULL))),
    CONSTRAINT staff_members_status_check CHECK ((status = ANY (ARRAY['active'::text, 'on_leave'::text, 'suspended'::text, 'exited'::text])))
);


--
-- Name: TABLE staff_members; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.staff_members IS '[identity] TechShop employees. A staff member is a user with an employee record.';


--
-- Name: staff_role_scopes; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.staff_role_scopes (
    user_id uuid NOT NULL,
    role_id uuid NOT NULL,
    scope_type text NOT NULL,
    scope_id uuid NOT NULL,
    CONSTRAINT staff_role_scopes_scope_type_check CHECK ((scope_type = ANY (ARRAY['store'::text, 'warehouse'::text, 'zone'::text])))
);


--
-- Name: TABLE staff_role_scopes; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.staff_role_scopes IS '[identity] Limits a staff role to specific stores, warehouses or delivery zones. No rows = unrestricted.';


--
-- Name: staff_roles; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.staff_roles (
    user_id uuid NOT NULL,
    role_id uuid NOT NULL,
    granted_by uuid,
    granted_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT staff_roles_check CHECK ((granted_by IS DISTINCT FROM user_id))
);


--
-- Name: TABLE staff_roles; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.staff_roles IS '[identity] Roles held by each staff member. Nobody grants a role to themselves.';


--
-- Name: user_mfa_factors; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.user_mfa_factors (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    kind text DEFAULT 'totp'::text NOT NULL,
    secret_encrypted bytea NOT NULL,
    encryption_key_id uuid,
    last_used_step bigint,
    confirmed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT user_mfa_factors_kind_check CHECK ((kind = 'totp'::text))
);


--
-- Name: TABLE user_mfa_factors; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.user_mfa_factors IS '[identity] Authenticator-app (TOTP) secrets, envelope-encrypted. last_used_step stops a code being used twice.';


--
-- Name: user_recovery_codes; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.user_recovery_codes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    code_hash bytea NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE user_recovery_codes; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.user_recovery_codes IS '[identity] Single-use MFA recovery codes, stored hashed.';


--
-- Name: user_sessions; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.user_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    family_id uuid NOT NULL,
    refresh_token_hash bytea NOT NULL,
    replaced_by uuid,
    aud text NOT NULL,
    platform text DEFAULT 'web'::text NOT NULL,
    amr text[] DEFAULT '{}'::text[] NOT NULL,
    device_id text,
    device_name text,
    app_version text,
    user_agent text,
    ip inet,
    mfa_at timestamp with time zone,
    last_used_at timestamp with time zone DEFAULT now() NOT NULL,
    idle_expires_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone,
    revoked_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT user_sessions_aud_check CHECK ((aud = ANY (ARRAY['market'::text, 'wholesale'::text, 'seller'::text, 'staff'::text, 'customer_app'::text, 'logistics'::text]))),
    CONSTRAINT user_sessions_check CHECK ((idle_expires_at <= expires_at)),
    CONSTRAINT user_sessions_check1 CHECK (((revoked_at IS NULL) OR (revoked_reason IS NOT NULL))),
    CONSTRAINT user_sessions_platform_check CHECK ((platform = ANY (ARRAY['web'::text, 'ios'::text, 'android'::text]))),
    CONSTRAINT user_sessions_revoked_reason_check CHECK ((revoked_reason = ANY (ARRAY['logout'::text, 'logout_all'::text, 'reuse_detected'::text, 'password_changed'::text, 'admin'::text, 'staff_exited'::text, 'expired'::text])))
);


--
-- Name: TABLE user_sessions; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.user_sessions IS '[identity] Refresh-token sessions per device and app. Tokens stored only as SHA-256 hashes; rotation tracked by family.';


--
-- Name: users; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email public.citext,
    phone text,
    password_hash text,
    first_name text NOT NULL,
    last_name text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    email_verified_at timestamp with time zone,
    phone_verified_at timestamp with time zone,
    marketing_opt_in boolean DEFAULT false NOT NULL,
    last_login_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT users_check CHECK (((email IS NOT NULL) OR (phone IS NOT NULL) OR (status = 'deleted'::text))),
    CONSTRAINT users_phone_check CHECK ((phone ~ '^\+234[0-9]{10}$'::text)),
    CONSTRAINT users_status_check CHECK ((status = ANY (ARRAY['active'::text, 'suspended'::text, 'deleted'::text])))
);


--
-- Name: TABLE users; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.users IS '[identity] Every person who signs in: customers, seller staff, business buyers, riders and TechShop staff. Phone in E.164 (+234…).';


--
-- Name: verification_codes; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.verification_codes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    channel text NOT NULL,
    destination text NOT NULL,
    purpose text NOT NULL,
    code_hash bytea NOT NULL,
    attempts smallint DEFAULT 0 NOT NULL,
    ip inet,
    user_agent text,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT verification_codes_attempts_check CHECK (((attempts >= 0) AND (attempts <= 5))),
    CONSTRAINT verification_codes_channel_check CHECK ((channel = ANY (ARRAY['sms'::text, 'email'::text]))),
    CONSTRAINT verification_codes_check CHECK ((expires_at > created_at)),
    CONSTRAINT verification_codes_purpose_check CHECK ((purpose = ANY (ARRAY['signup'::text, 'login'::text, 'password_reset'::text, 'phone_change'::text, 'email_change'::text, 'new_device'::text, 'staff_mfa'::text])))
);


--
-- Name: TABLE verification_codes; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.verification_codes IS '[identity] One-time codes sent by SMS or email. Stored as HMAC; at most 5 attempts. Delivery codes live on logistics.delivery_jobs.';


--
-- Name: webauthn_credentials; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.webauthn_credentials (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    credential_id bytea NOT NULL,
    public_key bytea NOT NULL,
    sign_count bigint DEFAULT 0 NOT NULL,
    transports text[] DEFAULT '{}'::text[] NOT NULL,
    name text,
    last_used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT webauthn_credentials_sign_count_check CHECK ((sign_count >= 0))
);


--
-- Name: TABLE webauthn_credentials; Type: COMMENT; Schema: identity; Owner: -
--

COMMENT ON TABLE identity.webauthn_credentials IS '[identity] Passkeys (planned for v2; created now so there are no migration surprises later).';


--
-- Name: bin_locations; Type: TABLE; Schema: inventory; Owner: -
--

CREATE TABLE inventory.bin_locations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    warehouse_id uuid NOT NULL,
    code text NOT NULL,
    zone text
);


--
-- Name: TABLE bin_locations; Type: COMMENT; Schema: inventory; Owner: -
--

COMMENT ON TABLE inventory.bin_locations IS '[inventory] Shelf/bin codes inside a warehouse (optional), used on pick lists.';


--
-- Name: device_units; Type: TABLE; Schema: inventory; Owner: -
--

CREATE TABLE inventory.device_units (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    variant_id uuid NOT NULL,
    condition text NOT NULL,
    grade text,
    imei text,
    serial_number text,
    warehouse_id uuid,
    owner_seller_id uuid NOT NULL,
    cost_kobo bigint,
    status text DEFAULT 'in_stock'::text NOT NULL,
    acquired_via text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT device_units_acquired_via_check CHECK ((acquired_via = ANY (ARRAY['purchase'::text, 'trade_in'::text, 'return'::text, 'consignment'::text]))),
    CONSTRAINT device_units_check CHECK (((imei IS NOT NULL) OR (serial_number IS NOT NULL))),
    CONSTRAINT device_units_condition_check CHECK ((condition = ANY (ARRAY['new'::text, 'uk_used'::text, 'refurbished'::text, 'open_box'::text]))),
    CONSTRAINT device_units_cost_kobo_check CHECK ((cost_kobo >= 0)),
    CONSTRAINT device_units_grade_check CHECK ((grade = ANY (ARRAY['A'::text, 'B'::text, 'C'::text]))),
    CONSTRAINT device_units_imei_check CHECK ((imei ~ '^[0-9]{15}$'::text)),
    CONSTRAINT device_units_status_check CHECK ((status = ANY (ARRAY['in_stock'::text, 'reserved'::text, 'sold'::text, 'in_repair'::text, 'returned'::text, 'quarantined'::text, 'written_off'::text])))
);


--
-- Name: TABLE device_units; Type: COMMENT; Schema: inventory; Owner: -
--

COMMENT ON TABLE inventory.device_units IS '[inventory] Individually tracked devices (IMEI/serial): picking, warranty, returns, theft checks, trade-ins.';


--
-- Name: inventory_costs; Type: TABLE; Schema: inventory; Owner: -
--

CREATE TABLE inventory.inventory_costs (
    variant_id uuid NOT NULL,
    condition text NOT NULL,
    warehouse_id uuid NOT NULL,
    owner_seller_id uuid NOT NULL,
    avg_cost_kobo bigint NOT NULL,
    quantity_basis integer NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT inventory_costs_avg_cost_kobo_check CHECK ((avg_cost_kobo >= 0)),
    CONSTRAINT inventory_costs_condition_check CHECK ((condition = ANY (ARRAY['new'::text, 'uk_used'::text, 'refurbished'::text, 'open_box'::text]))),
    CONSTRAINT inventory_costs_quantity_basis_check CHECK ((quantity_basis >= 0))
);


--
-- Name: TABLE inventory_costs; Type: COMMENT; Schema: inventory; Owner: -
--

COMMENT ON TABLE inventory.inventory_costs IS '[inventory] Weighted average cost for non-serialised stock (COGS once finance confirms the method).';


--
-- Name: inventory_count_lines; Type: TABLE; Schema: inventory; Owner: -
--

CREATE TABLE inventory.inventory_count_lines (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    count_id uuid NOT NULL,
    variant_id uuid NOT NULL,
    condition text NOT NULL,
    owner_seller_id uuid NOT NULL,
    bin_id uuid,
    expected_qty integer NOT NULL,
    counted_qty integer,
    variance integer GENERATED ALWAYS AS ((counted_qty - expected_qty)) STORED,
    CONSTRAINT inventory_count_lines_condition_check CHECK ((condition = ANY (ARRAY['new'::text, 'uk_used'::text, 'refurbished'::text, 'open_box'::text]))),
    CONSTRAINT inventory_count_lines_counted_qty_check CHECK ((counted_qty >= 0)),
    CONSTRAINT inventory_count_lines_expected_qty_check CHECK ((expected_qty >= 0))
);


--
-- Name: TABLE inventory_count_lines; Type: COMMENT; Schema: inventory; Owner: -
--

COMMENT ON TABLE inventory.inventory_count_lines IS '[inventory] Expected vs counted per item in a cycle count.';


--
-- Name: inventory_counts; Type: TABLE; Schema: inventory; Owner: -
--

CREATE TABLE inventory.inventory_counts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    warehouse_id uuid NOT NULL,
    status text DEFAULT 'open'::text NOT NULL,
    counted_by uuid NOT NULL,
    approved_by uuid,
    approved_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT inventory_counts_check CHECK ((approved_by IS DISTINCT FROM counted_by)),
    CONSTRAINT inventory_counts_check1 CHECK (((status <> 'approved'::text) OR (approved_by IS NOT NULL))),
    CONSTRAINT inventory_counts_status_check CHECK ((status = ANY (ARRAY['open'::text, 'submitted'::text, 'approved'::text, 'cancelled'::text])))
);


--
-- Name: TABLE inventory_counts; Type: COMMENT; Schema: inventory; Owner: -
--

COMMENT ON TABLE inventory.inventory_counts IS '[inventory] Cycle counts (blind). Variances above a threshold need a different person to approve.';


--
-- Name: inventory_levels; Type: TABLE; Schema: inventory; Owner: -
--

CREATE TABLE inventory.inventory_levels (
    warehouse_id uuid NOT NULL,
    variant_id uuid NOT NULL,
    condition text NOT NULL,
    owner_seller_id uuid NOT NULL,
    on_hand integer DEFAULT 0 NOT NULL,
    reserved integer DEFAULT 0 NOT NULL,
    reorder_point integer DEFAULT 0 NOT NULL,
    bin_id uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT inventory_levels_check CHECK ((reserved <= on_hand)),
    CONSTRAINT inventory_levels_condition_check CHECK ((condition = ANY (ARRAY['new'::text, 'uk_used'::text, 'refurbished'::text, 'open_box'::text]))),
    CONSTRAINT inventory_levels_on_hand_check CHECK ((on_hand >= 0)),
    CONSTRAINT inventory_levels_reorder_point_check CHECK ((reorder_point >= 0)),
    CONSTRAINT inventory_levels_reserved_check CHECK ((reserved >= 0))
);


--
-- Name: TABLE inventory_levels; Type: COMMENT; Schema: inventory; Owner: -
--

COMMENT ON TABLE inventory.inventory_levels IS '[inventory] Current stock per location × variant × condition × owner. Changed only together with a stock movement.';


--
-- Name: stock_movements; Type: TABLE; Schema: inventory; Owner: -
--

CREATE TABLE inventory.stock_movements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    warehouse_id uuid NOT NULL,
    variant_id uuid NOT NULL,
    condition text NOT NULL,
    owner_seller_id uuid NOT NULL,
    quantity_delta integer NOT NULL,
    reason text NOT NULL,
    device_unit_id uuid,
    reference_type text,
    reference_id uuid,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT stock_movements_condition_check CHECK ((condition = ANY (ARRAY['new'::text, 'uk_used'::text, 'refurbished'::text, 'open_box'::text]))),
    CONSTRAINT stock_movements_quantity_delta_check CHECK ((quantity_delta <> 0)),
    CONSTRAINT stock_movements_reason_check CHECK ((reason = ANY (ARRAY['purchase_receipt'::text, 'sale'::text, 'return'::text, 'transfer_in'::text, 'transfer_out'::text, 'adjustment'::text, 'trade_in'::text, 'write_off'::text, 'count_variance'::text])))
);


--
-- Name: TABLE stock_movements; Type: COMMENT; Schema: inventory; Owner: -
--

COMMENT ON TABLE inventory.stock_movements IS '[inventory] Append-only stock ledger; every change to inventory_levels has a movement (nightly check: sum = on hand).';


--
-- Name: stock_reservations; Type: TABLE; Schema: inventory; Owner: -
--

CREATE TABLE inventory.stock_reservations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    order_line_id uuid NOT NULL,
    warehouse_id uuid,
    listing_id uuid NOT NULL,
    variant_id uuid NOT NULL,
    condition text NOT NULL,
    owner_seller_id uuid NOT NULL,
    quantity integer NOT NULL,
    status text DEFAULT 'reserved'::text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT stock_reservations_condition_check CHECK ((condition = ANY (ARRAY['new'::text, 'uk_used'::text, 'refurbished'::text, 'open_box'::text]))),
    CONSTRAINT stock_reservations_quantity_check CHECK ((quantity > 0)),
    CONSTRAINT stock_reservations_status_check CHECK ((status = ANY (ARRAY['reserved'::text, 'committed'::text, 'released'::text])))
);


--
-- Name: TABLE stock_reservations; Type: COMMENT; Schema: inventory; Owner: -
--

COMMENT ON TABLE inventory.stock_reservations IS '[inventory] Stock held for an unpaid order (warehouse level or seller-held listing stock); committed on payment, released on expiry.';


--
-- Name: stock_transfer_lines; Type: TABLE; Schema: inventory; Owner: -
--

CREATE TABLE inventory.stock_transfer_lines (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    transfer_id uuid NOT NULL,
    variant_id uuid NOT NULL,
    condition text NOT NULL,
    owner_seller_id uuid NOT NULL,
    quantity integer NOT NULL,
    received_quantity integer,
    CONSTRAINT stock_transfer_lines_condition_check CHECK ((condition = ANY (ARRAY['new'::text, 'uk_used'::text, 'refurbished'::text, 'open_box'::text]))),
    CONSTRAINT stock_transfer_lines_quantity_check CHECK ((quantity > 0)),
    CONSTRAINT stock_transfer_lines_received_quantity_check CHECK ((received_quantity >= 0))
);


--
-- Name: TABLE stock_transfer_lines; Type: COMMENT; Schema: inventory; Owner: -
--

COMMENT ON TABLE inventory.stock_transfer_lines IS '[inventory] Items in a stock transfer.';


--
-- Name: stock_transfers; Type: TABLE; Schema: inventory; Owner: -
--

CREATE TABLE inventory.stock_transfers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    from_warehouse_id uuid NOT NULL,
    to_warehouse_id uuid NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    created_by uuid NOT NULL,
    shipped_at timestamp with time zone,
    received_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT stock_transfers_check CHECK ((from_warehouse_id <> to_warehouse_id)),
    CONSTRAINT stock_transfers_check1 CHECK (((status <> 'received'::text) OR (received_at IS NOT NULL))),
    CONSTRAINT stock_transfers_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'in_transit'::text, 'received'::text, 'cancelled'::text])))
);


--
-- Name: TABLE stock_transfers; Type: COMMENT; Schema: inventory; Owner: -
--

COMMENT ON TABLE inventory.stock_transfers IS '[inventory] Moving stock between locations.';


--
-- Name: warehouses; Type: TABLE; Schema: inventory; Owner: -
--

CREATE TABLE inventory.warehouses (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    kind text NOT NULL,
    address text NOT NULL,
    city text NOT NULL,
    state_code text NOT NULL,
    latitude numeric(9,6),
    longitude numeric(9,6),
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT warehouses_kind_check CHECK ((kind = ANY (ARRAY['warehouse'::text, 'store'::text, 'hub'::text])))
);


--
-- Name: TABLE warehouses; Type: COMMENT; Schema: inventory; Owner: -
--

COMMENT ON TABLE inventory.warehouses IS '[inventory] Physical locations: warehouses, retail stores (POS) and dispatch hubs.';


--
-- Name: delivery_events; Type: TABLE; Schema: logistics; Owner: -
--

CREATE TABLE logistics.delivery_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    delivery_job_id uuid NOT NULL,
    client_event_id uuid,
    kind text NOT NULL,
    reason text,
    latitude numeric(9,6),
    longitude numeric(9,6),
    accuracy_m real,
    note text,
    file_id uuid,
    created_by uuid,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT delivery_events_check CHECK (((kind <> 'attempt_failed'::text) OR (reason IS NOT NULL))),
    CONSTRAINT delivery_events_kind_check CHECK ((kind = ANY (ARRAY['assigned'::text, 'picked_up'::text, 'en_route'::text, 'location'::text, 'attempt_failed'::text, 'delivered'::text, 'returned'::text, 'note'::text]))),
    CONSTRAINT delivery_events_reason_check CHECK ((reason = ANY (ARRAY['customer_unreachable'::text, 'customer_refused'::text, 'wrong_address'::text, 'rescheduled'::text, 'unsafe'::text, 'other'::text])))
);


--
-- Name: TABLE delivery_events; Type: COMMENT; Schema: logistics; Owner: -
--

COMMENT ON TABLE logistics.delivery_events IS '[logistics] Append-only trail for a job. client_event_id makes offline syncs from the app idempotent.';


--
-- Name: delivery_jobs; Type: TABLE; Schema: logistics; Owner: -
--

CREATE TABLE logistics.delivery_jobs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    kind text DEFAULT 'delivery'::text NOT NULL,
    fulfilment_id uuid,
    return_request_id uuid,
    rider_id uuid,
    zone_id uuid NOT NULL,
    status text DEFAULT 'unassigned'::text NOT NULL,
    pickup_warehouse_id uuid,
    pickup_address jsonb,
    dropoff_latitude numeric(9,6),
    dropoff_longitude numeric(9,6),
    route_sequence smallint,
    window_start timestamp with time zone,
    window_end timestamp with time zone,
    attempts smallint DEFAULT 0 NOT NULL,
    assigned_by uuid,
    assigned_at timestamp with time zone,
    delivered_at timestamp with time zone,
    recipient_name text,
    otp_hash bytea,
    otp_expires_at timestamp with time zone,
    otp_attempts smallint DEFAULT 0 NOT NULL,
    otp_verified_at timestamp with time zone,
    proof_method text,
    proof_file_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT delivery_jobs_attempts_check CHECK (((attempts >= 0) AND (attempts <= 5))),
    CONSTRAINT delivery_jobs_check CHECK (((window_end IS NULL) OR (window_end > window_start))),
    CONSTRAINT delivery_jobs_check1 CHECK (((status = ANY (ARRAY['unassigned'::text, 'cancelled'::text])) OR (rider_id IS NOT NULL))),
    CONSTRAINT delivery_jobs_check2 CHECK (((kind <> 'delivery'::text) OR (fulfilment_id IS NOT NULL))),
    CONSTRAINT delivery_jobs_check3 CHECK (((kind <> 'return_pickup'::text) OR (return_request_id IS NOT NULL))),
    CONSTRAINT delivery_jobs_check4 CHECK (((pickup_warehouse_id IS NOT NULL) OR (pickup_address IS NOT NULL))),
    CONSTRAINT delivery_jobs_check5 CHECK (((status <> 'delivered'::text) OR ((delivered_at IS NOT NULL) AND (((proof_method = 'otp'::text) AND (otp_verified_at IS NOT NULL)) OR ((proof_method = ANY (ARRAY['photo'::text, 'photo_offline'::text])) AND (proof_file_id IS NOT NULL)))))),
    CONSTRAINT delivery_jobs_kind_check CHECK ((kind = ANY (ARRAY['delivery'::text, 'return_pickup'::text, 'seller_pickup'::text]))),
    CONSTRAINT delivery_jobs_otp_attempts_check CHECK (((otp_attempts >= 0) AND (otp_attempts <= 5))),
    CONSTRAINT delivery_jobs_proof_method_check CHECK ((proof_method = ANY (ARRAY['otp'::text, 'photo'::text, 'photo_offline'::text]))),
    CONSTRAINT delivery_jobs_status_check CHECK ((status = ANY (ARRAY['unassigned'::text, 'assigned'::text, 'picked_up'::text, 'en_route'::text, 'delivered'::text, 'failed'::text, 'returned'::text, 'cancelled'::text])))
);


--
-- Name: TABLE delivery_jobs; Type: COMMENT; Schema: logistics; Owner: -
--

COMMENT ON TABLE logistics.delivery_jobs IS '[logistics] A delivery or pickup run. Delivered requires a verified delivery code or photo proof. The delivery code lives only here.';


--
-- Name: delivery_rates; Type: TABLE; Schema: logistics; Owner: -
--

CREATE TABLE logistics.delivery_rates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    zone_id uuid NOT NULL,
    fulfilled_by text NOT NULL,
    max_weight_grams integer NOT NULL,
    fee_kobo bigint NOT NULL,
    eta_min_days smallint NOT NULL,
    eta_max_days smallint NOT NULL,
    valid_from date DEFAULT CURRENT_DATE NOT NULL,
    CONSTRAINT delivery_rates_check CHECK ((eta_max_days >= eta_min_days)),
    CONSTRAINT delivery_rates_eta_min_days_check CHECK ((eta_min_days >= 0)),
    CONSTRAINT delivery_rates_fee_kobo_check CHECK ((fee_kobo >= 0)),
    CONSTRAINT delivery_rates_fulfilled_by_check CHECK ((fulfilled_by = ANY (ARRAY['techshop'::text, 'seller'::text]))),
    CONSTRAINT delivery_rates_max_weight_grams_check CHECK ((max_weight_grams > 0))
);


--
-- Name: TABLE delivery_rates; Type: COMMENT; Schema: logistics; Owner: -
--

COMMENT ON TABLE logistics.delivery_rates IS '[logistics] Delivery fee and ETA per zone and weight band.';


--
-- Name: delivery_zones; Type: TABLE; Schema: logistics; Owner: -
--

CREATE TABLE logistics.delivery_zones (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    state_code text NOT NULL,
    lgas text[] DEFAULT '{}'::text[] NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE delivery_zones; Type: COMMENT; Schema: logistics; Owner: -
--

COMMENT ON TABLE logistics.delivery_zones IS '[logistics] Delivery areas (state + LGAs) used for fees, ETAs and rider assignment.';


--
-- Name: rider_devices; Type: TABLE; Schema: logistics; Owner: -
--

CREATE TABLE logistics.rider_devices (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    rider_id uuid NOT NULL,
    device_id text NOT NULL,
    platform text NOT NULL,
    approved_by uuid,
    approved_at timestamp with time zone,
    revoked_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT rider_devices_check CHECK ((approved_by IS DISTINCT FROM rider_id)),
    CONSTRAINT rider_devices_platform_check CHECK ((platform = ANY (ARRAY['ios'::text, 'android'::text])))
);


--
-- Name: TABLE rider_devices; Type: COMMENT; Schema: logistics; Owner: -
--

COMMENT ON TABLE logistics.rider_devices IS '[logistics] Phones bound to a rider; a dispatcher approves each one and can revoke a lost phone.';


--
-- Name: rider_location_pings; Type: TABLE; Schema: logistics; Owner: -
--

CREATE TABLE logistics.rider_location_pings (
    rider_id uuid NOT NULL,
    latitude numeric(9,6) NOT NULL,
    longitude numeric(9,6) NOT NULL,
    accuracy_m real,
    speed real,
    heading real,
    recorded_at timestamp with time zone NOT NULL,
    received_at timestamp with time zone DEFAULT now() NOT NULL
)
PARTITION BY RANGE (recorded_at);


--
-- Name: TABLE rider_location_pings; Type: COMMENT; Schema: logistics; Owner: -
--

COMMENT ON TABLE logistics.rider_location_pings IS '[logistics] Full location trail while on shift (optional). Partitioned monthly; about 90 days retention.';


--
-- Name: rider_location_pings_default; Type: TABLE; Schema: logistics; Owner: -
--

CREATE TABLE logistics.rider_location_pings_default (
    rider_id uuid NOT NULL,
    latitude numeric(9,6) NOT NULL,
    longitude numeric(9,6) NOT NULL,
    accuracy_m real,
    speed real,
    heading real,
    recorded_at timestamp with time zone NOT NULL,
    received_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: rider_locations; Type: TABLE; Schema: logistics; Owner: -
--

CREATE TABLE logistics.rider_locations (
    rider_id uuid NOT NULL,
    latitude numeric(9,6) NOT NULL,
    longitude numeric(9,6) NOT NULL,
    accuracy_m real,
    job_id uuid,
    recorded_at timestamp with time zone NOT NULL
);


--
-- Name: TABLE rider_locations; Type: COMMENT; Schema: logistics; Owner: -
--

COMMENT ON TABLE logistics.rider_locations IS '[logistics] Latest known position per rider (live tracking and the dispatch map).';


--
-- Name: rider_shifts; Type: TABLE; Schema: logistics; Owner: -
--

CREATE TABLE logistics.rider_shifts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    rider_id uuid NOT NULL,
    device_session_id uuid,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    ended_at timestamp with time zone,
    start_latitude numeric(9,6),
    start_longitude numeric(9,6),
    location_consent_at timestamp with time zone NOT NULL,
    CONSTRAINT rider_shifts_check CHECK (((ended_at IS NULL) OR (ended_at > started_at)))
);


--
-- Name: TABLE rider_shifts; Type: COMMENT; Schema: logistics; Owner: -
--

COMMENT ON TABLE logistics.rider_shifts IS '[logistics] Shift history; location is only shared while a shift is open and with consent.';


--
-- Name: rider_zones; Type: TABLE; Schema: logistics; Owner: -
--

CREATE TABLE logistics.rider_zones (
    rider_id uuid NOT NULL,
    zone_id uuid NOT NULL
);


--
-- Name: TABLE rider_zones; Type: COMMENT; Schema: logistics; Owner: -
--

COMMENT ON TABLE logistics.rider_zones IS '[logistics] Zones each rider works in (assignment by zone).';


--
-- Name: riders; Type: TABLE; Schema: logistics; Owner: -
--

CREATE TABLE logistics.riders (
    user_id uuid NOT NULL,
    vehicle_type text NOT NULL,
    plate_number text,
    home_warehouse_id uuid NOT NULL,
    status text DEFAULT 'off_shift'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT riders_status_check CHECK ((status = ANY (ARRAY['off_shift'::text, 'available'::text, 'on_delivery'::text, 'suspended'::text]))),
    CONSTRAINT riders_vehicle_type_check CHECK ((vehicle_type = ANY (ARRAY['bike'::text, 'car'::text, 'van'::text])))
);


--
-- Name: TABLE riders; Type: COMMENT; Schema: logistics; Owner: -
--

COMMENT ON TABLE logistics.riders IS '[logistics] Dispatch riders (staff using the logistics app).';


--
-- Name: campaign_recipients; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.campaign_recipients (
    campaign_id uuid NOT NULL,
    user_id uuid NOT NULL,
    variant_id uuid,
    holdout boolean DEFAULT false NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    CONSTRAINT campaign_recipients_check CHECK (((NOT holdout) OR (variant_id IS NULL))),
    CONSTRAINT campaign_recipients_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'sent'::text, 'skipped'::text, 'failed'::text])))
);


--
-- Name: TABLE campaign_recipients; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.campaign_recipients IS '[marketing] Recipient snapshot taken when sending starts (stable and auditable); holdout rows get nothing.';


--
-- Name: campaign_variants; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.campaign_variants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    campaign_id uuid NOT NULL,
    label text NOT NULL,
    channel text NOT NULL,
    template_version_id uuid,
    share_pct numeric(5,2) NOT NULL,
    is_winner boolean DEFAULT false NOT NULL,
    CONSTRAINT campaign_variants_channel_check CHECK ((channel = ANY (ARRAY['sms'::text, 'email'::text, 'push'::text, 'whatsapp'::text]))),
    CONSTRAINT campaign_variants_share_pct_check CHECK (((share_pct > (0)::numeric) AND (share_pct <= (100)::numeric)))
);


--
-- Name: TABLE campaign_variants; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.campaign_variants IS '[marketing] Content variants for A/B tests (winner by clicks).';


--
-- Name: campaigns; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.campaigns (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    segment_id uuid NOT NULL,
    exclusions jsonb DEFAULT '{}'::jsonb NOT NULL,
    channels text[] NOT NULL,
    scheduled_at timestamp with time zone,
    holdout_pct numeric(4,2) DEFAULT 5 NOT NULL,
    ab_test jsonb,
    cost_estimate_kobo bigint,
    created_by uuid NOT NULL,
    approved_by uuid,
    approved_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT campaigns_channels_check CHECK (((cardinality(channels) >= 1) AND (channels <@ ARRAY['sms'::text, 'email'::text, 'push'::text, 'whatsapp'::text]))),
    CONSTRAINT campaigns_check CHECK (((approved_by IS DISTINCT FROM created_by) OR (approved_by IS NULL))),
    CONSTRAINT campaigns_check1 CHECK (((status <> 'scheduled'::text) OR (scheduled_at IS NOT NULL))),
    CONSTRAINT campaigns_cost_estimate_kobo_check CHECK ((cost_estimate_kobo >= 0)),
    CONSTRAINT campaigns_holdout_pct_check CHECK (((holdout_pct >= (0)::numeric) AND (holdout_pct <= (50)::numeric))),
    CONSTRAINT campaigns_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'in_review'::text, 'scheduled'::text, 'sending'::text, 'paused'::text, 'sent'::text, 'cancelled'::text])))
);


--
-- Name: TABLE campaigns; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.campaigns IS '[marketing] One-off sends. Large or costly campaigns need approval by someone other than the creator.';


--
-- Name: coupon_codes; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.coupon_codes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    coupon_id uuid NOT NULL,
    code public.citext NOT NULL,
    issued_to_user_id uuid,
    notification_id uuid,
    expires_at timestamp with time zone,
    redeemed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE coupon_codes; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.coupon_codes IS '[marketing] Unique single-use codes issued to one person (e.g. CART-7KQ2M9); a leaked code works once.';


--
-- Name: coupon_redemptions; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.coupon_redemptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    coupon_id uuid NOT NULL,
    coupon_code_id uuid,
    order_id uuid NOT NULL,
    user_id uuid,
    discount_kobo bigint NOT NULL,
    status text DEFAULT 'applied'::text NOT NULL,
    redeemed_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT coupon_redemptions_discount_kobo_check CHECK ((discount_kobo > 0)),
    CONSTRAINT coupon_redemptions_status_check CHECK ((status = ANY (ARRAY['applied'::text, 'released'::text])))
);


--
-- Name: TABLE coupon_redemptions; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.coupon_redemptions IS '[marketing] Each coupon use (one per order); released when the order is cancelled.';


--
-- Name: coupons; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.coupons (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code public.citext,
    promotion_id uuid NOT NULL,
    is_unique_codes boolean DEFAULT false NOT NULL,
    max_redemptions integer,
    redeemed_count integer DEFAULT 0 NOT NULL,
    per_user_limit integer DEFAULT 1 NOT NULL,
    min_order_kobo bigint,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT coupons_check CHECK (((max_redemptions IS NULL) OR (redeemed_count <= max_redemptions))),
    CONSTRAINT coupons_check1 CHECK ((is_unique_codes OR (code IS NOT NULL))),
    CONSTRAINT coupons_max_redemptions_check CHECK ((max_redemptions > 0)),
    CONSTRAINT coupons_min_order_kobo_check CHECK ((min_order_kobo >= 0)),
    CONSTRAINT coupons_per_user_limit_check CHECK ((per_user_limit > 0)),
    CONSTRAINT coupons_redeemed_count_check CHECK ((redeemed_count >= 0))
);


--
-- Name: TABLE coupons; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.coupons IS '[marketing] Coupons: one shared code, or many single-use codes (coupon_codes). Counters are concurrency-safe.';


--
-- Name: flash_claims; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.flash_claims (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    promotion_id uuid NOT NULL,
    listing_id uuid NOT NULL,
    user_id uuid NOT NULL,
    order_id uuid NOT NULL,
    quantity integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT flash_claims_quantity_check CHECK ((quantity > 0))
);


--
-- Name: TABLE flash_claims; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.flash_claims IS '[marketing] Who claimed flash-deal units, for per-user limits and release on cancellation.';


--
-- Name: journey_enrollments; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.journey_enrollments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    journey_id uuid NOT NULL,
    user_id uuid NOT NULL,
    current_step integer DEFAULT 0 NOT NULL,
    next_run_at timestamp with time zone,
    status text DEFAULT 'active'::text NOT NULL,
    exit_reason text,
    context jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT journey_enrollments_check CHECK (((status <> 'active'::text) OR (next_run_at IS NOT NULL))),
    CONSTRAINT journey_enrollments_status_check CHECK ((status = ANY (ARRAY['active'::text, 'exited'::text, 'goal_reached'::text, 'completed'::text, 'failed'::text])))
);


--
-- Name: TABLE journey_enrollments; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.journey_enrollments IS '[marketing] People inside a journey; a worker runs due steps every minute; goal events exit immediately.';


--
-- Name: journey_steps; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.journey_steps (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    journey_id uuid NOT NULL,
    "position" integer NOT NULL,
    kind text NOT NULL,
    config jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT journey_steps_kind_check CHECK ((kind = ANY (ARRAY['wait'::text, 'condition'::text, 'send'::text, 'issue_coupon'::text, 'branch'::text, 'exit'::text]))),
    CONSTRAINT journey_steps_position_check CHECK (("position" >= 0))
);


--
-- Name: TABLE journey_steps; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.journey_steps IS '[marketing] Ordered steps of a journey.';


--
-- Name: journeys; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.journeys (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    trigger text NOT NULL,
    entry_rules jsonb DEFAULT '{}'::jsonb NOT NULL,
    reentry_days integer,
    goal_event text,
    status text DEFAULT 'draft'::text NOT NULL,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT journeys_reentry_days_check CHECK ((reentry_days >= 0)),
    CONSTRAINT journeys_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'active'::text, 'paused'::text, 'archived'::text])))
);


--
-- Name: TABLE journeys; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.journeys IS '[marketing] Automations (abandoned cart, back in stock, welcome…): a trigger plus steps.';


--
-- Name: promotion_listing_prices; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.promotion_listing_prices (
    promotion_id uuid NOT NULL,
    listing_id uuid NOT NULL,
    price_kobo bigint NOT NULL,
    stock_limit integer,
    claimed integer DEFAULT 0 NOT NULL,
    per_user_limit integer,
    CONSTRAINT promotion_listing_prices_check CHECK (((stock_limit IS NULL) OR (claimed <= stock_limit))),
    CONSTRAINT promotion_listing_prices_claimed_check CHECK ((claimed >= 0)),
    CONSTRAINT promotion_listing_prices_per_user_limit_check CHECK ((per_user_limit > 0)),
    CONSTRAINT promotion_listing_prices_price_kobo_check CHECK ((price_kobo > 0)),
    CONSTRAINT promotion_listing_prices_stock_limit_check CHECK ((stock_limit > 0))
);


--
-- Name: TABLE promotion_listing_prices; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.promotion_listing_prices IS '[marketing] Flash-deal prices per listing; claimed can never exceed stock_limit (conditional update).';


--
-- Name: promotion_targets; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.promotion_targets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    promotion_id uuid NOT NULL,
    category_id uuid,
    listing_id uuid,
    CONSTRAINT promotion_targets_check CHECK (((category_id IS NULL) <> (listing_id IS NULL)))
);


--
-- Name: TABLE promotion_targets; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.promotion_targets IS '[marketing] What a promotion applies to (a category or a listing); none = sitewide.';


--
-- Name: promotions; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.promotions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    kind text NOT NULL,
    value integer,
    funded_by text DEFAULT 'techshop'::text NOT NULL,
    starts_at timestamp with time zone NOT NULL,
    ends_at timestamp with time zone NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT promotions_check CHECK ((ends_at > starts_at)),
    CONSTRAINT promotions_check1 CHECK (((kind = ANY (ARRAY['free_delivery'::text, 'flash_price'::text])) OR (value IS NOT NULL))),
    CONSTRAINT promotions_check2 CHECK (((kind <> 'percentage'::text) OR (value <= 90))),
    CONSTRAINT promotions_funded_by_check CHECK ((funded_by = ANY (ARRAY['techshop'::text, 'seller'::text]))),
    CONSTRAINT promotions_kind_check CHECK ((kind = ANY (ARRAY['percentage'::text, 'fixed_amount'::text, 'free_delivery'::text, 'flash_price'::text]))),
    CONSTRAINT promotions_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'scheduled'::text, 'live'::text, 'ended'::text, 'cancelled'::text]))),
    CONSTRAINT promotions_value_check CHECK ((value > 0))
);


--
-- Name: TABLE promotions; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.promotions IS '[marketing] Deals with real start and end times (countdowns read ends_at; no fake timers).';


--
-- Name: segments; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.segments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    rules jsonb NOT NULL,
    created_by uuid NOT NULL,
    last_count integer,
    last_counted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT segments_last_count_check CHECK ((last_count >= 0))
);


--
-- Name: TABLE segments; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.segments IS '[marketing] Saved audiences: JSON rules compiled to SQL over an allowlist of fields.';


--
-- Name: tracked_links; Type: TABLE; Schema: marketing; Owner: -
--

CREATE TABLE marketing.tracked_links (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    url text NOT NULL,
    notification_id uuid,
    campaign_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT tracked_links_url_check CHECK ((url ~ '^https://([a-z0-9-]+\.)*techshop\.ng(/|$)'::text))
);


--
-- Name: TABLE tracked_links; Type: COMMENT; Schema: marketing; Owner: -
--

COMMENT ON TABLE marketing.tracked_links IS '[marketing] Signed click-redirect targets (only techshop.ng URLs, so it can never be an open redirect).';


--
-- Name: message_events; Type: TABLE; Schema: messaging; Owner: -
--

CREATE TABLE messaging.message_events (
    id bigint NOT NULL,
    notification_id uuid NOT NULL,
    kind text NOT NULL,
    meta jsonb DEFAULT '{}'::jsonb NOT NULL,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT message_events_kind_check CHECK ((kind = ANY (ARRAY['delivered'::text, 'bounced'::text, 'complained'::text, 'opened'::text, 'clicked'::text, 'unsubscribed'::text, 'failed'::text])))
)
PARTITION BY RANGE (occurred_at);


--
-- Name: TABLE message_events; Type: COMMENT; Schema: messaging; Owner: -
--

COMMENT ON TABLE messaging.message_events IS '[messaging] Append-only provider callbacks and clicks per message. Partitioned monthly; 13-month retention.';


--
-- Name: message_events_default; Type: TABLE; Schema: messaging; Owner: -
--

CREATE TABLE messaging.message_events_default (
    id bigint NOT NULL,
    notification_id uuid NOT NULL,
    kind text NOT NULL,
    meta jsonb DEFAULT '{}'::jsonb NOT NULL,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT message_events_kind_check CHECK ((kind = ANY (ARRAY['delivered'::text, 'bounced'::text, 'complained'::text, 'opened'::text, 'clicked'::text, 'unsubscribed'::text, 'failed'::text])))
);


--
-- Name: message_events_id_seq; Type: SEQUENCE; Schema: messaging; Owner: -
--

ALTER TABLE messaging.message_events ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME messaging.message_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: message_suppressions; Type: TABLE; Schema: messaging; Owner: -
--

CREATE TABLE messaging.message_suppressions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    channel text NOT NULL,
    address_hash bytea NOT NULL,
    reason text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT message_suppressions_channel_check CHECK ((channel = ANY (ARRAY['sms'::text, 'email'::text, 'push'::text, 'whatsapp'::text]))),
    CONSTRAINT message_suppressions_reason_check CHECK ((reason = ANY (ARRAY['hard_bounce'::text, 'complaint'::text, 'unsubscribed'::text, 'invalid'::text, 'manual'::text])))
);


--
-- Name: TABLE message_suppressions; Type: COMMENT; Schema: messaging; Owner: -
--

COMMENT ON TABLE messaging.message_suppressions IS '[messaging] Addresses never to send to on a channel (bounces, complaints, unsubscribes).';


--
-- Name: message_templates; Type: TABLE; Schema: messaging; Owner: -
--

CREATE TABLE messaging.message_templates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    key text NOT NULL,
    channel text NOT NULL,
    category text NOT NULL,
    locale text DEFAULT 'en'::text NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT message_templates_category_check CHECK ((category = ANY (ARRAY['security'::text, 'delivery'::text, 'orders'::text, 'account'::text, 'seller'::text, 'b2b'::text, 'deals'::text, 'recommendations'::text, 'reminders'::text]))),
    CONSTRAINT message_templates_channel_check CHECK ((channel = ANY (ARRAY['sms'::text, 'email'::text, 'push'::text, 'in_app'::text, 'whatsapp'::text]))),
    CONSTRAINT message_templates_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'active'::text, 'archived'::text])))
);


--
-- Name: TABLE message_templates; Type: COMMENT; Schema: messaging; Owner: -
--

COMMENT ON TABLE messaging.message_templates IS '[messaging] Marketing-authored templates (transactional ones live in code).';


--
-- Name: notification_preferences; Type: TABLE; Schema: messaging; Owner: -
--

CREATE TABLE messaging.notification_preferences (
    user_id uuid NOT NULL,
    category text NOT NULL,
    channel text NOT NULL,
    enabled boolean NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT notification_preferences_category_check CHECK ((category = ANY (ARRAY['orders'::text, 'seller'::text, 'deals'::text, 'recommendations'::text, 'reminders'::text]))),
    CONSTRAINT notification_preferences_channel_check CHECK ((channel = ANY (ARRAY['sms'::text, 'email'::text, 'push'::text, 'whatsapp'::text])))
);


--
-- Name: TABLE notification_preferences; Type: COMMENT; Schema: messaging; Owner: -
--

COMMENT ON TABLE messaging.notification_preferences IS '[messaging] Per category and channel choices. Security and delivery messages cannot be turned off, so they are not listed.';


--
-- Name: notification_settings; Type: TABLE; Schema: messaging; Owner: -
--

CREATE TABLE messaging.notification_settings (
    user_id uuid NOT NULL,
    quiet_start time without time zone DEFAULT '21:00:00'::time without time zone NOT NULL,
    quiet_end time without time zone DEFAULT '08:00:00'::time without time zone NOT NULL,
    quiet_hours boolean DEFAULT true NOT NULL,
    timezone text DEFAULT 'Africa/Lagos'::text NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE notification_settings; Type: COMMENT; Schema: messaging; Owner: -
--

COMMENT ON TABLE messaging.notification_settings IS '[messaging] Quiet hours per user (marketing only).';


--
-- Name: notifications; Type: TABLE; Schema: messaging; Owner: -
--

CREATE TABLE messaging.notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    channel text NOT NULL,
    category text NOT NULL,
    priority text NOT NULL,
    template text NOT NULL,
    template_version_id uuid,
    locale text DEFAULT 'en'::text NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    idempotency_key text NOT NULL,
    campaign_id uuid,
    journey_enrollment_id uuid,
    to_hash bytea,
    provider text,
    provider_message_id text,
    cost_kobo bigint,
    status text DEFAULT 'queued'::text NOT NULL,
    skip_reason text,
    error text,
    scheduled_for timestamp with time zone,
    sent_at timestamp with time zone,
    delivered_at timestamp with time zone,
    clicked_at timestamp with time zone,
    read_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT notifications_category_check CHECK ((category = ANY (ARRAY['security'::text, 'delivery'::text, 'orders'::text, 'account'::text, 'seller'::text, 'b2b'::text, 'deals'::text, 'recommendations'::text, 'reminders'::text]))),
    CONSTRAINT notifications_channel_check CHECK ((channel = ANY (ARRAY['sms'::text, 'email'::text, 'push'::text, 'in_app'::text, 'whatsapp'::text]))),
    CONSTRAINT notifications_check CHECK (((status <> 'skipped'::text) OR (skip_reason IS NOT NULL))),
    CONSTRAINT notifications_check1 CHECK (((user_id IS NOT NULL) OR (to_hash IS NOT NULL))),
    CONSTRAINT notifications_cost_kobo_check CHECK ((cost_kobo >= 0)),
    CONSTRAINT notifications_priority_check CHECK ((priority = ANY (ARRAY['critical'::text, 'transactional'::text, 'marketing'::text]))),
    CONSTRAINT notifications_skip_reason_check CHECK ((skip_reason = ANY (ARRAY['no_consent'::text, 'suppressed'::text, 'capped'::text, 'quiet_hours'::text, 'preference'::text, 'invalid_address'::text, 'duplicate'::text]))),
    CONSTRAINT notifications_status_check CHECK ((status = ANY (ARRAY['queued'::text, 'skipped'::text, 'sent'::text, 'delivered'::text, 'bounced'::text, 'failed'::text, 'read'::text])))
);


--
-- Name: TABLE notifications; Type: COMMENT; Schema: messaging; Owner: -
--

COMMENT ON TABLE messaging.notifications IS '[messaging] Every message to a recipient, any channel. Doubles as the in-app inbox. Bodies trimmed after 90 days.';


--
-- Name: push_tokens; Type: TABLE; Schema: messaging; Owner: -
--

CREATE TABLE messaging.push_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    app text NOT NULL,
    platform text NOT NULL,
    token text NOT NULL,
    provider text DEFAULT 'expo'::text NOT NULL,
    app_version text,
    locale text,
    last_seen_at timestamp with time zone DEFAULT now() NOT NULL,
    revoked_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT push_tokens_app_check CHECK ((app = ANY (ARRAY['customer'::text, 'logistics'::text]))),
    CONSTRAINT push_tokens_platform_check CHECK ((platform = ANY (ARRAY['ios'::text, 'android'::text]))),
    CONSTRAINT push_tokens_provider_check CHECK ((provider = ANY (ARRAY['expo'::text, 'fcm'::text, 'apns'::text])))
);


--
-- Name: TABLE push_tokens; Type: COMMENT; Schema: messaging; Owner: -
--

COMMENT ON TABLE messaging.push_tokens IS '[messaging] Device push tokens per app; dead tokens are revoked by the receipts job.';


--
-- Name: template_versions; Type: TABLE; Schema: messaging; Owner: -
--

CREATE TABLE messaging.template_versions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    template_id uuid NOT NULL,
    version integer NOT NULL,
    subject text,
    blocks jsonb NOT NULL,
    created_by uuid NOT NULL,
    published_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT template_versions_version_check CHECK ((version > 0))
);


--
-- Name: TABLE template_versions; Type: COMMENT; Schema: messaging; Owner: -
--

COMMENT ON TABLE messaging.template_versions IS '[messaging] Immutable versions of a template; each message records the version it was rendered from.';


--
-- Name: bank_statement_lines; Type: TABLE; Schema: payments; Owner: -
--

CREATE TABLE payments.bank_statement_lines (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    bank_account text NOT NULL,
    value_date date NOT NULL,
    amount_kobo bigint NOT NULL,
    narration text,
    reference text,
    match_status text DEFAULT 'unmatched'::text NOT NULL,
    matched_type text,
    matched_id uuid,
    imported_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT bank_statement_lines_check CHECK (((match_status <> 'matched'::text) OR (matched_type IS NOT NULL))),
    CONSTRAINT bank_statement_lines_match_status_check CHECK ((match_status = ANY (ARRAY['unmatched'::text, 'matched'::text, 'exception'::text]))),
    CONSTRAINT bank_statement_lines_matched_type_check CHECK ((matched_type = ANY (ARRAY['settlement'::text, 'payout'::text, 'b2b_payment'::text, 'other'::text])))
);


--
-- Name: TABLE bank_statement_lines; Type: COMMENT; Schema: payments; Owner: -
--

COMMENT ON TABLE payments.bank_statement_lines IS '[payments] Imported bank statement lines, matched to settlements and payouts.';


--
-- Name: dispute_evidence; Type: TABLE; Schema: payments; Owner: -
--

CREATE TABLE payments.dispute_evidence (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    dispute_id uuid NOT NULL,
    kind text NOT NULL,
    file_id uuid,
    summary text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT dispute_evidence_check CHECK (((file_id IS NOT NULL) OR (summary IS NOT NULL))),
    CONSTRAINT dispute_evidence_kind_check CHECK ((kind = ANY (ARRAY['delivery_proof'::text, 'otp_record'::text, 'invoice'::text, 'device_serial'::text, 'sign_in_history'::text, 'messages'::text, 'other'::text])))
);


--
-- Name: TABLE dispute_evidence; Type: COMMENT; Schema: payments; Owner: -
--

COMMENT ON TABLE payments.dispute_evidence IS '[payments] Evidence gathered automatically for a dispute.';


--
-- Name: disputes; Type: TABLE; Schema: payments; Owner: -
--

CREATE TABLE payments.disputes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    payment_id uuid NOT NULL,
    provider_dispute_id text NOT NULL,
    provider text NOT NULL,
    reason text NOT NULL,
    amount_kobo bigint NOT NULL,
    due_by timestamp with time zone NOT NULL,
    status text DEFAULT 'open'::text NOT NULL,
    decided_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT disputes_amount_kobo_check CHECK ((amount_kobo > 0)),
    CONSTRAINT disputes_provider_check CHECK ((provider = ANY (ARRAY['paystack'::text, 'opay'::text, 'moniepoint'::text]))),
    CONSTRAINT disputes_status_check CHECK ((status = ANY (ARRAY['open'::text, 'evidence_submitted'::text, 'won'::text, 'lost'::text, 'accepted'::text])))
);


--
-- Name: TABLE disputes; Type: COMMENT; Schema: payments; Owner: -
--

COMMENT ON TABLE payments.disputes IS '[payments] Chargebacks from card schemes, with an evidence deadline.';


--
-- Name: payment_events; Type: TABLE; Schema: payments; Owner: -
--

CREATE TABLE payments.payment_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    provider text NOT NULL,
    provider_event_id text NOT NULL,
    event_type text NOT NULL,
    payment_id uuid,
    refund_id uuid,
    payout_id uuid,
    payload jsonb NOT NULL,
    signature_valid boolean NOT NULL,
    received_at timestamp with time zone DEFAULT now() NOT NULL,
    processed_at timestamp with time zone,
    CONSTRAINT payment_events_provider_check CHECK ((provider = ANY (ARRAY['paystack'::text, 'opay'::text, 'moniepoint'::text])))
);


--
-- Name: TABLE payment_events; Type: COMMENT; Schema: payments; Owner: -
--

COMMENT ON TABLE payments.payment_events IS '[payments] Raw provider webhooks, stored once per provider event so retries are ignored.';


--
-- Name: payments; Type: TABLE; Schema: payments; Owner: -
--

CREATE TABLE payments.payments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid,
    invoice_id uuid,
    provider text NOT NULL,
    channel text,
    provider_reference text,
    provider_status text,
    checkout_url text,
    idempotency_key text NOT NULL,
    amount_kobo bigint NOT NULL,
    fees_kobo bigint,
    currency character(3) DEFAULT 'NGN'::bpchar NOT NULL,
    status text DEFAULT 'initiated'::text NOT NULL,
    expires_at timestamp with time zone,
    paid_at timestamp with time zone,
    failure_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT payments_amount_kobo_check CHECK ((amount_kobo > 0)),
    CONSTRAINT payments_channel_check CHECK ((channel = ANY (ARRAY['card'::text, 'bank_transfer'::text, 'ussd'::text, 'wallet'::text, 'cash'::text, 'credit'::text]))),
    CONSTRAINT payments_check CHECK (((order_id IS NULL) <> (invoice_id IS NULL))),
    CONSTRAINT payments_check1 CHECK (((status <> 'succeeded'::text) OR (paid_at IS NOT NULL))),
    CONSTRAINT payments_fees_kobo_check CHECK ((fees_kobo >= 0)),
    CONSTRAINT payments_provider_check CHECK ((provider = ANY (ARRAY['paystack'::text, 'opay'::text, 'moniepoint'::text, 'bank_transfer'::text, 'credit'::text, 'cash'::text, 'store_credit'::text, 'pos_terminal'::text]))),
    CONSTRAINT payments_status_check CHECK ((status = ANY (ARRAY['initiated'::text, 'pending'::text, 'pending_review'::text, 'succeeded'::text, 'failed'::text, 'cancelled'::text, 'reversed'::text])))
);


--
-- Name: TABLE payments; Type: COMMENT; Schema: payments; Owner: -
--

COMMENT ON TABLE payments.payments IS '[payments] A charge for an order or invoice. Always re-verified with the provider; card data is never stored.';


--
-- Name: provider_health; Type: TABLE; Schema: payments; Owner: -
--

CREATE TABLE payments.provider_health (
    provider text NOT NULL,
    breaker_state text DEFAULT 'closed'::text NOT NULL,
    failures integer DEFAULT 0 NOT NULL,
    last_failure_at timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT provider_health_breaker_state_check CHECK ((breaker_state = ANY (ARRAY['closed'::text, 'open'::text, 'half_open'::text]))),
    CONSTRAINT provider_health_failures_check CHECK ((failures >= 0)),
    CONSTRAINT provider_health_provider_check CHECK ((provider = ANY (ARRAY['paystack'::text, 'opay'::text, 'moniepoint'::text])))
);


--
-- Name: TABLE provider_health; Type: COMMENT; Schema: payments; Owner: -
--

COMMENT ON TABLE payments.provider_health IS '[payments] Shared circuit-breaker state, so checkout hides a provider that is failing right now.';


--
-- Name: reconciliation_exceptions; Type: TABLE; Schema: payments; Owner: -
--

CREATE TABLE payments.reconciliation_exceptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    kind text NOT NULL,
    provider text,
    reference text,
    amount_kobo bigint,
    detail text NOT NULL,
    status text DEFAULT 'open'::text NOT NULL,
    owner_id uuid,
    resolution text,
    resolved_by uuid,
    resolved_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reconciliation_exceptions_check CHECK (((status = 'open'::text) OR ((resolution IS NOT NULL) AND (resolved_by IS NOT NULL) AND (resolved_at IS NOT NULL)))),
    CONSTRAINT reconciliation_exceptions_kind_check CHECK ((kind = ANY (ARRAY['missing_in_provider'::text, 'missing_in_ours'::text, 'amount_mismatch'::text, 'fee_mismatch'::text, 'duplicate'::text, 'unsettled_batch'::text]))),
    CONSTRAINT reconciliation_exceptions_status_check CHECK ((status = ANY (ARRAY['open'::text, 'resolved'::text, 'escalated'::text])))
);


--
-- Name: TABLE reconciliation_exceptions; Type: COMMENT; Schema: payments; Owner: -
--

COMMENT ON TABLE payments.reconciliation_exceptions IS '[payments] Anything that did not reconcile, with an owner and a 48-hour target.';


--
-- Name: refunds; Type: TABLE; Schema: payments; Owner: -
--

CREATE TABLE payments.refunds (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    payment_id uuid NOT NULL,
    order_id uuid NOT NULL,
    return_request_id uuid,
    amount_kobo bigint NOT NULL,
    currency character(3) DEFAULT 'NGN'::bpchar NOT NULL,
    method text DEFAULT 'original'::text NOT NULL,
    reason text NOT NULL,
    status text DEFAULT 'requested'::text NOT NULL,
    requested_by uuid,
    approved_by uuid,
    provider_reference text,
    failure_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT refunds_amount_kobo_check CHECK ((amount_kobo > 0)),
    CONSTRAINT refunds_check CHECK (((status <> ALL (ARRAY['approved'::text, 'processing'::text, 'succeeded'::text])) OR (approved_by IS NOT NULL))),
    CONSTRAINT refunds_check1 CHECK (((approved_by IS DISTINCT FROM requested_by) OR (approved_by IS NULL))),
    CONSTRAINT refunds_check2 CHECK (((status <> 'failed'::text) OR (failure_reason IS NOT NULL))),
    CONSTRAINT refunds_method_check CHECK ((method = ANY (ARRAY['original'::text, 'bank_transfer'::text, 'store_credit'::text]))),
    CONSTRAINT refunds_status_check CHECK ((status = ANY (ARRAY['requested'::text, 'approved'::text, 'processing'::text, 'succeeded'::text, 'failed'::text, 'rejected'::text])))
);


--
-- Name: TABLE refunds; Type: COMMENT; Schema: payments; Owner: -
--

COMMENT ON TABLE payments.refunds IS '[payments] Money returned to a customer. Requester and approver must be different people; requested_by is null for automatic refunds.';


--
-- Name: settlement_lines; Type: TABLE; Schema: payments; Owner: -
--

CREATE TABLE payments.settlement_lines (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    report_id uuid NOT NULL,
    provider_reference text NOT NULL,
    amount_kobo bigint NOT NULL,
    fee_kobo bigint DEFAULT 0 NOT NULL,
    payment_id uuid,
    match_status text DEFAULT 'unmatched'::text NOT NULL,
    CONSTRAINT settlement_lines_match_status_check CHECK ((match_status = ANY (ARRAY['unmatched'::text, 'matched'::text, 'exception'::text])))
);


--
-- Name: TABLE settlement_lines; Type: COMMENT; Schema: payments; Owner: -
--

COMMENT ON TABLE payments.settlement_lines IS '[payments] One line per settled transaction; matched to our payments by reference and amount.';


--
-- Name: settlement_reports; Type: TABLE; Schema: payments; Owner: -
--

CREATE TABLE payments.settlement_reports (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    provider text NOT NULL,
    report_date date NOT NULL,
    batch_ref text NOT NULL,
    gross_kobo bigint NOT NULL,
    fees_kobo bigint NOT NULL,
    net_kobo bigint NOT NULL,
    imported_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT settlement_reports_check CHECK ((net_kobo = (gross_kobo - fees_kobo))),
    CONSTRAINT settlement_reports_fees_kobo_check CHECK ((fees_kobo >= 0)),
    CONSTRAINT settlement_reports_provider_check CHECK ((provider = ANY (ARRAY['paystack'::text, 'opay'::text, 'moniepoint'::text])))
);


--
-- Name: TABLE settlement_reports; Type: COMMENT; Schema: payments; Owner: -
--

COMMENT ON TABLE payments.settlement_reports IS '[payments] Provider settlement batches imported for daily reconciliation.';


--
-- Name: virtual_accounts; Type: TABLE; Schema: payments; Owner: -
--

CREATE TABLE payments.virtual_accounts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    provider text NOT NULL,
    account_number text NOT NULL,
    bank_name text NOT NULL,
    owner_type text NOT NULL,
    order_id uuid,
    business_id uuid,
    expected_kobo bigint,
    received_kobo bigint DEFAULT 0 NOT NULL,
    expires_at timestamp with time zone,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT virtual_accounts_account_number_check CHECK ((account_number ~ '^[0-9]{10}$'::text)),
    CONSTRAINT virtual_accounts_check CHECK ((((owner_type = 'order'::text) AND (order_id IS NOT NULL) AND (business_id IS NULL) AND (expected_kobo IS NOT NULL) AND (expires_at IS NOT NULL)) OR ((owner_type = 'business'::text) AND (business_id IS NOT NULL) AND (order_id IS NULL)))),
    CONSTRAINT virtual_accounts_expected_kobo_check CHECK ((expected_kobo > 0)),
    CONSTRAINT virtual_accounts_owner_type_check CHECK ((owner_type = ANY (ARRAY['order'::text, 'business'::text]))),
    CONSTRAINT virtual_accounts_provider_check CHECK ((provider = ANY (ARRAY['paystack'::text, 'moniepoint'::text]))),
    CONSTRAINT virtual_accounts_received_kobo_check CHECK ((received_kobo >= 0)),
    CONSTRAINT virtual_accounts_status_check CHECK ((status = ANY (ARRAY['active'::text, 'paid'::text, 'expired'::text, 'closed'::text])))
);


--
-- Name: TABLE virtual_accounts; Type: COMMENT; Schema: payments; Owner: -
--

COMMENT ON TABLE payments.virtual_accounts IS '[payments] Pay-by-transfer accounts: one-time per order (exact amount, expires) or permanent per business.';


--
-- Name: identity_links; Type: TABLE; Schema: personalisation; Owner: -
--

CREATE TABLE personalisation.identity_links (
    anonymous_id uuid NOT NULL,
    user_id uuid NOT NULL,
    linked_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE identity_links; Type: COMMENT; Schema: personalisation; Owner: -
--

COMMENT ON TABLE personalisation.identity_links IS '[personalisation] Joins a guest''s anonymous history to their account at sign-in.';


--
-- Name: rec_item_neighbours; Type: TABLE; Schema: personalisation; Owner: -
--

CREATE TABLE personalisation.rec_item_neighbours (
    model_version_id uuid NOT NULL,
    kind text NOT NULL,
    product_id uuid NOT NULL,
    neighbour_product_id uuid NOT NULL,
    score real NOT NULL,
    rank smallint NOT NULL,
    CONSTRAINT rec_item_neighbours_check CHECK ((neighbour_product_id <> product_id)),
    CONSTRAINT rec_item_neighbours_kind_check CHECK ((kind = ANY (ARRAY['similar'::text, 'bought_together'::text, 'also_viewed'::text]))),
    CONSTRAINT rec_item_neighbours_rank_check CHECK ((rank > 0))
);


--
-- Name: TABLE rec_item_neighbours; Type: COMMENT; Schema: personalisation; Owner: -
--

COMMENT ON TABLE personalisation.rec_item_neighbours IS '[personalisation] Similar, bought-together and also-viewed products per product.';


--
-- Name: rec_item_vectors; Type: TABLE; Schema: personalisation; Owner: -
--

CREATE TABLE personalisation.rec_item_vectors (
    model_version_id uuid NOT NULL,
    product_id uuid NOT NULL,
    vector real[] NOT NULL,
    CONSTRAINT rec_item_vectors_vector_check CHECK (((cardinality(vector) >= 8) AND (cardinality(vector) <= 512)))
);


--
-- Name: TABLE rec_item_vectors; Type: COMMENT; Schema: personalisation; Owner: -
--

COMMENT ON TABLE personalisation.rec_item_vectors IS '[personalisation] Item embeddings used by Go for live session re-ranking.';


--
-- Name: rec_model_versions; Type: TABLE; Schema: personalisation; Owner: -
--

CREATE TABLE personalisation.rec_model_versions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    model_kind text DEFAULT 'hybrid'::text NOT NULL,
    status text DEFAULT 'staged'::text NOT NULL,
    trained_at timestamp with time zone NOT NULL,
    data_cutoff timestamp with time zone NOT NULL,
    code_version text NOT NULL,
    metrics jsonb DEFAULT '{}'::jsonb NOT NULL,
    row_counts jsonb DEFAULT '{}'::jsonb NOT NULL,
    activated_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT rec_model_versions_check CHECK (((status <> 'active'::text) OR (activated_at IS NOT NULL))),
    CONSTRAINT rec_model_versions_model_kind_check CHECK ((model_kind = ANY (ARRAY['hybrid'::text, 'popularity'::text, 'content'::text, 'cooccurrence'::text, 'als'::text, 'bpr'::text, 'item2vec'::text, 'two_tower'::text, 'sasrec'::text, 'ranker'::text]))),
    CONSTRAINT rec_model_versions_status_check CHECK ((status = ANY (ARRAY['staged'::text, 'active'::text, 'retired'::text, 'failed'::text])))
);


--
-- Name: TABLE rec_model_versions; Type: COMMENT; Schema: personalisation; Owner: -
--

COMMENT ON TABLE personalisation.rec_model_versions IS '[personalisation] Imported model versions; one active at a time, the previous kept for one-click rollback.';


--
-- Name: rec_overrides; Type: TABLE; Schema: personalisation; Owner: -
--

CREATE TABLE personalisation.rec_overrides (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    surface text NOT NULL,
    anchor_product_id uuid,
    product_id uuid NOT NULL,
    action text NOT NULL,
    starts_at timestamp with time zone DEFAULT now() NOT NULL,
    ends_at timestamp with time zone,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT rec_overrides_action_check CHECK ((action = ANY (ARRAY['pin'::text, 'block'::text]))),
    CONSTRAINT rec_overrides_check CHECK (((ends_at IS NULL) OR (ends_at > starts_at)))
);


--
-- Name: TABLE rec_overrides; Type: COMMENT; Schema: personalisation; Owner: -
--

COMMENT ON TABLE personalisation.rec_overrides IS '[personalisation] Staff merchandising pins and blocks per shelf (audited, time-boxed).';


--
-- Name: rec_popularity; Type: TABLE; Schema: personalisation; Owner: -
--

CREATE TABLE personalisation.rec_popularity (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    computed_at timestamp with time zone NOT NULL,
    "window" text NOT NULL,
    category_id uuid,
    state_code text,
    product_id uuid NOT NULL,
    score real NOT NULL,
    rank smallint NOT NULL,
    CONSTRAINT rec_popularity_rank_check CHECK ((rank > 0)),
    CONSTRAINT rec_popularity_window_check CHECK (("window" = ANY (ARRAY['7d'::text, '30d'::text])))
);


--
-- Name: TABLE rec_popularity; Type: COMMENT; Schema: personalisation; Owner: -
--

COMMENT ON TABLE personalisation.rec_popularity IS '[personalisation] Trending products by window, category and state (the last-resort fallback for every shelf).';


--
-- Name: rec_staging_item_neighbours; Type: TABLE; Schema: personalisation; Owner: -
--

CREATE TABLE personalisation.rec_staging_item_neighbours (
    model_version_id uuid NOT NULL,
    kind text NOT NULL,
    product_id uuid NOT NULL,
    neighbour_product_id uuid NOT NULL,
    score real NOT NULL,
    rank smallint NOT NULL,
    CONSTRAINT rec_item_neighbours_check CHECK ((neighbour_product_id <> product_id)),
    CONSTRAINT rec_item_neighbours_kind_check CHECK ((kind = ANY (ARRAY['similar'::text, 'bought_together'::text, 'also_viewed'::text]))),
    CONSTRAINT rec_item_neighbours_rank_check CHECK ((rank > 0))
);


--
-- Name: TABLE rec_staging_item_neighbours; Type: COMMENT; Schema: personalisation; Owner: -
--

COMMENT ON TABLE personalisation.rec_staging_item_neighbours IS '[personalisation] Staging copy of rec_item_neighbours for validating imports.';


--
-- Name: rec_staging_item_vectors; Type: TABLE; Schema: personalisation; Owner: -
--

CREATE TABLE personalisation.rec_staging_item_vectors (
    model_version_id uuid NOT NULL,
    product_id uuid NOT NULL,
    vector real[] NOT NULL,
    CONSTRAINT rec_item_vectors_vector_check CHECK (((cardinality(vector) >= 8) AND (cardinality(vector) <= 512)))
);


--
-- Name: TABLE rec_staging_item_vectors; Type: COMMENT; Schema: personalisation; Owner: -
--

COMMENT ON TABLE personalisation.rec_staging_item_vectors IS '[personalisation] Staging copy of rec_item_vectors for validating imports.';


--
-- Name: rec_staging_popularity; Type: TABLE; Schema: personalisation; Owner: -
--

CREATE TABLE personalisation.rec_staging_popularity (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    computed_at timestamp with time zone NOT NULL,
    "window" text NOT NULL,
    category_id uuid,
    state_code text,
    product_id uuid NOT NULL,
    score real NOT NULL,
    rank smallint NOT NULL,
    CONSTRAINT rec_popularity_rank_check CHECK ((rank > 0)),
    CONSTRAINT rec_popularity_window_check CHECK (("window" = ANY (ARRAY['7d'::text, '30d'::text])))
);


--
-- Name: TABLE rec_staging_popularity; Type: COMMENT; Schema: personalisation; Owner: -
--

COMMENT ON TABLE personalisation.rec_staging_popularity IS '[personalisation] Staging copy of rec_popularity for validating imports.';


--
-- Name: rec_staging_subject_candidates; Type: TABLE; Schema: personalisation; Owner: -
--

CREATE TABLE personalisation.rec_staging_subject_candidates (
    model_version_id uuid NOT NULL,
    user_id uuid NOT NULL,
    product_id uuid NOT NULL,
    score real NOT NULL,
    rank smallint NOT NULL,
    CONSTRAINT rec_subject_candidates_rank_check CHECK ((rank > 0))
);


--
-- Name: TABLE rec_staging_subject_candidates; Type: COMMENT; Schema: personalisation; Owner: -
--

COMMENT ON TABLE personalisation.rec_staging_subject_candidates IS '[personalisation] Staging copy of rec_subject_candidates for validating imports.';


--
-- Name: rec_subject_candidates; Type: TABLE; Schema: personalisation; Owner: -
--

CREATE TABLE personalisation.rec_subject_candidates (
    model_version_id uuid NOT NULL,
    user_id uuid NOT NULL,
    product_id uuid NOT NULL,
    score real NOT NULL,
    rank smallint NOT NULL,
    CONSTRAINT rec_subject_candidates_rank_check CHECK ((rank > 0))
);


--
-- Name: TABLE rec_subject_candidates; Type: COMMENT; Schema: personalisation; Owner: -
--

COMMENT ON TABLE personalisation.rec_subject_candidates IS '[personalisation] "For you" candidates per consented user.';


--
-- Name: user_events; Type: TABLE; Schema: personalisation; Owner: -
--

CREATE TABLE personalisation.user_events (
    id bigint NOT NULL,
    occurred_at timestamp with time zone NOT NULL,
    received_at timestamp with time zone DEFAULT now() NOT NULL,
    anonymous_id uuid NOT NULL,
    user_id uuid,
    session_id text,
    event_type text NOT NULL,
    app text NOT NULL,
    surface text,
    product_id uuid,
    listing_id uuid,
    car_listing_id uuid,
    category_id uuid,
    query_norm text,
    results_count integer,
    "position" smallint,
    rec_request_id text,
    strategy text,
    state_code text,
    device_class text,
    properties jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT user_events_app_check CHECK ((app = ANY (ARRAY['market'::text, 'wholesale'::text, 'customer_app'::text, 'server'::text]))),
    CONSTRAINT user_events_device_class_check CHECK ((device_class = ANY (ARRAY['mobile'::text, 'tablet'::text, 'desktop'::text]))),
    CONSTRAINT user_events_event_type_check CHECK ((event_type = ANY (ARRAY['product_view'::text, 'list_impression'::text, 'rec_impression'::text, 'rec_click'::text, 'search'::text, 'search_click'::text, 'add_to_cart'::text, 'remove_from_cart'::text, 'save'::text, 'unsave'::text, 'begin_checkout'::text, 'purchase'::text, 'car_view'::text, 'car_enquiry'::text, 'share'::text]))),
    CONSTRAINT user_events_position_check CHECK (("position" >= 0)),
    CONSTRAINT user_events_results_count_check CHECK ((results_count >= 0))
)
PARTITION BY RANGE (received_at);


--
-- Name: TABLE user_events; Type: COMMENT; Schema: personalisation; Owner: -
--

COMMENT ON TABLE personalisation.user_events IS '[personalisation] Behaviour events from consented visitors (views, clicks, carts, searches). Partitioned monthly.';


--
-- Name: user_events_default; Type: TABLE; Schema: personalisation; Owner: -
--

CREATE TABLE personalisation.user_events_default (
    id bigint NOT NULL,
    occurred_at timestamp with time zone NOT NULL,
    received_at timestamp with time zone DEFAULT now() NOT NULL,
    anonymous_id uuid NOT NULL,
    user_id uuid,
    session_id text,
    event_type text NOT NULL,
    app text NOT NULL,
    surface text,
    product_id uuid,
    listing_id uuid,
    car_listing_id uuid,
    category_id uuid,
    query_norm text,
    results_count integer,
    "position" smallint,
    rec_request_id text,
    strategy text,
    state_code text,
    device_class text,
    properties jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT user_events_app_check CHECK ((app = ANY (ARRAY['market'::text, 'wholesale'::text, 'customer_app'::text, 'server'::text]))),
    CONSTRAINT user_events_device_class_check CHECK ((device_class = ANY (ARRAY['mobile'::text, 'tablet'::text, 'desktop'::text]))),
    CONSTRAINT user_events_event_type_check CHECK ((event_type = ANY (ARRAY['product_view'::text, 'list_impression'::text, 'rec_impression'::text, 'rec_click'::text, 'search'::text, 'search_click'::text, 'add_to_cart'::text, 'remove_from_cart'::text, 'save'::text, 'unsave'::text, 'begin_checkout'::text, 'purchase'::text, 'car_view'::text, 'car_enquiry'::text, 'share'::text]))),
    CONSTRAINT user_events_position_check CHECK (("position" >= 0)),
    CONSTRAINT user_events_results_count_check CHECK ((results_count >= 0))
);


--
-- Name: user_events_id_seq; Type: SEQUENCE; Schema: personalisation; Owner: -
--

ALTER TABLE personalisation.user_events ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME personalisation.user_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: audit_log; Type: TABLE; Schema: platform; Owner: -
--

CREATE TABLE platform.audit_log (
    id bigint NOT NULL,
    actor_user_id uuid,
    actor_kind text DEFAULT 'user'::text NOT NULL,
    action text NOT NULL,
    entity_type text NOT NULL,
    entity_id text NOT NULL,
    changes jsonb,
    request_id text,
    ip inet,
    user_agent text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT audit_log_actor_kind_check CHECK ((actor_kind = ANY (ARRAY['user'::text, 'staff'::text, 'seller'::text, 'system'::text, 'service'::text])))
);


--
-- Name: TABLE audit_log; Type: COMMENT; Schema: platform; Owner: -
--

COMMENT ON TABLE platform.audit_log IS '[platform] Append-only record of who changed what (staff and seller actions, sensitive reads such as ID reveals).';


--
-- Name: audit_log_id_seq; Type: SEQUENCE; Schema: platform; Owner: -
--

ALTER TABLE platform.audit_log ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME platform.audit_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: encryption_keys; Type: TABLE; Schema: platform; Owner: -
--

CREATE TABLE platform.encryption_keys (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    purpose text NOT NULL,
    wrapped_dek bytea NOT NULL,
    kms_key_id text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    retired_at timestamp with time zone,
    CONSTRAINT encryption_keys_purpose_check CHECK ((purpose = ANY (ARRAY['kyc_id'::text, 'bank_account'::text, 'mfa_secret'::text, 'trade_in_bank'::text])))
);


--
-- Name: TABLE encryption_keys; Type: COMMENT; Schema: platform; Owner: -
--

COMMENT ON TABLE platform.encryption_keys IS '[platform] Envelope encryption: data keys wrapped by the KMS key. Only one active key per purpose.';


--
-- Name: feature_flags; Type: TABLE; Schema: platform; Owner: -
--

CREATE TABLE platform.feature_flags (
    key text NOT NULL,
    description text NOT NULL,
    enabled boolean DEFAULT false NOT NULL,
    rules jsonb,
    updated_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE feature_flags; Type: COMMENT; Schema: platform; Owner: -
--

COMMENT ON TABLE platform.feature_flags IS '[platform] Runtime feature switches (IT tools module), e.g. the live recommender service.';


--
-- Name: files; Type: TABLE; Schema: platform; Owner: -
--

CREATE TABLE platform.files (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    storage_key text NOT NULL,
    purpose text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    original_name text,
    content_type text NOT NULL,
    size_bytes bigint NOT NULL,
    checksum text,
    width integer,
    height integer,
    blurhash text,
    visibility text DEFAULT 'private'::text NOT NULL,
    uploaded_by uuid,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT files_height_check CHECK ((height > 0)),
    CONSTRAINT files_purpose_check CHECK ((purpose = ANY (ARRAY['product_image'::text, 'car_image'::text, 'kyc_document'::text, 'delivery_proof'::text, 'invoice_pdf'::text, 'receipt_pdf'::text, 'return_photo'::text, 'ticket_attachment'::text, 'privacy_export'::text, 'car_document'::text, 'inspection_report'::text, 'brand_logo'::text, 'banner'::text, 'review_image'::text, 'dispute_evidence'::text, 'other'::text]))),
    CONSTRAINT files_size_bytes_check CHECK ((size_bytes > 0)),
    CONSTRAINT files_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'ready'::text, 'rejected'::text]))),
    CONSTRAINT files_visibility_check CHECK ((visibility = ANY (ARRAY['public'::text, 'private'::text]))),
    CONSTRAINT files_width_check CHECK ((width > 0))
);


--
-- Name: TABLE files; Type: COMMENT; Schema: platform; Owner: -
--

COMMENT ON TABLE platform.files IS '[platform] Uploaded files (photos, KYC documents, delivery proofs, PDFs). Bytes live in object storage via presigned uploads.';


--
-- Name: idempotency_keys; Type: TABLE; Schema: platform; Owner: -
--

CREATE TABLE platform.idempotency_keys (
    scope text NOT NULL,
    key text NOT NULL,
    user_id uuid,
    method text NOT NULL,
    path text NOT NULL,
    request_hash bytea NOT NULL,
    response_code smallint,
    response_body jsonb,
    locked_at timestamp with time zone,
    completed_at timestamp with time zone,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE idempotency_keys; Type: COMMENT; Schema: platform; Owner: -
--

COMMENT ON TABLE platform.idempotency_keys IS '[platform] Remembers responses to retried requests (checkout, payments, refunds) so they run once; locked_at detects in-flight duplicates.';


--
-- Name: outbox_events; Type: TABLE; Schema: platform; Owner: -
--

CREATE TABLE platform.outbox_events (
    id bigint NOT NULL,
    aggregate_type text NOT NULL,
    aggregate_id uuid NOT NULL,
    event_type text NOT NULL,
    payload jsonb NOT NULL,
    trace_context jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    published_at timestamp with time zone
);


--
-- Name: TABLE outbox_events; Type: COMMENT; Schema: platform; Owner: -
--

COMMENT ON TABLE platform.outbox_events IS '[platform] Transactional outbox: events written in the same transaction as the change, then relayed to jobs.';


--
-- Name: outbox_events_id_seq; Type: SEQUENCE; Schema: platform; Owner: -
--

ALTER TABLE platform.outbox_events ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME platform.outbox_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: public_holidays; Type: TABLE; Schema: platform; Owner: -
--

CREATE TABLE platform.public_holidays (
    date date NOT NULL,
    name text NOT NULL
);


--
-- Name: TABLE public_holidays; Type: COMMENT; Schema: platform; Owner: -
--

COMMENT ON TABLE platform.public_holidays IS '[platform] Nigerian public holidays, so payout runs skip non-business days.';


--
-- Name: pos_shifts; Type: TABLE; Schema: pos; Owner: -
--

CREATE TABLE pos.pos_shifts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    terminal_id uuid NOT NULL,
    cashier_id uuid NOT NULL,
    opened_at timestamp with time zone DEFAULT now() NOT NULL,
    closed_at timestamp with time zone,
    opening_float_kobo bigint DEFAULT 0 NOT NULL,
    expected_cash_kobo bigint,
    counted_cash_kobo bigint,
    variance_kobo bigint GENERATED ALWAYS AS ((counted_cash_kobo - expected_cash_kobo)) STORED,
    CONSTRAINT pos_shifts_check CHECK (((closed_at IS NULL) OR (closed_at > opened_at))),
    CONSTRAINT pos_shifts_check1 CHECK (((closed_at IS NULL) OR ((expected_cash_kobo IS NOT NULL) AND (counted_cash_kobo IS NOT NULL)))),
    CONSTRAINT pos_shifts_opening_float_kobo_check CHECK ((opening_float_kobo >= 0))
);


--
-- Name: TABLE pos_shifts; Type: COMMENT; Schema: pos; Owner: -
--

COMMENT ON TABLE pos.pos_shifts IS '[pos] A cashier''s shift on a till, with cash reconciliation (variance flagged).';


--
-- Name: pos_terminals; Type: TABLE; Schema: pos; Owner: -
--

CREATE TABLE pos.pos_terminals (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    warehouse_id uuid NOT NULL,
    label text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT pos_terminals_status_check CHECK ((status = ANY (ARRAY['active'::text, 'retired'::text])))
);


--
-- Name: TABLE pos_terminals; Type: COMMENT; Schema: pos; Owner: -
--

COMMENT ON TABLE pos.pos_terminals IS '[pos] Tills in physical stores (warehouses of kind store).';


--
-- Name: goods_receipts; Type: TABLE; Schema: purchasing; Owner: -
--

CREATE TABLE purchasing.goods_receipts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    purchase_order_id uuid NOT NULL,
    received_by uuid NOT NULL,
    received_at timestamp with time zone DEFAULT now() NOT NULL,
    notes text
);


--
-- Name: TABLE goods_receipts; Type: COMMENT; Schema: purchasing; Owner: -
--

COMMENT ON TABLE purchasing.goods_receipts IS '[purchasing] A delivery received against a PO; posts stock movements (purchase_receipt) and device units.';


--
-- Name: po_number_seq; Type: SEQUENCE; Schema: purchasing; Owner: -
--

CREATE SEQUENCE purchasing.po_number_seq
    START WITH 2001
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: purchase_order_lines; Type: TABLE; Schema: purchasing; Owner: -
--

CREATE TABLE purchasing.purchase_order_lines (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    purchase_order_id uuid NOT NULL,
    variant_id uuid NOT NULL,
    condition text DEFAULT 'new'::text NOT NULL,
    quantity_ordered integer NOT NULL,
    quantity_received integer DEFAULT 0 NOT NULL,
    unit_cost_kobo bigint NOT NULL,
    CONSTRAINT purchase_order_lines_check CHECK (((quantity_received)::numeric <= ((quantity_ordered)::numeric * 1.05))),
    CONSTRAINT purchase_order_lines_condition_check CHECK ((condition = ANY (ARRAY['new'::text, 'uk_used'::text, 'refurbished'::text, 'open_box'::text]))),
    CONSTRAINT purchase_order_lines_quantity_ordered_check CHECK ((quantity_ordered > 0)),
    CONSTRAINT purchase_order_lines_quantity_received_check CHECK ((quantity_received >= 0)),
    CONSTRAINT purchase_order_lines_unit_cost_kobo_check CHECK ((unit_cost_kobo > 0))
);


--
-- Name: TABLE purchase_order_lines; Type: COMMENT; Schema: purchasing; Owner: -
--

COMMENT ON TABLE purchasing.purchase_order_lines IS '[purchasing] Items, quantities and costs on a purchase order (over-receipt above 5% needs approval).';


--
-- Name: purchase_orders; Type: TABLE; Schema: purchasing; Owner: -
--

CREATE TABLE purchasing.purchase_orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    po_number text DEFAULT platform.next_ref('PO-'::text, 'purchasing.po_number_seq'::regclass) NOT NULL,
    supplier_id uuid NOT NULL,
    warehouse_id uuid NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    expected_on date,
    total_kobo bigint DEFAULT 0 NOT NULL,
    currency character(3) DEFAULT 'NGN'::bpchar NOT NULL,
    created_by uuid NOT NULL,
    approved_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT purchase_orders_check CHECK ((approved_by IS DISTINCT FROM created_by)),
    CONSTRAINT purchase_orders_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'awaiting_approval'::text, 'sent'::text, 'partially_received'::text, 'received'::text, 'cancelled'::text]))),
    CONSTRAINT purchase_orders_total_kobo_check CHECK ((total_kobo >= 0))
);


--
-- Name: TABLE purchase_orders; Type: COMMENT; Schema: purchasing; Owner: -
--

COMMENT ON TABLE purchasing.purchase_orders IS '[purchasing] Orders placed with suppliers, delivered to a warehouse. Large POs need finance approval.';


--
-- Name: suppliers; Type: TABLE; Schema: purchasing; Owner: -
--

CREATE TABLE purchasing.suppliers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    rc_number text,
    contact_name text,
    email public.citext,
    phone text,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT suppliers_status_check CHECK ((status = ANY (ARRAY['active'::text, 'on_hold'::text, 'inactive'::text])))
);


--
-- Name: TABLE suppliers; Type: COMMENT; Schema: purchasing; Owner: -
--

COMMENT ON TABLE purchasing.suppliers IS '[purchasing] Companies TechShop buys stock from.';


--
-- Name: device_blocklist; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.device_blocklist (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    imei text NOT NULL,
    reason text NOT NULL,
    source text NOT NULL,
    reference text,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT device_blocklist_imei_check CHECK ((imei ~ '^[0-9]{15}$'::text)),
    CONSTRAINT device_blocklist_reason_check CHECK ((reason = ANY (ARRAY['reported_stolen'::text, 'fraud'::text, 'counterfeit'::text])))
);


--
-- Name: TABLE device_blocklist; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.device_blocklist IS '[risk] IMEIs that may not be sold, traded in, returned or repaired (stolen, fraud, counterfeit).';


--
-- Name: enforcement_actions; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.enforcement_actions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    seller_id uuid NOT NULL,
    step text NOT NULL,
    reason text NOT NULL,
    created_by uuid NOT NULL,
    appeal_status text,
    appealed_at timestamp with time zone,
    decided_by uuid,
    decided_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT enforcement_actions_appeal_status_check CHECK ((appeal_status = ANY (ARRAY['submitted'::text, 'upheld'::text, 'rejected'::text]))),
    CONSTRAINT enforcement_actions_check CHECK (((decided_by IS DISTINCT FROM created_by) OR (decided_by IS NULL))),
    CONSTRAINT enforcement_actions_check1 CHECK (((appeal_status IS NULL) OR (appealed_at IS NOT NULL))),
    CONSTRAINT enforcement_actions_check2 CHECK (((appeal_status <> ALL (ARRAY['upheld'::text, 'rejected'::text])) OR ((decided_by IS NOT NULL) AND (decided_at IS NOT NULL)))),
    CONSTRAINT enforcement_actions_step_check CHECK ((step = ANY (ARRAY['warning'::text, 'ranking_demotion'::text, 'listing_review'::text, 'payouts_held'::text, 'suspended'::text, 'reinstated'::text])))
);


--
-- Name: TABLE enforcement_actions; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.enforcement_actions IS '[risk] Seller enforcement ladder steps, each with reasons and an appeal decided by a different person.';


--
-- Name: kyc_checks; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.kyc_checks (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    subject_type text NOT NULL,
    subject_id uuid NOT NULL,
    kind text NOT NULL,
    provider text NOT NULL,
    result text NOT NULL,
    score numeric(4,3),
    raw_ref text,
    checked_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT kyc_checks_kind_check CHECK ((kind = ANY (ARRAY['nin'::text, 'selfie'::text, 'cac'::text, 'bank_name'::text, 'director'::text, 'guarantor'::text, 'licence'::text]))),
    CONSTRAINT kyc_checks_result_check CHECK ((result = ANY (ARRAY['match'::text, 'partial'::text, 'no_match'::text, 'error'::text]))),
    CONSTRAINT kyc_checks_score_check CHECK (((score >= (0)::numeric) AND (score <= (1)::numeric))),
    CONSTRAINT kyc_checks_subject_type_check CHECK ((subject_type = ANY (ARRAY['seller'::text, 'business'::text, 'rider'::text])))
);


--
-- Name: TABLE kyc_checks; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.kyc_checks IS '[risk] Results of each automated KYC check (provider or manual), shown to the reviewer.';


--
-- Name: listing_reports; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.listing_reports (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    listing_id uuid NOT NULL,
    reporter_id uuid,
    source text NOT NULL,
    reason text NOT NULL,
    details text,
    status text DEFAULT 'open'::text NOT NULL,
    decided_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT listing_reports_check CHECK (((status = 'open'::text) OR (decided_by IS NOT NULL))),
    CONSTRAINT listing_reports_reason_check CHECK ((reason = ANY (ARRAY['counterfeit'::text, 'wrong_info'::text, 'prohibited'::text, 'scam'::text, 'other'::text]))),
    CONSTRAINT listing_reports_source_check CHECK ((source = ANY (ARRAY['buyer'::text, 'brand_owner'::text, 'staff'::text, 'mystery_shop'::text]))),
    CONSTRAINT listing_reports_status_check CHECK ((status = ANY (ARRAY['open'::text, 'upheld'::text, 'dismissed'::text])))
);


--
-- Name: TABLE listing_reports; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.listing_reports IS '[risk] Reports of counterfeit or misleading listings; upheld reports are seller strikes.';


--
-- Name: review_reports; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.review_reports (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    review_id uuid NOT NULL,
    reporter_id uuid,
    reason text NOT NULL,
    status text DEFAULT 'open'::text NOT NULL,
    decided_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT review_reports_check CHECK (((status = 'open'::text) OR (decided_by IS NOT NULL))),
    CONSTRAINT review_reports_reason_check CHECK ((reason = ANY (ARRAY['abuse'::text, 'personal_data'::text, 'off_topic'::text, 'fake'::text, 'other'::text]))),
    CONSTRAINT review_reports_status_check CHECK ((status = ANY (ARRAY['open'::text, 'upheld'::text, 'dismissed'::text])))
);


--
-- Name: TABLE review_reports; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.review_reports IS '[risk] Reports of reviews that break policy (sellers can report, never delete).';


--
-- Name: risk_cases; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.risk_cases (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    kind text NOT NULL,
    subject_type text NOT NULL,
    subject_id uuid NOT NULL,
    reason text NOT NULL,
    score smallint,
    money_at_risk_kobo bigint,
    decision_id bigint,
    status text DEFAULT 'open'::text NOT NULL,
    label text,
    sla_due_at timestamp with time zone,
    assigned_to uuid,
    resolved_by uuid,
    resolved_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT risk_cases_check CHECK (((status = ANY (ARRAY['open'::text, 'in_review'::text, 'escalated'::text])) OR ((resolved_by IS NOT NULL) AND (resolved_at IS NOT NULL) AND (label IS NOT NULL)))),
    CONSTRAINT risk_cases_kind_check CHECK ((kind = ANY (ARRAY['order'::text, 'payment'::text, 'device'::text, 'seller'::text, 'account'::text, 'trade_in'::text, 'return'::text, 'review'::text]))),
    CONSTRAINT risk_cases_label_check CHECK ((label = ANY (ARRAY['fraud'::text, 'legit'::text, 'unclear'::text]))),
    CONSTRAINT risk_cases_money_at_risk_kobo_check CHECK ((money_at_risk_kobo >= 0)),
    CONSTRAINT risk_cases_score_check CHECK (((score >= 0) AND (score <= 100))),
    CONSTRAINT risk_cases_status_check CHECK ((status = ANY (ARRAY['open'::text, 'in_review'::text, 'cleared'::text, 'blocked'::text, 'escalated'::text])))
);


--
-- Name: TABLE risk_cases; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.risk_cases IS '[risk] Fraud and risk reviews. Analysts'' clear/block decisions become labels for rule tuning.';


--
-- Name: risk_decisions; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.risk_decisions (
    id bigint NOT NULL,
    checkpoint text NOT NULL,
    subject_type text NOT NULL,
    subject_id uuid NOT NULL,
    user_id uuid,
    score smallint NOT NULL,
    outcome text NOT NULL,
    reasons jsonb DEFAULT '[]'::jsonb NOT NULL,
    rule_set_version integer,
    latency_ms integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT risk_decisions_latency_ms_check CHECK ((latency_ms >= 0)),
    CONSTRAINT risk_decisions_outcome_check CHECK ((outcome = ANY (ARRAY['allow'::text, 'step_up'::text, 'hold'::text, 'block'::text, 'fallback_allow'::text, 'fallback_hold'::text]))),
    CONSTRAINT risk_decisions_score_check CHECK (((score >= 0) AND (score <= 100)))
)
PARTITION BY RANGE (created_at);


--
-- Name: TABLE risk_decisions; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.risk_decisions IS '[risk] Every risk decision with score, reasons and rule-set version. Partitioned monthly; 24-month retention.';


--
-- Name: risk_decisions_default; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.risk_decisions_default (
    id bigint NOT NULL,
    checkpoint text NOT NULL,
    subject_type text NOT NULL,
    subject_id uuid NOT NULL,
    user_id uuid,
    score smallint NOT NULL,
    outcome text NOT NULL,
    reasons jsonb DEFAULT '[]'::jsonb NOT NULL,
    rule_set_version integer,
    latency_ms integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT risk_decisions_latency_ms_check CHECK ((latency_ms >= 0)),
    CONSTRAINT risk_decisions_outcome_check CHECK ((outcome = ANY (ARRAY['allow'::text, 'step_up'::text, 'hold'::text, 'block'::text, 'fallback_allow'::text, 'fallback_hold'::text]))),
    CONSTRAINT risk_decisions_score_check CHECK (((score >= 0) AND (score <= 100)))
);


--
-- Name: risk_decisions_id_seq; Type: SEQUENCE; Schema: risk; Owner: -
--

ALTER TABLE risk.risk_decisions ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME risk.risk_decisions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: risk_entities; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.risk_entities (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    kind text NOT NULL,
    value_hash bytea NOT NULL,
    first_seen timestamp with time zone DEFAULT now() NOT NULL,
    last_seen timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT risk_entities_kind_check CHECK ((kind = ANY (ARRAY['account'::text, 'device'::text, 'phone'::text, 'email'::text, 'address'::text, 'card_fp'::text, 'bank_hmac'::text, 'imei'::text, 'ip'::text])))
);


--
-- Name: TABLE risk_entities; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.risk_entities IS '[risk] Nodes of the link graph (accounts, devices, phones, cards…), stored as hashes.';


--
-- Name: risk_links; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.risk_links (
    entity_a uuid NOT NULL,
    entity_b uuid NOT NULL,
    first_seen timestamp with time zone DEFAULT now() NOT NULL,
    last_seen timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT risk_links_check CHECK ((entity_a < entity_b))
);


--
-- Name: TABLE risk_links; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.risk_links IS '[risk] Edges of the link graph: two entities seen together (stored once, a < b).';


--
-- Name: risk_lists; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.risk_lists (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    kind text NOT NULL,
    value_hash bytea NOT NULL,
    reason text NOT NULL,
    expires_at timestamp with time zone,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT risk_lists_kind_check CHECK ((kind = ANY (ARRAY['phone'::text, 'email'::text, 'bank'::text, 'address'::text, 'email_domain'::text, 'device'::text])))
);


--
-- Name: TABLE risk_lists; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.risk_lists IS '[risk] Ban lists (phones, emails, bank accounts, addresses, disposable email domains), hashed.';


--
-- Name: risk_rule_sets; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.risk_rule_sets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    checkpoint text NOT NULL,
    version integer NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    thresholds jsonb DEFAULT '{"hold": 70, "step": 40, "block": 90}'::jsonb NOT NULL,
    published_by uuid,
    published_at timestamp with time zone,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT risk_rule_sets_check CHECK (((status = 'draft'::text) OR ((published_by IS NOT NULL) AND (published_at IS NOT NULL)))),
    CONSTRAINT risk_rule_sets_checkpoint_check CHECK ((checkpoint = ANY (ARRAY['signup'::text, 'signin'::text, 'checkout'::text, 'payout'::text, 'bank_change'::text, 'trade_in'::text, 'return'::text, 'review'::text]))),
    CONSTRAINT risk_rule_sets_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'active'::text, 'retired'::text]))),
    CONSTRAINT risk_rule_sets_version_check CHECK ((version > 0))
);


--
-- Name: TABLE risk_rule_sets; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.risk_rule_sets IS '[risk] Versioned rule sets per checkpoint; only one active per checkpoint. Publishing needs a step-up.';


--
-- Name: risk_rules; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.risk_rules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    rule_set_id uuid NOT NULL,
    key text NOT NULL,
    label text NOT NULL,
    condition jsonb NOT NULL,
    weight integer DEFAULT 0 NOT NULL,
    action text DEFAULT 'score'::text NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    CONSTRAINT risk_rules_action_check CHECK ((action = ANY (ARRAY['score'::text, 'hold'::text, 'block'::text]))),
    CONSTRAINT risk_rules_key_check CHECK ((key ~ '^[a-z][a-z0-9_]*$'::text)),
    CONSTRAINT risk_rules_weight_check CHECK (((weight >= '-100'::integer) AND (weight <= 100)))
);


--
-- Name: TABLE risk_rules; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.risk_rules IS '[risk] Rules in a rule set: a condition over features, a weight, or a deterministic hold/block.';


--
-- Name: seller_metrics_daily; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.seller_metrics_daily (
    seller_id uuid NOT NULL,
    day date NOT NULL,
    orders integer DEFAULT 0 NOT NULL,
    defects integer DEFAULT 0 NOT NULL,
    late_shipments integer DEFAULT 0 NOT NULL,
    seller_cancels integer DEFAULT 0 NOT NULL,
    returns integer DEFAULT 0 NOT NULL,
    counterfeit_strikes integer DEFAULT 0 NOT NULL,
    CONSTRAINT seller_metrics_daily_counterfeit_strikes_check CHECK ((counterfeit_strikes >= 0)),
    CONSTRAINT seller_metrics_daily_defects_check CHECK ((defects >= 0)),
    CONSTRAINT seller_metrics_daily_late_shipments_check CHECK ((late_shipments >= 0)),
    CONSTRAINT seller_metrics_daily_orders_check CHECK ((orders >= 0)),
    CONSTRAINT seller_metrics_daily_returns_check CHECK ((returns >= 0)),
    CONSTRAINT seller_metrics_daily_seller_cancels_check CHECK ((seller_cancels >= 0))
);


--
-- Name: TABLE seller_metrics_daily; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.seller_metrics_daily IS '[risk] Per-seller daily counts used for the rolling 60-day performance rates.';


--
-- Name: velocity_counters; Type: TABLE; Schema: risk; Owner: -
--

CREATE TABLE risk.velocity_counters (
    key text NOT NULL,
    window_start timestamp with time zone NOT NULL,
    count integer DEFAULT 0 NOT NULL,
    CONSTRAINT velocity_counters_count_check CHECK ((count >= 0))
);


--
-- Name: TABLE velocity_counters; Type: COMMENT; Schema: risk; Owner: -
--

COMMENT ON TABLE risk.velocity_counters IS '[risk] Window counters for velocity features (orders per device, cards per account…). Same design as identity.auth_rate_limits.';


--
-- Name: cart_items; Type: TABLE; Schema: sales; Owner: -
--

CREATE TABLE sales.cart_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    cart_id uuid NOT NULL,
    listing_id uuid NOT NULL,
    quantity integer NOT NULL,
    added_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT cart_items_quantity_check CHECK ((quantity > 0))
);


--
-- Name: TABLE cart_items; Type: COMMENT; Schema: sales; Owner: -
--

COMMENT ON TABLE sales.cart_items IS '[sales] Listings in a cart. Prices are read live; they are snapshotted only on order lines.';


--
-- Name: carts; Type: TABLE; Schema: sales; Owner: -
--

CREATE TABLE sales.carts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    session_token_hash bytea,
    business_id uuid,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT carts_check CHECK (((user_id IS NOT NULL) OR (session_token_hash IS NOT NULL))),
    CONSTRAINT carts_status_check CHECK ((status = ANY (ARRAY['active'::text, 'converted'::text, 'merged'::text, 'abandoned'::text])))
);


--
-- Name: TABLE carts; Type: COMMENT; Schema: sales; Owner: -
--

COMMENT ON TABLE sales.carts IS '[sales] Shopping carts for signed-in users or guests (by session token hash).';


--
-- Name: fulfilments; Type: TABLE; Schema: sales; Owner: -
--

CREATE TABLE sales.fulfilments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    seller_id uuid NOT NULL,
    fulfilled_by text NOT NULL,
    method text DEFAULT 'delivery'::text NOT NULL,
    warehouse_id uuid,
    collection_store_id uuid,
    status text DEFAULT 'pending'::text NOT NULL,
    delivery_fee_kobo bigint DEFAULT 0 NOT NULL,
    accept_by timestamp with time zone,
    pack_by timestamp with time zone,
    delivered_at timestamp with time zone,
    cancelled_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT fulfilments_check CHECK (((status <> 'delivered'::text) OR (delivered_at IS NOT NULL))),
    CONSTRAINT fulfilments_check1 CHECK (((method <> 'collection'::text) OR (collection_store_id IS NOT NULL))),
    CONSTRAINT fulfilments_check2 CHECK (((status <> 'cancelled'::text) OR (cancelled_reason IS NOT NULL))),
    CONSTRAINT fulfilments_delivery_fee_kobo_check CHECK ((delivery_fee_kobo >= 0)),
    CONSTRAINT fulfilments_fulfilled_by_check CHECK ((fulfilled_by = ANY (ARRAY['techshop'::text, 'seller'::text]))),
    CONSTRAINT fulfilments_method_check CHECK ((method = ANY (ARRAY['delivery'::text, 'carrier'::text, 'collection'::text]))),
    CONSTRAINT fulfilments_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'accepted'::text, 'picking'::text, 'packed'::text, 'handed_over'::text, 'in_transit'::text, 'delivered'::text, 'failed'::text, 'returned'::text, 'cancelled'::text])))
);


--
-- Name: TABLE fulfilments; Type: COMMENT; Schema: sales; Owner: -
--

COMMENT ON TABLE sales.fulfilments IS '[sales] The part of an order handled by one seller. Moves on its own; drives delivery, earnings and payouts.';


--
-- Name: order_lines; Type: TABLE; Schema: sales; Owner: -
--

CREATE TABLE sales.order_lines (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    fulfilment_id uuid NOT NULL,
    listing_id uuid NOT NULL,
    seller_id uuid NOT NULL,
    variant_id uuid NOT NULL,
    product_name text NOT NULL,
    variant_name text NOT NULL,
    sku text NOT NULL,
    condition text NOT NULL,
    quantity integer NOT NULL,
    unit_price_kobo bigint NOT NULL,
    discount_kobo bigint DEFAULT 0 NOT NULL,
    line_total_kobo bigint NOT NULL,
    vat_kobo bigint DEFAULT 0 NOT NULL,
    commission_bps integer NOT NULL,
    device_unit_id uuid,
    cancelled_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT order_lines_check CHECK ((line_total_kobo = ((unit_price_kobo * quantity) - discount_kobo))),
    CONSTRAINT order_lines_check1 CHECK ((vat_kobo <= line_total_kobo)),
    CONSTRAINT order_lines_check2 CHECK (((device_unit_id IS NULL) OR (quantity = 1))),
    CONSTRAINT order_lines_commission_bps_check CHECK (((commission_bps >= 0) AND (commission_bps <= 10000))),
    CONSTRAINT order_lines_condition_check CHECK ((condition = ANY (ARRAY['new'::text, 'uk_used'::text, 'refurbished'::text, 'open_box'::text]))),
    CONSTRAINT order_lines_discount_kobo_check CHECK ((discount_kobo >= 0)),
    CONSTRAINT order_lines_line_total_kobo_check CHECK ((line_total_kobo >= 0)),
    CONSTRAINT order_lines_quantity_check CHECK ((quantity > 0)),
    CONSTRAINT order_lines_unit_price_kobo_check CHECK ((unit_price_kobo > 0)),
    CONSTRAINT order_lines_vat_kobo_check CHECK ((vat_kobo >= 0))
);


--
-- Name: TABLE order_lines; Type: COMMENT; Schema: sales; Owner: -
--

COMMENT ON TABLE sales.order_lines IS '[sales] Snapshot of what was bought (name, price, VAT, condition, commission), optionally the exact device unit.';


--
-- Name: order_number_seq; Type: SEQUENCE; Schema: sales; Owner: -
--

CREATE SEQUENCE sales.order_number_seq
    START WITH 10001
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: order_status_history; Type: TABLE; Schema: sales; Owner: -
--

CREATE TABLE sales.order_status_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    fulfilment_id uuid,
    from_status text,
    to_status text NOT NULL,
    actor_user_id uuid,
    note text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE order_status_history; Type: COMMENT; Schema: sales; Owner: -
--

COMMENT ON TABLE sales.order_status_history IS '[sales] Every order and fulfilment status change and who made it (shown on tracking).';


--
-- Name: orders; Type: TABLE; Schema: sales; Owner: -
--

CREATE TABLE sales.orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_number text DEFAULT platform.next_ref('TS-'::text, 'sales.order_number_seq'::regclass) NOT NULL,
    channel text NOT NULL,
    customer_user_id uuid,
    business_id uuid,
    quote_id uuid,
    pos_shift_id uuid,
    status text DEFAULT 'pending_payment'::text NOT NULL,
    subtotal_kobo bigint NOT NULL,
    delivery_fee_kobo bigint DEFAULT 0 NOT NULL,
    discount_kobo bigint DEFAULT 0 NOT NULL,
    vat_kobo bigint DEFAULT 0 NOT NULL,
    total_kobo bigint NOT NULL,
    currency character(3) DEFAULT 'NGN'::bpchar NOT NULL,
    ship_to jsonb,
    contact_phone text,
    price_hash text,
    payment_due_at timestamp with time zone,
    attributed_notification_id uuid,
    placed_at timestamp with time zone DEFAULT now() NOT NULL,
    cancelled_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT orders_channel_check CHECK ((channel = ANY (ARRAY['market'::text, 'wholesale'::text, 'app'::text, 'pos'::text]))),
    CONSTRAINT orders_check CHECK ((total_kobo = (((subtotal_kobo + delivery_fee_kobo) - discount_kobo) + vat_kobo))),
    CONSTRAINT orders_check1 CHECK (((channel <> 'pos'::text) OR (pos_shift_id IS NOT NULL))),
    CONSTRAINT orders_check2 CHECK (((channel = 'pos'::text) OR (ship_to IS NOT NULL))),
    CONSTRAINT orders_check3 CHECK (((status <> 'cancelled'::text) OR (cancelled_at IS NOT NULL))),
    CONSTRAINT orders_delivery_fee_kobo_check CHECK ((delivery_fee_kobo >= 0)),
    CONSTRAINT orders_discount_kobo_check CHECK ((discount_kobo >= 0)),
    CONSTRAINT orders_status_check CHECK ((status = ANY (ARRAY['pending_payment'::text, 'paid'::text, 'processing'::text, 'partially_shipped'::text, 'shipped'::text, 'delivered'::text, 'cancelled'::text, 'refunded'::text, 'partially_refunded'::text]))),
    CONSTRAINT orders_subtotal_kobo_check CHECK ((subtotal_kobo >= 0)),
    CONSTRAINT orders_total_kobo_check CHECK ((total_kobo >= 0)),
    CONSTRAINT orders_vat_kobo_check CHECK ((vat_kobo >= 0))
);


--
-- Name: TABLE orders; Type: COMMENT; Schema: sales; Owner: -
--

COMMENT ON TABLE sales.orders IS '[sales] One checkout. Totals must add up; ship_to snapshots the address; status is derived from fulfilments.';


--
-- Name: catalogue_gaps; Type: TABLE; Schema: search; Owner: -
--

CREATE TABLE search.catalogue_gaps (
    query_norm text NOT NULL,
    searches_30d integer DEFAULT 0 NOT NULL,
    status text DEFAULT 'new'::text NOT NULL,
    owner_id uuid,
    note text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT catalogue_gaps_searches_30d_check CHECK ((searches_30d >= 0)),
    CONSTRAINT catalogue_gaps_status_check CHECK ((status = ANY (ARRAY['new'::text, 'sourcing'::text, 'wont_stock'::text, 'resolved'::text])))
);


--
-- Name: TABLE catalogue_gaps; Type: COMMENT; Schema: search; Owner: -
--

COMMENT ON TABLE search.catalogue_gaps IS '[search] Things customers search for that TechShop doesn''t sell yet; demand data for purchasing.';


--
-- Name: search_pins; Type: TABLE; Schema: search; Owner: -
--

CREATE TABLE search.search_pins (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    query_norm text NOT NULL,
    product_id uuid NOT NULL,
    action text NOT NULL,
    "position" smallint,
    starts_at timestamp with time zone DEFAULT now() NOT NULL,
    ends_at timestamp with time zone NOT NULL,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT search_pins_action_check CHECK ((action = ANY (ARRAY['pin'::text, 'bury'::text]))),
    CONSTRAINT search_pins_check CHECK ((ends_at > starts_at)),
    CONSTRAINT search_pins_check1 CHECK (((action = 'bury'::text) OR ("position" IS NOT NULL))),
    CONSTRAINT search_pins_position_check CHECK ((("position" >= 1) AND ("position" <= 50)))
);


--
-- Name: TABLE search_pins; Type: COMMENT; Schema: search; Owner: -
--

COMMENT ON TABLE search.search_pins IS '[search] Time-boxed, audited pins or burials of a product for a query.';


--
-- Name: search_queries; Type: TABLE; Schema: search; Owner: -
--

CREATE TABLE search.search_queries (
    id bigint NOT NULL,
    session_id text,
    user_id uuid,
    query_norm text NOT NULL,
    filters jsonb DEFAULT '{}'::jsonb NOT NULL,
    results_count integer NOT NULL,
    corrected_to text,
    latency_ms integer,
    channel text DEFAULT 'retail'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT search_queries_channel_check CHECK ((channel = ANY (ARRAY['retail'::text, 'wholesale'::text]))),
    CONSTRAINT search_queries_latency_ms_check CHECK ((latency_ms >= 0)),
    CONSTRAINT search_queries_results_count_check CHECK ((results_count >= 0))
)
PARTITION BY RANGE (created_at);


--
-- Name: TABLE search_queries; Type: COMMENT; Schema: search; Owner: -
--

COMMENT ON TABLE search.search_queries IS '[search] Every search with result count and latency (zero-result queue, insights). Partitioned monthly; 13-month retention.';


--
-- Name: search_queries_default; Type: TABLE; Schema: search; Owner: -
--

CREATE TABLE search.search_queries_default (
    id bigint NOT NULL,
    session_id text,
    user_id uuid,
    query_norm text NOT NULL,
    filters jsonb DEFAULT '{}'::jsonb NOT NULL,
    results_count integer NOT NULL,
    corrected_to text,
    latency_ms integer,
    channel text DEFAULT 'retail'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT search_queries_channel_check CHECK ((channel = ANY (ARRAY['retail'::text, 'wholesale'::text]))),
    CONSTRAINT search_queries_latency_ms_check CHECK ((latency_ms >= 0)),
    CONSTRAINT search_queries_results_count_check CHECK ((results_count >= 0))
);


--
-- Name: search_queries_id_seq; Type: SEQUENCE; Schema: search; Owner: -
--

ALTER TABLE search.search_queries ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME search.search_queries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: search_redirects; Type: TABLE; Schema: search; Owner: -
--

CREATE TABLE search.search_redirects (
    query_norm text NOT NULL,
    url text NOT NULL,
    created_by uuid,
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT search_redirects_url_check CHECK ((url ~ '^/'::text))
);


--
-- Name: TABLE search_redirects; Type: COMMENT; Schema: search; Owner: -
--

COMMENT ON TABLE search.search_redirects IS '[search] Queries that go straight to a page (e.g. "cars" → /c/cars). Internal paths only.';


--
-- Name: search_suggestions; Type: TABLE; Schema: search; Owner: -
--

CREATE TABLE search.search_suggestions (
    prefix text NOT NULL,
    suggestion text NOT NULL,
    weight real DEFAULT 0 NOT NULL
);


--
-- Name: TABLE search_suggestions; Type: COMMENT; Schema: search; Owner: -
--

COMMENT ON TABLE search.search_suggestions IS '[search] Autocomplete entries built nightly from popular queries.';


--
-- Name: search_synonyms; Type: TABLE; Schema: search; Owner: -
--

CREATE TABLE search.search_synonyms (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    terms text[] NOT NULL,
    target text NOT NULL,
    kind text NOT NULL,
    active boolean DEFAULT true NOT NULL,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT search_synonyms_kind_check CHECK ((kind = ANY (ARRAY['one_way'::text, 'two_way'::text]))),
    CONSTRAINT search_synonyms_terms_check CHECK ((cardinality(terms) >= 1))
);


--
-- Name: TABLE search_synonyms; Type: COMMENT; Schema: search; Owner: -
--

COMMENT ON TABLE search.search_synonyms IS '[search] Query rewrites, e.g. tokunbo → UK-used, ps5 → Play 5. Reloaded instantly via NOTIFY.';


--
-- Name: kyc_documents; Type: TABLE; Schema: sellers; Owner: -
--

CREATE TABLE sellers.kyc_documents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    submission_id uuid NOT NULL,
    kind text NOT NULL,
    file_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT kyc_documents_kind_check CHECK ((kind = ANY (ARRAY['id_front'::text, 'id_back'::text, 'cac_certificate'::text, 'utility_bill'::text, 'selfie'::text])))
);


--
-- Name: TABLE kyc_documents; Type: COMMENT; Schema: sellers; Owner: -
--

COMMENT ON TABLE sellers.kyc_documents IS '[sellers] Documents uploaded with a KYC submission (private files; every view audited).';


--
-- Name: seller_bank_accounts; Type: TABLE; Schema: sellers; Owner: -
--

CREATE TABLE sellers.seller_bank_accounts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    seller_id uuid NOT NULL,
    bank_code text NOT NULL,
    account_name text NOT NULL,
    resolved_account_name text,
    account_number_encrypted bytea NOT NULL,
    account_number_hmac bytea NOT NULL,
    account_number_last4 text NOT NULL,
    encryption_key_id uuid,
    provider text,
    provider_recipient_code text,
    verified_at timestamp with time zone,
    is_default boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT seller_bank_accounts_account_number_last4_check CHECK ((length(account_number_last4) = 4)),
    CONSTRAINT seller_bank_accounts_provider_check CHECK ((provider = ANY (ARRAY['paystack'::text, 'moniepoint'::text, 'opay'::text])))
);


--
-- Name: TABLE seller_bank_accounts; Type: COMMENT; Schema: sellers; Owner: -
--

COMMENT ON TABLE sellers.seller_bank_accounts IS '[sellers] Payout bank accounts (NUBAN encrypted, last 4 shown). A new account triggers a 24-hour payout hold.';


--
-- Name: seller_kyc_submissions; Type: TABLE; Schema: sellers; Owner: -
--

CREATE TABLE sellers.seller_kyc_submissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    seller_id uuid NOT NULL,
    status text DEFAULT 'submitted'::text NOT NULL,
    id_type text NOT NULL,
    id_number_encrypted bytea NOT NULL,
    id_number_hmac bytea NOT NULL,
    id_number_last4 text NOT NULL,
    encryption_key_id uuid,
    cac_rc_number text,
    tin text,
    verification_provider text,
    verification_ref text,
    verified_at timestamp with time zone,
    reviewed_by uuid,
    reviewed_at timestamp with time zone,
    rejection_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT seller_kyc_submissions_check CHECK (((status <> ALL (ARRAY['approved'::text, 'rejected'::text])) OR ((reviewed_by IS NOT NULL) AND (reviewed_at IS NOT NULL)))),
    CONSTRAINT seller_kyc_submissions_check1 CHECK (((status <> 'rejected'::text) OR (rejection_reason IS NOT NULL))),
    CONSTRAINT seller_kyc_submissions_id_number_last4_check CHECK ((length(id_number_last4) = 4)),
    CONSTRAINT seller_kyc_submissions_id_type_check CHECK ((id_type = ANY (ARRAY['nin'::text, 'nin_slip'::text, 'passport'::text, 'drivers_licence'::text, 'voters_card'::text]))),
    CONSTRAINT seller_kyc_submissions_status_check CHECK ((status = ANY (ARRAY['submitted'::text, 'in_review'::text, 'approved'::text, 'rejected'::text, 'needs_more_info'::text])))
);


--
-- Name: TABLE seller_kyc_submissions; Type: COMMENT; Schema: sellers; Owner: -
--

COMMENT ON TABLE sellers.seller_kyc_submissions IS '[sellers] Identity and business verification. ID numbers encrypted (NDPA); the HMAC finds the same ID across sellers.';


--
-- Name: seller_members; Type: TABLE; Schema: sellers; Owner: -
--

CREATE TABLE sellers.seller_members (
    seller_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT seller_members_role_check CHECK ((role = ANY (ARRAY['owner'::text, 'manager'::text, 'staff'::text, 'finance'::text])))
);


--
-- Name: TABLE seller_members; Type: COMMENT; Schema: sellers; Owner: -
--

COMMENT ON TABLE sellers.seller_members IS '[sellers] Users who act for a seller in Seller Centre. The finance role needs TOTP.';


--
-- Name: sellers; Type: TABLE; Schema: sellers; Owner: -
--

CREATE TABLE sellers.sellers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    type text NOT NULL,
    display_name text NOT NULL,
    slug text NOT NULL,
    owner_user_id uuid,
    status text DEFAULT 'pending_kyc'::text NOT NULL,
    state_code text,
    city text,
    rating_avg numeric(3,2),
    rating_count integer DEFAULT 0 NOT NULL,
    commission_override_bps integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT sellers_check CHECK (((type = 'first_party'::text) OR (owner_user_id IS NOT NULL))),
    CONSTRAINT sellers_commission_override_bps_check CHECK (((commission_override_bps >= 0) AND (commission_override_bps <= 10000))),
    CONSTRAINT sellers_rating_avg_check CHECK (((rating_avg >= (1)::numeric) AND (rating_avg <= (5)::numeric))),
    CONSTRAINT sellers_rating_count_check CHECK ((rating_count >= 0)),
    CONSTRAINT sellers_slug_check CHECK ((slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'::text)),
    CONSTRAINT sellers_status_check CHECK ((status = ANY (ARRAY['pending_kyc'::text, 'active'::text, 'suspended'::text, 'closed'::text]))),
    CONSTRAINT sellers_type_check CHECK ((type = ANY (ARRAY['first_party'::text, 'business'::text, 'individual'::text])))
);


--
-- Name: TABLE sellers; Type: COMMENT; Schema: sellers; Owner: -
--

COMMENT ON TABLE sellers.sellers IS '[sellers] Everyone who sells: TechShop itself (first_party), marketplace businesses and individuals.';


--
-- Name: message_attachments; Type: TABLE; Schema: support; Owner: -
--

CREATE TABLE support.message_attachments (
    message_id uuid NOT NULL,
    file_id uuid NOT NULL
);


--
-- Name: TABLE message_attachments; Type: COMMENT; Schema: support; Owner: -
--

COMMENT ON TABLE support.message_attachments IS '[support] Files attached to ticket messages.';


--
-- Name: ticket_number_seq; Type: SEQUENCE; Schema: support; Owner: -
--

CREATE SEQUENCE support.ticket_number_seq
    START WITH 5001
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: support_tickets; Type: TABLE; Schema: support; Owner: -
--

CREATE TABLE support.support_tickets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    ticket_number text DEFAULT platform.next_ref('TKT-'::text, 'support.ticket_number_seq'::regclass) NOT NULL,
    customer_user_id uuid,
    seller_id uuid,
    order_id uuid,
    site text,
    contact_name text,
    contact_email public.citext,
    contact_phone text,
    channel text NOT NULL,
    topic text NOT NULL,
    priority text DEFAULT 'normal'::text NOT NULL,
    status text DEFAULT 'open'::text NOT NULL,
    assigned_to uuid,
    first_response_at timestamp with time zone,
    resolved_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT support_tickets_channel_check CHECK ((channel = ANY (ARRAY['web'::text, 'app'::text, 'whatsapp'::text, 'email'::text, 'phone'::text]))),
    CONSTRAINT support_tickets_check CHECK (((customer_user_id IS NOT NULL) OR (seller_id IS NOT NULL) OR (contact_email IS NOT NULL) OR (contact_phone IS NOT NULL))),
    CONSTRAINT support_tickets_contact_phone_check CHECK ((contact_phone ~ '^\+234[0-9]{10}$'::text)),
    CONSTRAINT support_tickets_priority_check CHECK ((priority = ANY (ARRAY['low'::text, 'normal'::text, 'high'::text, 'urgent'::text]))),
    CONSTRAINT support_tickets_site_check CHECK ((site = ANY (ARRAY['corporate'::text, 'market'::text, 'wholesale'::text, 'seller'::text, 'app'::text]))),
    CONSTRAINT support_tickets_status_check CHECK ((status = ANY (ARRAY['open'::text, 'pending_customer'::text, 'resolved'::text, 'closed'::text])))
);


--
-- Name: TABLE support_tickets; Type: COMMENT; Schema: support; Owner: -
--

COMMENT ON TABLE support.support_tickets IS '[support] Support conversations with customers, sellers or public contact-form visitors.';


--
-- Name: ticket_messages; Type: TABLE; Schema: support; Owner: -
--

CREATE TABLE support.ticket_messages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    ticket_id uuid NOT NULL,
    author_user_id uuid,
    author_kind text NOT NULL,
    body text NOT NULL,
    is_internal boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ticket_messages_author_kind_check CHECK ((author_kind = ANY (ARRAY['customer'::text, 'seller'::text, 'staff'::text, 'system'::text]))),
    CONSTRAINT ticket_messages_check CHECK (((NOT is_internal) OR (author_kind = ANY (ARRAY['staff'::text, 'system'::text]))))
);


--
-- Name: TABLE ticket_messages; Type: COMMENT; Schema: support; Owner: -
--

COMMENT ON TABLE support.ticket_messages IS '[support] Messages in a ticket; internal notes are never returned to customers.';


--
-- Name: auth_events_default; Type: TABLE ATTACH; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.auth_events ATTACH PARTITION identity.auth_events_default DEFAULT;


--
-- Name: rider_location_pings_default; Type: TABLE ATTACH; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_location_pings ATTACH PARTITION logistics.rider_location_pings_default DEFAULT;


--
-- Name: message_events_default; Type: TABLE ATTACH; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.message_events ATTACH PARTITION messaging.message_events_default DEFAULT;


--
-- Name: user_events_default; Type: TABLE ATTACH; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.user_events ATTACH PARTITION personalisation.user_events_default DEFAULT;


--
-- Name: risk_decisions_default; Type: TABLE ATTACH; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_decisions ATTACH PARTITION risk.risk_decisions_default DEFAULT;


--
-- Name: search_queries_default; Type: TABLE ATTACH; Schema: search; Owner: -
--

ALTER TABLE ONLY search.search_queries ATTACH PARTITION search.search_queries_default DEFAULT;


--
-- Name: repair_jobs repair_jobs_pkey; Type: CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.repair_jobs
    ADD CONSTRAINT repair_jobs_pkey PRIMARY KEY (id);


--
-- Name: return_inspections return_inspections_pkey; Type: CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.return_inspections
    ADD CONSTRAINT return_inspections_pkey PRIMARY KEY (id);


--
-- Name: return_items return_items_pkey; Type: CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.return_items
    ADD CONSTRAINT return_items_pkey PRIMARY KEY (id);


--
-- Name: return_items return_items_return_request_id_order_line_id_key; Type: CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.return_items
    ADD CONSTRAINT return_items_return_request_id_order_line_id_key UNIQUE (return_request_id, order_line_id);


--
-- Name: return_requests return_requests_pkey; Type: CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.return_requests
    ADD CONSTRAINT return_requests_pkey PRIMARY KEY (id);


--
-- Name: return_requests return_requests_rma_number_key; Type: CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.return_requests
    ADD CONSTRAINT return_requests_rma_number_key UNIQUE (rma_number);


--
-- Name: trade_ins trade_ins_device_unit_id_key; Type: CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.trade_ins
    ADD CONSTRAINT trade_ins_device_unit_id_key UNIQUE (device_unit_id);


--
-- Name: trade_ins trade_ins_pkey; Type: CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.trade_ins
    ADD CONSTRAINT trade_ins_pkey PRIMARY KEY (id);


--
-- Name: trade_ins trade_ins_reference_key; Type: CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.trade_ins
    ADD CONSTRAINT trade_ins_reference_key UNIQUE (reference);


--
-- Name: warranty_claims warranty_claims_claim_number_key; Type: CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.warranty_claims
    ADD CONSTRAINT warranty_claims_claim_number_key UNIQUE (claim_number);


--
-- Name: warranty_claims warranty_claims_pkey; Type: CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.warranty_claims
    ADD CONSTRAINT warranty_claims_pkey PRIMARY KEY (id);


--
-- Name: business_members business_members_pkey; Type: CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.business_members
    ADD CONSTRAINT business_members_pkey PRIMARY KEY (business_id, user_id);


--
-- Name: businesses businesses_pkey; Type: CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.businesses
    ADD CONSTRAINT businesses_pkey PRIMARY KEY (id);


--
-- Name: businesses businesses_rc_number_key; Type: CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.businesses
    ADD CONSTRAINT businesses_rc_number_key UNIQUE (rc_number);


--
-- Name: credit_accounts credit_accounts_business_id_key; Type: CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.credit_accounts
    ADD CONSTRAINT credit_accounts_business_id_key UNIQUE (business_id);


--
-- Name: credit_accounts credit_accounts_pkey; Type: CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.credit_accounts
    ADD CONSTRAINT credit_accounts_pkey PRIMARY KEY (id);


--
-- Name: quote_lines quote_lines_pkey; Type: CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.quote_lines
    ADD CONSTRAINT quote_lines_pkey PRIMARY KEY (id);


--
-- Name: quotes quotes_pkey; Type: CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.quotes
    ADD CONSTRAINT quotes_pkey PRIMARY KEY (id);


--
-- Name: quotes quotes_quote_number_key; Type: CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.quotes
    ADD CONSTRAINT quotes_quote_number_key UNIQUE (quote_number);


--
-- Name: car_documents car_documents_pkey; Type: CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_documents
    ADD CONSTRAINT car_documents_pkey PRIMARY KEY (id);


--
-- Name: car_images car_images_pkey; Type: CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_images
    ADD CONSTRAINT car_images_pkey PRIMARY KEY (id);


--
-- Name: car_inspections car_inspections_pkey; Type: CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_inspections
    ADD CONSTRAINT car_inspections_pkey PRIMARY KEY (id);


--
-- Name: car_listings car_listings_pkey; Type: CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_listings
    ADD CONSTRAINT car_listings_pkey PRIMARY KEY (id);


--
-- Name: car_listings car_listings_reference_key; Type: CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_listings
    ADD CONSTRAINT car_listings_reference_key UNIQUE (reference);


--
-- Name: car_listings car_listings_slug_key; Type: CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_listings
    ADD CONSTRAINT car_listings_slug_key UNIQUE (slug);


--
-- Name: car_listings car_listings_vin_key; Type: CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_listings
    ADD CONSTRAINT car_listings_vin_key UNIQUE (vin);


--
-- Name: financing_enquiries financing_enquiries_pkey; Type: CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.financing_enquiries
    ADD CONSTRAINT financing_enquiries_pkey PRIMARY KEY (id);


--
-- Name: viewing_bookings viewing_bookings_pkey; Type: CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.viewing_bookings
    ADD CONSTRAINT viewing_bookings_pkey PRIMARY KEY (id);


--
-- Name: attribute_definitions attribute_definitions_category_id_key_key; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.attribute_definitions
    ADD CONSTRAINT attribute_definitions_category_id_key_key UNIQUE (category_id, key);


--
-- Name: attribute_definitions attribute_definitions_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.attribute_definitions
    ADD CONSTRAINT attribute_definitions_pkey PRIMARY KEY (id);


--
-- Name: brands brands_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.brands
    ADD CONSTRAINT brands_pkey PRIMARY KEY (id);


--
-- Name: brands brands_slug_key; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.brands
    ADD CONSTRAINT brands_slug_key UNIQUE (slug);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: categories categories_slug_key; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.categories
    ADD CONSTRAINT categories_slug_key UNIQUE (slug);


--
-- Name: image_hashes image_hashes_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.image_hashes
    ADD CONSTRAINT image_hashes_pkey PRIMARY KEY (product_image_id);


--
-- Name: listing_price_history listing_price_history_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.listing_price_history
    ADD CONSTRAINT listing_price_history_pkey PRIMARY KEY (id);


--
-- Name: listing_reviews listing_reviews_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.listing_reviews
    ADD CONSTRAINT listing_reviews_pkey PRIMARY KEY (id);


--
-- Name: listings listings_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.listings
    ADD CONSTRAINT listings_pkey PRIMARY KEY (id);


--
-- Name: listings listings_seller_id_variant_id_condition_key; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.listings
    ADD CONSTRAINT listings_seller_id_variant_id_condition_key UNIQUE (seller_id, variant_id, condition);


--
-- Name: price_tiers price_tiers_listing_id_min_quantity_key; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.price_tiers
    ADD CONSTRAINT price_tiers_listing_id_min_quantity_key UNIQUE (listing_id, min_quantity);


--
-- Name: price_tiers price_tiers_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.price_tiers
    ADD CONSTRAINT price_tiers_pkey PRIMARY KEY (id);


--
-- Name: product_attribute_values product_attribute_values_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_attribute_values
    ADD CONSTRAINT product_attribute_values_pkey PRIMARY KEY (product_id, attribute_id);


--
-- Name: product_compatibility product_compatibility_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_compatibility
    ADD CONSTRAINT product_compatibility_pkey PRIMARY KEY (id);


--
-- Name: product_images product_images_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_images
    ADD CONSTRAINT product_images_pkey PRIMARY KEY (id);


--
-- Name: product_offer_summary product_offer_summary_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_offer_summary
    ADD CONSTRAINT product_offer_summary_pkey PRIMARY KEY (product_id, channel);


--
-- Name: product_reviews product_reviews_order_line_id_key; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_reviews
    ADD CONSTRAINT product_reviews_order_line_id_key UNIQUE (order_line_id);


--
-- Name: product_reviews product_reviews_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_reviews
    ADD CONSTRAINT product_reviews_pkey PRIMARY KEY (id);


--
-- Name: product_suggestions product_suggestions_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_suggestions
    ADD CONSTRAINT product_suggestions_pkey PRIMARY KEY (id);


--
-- Name: product_variants product_variants_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_variants
    ADD CONSTRAINT product_variants_pkey PRIMARY KEY (id);


--
-- Name: product_variants product_variants_sku_key; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_variants
    ADD CONSTRAINT product_variants_sku_key UNIQUE (sku);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: products products_slug_key; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.products
    ADD CONSTRAINT products_slug_key UNIQUE (slug);


--
-- Name: review_images review_images_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.review_images
    ADD CONSTRAINT review_images_pkey PRIMARY KEY (review_id, file_id);


--
-- Name: saved_items saved_items_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.saved_items
    ADD CONSTRAINT saved_items_pkey PRIMARY KEY (user_id, product_id);


--
-- Name: seller_ratings seller_ratings_fulfilment_id_key; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.seller_ratings
    ADD CONSTRAINT seller_ratings_fulfilment_id_key UNIQUE (fulfilment_id);


--
-- Name: seller_ratings seller_ratings_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.seller_ratings
    ADD CONSTRAINT seller_ratings_pkey PRIMARY KEY (id);


--
-- Name: stock_alerts stock_alerts_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.stock_alerts
    ADD CONSTRAINT stock_alerts_pkey PRIMARY KEY (user_id, product_id);


--
-- Name: banners banners_pkey; Type: CONSTRAINT; Schema: content; Owner: -
--

ALTER TABLE ONLY content.banners
    ADD CONSTRAINT banners_pkey PRIMARY KEY (id);


--
-- Name: cms_pages cms_pages_pkey; Type: CONSTRAINT; Schema: content; Owner: -
--

ALTER TABLE ONLY content.cms_pages
    ADD CONSTRAINT cms_pages_pkey PRIMARY KEY (id);


--
-- Name: cms_pages cms_pages_site_slug_key; Type: CONSTRAINT; Schema: content; Owner: -
--

ALTER TABLE ONLY content.cms_pages
    ADD CONSTRAINT cms_pages_site_slug_key UNIQUE (site, slug);


--
-- Name: help_articles help_articles_pkey; Type: CONSTRAINT; Schema: content; Owner: -
--

ALTER TABLE ONLY content.help_articles
    ADD CONSTRAINT help_articles_pkey PRIMARY KEY (id);


--
-- Name: help_articles help_articles_site_slug_key; Type: CONSTRAINT; Schema: content; Owner: -
--

ALTER TABLE ONLY content.help_articles
    ADD CONSTRAINT help_articles_site_slug_key UNIQUE (site, slug);


--
-- Name: accounting_periods accounting_periods_pkey; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.accounting_periods
    ADD CONSTRAINT accounting_periods_pkey PRIMARY KEY (month);


--
-- Name: commission_rules commission_rules_pkey; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.commission_rules
    ADD CONSTRAINT commission_rules_pkey PRIMARY KEY (id);


--
-- Name: invoices invoices_invoice_number_key; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.invoices
    ADD CONSTRAINT invoices_invoice_number_key UNIQUE (invoice_number);


--
-- Name: invoices invoices_order_id_key; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.invoices
    ADD CONSTRAINT invoices_order_id_key UNIQUE (order_id);


--
-- Name: invoices invoices_pkey; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.invoices
    ADD CONSTRAINT invoices_pkey PRIMARY KEY (id);


--
-- Name: ledger_accounts ledger_accounts_code_key; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_accounts
    ADD CONSTRAINT ledger_accounts_code_key UNIQUE (code);


--
-- Name: ledger_accounts ledger_accounts_pkey; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_accounts
    ADD CONSTRAINT ledger_accounts_pkey PRIMARY KEY (id);


--
-- Name: ledger_accounts ledger_accounts_seller_id_key; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_accounts
    ADD CONSTRAINT ledger_accounts_seller_id_key UNIQUE (seller_id);


--
-- Name: ledger_accounts ledger_accounts_user_id_key; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_accounts
    ADD CONSTRAINT ledger_accounts_user_id_key UNIQUE (user_id);


--
-- Name: ledger_entries ledger_entries_pkey; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_entries
    ADD CONSTRAINT ledger_entries_pkey PRIMARY KEY (id);


--
-- Name: ledger_journals ledger_journals_pkey; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_journals
    ADD CONSTRAINT ledger_journals_pkey PRIMARY KEY (id);


--
-- Name: ledger_journals ledger_journals_posting_key_key; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_journals
    ADD CONSTRAINT ledger_journals_posting_key_key UNIQUE (posting_key);


--
-- Name: ledger_journals ledger_journals_reverses_journal_id_key; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_journals
    ADD CONSTRAINT ledger_journals_reverses_journal_id_key UNIQUE (reverses_journal_id);


--
-- Name: payout_batches payout_batches_pkey; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payout_batches
    ADD CONSTRAINT payout_batches_pkey PRIMARY KEY (id);


--
-- Name: payout_holds payout_holds_pkey; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payout_holds
    ADD CONSTRAINT payout_holds_pkey PRIMARY KEY (id);


--
-- Name: payout_items payout_items_fulfilment_id_key; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payout_items
    ADD CONSTRAINT payout_items_fulfilment_id_key UNIQUE (fulfilment_id);


--
-- Name: payout_items payout_items_pkey; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payout_items
    ADD CONSTRAINT payout_items_pkey PRIMARY KEY (payout_id, fulfilment_id);


--
-- Name: payouts payouts_pkey; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payouts
    ADD CONSTRAINT payouts_pkey PRIMARY KEY (id);


--
-- Name: payouts payouts_provider_reference_key; Type: CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payouts
    ADD CONSTRAINT payouts_provider_reference_key UNIQUE (provider_reference);


--
-- Name: carriers carriers_code_key; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.carriers
    ADD CONSTRAINT carriers_code_key UNIQUE (code);


--
-- Name: carriers carriers_pkey; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.carriers
    ADD CONSTRAINT carriers_pkey PRIMARY KEY (id);


--
-- Name: manifest_parcels manifest_parcels_parcel_id_key; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.manifest_parcels
    ADD CONSTRAINT manifest_parcels_parcel_id_key UNIQUE (parcel_id);


--
-- Name: manifest_parcels manifest_parcels_pkey; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.manifest_parcels
    ADD CONSTRAINT manifest_parcels_pkey PRIMARY KEY (manifest_id, parcel_id);


--
-- Name: manifests manifests_pkey; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.manifests
    ADD CONSTRAINT manifests_pkey PRIMARY KEY (id);


--
-- Name: pack_records pack_records_parcel_id_key; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pack_records
    ADD CONSTRAINT pack_records_parcel_id_key UNIQUE (parcel_id);


--
-- Name: pack_records pack_records_pkey; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pack_records
    ADD CONSTRAINT pack_records_pkey PRIMARY KEY (id);


--
-- Name: parcels parcels_label_code_key; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.parcels
    ADD CONSTRAINT parcels_label_code_key UNIQUE (label_code);


--
-- Name: parcels parcels_pkey; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.parcels
    ADD CONSTRAINT parcels_pkey PRIMARY KEY (id);


--
-- Name: pick_list_items pick_list_items_pick_list_id_order_line_id_key; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pick_list_items
    ADD CONSTRAINT pick_list_items_pick_list_id_order_line_id_key UNIQUE (pick_list_id, order_line_id);


--
-- Name: pick_list_items pick_list_items_pkey; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pick_list_items
    ADD CONSTRAINT pick_list_items_pkey PRIMARY KEY (id);


--
-- Name: pick_lists pick_lists_pkey; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pick_lists
    ADD CONSTRAINT pick_lists_pkey PRIMARY KEY (id);


--
-- Name: seller_sla_events seller_sla_events_fulfilment_id_kind_key; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.seller_sla_events
    ADD CONSTRAINT seller_sla_events_fulfilment_id_kind_key UNIQUE (fulfilment_id, kind);


--
-- Name: seller_sla_events seller_sla_events_pkey; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.seller_sla_events
    ADD CONSTRAINT seller_sla_events_pkey PRIMARY KEY (id);


--
-- Name: shipment_events shipment_events_pkey; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.shipment_events
    ADD CONSTRAINT shipment_events_pkey PRIMARY KEY (id);


--
-- Name: shipment_events shipment_events_shipment_id_carrier_event_id_key; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.shipment_events
    ADD CONSTRAINT shipment_events_shipment_id_carrier_event_id_key UNIQUE (shipment_id, carrier_event_id);


--
-- Name: shipments shipments_carrier_id_tracking_number_key; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.shipments
    ADD CONSTRAINT shipments_carrier_id_tracking_number_key UNIQUE (carrier_id, tracking_number);


--
-- Name: shipments shipments_pkey; Type: CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.shipments
    ADD CONSTRAINT shipments_pkey PRIMARY KEY (id);


--
-- Name: addresses addresses_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.addresses
    ADD CONSTRAINT addresses_pkey PRIMARY KEY (id);


--
-- Name: auth_events auth_events_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.auth_events
    ADD CONSTRAINT auth_events_pkey PRIMARY KEY (id, created_at);


--
-- Name: auth_events_default auth_events_default_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.auth_events_default
    ADD CONSTRAINT auth_events_default_pkey PRIMARY KEY (id, created_at);


--
-- Name: auth_rate_limits auth_rate_limits_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.auth_rate_limits
    ADD CONSTRAINT auth_rate_limits_pkey PRIMARY KEY (key, window_start);


--
-- Name: consents consents_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.consents
    ADD CONSTRAINT consents_pkey PRIMARY KEY (id);


--
-- Name: nigerian_lgas nigerian_lgas_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.nigerian_lgas
    ADD CONSTRAINT nigerian_lgas_pkey PRIMARY KEY (id);


--
-- Name: nigerian_lgas nigerian_lgas_state_code_name_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.nigerian_lgas
    ADD CONSTRAINT nigerian_lgas_state_code_name_key UNIQUE (state_code, name);


--
-- Name: nigerian_states nigerian_states_name_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.nigerian_states
    ADD CONSTRAINT nigerian_states_name_key UNIQUE (name);


--
-- Name: nigerian_states nigerian_states_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.nigerian_states
    ADD CONSTRAINT nigerian_states_pkey PRIMARY KEY (code);


--
-- Name: permissions permissions_key_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.permissions
    ADD CONSTRAINT permissions_key_key UNIQUE (key);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: privacy_requests privacy_requests_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.privacy_requests
    ADD CONSTRAINT privacy_requests_pkey PRIMARY KEY (id);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_id);


--
-- Name: roles roles_key_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.roles
    ADD CONSTRAINT roles_key_key UNIQUE (key);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: service_clients service_clients_name_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.service_clients
    ADD CONSTRAINT service_clients_name_key UNIQUE (name);


--
-- Name: service_clients service_clients_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.service_clients
    ADD CONSTRAINT service_clients_pkey PRIMARY KEY (id);


--
-- Name: staff_invites staff_invites_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.staff_invites
    ADD CONSTRAINT staff_invites_pkey PRIMARY KEY (id);


--
-- Name: staff_invites staff_invites_token_hash_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.staff_invites
    ADD CONSTRAINT staff_invites_token_hash_key UNIQUE (token_hash);


--
-- Name: staff_members staff_members_employee_no_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.staff_members
    ADD CONSTRAINT staff_members_employee_no_key UNIQUE (employee_no);


--
-- Name: staff_members staff_members_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.staff_members
    ADD CONSTRAINT staff_members_pkey PRIMARY KEY (user_id);


--
-- Name: staff_role_scopes staff_role_scopes_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.staff_role_scopes
    ADD CONSTRAINT staff_role_scopes_pkey PRIMARY KEY (user_id, role_id, scope_type, scope_id);


--
-- Name: staff_roles staff_roles_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.staff_roles
    ADD CONSTRAINT staff_roles_pkey PRIMARY KEY (user_id, role_id);


--
-- Name: user_mfa_factors user_mfa_factors_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.user_mfa_factors
    ADD CONSTRAINT user_mfa_factors_pkey PRIMARY KEY (id);


--
-- Name: user_recovery_codes user_recovery_codes_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.user_recovery_codes
    ADD CONSTRAINT user_recovery_codes_pkey PRIMARY KEY (id);


--
-- Name: user_recovery_codes user_recovery_codes_user_id_code_hash_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.user_recovery_codes
    ADD CONSTRAINT user_recovery_codes_user_id_code_hash_key UNIQUE (user_id, code_hash);


--
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);


--
-- Name: user_sessions user_sessions_refresh_token_hash_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.user_sessions
    ADD CONSTRAINT user_sessions_refresh_token_hash_key UNIQUE (refresh_token_hash);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_phone_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.users
    ADD CONSTRAINT users_phone_key UNIQUE (phone);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: verification_codes verification_codes_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.verification_codes
    ADD CONSTRAINT verification_codes_pkey PRIMARY KEY (id);


--
-- Name: webauthn_credentials webauthn_credentials_credential_id_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.webauthn_credentials
    ADD CONSTRAINT webauthn_credentials_credential_id_key UNIQUE (credential_id);


--
-- Name: webauthn_credentials webauthn_credentials_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.webauthn_credentials
    ADD CONSTRAINT webauthn_credentials_pkey PRIMARY KEY (id);


--
-- Name: bin_locations bin_locations_pkey; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.bin_locations
    ADD CONSTRAINT bin_locations_pkey PRIMARY KEY (id);


--
-- Name: bin_locations bin_locations_warehouse_id_code_key; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.bin_locations
    ADD CONSTRAINT bin_locations_warehouse_id_code_key UNIQUE (warehouse_id, code);


--
-- Name: device_units device_units_imei_key; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.device_units
    ADD CONSTRAINT device_units_imei_key UNIQUE (imei);


--
-- Name: device_units device_units_pkey; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.device_units
    ADD CONSTRAINT device_units_pkey PRIMARY KEY (id);


--
-- Name: inventory_costs inventory_costs_pkey; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_costs
    ADD CONSTRAINT inventory_costs_pkey PRIMARY KEY (variant_id, condition, warehouse_id, owner_seller_id);


--
-- Name: inventory_count_lines inventory_count_lines_count_id_variant_id_condition_owner_s_key; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_count_lines
    ADD CONSTRAINT inventory_count_lines_count_id_variant_id_condition_owner_s_key UNIQUE (count_id, variant_id, condition, owner_seller_id);


--
-- Name: inventory_count_lines inventory_count_lines_pkey; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_count_lines
    ADD CONSTRAINT inventory_count_lines_pkey PRIMARY KEY (id);


--
-- Name: inventory_counts inventory_counts_pkey; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_counts
    ADD CONSTRAINT inventory_counts_pkey PRIMARY KEY (id);


--
-- Name: inventory_levels inventory_levels_pkey; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_levels
    ADD CONSTRAINT inventory_levels_pkey PRIMARY KEY (warehouse_id, variant_id, condition, owner_seller_id);


--
-- Name: stock_movements stock_movements_pkey; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_movements
    ADD CONSTRAINT stock_movements_pkey PRIMARY KEY (id);


--
-- Name: stock_reservations stock_reservations_pkey; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_reservations
    ADD CONSTRAINT stock_reservations_pkey PRIMARY KEY (id);


--
-- Name: stock_transfer_lines stock_transfer_lines_pkey; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_transfer_lines
    ADD CONSTRAINT stock_transfer_lines_pkey PRIMARY KEY (id);


--
-- Name: stock_transfers stock_transfers_pkey; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_transfers
    ADD CONSTRAINT stock_transfers_pkey PRIMARY KEY (id);


--
-- Name: warehouses warehouses_code_key; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.warehouses
    ADD CONSTRAINT warehouses_code_key UNIQUE (code);


--
-- Name: warehouses warehouses_pkey; Type: CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.warehouses
    ADD CONSTRAINT warehouses_pkey PRIMARY KEY (id);


--
-- Name: delivery_events delivery_events_client_event_id_key; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_events
    ADD CONSTRAINT delivery_events_client_event_id_key UNIQUE (client_event_id);


--
-- Name: delivery_events delivery_events_pkey; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_events
    ADD CONSTRAINT delivery_events_pkey PRIMARY KEY (id);


--
-- Name: delivery_jobs delivery_jobs_pkey; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_jobs
    ADD CONSTRAINT delivery_jobs_pkey PRIMARY KEY (id);


--
-- Name: delivery_rates delivery_rates_pkey; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_rates
    ADD CONSTRAINT delivery_rates_pkey PRIMARY KEY (id);


--
-- Name: delivery_rates delivery_rates_zone_id_fulfilled_by_max_weight_grams_valid__key; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_rates
    ADD CONSTRAINT delivery_rates_zone_id_fulfilled_by_max_weight_grams_valid__key UNIQUE (zone_id, fulfilled_by, max_weight_grams, valid_from);


--
-- Name: delivery_zones delivery_zones_name_key; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_zones
    ADD CONSTRAINT delivery_zones_name_key UNIQUE (name);


--
-- Name: delivery_zones delivery_zones_pkey; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_zones
    ADD CONSTRAINT delivery_zones_pkey PRIMARY KEY (id);


--
-- Name: rider_devices rider_devices_pkey; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_devices
    ADD CONSTRAINT rider_devices_pkey PRIMARY KEY (id);


--
-- Name: rider_devices rider_devices_rider_id_device_id_key; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_devices
    ADD CONSTRAINT rider_devices_rider_id_device_id_key UNIQUE (rider_id, device_id);


--
-- Name: rider_location_pings rider_location_pings_pkey; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_location_pings
    ADD CONSTRAINT rider_location_pings_pkey PRIMARY KEY (rider_id, recorded_at);


--
-- Name: rider_location_pings_default rider_location_pings_default_pkey; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_location_pings_default
    ADD CONSTRAINT rider_location_pings_default_pkey PRIMARY KEY (rider_id, recorded_at);


--
-- Name: rider_locations rider_locations_pkey; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_locations
    ADD CONSTRAINT rider_locations_pkey PRIMARY KEY (rider_id);


--
-- Name: rider_shifts rider_shifts_pkey; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_shifts
    ADD CONSTRAINT rider_shifts_pkey PRIMARY KEY (id);


--
-- Name: rider_zones rider_zones_pkey; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_zones
    ADD CONSTRAINT rider_zones_pkey PRIMARY KEY (rider_id, zone_id);


--
-- Name: riders riders_pkey; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.riders
    ADD CONSTRAINT riders_pkey PRIMARY KEY (user_id);


--
-- Name: riders riders_plate_number_key; Type: CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.riders
    ADD CONSTRAINT riders_plate_number_key UNIQUE (plate_number);


--
-- Name: campaign_recipients campaign_recipients_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.campaign_recipients
    ADD CONSTRAINT campaign_recipients_pkey PRIMARY KEY (campaign_id, user_id);


--
-- Name: campaign_variants campaign_variants_campaign_id_label_channel_key; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.campaign_variants
    ADD CONSTRAINT campaign_variants_campaign_id_label_channel_key UNIQUE (campaign_id, label, channel);


--
-- Name: campaign_variants campaign_variants_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.campaign_variants
    ADD CONSTRAINT campaign_variants_pkey PRIMARY KEY (id);


--
-- Name: campaigns campaigns_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.campaigns
    ADD CONSTRAINT campaigns_pkey PRIMARY KEY (id);


--
-- Name: coupon_codes coupon_codes_code_key; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupon_codes
    ADD CONSTRAINT coupon_codes_code_key UNIQUE (code);


--
-- Name: coupon_codes coupon_codes_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupon_codes
    ADD CONSTRAINT coupon_codes_pkey PRIMARY KEY (id);


--
-- Name: coupon_redemptions coupon_redemptions_coupon_code_id_key; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupon_redemptions
    ADD CONSTRAINT coupon_redemptions_coupon_code_id_key UNIQUE (coupon_code_id);


--
-- Name: coupon_redemptions coupon_redemptions_order_id_key; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupon_redemptions
    ADD CONSTRAINT coupon_redemptions_order_id_key UNIQUE (order_id);


--
-- Name: coupon_redemptions coupon_redemptions_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupon_redemptions
    ADD CONSTRAINT coupon_redemptions_pkey PRIMARY KEY (id);


--
-- Name: coupons coupons_code_key; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupons
    ADD CONSTRAINT coupons_code_key UNIQUE (code);


--
-- Name: coupons coupons_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupons
    ADD CONSTRAINT coupons_pkey PRIMARY KEY (id);


--
-- Name: flash_claims flash_claims_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.flash_claims
    ADD CONSTRAINT flash_claims_pkey PRIMARY KEY (id);


--
-- Name: flash_claims flash_claims_promotion_id_listing_id_order_id_key; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.flash_claims
    ADD CONSTRAINT flash_claims_promotion_id_listing_id_order_id_key UNIQUE (promotion_id, listing_id, order_id);


--
-- Name: journey_enrollments journey_enrollments_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.journey_enrollments
    ADD CONSTRAINT journey_enrollments_pkey PRIMARY KEY (id);


--
-- Name: journey_steps journey_steps_journey_id_position_key; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.journey_steps
    ADD CONSTRAINT journey_steps_journey_id_position_key UNIQUE (journey_id, "position");


--
-- Name: journey_steps journey_steps_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.journey_steps
    ADD CONSTRAINT journey_steps_pkey PRIMARY KEY (id);


--
-- Name: journeys journeys_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.journeys
    ADD CONSTRAINT journeys_pkey PRIMARY KEY (id);


--
-- Name: promotion_listing_prices promotion_listing_prices_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.promotion_listing_prices
    ADD CONSTRAINT promotion_listing_prices_pkey PRIMARY KEY (promotion_id, listing_id);


--
-- Name: promotion_targets promotion_targets_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.promotion_targets
    ADD CONSTRAINT promotion_targets_pkey PRIMARY KEY (id);


--
-- Name: promotions promotions_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.promotions
    ADD CONSTRAINT promotions_pkey PRIMARY KEY (id);


--
-- Name: segments segments_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.segments
    ADD CONSTRAINT segments_pkey PRIMARY KEY (id);


--
-- Name: tracked_links tracked_links_pkey; Type: CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.tracked_links
    ADD CONSTRAINT tracked_links_pkey PRIMARY KEY (id);


--
-- Name: message_events message_events_pkey; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.message_events
    ADD CONSTRAINT message_events_pkey PRIMARY KEY (id, occurred_at);


--
-- Name: message_events_default message_events_default_pkey; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.message_events_default
    ADD CONSTRAINT message_events_default_pkey PRIMARY KEY (id, occurred_at);


--
-- Name: message_suppressions message_suppressions_channel_address_hash_key; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.message_suppressions
    ADD CONSTRAINT message_suppressions_channel_address_hash_key UNIQUE (channel, address_hash);


--
-- Name: message_suppressions message_suppressions_pkey; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.message_suppressions
    ADD CONSTRAINT message_suppressions_pkey PRIMARY KEY (id);


--
-- Name: message_templates message_templates_key_channel_locale_key; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.message_templates
    ADD CONSTRAINT message_templates_key_channel_locale_key UNIQUE (key, channel, locale);


--
-- Name: message_templates message_templates_pkey; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.message_templates
    ADD CONSTRAINT message_templates_pkey PRIMARY KEY (id);


--
-- Name: notification_preferences notification_preferences_pkey; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notification_preferences
    ADD CONSTRAINT notification_preferences_pkey PRIMARY KEY (user_id, category, channel);


--
-- Name: notification_settings notification_settings_pkey; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notification_settings
    ADD CONSTRAINT notification_settings_pkey PRIMARY KEY (user_id);


--
-- Name: notifications notifications_idempotency_key_key; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notifications
    ADD CONSTRAINT notifications_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: push_tokens push_tokens_pkey; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.push_tokens
    ADD CONSTRAINT push_tokens_pkey PRIMARY KEY (id);


--
-- Name: push_tokens push_tokens_token_key; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.push_tokens
    ADD CONSTRAINT push_tokens_token_key UNIQUE (token);


--
-- Name: template_versions template_versions_pkey; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.template_versions
    ADD CONSTRAINT template_versions_pkey PRIMARY KEY (id);


--
-- Name: template_versions template_versions_template_id_version_key; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.template_versions
    ADD CONSTRAINT template_versions_template_id_version_key UNIQUE (template_id, version);


--
-- Name: bank_statement_lines bank_statement_lines_pkey; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.bank_statement_lines
    ADD CONSTRAINT bank_statement_lines_pkey PRIMARY KEY (id);


--
-- Name: dispute_evidence dispute_evidence_pkey; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.dispute_evidence
    ADD CONSTRAINT dispute_evidence_pkey PRIMARY KEY (id);


--
-- Name: disputes disputes_pkey; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.disputes
    ADD CONSTRAINT disputes_pkey PRIMARY KEY (id);


--
-- Name: disputes disputes_provider_provider_dispute_id_key; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.disputes
    ADD CONSTRAINT disputes_provider_provider_dispute_id_key UNIQUE (provider, provider_dispute_id);


--
-- Name: payment_events payment_events_pkey; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.payment_events
    ADD CONSTRAINT payment_events_pkey PRIMARY KEY (id);


--
-- Name: payment_events payment_events_provider_provider_event_id_key; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.payment_events
    ADD CONSTRAINT payment_events_provider_provider_event_id_key UNIQUE (provider, provider_event_id);


--
-- Name: payments payments_idempotency_key_key; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.payments
    ADD CONSTRAINT payments_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: payments payments_pkey; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);


--
-- Name: payments payments_provider_provider_reference_key; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.payments
    ADD CONSTRAINT payments_provider_provider_reference_key UNIQUE (provider, provider_reference);


--
-- Name: provider_health provider_health_pkey; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.provider_health
    ADD CONSTRAINT provider_health_pkey PRIMARY KEY (provider);


--
-- Name: reconciliation_exceptions reconciliation_exceptions_pkey; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.reconciliation_exceptions
    ADD CONSTRAINT reconciliation_exceptions_pkey PRIMARY KEY (id);


--
-- Name: refunds refunds_pkey; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.refunds
    ADD CONSTRAINT refunds_pkey PRIMARY KEY (id);


--
-- Name: settlement_lines settlement_lines_pkey; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.settlement_lines
    ADD CONSTRAINT settlement_lines_pkey PRIMARY KEY (id);


--
-- Name: settlement_lines settlement_lines_report_id_provider_reference_key; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.settlement_lines
    ADD CONSTRAINT settlement_lines_report_id_provider_reference_key UNIQUE (report_id, provider_reference);


--
-- Name: settlement_reports settlement_reports_pkey; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.settlement_reports
    ADD CONSTRAINT settlement_reports_pkey PRIMARY KEY (id);


--
-- Name: settlement_reports settlement_reports_provider_batch_ref_key; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.settlement_reports
    ADD CONSTRAINT settlement_reports_provider_batch_ref_key UNIQUE (provider, batch_ref);


--
-- Name: virtual_accounts virtual_accounts_pkey; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.virtual_accounts
    ADD CONSTRAINT virtual_accounts_pkey PRIMARY KEY (id);


--
-- Name: virtual_accounts virtual_accounts_provider_account_number_key; Type: CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.virtual_accounts
    ADD CONSTRAINT virtual_accounts_provider_account_number_key UNIQUE (provider, account_number);


--
-- Name: identity_links identity_links_pkey; Type: CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.identity_links
    ADD CONSTRAINT identity_links_pkey PRIMARY KEY (anonymous_id);


--
-- Name: rec_item_neighbours rec_item_neighbours_pkey; Type: CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_item_neighbours
    ADD CONSTRAINT rec_item_neighbours_pkey PRIMARY KEY (model_version_id, kind, product_id, rank);


--
-- Name: rec_item_vectors rec_item_vectors_pkey; Type: CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_item_vectors
    ADD CONSTRAINT rec_item_vectors_pkey PRIMARY KEY (model_version_id, product_id);


--
-- Name: rec_model_versions rec_model_versions_pkey; Type: CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_model_versions
    ADD CONSTRAINT rec_model_versions_pkey PRIMARY KEY (id);


--
-- Name: rec_overrides rec_overrides_pkey; Type: CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_overrides
    ADD CONSTRAINT rec_overrides_pkey PRIMARY KEY (id);


--
-- Name: rec_popularity rec_popularity_pkey; Type: CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_popularity
    ADD CONSTRAINT rec_popularity_pkey PRIMARY KEY (id);


--
-- Name: rec_staging_item_neighbours rec_staging_item_neighbours_pkey; Type: CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_staging_item_neighbours
    ADD CONSTRAINT rec_staging_item_neighbours_pkey PRIMARY KEY (model_version_id, kind, product_id, rank);


--
-- Name: rec_staging_item_vectors rec_staging_item_vectors_pkey; Type: CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_staging_item_vectors
    ADD CONSTRAINT rec_staging_item_vectors_pkey PRIMARY KEY (model_version_id, product_id);


--
-- Name: rec_staging_popularity rec_staging_popularity_pkey; Type: CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_staging_popularity
    ADD CONSTRAINT rec_staging_popularity_pkey PRIMARY KEY (id);


--
-- Name: rec_staging_subject_candidates rec_staging_subject_candidates_pkey; Type: CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_staging_subject_candidates
    ADD CONSTRAINT rec_staging_subject_candidates_pkey PRIMARY KEY (model_version_id, user_id, rank);


--
-- Name: rec_subject_candidates rec_subject_candidates_pkey; Type: CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_subject_candidates
    ADD CONSTRAINT rec_subject_candidates_pkey PRIMARY KEY (model_version_id, user_id, rank);


--
-- Name: user_events user_events_pkey; Type: CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.user_events
    ADD CONSTRAINT user_events_pkey PRIMARY KEY (id, received_at);


--
-- Name: user_events_default user_events_default_pkey; Type: CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.user_events_default
    ADD CONSTRAINT user_events_default_pkey PRIMARY KEY (id, received_at);


--
-- Name: audit_log audit_log_pkey; Type: CONSTRAINT; Schema: platform; Owner: -
--

ALTER TABLE ONLY platform.audit_log
    ADD CONSTRAINT audit_log_pkey PRIMARY KEY (id);


--
-- Name: encryption_keys encryption_keys_pkey; Type: CONSTRAINT; Schema: platform; Owner: -
--

ALTER TABLE ONLY platform.encryption_keys
    ADD CONSTRAINT encryption_keys_pkey PRIMARY KEY (id);


--
-- Name: feature_flags feature_flags_pkey; Type: CONSTRAINT; Schema: platform; Owner: -
--

ALTER TABLE ONLY platform.feature_flags
    ADD CONSTRAINT feature_flags_pkey PRIMARY KEY (key);


--
-- Name: files files_pkey; Type: CONSTRAINT; Schema: platform; Owner: -
--

ALTER TABLE ONLY platform.files
    ADD CONSTRAINT files_pkey PRIMARY KEY (id);


--
-- Name: files files_storage_key_key; Type: CONSTRAINT; Schema: platform; Owner: -
--

ALTER TABLE ONLY platform.files
    ADD CONSTRAINT files_storage_key_key UNIQUE (storage_key);


--
-- Name: idempotency_keys idempotency_keys_pkey; Type: CONSTRAINT; Schema: platform; Owner: -
--

ALTER TABLE ONLY platform.idempotency_keys
    ADD CONSTRAINT idempotency_keys_pkey PRIMARY KEY (scope, key);


--
-- Name: outbox_events outbox_events_pkey; Type: CONSTRAINT; Schema: platform; Owner: -
--

ALTER TABLE ONLY platform.outbox_events
    ADD CONSTRAINT outbox_events_pkey PRIMARY KEY (id);


--
-- Name: public_holidays public_holidays_pkey; Type: CONSTRAINT; Schema: platform; Owner: -
--

ALTER TABLE ONLY platform.public_holidays
    ADD CONSTRAINT public_holidays_pkey PRIMARY KEY (date);


--
-- Name: pos_shifts pos_shifts_pkey; Type: CONSTRAINT; Schema: pos; Owner: -
--

ALTER TABLE ONLY pos.pos_shifts
    ADD CONSTRAINT pos_shifts_pkey PRIMARY KEY (id);


--
-- Name: pos_terminals pos_terminals_pkey; Type: CONSTRAINT; Schema: pos; Owner: -
--

ALTER TABLE ONLY pos.pos_terminals
    ADD CONSTRAINT pos_terminals_pkey PRIMARY KEY (id);


--
-- Name: pos_terminals pos_terminals_warehouse_id_label_key; Type: CONSTRAINT; Schema: pos; Owner: -
--

ALTER TABLE ONLY pos.pos_terminals
    ADD CONSTRAINT pos_terminals_warehouse_id_label_key UNIQUE (warehouse_id, label);


--
-- Name: goods_receipts goods_receipts_pkey; Type: CONSTRAINT; Schema: purchasing; Owner: -
--

ALTER TABLE ONLY purchasing.goods_receipts
    ADD CONSTRAINT goods_receipts_pkey PRIMARY KEY (id);


--
-- Name: purchase_order_lines purchase_order_lines_pkey; Type: CONSTRAINT; Schema: purchasing; Owner: -
--

ALTER TABLE ONLY purchasing.purchase_order_lines
    ADD CONSTRAINT purchase_order_lines_pkey PRIMARY KEY (id);


--
-- Name: purchase_orders purchase_orders_pkey; Type: CONSTRAINT; Schema: purchasing; Owner: -
--

ALTER TABLE ONLY purchasing.purchase_orders
    ADD CONSTRAINT purchase_orders_pkey PRIMARY KEY (id);


--
-- Name: purchase_orders purchase_orders_po_number_key; Type: CONSTRAINT; Schema: purchasing; Owner: -
--

ALTER TABLE ONLY purchasing.purchase_orders
    ADD CONSTRAINT purchase_orders_po_number_key UNIQUE (po_number);


--
-- Name: suppliers suppliers_pkey; Type: CONSTRAINT; Schema: purchasing; Owner: -
--

ALTER TABLE ONLY purchasing.suppliers
    ADD CONSTRAINT suppliers_pkey PRIMARY KEY (id);


--
-- Name: device_blocklist device_blocklist_imei_key; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.device_blocklist
    ADD CONSTRAINT device_blocklist_imei_key UNIQUE (imei);


--
-- Name: device_blocklist device_blocklist_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.device_blocklist
    ADD CONSTRAINT device_blocklist_pkey PRIMARY KEY (id);


--
-- Name: enforcement_actions enforcement_actions_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.enforcement_actions
    ADD CONSTRAINT enforcement_actions_pkey PRIMARY KEY (id);


--
-- Name: kyc_checks kyc_checks_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.kyc_checks
    ADD CONSTRAINT kyc_checks_pkey PRIMARY KEY (id);


--
-- Name: listing_reports listing_reports_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.listing_reports
    ADD CONSTRAINT listing_reports_pkey PRIMARY KEY (id);


--
-- Name: review_reports review_reports_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.review_reports
    ADD CONSTRAINT review_reports_pkey PRIMARY KEY (id);


--
-- Name: risk_cases risk_cases_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_cases
    ADD CONSTRAINT risk_cases_pkey PRIMARY KEY (id);


--
-- Name: risk_decisions risk_decisions_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_decisions
    ADD CONSTRAINT risk_decisions_pkey PRIMARY KEY (id, created_at);


--
-- Name: risk_decisions_default risk_decisions_default_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_decisions_default
    ADD CONSTRAINT risk_decisions_default_pkey PRIMARY KEY (id, created_at);


--
-- Name: risk_entities risk_entities_kind_value_hash_key; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_entities
    ADD CONSTRAINT risk_entities_kind_value_hash_key UNIQUE (kind, value_hash);


--
-- Name: risk_entities risk_entities_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_entities
    ADD CONSTRAINT risk_entities_pkey PRIMARY KEY (id);


--
-- Name: risk_links risk_links_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_links
    ADD CONSTRAINT risk_links_pkey PRIMARY KEY (entity_a, entity_b);


--
-- Name: risk_lists risk_lists_kind_value_hash_key; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_lists
    ADD CONSTRAINT risk_lists_kind_value_hash_key UNIQUE (kind, value_hash);


--
-- Name: risk_lists risk_lists_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_lists
    ADD CONSTRAINT risk_lists_pkey PRIMARY KEY (id);


--
-- Name: risk_rule_sets risk_rule_sets_checkpoint_version_key; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_rule_sets
    ADD CONSTRAINT risk_rule_sets_checkpoint_version_key UNIQUE (checkpoint, version);


--
-- Name: risk_rule_sets risk_rule_sets_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_rule_sets
    ADD CONSTRAINT risk_rule_sets_pkey PRIMARY KEY (id);


--
-- Name: risk_rules risk_rules_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_rules
    ADD CONSTRAINT risk_rules_pkey PRIMARY KEY (id);


--
-- Name: risk_rules risk_rules_rule_set_id_key_key; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_rules
    ADD CONSTRAINT risk_rules_rule_set_id_key_key UNIQUE (rule_set_id, key);


--
-- Name: seller_metrics_daily seller_metrics_daily_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.seller_metrics_daily
    ADD CONSTRAINT seller_metrics_daily_pkey PRIMARY KEY (seller_id, day);


--
-- Name: velocity_counters velocity_counters_pkey; Type: CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.velocity_counters
    ADD CONSTRAINT velocity_counters_pkey PRIMARY KEY (key, window_start);


--
-- Name: cart_items cart_items_cart_id_listing_id_key; Type: CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.cart_items
    ADD CONSTRAINT cart_items_cart_id_listing_id_key UNIQUE (cart_id, listing_id);


--
-- Name: cart_items cart_items_pkey; Type: CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.cart_items
    ADD CONSTRAINT cart_items_pkey PRIMARY KEY (id);


--
-- Name: carts carts_pkey; Type: CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.carts
    ADD CONSTRAINT carts_pkey PRIMARY KEY (id);


--
-- Name: carts carts_session_token_hash_key; Type: CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.carts
    ADD CONSTRAINT carts_session_token_hash_key UNIQUE (session_token_hash);


--
-- Name: fulfilments fulfilments_order_id_seller_id_key; Type: CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.fulfilments
    ADD CONSTRAINT fulfilments_order_id_seller_id_key UNIQUE (order_id, seller_id);


--
-- Name: fulfilments fulfilments_pkey; Type: CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.fulfilments
    ADD CONSTRAINT fulfilments_pkey PRIMARY KEY (id);


--
-- Name: order_lines order_lines_device_unit_id_key; Type: CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.order_lines
    ADD CONSTRAINT order_lines_device_unit_id_key UNIQUE (device_unit_id);


--
-- Name: order_lines order_lines_pkey; Type: CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.order_lines
    ADD CONSTRAINT order_lines_pkey PRIMARY KEY (id);


--
-- Name: order_status_history order_status_history_pkey; Type: CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.order_status_history
    ADD CONSTRAINT order_status_history_pkey PRIMARY KEY (id);


--
-- Name: orders orders_order_number_key; Type: CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.orders
    ADD CONSTRAINT orders_order_number_key UNIQUE (order_number);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: orders orders_quote_id_key; Type: CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.orders
    ADD CONSTRAINT orders_quote_id_key UNIQUE (quote_id);


--
-- Name: catalogue_gaps catalogue_gaps_pkey; Type: CONSTRAINT; Schema: search; Owner: -
--

ALTER TABLE ONLY search.catalogue_gaps
    ADD CONSTRAINT catalogue_gaps_pkey PRIMARY KEY (query_norm);


--
-- Name: search_pins search_pins_pkey; Type: CONSTRAINT; Schema: search; Owner: -
--

ALTER TABLE ONLY search.search_pins
    ADD CONSTRAINT search_pins_pkey PRIMARY KEY (id);


--
-- Name: search_queries search_queries_pkey; Type: CONSTRAINT; Schema: search; Owner: -
--

ALTER TABLE ONLY search.search_queries
    ADD CONSTRAINT search_queries_pkey PRIMARY KEY (id, created_at);


--
-- Name: search_queries_default search_queries_default_pkey; Type: CONSTRAINT; Schema: search; Owner: -
--

ALTER TABLE ONLY search.search_queries_default
    ADD CONSTRAINT search_queries_default_pkey PRIMARY KEY (id, created_at);


--
-- Name: search_redirects search_redirects_pkey; Type: CONSTRAINT; Schema: search; Owner: -
--

ALTER TABLE ONLY search.search_redirects
    ADD CONSTRAINT search_redirects_pkey PRIMARY KEY (query_norm);


--
-- Name: search_suggestions search_suggestions_pkey; Type: CONSTRAINT; Schema: search; Owner: -
--

ALTER TABLE ONLY search.search_suggestions
    ADD CONSTRAINT search_suggestions_pkey PRIMARY KEY (prefix, suggestion);


--
-- Name: search_synonyms search_synonyms_pkey; Type: CONSTRAINT; Schema: search; Owner: -
--

ALTER TABLE ONLY search.search_synonyms
    ADD CONSTRAINT search_synonyms_pkey PRIMARY KEY (id);


--
-- Name: kyc_documents kyc_documents_pkey; Type: CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.kyc_documents
    ADD CONSTRAINT kyc_documents_pkey PRIMARY KEY (id);


--
-- Name: seller_bank_accounts seller_bank_accounts_pkey; Type: CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.seller_bank_accounts
    ADD CONSTRAINT seller_bank_accounts_pkey PRIMARY KEY (id);


--
-- Name: seller_kyc_submissions seller_kyc_submissions_pkey; Type: CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.seller_kyc_submissions
    ADD CONSTRAINT seller_kyc_submissions_pkey PRIMARY KEY (id);


--
-- Name: seller_members seller_members_pkey; Type: CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.seller_members
    ADD CONSTRAINT seller_members_pkey PRIMARY KEY (seller_id, user_id);


--
-- Name: sellers sellers_pkey; Type: CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.sellers
    ADD CONSTRAINT sellers_pkey PRIMARY KEY (id);


--
-- Name: sellers sellers_slug_key; Type: CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.sellers
    ADD CONSTRAINT sellers_slug_key UNIQUE (slug);


--
-- Name: message_attachments message_attachments_pkey; Type: CONSTRAINT; Schema: support; Owner: -
--

ALTER TABLE ONLY support.message_attachments
    ADD CONSTRAINT message_attachments_pkey PRIMARY KEY (message_id, file_id);


--
-- Name: support_tickets support_tickets_pkey; Type: CONSTRAINT; Schema: support; Owner: -
--

ALTER TABLE ONLY support.support_tickets
    ADD CONSTRAINT support_tickets_pkey PRIMARY KEY (id);


--
-- Name: support_tickets support_tickets_ticket_number_key; Type: CONSTRAINT; Schema: support; Owner: -
--

ALTER TABLE ONLY support.support_tickets
    ADD CONSTRAINT support_tickets_ticket_number_key UNIQUE (ticket_number);


--
-- Name: ticket_messages ticket_messages_pkey; Type: CONSTRAINT; Schema: support; Owner: -
--

ALTER TABLE ONLY support.ticket_messages
    ADD CONSTRAINT ticket_messages_pkey PRIMARY KEY (id);


--
-- Name: repair_jobs_device_unit_id_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX repair_jobs_device_unit_id_idx ON aftersales.repair_jobs USING btree (device_unit_id);


--
-- Name: repair_jobs_technician_id_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX repair_jobs_technician_id_idx ON aftersales.repair_jobs USING btree (technician_id);


--
-- Name: repair_jobs_warranty_claim_id_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX repair_jobs_warranty_claim_id_idx ON aftersales.repair_jobs USING btree (warranty_claim_id);


--
-- Name: return_inspections_device_unit_id_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX return_inspections_device_unit_id_idx ON aftersales.return_inspections USING btree (device_unit_id);


--
-- Name: return_inspections_inspected_by_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX return_inspections_inspected_by_idx ON aftersales.return_inspections USING btree (inspected_by);


--
-- Name: return_inspections_return_item_id_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX return_inspections_return_item_id_idx ON aftersales.return_inspections USING btree (return_item_id);


--
-- Name: return_items_order_line_id_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX return_items_order_line_id_idx ON aftersales.return_items USING btree (order_line_id);


--
-- Name: return_requests_decided_by_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX return_requests_decided_by_idx ON aftersales.return_requests USING btree (decided_by);


--
-- Name: return_requests_order_id_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX return_requests_order_id_idx ON aftersales.return_requests USING btree (order_id);


--
-- Name: return_requests_requested_by_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX return_requests_requested_by_idx ON aftersales.return_requests USING btree (requested_by);


--
-- Name: trade_ins_encryption_key_id_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX trade_ins_encryption_key_id_idx ON aftersales.trade_ins USING btree (encryption_key_id);


--
-- Name: trade_ins_inspected_by_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX trade_ins_inspected_by_idx ON aftersales.trade_ins USING btree (inspected_by);


--
-- Name: trade_ins_user_id_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX trade_ins_user_id_idx ON aftersales.trade_ins USING btree (user_id);


--
-- Name: warranty_claims_customer_user_id_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX warranty_claims_customer_user_id_idx ON aftersales.warranty_claims USING btree (customer_user_id);


--
-- Name: warranty_claims_device_unit_id_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX warranty_claims_device_unit_id_idx ON aftersales.warranty_claims USING btree (device_unit_id);


--
-- Name: warranty_claims_order_line_id_idx; Type: INDEX; Schema: aftersales; Owner: -
--

CREATE INDEX warranty_claims_order_line_id_idx ON aftersales.warranty_claims USING btree (order_line_id);


--
-- Name: business_members_user_idx; Type: INDEX; Schema: b2b; Owner: -
--

CREATE INDEX business_members_user_idx ON b2b.business_members USING btree (user_id);


--
-- Name: businesses_account_manager_id_idx; Type: INDEX; Schema: b2b; Owner: -
--

CREATE INDEX businesses_account_manager_id_idx ON b2b.businesses USING btree (account_manager_id);


--
-- Name: credit_accounts_approved_by_idx; Type: INDEX; Schema: b2b; Owner: -
--

CREATE INDEX credit_accounts_approved_by_idx ON b2b.credit_accounts USING btree (approved_by);


--
-- Name: quote_lines_listing_id_idx; Type: INDEX; Schema: b2b; Owner: -
--

CREATE INDEX quote_lines_listing_id_idx ON b2b.quote_lines USING btree (listing_id);


--
-- Name: quote_lines_quote_id_idx; Type: INDEX; Schema: b2b; Owner: -
--

CREATE INDEX quote_lines_quote_id_idx ON b2b.quote_lines USING btree (quote_id);


--
-- Name: quotes_assigned_to_idx; Type: INDEX; Schema: b2b; Owner: -
--

CREATE INDEX quotes_assigned_to_idx ON b2b.quotes USING btree (assigned_to);


--
-- Name: quotes_business_id_idx; Type: INDEX; Schema: b2b; Owner: -
--

CREATE INDEX quotes_business_id_idx ON b2b.quotes USING btree (business_id);


--
-- Name: quotes_delivery_state_idx; Type: INDEX; Schema: b2b; Owner: -
--

CREATE INDEX quotes_delivery_state_idx ON b2b.quotes USING btree (delivery_state);


--
-- Name: quotes_requested_by_idx; Type: INDEX; Schema: b2b; Owner: -
--

CREATE INDEX quotes_requested_by_idx ON b2b.quotes USING btree (requested_by);


--
-- Name: car_documents_car_listing_id_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX car_documents_car_listing_id_idx ON cars.car_documents USING btree (car_listing_id);


--
-- Name: car_documents_file_id_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX car_documents_file_id_idx ON cars.car_documents USING btree (file_id);


--
-- Name: car_documents_verified_by_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX car_documents_verified_by_idx ON cars.car_documents USING btree (verified_by);


--
-- Name: car_images_car_listing_id_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX car_images_car_listing_id_idx ON cars.car_images USING btree (car_listing_id);


--
-- Name: car_images_file_id_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX car_images_file_id_idx ON cars.car_images USING btree (file_id);


--
-- Name: car_inspections_car_listing_id_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX car_inspections_car_listing_id_idx ON cars.car_inspections USING btree (car_listing_id);


--
-- Name: car_inspections_inspector_id_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX car_inspections_inspector_id_idx ON cars.car_inspections USING btree (inspector_id);


--
-- Name: car_inspections_report_file_id_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX car_inspections_report_file_id_idx ON cars.car_inspections USING btree (report_file_id);


--
-- Name: car_listings_seller_id_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX car_listings_seller_id_idx ON cars.car_listings USING btree (seller_id);


--
-- Name: car_listings_state_code_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX car_listings_state_code_idx ON cars.car_listings USING btree (state_code);


--
-- Name: financing_enquiries_car_listing_id_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX financing_enquiries_car_listing_id_idx ON cars.financing_enquiries USING btree (car_listing_id);


--
-- Name: financing_enquiries_customer_user_id_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX financing_enquiries_customer_user_id_idx ON cars.financing_enquiries USING btree (customer_user_id);


--
-- Name: viewing_bookings_car_listing_id_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX viewing_bookings_car_listing_id_idx ON cars.viewing_bookings USING btree (car_listing_id);


--
-- Name: viewing_bookings_customer_user_id_idx; Type: INDEX; Schema: cars; Owner: -
--

CREATE INDEX viewing_bookings_customer_user_id_idx ON cars.viewing_bookings USING btree (customer_user_id);


--
-- Name: brands_logo_file_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX brands_logo_file_id_idx ON catalog.brands USING btree (logo_file_id);


--
-- Name: categories_parent_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX categories_parent_id_idx ON catalog.categories USING btree (parent_id);


--
-- Name: image_hashes_phash_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX image_hashes_phash_idx ON catalog.image_hashes USING btree (phash);


--
-- Name: listing_price_history_changed_by_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX listing_price_history_changed_by_idx ON catalog.listing_price_history USING btree (changed_by);


--
-- Name: listing_price_history_listing_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX listing_price_history_listing_idx ON catalog.listing_price_history USING btree (listing_id, changed_at DESC);


--
-- Name: listing_reviews_listing_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX listing_reviews_listing_id_idx ON catalog.listing_reviews USING btree (listing_id);


--
-- Name: listing_reviews_reviewer_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX listing_reviews_reviewer_id_idx ON catalog.listing_reviews USING btree (reviewer_id);


--
-- Name: listings_moderation_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX listings_moderation_idx ON catalog.listings USING btree (risk_score DESC, created_at) WHERE (status = 'in_review'::text);


--
-- Name: listings_variant_status_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX listings_variant_status_idx ON catalog.listings USING btree (variant_id, status);


--
-- Name: product_attribute_values_attribute_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_attribute_values_attribute_id_idx ON catalog.product_attribute_values USING btree (attribute_id);


--
-- Name: product_compatibility_accessory_product_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_compatibility_accessory_product_id_idx ON catalog.product_compatibility USING btree (accessory_product_id);


--
-- Name: product_compatibility_device_product_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_compatibility_device_product_id_idx ON catalog.product_compatibility USING btree (device_product_id);


--
-- Name: product_images_file_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_images_file_id_idx ON catalog.product_images USING btree (file_id);


--
-- Name: product_images_product_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_images_product_id_idx ON catalog.product_images USING btree (product_id);


--
-- Name: product_images_variant_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_images_variant_id_idx ON catalog.product_images USING btree (variant_id);


--
-- Name: product_offer_summary_attrs_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_offer_summary_attrs_idx ON catalog.product_offer_summary USING gin (attrs jsonb_path_ops);


--
-- Name: product_offer_summary_best_listing_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_offer_summary_best_listing_id_idx ON catalog.product_offer_summary USING btree (best_listing_id);


--
-- Name: product_offer_summary_brand_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_offer_summary_brand_id_idx ON catalog.product_offer_summary USING btree (brand_id);


--
-- Name: product_offer_summary_browse_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_offer_summary_browse_idx ON catalog.product_offer_summary USING btree (channel, in_stock, sales_30d DESC);


--
-- Name: product_offer_summary_categories_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_offer_summary_categories_idx ON catalog.product_offer_summary USING gin (category_ids);


--
-- Name: product_offer_summary_facets_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_offer_summary_facets_idx ON catalog.product_offer_summary USING gin (facet_keys);


--
-- Name: product_offer_summary_search_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_offer_summary_search_idx ON catalog.product_offer_summary USING gin (search_vector);


--
-- Name: product_offer_summary_title_trgm_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_offer_summary_title_trgm_idx ON catalog.product_offer_summary USING gin (title public.gin_trgm_ops);


--
-- Name: product_reviews_one_per_user_product; Type: INDEX; Schema: catalog; Owner: -
--

CREATE UNIQUE INDEX product_reviews_one_per_user_product ON catalog.product_reviews USING btree (user_id, product_id);


--
-- Name: product_reviews_product_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_reviews_product_id_idx ON catalog.product_reviews USING btree (product_id);


--
-- Name: product_suggestions_product_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_suggestions_product_id_idx ON catalog.product_suggestions USING btree (product_id);


--
-- Name: product_suggestions_reviewed_by_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_suggestions_reviewed_by_idx ON catalog.product_suggestions USING btree (reviewed_by);


--
-- Name: product_suggestions_seller_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_suggestions_seller_id_idx ON catalog.product_suggestions USING btree (seller_id);


--
-- Name: product_variants_gtin_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_variants_gtin_idx ON catalog.product_variants USING btree (gtin) WHERE (gtin IS NOT NULL);


--
-- Name: product_variants_product_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX product_variants_product_id_idx ON catalog.product_variants USING btree (product_id);


--
-- Name: products_brand_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX products_brand_id_idx ON catalog.products USING btree (brand_id);


--
-- Name: products_category_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX products_category_id_idx ON catalog.products USING btree (category_id);


--
-- Name: products_created_by_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX products_created_by_idx ON catalog.products USING btree (created_by);


--
-- Name: products_name_trgm_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX products_name_trgm_idx ON catalog.products USING gin (name public.gin_trgm_ops);


--
-- Name: products_search_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX products_search_idx ON catalog.products USING gin (search_vector);


--
-- Name: review_images_file_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX review_images_file_id_idx ON catalog.review_images USING btree (file_id);


--
-- Name: saved_items_product_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX saved_items_product_id_idx ON catalog.saved_items USING btree (product_id);


--
-- Name: seller_ratings_seller_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX seller_ratings_seller_id_idx ON catalog.seller_ratings USING btree (seller_id);


--
-- Name: seller_ratings_user_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX seller_ratings_user_id_idx ON catalog.seller_ratings USING btree (user_id);


--
-- Name: stock_alerts_product_id_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX stock_alerts_product_id_idx ON catalog.stock_alerts USING btree (product_id);


--
-- Name: banners_image_file_id_idx; Type: INDEX; Schema: content; Owner: -
--

CREATE INDEX banners_image_file_id_idx ON content.banners USING btree (image_file_id);


--
-- Name: cms_pages_approved_by_idx; Type: INDEX; Schema: content; Owner: -
--

CREATE INDEX cms_pages_approved_by_idx ON content.cms_pages USING btree (approved_by);


--
-- Name: cms_pages_updated_by_idx; Type: INDEX; Schema: content; Owner: -
--

CREATE INDEX cms_pages_updated_by_idx ON content.cms_pages USING btree (updated_by);


--
-- Name: accounting_periods_closed_by_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX accounting_periods_closed_by_idx ON finance.accounting_periods USING btree (closed_by);


--
-- Name: commission_rules_category_id_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX commission_rules_category_id_idx ON finance.commission_rules USING btree (category_id);


--
-- Name: commission_rules_created_by_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX commission_rules_created_by_idx ON finance.commission_rules USING btree (created_by);


--
-- Name: commission_rules_seller_id_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX commission_rules_seller_id_idx ON finance.commission_rules USING btree (seller_id);


--
-- Name: invoices_business_id_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX invoices_business_id_idx ON finance.invoices USING btree (business_id);


--
-- Name: invoices_open_due_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX invoices_open_due_idx ON finance.invoices USING btree (due_at) WHERE (status = ANY (ARRAY['issued'::text, 'partially_paid'::text, 'overdue'::text]));


--
-- Name: invoices_pdf_file_id_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX invoices_pdf_file_id_idx ON finance.invoices USING btree (pdf_file_id);


--
-- Name: ledger_entries_account_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX ledger_entries_account_idx ON finance.ledger_entries USING btree (account_id, created_at);


--
-- Name: ledger_entries_journal_id_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX ledger_entries_journal_id_idx ON finance.ledger_entries USING btree (journal_id);


--
-- Name: ledger_journals_posted_by_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX ledger_journals_posted_by_idx ON finance.ledger_journals USING btree (posted_by);


--
-- Name: ledger_journals_reference_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX ledger_journals_reference_idx ON finance.ledger_journals USING btree (reference_type, reference_id);


--
-- Name: payout_batches_approved_by_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX payout_batches_approved_by_idx ON finance.payout_batches USING btree (approved_by);


--
-- Name: payout_batches_prepared_by_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX payout_batches_prepared_by_idx ON finance.payout_batches USING btree (prepared_by);


--
-- Name: payout_holds_active_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX payout_holds_active_idx ON finance.payout_holds USING btree (seller_id) WHERE (released_at IS NULL);


--
-- Name: payout_holds_created_by_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX payout_holds_created_by_idx ON finance.payout_holds USING btree (created_by);


--
-- Name: payouts_bank_account_id_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX payouts_bank_account_id_idx ON finance.payouts USING btree (bank_account_id);


--
-- Name: payouts_batch_id_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX payouts_batch_id_idx ON finance.payouts USING btree (batch_id);


--
-- Name: payouts_seller_id_idx; Type: INDEX; Schema: finance; Owner: -
--

CREATE INDEX payouts_seller_id_idx ON finance.payouts USING btree (seller_id);


--
-- Name: manifests_carrier_id_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX manifests_carrier_id_idx ON fulfilment.manifests USING btree (carrier_id);


--
-- Name: manifests_created_by_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX manifests_created_by_idx ON fulfilment.manifests USING btree (created_by);


--
-- Name: manifests_rider_id_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX manifests_rider_id_idx ON fulfilment.manifests USING btree (rider_id);


--
-- Name: manifests_warehouse_id_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX manifests_warehouse_id_idx ON fulfilment.manifests USING btree (warehouse_id);


--
-- Name: pack_records_fulfilment_id_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX pack_records_fulfilment_id_idx ON fulfilment.pack_records USING btree (fulfilment_id);


--
-- Name: pack_records_packed_by_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX pack_records_packed_by_idx ON fulfilment.pack_records USING btree (packed_by);


--
-- Name: parcels_fulfilment_id_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX parcels_fulfilment_id_idx ON fulfilment.parcels USING btree (fulfilment_id);


--
-- Name: pick_list_items_bin_id_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX pick_list_items_bin_id_idx ON fulfilment.pick_list_items USING btree (bin_id);


--
-- Name: pick_list_items_device_unit_id_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX pick_list_items_device_unit_id_idx ON fulfilment.pick_list_items USING btree (device_unit_id);


--
-- Name: pick_list_items_fulfilment_id_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX pick_list_items_fulfilment_id_idx ON fulfilment.pick_list_items USING btree (fulfilment_id);


--
-- Name: pick_list_items_order_line_id_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX pick_list_items_order_line_id_idx ON fulfilment.pick_list_items USING btree (order_line_id);


--
-- Name: pick_lists_picker_id_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX pick_lists_picker_id_idx ON fulfilment.pick_lists USING btree (picker_id);


--
-- Name: pick_lists_warehouse_id_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX pick_lists_warehouse_id_idx ON fulfilment.pick_lists USING btree (warehouse_id);


--
-- Name: seller_sla_events_seller_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX seller_sla_events_seller_idx ON fulfilment.seller_sla_events USING btree (seller_id, created_at DESC);


--
-- Name: shipments_fulfilment_id_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX shipments_fulfilment_id_idx ON fulfilment.shipments USING btree (fulfilment_id);


--
-- Name: shipments_label_file_id_idx; Type: INDEX; Schema: fulfilment; Owner: -
--

CREATE INDEX shipments_label_file_id_idx ON fulfilment.shipments USING btree (label_file_id);


--
-- Name: addresses_one_default_per_business; Type: INDEX; Schema: identity; Owner: -
--

CREATE UNIQUE INDEX addresses_one_default_per_business ON identity.addresses USING btree (business_id) WHERE (is_default AND (business_id IS NOT NULL));


--
-- Name: addresses_one_default_per_user; Type: INDEX; Schema: identity; Owner: -
--

CREATE UNIQUE INDEX addresses_one_default_per_user ON identity.addresses USING btree (user_id) WHERE (is_default AND (user_id IS NOT NULL));


--
-- Name: addresses_state_code_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX addresses_state_code_idx ON identity.addresses USING btree (state_code);


--
-- Name: auth_events_user_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX auth_events_user_idx ON ONLY identity.auth_events USING btree (user_id, created_at DESC);


--
-- Name: auth_events_default_user_id_created_at_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX auth_events_default_user_id_created_at_idx ON identity.auth_events_default USING btree (user_id, created_at DESC);


--
-- Name: consents_anon_purpose_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX consents_anon_purpose_idx ON identity.consents USING btree (anonymous_id, purpose, recorded_at DESC) WHERE (anonymous_id IS NOT NULL);


--
-- Name: consents_user_purpose_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX consents_user_purpose_idx ON identity.consents USING btree (user_id, purpose, recorded_at DESC) WHERE (user_id IS NOT NULL);


--
-- Name: privacy_requests_file_id_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX privacy_requests_file_id_idx ON identity.privacy_requests USING btree (file_id);


--
-- Name: privacy_requests_user_id_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX privacy_requests_user_id_idx ON identity.privacy_requests USING btree (user_id);


--
-- Name: role_permissions_permission_id_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX role_permissions_permission_id_idx ON identity.role_permissions USING btree (permission_id);


--
-- Name: staff_invites_created_by_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX staff_invites_created_by_idx ON identity.staff_invites USING btree (created_by);


--
-- Name: staff_invites_staff_user_id_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX staff_invites_staff_user_id_idx ON identity.staff_invites USING btree (staff_user_id);


--
-- Name: staff_roles_granted_by_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX staff_roles_granted_by_idx ON identity.staff_roles USING btree (granted_by);


--
-- Name: staff_roles_role_id_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX staff_roles_role_id_idx ON identity.staff_roles USING btree (role_id);


--
-- Name: user_mfa_factors_encryption_key_id_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX user_mfa_factors_encryption_key_id_idx ON identity.user_mfa_factors USING btree (encryption_key_id);


--
-- Name: user_mfa_factors_one_totp; Type: INDEX; Schema: identity; Owner: -
--

CREATE UNIQUE INDEX user_mfa_factors_one_totp ON identity.user_mfa_factors USING btree (user_id) WHERE (kind = 'totp'::text);


--
-- Name: user_sessions_active_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX user_sessions_active_idx ON identity.user_sessions USING btree (user_id, last_used_at DESC) WHERE (revoked_at IS NULL);


--
-- Name: user_sessions_family_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX user_sessions_family_idx ON identity.user_sessions USING btree (family_id);


--
-- Name: user_sessions_replaced_by_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX user_sessions_replaced_by_idx ON identity.user_sessions USING btree (replaced_by);


--
-- Name: verification_codes_destination_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX verification_codes_destination_idx ON identity.verification_codes USING btree (destination, created_at DESC);


--
-- Name: verification_codes_user_id_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX verification_codes_user_id_idx ON identity.verification_codes USING btree (user_id);


--
-- Name: webauthn_credentials_user_id_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX webauthn_credentials_user_id_idx ON identity.webauthn_credentials USING btree (user_id);


--
-- Name: device_units_owner_seller_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX device_units_owner_seller_id_idx ON inventory.device_units USING btree (owner_seller_id);


--
-- Name: device_units_serial_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE UNIQUE INDEX device_units_serial_idx ON inventory.device_units USING btree (variant_id, serial_number) WHERE (serial_number IS NOT NULL);


--
-- Name: device_units_warehouse_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX device_units_warehouse_id_idx ON inventory.device_units USING btree (warehouse_id);


--
-- Name: inventory_costs_owner_seller_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX inventory_costs_owner_seller_id_idx ON inventory.inventory_costs USING btree (owner_seller_id);


--
-- Name: inventory_costs_warehouse_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX inventory_costs_warehouse_id_idx ON inventory.inventory_costs USING btree (warehouse_id);


--
-- Name: inventory_count_lines_bin_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX inventory_count_lines_bin_id_idx ON inventory.inventory_count_lines USING btree (bin_id);


--
-- Name: inventory_count_lines_owner_seller_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX inventory_count_lines_owner_seller_id_idx ON inventory.inventory_count_lines USING btree (owner_seller_id);


--
-- Name: inventory_count_lines_variant_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX inventory_count_lines_variant_id_idx ON inventory.inventory_count_lines USING btree (variant_id);


--
-- Name: inventory_counts_approved_by_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX inventory_counts_approved_by_idx ON inventory.inventory_counts USING btree (approved_by);


--
-- Name: inventory_counts_counted_by_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX inventory_counts_counted_by_idx ON inventory.inventory_counts USING btree (counted_by);


--
-- Name: inventory_counts_warehouse_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX inventory_counts_warehouse_id_idx ON inventory.inventory_counts USING btree (warehouse_id);


--
-- Name: inventory_levels_bin_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX inventory_levels_bin_id_idx ON inventory.inventory_levels USING btree (bin_id);


--
-- Name: inventory_levels_owner_seller_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX inventory_levels_owner_seller_id_idx ON inventory.inventory_levels USING btree (owner_seller_id);


--
-- Name: inventory_levels_variant_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX inventory_levels_variant_id_idx ON inventory.inventory_levels USING btree (variant_id);


--
-- Name: stock_movements_created_by_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_movements_created_by_idx ON inventory.stock_movements USING btree (created_by);


--
-- Name: stock_movements_device_unit_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_movements_device_unit_id_idx ON inventory.stock_movements USING btree (device_unit_id);


--
-- Name: stock_movements_owner_seller_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_movements_owner_seller_id_idx ON inventory.stock_movements USING btree (owner_seller_id);


--
-- Name: stock_movements_variant_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_movements_variant_id_idx ON inventory.stock_movements USING btree (variant_id);


--
-- Name: stock_movements_warehouse_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_movements_warehouse_id_idx ON inventory.stock_movements USING btree (warehouse_id);


--
-- Name: stock_reservations_expiry_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_reservations_expiry_idx ON inventory.stock_reservations USING btree (expires_at) WHERE (status = 'reserved'::text);


--
-- Name: stock_reservations_listing_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_reservations_listing_id_idx ON inventory.stock_reservations USING btree (listing_id);


--
-- Name: stock_reservations_order_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_reservations_order_id_idx ON inventory.stock_reservations USING btree (order_id);


--
-- Name: stock_reservations_order_line_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_reservations_order_line_id_idx ON inventory.stock_reservations USING btree (order_line_id);


--
-- Name: stock_reservations_owner_seller_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_reservations_owner_seller_id_idx ON inventory.stock_reservations USING btree (owner_seller_id);


--
-- Name: stock_reservations_variant_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_reservations_variant_id_idx ON inventory.stock_reservations USING btree (variant_id);


--
-- Name: stock_reservations_warehouse_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_reservations_warehouse_id_idx ON inventory.stock_reservations USING btree (warehouse_id);


--
-- Name: stock_transfer_lines_owner_seller_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_transfer_lines_owner_seller_id_idx ON inventory.stock_transfer_lines USING btree (owner_seller_id);


--
-- Name: stock_transfer_lines_transfer_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_transfer_lines_transfer_id_idx ON inventory.stock_transfer_lines USING btree (transfer_id);


--
-- Name: stock_transfer_lines_variant_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_transfer_lines_variant_id_idx ON inventory.stock_transfer_lines USING btree (variant_id);


--
-- Name: stock_transfers_created_by_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_transfers_created_by_idx ON inventory.stock_transfers USING btree (created_by);


--
-- Name: stock_transfers_from_warehouse_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_transfers_from_warehouse_id_idx ON inventory.stock_transfers USING btree (from_warehouse_id);


--
-- Name: stock_transfers_to_warehouse_id_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX stock_transfers_to_warehouse_id_idx ON inventory.stock_transfers USING btree (to_warehouse_id);


--
-- Name: warehouses_state_code_idx; Type: INDEX; Schema: inventory; Owner: -
--

CREATE INDEX warehouses_state_code_idx ON inventory.warehouses USING btree (state_code);


--
-- Name: delivery_events_created_by_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX delivery_events_created_by_idx ON logistics.delivery_events USING btree (created_by);


--
-- Name: delivery_events_delivery_job_id_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX delivery_events_delivery_job_id_idx ON logistics.delivery_events USING btree (delivery_job_id);


--
-- Name: delivery_events_file_id_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX delivery_events_file_id_idx ON logistics.delivery_events USING btree (file_id);


--
-- Name: delivery_jobs_assigned_by_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX delivery_jobs_assigned_by_idx ON logistics.delivery_jobs USING btree (assigned_by);


--
-- Name: delivery_jobs_fulfilment_id_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX delivery_jobs_fulfilment_id_idx ON logistics.delivery_jobs USING btree (fulfilment_id);


--
-- Name: delivery_jobs_open_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX delivery_jobs_open_idx ON logistics.delivery_jobs USING btree (zone_id, status) WHERE (status = ANY (ARRAY['unassigned'::text, 'assigned'::text]));


--
-- Name: delivery_jobs_pickup_warehouse_id_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX delivery_jobs_pickup_warehouse_id_idx ON logistics.delivery_jobs USING btree (pickup_warehouse_id);


--
-- Name: delivery_jobs_proof_file_id_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX delivery_jobs_proof_file_id_idx ON logistics.delivery_jobs USING btree (proof_file_id);


--
-- Name: delivery_jobs_return_request_id_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX delivery_jobs_return_request_id_idx ON logistics.delivery_jobs USING btree (return_request_id);


--
-- Name: delivery_jobs_rider_status_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX delivery_jobs_rider_status_idx ON logistics.delivery_jobs USING btree (rider_id, status);


--
-- Name: delivery_zones_state_code_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX delivery_zones_state_code_idx ON logistics.delivery_zones USING btree (state_code);


--
-- Name: rider_devices_approved_by_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX rider_devices_approved_by_idx ON logistics.rider_devices USING btree (approved_by);


--
-- Name: rider_locations_job_id_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX rider_locations_job_id_idx ON logistics.rider_locations USING btree (job_id);


--
-- Name: rider_shifts_device_session_id_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX rider_shifts_device_session_id_idx ON logistics.rider_shifts USING btree (device_session_id);


--
-- Name: rider_shifts_one_open; Type: INDEX; Schema: logistics; Owner: -
--

CREATE UNIQUE INDEX rider_shifts_one_open ON logistics.rider_shifts USING btree (rider_id) WHERE (ended_at IS NULL);


--
-- Name: rider_zones_zone_id_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX rider_zones_zone_id_idx ON logistics.rider_zones USING btree (zone_id);


--
-- Name: riders_home_warehouse_id_idx; Type: INDEX; Schema: logistics; Owner: -
--

CREATE INDEX riders_home_warehouse_id_idx ON logistics.riders USING btree (home_warehouse_id);


--
-- Name: campaign_recipients_user_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX campaign_recipients_user_id_idx ON marketing.campaign_recipients USING btree (user_id);


--
-- Name: campaign_recipients_variant_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX campaign_recipients_variant_id_idx ON marketing.campaign_recipients USING btree (variant_id);


--
-- Name: campaign_variants_template_version_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX campaign_variants_template_version_id_idx ON marketing.campaign_variants USING btree (template_version_id);


--
-- Name: campaigns_approved_by_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX campaigns_approved_by_idx ON marketing.campaigns USING btree (approved_by);


--
-- Name: campaigns_created_by_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX campaigns_created_by_idx ON marketing.campaigns USING btree (created_by);


--
-- Name: campaigns_segment_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX campaigns_segment_id_idx ON marketing.campaigns USING btree (segment_id);


--
-- Name: coupon_codes_coupon_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX coupon_codes_coupon_id_idx ON marketing.coupon_codes USING btree (coupon_id);


--
-- Name: coupon_codes_issued_to_user_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX coupon_codes_issued_to_user_id_idx ON marketing.coupon_codes USING btree (issued_to_user_id);


--
-- Name: coupon_codes_notification_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX coupon_codes_notification_id_idx ON marketing.coupon_codes USING btree (notification_id);


--
-- Name: coupon_redemptions_coupon_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX coupon_redemptions_coupon_id_idx ON marketing.coupon_redemptions USING btree (coupon_id);


--
-- Name: coupon_redemptions_user_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX coupon_redemptions_user_id_idx ON marketing.coupon_redemptions USING btree (user_id);


--
-- Name: coupons_promotion_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX coupons_promotion_id_idx ON marketing.coupons USING btree (promotion_id);


--
-- Name: flash_claims_order_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX flash_claims_order_id_idx ON marketing.flash_claims USING btree (order_id);


--
-- Name: flash_claims_user_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX flash_claims_user_id_idx ON marketing.flash_claims USING btree (user_id);


--
-- Name: flash_claims_user_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX flash_claims_user_idx ON marketing.flash_claims USING btree (promotion_id, listing_id, user_id);


--
-- Name: journey_enrollments_due_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX journey_enrollments_due_idx ON marketing.journey_enrollments USING btree (next_run_at) WHERE (status = 'active'::text);


--
-- Name: journey_enrollments_one_active; Type: INDEX; Schema: marketing; Owner: -
--

CREATE UNIQUE INDEX journey_enrollments_one_active ON marketing.journey_enrollments USING btree (journey_id, user_id) WHERE (status = 'active'::text);


--
-- Name: journey_enrollments_user_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX journey_enrollments_user_id_idx ON marketing.journey_enrollments USING btree (user_id);


--
-- Name: journeys_created_by_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX journeys_created_by_idx ON marketing.journeys USING btree (created_by);


--
-- Name: promotion_listing_prices_listing_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX promotion_listing_prices_listing_id_idx ON marketing.promotion_listing_prices USING btree (listing_id);


--
-- Name: promotion_targets_category_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX promotion_targets_category_id_idx ON marketing.promotion_targets USING btree (category_id);


--
-- Name: promotion_targets_listing_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX promotion_targets_listing_id_idx ON marketing.promotion_targets USING btree (listing_id);


--
-- Name: promotion_targets_promotion_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX promotion_targets_promotion_id_idx ON marketing.promotion_targets USING btree (promotion_id);


--
-- Name: promotions_created_by_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX promotions_created_by_idx ON marketing.promotions USING btree (created_by);


--
-- Name: segments_created_by_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX segments_created_by_idx ON marketing.segments USING btree (created_by);


--
-- Name: tracked_links_campaign_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX tracked_links_campaign_id_idx ON marketing.tracked_links USING btree (campaign_id);


--
-- Name: tracked_links_notification_id_idx; Type: INDEX; Schema: marketing; Owner: -
--

CREATE INDEX tracked_links_notification_id_idx ON marketing.tracked_links USING btree (notification_id);


--
-- Name: message_events_notification_idx; Type: INDEX; Schema: messaging; Owner: -
--

CREATE INDEX message_events_notification_idx ON ONLY messaging.message_events USING btree (notification_id);


--
-- Name: message_events_default_notification_id_idx; Type: INDEX; Schema: messaging; Owner: -
--

CREATE INDEX message_events_default_notification_id_idx ON messaging.message_events_default USING btree (notification_id);


--
-- Name: notifications_campaign_id_idx; Type: INDEX; Schema: messaging; Owner: -
--

CREATE INDEX notifications_campaign_id_idx ON messaging.notifications USING btree (campaign_id);


--
-- Name: notifications_inbox_idx; Type: INDEX; Schema: messaging; Owner: -
--

CREATE INDEX notifications_inbox_idx ON messaging.notifications USING btree (user_id, created_at DESC) WHERE (channel = 'in_app'::text);


--
-- Name: notifications_journey_enrollment_id_idx; Type: INDEX; Schema: messaging; Owner: -
--

CREATE INDEX notifications_journey_enrollment_id_idx ON messaging.notifications USING btree (journey_enrollment_id);


--
-- Name: notifications_provider_idx; Type: INDEX; Schema: messaging; Owner: -
--

CREATE INDEX notifications_provider_idx ON messaging.notifications USING btree (provider, provider_message_id) WHERE (provider_message_id IS NOT NULL);


--
-- Name: notifications_template_version_id_idx; Type: INDEX; Schema: messaging; Owner: -
--

CREATE INDEX notifications_template_version_id_idx ON messaging.notifications USING btree (template_version_id);


--
-- Name: notifications_unread_idx; Type: INDEX; Schema: messaging; Owner: -
--

CREATE INDEX notifications_unread_idx ON messaging.notifications USING btree (user_id) WHERE ((channel = 'in_app'::text) AND (read_at IS NULL));


--
-- Name: push_tokens_user_idx; Type: INDEX; Schema: messaging; Owner: -
--

CREATE INDEX push_tokens_user_idx ON messaging.push_tokens USING btree (user_id) WHERE (revoked_at IS NULL);


--
-- Name: template_versions_created_by_idx; Type: INDEX; Schema: messaging; Owner: -
--

CREATE INDEX template_versions_created_by_idx ON messaging.template_versions USING btree (created_by);


--
-- Name: dispute_evidence_dispute_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX dispute_evidence_dispute_id_idx ON payments.dispute_evidence USING btree (dispute_id);


--
-- Name: dispute_evidence_file_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX dispute_evidence_file_id_idx ON payments.dispute_evidence USING btree (file_id);


--
-- Name: disputes_decided_by_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX disputes_decided_by_idx ON payments.disputes USING btree (decided_by);


--
-- Name: disputes_payment_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX disputes_payment_id_idx ON payments.disputes USING btree (payment_id);


--
-- Name: payment_events_payment_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX payment_events_payment_id_idx ON payments.payment_events USING btree (payment_id);


--
-- Name: payment_events_payout_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX payment_events_payout_id_idx ON payments.payment_events USING btree (payout_id);


--
-- Name: payment_events_refund_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX payment_events_refund_id_idx ON payments.payment_events USING btree (refund_id);


--
-- Name: payments_invoice_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX payments_invoice_id_idx ON payments.payments USING btree (invoice_id);


--
-- Name: payments_order_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX payments_order_id_idx ON payments.payments USING btree (order_id);


--
-- Name: payments_pending_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX payments_pending_idx ON payments.payments USING btree (created_at) WHERE (status = ANY (ARRAY['pending'::text, 'pending_review'::text]));


--
-- Name: reconciliation_exceptions_open_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX reconciliation_exceptions_open_idx ON payments.reconciliation_exceptions USING btree (created_at) WHERE (status = 'open'::text);


--
-- Name: reconciliation_exceptions_owner_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX reconciliation_exceptions_owner_id_idx ON payments.reconciliation_exceptions USING btree (owner_id);


--
-- Name: reconciliation_exceptions_resolved_by_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX reconciliation_exceptions_resolved_by_idx ON payments.reconciliation_exceptions USING btree (resolved_by);


--
-- Name: refunds_approved_by_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX refunds_approved_by_idx ON payments.refunds USING btree (approved_by);


--
-- Name: refunds_order_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX refunds_order_id_idx ON payments.refunds USING btree (order_id);


--
-- Name: refunds_payment_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX refunds_payment_id_idx ON payments.refunds USING btree (payment_id);


--
-- Name: refunds_requested_by_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX refunds_requested_by_idx ON payments.refunds USING btree (requested_by);


--
-- Name: refunds_return_request_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX refunds_return_request_id_idx ON payments.refunds USING btree (return_request_id);


--
-- Name: settlement_lines_payment_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX settlement_lines_payment_id_idx ON payments.settlement_lines USING btree (payment_id);


--
-- Name: virtual_accounts_business_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX virtual_accounts_business_id_idx ON payments.virtual_accounts USING btree (business_id);


--
-- Name: virtual_accounts_order_id_idx; Type: INDEX; Schema: payments; Owner: -
--

CREATE INDEX virtual_accounts_order_id_idx ON payments.virtual_accounts USING btree (order_id);


--
-- Name: identity_links_user_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX identity_links_user_id_idx ON personalisation.identity_links USING btree (user_id);


--
-- Name: rec_item_neighbours_neighbour_product_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX rec_item_neighbours_neighbour_product_id_idx ON personalisation.rec_item_neighbours USING btree (neighbour_product_id);


--
-- Name: rec_item_neighbours_product_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX rec_item_neighbours_product_id_idx ON personalisation.rec_item_neighbours USING btree (product_id);


--
-- Name: rec_item_vectors_product_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX rec_item_vectors_product_id_idx ON personalisation.rec_item_vectors USING btree (product_id);


--
-- Name: rec_model_versions_one_active; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE UNIQUE INDEX rec_model_versions_one_active ON personalisation.rec_model_versions USING btree (model_kind) WHERE (status = 'active'::text);


--
-- Name: rec_overrides_anchor_product_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX rec_overrides_anchor_product_id_idx ON personalisation.rec_overrides USING btree (anchor_product_id);


--
-- Name: rec_overrides_created_by_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX rec_overrides_created_by_idx ON personalisation.rec_overrides USING btree (created_by);


--
-- Name: rec_overrides_product_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX rec_overrides_product_id_idx ON personalisation.rec_overrides USING btree (product_id);


--
-- Name: rec_popularity_category_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX rec_popularity_category_id_idx ON personalisation.rec_popularity USING btree (category_id);


--
-- Name: rec_popularity_lookup_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX rec_popularity_lookup_idx ON personalisation.rec_popularity USING btree ("window", category_id, state_code, rank);


--
-- Name: rec_popularity_product_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX rec_popularity_product_id_idx ON personalisation.rec_popularity USING btree (product_id);


--
-- Name: rec_popularity_state_code_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX rec_popularity_state_code_idx ON personalisation.rec_popularity USING btree (state_code);


--
-- Name: rec_staging_popularity_window_category_id_state_code_rank_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX rec_staging_popularity_window_category_id_state_code_rank_idx ON personalisation.rec_staging_popularity USING btree ("window", category_id, state_code, rank);


--
-- Name: rec_subject_candidates_product_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX rec_subject_candidates_product_id_idx ON personalisation.rec_subject_candidates USING btree (product_id);


--
-- Name: rec_subject_candidates_user_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX rec_subject_candidates_user_id_idx ON personalisation.rec_subject_candidates USING btree (user_id);


--
-- Name: user_events_car_listing_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_car_listing_id_idx ON ONLY personalisation.user_events USING btree (car_listing_id);


--
-- Name: user_events_category_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_category_id_idx ON ONLY personalisation.user_events USING btree (category_id);


--
-- Name: user_events_subject_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_subject_idx ON ONLY personalisation.user_events USING btree (anonymous_id, occurred_at);


--
-- Name: user_events_default_anonymous_id_occurred_at_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_default_anonymous_id_occurred_at_idx ON personalisation.user_events_default USING btree (anonymous_id, occurred_at);


--
-- Name: user_events_default_car_listing_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_default_car_listing_id_idx ON personalisation.user_events_default USING btree (car_listing_id);


--
-- Name: user_events_default_category_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_default_category_id_idx ON personalisation.user_events_default USING btree (category_id);


--
-- Name: user_events_listing_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_listing_id_idx ON ONLY personalisation.user_events USING btree (listing_id);


--
-- Name: user_events_default_listing_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_default_listing_id_idx ON personalisation.user_events_default USING btree (listing_id);


--
-- Name: user_events_product_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_product_id_idx ON ONLY personalisation.user_events USING btree (product_id);


--
-- Name: user_events_default_product_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_default_product_id_idx ON personalisation.user_events_default USING btree (product_id);


--
-- Name: user_events_received_brin; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_received_brin ON ONLY personalisation.user_events USING brin (received_at);


--
-- Name: user_events_default_received_at_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_default_received_at_idx ON personalisation.user_events_default USING brin (received_at);


--
-- Name: user_events_user_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_user_id_idx ON ONLY personalisation.user_events USING btree (user_id);


--
-- Name: user_events_default_user_id_idx; Type: INDEX; Schema: personalisation; Owner: -
--

CREATE INDEX user_events_default_user_id_idx ON personalisation.user_events_default USING btree (user_id);


--
-- Name: audit_log_actor_user_id_idx; Type: INDEX; Schema: platform; Owner: -
--

CREATE INDEX audit_log_actor_user_id_idx ON platform.audit_log USING btree (actor_user_id);


--
-- Name: audit_log_entity_idx; Type: INDEX; Schema: platform; Owner: -
--

CREATE INDEX audit_log_entity_idx ON platform.audit_log USING btree (entity_type, entity_id, created_at DESC);


--
-- Name: encryption_keys_one_active; Type: INDEX; Schema: platform; Owner: -
--

CREATE UNIQUE INDEX encryption_keys_one_active ON platform.encryption_keys USING btree (purpose) WHERE (retired_at IS NULL);


--
-- Name: feature_flags_updated_by_idx; Type: INDEX; Schema: platform; Owner: -
--

CREATE INDEX feature_flags_updated_by_idx ON platform.feature_flags USING btree (updated_by);


--
-- Name: files_uploaded_by_idx; Type: INDEX; Schema: platform; Owner: -
--

CREATE INDEX files_uploaded_by_idx ON platform.files USING btree (uploaded_by);


--
-- Name: idempotency_keys_expires_idx; Type: INDEX; Schema: platform; Owner: -
--

CREATE INDEX idempotency_keys_expires_idx ON platform.idempotency_keys USING btree (expires_at);


--
-- Name: idempotency_keys_user_id_idx; Type: INDEX; Schema: platform; Owner: -
--

CREATE INDEX idempotency_keys_user_id_idx ON platform.idempotency_keys USING btree (user_id);


--
-- Name: outbox_events_unpublished_idx; Type: INDEX; Schema: platform; Owner: -
--

CREATE INDEX outbox_events_unpublished_idx ON platform.outbox_events USING btree (id) WHERE (published_at IS NULL);


--
-- Name: pos_shifts_cashier_id_idx; Type: INDEX; Schema: pos; Owner: -
--

CREATE INDEX pos_shifts_cashier_id_idx ON pos.pos_shifts USING btree (cashier_id);


--
-- Name: pos_shifts_one_open_per_terminal; Type: INDEX; Schema: pos; Owner: -
--

CREATE UNIQUE INDEX pos_shifts_one_open_per_terminal ON pos.pos_shifts USING btree (terminal_id) WHERE (closed_at IS NULL);


--
-- Name: goods_receipts_purchase_order_id_idx; Type: INDEX; Schema: purchasing; Owner: -
--

CREATE INDEX goods_receipts_purchase_order_id_idx ON purchasing.goods_receipts USING btree (purchase_order_id);


--
-- Name: goods_receipts_received_by_idx; Type: INDEX; Schema: purchasing; Owner: -
--

CREATE INDEX goods_receipts_received_by_idx ON purchasing.goods_receipts USING btree (received_by);


--
-- Name: purchase_order_lines_purchase_order_id_idx; Type: INDEX; Schema: purchasing; Owner: -
--

CREATE INDEX purchase_order_lines_purchase_order_id_idx ON purchasing.purchase_order_lines USING btree (purchase_order_id);


--
-- Name: purchase_order_lines_variant_id_idx; Type: INDEX; Schema: purchasing; Owner: -
--

CREATE INDEX purchase_order_lines_variant_id_idx ON purchasing.purchase_order_lines USING btree (variant_id);


--
-- Name: purchase_orders_approved_by_idx; Type: INDEX; Schema: purchasing; Owner: -
--

CREATE INDEX purchase_orders_approved_by_idx ON purchasing.purchase_orders USING btree (approved_by);


--
-- Name: purchase_orders_created_by_idx; Type: INDEX; Schema: purchasing; Owner: -
--

CREATE INDEX purchase_orders_created_by_idx ON purchasing.purchase_orders USING btree (created_by);


--
-- Name: purchase_orders_supplier_id_idx; Type: INDEX; Schema: purchasing; Owner: -
--

CREATE INDEX purchase_orders_supplier_id_idx ON purchasing.purchase_orders USING btree (supplier_id);


--
-- Name: purchase_orders_warehouse_id_idx; Type: INDEX; Schema: purchasing; Owner: -
--

CREATE INDEX purchase_orders_warehouse_id_idx ON purchasing.purchase_orders USING btree (warehouse_id);


--
-- Name: device_blocklist_created_by_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX device_blocklist_created_by_idx ON risk.device_blocklist USING btree (created_by);


--
-- Name: enforcement_actions_created_by_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX enforcement_actions_created_by_idx ON risk.enforcement_actions USING btree (created_by);


--
-- Name: enforcement_actions_decided_by_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX enforcement_actions_decided_by_idx ON risk.enforcement_actions USING btree (decided_by);


--
-- Name: enforcement_actions_seller_id_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX enforcement_actions_seller_id_idx ON risk.enforcement_actions USING btree (seller_id);


--
-- Name: kyc_checks_subject_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX kyc_checks_subject_idx ON risk.kyc_checks USING btree (subject_type, subject_id);


--
-- Name: listing_reports_decided_by_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX listing_reports_decided_by_idx ON risk.listing_reports USING btree (decided_by);


--
-- Name: listing_reports_listing_id_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX listing_reports_listing_id_idx ON risk.listing_reports USING btree (listing_id);


--
-- Name: listing_reports_reporter_id_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX listing_reports_reporter_id_idx ON risk.listing_reports USING btree (reporter_id);


--
-- Name: review_reports_decided_by_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX review_reports_decided_by_idx ON risk.review_reports USING btree (decided_by);


--
-- Name: review_reports_reporter_id_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX review_reports_reporter_id_idx ON risk.review_reports USING btree (reporter_id);


--
-- Name: review_reports_review_id_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX review_reports_review_id_idx ON risk.review_reports USING btree (review_id);


--
-- Name: risk_cases_assigned_to_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX risk_cases_assigned_to_idx ON risk.risk_cases USING btree (assigned_to);


--
-- Name: risk_cases_queue_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX risk_cases_queue_idx ON risk.risk_cases USING btree (status, score DESC, created_at) WHERE (status = ANY (ARRAY['open'::text, 'in_review'::text, 'escalated'::text]));


--
-- Name: risk_cases_resolved_by_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX risk_cases_resolved_by_idx ON risk.risk_cases USING btree (resolved_by);


--
-- Name: risk_decisions_subject_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX risk_decisions_subject_idx ON ONLY risk.risk_decisions USING btree (subject_type, subject_id);


--
-- Name: risk_decisions_default_subject_type_subject_id_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX risk_decisions_default_subject_type_subject_id_idx ON risk.risk_decisions_default USING btree (subject_type, subject_id);


--
-- Name: risk_decisions_user_id_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX risk_decisions_user_id_idx ON ONLY risk.risk_decisions USING btree (user_id);


--
-- Name: risk_decisions_default_user_id_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX risk_decisions_default_user_id_idx ON risk.risk_decisions_default USING btree (user_id);


--
-- Name: risk_links_entity_b_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX risk_links_entity_b_idx ON risk.risk_links USING btree (entity_b);


--
-- Name: risk_lists_created_by_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX risk_lists_created_by_idx ON risk.risk_lists USING btree (created_by);


--
-- Name: risk_rule_sets_created_by_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX risk_rule_sets_created_by_idx ON risk.risk_rule_sets USING btree (created_by);


--
-- Name: risk_rule_sets_one_active; Type: INDEX; Schema: risk; Owner: -
--

CREATE UNIQUE INDEX risk_rule_sets_one_active ON risk.risk_rule_sets USING btree (checkpoint) WHERE (status = 'active'::text);


--
-- Name: risk_rule_sets_published_by_idx; Type: INDEX; Schema: risk; Owner: -
--

CREATE INDEX risk_rule_sets_published_by_idx ON risk.risk_rule_sets USING btree (published_by);


--
-- Name: cart_items_listing_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX cart_items_listing_id_idx ON sales.cart_items USING btree (listing_id);


--
-- Name: carts_business_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX carts_business_id_idx ON sales.carts USING btree (business_id);


--
-- Name: carts_one_active_per_user; Type: INDEX; Schema: sales; Owner: -
--

CREATE UNIQUE INDEX carts_one_active_per_user ON sales.carts USING btree (user_id) WHERE ((status = 'active'::text) AND (user_id IS NOT NULL) AND (business_id IS NULL));


--
-- Name: fulfilments_collection_store_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX fulfilments_collection_store_id_idx ON sales.fulfilments USING btree (collection_store_id);


--
-- Name: fulfilments_seller_status_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX fulfilments_seller_status_idx ON sales.fulfilments USING btree (seller_id, status);


--
-- Name: fulfilments_sla_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX fulfilments_sla_idx ON sales.fulfilments USING btree (accept_by) WHERE (status = 'pending'::text);


--
-- Name: fulfilments_warehouse_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX fulfilments_warehouse_id_idx ON sales.fulfilments USING btree (warehouse_id);


--
-- Name: order_lines_fulfilment_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX order_lines_fulfilment_id_idx ON sales.order_lines USING btree (fulfilment_id);


--
-- Name: order_lines_listing_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX order_lines_listing_id_idx ON sales.order_lines USING btree (listing_id);


--
-- Name: order_lines_order_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX order_lines_order_id_idx ON sales.order_lines USING btree (order_id);


--
-- Name: order_lines_seller_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX order_lines_seller_id_idx ON sales.order_lines USING btree (seller_id);


--
-- Name: order_lines_variant_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX order_lines_variant_id_idx ON sales.order_lines USING btree (variant_id);


--
-- Name: order_status_history_actor_user_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX order_status_history_actor_user_id_idx ON sales.order_status_history USING btree (actor_user_id);


--
-- Name: order_status_history_fulfilment_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX order_status_history_fulfilment_id_idx ON sales.order_status_history USING btree (fulfilment_id);


--
-- Name: order_status_history_order_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX order_status_history_order_id_idx ON sales.order_status_history USING btree (order_id);


--
-- Name: orders_attributed_notification_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX orders_attributed_notification_id_idx ON sales.orders USING btree (attributed_notification_id);


--
-- Name: orders_business_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX orders_business_id_idx ON sales.orders USING btree (business_id);


--
-- Name: orders_customer_placed_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX orders_customer_placed_idx ON sales.orders USING btree (customer_user_id, placed_at DESC);


--
-- Name: orders_pos_shift_id_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX orders_pos_shift_id_idx ON sales.orders USING btree (pos_shift_id);


--
-- Name: orders_unpaid_due_idx; Type: INDEX; Schema: sales; Owner: -
--

CREATE INDEX orders_unpaid_due_idx ON sales.orders USING btree (payment_due_at) WHERE (status = 'pending_payment'::text);


--
-- Name: catalogue_gaps_owner_id_idx; Type: INDEX; Schema: search; Owner: -
--

CREATE INDEX catalogue_gaps_owner_id_idx ON search.catalogue_gaps USING btree (owner_id);


--
-- Name: search_pins_created_by_idx; Type: INDEX; Schema: search; Owner: -
--

CREATE INDEX search_pins_created_by_idx ON search.search_pins USING btree (created_by);


--
-- Name: search_pins_product_id_idx; Type: INDEX; Schema: search; Owner: -
--

CREATE INDEX search_pins_product_id_idx ON search.search_pins USING btree (product_id);


--
-- Name: search_pins_query_idx; Type: INDEX; Schema: search; Owner: -
--

CREATE INDEX search_pins_query_idx ON search.search_pins USING btree (query_norm, ends_at);


--
-- Name: search_queries_zero_idx; Type: INDEX; Schema: search; Owner: -
--

CREATE INDEX search_queries_zero_idx ON ONLY search.search_queries USING btree (query_norm, created_at) WHERE (results_count = 0);


--
-- Name: search_queries_default_query_norm_created_at_idx; Type: INDEX; Schema: search; Owner: -
--

CREATE INDEX search_queries_default_query_norm_created_at_idx ON search.search_queries_default USING btree (query_norm, created_at) WHERE (results_count = 0);


--
-- Name: search_queries_user_id_idx; Type: INDEX; Schema: search; Owner: -
--

CREATE INDEX search_queries_user_id_idx ON ONLY search.search_queries USING btree (user_id);


--
-- Name: search_queries_default_user_id_idx; Type: INDEX; Schema: search; Owner: -
--

CREATE INDEX search_queries_default_user_id_idx ON search.search_queries_default USING btree (user_id);


--
-- Name: search_redirects_created_by_idx; Type: INDEX; Schema: search; Owner: -
--

CREATE INDEX search_redirects_created_by_idx ON search.search_redirects USING btree (created_by);


--
-- Name: search_synonyms_created_by_idx; Type: INDEX; Schema: search; Owner: -
--

CREATE INDEX search_synonyms_created_by_idx ON search.search_synonyms USING btree (created_by);


--
-- Name: kyc_documents_file_id_idx; Type: INDEX; Schema: sellers; Owner: -
--

CREATE INDEX kyc_documents_file_id_idx ON sellers.kyc_documents USING btree (file_id);


--
-- Name: kyc_documents_submission_id_idx; Type: INDEX; Schema: sellers; Owner: -
--

CREATE INDEX kyc_documents_submission_id_idx ON sellers.kyc_documents USING btree (submission_id);


--
-- Name: seller_bank_accounts_encryption_key_id_idx; Type: INDEX; Schema: sellers; Owner: -
--

CREATE INDEX seller_bank_accounts_encryption_key_id_idx ON sellers.seller_bank_accounts USING btree (encryption_key_id);


--
-- Name: seller_bank_accounts_hmac_idx; Type: INDEX; Schema: sellers; Owner: -
--

CREATE INDEX seller_bank_accounts_hmac_idx ON sellers.seller_bank_accounts USING btree (account_number_hmac);


--
-- Name: seller_bank_accounts_one_default; Type: INDEX; Schema: sellers; Owner: -
--

CREATE UNIQUE INDEX seller_bank_accounts_one_default ON sellers.seller_bank_accounts USING btree (seller_id) WHERE is_default;


--
-- Name: seller_kyc_submissions_encryption_key_id_idx; Type: INDEX; Schema: sellers; Owner: -
--

CREATE INDEX seller_kyc_submissions_encryption_key_id_idx ON sellers.seller_kyc_submissions USING btree (encryption_key_id);


--
-- Name: seller_kyc_submissions_hmac_idx; Type: INDEX; Schema: sellers; Owner: -
--

CREATE INDEX seller_kyc_submissions_hmac_idx ON sellers.seller_kyc_submissions USING btree (id_number_hmac);


--
-- Name: seller_kyc_submissions_reviewed_by_idx; Type: INDEX; Schema: sellers; Owner: -
--

CREATE INDEX seller_kyc_submissions_reviewed_by_idx ON sellers.seller_kyc_submissions USING btree (reviewed_by);


--
-- Name: seller_kyc_submissions_seller_id_idx; Type: INDEX; Schema: sellers; Owner: -
--

CREATE INDEX seller_kyc_submissions_seller_id_idx ON sellers.seller_kyc_submissions USING btree (seller_id);


--
-- Name: seller_members_user_idx; Type: INDEX; Schema: sellers; Owner: -
--

CREATE INDEX seller_members_user_idx ON sellers.seller_members USING btree (user_id);


--
-- Name: sellers_one_first_party; Type: INDEX; Schema: sellers; Owner: -
--

CREATE UNIQUE INDEX sellers_one_first_party ON sellers.sellers USING btree ((true)) WHERE (type = 'first_party'::text);


--
-- Name: sellers_owner_user_id_idx; Type: INDEX; Schema: sellers; Owner: -
--

CREATE INDEX sellers_owner_user_id_idx ON sellers.sellers USING btree (owner_user_id);


--
-- Name: sellers_state_code_idx; Type: INDEX; Schema: sellers; Owner: -
--

CREATE INDEX sellers_state_code_idx ON sellers.sellers USING btree (state_code);


--
-- Name: message_attachments_file_id_idx; Type: INDEX; Schema: support; Owner: -
--

CREATE INDEX message_attachments_file_id_idx ON support.message_attachments USING btree (file_id);


--
-- Name: support_tickets_assigned_to_idx; Type: INDEX; Schema: support; Owner: -
--

CREATE INDEX support_tickets_assigned_to_idx ON support.support_tickets USING btree (assigned_to);


--
-- Name: support_tickets_customer_user_id_idx; Type: INDEX; Schema: support; Owner: -
--

CREATE INDEX support_tickets_customer_user_id_idx ON support.support_tickets USING btree (customer_user_id);


--
-- Name: support_tickets_order_id_idx; Type: INDEX; Schema: support; Owner: -
--

CREATE INDEX support_tickets_order_id_idx ON support.support_tickets USING btree (order_id);


--
-- Name: support_tickets_queue_idx; Type: INDEX; Schema: support; Owner: -
--

CREATE INDEX support_tickets_queue_idx ON support.support_tickets USING btree (status, priority, created_at);


--
-- Name: support_tickets_seller_id_idx; Type: INDEX; Schema: support; Owner: -
--

CREATE INDEX support_tickets_seller_id_idx ON support.support_tickets USING btree (seller_id);


--
-- Name: ticket_messages_author_user_id_idx; Type: INDEX; Schema: support; Owner: -
--

CREATE INDEX ticket_messages_author_user_id_idx ON support.ticket_messages USING btree (author_user_id);


--
-- Name: ticket_messages_ticket_id_idx; Type: INDEX; Schema: support; Owner: -
--

CREATE INDEX ticket_messages_ticket_id_idx ON support.ticket_messages USING btree (ticket_id);


--
-- Name: auth_events_default_pkey; Type: INDEX ATTACH; Schema: identity; Owner: -
--

ALTER INDEX identity.auth_events_pkey ATTACH PARTITION identity.auth_events_default_pkey;


--
-- Name: auth_events_default_user_id_created_at_idx; Type: INDEX ATTACH; Schema: identity; Owner: -
--

ALTER INDEX identity.auth_events_user_idx ATTACH PARTITION identity.auth_events_default_user_id_created_at_idx;


--
-- Name: rider_location_pings_default_pkey; Type: INDEX ATTACH; Schema: logistics; Owner: -
--

ALTER INDEX logistics.rider_location_pings_pkey ATTACH PARTITION logistics.rider_location_pings_default_pkey;


--
-- Name: message_events_default_notification_id_idx; Type: INDEX ATTACH; Schema: messaging; Owner: -
--

ALTER INDEX messaging.message_events_notification_idx ATTACH PARTITION messaging.message_events_default_notification_id_idx;


--
-- Name: message_events_default_pkey; Type: INDEX ATTACH; Schema: messaging; Owner: -
--

ALTER INDEX messaging.message_events_pkey ATTACH PARTITION messaging.message_events_default_pkey;


--
-- Name: user_events_default_anonymous_id_occurred_at_idx; Type: INDEX ATTACH; Schema: personalisation; Owner: -
--

ALTER INDEX personalisation.user_events_subject_idx ATTACH PARTITION personalisation.user_events_default_anonymous_id_occurred_at_idx;


--
-- Name: user_events_default_car_listing_id_idx; Type: INDEX ATTACH; Schema: personalisation; Owner: -
--

ALTER INDEX personalisation.user_events_car_listing_id_idx ATTACH PARTITION personalisation.user_events_default_car_listing_id_idx;


--
-- Name: user_events_default_category_id_idx; Type: INDEX ATTACH; Schema: personalisation; Owner: -
--

ALTER INDEX personalisation.user_events_category_id_idx ATTACH PARTITION personalisation.user_events_default_category_id_idx;


--
-- Name: user_events_default_listing_id_idx; Type: INDEX ATTACH; Schema: personalisation; Owner: -
--

ALTER INDEX personalisation.user_events_listing_id_idx ATTACH PARTITION personalisation.user_events_default_listing_id_idx;


--
-- Name: user_events_default_pkey; Type: INDEX ATTACH; Schema: personalisation; Owner: -
--

ALTER INDEX personalisation.user_events_pkey ATTACH PARTITION personalisation.user_events_default_pkey;


--
-- Name: user_events_default_product_id_idx; Type: INDEX ATTACH; Schema: personalisation; Owner: -
--

ALTER INDEX personalisation.user_events_product_id_idx ATTACH PARTITION personalisation.user_events_default_product_id_idx;


--
-- Name: user_events_default_received_at_idx; Type: INDEX ATTACH; Schema: personalisation; Owner: -
--

ALTER INDEX personalisation.user_events_received_brin ATTACH PARTITION personalisation.user_events_default_received_at_idx;


--
-- Name: user_events_default_user_id_idx; Type: INDEX ATTACH; Schema: personalisation; Owner: -
--

ALTER INDEX personalisation.user_events_user_id_idx ATTACH PARTITION personalisation.user_events_default_user_id_idx;


--
-- Name: risk_decisions_default_pkey; Type: INDEX ATTACH; Schema: risk; Owner: -
--

ALTER INDEX risk.risk_decisions_pkey ATTACH PARTITION risk.risk_decisions_default_pkey;


--
-- Name: risk_decisions_default_subject_type_subject_id_idx; Type: INDEX ATTACH; Schema: risk; Owner: -
--

ALTER INDEX risk.risk_decisions_subject_idx ATTACH PARTITION risk.risk_decisions_default_subject_type_subject_id_idx;


--
-- Name: risk_decisions_default_user_id_idx; Type: INDEX ATTACH; Schema: risk; Owner: -
--

ALTER INDEX risk.risk_decisions_user_id_idx ATTACH PARTITION risk.risk_decisions_default_user_id_idx;


--
-- Name: search_queries_default_pkey; Type: INDEX ATTACH; Schema: search; Owner: -
--

ALTER INDEX search.search_queries_pkey ATTACH PARTITION search.search_queries_default_pkey;


--
-- Name: search_queries_default_query_norm_created_at_idx; Type: INDEX ATTACH; Schema: search; Owner: -
--

ALTER INDEX search.search_queries_zero_idx ATTACH PARTITION search.search_queries_default_query_norm_created_at_idx;


--
-- Name: search_queries_default_user_id_idx; Type: INDEX ATTACH; Schema: search; Owner: -
--

ALTER INDEX search.search_queries_user_id_idx ATTACH PARTITION search.search_queries_default_user_id_idx;


--
-- Name: repair_jobs repair_jobs_set_updated_at; Type: TRIGGER; Schema: aftersales; Owner: -
--

CREATE TRIGGER repair_jobs_set_updated_at BEFORE UPDATE ON aftersales.repair_jobs FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: return_requests return_requests_set_updated_at; Type: TRIGGER; Schema: aftersales; Owner: -
--

CREATE TRIGGER return_requests_set_updated_at BEFORE UPDATE ON aftersales.return_requests FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: trade_ins trade_ins_set_updated_at; Type: TRIGGER; Schema: aftersales; Owner: -
--

CREATE TRIGGER trade_ins_set_updated_at BEFORE UPDATE ON aftersales.trade_ins FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: warranty_claims warranty_claims_set_updated_at; Type: TRIGGER; Schema: aftersales; Owner: -
--

CREATE TRIGGER warranty_claims_set_updated_at BEFORE UPDATE ON aftersales.warranty_claims FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: businesses businesses_set_updated_at; Type: TRIGGER; Schema: b2b; Owner: -
--

CREATE TRIGGER businesses_set_updated_at BEFORE UPDATE ON b2b.businesses FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: credit_accounts credit_accounts_set_updated_at; Type: TRIGGER; Schema: b2b; Owner: -
--

CREATE TRIGGER credit_accounts_set_updated_at BEFORE UPDATE ON b2b.credit_accounts FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: quotes quotes_set_updated_at; Type: TRIGGER; Schema: b2b; Owner: -
--

CREATE TRIGGER quotes_set_updated_at BEFORE UPDATE ON b2b.quotes FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: car_listings car_listings_set_updated_at; Type: TRIGGER; Schema: cars; Owner: -
--

CREATE TRIGGER car_listings_set_updated_at BEFORE UPDATE ON cars.car_listings FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: financing_enquiries financing_enquiries_set_updated_at; Type: TRIGGER; Schema: cars; Owner: -
--

CREATE TRIGGER financing_enquiries_set_updated_at BEFORE UPDATE ON cars.financing_enquiries FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: viewing_bookings viewing_bookings_set_updated_at; Type: TRIGGER; Schema: cars; Owner: -
--

CREATE TRIGGER viewing_bookings_set_updated_at BEFORE UPDATE ON cars.viewing_bookings FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: categories categories_set_updated_at; Type: TRIGGER; Schema: catalog; Owner: -
--

CREATE TRIGGER categories_set_updated_at BEFORE UPDATE ON catalog.categories FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: listings listings_set_updated_at; Type: TRIGGER; Schema: catalog; Owner: -
--

CREATE TRIGGER listings_set_updated_at BEFORE UPDATE ON catalog.listings FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: product_reviews product_reviews_set_updated_at; Type: TRIGGER; Schema: catalog; Owner: -
--

CREATE TRIGGER product_reviews_set_updated_at BEFORE UPDATE ON catalog.product_reviews FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: product_variants product_variants_set_updated_at; Type: TRIGGER; Schema: catalog; Owner: -
--

CREATE TRIGGER product_variants_set_updated_at BEFORE UPDATE ON catalog.product_variants FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: products products_set_updated_at; Type: TRIGGER; Schema: catalog; Owner: -
--

CREATE TRIGGER products_set_updated_at BEFORE UPDATE ON catalog.products FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: banners banners_set_updated_at; Type: TRIGGER; Schema: content; Owner: -
--

CREATE TRIGGER banners_set_updated_at BEFORE UPDATE ON content.banners FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: cms_pages cms_pages_set_updated_at; Type: TRIGGER; Schema: content; Owner: -
--

CREATE TRIGGER cms_pages_set_updated_at BEFORE UPDATE ON content.cms_pages FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: help_articles help_articles_set_updated_at; Type: TRIGGER; Schema: content; Owner: -
--

CREATE TRIGGER help_articles_set_updated_at BEFORE UPDATE ON content.help_articles FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: invoices invoices_set_updated_at; Type: TRIGGER; Schema: finance; Owner: -
--

CREATE TRIGGER invoices_set_updated_at BEFORE UPDATE ON finance.invoices FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: ledger_entries ledger_entries_append_only; Type: TRIGGER; Schema: finance; Owner: -
--

CREATE TRIGGER ledger_entries_append_only BEFORE DELETE OR UPDATE ON finance.ledger_entries FOR EACH ROW EXECUTE FUNCTION platform.forbid_change();


--
-- Name: ledger_entries ledger_entries_balanced; Type: TRIGGER; Schema: finance; Owner: -
--

CREATE CONSTRAINT TRIGGER ledger_entries_balanced AFTER INSERT ON finance.ledger_entries DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION finance.assert_journal_balanced();


--
-- Name: ledger_journals ledger_journals_append_only; Type: TRIGGER; Schema: finance; Owner: -
--

CREATE TRIGGER ledger_journals_append_only BEFORE DELETE OR UPDATE ON finance.ledger_journals FOR EACH ROW EXECUTE FUNCTION platform.forbid_change();


--
-- Name: ledger_journals ledger_journals_period_open; Type: TRIGGER; Schema: finance; Owner: -
--

CREATE TRIGGER ledger_journals_period_open BEFORE INSERT ON finance.ledger_journals FOR EACH ROW EXECUTE FUNCTION finance.assert_period_open();


--
-- Name: payout_batches payout_batches_set_updated_at; Type: TRIGGER; Schema: finance; Owner: -
--

CREATE TRIGGER payout_batches_set_updated_at BEFORE UPDATE ON finance.payout_batches FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: payouts payouts_set_updated_at; Type: TRIGGER; Schema: finance; Owner: -
--

CREATE TRIGGER payouts_set_updated_at BEFORE UPDATE ON finance.payouts FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: pick_lists pick_lists_set_updated_at; Type: TRIGGER; Schema: fulfilment; Owner: -
--

CREATE TRIGGER pick_lists_set_updated_at BEFORE UPDATE ON fulfilment.pick_lists FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: shipments shipments_set_updated_at; Type: TRIGGER; Schema: fulfilment; Owner: -
--

CREATE TRIGGER shipments_set_updated_at BEFORE UPDATE ON fulfilment.shipments FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: addresses addresses_set_updated_at; Type: TRIGGER; Schema: identity; Owner: -
--

CREATE TRIGGER addresses_set_updated_at BEFORE UPDATE ON identity.addresses FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: privacy_requests privacy_requests_set_updated_at; Type: TRIGGER; Schema: identity; Owner: -
--

CREATE TRIGGER privacy_requests_set_updated_at BEFORE UPDATE ON identity.privacy_requests FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: roles roles_set_updated_at; Type: TRIGGER; Schema: identity; Owner: -
--

CREATE TRIGGER roles_set_updated_at BEFORE UPDATE ON identity.roles FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: staff_members staff_members_set_updated_at; Type: TRIGGER; Schema: identity; Owner: -
--

CREATE TRIGGER staff_members_set_updated_at BEFORE UPDATE ON identity.staff_members FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: users users_set_updated_at; Type: TRIGGER; Schema: identity; Owner: -
--

CREATE TRIGGER users_set_updated_at BEFORE UPDATE ON identity.users FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: device_units device_units_set_updated_at; Type: TRIGGER; Schema: inventory; Owner: -
--

CREATE TRIGGER device_units_set_updated_at BEFORE UPDATE ON inventory.device_units FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: inventory_costs inventory_costs_set_updated_at; Type: TRIGGER; Schema: inventory; Owner: -
--

CREATE TRIGGER inventory_costs_set_updated_at BEFORE UPDATE ON inventory.inventory_costs FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: inventory_counts inventory_counts_set_updated_at; Type: TRIGGER; Schema: inventory; Owner: -
--

CREATE TRIGGER inventory_counts_set_updated_at BEFORE UPDATE ON inventory.inventory_counts FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: inventory_levels inventory_levels_set_updated_at; Type: TRIGGER; Schema: inventory; Owner: -
--

CREATE TRIGGER inventory_levels_set_updated_at BEFORE UPDATE ON inventory.inventory_levels FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: stock_movements stock_movements_append_only; Type: TRIGGER; Schema: inventory; Owner: -
--

CREATE TRIGGER stock_movements_append_only BEFORE DELETE OR UPDATE ON inventory.stock_movements FOR EACH ROW EXECUTE FUNCTION platform.forbid_change();


--
-- Name: stock_reservations stock_reservations_set_updated_at; Type: TRIGGER; Schema: inventory; Owner: -
--

CREATE TRIGGER stock_reservations_set_updated_at BEFORE UPDATE ON inventory.stock_reservations FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: stock_transfers stock_transfers_set_updated_at; Type: TRIGGER; Schema: inventory; Owner: -
--

CREATE TRIGGER stock_transfers_set_updated_at BEFORE UPDATE ON inventory.stock_transfers FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: warehouses warehouses_set_updated_at; Type: TRIGGER; Schema: inventory; Owner: -
--

CREATE TRIGGER warehouses_set_updated_at BEFORE UPDATE ON inventory.warehouses FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: delivery_events delivery_events_append_only; Type: TRIGGER; Schema: logistics; Owner: -
--

CREATE TRIGGER delivery_events_append_only BEFORE DELETE OR UPDATE ON logistics.delivery_events FOR EACH ROW EXECUTE FUNCTION platform.forbid_change();


--
-- Name: delivery_jobs delivery_jobs_set_updated_at; Type: TRIGGER; Schema: logistics; Owner: -
--

CREATE TRIGGER delivery_jobs_set_updated_at BEFORE UPDATE ON logistics.delivery_jobs FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: riders riders_set_updated_at; Type: TRIGGER; Schema: logistics; Owner: -
--

CREATE TRIGGER riders_set_updated_at BEFORE UPDATE ON logistics.riders FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: campaigns campaigns_set_updated_at; Type: TRIGGER; Schema: marketing; Owner: -
--

CREATE TRIGGER campaigns_set_updated_at BEFORE UPDATE ON marketing.campaigns FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: journey_enrollments journey_enrollments_set_updated_at; Type: TRIGGER; Schema: marketing; Owner: -
--

CREATE TRIGGER journey_enrollments_set_updated_at BEFORE UPDATE ON marketing.journey_enrollments FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: journeys journeys_set_updated_at; Type: TRIGGER; Schema: marketing; Owner: -
--

CREATE TRIGGER journeys_set_updated_at BEFORE UPDATE ON marketing.journeys FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: promotions promotions_set_updated_at; Type: TRIGGER; Schema: marketing; Owner: -
--

CREATE TRIGGER promotions_set_updated_at BEFORE UPDATE ON marketing.promotions FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: segments segments_set_updated_at; Type: TRIGGER; Schema: marketing; Owner: -
--

CREATE TRIGGER segments_set_updated_at BEFORE UPDATE ON marketing.segments FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: message_templates message_templates_set_updated_at; Type: TRIGGER; Schema: messaging; Owner: -
--

CREATE TRIGGER message_templates_set_updated_at BEFORE UPDATE ON messaging.message_templates FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: notification_preferences notification_preferences_set_updated_at; Type: TRIGGER; Schema: messaging; Owner: -
--

CREATE TRIGGER notification_preferences_set_updated_at BEFORE UPDATE ON messaging.notification_preferences FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: notification_settings notification_settings_set_updated_at; Type: TRIGGER; Schema: messaging; Owner: -
--

CREATE TRIGGER notification_settings_set_updated_at BEFORE UPDATE ON messaging.notification_settings FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: disputes disputes_set_updated_at; Type: TRIGGER; Schema: payments; Owner: -
--

CREATE TRIGGER disputes_set_updated_at BEFORE UPDATE ON payments.disputes FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: payments payments_set_updated_at; Type: TRIGGER; Schema: payments; Owner: -
--

CREATE TRIGGER payments_set_updated_at BEFORE UPDATE ON payments.payments FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: provider_health provider_health_set_updated_at; Type: TRIGGER; Schema: payments; Owner: -
--

CREATE TRIGGER provider_health_set_updated_at BEFORE UPDATE ON payments.provider_health FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: reconciliation_exceptions reconciliation_exceptions_set_updated_at; Type: TRIGGER; Schema: payments; Owner: -
--

CREATE TRIGGER reconciliation_exceptions_set_updated_at BEFORE UPDATE ON payments.reconciliation_exceptions FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: refunds refunds_set_updated_at; Type: TRIGGER; Schema: payments; Owner: -
--

CREATE TRIGGER refunds_set_updated_at BEFORE UPDATE ON payments.refunds FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: virtual_accounts virtual_accounts_set_updated_at; Type: TRIGGER; Schema: payments; Owner: -
--

CREATE TRIGGER virtual_accounts_set_updated_at BEFORE UPDATE ON payments.virtual_accounts FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: feature_flags feature_flags_set_updated_at; Type: TRIGGER; Schema: platform; Owner: -
--

CREATE TRIGGER feature_flags_set_updated_at BEFORE UPDATE ON platform.feature_flags FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: outbox_events outbox_events_notify; Type: TRIGGER; Schema: platform; Owner: -
--

CREATE TRIGGER outbox_events_notify AFTER INSERT ON platform.outbox_events FOR EACH STATEMENT EXECUTE FUNCTION platform.notify_outbox();


--
-- Name: purchase_orders purchase_orders_set_updated_at; Type: TRIGGER; Schema: purchasing; Owner: -
--

CREATE TRIGGER purchase_orders_set_updated_at BEFORE UPDATE ON purchasing.purchase_orders FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: suppliers suppliers_set_updated_at; Type: TRIGGER; Schema: purchasing; Owner: -
--

CREATE TRIGGER suppliers_set_updated_at BEFORE UPDATE ON purchasing.suppliers FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: listing_reports listing_reports_set_updated_at; Type: TRIGGER; Schema: risk; Owner: -
--

CREATE TRIGGER listing_reports_set_updated_at BEFORE UPDATE ON risk.listing_reports FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: review_reports review_reports_set_updated_at; Type: TRIGGER; Schema: risk; Owner: -
--

CREATE TRIGGER review_reports_set_updated_at BEFORE UPDATE ON risk.review_reports FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: risk_cases risk_cases_set_updated_at; Type: TRIGGER; Schema: risk; Owner: -
--

CREATE TRIGGER risk_cases_set_updated_at BEFORE UPDATE ON risk.risk_cases FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: carts carts_set_updated_at; Type: TRIGGER; Schema: sales; Owner: -
--

CREATE TRIGGER carts_set_updated_at BEFORE UPDATE ON sales.carts FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: fulfilments fulfilments_set_updated_at; Type: TRIGGER; Schema: sales; Owner: -
--

CREATE TRIGGER fulfilments_set_updated_at BEFORE UPDATE ON sales.fulfilments FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: orders orders_set_updated_at; Type: TRIGGER; Schema: sales; Owner: -
--

CREATE TRIGGER orders_set_updated_at BEFORE UPDATE ON sales.orders FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: catalogue_gaps catalogue_gaps_set_updated_at; Type: TRIGGER; Schema: search; Owner: -
--

CREATE TRIGGER catalogue_gaps_set_updated_at BEFORE UPDATE ON search.catalogue_gaps FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: search_synonyms search_synonyms_set_updated_at; Type: TRIGGER; Schema: search; Owner: -
--

CREATE TRIGGER search_synonyms_set_updated_at BEFORE UPDATE ON search.search_synonyms FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: seller_bank_accounts seller_bank_accounts_set_updated_at; Type: TRIGGER; Schema: sellers; Owner: -
--

CREATE TRIGGER seller_bank_accounts_set_updated_at BEFORE UPDATE ON sellers.seller_bank_accounts FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: seller_kyc_submissions seller_kyc_submissions_set_updated_at; Type: TRIGGER; Schema: sellers; Owner: -
--

CREATE TRIGGER seller_kyc_submissions_set_updated_at BEFORE UPDATE ON sellers.seller_kyc_submissions FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: sellers sellers_set_updated_at; Type: TRIGGER; Schema: sellers; Owner: -
--

CREATE TRIGGER sellers_set_updated_at BEFORE UPDATE ON sellers.sellers FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: support_tickets support_tickets_set_updated_at; Type: TRIGGER; Schema: support; Owner: -
--

CREATE TRIGGER support_tickets_set_updated_at BEFORE UPDATE ON support.support_tickets FOR EACH ROW EXECUTE FUNCTION platform.set_updated_at();


--
-- Name: repair_jobs repair_jobs_device_unit_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.repair_jobs
    ADD CONSTRAINT repair_jobs_device_unit_id_fkey FOREIGN KEY (device_unit_id) REFERENCES inventory.device_units(id);


--
-- Name: repair_jobs repair_jobs_technician_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.repair_jobs
    ADD CONSTRAINT repair_jobs_technician_id_fkey FOREIGN KEY (technician_id) REFERENCES identity.staff_members(user_id);


--
-- Name: repair_jobs repair_jobs_warranty_claim_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.repair_jobs
    ADD CONSTRAINT repair_jobs_warranty_claim_id_fkey FOREIGN KEY (warranty_claim_id) REFERENCES aftersales.warranty_claims(id);


--
-- Name: return_inspections return_inspections_device_unit_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.return_inspections
    ADD CONSTRAINT return_inspections_device_unit_id_fkey FOREIGN KEY (device_unit_id) REFERENCES inventory.device_units(id);


--
-- Name: return_inspections return_inspections_inspected_by_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.return_inspections
    ADD CONSTRAINT return_inspections_inspected_by_fkey FOREIGN KEY (inspected_by) REFERENCES identity.users(id);


--
-- Name: return_inspections return_inspections_return_item_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.return_inspections
    ADD CONSTRAINT return_inspections_return_item_id_fkey FOREIGN KEY (return_item_id) REFERENCES aftersales.return_items(id) ON DELETE CASCADE;


--
-- Name: return_items return_items_order_line_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.return_items
    ADD CONSTRAINT return_items_order_line_id_fkey FOREIGN KEY (order_line_id) REFERENCES sales.order_lines(id);


--
-- Name: return_items return_items_return_request_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.return_items
    ADD CONSTRAINT return_items_return_request_id_fkey FOREIGN KEY (return_request_id) REFERENCES aftersales.return_requests(id) ON DELETE CASCADE;


--
-- Name: return_requests return_requests_decided_by_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.return_requests
    ADD CONSTRAINT return_requests_decided_by_fkey FOREIGN KEY (decided_by) REFERENCES identity.users(id);


--
-- Name: return_requests return_requests_order_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.return_requests
    ADD CONSTRAINT return_requests_order_id_fkey FOREIGN KEY (order_id) REFERENCES sales.orders(id);


--
-- Name: return_requests return_requests_requested_by_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.return_requests
    ADD CONSTRAINT return_requests_requested_by_fkey FOREIGN KEY (requested_by) REFERENCES identity.users(id);


--
-- Name: trade_ins trade_ins_device_unit_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.trade_ins
    ADD CONSTRAINT trade_ins_device_unit_id_fkey FOREIGN KEY (device_unit_id) REFERENCES inventory.device_units(id);


--
-- Name: trade_ins trade_ins_encryption_key_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.trade_ins
    ADD CONSTRAINT trade_ins_encryption_key_id_fkey FOREIGN KEY (encryption_key_id) REFERENCES platform.encryption_keys(id);


--
-- Name: trade_ins trade_ins_inspected_by_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.trade_ins
    ADD CONSTRAINT trade_ins_inspected_by_fkey FOREIGN KEY (inspected_by) REFERENCES identity.staff_members(user_id);


--
-- Name: trade_ins trade_ins_user_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.trade_ins
    ADD CONSTRAINT trade_ins_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: warranty_claims warranty_claims_customer_user_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.warranty_claims
    ADD CONSTRAINT warranty_claims_customer_user_id_fkey FOREIGN KEY (customer_user_id) REFERENCES identity.users(id);


--
-- Name: warranty_claims warranty_claims_device_unit_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.warranty_claims
    ADD CONSTRAINT warranty_claims_device_unit_id_fkey FOREIGN KEY (device_unit_id) REFERENCES inventory.device_units(id);


--
-- Name: warranty_claims warranty_claims_order_line_id_fkey; Type: FK CONSTRAINT; Schema: aftersales; Owner: -
--

ALTER TABLE ONLY aftersales.warranty_claims
    ADD CONSTRAINT warranty_claims_order_line_id_fkey FOREIGN KEY (order_line_id) REFERENCES sales.order_lines(id);


--
-- Name: business_members business_members_business_id_fkey; Type: FK CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.business_members
    ADD CONSTRAINT business_members_business_id_fkey FOREIGN KEY (business_id) REFERENCES b2b.businesses(id) ON DELETE CASCADE;


--
-- Name: business_members business_members_user_id_fkey; Type: FK CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.business_members
    ADD CONSTRAINT business_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: businesses businesses_account_manager_id_fkey; Type: FK CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.businesses
    ADD CONSTRAINT businesses_account_manager_id_fkey FOREIGN KEY (account_manager_id) REFERENCES identity.staff_members(user_id);


--
-- Name: credit_accounts credit_accounts_approved_by_fkey; Type: FK CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.credit_accounts
    ADD CONSTRAINT credit_accounts_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES identity.users(id);


--
-- Name: credit_accounts credit_accounts_business_id_fkey; Type: FK CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.credit_accounts
    ADD CONSTRAINT credit_accounts_business_id_fkey FOREIGN KEY (business_id) REFERENCES b2b.businesses(id) ON DELETE CASCADE;


--
-- Name: quote_lines quote_lines_listing_id_fkey; Type: FK CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.quote_lines
    ADD CONSTRAINT quote_lines_listing_id_fkey FOREIGN KEY (listing_id) REFERENCES catalog.listings(id);


--
-- Name: quote_lines quote_lines_quote_id_fkey; Type: FK CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.quote_lines
    ADD CONSTRAINT quote_lines_quote_id_fkey FOREIGN KEY (quote_id) REFERENCES b2b.quotes(id) ON DELETE CASCADE;


--
-- Name: quotes quotes_assigned_to_fkey; Type: FK CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.quotes
    ADD CONSTRAINT quotes_assigned_to_fkey FOREIGN KEY (assigned_to) REFERENCES identity.staff_members(user_id);


--
-- Name: quotes quotes_business_id_fkey; Type: FK CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.quotes
    ADD CONSTRAINT quotes_business_id_fkey FOREIGN KEY (business_id) REFERENCES b2b.businesses(id);


--
-- Name: quotes quotes_delivery_state_fkey; Type: FK CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.quotes
    ADD CONSTRAINT quotes_delivery_state_fkey FOREIGN KEY (delivery_state) REFERENCES identity.nigerian_states(code);


--
-- Name: quotes quotes_requested_by_fkey; Type: FK CONSTRAINT; Schema: b2b; Owner: -
--

ALTER TABLE ONLY b2b.quotes
    ADD CONSTRAINT quotes_requested_by_fkey FOREIGN KEY (requested_by) REFERENCES identity.users(id);


--
-- Name: car_documents car_documents_car_listing_id_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_documents
    ADD CONSTRAINT car_documents_car_listing_id_fkey FOREIGN KEY (car_listing_id) REFERENCES cars.car_listings(id) ON DELETE CASCADE;


--
-- Name: car_documents car_documents_file_id_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_documents
    ADD CONSTRAINT car_documents_file_id_fkey FOREIGN KEY (file_id) REFERENCES platform.files(id);


--
-- Name: car_documents car_documents_verified_by_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_documents
    ADD CONSTRAINT car_documents_verified_by_fkey FOREIGN KEY (verified_by) REFERENCES identity.users(id);


--
-- Name: car_images car_images_car_listing_id_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_images
    ADD CONSTRAINT car_images_car_listing_id_fkey FOREIGN KEY (car_listing_id) REFERENCES cars.car_listings(id) ON DELETE CASCADE;


--
-- Name: car_images car_images_file_id_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_images
    ADD CONSTRAINT car_images_file_id_fkey FOREIGN KEY (file_id) REFERENCES platform.files(id);


--
-- Name: car_inspections car_inspections_car_listing_id_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_inspections
    ADD CONSTRAINT car_inspections_car_listing_id_fkey FOREIGN KEY (car_listing_id) REFERENCES cars.car_listings(id) ON DELETE CASCADE;


--
-- Name: car_inspections car_inspections_inspector_id_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_inspections
    ADD CONSTRAINT car_inspections_inspector_id_fkey FOREIGN KEY (inspector_id) REFERENCES identity.staff_members(user_id);


--
-- Name: car_inspections car_inspections_report_file_id_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_inspections
    ADD CONSTRAINT car_inspections_report_file_id_fkey FOREIGN KEY (report_file_id) REFERENCES platform.files(id);


--
-- Name: car_listings car_listings_seller_id_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_listings
    ADD CONSTRAINT car_listings_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id);


--
-- Name: car_listings car_listings_state_code_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.car_listings
    ADD CONSTRAINT car_listings_state_code_fkey FOREIGN KEY (state_code) REFERENCES identity.nigerian_states(code);


--
-- Name: financing_enquiries financing_enquiries_car_listing_id_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.financing_enquiries
    ADD CONSTRAINT financing_enquiries_car_listing_id_fkey FOREIGN KEY (car_listing_id) REFERENCES cars.car_listings(id);


--
-- Name: financing_enquiries financing_enquiries_customer_user_id_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.financing_enquiries
    ADD CONSTRAINT financing_enquiries_customer_user_id_fkey FOREIGN KEY (customer_user_id) REFERENCES identity.users(id);


--
-- Name: viewing_bookings viewing_bookings_car_listing_id_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.viewing_bookings
    ADD CONSTRAINT viewing_bookings_car_listing_id_fkey FOREIGN KEY (car_listing_id) REFERENCES cars.car_listings(id);


--
-- Name: viewing_bookings viewing_bookings_customer_user_id_fkey; Type: FK CONSTRAINT; Schema: cars; Owner: -
--

ALTER TABLE ONLY cars.viewing_bookings
    ADD CONSTRAINT viewing_bookings_customer_user_id_fkey FOREIGN KEY (customer_user_id) REFERENCES identity.users(id);


--
-- Name: attribute_definitions attribute_definitions_category_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.attribute_definitions
    ADD CONSTRAINT attribute_definitions_category_id_fkey FOREIGN KEY (category_id) REFERENCES catalog.categories(id) ON DELETE CASCADE;


--
-- Name: brands brands_logo_file_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.brands
    ADD CONSTRAINT brands_logo_file_id_fkey FOREIGN KEY (logo_file_id) REFERENCES platform.files(id);


--
-- Name: categories categories_parent_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.categories
    ADD CONSTRAINT categories_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES catalog.categories(id);


--
-- Name: image_hashes image_hashes_product_image_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.image_hashes
    ADD CONSTRAINT image_hashes_product_image_id_fkey FOREIGN KEY (product_image_id) REFERENCES catalog.product_images(id) ON DELETE CASCADE;


--
-- Name: listing_price_history listing_price_history_changed_by_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.listing_price_history
    ADD CONSTRAINT listing_price_history_changed_by_fkey FOREIGN KEY (changed_by) REFERENCES identity.users(id);


--
-- Name: listing_price_history listing_price_history_listing_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.listing_price_history
    ADD CONSTRAINT listing_price_history_listing_id_fkey FOREIGN KEY (listing_id) REFERENCES catalog.listings(id) ON DELETE CASCADE;


--
-- Name: listing_reviews listing_reviews_listing_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.listing_reviews
    ADD CONSTRAINT listing_reviews_listing_id_fkey FOREIGN KEY (listing_id) REFERENCES catalog.listings(id) ON DELETE CASCADE;


--
-- Name: listing_reviews listing_reviews_reviewer_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.listing_reviews
    ADD CONSTRAINT listing_reviews_reviewer_id_fkey FOREIGN KEY (reviewer_id) REFERENCES identity.users(id);


--
-- Name: listings listings_seller_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.listings
    ADD CONSTRAINT listings_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id);


--
-- Name: listings listings_variant_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.listings
    ADD CONSTRAINT listings_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES catalog.product_variants(id);


--
-- Name: price_tiers price_tiers_listing_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.price_tiers
    ADD CONSTRAINT price_tiers_listing_id_fkey FOREIGN KEY (listing_id) REFERENCES catalog.listings(id) ON DELETE CASCADE;


--
-- Name: product_attribute_values product_attribute_values_attribute_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_attribute_values
    ADD CONSTRAINT product_attribute_values_attribute_id_fkey FOREIGN KEY (attribute_id) REFERENCES catalog.attribute_definitions(id) ON DELETE CASCADE;


--
-- Name: product_attribute_values product_attribute_values_product_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_attribute_values
    ADD CONSTRAINT product_attribute_values_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: product_compatibility product_compatibility_accessory_product_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_compatibility
    ADD CONSTRAINT product_compatibility_accessory_product_id_fkey FOREIGN KEY (accessory_product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: product_compatibility product_compatibility_device_product_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_compatibility
    ADD CONSTRAINT product_compatibility_device_product_id_fkey FOREIGN KEY (device_product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: product_images product_images_file_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_images
    ADD CONSTRAINT product_images_file_id_fkey FOREIGN KEY (file_id) REFERENCES platform.files(id);


--
-- Name: product_images product_images_product_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_images
    ADD CONSTRAINT product_images_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: product_images product_images_variant_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_images
    ADD CONSTRAINT product_images_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES catalog.product_variants(id) ON DELETE CASCADE;


--
-- Name: product_offer_summary product_offer_summary_best_listing_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_offer_summary
    ADD CONSTRAINT product_offer_summary_best_listing_id_fkey FOREIGN KEY (best_listing_id) REFERENCES catalog.listings(id) ON DELETE SET NULL;


--
-- Name: product_offer_summary product_offer_summary_brand_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_offer_summary
    ADD CONSTRAINT product_offer_summary_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES catalog.brands(id);


--
-- Name: product_offer_summary product_offer_summary_product_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_offer_summary
    ADD CONSTRAINT product_offer_summary_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: product_reviews product_reviews_order_line_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_reviews
    ADD CONSTRAINT product_reviews_order_line_id_fkey FOREIGN KEY (order_line_id) REFERENCES sales.order_lines(id);


--
-- Name: product_reviews product_reviews_product_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_reviews
    ADD CONSTRAINT product_reviews_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: product_reviews product_reviews_user_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_reviews
    ADD CONSTRAINT product_reviews_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: product_suggestions product_suggestions_product_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_suggestions
    ADD CONSTRAINT product_suggestions_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id);


--
-- Name: product_suggestions product_suggestions_reviewed_by_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_suggestions
    ADD CONSTRAINT product_suggestions_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES identity.users(id);


--
-- Name: product_suggestions product_suggestions_seller_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_suggestions
    ADD CONSTRAINT product_suggestions_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id);


--
-- Name: product_variants product_variants_product_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.product_variants
    ADD CONSTRAINT product_variants_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: products products_brand_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.products
    ADD CONSTRAINT products_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES catalog.brands(id);


--
-- Name: products products_category_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.products
    ADD CONSTRAINT products_category_id_fkey FOREIGN KEY (category_id) REFERENCES catalog.categories(id);


--
-- Name: products products_created_by_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.products
    ADD CONSTRAINT products_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: review_images review_images_file_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.review_images
    ADD CONSTRAINT review_images_file_id_fkey FOREIGN KEY (file_id) REFERENCES platform.files(id);


--
-- Name: review_images review_images_review_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.review_images
    ADD CONSTRAINT review_images_review_id_fkey FOREIGN KEY (review_id) REFERENCES catalog.product_reviews(id) ON DELETE CASCADE;


--
-- Name: saved_items saved_items_product_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.saved_items
    ADD CONSTRAINT saved_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: saved_items saved_items_user_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.saved_items
    ADD CONSTRAINT saved_items_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: seller_ratings seller_ratings_fulfilment_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.seller_ratings
    ADD CONSTRAINT seller_ratings_fulfilment_id_fkey FOREIGN KEY (fulfilment_id) REFERENCES sales.fulfilments(id);


--
-- Name: seller_ratings seller_ratings_seller_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.seller_ratings
    ADD CONSTRAINT seller_ratings_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id);


--
-- Name: seller_ratings seller_ratings_user_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.seller_ratings
    ADD CONSTRAINT seller_ratings_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: stock_alerts stock_alerts_product_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.stock_alerts
    ADD CONSTRAINT stock_alerts_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: stock_alerts stock_alerts_user_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.stock_alerts
    ADD CONSTRAINT stock_alerts_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: banners banners_image_file_id_fkey; Type: FK CONSTRAINT; Schema: content; Owner: -
--

ALTER TABLE ONLY content.banners
    ADD CONSTRAINT banners_image_file_id_fkey FOREIGN KEY (image_file_id) REFERENCES platform.files(id);


--
-- Name: cms_pages cms_pages_approved_by_fkey; Type: FK CONSTRAINT; Schema: content; Owner: -
--

ALTER TABLE ONLY content.cms_pages
    ADD CONSTRAINT cms_pages_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES identity.users(id);


--
-- Name: cms_pages cms_pages_updated_by_fkey; Type: FK CONSTRAINT; Schema: content; Owner: -
--

ALTER TABLE ONLY content.cms_pages
    ADD CONSTRAINT cms_pages_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES identity.users(id);


--
-- Name: accounting_periods accounting_periods_closed_by_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.accounting_periods
    ADD CONSTRAINT accounting_periods_closed_by_fkey FOREIGN KEY (closed_by) REFERENCES identity.users(id);


--
-- Name: commission_rules commission_rules_category_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.commission_rules
    ADD CONSTRAINT commission_rules_category_id_fkey FOREIGN KEY (category_id) REFERENCES catalog.categories(id);


--
-- Name: commission_rules commission_rules_created_by_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.commission_rules
    ADD CONSTRAINT commission_rules_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: commission_rules commission_rules_seller_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.commission_rules
    ADD CONSTRAINT commission_rules_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id);


--
-- Name: invoices invoices_business_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.invoices
    ADD CONSTRAINT invoices_business_id_fkey FOREIGN KEY (business_id) REFERENCES b2b.businesses(id);


--
-- Name: invoices invoices_order_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.invoices
    ADD CONSTRAINT invoices_order_id_fkey FOREIGN KEY (order_id) REFERENCES sales.orders(id);


--
-- Name: invoices invoices_pdf_file_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.invoices
    ADD CONSTRAINT invoices_pdf_file_id_fkey FOREIGN KEY (pdf_file_id) REFERENCES platform.files(id);


--
-- Name: ledger_accounts ledger_accounts_seller_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_accounts
    ADD CONSTRAINT ledger_accounts_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id);


--
-- Name: ledger_accounts ledger_accounts_user_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_accounts
    ADD CONSTRAINT ledger_accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: ledger_entries ledger_entries_account_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_entries
    ADD CONSTRAINT ledger_entries_account_id_fkey FOREIGN KEY (account_id) REFERENCES finance.ledger_accounts(id);


--
-- Name: ledger_entries ledger_entries_journal_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_entries
    ADD CONSTRAINT ledger_entries_journal_id_fkey FOREIGN KEY (journal_id) REFERENCES finance.ledger_journals(id);


--
-- Name: ledger_journals ledger_journals_posted_by_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_journals
    ADD CONSTRAINT ledger_journals_posted_by_fkey FOREIGN KEY (posted_by) REFERENCES identity.users(id);


--
-- Name: ledger_journals ledger_journals_reverses_journal_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.ledger_journals
    ADD CONSTRAINT ledger_journals_reverses_journal_id_fkey FOREIGN KEY (reverses_journal_id) REFERENCES finance.ledger_journals(id);


--
-- Name: payout_batches payout_batches_approved_by_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payout_batches
    ADD CONSTRAINT payout_batches_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES identity.users(id);


--
-- Name: payout_batches payout_batches_prepared_by_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payout_batches
    ADD CONSTRAINT payout_batches_prepared_by_fkey FOREIGN KEY (prepared_by) REFERENCES identity.users(id);


--
-- Name: payout_holds payout_holds_created_by_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payout_holds
    ADD CONSTRAINT payout_holds_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: payout_holds payout_holds_seller_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payout_holds
    ADD CONSTRAINT payout_holds_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id);


--
-- Name: payout_items payout_items_fulfilment_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payout_items
    ADD CONSTRAINT payout_items_fulfilment_id_fkey FOREIGN KEY (fulfilment_id) REFERENCES sales.fulfilments(id);


--
-- Name: payout_items payout_items_payout_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payout_items
    ADD CONSTRAINT payout_items_payout_id_fkey FOREIGN KEY (payout_id) REFERENCES finance.payouts(id) ON DELETE CASCADE;


--
-- Name: payouts payouts_bank_account_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payouts
    ADD CONSTRAINT payouts_bank_account_id_fkey FOREIGN KEY (bank_account_id) REFERENCES sellers.seller_bank_accounts(id);


--
-- Name: payouts payouts_batch_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payouts
    ADD CONSTRAINT payouts_batch_id_fkey FOREIGN KEY (batch_id) REFERENCES finance.payout_batches(id);


--
-- Name: payouts payouts_seller_id_fkey; Type: FK CONSTRAINT; Schema: finance; Owner: -
--

ALTER TABLE ONLY finance.payouts
    ADD CONSTRAINT payouts_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id);


--
-- Name: manifest_parcels manifest_parcels_manifest_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.manifest_parcels
    ADD CONSTRAINT manifest_parcels_manifest_id_fkey FOREIGN KEY (manifest_id) REFERENCES fulfilment.manifests(id) ON DELETE CASCADE;


--
-- Name: manifest_parcels manifest_parcels_parcel_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.manifest_parcels
    ADD CONSTRAINT manifest_parcels_parcel_id_fkey FOREIGN KEY (parcel_id) REFERENCES fulfilment.parcels(id);


--
-- Name: manifests manifests_carrier_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.manifests
    ADD CONSTRAINT manifests_carrier_id_fkey FOREIGN KEY (carrier_id) REFERENCES fulfilment.carriers(id);


--
-- Name: manifests manifests_created_by_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.manifests
    ADD CONSTRAINT manifests_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: manifests manifests_rider_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.manifests
    ADD CONSTRAINT manifests_rider_id_fkey FOREIGN KEY (rider_id) REFERENCES logistics.riders(user_id);


--
-- Name: manifests manifests_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.manifests
    ADD CONSTRAINT manifests_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: pack_records pack_records_fulfilment_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pack_records
    ADD CONSTRAINT pack_records_fulfilment_id_fkey FOREIGN KEY (fulfilment_id) REFERENCES sales.fulfilments(id);


--
-- Name: pack_records pack_records_packed_by_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pack_records
    ADD CONSTRAINT pack_records_packed_by_fkey FOREIGN KEY (packed_by) REFERENCES identity.users(id);


--
-- Name: pack_records pack_records_parcel_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pack_records
    ADD CONSTRAINT pack_records_parcel_id_fkey FOREIGN KEY (parcel_id) REFERENCES fulfilment.parcels(id) ON DELETE CASCADE;


--
-- Name: parcels parcels_fulfilment_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.parcels
    ADD CONSTRAINT parcels_fulfilment_id_fkey FOREIGN KEY (fulfilment_id) REFERENCES sales.fulfilments(id) ON DELETE CASCADE;


--
-- Name: pick_list_items pick_list_items_bin_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pick_list_items
    ADD CONSTRAINT pick_list_items_bin_id_fkey FOREIGN KEY (bin_id) REFERENCES inventory.bin_locations(id);


--
-- Name: pick_list_items pick_list_items_device_unit_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pick_list_items
    ADD CONSTRAINT pick_list_items_device_unit_id_fkey FOREIGN KEY (device_unit_id) REFERENCES inventory.device_units(id);


--
-- Name: pick_list_items pick_list_items_fulfilment_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pick_list_items
    ADD CONSTRAINT pick_list_items_fulfilment_id_fkey FOREIGN KEY (fulfilment_id) REFERENCES sales.fulfilments(id);


--
-- Name: pick_list_items pick_list_items_order_line_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pick_list_items
    ADD CONSTRAINT pick_list_items_order_line_id_fkey FOREIGN KEY (order_line_id) REFERENCES sales.order_lines(id);


--
-- Name: pick_list_items pick_list_items_pick_list_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pick_list_items
    ADD CONSTRAINT pick_list_items_pick_list_id_fkey FOREIGN KEY (pick_list_id) REFERENCES fulfilment.pick_lists(id) ON DELETE CASCADE;


--
-- Name: pick_lists pick_lists_picker_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pick_lists
    ADD CONSTRAINT pick_lists_picker_id_fkey FOREIGN KEY (picker_id) REFERENCES identity.staff_members(user_id);


--
-- Name: pick_lists pick_lists_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.pick_lists
    ADD CONSTRAINT pick_lists_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: seller_sla_events seller_sla_events_fulfilment_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.seller_sla_events
    ADD CONSTRAINT seller_sla_events_fulfilment_id_fkey FOREIGN KEY (fulfilment_id) REFERENCES sales.fulfilments(id);


--
-- Name: seller_sla_events seller_sla_events_seller_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.seller_sla_events
    ADD CONSTRAINT seller_sla_events_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id);


--
-- Name: shipment_events shipment_events_shipment_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.shipment_events
    ADD CONSTRAINT shipment_events_shipment_id_fkey FOREIGN KEY (shipment_id) REFERENCES fulfilment.shipments(id) ON DELETE CASCADE;


--
-- Name: shipments shipments_carrier_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.shipments
    ADD CONSTRAINT shipments_carrier_id_fkey FOREIGN KEY (carrier_id) REFERENCES fulfilment.carriers(id);


--
-- Name: shipments shipments_fulfilment_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.shipments
    ADD CONSTRAINT shipments_fulfilment_id_fkey FOREIGN KEY (fulfilment_id) REFERENCES sales.fulfilments(id);


--
-- Name: shipments shipments_label_file_id_fkey; Type: FK CONSTRAINT; Schema: fulfilment; Owner: -
--

ALTER TABLE ONLY fulfilment.shipments
    ADD CONSTRAINT shipments_label_file_id_fkey FOREIGN KEY (label_file_id) REFERENCES platform.files(id);


--
-- Name: addresses addresses_business_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.addresses
    ADD CONSTRAINT addresses_business_id_fkey FOREIGN KEY (business_id) REFERENCES b2b.businesses(id) ON DELETE CASCADE;


--
-- Name: addresses addresses_state_code_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.addresses
    ADD CONSTRAINT addresses_state_code_fkey FOREIGN KEY (state_code) REFERENCES identity.nigerian_states(code);


--
-- Name: addresses addresses_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.addresses
    ADD CONSTRAINT addresses_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: auth_events auth_events_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE identity.auth_events
    ADD CONSTRAINT auth_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: consents consents_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.consents
    ADD CONSTRAINT consents_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: nigerian_lgas nigerian_lgas_state_code_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.nigerian_lgas
    ADD CONSTRAINT nigerian_lgas_state_code_fkey FOREIGN KEY (state_code) REFERENCES identity.nigerian_states(code);


--
-- Name: privacy_requests privacy_requests_file_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.privacy_requests
    ADD CONSTRAINT privacy_requests_file_id_fkey FOREIGN KEY (file_id) REFERENCES platform.files(id);


--
-- Name: privacy_requests privacy_requests_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.privacy_requests
    ADD CONSTRAINT privacy_requests_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: role_permissions role_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.role_permissions
    ADD CONSTRAINT role_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES identity.permissions(id) ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_role_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES identity.roles(id) ON DELETE CASCADE;


--
-- Name: staff_invites staff_invites_created_by_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.staff_invites
    ADD CONSTRAINT staff_invites_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: staff_invites staff_invites_staff_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.staff_invites
    ADD CONSTRAINT staff_invites_staff_user_id_fkey FOREIGN KEY (staff_user_id) REFERENCES identity.staff_members(user_id) ON DELETE CASCADE;


--
-- Name: staff_members staff_members_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.staff_members
    ADD CONSTRAINT staff_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: staff_role_scopes staff_role_scopes_user_id_role_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.staff_role_scopes
    ADD CONSTRAINT staff_role_scopes_user_id_role_id_fkey FOREIGN KEY (user_id, role_id) REFERENCES identity.staff_roles(user_id, role_id) ON DELETE CASCADE;


--
-- Name: staff_roles staff_roles_granted_by_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.staff_roles
    ADD CONSTRAINT staff_roles_granted_by_fkey FOREIGN KEY (granted_by) REFERENCES identity.users(id);


--
-- Name: staff_roles staff_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.staff_roles
    ADD CONSTRAINT staff_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES identity.roles(id) ON DELETE CASCADE;


--
-- Name: staff_roles staff_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.staff_roles
    ADD CONSTRAINT staff_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.staff_members(user_id) ON DELETE CASCADE;


--
-- Name: user_mfa_factors user_mfa_factors_encryption_key_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.user_mfa_factors
    ADD CONSTRAINT user_mfa_factors_encryption_key_id_fkey FOREIGN KEY (encryption_key_id) REFERENCES platform.encryption_keys(id);


--
-- Name: user_mfa_factors user_mfa_factors_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.user_mfa_factors
    ADD CONSTRAINT user_mfa_factors_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: user_recovery_codes user_recovery_codes_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.user_recovery_codes
    ADD CONSTRAINT user_recovery_codes_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: user_sessions user_sessions_replaced_by_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.user_sessions
    ADD CONSTRAINT user_sessions_replaced_by_fkey FOREIGN KEY (replaced_by) REFERENCES identity.user_sessions(id);


--
-- Name: user_sessions user_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.user_sessions
    ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: verification_codes verification_codes_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.verification_codes
    ADD CONSTRAINT verification_codes_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: webauthn_credentials webauthn_credentials_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.webauthn_credentials
    ADD CONSTRAINT webauthn_credentials_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: bin_locations bin_locations_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.bin_locations
    ADD CONSTRAINT bin_locations_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id) ON DELETE CASCADE;


--
-- Name: device_units device_units_owner_seller_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.device_units
    ADD CONSTRAINT device_units_owner_seller_id_fkey FOREIGN KEY (owner_seller_id) REFERENCES sellers.sellers(id);


--
-- Name: device_units device_units_variant_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.device_units
    ADD CONSTRAINT device_units_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES catalog.product_variants(id);


--
-- Name: device_units device_units_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.device_units
    ADD CONSTRAINT device_units_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: inventory_costs inventory_costs_owner_seller_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_costs
    ADD CONSTRAINT inventory_costs_owner_seller_id_fkey FOREIGN KEY (owner_seller_id) REFERENCES sellers.sellers(id);


--
-- Name: inventory_costs inventory_costs_variant_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_costs
    ADD CONSTRAINT inventory_costs_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES catalog.product_variants(id);


--
-- Name: inventory_costs inventory_costs_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_costs
    ADD CONSTRAINT inventory_costs_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: inventory_count_lines inventory_count_lines_bin_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_count_lines
    ADD CONSTRAINT inventory_count_lines_bin_id_fkey FOREIGN KEY (bin_id) REFERENCES inventory.bin_locations(id);


--
-- Name: inventory_count_lines inventory_count_lines_count_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_count_lines
    ADD CONSTRAINT inventory_count_lines_count_id_fkey FOREIGN KEY (count_id) REFERENCES inventory.inventory_counts(id) ON DELETE CASCADE;


--
-- Name: inventory_count_lines inventory_count_lines_owner_seller_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_count_lines
    ADD CONSTRAINT inventory_count_lines_owner_seller_id_fkey FOREIGN KEY (owner_seller_id) REFERENCES sellers.sellers(id);


--
-- Name: inventory_count_lines inventory_count_lines_variant_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_count_lines
    ADD CONSTRAINT inventory_count_lines_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES catalog.product_variants(id);


--
-- Name: inventory_counts inventory_counts_approved_by_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_counts
    ADD CONSTRAINT inventory_counts_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES identity.users(id);


--
-- Name: inventory_counts inventory_counts_counted_by_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_counts
    ADD CONSTRAINT inventory_counts_counted_by_fkey FOREIGN KEY (counted_by) REFERENCES identity.users(id);


--
-- Name: inventory_counts inventory_counts_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_counts
    ADD CONSTRAINT inventory_counts_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: inventory_levels inventory_levels_bin_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_levels
    ADD CONSTRAINT inventory_levels_bin_id_fkey FOREIGN KEY (bin_id) REFERENCES inventory.bin_locations(id);


--
-- Name: inventory_levels inventory_levels_owner_seller_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_levels
    ADD CONSTRAINT inventory_levels_owner_seller_id_fkey FOREIGN KEY (owner_seller_id) REFERENCES sellers.sellers(id);


--
-- Name: inventory_levels inventory_levels_variant_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_levels
    ADD CONSTRAINT inventory_levels_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES catalog.product_variants(id);


--
-- Name: inventory_levels inventory_levels_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.inventory_levels
    ADD CONSTRAINT inventory_levels_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: stock_movements stock_movements_created_by_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_movements
    ADD CONSTRAINT stock_movements_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: stock_movements stock_movements_device_unit_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_movements
    ADD CONSTRAINT stock_movements_device_unit_id_fkey FOREIGN KEY (device_unit_id) REFERENCES inventory.device_units(id);


--
-- Name: stock_movements stock_movements_owner_seller_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_movements
    ADD CONSTRAINT stock_movements_owner_seller_id_fkey FOREIGN KEY (owner_seller_id) REFERENCES sellers.sellers(id);


--
-- Name: stock_movements stock_movements_variant_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_movements
    ADD CONSTRAINT stock_movements_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES catalog.product_variants(id);


--
-- Name: stock_movements stock_movements_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_movements
    ADD CONSTRAINT stock_movements_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: stock_reservations stock_reservations_listing_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_reservations
    ADD CONSTRAINT stock_reservations_listing_id_fkey FOREIGN KEY (listing_id) REFERENCES catalog.listings(id);


--
-- Name: stock_reservations stock_reservations_order_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_reservations
    ADD CONSTRAINT stock_reservations_order_id_fkey FOREIGN KEY (order_id) REFERENCES sales.orders(id) ON DELETE CASCADE;


--
-- Name: stock_reservations stock_reservations_order_line_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_reservations
    ADD CONSTRAINT stock_reservations_order_line_id_fkey FOREIGN KEY (order_line_id) REFERENCES sales.order_lines(id) ON DELETE CASCADE;


--
-- Name: stock_reservations stock_reservations_owner_seller_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_reservations
    ADD CONSTRAINT stock_reservations_owner_seller_id_fkey FOREIGN KEY (owner_seller_id) REFERENCES sellers.sellers(id);


--
-- Name: stock_reservations stock_reservations_variant_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_reservations
    ADD CONSTRAINT stock_reservations_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES catalog.product_variants(id);


--
-- Name: stock_reservations stock_reservations_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_reservations
    ADD CONSTRAINT stock_reservations_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: stock_transfer_lines stock_transfer_lines_owner_seller_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_transfer_lines
    ADD CONSTRAINT stock_transfer_lines_owner_seller_id_fkey FOREIGN KEY (owner_seller_id) REFERENCES sellers.sellers(id);


--
-- Name: stock_transfer_lines stock_transfer_lines_transfer_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_transfer_lines
    ADD CONSTRAINT stock_transfer_lines_transfer_id_fkey FOREIGN KEY (transfer_id) REFERENCES inventory.stock_transfers(id) ON DELETE CASCADE;


--
-- Name: stock_transfer_lines stock_transfer_lines_variant_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_transfer_lines
    ADD CONSTRAINT stock_transfer_lines_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES catalog.product_variants(id);


--
-- Name: stock_transfers stock_transfers_created_by_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_transfers
    ADD CONSTRAINT stock_transfers_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: stock_transfers stock_transfers_from_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_transfers
    ADD CONSTRAINT stock_transfers_from_warehouse_id_fkey FOREIGN KEY (from_warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: stock_transfers stock_transfers_to_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.stock_transfers
    ADD CONSTRAINT stock_transfers_to_warehouse_id_fkey FOREIGN KEY (to_warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: warehouses warehouses_state_code_fkey; Type: FK CONSTRAINT; Schema: inventory; Owner: -
--

ALTER TABLE ONLY inventory.warehouses
    ADD CONSTRAINT warehouses_state_code_fkey FOREIGN KEY (state_code) REFERENCES identity.nigerian_states(code);


--
-- Name: delivery_events delivery_events_created_by_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_events
    ADD CONSTRAINT delivery_events_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: delivery_events delivery_events_delivery_job_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_events
    ADD CONSTRAINT delivery_events_delivery_job_id_fkey FOREIGN KEY (delivery_job_id) REFERENCES logistics.delivery_jobs(id) ON DELETE CASCADE;


--
-- Name: delivery_events delivery_events_file_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_events
    ADD CONSTRAINT delivery_events_file_id_fkey FOREIGN KEY (file_id) REFERENCES platform.files(id);


--
-- Name: delivery_jobs delivery_jobs_assigned_by_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_jobs
    ADD CONSTRAINT delivery_jobs_assigned_by_fkey FOREIGN KEY (assigned_by) REFERENCES identity.users(id);


--
-- Name: delivery_jobs delivery_jobs_fulfilment_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_jobs
    ADD CONSTRAINT delivery_jobs_fulfilment_id_fkey FOREIGN KEY (fulfilment_id) REFERENCES sales.fulfilments(id);


--
-- Name: delivery_jobs delivery_jobs_pickup_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_jobs
    ADD CONSTRAINT delivery_jobs_pickup_warehouse_id_fkey FOREIGN KEY (pickup_warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: delivery_jobs delivery_jobs_proof_file_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_jobs
    ADD CONSTRAINT delivery_jobs_proof_file_id_fkey FOREIGN KEY (proof_file_id) REFERENCES platform.files(id);


--
-- Name: delivery_jobs delivery_jobs_return_request_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_jobs
    ADD CONSTRAINT delivery_jobs_return_request_id_fkey FOREIGN KEY (return_request_id) REFERENCES aftersales.return_requests(id);


--
-- Name: delivery_jobs delivery_jobs_rider_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_jobs
    ADD CONSTRAINT delivery_jobs_rider_id_fkey FOREIGN KEY (rider_id) REFERENCES logistics.riders(user_id);


--
-- Name: delivery_jobs delivery_jobs_zone_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_jobs
    ADD CONSTRAINT delivery_jobs_zone_id_fkey FOREIGN KEY (zone_id) REFERENCES logistics.delivery_zones(id);


--
-- Name: delivery_rates delivery_rates_zone_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_rates
    ADD CONSTRAINT delivery_rates_zone_id_fkey FOREIGN KEY (zone_id) REFERENCES logistics.delivery_zones(id) ON DELETE CASCADE;


--
-- Name: delivery_zones delivery_zones_state_code_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.delivery_zones
    ADD CONSTRAINT delivery_zones_state_code_fkey FOREIGN KEY (state_code) REFERENCES identity.nigerian_states(code);


--
-- Name: rider_devices rider_devices_approved_by_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_devices
    ADD CONSTRAINT rider_devices_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES identity.users(id);


--
-- Name: rider_devices rider_devices_rider_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_devices
    ADD CONSTRAINT rider_devices_rider_id_fkey FOREIGN KEY (rider_id) REFERENCES logistics.riders(user_id) ON DELETE CASCADE;


--
-- Name: rider_location_pings rider_location_pings_rider_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE logistics.rider_location_pings
    ADD CONSTRAINT rider_location_pings_rider_id_fkey FOREIGN KEY (rider_id) REFERENCES logistics.riders(user_id);


--
-- Name: rider_locations rider_locations_job_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_locations
    ADD CONSTRAINT rider_locations_job_id_fkey FOREIGN KEY (job_id) REFERENCES logistics.delivery_jobs(id) ON DELETE SET NULL;


--
-- Name: rider_locations rider_locations_rider_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_locations
    ADD CONSTRAINT rider_locations_rider_id_fkey FOREIGN KEY (rider_id) REFERENCES logistics.riders(user_id) ON DELETE CASCADE;


--
-- Name: rider_shifts rider_shifts_device_session_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_shifts
    ADD CONSTRAINT rider_shifts_device_session_id_fkey FOREIGN KEY (device_session_id) REFERENCES identity.user_sessions(id);


--
-- Name: rider_shifts rider_shifts_rider_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_shifts
    ADD CONSTRAINT rider_shifts_rider_id_fkey FOREIGN KEY (rider_id) REFERENCES logistics.riders(user_id);


--
-- Name: rider_zones rider_zones_rider_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_zones
    ADD CONSTRAINT rider_zones_rider_id_fkey FOREIGN KEY (rider_id) REFERENCES logistics.riders(user_id) ON DELETE CASCADE;


--
-- Name: rider_zones rider_zones_zone_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.rider_zones
    ADD CONSTRAINT rider_zones_zone_id_fkey FOREIGN KEY (zone_id) REFERENCES logistics.delivery_zones(id) ON DELETE CASCADE;


--
-- Name: riders riders_home_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.riders
    ADD CONSTRAINT riders_home_warehouse_id_fkey FOREIGN KEY (home_warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: riders riders_user_id_fkey; Type: FK CONSTRAINT; Schema: logistics; Owner: -
--

ALTER TABLE ONLY logistics.riders
    ADD CONSTRAINT riders_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.staff_members(user_id);


--
-- Name: campaign_recipients campaign_recipients_campaign_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.campaign_recipients
    ADD CONSTRAINT campaign_recipients_campaign_id_fkey FOREIGN KEY (campaign_id) REFERENCES marketing.campaigns(id) ON DELETE CASCADE;


--
-- Name: campaign_recipients campaign_recipients_user_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.campaign_recipients
    ADD CONSTRAINT campaign_recipients_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: campaign_recipients campaign_recipients_variant_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.campaign_recipients
    ADD CONSTRAINT campaign_recipients_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES marketing.campaign_variants(id);


--
-- Name: campaign_variants campaign_variants_campaign_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.campaign_variants
    ADD CONSTRAINT campaign_variants_campaign_id_fkey FOREIGN KEY (campaign_id) REFERENCES marketing.campaigns(id) ON DELETE CASCADE;


--
-- Name: campaign_variants campaign_variants_template_version_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.campaign_variants
    ADD CONSTRAINT campaign_variants_template_version_id_fkey FOREIGN KEY (template_version_id) REFERENCES messaging.template_versions(id);


--
-- Name: campaigns campaigns_approved_by_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.campaigns
    ADD CONSTRAINT campaigns_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES identity.users(id);


--
-- Name: campaigns campaigns_created_by_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.campaigns
    ADD CONSTRAINT campaigns_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: campaigns campaigns_segment_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.campaigns
    ADD CONSTRAINT campaigns_segment_id_fkey FOREIGN KEY (segment_id) REFERENCES marketing.segments(id);


--
-- Name: coupon_codes coupon_codes_coupon_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupon_codes
    ADD CONSTRAINT coupon_codes_coupon_id_fkey FOREIGN KEY (coupon_id) REFERENCES marketing.coupons(id) ON DELETE CASCADE;


--
-- Name: coupon_codes coupon_codes_issued_to_user_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupon_codes
    ADD CONSTRAINT coupon_codes_issued_to_user_id_fkey FOREIGN KEY (issued_to_user_id) REFERENCES identity.users(id);


--
-- Name: coupon_codes coupon_codes_notification_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupon_codes
    ADD CONSTRAINT coupon_codes_notification_id_fkey FOREIGN KEY (notification_id) REFERENCES messaging.notifications(id);


--
-- Name: coupon_redemptions coupon_redemptions_coupon_code_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupon_redemptions
    ADD CONSTRAINT coupon_redemptions_coupon_code_id_fkey FOREIGN KEY (coupon_code_id) REFERENCES marketing.coupon_codes(id);


--
-- Name: coupon_redemptions coupon_redemptions_coupon_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupon_redemptions
    ADD CONSTRAINT coupon_redemptions_coupon_id_fkey FOREIGN KEY (coupon_id) REFERENCES marketing.coupons(id);


--
-- Name: coupon_redemptions coupon_redemptions_order_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupon_redemptions
    ADD CONSTRAINT coupon_redemptions_order_id_fkey FOREIGN KEY (order_id) REFERENCES sales.orders(id);


--
-- Name: coupon_redemptions coupon_redemptions_user_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupon_redemptions
    ADD CONSTRAINT coupon_redemptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: coupons coupons_promotion_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.coupons
    ADD CONSTRAINT coupons_promotion_id_fkey FOREIGN KEY (promotion_id) REFERENCES marketing.promotions(id) ON DELETE CASCADE;


--
-- Name: flash_claims flash_claims_order_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.flash_claims
    ADD CONSTRAINT flash_claims_order_id_fkey FOREIGN KEY (order_id) REFERENCES sales.orders(id) ON DELETE CASCADE;


--
-- Name: flash_claims flash_claims_promotion_id_listing_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.flash_claims
    ADD CONSTRAINT flash_claims_promotion_id_listing_id_fkey FOREIGN KEY (promotion_id, listing_id) REFERENCES marketing.promotion_listing_prices(promotion_id, listing_id);


--
-- Name: flash_claims flash_claims_user_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.flash_claims
    ADD CONSTRAINT flash_claims_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: journey_enrollments journey_enrollments_journey_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.journey_enrollments
    ADD CONSTRAINT journey_enrollments_journey_id_fkey FOREIGN KEY (journey_id) REFERENCES marketing.journeys(id) ON DELETE CASCADE;


--
-- Name: journey_enrollments journey_enrollments_user_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.journey_enrollments
    ADD CONSTRAINT journey_enrollments_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: journey_steps journey_steps_journey_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.journey_steps
    ADD CONSTRAINT journey_steps_journey_id_fkey FOREIGN KEY (journey_id) REFERENCES marketing.journeys(id) ON DELETE CASCADE;


--
-- Name: journeys journeys_created_by_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.journeys
    ADD CONSTRAINT journeys_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: promotion_listing_prices promotion_listing_prices_listing_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.promotion_listing_prices
    ADD CONSTRAINT promotion_listing_prices_listing_id_fkey FOREIGN KEY (listing_id) REFERENCES catalog.listings(id);


--
-- Name: promotion_listing_prices promotion_listing_prices_promotion_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.promotion_listing_prices
    ADD CONSTRAINT promotion_listing_prices_promotion_id_fkey FOREIGN KEY (promotion_id) REFERENCES marketing.promotions(id) ON DELETE CASCADE;


--
-- Name: promotion_targets promotion_targets_category_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.promotion_targets
    ADD CONSTRAINT promotion_targets_category_id_fkey FOREIGN KEY (category_id) REFERENCES catalog.categories(id);


--
-- Name: promotion_targets promotion_targets_listing_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.promotion_targets
    ADD CONSTRAINT promotion_targets_listing_id_fkey FOREIGN KEY (listing_id) REFERENCES catalog.listings(id);


--
-- Name: promotion_targets promotion_targets_promotion_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.promotion_targets
    ADD CONSTRAINT promotion_targets_promotion_id_fkey FOREIGN KEY (promotion_id) REFERENCES marketing.promotions(id) ON DELETE CASCADE;


--
-- Name: promotions promotions_created_by_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.promotions
    ADD CONSTRAINT promotions_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: segments segments_created_by_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.segments
    ADD CONSTRAINT segments_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: tracked_links tracked_links_campaign_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.tracked_links
    ADD CONSTRAINT tracked_links_campaign_id_fkey FOREIGN KEY (campaign_id) REFERENCES marketing.campaigns(id) ON DELETE CASCADE;


--
-- Name: tracked_links tracked_links_notification_id_fkey; Type: FK CONSTRAINT; Schema: marketing; Owner: -
--

ALTER TABLE ONLY marketing.tracked_links
    ADD CONSTRAINT tracked_links_notification_id_fkey FOREIGN KEY (notification_id) REFERENCES messaging.notifications(id) ON DELETE CASCADE;


--
-- Name: message_events message_events_notification_id_fkey; Type: FK CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE messaging.message_events
    ADD CONSTRAINT message_events_notification_id_fkey FOREIGN KEY (notification_id) REFERENCES messaging.notifications(id) ON DELETE CASCADE;


--
-- Name: notification_preferences notification_preferences_user_id_fkey; Type: FK CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notification_preferences
    ADD CONSTRAINT notification_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: notification_settings notification_settings_user_id_fkey; Type: FK CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notification_settings
    ADD CONSTRAINT notification_settings_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_campaign_id_fkey; Type: FK CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notifications
    ADD CONSTRAINT notifications_campaign_id_fkey FOREIGN KEY (campaign_id) REFERENCES marketing.campaigns(id);


--
-- Name: notifications notifications_journey_enrollment_id_fkey; Type: FK CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notifications
    ADD CONSTRAINT notifications_journey_enrollment_id_fkey FOREIGN KEY (journey_enrollment_id) REFERENCES marketing.journey_enrollments(id);


--
-- Name: notifications notifications_template_version_id_fkey; Type: FK CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notifications
    ADD CONSTRAINT notifications_template_version_id_fkey FOREIGN KEY (template_version_id) REFERENCES messaging.template_versions(id);


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: push_tokens push_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.push_tokens
    ADD CONSTRAINT push_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: template_versions template_versions_created_by_fkey; Type: FK CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.template_versions
    ADD CONSTRAINT template_versions_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: template_versions template_versions_template_id_fkey; Type: FK CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.template_versions
    ADD CONSTRAINT template_versions_template_id_fkey FOREIGN KEY (template_id) REFERENCES messaging.message_templates(id) ON DELETE CASCADE;


--
-- Name: dispute_evidence dispute_evidence_dispute_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.dispute_evidence
    ADD CONSTRAINT dispute_evidence_dispute_id_fkey FOREIGN KEY (dispute_id) REFERENCES payments.disputes(id) ON DELETE CASCADE;


--
-- Name: dispute_evidence dispute_evidence_file_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.dispute_evidence
    ADD CONSTRAINT dispute_evidence_file_id_fkey FOREIGN KEY (file_id) REFERENCES platform.files(id);


--
-- Name: disputes disputes_decided_by_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.disputes
    ADD CONSTRAINT disputes_decided_by_fkey FOREIGN KEY (decided_by) REFERENCES identity.users(id);


--
-- Name: disputes disputes_payment_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.disputes
    ADD CONSTRAINT disputes_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES payments.payments(id);


--
-- Name: payment_events payment_events_payment_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.payment_events
    ADD CONSTRAINT payment_events_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES payments.payments(id);


--
-- Name: payment_events payment_events_payout_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.payment_events
    ADD CONSTRAINT payment_events_payout_id_fkey FOREIGN KEY (payout_id) REFERENCES finance.payouts(id);


--
-- Name: payment_events payment_events_refund_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.payment_events
    ADD CONSTRAINT payment_events_refund_id_fkey FOREIGN KEY (refund_id) REFERENCES payments.refunds(id);


--
-- Name: payments payments_invoice_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.payments
    ADD CONSTRAINT payments_invoice_id_fkey FOREIGN KEY (invoice_id) REFERENCES finance.invoices(id);


--
-- Name: payments payments_order_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.payments
    ADD CONSTRAINT payments_order_id_fkey FOREIGN KEY (order_id) REFERENCES sales.orders(id);


--
-- Name: reconciliation_exceptions reconciliation_exceptions_owner_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.reconciliation_exceptions
    ADD CONSTRAINT reconciliation_exceptions_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES identity.users(id);


--
-- Name: reconciliation_exceptions reconciliation_exceptions_resolved_by_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.reconciliation_exceptions
    ADD CONSTRAINT reconciliation_exceptions_resolved_by_fkey FOREIGN KEY (resolved_by) REFERENCES identity.users(id);


--
-- Name: refunds refunds_approved_by_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.refunds
    ADD CONSTRAINT refunds_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES identity.users(id);


--
-- Name: refunds refunds_order_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.refunds
    ADD CONSTRAINT refunds_order_id_fkey FOREIGN KEY (order_id) REFERENCES sales.orders(id);


--
-- Name: refunds refunds_payment_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.refunds
    ADD CONSTRAINT refunds_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES payments.payments(id);


--
-- Name: refunds refunds_requested_by_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.refunds
    ADD CONSTRAINT refunds_requested_by_fkey FOREIGN KEY (requested_by) REFERENCES identity.users(id);


--
-- Name: refunds refunds_return_request_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.refunds
    ADD CONSTRAINT refunds_return_request_id_fkey FOREIGN KEY (return_request_id) REFERENCES aftersales.return_requests(id);


--
-- Name: settlement_lines settlement_lines_payment_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.settlement_lines
    ADD CONSTRAINT settlement_lines_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES payments.payments(id);


--
-- Name: settlement_lines settlement_lines_report_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.settlement_lines
    ADD CONSTRAINT settlement_lines_report_id_fkey FOREIGN KEY (report_id) REFERENCES payments.settlement_reports(id) ON DELETE CASCADE;


--
-- Name: virtual_accounts virtual_accounts_business_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.virtual_accounts
    ADD CONSTRAINT virtual_accounts_business_id_fkey FOREIGN KEY (business_id) REFERENCES b2b.businesses(id);


--
-- Name: virtual_accounts virtual_accounts_order_id_fkey; Type: FK CONSTRAINT; Schema: payments; Owner: -
--

ALTER TABLE ONLY payments.virtual_accounts
    ADD CONSTRAINT virtual_accounts_order_id_fkey FOREIGN KEY (order_id) REFERENCES sales.orders(id);


--
-- Name: identity_links identity_links_user_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.identity_links
    ADD CONSTRAINT identity_links_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: rec_item_neighbours rec_item_neighbours_model_version_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_item_neighbours
    ADD CONSTRAINT rec_item_neighbours_model_version_id_fkey FOREIGN KEY (model_version_id) REFERENCES personalisation.rec_model_versions(id) ON DELETE CASCADE;


--
-- Name: rec_item_neighbours rec_item_neighbours_neighbour_product_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_item_neighbours
    ADD CONSTRAINT rec_item_neighbours_neighbour_product_id_fkey FOREIGN KEY (neighbour_product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: rec_item_neighbours rec_item_neighbours_product_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_item_neighbours
    ADD CONSTRAINT rec_item_neighbours_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: rec_item_vectors rec_item_vectors_model_version_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_item_vectors
    ADD CONSTRAINT rec_item_vectors_model_version_id_fkey FOREIGN KEY (model_version_id) REFERENCES personalisation.rec_model_versions(id) ON DELETE CASCADE;


--
-- Name: rec_item_vectors rec_item_vectors_product_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_item_vectors
    ADD CONSTRAINT rec_item_vectors_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: rec_overrides rec_overrides_anchor_product_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_overrides
    ADD CONSTRAINT rec_overrides_anchor_product_id_fkey FOREIGN KEY (anchor_product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: rec_overrides rec_overrides_created_by_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_overrides
    ADD CONSTRAINT rec_overrides_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: rec_overrides rec_overrides_product_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_overrides
    ADD CONSTRAINT rec_overrides_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: rec_popularity rec_popularity_category_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_popularity
    ADD CONSTRAINT rec_popularity_category_id_fkey FOREIGN KEY (category_id) REFERENCES catalog.categories(id) ON DELETE CASCADE;


--
-- Name: rec_popularity rec_popularity_product_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_popularity
    ADD CONSTRAINT rec_popularity_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: rec_popularity rec_popularity_state_code_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_popularity
    ADD CONSTRAINT rec_popularity_state_code_fkey FOREIGN KEY (state_code) REFERENCES identity.nigerian_states(code);


--
-- Name: rec_subject_candidates rec_subject_candidates_model_version_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_subject_candidates
    ADD CONSTRAINT rec_subject_candidates_model_version_id_fkey FOREIGN KEY (model_version_id) REFERENCES personalisation.rec_model_versions(id) ON DELETE CASCADE;


--
-- Name: rec_subject_candidates rec_subject_candidates_product_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_subject_candidates
    ADD CONSTRAINT rec_subject_candidates_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: rec_subject_candidates rec_subject_candidates_user_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE ONLY personalisation.rec_subject_candidates
    ADD CONSTRAINT rec_subject_candidates_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: user_events user_events_car_listing_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE personalisation.user_events
    ADD CONSTRAINT user_events_car_listing_id_fkey FOREIGN KEY (car_listing_id) REFERENCES cars.car_listings(id) ON DELETE SET NULL;


--
-- Name: user_events user_events_category_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE personalisation.user_events
    ADD CONSTRAINT user_events_category_id_fkey FOREIGN KEY (category_id) REFERENCES catalog.categories(id) ON DELETE SET NULL;


--
-- Name: user_events user_events_listing_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE personalisation.user_events
    ADD CONSTRAINT user_events_listing_id_fkey FOREIGN KEY (listing_id) REFERENCES catalog.listings(id) ON DELETE SET NULL;


--
-- Name: user_events user_events_product_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE personalisation.user_events
    ADD CONSTRAINT user_events_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE SET NULL;


--
-- Name: user_events user_events_user_id_fkey; Type: FK CONSTRAINT; Schema: personalisation; Owner: -
--

ALTER TABLE personalisation.user_events
    ADD CONSTRAINT user_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: audit_log audit_log_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: platform; Owner: -
--

ALTER TABLE ONLY platform.audit_log
    ADD CONSTRAINT audit_log_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES identity.users(id);


--
-- Name: feature_flags feature_flags_updated_by_fkey; Type: FK CONSTRAINT; Schema: platform; Owner: -
--

ALTER TABLE ONLY platform.feature_flags
    ADD CONSTRAINT feature_flags_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES identity.users(id);


--
-- Name: files files_uploaded_by_fkey; Type: FK CONSTRAINT; Schema: platform; Owner: -
--

ALTER TABLE ONLY platform.files
    ADD CONSTRAINT files_uploaded_by_fkey FOREIGN KEY (uploaded_by) REFERENCES identity.users(id);


--
-- Name: idempotency_keys idempotency_keys_user_id_fkey; Type: FK CONSTRAINT; Schema: platform; Owner: -
--

ALTER TABLE ONLY platform.idempotency_keys
    ADD CONSTRAINT idempotency_keys_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: pos_shifts pos_shifts_cashier_id_fkey; Type: FK CONSTRAINT; Schema: pos; Owner: -
--

ALTER TABLE ONLY pos.pos_shifts
    ADD CONSTRAINT pos_shifts_cashier_id_fkey FOREIGN KEY (cashier_id) REFERENCES identity.staff_members(user_id);


--
-- Name: pos_shifts pos_shifts_terminal_id_fkey; Type: FK CONSTRAINT; Schema: pos; Owner: -
--

ALTER TABLE ONLY pos.pos_shifts
    ADD CONSTRAINT pos_shifts_terminal_id_fkey FOREIGN KEY (terminal_id) REFERENCES pos.pos_terminals(id);


--
-- Name: pos_terminals pos_terminals_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: pos; Owner: -
--

ALTER TABLE ONLY pos.pos_terminals
    ADD CONSTRAINT pos_terminals_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: goods_receipts goods_receipts_purchase_order_id_fkey; Type: FK CONSTRAINT; Schema: purchasing; Owner: -
--

ALTER TABLE ONLY purchasing.goods_receipts
    ADD CONSTRAINT goods_receipts_purchase_order_id_fkey FOREIGN KEY (purchase_order_id) REFERENCES purchasing.purchase_orders(id);


--
-- Name: goods_receipts goods_receipts_received_by_fkey; Type: FK CONSTRAINT; Schema: purchasing; Owner: -
--

ALTER TABLE ONLY purchasing.goods_receipts
    ADD CONSTRAINT goods_receipts_received_by_fkey FOREIGN KEY (received_by) REFERENCES identity.users(id);


--
-- Name: purchase_order_lines purchase_order_lines_purchase_order_id_fkey; Type: FK CONSTRAINT; Schema: purchasing; Owner: -
--

ALTER TABLE ONLY purchasing.purchase_order_lines
    ADD CONSTRAINT purchase_order_lines_purchase_order_id_fkey FOREIGN KEY (purchase_order_id) REFERENCES purchasing.purchase_orders(id) ON DELETE CASCADE;


--
-- Name: purchase_order_lines purchase_order_lines_variant_id_fkey; Type: FK CONSTRAINT; Schema: purchasing; Owner: -
--

ALTER TABLE ONLY purchasing.purchase_order_lines
    ADD CONSTRAINT purchase_order_lines_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES catalog.product_variants(id);


--
-- Name: purchase_orders purchase_orders_approved_by_fkey; Type: FK CONSTRAINT; Schema: purchasing; Owner: -
--

ALTER TABLE ONLY purchasing.purchase_orders
    ADD CONSTRAINT purchase_orders_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES identity.users(id);


--
-- Name: purchase_orders purchase_orders_created_by_fkey; Type: FK CONSTRAINT; Schema: purchasing; Owner: -
--

ALTER TABLE ONLY purchasing.purchase_orders
    ADD CONSTRAINT purchase_orders_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: purchase_orders purchase_orders_supplier_id_fkey; Type: FK CONSTRAINT; Schema: purchasing; Owner: -
--

ALTER TABLE ONLY purchasing.purchase_orders
    ADD CONSTRAINT purchase_orders_supplier_id_fkey FOREIGN KEY (supplier_id) REFERENCES purchasing.suppliers(id);


--
-- Name: purchase_orders purchase_orders_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: purchasing; Owner: -
--

ALTER TABLE ONLY purchasing.purchase_orders
    ADD CONSTRAINT purchase_orders_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: device_blocklist device_blocklist_created_by_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.device_blocklist
    ADD CONSTRAINT device_blocklist_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: enforcement_actions enforcement_actions_created_by_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.enforcement_actions
    ADD CONSTRAINT enforcement_actions_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: enforcement_actions enforcement_actions_decided_by_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.enforcement_actions
    ADD CONSTRAINT enforcement_actions_decided_by_fkey FOREIGN KEY (decided_by) REFERENCES identity.users(id);


--
-- Name: enforcement_actions enforcement_actions_seller_id_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.enforcement_actions
    ADD CONSTRAINT enforcement_actions_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id);


--
-- Name: listing_reports listing_reports_decided_by_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.listing_reports
    ADD CONSTRAINT listing_reports_decided_by_fkey FOREIGN KEY (decided_by) REFERENCES identity.users(id);


--
-- Name: listing_reports listing_reports_listing_id_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.listing_reports
    ADD CONSTRAINT listing_reports_listing_id_fkey FOREIGN KEY (listing_id) REFERENCES catalog.listings(id) ON DELETE CASCADE;


--
-- Name: listing_reports listing_reports_reporter_id_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.listing_reports
    ADD CONSTRAINT listing_reports_reporter_id_fkey FOREIGN KEY (reporter_id) REFERENCES identity.users(id);


--
-- Name: review_reports review_reports_decided_by_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.review_reports
    ADD CONSTRAINT review_reports_decided_by_fkey FOREIGN KEY (decided_by) REFERENCES identity.users(id);


--
-- Name: review_reports review_reports_reporter_id_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.review_reports
    ADD CONSTRAINT review_reports_reporter_id_fkey FOREIGN KEY (reporter_id) REFERENCES identity.users(id);


--
-- Name: review_reports review_reports_review_id_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.review_reports
    ADD CONSTRAINT review_reports_review_id_fkey FOREIGN KEY (review_id) REFERENCES catalog.product_reviews(id) ON DELETE CASCADE;


--
-- Name: risk_cases risk_cases_assigned_to_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_cases
    ADD CONSTRAINT risk_cases_assigned_to_fkey FOREIGN KEY (assigned_to) REFERENCES identity.staff_members(user_id);


--
-- Name: risk_cases risk_cases_resolved_by_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_cases
    ADD CONSTRAINT risk_cases_resolved_by_fkey FOREIGN KEY (resolved_by) REFERENCES identity.users(id);


--
-- Name: risk_decisions risk_decisions_user_id_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE risk.risk_decisions
    ADD CONSTRAINT risk_decisions_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: risk_links risk_links_entity_a_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_links
    ADD CONSTRAINT risk_links_entity_a_fkey FOREIGN KEY (entity_a) REFERENCES risk.risk_entities(id) ON DELETE CASCADE;


--
-- Name: risk_links risk_links_entity_b_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_links
    ADD CONSTRAINT risk_links_entity_b_fkey FOREIGN KEY (entity_b) REFERENCES risk.risk_entities(id) ON DELETE CASCADE;


--
-- Name: risk_lists risk_lists_created_by_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_lists
    ADD CONSTRAINT risk_lists_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: risk_rule_sets risk_rule_sets_created_by_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_rule_sets
    ADD CONSTRAINT risk_rule_sets_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: risk_rule_sets risk_rule_sets_published_by_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_rule_sets
    ADD CONSTRAINT risk_rule_sets_published_by_fkey FOREIGN KEY (published_by) REFERENCES identity.users(id);


--
-- Name: risk_rules risk_rules_rule_set_id_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.risk_rules
    ADD CONSTRAINT risk_rules_rule_set_id_fkey FOREIGN KEY (rule_set_id) REFERENCES risk.risk_rule_sets(id) ON DELETE CASCADE;


--
-- Name: seller_metrics_daily seller_metrics_daily_seller_id_fkey; Type: FK CONSTRAINT; Schema: risk; Owner: -
--

ALTER TABLE ONLY risk.seller_metrics_daily
    ADD CONSTRAINT seller_metrics_daily_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id) ON DELETE CASCADE;


--
-- Name: cart_items cart_items_cart_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.cart_items
    ADD CONSTRAINT cart_items_cart_id_fkey FOREIGN KEY (cart_id) REFERENCES sales.carts(id) ON DELETE CASCADE;


--
-- Name: cart_items cart_items_listing_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.cart_items
    ADD CONSTRAINT cart_items_listing_id_fkey FOREIGN KEY (listing_id) REFERENCES catalog.listings(id);


--
-- Name: carts carts_business_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.carts
    ADD CONSTRAINT carts_business_id_fkey FOREIGN KEY (business_id) REFERENCES b2b.businesses(id);


--
-- Name: carts carts_user_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.carts
    ADD CONSTRAINT carts_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: fulfilments fulfilments_collection_store_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.fulfilments
    ADD CONSTRAINT fulfilments_collection_store_id_fkey FOREIGN KEY (collection_store_id) REFERENCES inventory.warehouses(id);


--
-- Name: fulfilments fulfilments_order_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.fulfilments
    ADD CONSTRAINT fulfilments_order_id_fkey FOREIGN KEY (order_id) REFERENCES sales.orders(id) ON DELETE CASCADE;


--
-- Name: fulfilments fulfilments_seller_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.fulfilments
    ADD CONSTRAINT fulfilments_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id);


--
-- Name: fulfilments fulfilments_warehouse_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.fulfilments
    ADD CONSTRAINT fulfilments_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id);


--
-- Name: order_lines order_lines_device_unit_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.order_lines
    ADD CONSTRAINT order_lines_device_unit_id_fkey FOREIGN KEY (device_unit_id) REFERENCES inventory.device_units(id);


--
-- Name: order_lines order_lines_fulfilment_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.order_lines
    ADD CONSTRAINT order_lines_fulfilment_id_fkey FOREIGN KEY (fulfilment_id) REFERENCES sales.fulfilments(id) ON DELETE CASCADE;


--
-- Name: order_lines order_lines_listing_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.order_lines
    ADD CONSTRAINT order_lines_listing_id_fkey FOREIGN KEY (listing_id) REFERENCES catalog.listings(id);


--
-- Name: order_lines order_lines_order_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.order_lines
    ADD CONSTRAINT order_lines_order_id_fkey FOREIGN KEY (order_id) REFERENCES sales.orders(id) ON DELETE CASCADE;


--
-- Name: order_lines order_lines_seller_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.order_lines
    ADD CONSTRAINT order_lines_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id);


--
-- Name: order_lines order_lines_variant_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.order_lines
    ADD CONSTRAINT order_lines_variant_id_fkey FOREIGN KEY (variant_id) REFERENCES catalog.product_variants(id);


--
-- Name: order_status_history order_status_history_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.order_status_history
    ADD CONSTRAINT order_status_history_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES identity.users(id);


--
-- Name: order_status_history order_status_history_fulfilment_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.order_status_history
    ADD CONSTRAINT order_status_history_fulfilment_id_fkey FOREIGN KEY (fulfilment_id) REFERENCES sales.fulfilments(id) ON DELETE CASCADE;


--
-- Name: order_status_history order_status_history_order_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.order_status_history
    ADD CONSTRAINT order_status_history_order_id_fkey FOREIGN KEY (order_id) REFERENCES sales.orders(id) ON DELETE CASCADE;


--
-- Name: orders orders_attributed_notification_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.orders
    ADD CONSTRAINT orders_attributed_notification_id_fkey FOREIGN KEY (attributed_notification_id) REFERENCES messaging.notifications(id);


--
-- Name: orders orders_business_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.orders
    ADD CONSTRAINT orders_business_id_fkey FOREIGN KEY (business_id) REFERENCES b2b.businesses(id);


--
-- Name: orders orders_customer_user_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.orders
    ADD CONSTRAINT orders_customer_user_id_fkey FOREIGN KEY (customer_user_id) REFERENCES identity.users(id);


--
-- Name: orders orders_pos_shift_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.orders
    ADD CONSTRAINT orders_pos_shift_id_fkey FOREIGN KEY (pos_shift_id) REFERENCES pos.pos_shifts(id);


--
-- Name: orders orders_quote_id_fkey; Type: FK CONSTRAINT; Schema: sales; Owner: -
--

ALTER TABLE ONLY sales.orders
    ADD CONSTRAINT orders_quote_id_fkey FOREIGN KEY (quote_id) REFERENCES b2b.quotes(id);


--
-- Name: catalogue_gaps catalogue_gaps_owner_id_fkey; Type: FK CONSTRAINT; Schema: search; Owner: -
--

ALTER TABLE ONLY search.catalogue_gaps
    ADD CONSTRAINT catalogue_gaps_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES identity.users(id);


--
-- Name: search_pins search_pins_created_by_fkey; Type: FK CONSTRAINT; Schema: search; Owner: -
--

ALTER TABLE ONLY search.search_pins
    ADD CONSTRAINT search_pins_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: search_pins search_pins_product_id_fkey; Type: FK CONSTRAINT; Schema: search; Owner: -
--

ALTER TABLE ONLY search.search_pins
    ADD CONSTRAINT search_pins_product_id_fkey FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;


--
-- Name: search_queries search_queries_user_id_fkey; Type: FK CONSTRAINT; Schema: search; Owner: -
--

ALTER TABLE search.search_queries
    ADD CONSTRAINT search_queries_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: search_redirects search_redirects_created_by_fkey; Type: FK CONSTRAINT; Schema: search; Owner: -
--

ALTER TABLE ONLY search.search_redirects
    ADD CONSTRAINT search_redirects_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: search_synonyms search_synonyms_created_by_fkey; Type: FK CONSTRAINT; Schema: search; Owner: -
--

ALTER TABLE ONLY search.search_synonyms
    ADD CONSTRAINT search_synonyms_created_by_fkey FOREIGN KEY (created_by) REFERENCES identity.users(id);


--
-- Name: kyc_documents kyc_documents_file_id_fkey; Type: FK CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.kyc_documents
    ADD CONSTRAINT kyc_documents_file_id_fkey FOREIGN KEY (file_id) REFERENCES platform.files(id);


--
-- Name: kyc_documents kyc_documents_submission_id_fkey; Type: FK CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.kyc_documents
    ADD CONSTRAINT kyc_documents_submission_id_fkey FOREIGN KEY (submission_id) REFERENCES sellers.seller_kyc_submissions(id) ON DELETE CASCADE;


--
-- Name: seller_bank_accounts seller_bank_accounts_encryption_key_id_fkey; Type: FK CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.seller_bank_accounts
    ADD CONSTRAINT seller_bank_accounts_encryption_key_id_fkey FOREIGN KEY (encryption_key_id) REFERENCES platform.encryption_keys(id);


--
-- Name: seller_bank_accounts seller_bank_accounts_seller_id_fkey; Type: FK CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.seller_bank_accounts
    ADD CONSTRAINT seller_bank_accounts_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id) ON DELETE CASCADE;


--
-- Name: seller_kyc_submissions seller_kyc_submissions_encryption_key_id_fkey; Type: FK CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.seller_kyc_submissions
    ADD CONSTRAINT seller_kyc_submissions_encryption_key_id_fkey FOREIGN KEY (encryption_key_id) REFERENCES platform.encryption_keys(id);


--
-- Name: seller_kyc_submissions seller_kyc_submissions_reviewed_by_fkey; Type: FK CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.seller_kyc_submissions
    ADD CONSTRAINT seller_kyc_submissions_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES identity.users(id);


--
-- Name: seller_kyc_submissions seller_kyc_submissions_seller_id_fkey; Type: FK CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.seller_kyc_submissions
    ADD CONSTRAINT seller_kyc_submissions_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id) ON DELETE CASCADE;


--
-- Name: seller_members seller_members_seller_id_fkey; Type: FK CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.seller_members
    ADD CONSTRAINT seller_members_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id) ON DELETE CASCADE;


--
-- Name: seller_members seller_members_user_id_fkey; Type: FK CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.seller_members
    ADD CONSTRAINT seller_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: sellers sellers_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.sellers
    ADD CONSTRAINT sellers_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES identity.users(id);


--
-- Name: sellers sellers_state_code_fkey; Type: FK CONSTRAINT; Schema: sellers; Owner: -
--

ALTER TABLE ONLY sellers.sellers
    ADD CONSTRAINT sellers_state_code_fkey FOREIGN KEY (state_code) REFERENCES identity.nigerian_states(code);


--
-- Name: message_attachments message_attachments_file_id_fkey; Type: FK CONSTRAINT; Schema: support; Owner: -
--

ALTER TABLE ONLY support.message_attachments
    ADD CONSTRAINT message_attachments_file_id_fkey FOREIGN KEY (file_id) REFERENCES platform.files(id);


--
-- Name: message_attachments message_attachments_message_id_fkey; Type: FK CONSTRAINT; Schema: support; Owner: -
--

ALTER TABLE ONLY support.message_attachments
    ADD CONSTRAINT message_attachments_message_id_fkey FOREIGN KEY (message_id) REFERENCES support.ticket_messages(id) ON DELETE CASCADE;


--
-- Name: support_tickets support_tickets_assigned_to_fkey; Type: FK CONSTRAINT; Schema: support; Owner: -
--

ALTER TABLE ONLY support.support_tickets
    ADD CONSTRAINT support_tickets_assigned_to_fkey FOREIGN KEY (assigned_to) REFERENCES identity.staff_members(user_id);


--
-- Name: support_tickets support_tickets_customer_user_id_fkey; Type: FK CONSTRAINT; Schema: support; Owner: -
--

ALTER TABLE ONLY support.support_tickets
    ADD CONSTRAINT support_tickets_customer_user_id_fkey FOREIGN KEY (customer_user_id) REFERENCES identity.users(id);


--
-- Name: support_tickets support_tickets_order_id_fkey; Type: FK CONSTRAINT; Schema: support; Owner: -
--

ALTER TABLE ONLY support.support_tickets
    ADD CONSTRAINT support_tickets_order_id_fkey FOREIGN KEY (order_id) REFERENCES sales.orders(id);


--
-- Name: support_tickets support_tickets_seller_id_fkey; Type: FK CONSTRAINT; Schema: support; Owner: -
--

ALTER TABLE ONLY support.support_tickets
    ADD CONSTRAINT support_tickets_seller_id_fkey FOREIGN KEY (seller_id) REFERENCES sellers.sellers(id);


--
-- Name: ticket_messages ticket_messages_author_user_id_fkey; Type: FK CONSTRAINT; Schema: support; Owner: -
--

ALTER TABLE ONLY support.ticket_messages
    ADD CONSTRAINT ticket_messages_author_user_id_fkey FOREIGN KEY (author_user_id) REFERENCES identity.users(id);


--
-- Name: ticket_messages ticket_messages_ticket_id_fkey; Type: FK CONSTRAINT; Schema: support; Owner: -
--

ALTER TABLE ONLY support.ticket_messages
    ADD CONSTRAINT ticket_messages_ticket_id_fkey FOREIGN KEY (ticket_id) REFERENCES support.support_tickets(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--


