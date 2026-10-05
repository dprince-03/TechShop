-- Reference data every environment needs: states, TechShop as the first-party seller, the chart of
-- accounts, payment provider health rows, open accounting periods, the admin role and default flags.
-- Not seeded here (on purpose): LGAs, public holidays, other staff roles (owner decision), dev test
-- accounts (dev-only seed script, when backend work starts).

-- +goose Up
insert into identity.nigerian_states (code, name) values
  ('AB', 'Abia'), ('AD', 'Adamawa'), ('AK', 'Akwa Ibom'), ('AN', 'Anambra'), ('BA', 'Bauchi'),
  ('BY', 'Bayelsa'), ('BE', 'Benue'), ('BO', 'Borno'), ('CR', 'Cross River'), ('DE', 'Delta'),
  ('EB', 'Ebonyi'), ('ED', 'Edo'), ('EK', 'Ekiti'), ('EN', 'Enugu'), ('FC', 'FCT (Abuja)'),
  ('GO', 'Gombe'), ('IM', 'Imo'), ('JI', 'Jigawa'), ('KD', 'Kaduna'), ('KN', 'Kano'),
  ('KT', 'Katsina'), ('KE', 'Kebbi'), ('KO', 'Kogi'), ('KW', 'Kwara'), ('LA', 'Lagos'),
  ('NA', 'Nasarawa'), ('NI', 'Niger'), ('OG', 'Ogun'), ('ON', 'Ondo'), ('OS', 'Osun'),
  ('OY', 'Oyo'), ('PL', 'Plateau'), ('RI', 'Rivers'), ('SO', 'Sokoto'), ('TA', 'Taraba'),
  ('YO', 'Yobe'), ('ZA', 'Zamfara');

-- TechShop itself, as the first-party seller (its own stock is sold through first_party listings).
insert into sellers.sellers (type, display_name, slug, status, state_code, city)
values ('first_party', 'TechShop', 'techshop', 'active', 'LA', 'Lagos');

-- Platform ledger accounts. Seller payables and customer store-credit accounts are created per seller/user.
insert into finance.ledger_accounts (code, name, kind) values
  ('clearing:paystack', 'Paystack clearing', 'asset'),
  ('clearing:opay', 'OPay clearing', 'asset'),
  ('clearing:moniepoint', 'Monnify (Moniepoint) clearing', 'asset'),
  ('bank:operating', 'Operating bank account', 'asset'),
  ('bank:payouts', 'Payouts bank account', 'asset'),
  ('receivable:b2b', 'B2B invoices receivable', 'asset'),
  ('asset:inventory', 'Inventory at cost', 'asset'),
  ('revenue:first_party_sales', 'TechShop own-stock sales', 'revenue'),
  ('revenue:commission', 'Marketplace commission', 'revenue'),
  ('revenue:delivery', 'Delivery fees', 'revenue'),
  ('liability:vat', 'VAT payable', 'liability'),
  ('liability:customer_refunds', 'Refunds owed to customers', 'liability'),
  ('liability:accounts_payable', 'Supplier invoices payable', 'liability'),
  ('expense:payment_fees', 'Payment provider fees', 'expense'),
  ('expense:cogs', 'Cost of goods sold', 'expense'),
  ('expense:chargebacks', 'Chargebacks lost', 'expense'),
  ('expense:write_offs', 'Bad debt and write-offs', 'expense'),
  ('expense:inventory_shrinkage', 'Stock count losses', 'expense');

insert into payments.provider_health (provider) values ('paystack'), ('opay'), ('moniepoint');

-- The current and next month are open for posting.
insert into finance.accounting_periods (month) values
  (date_trunc('month', now())::date),
  ((date_trunc('month', now()) + interval '1 month')::date)
on conflict do nothing;

-- Only the admin role is seeded; the rest of the role list is waiting for the owner (docs/identity-access.md §7.2).
insert into identity.roles (key, name, description, is_system)
values ('admin', 'Admin', 'Full access. Held by a small number of named people; every sign-in is alerted.', true);

insert into platform.feature_flags (key, description, enabled)
values ('recs.live_service', 'Call the Python live recommendation service (80 ms budget, Go fallback).', false);

-- +goose Down
delete from platform.feature_flags where key = 'recs.live_service';
delete from identity.roles where key = 'admin';
delete from finance.accounting_periods p
where p.status = 'open'
  and not exists (select 1 from finance.ledger_journals j where j.period = p.month);
delete from payments.provider_health where provider in ('paystack', 'opay', 'moniepoint');
delete from finance.ledger_accounts where code in (
  'clearing:paystack', 'clearing:opay', 'clearing:moniepoint', 'bank:operating', 'bank:payouts', 'receivable:b2b',
  'asset:inventory', 'revenue:first_party_sales', 'revenue:commission', 'revenue:delivery', 'liability:vat',
  'liability:customer_refunds', 'liability:accounts_payable', 'expense:payment_fees', 'expense:cogs',
  'expense:chargebacks', 'expense:write_offs', 'expense:inventory_shrinkage');
delete from sellers.sellers where type = 'first_party' and slug = 'techshop';
delete from identity.nigerian_states where code in (
  'AB', 'AD', 'AK', 'AN', 'BA', 'BY', 'BE', 'BO', 'CR', 'DE', 'EB', 'ED', 'EK', 'EN', 'FC', 'GO', 'IM', 'JI', 'KD',
  'KN', 'KT', 'KE', 'KO', 'KW', 'LA', 'NA', 'NI', 'OG', 'ON', 'OS', 'OY', 'PL', 'RI', 'SO', 'TA', 'YO', 'ZA');
