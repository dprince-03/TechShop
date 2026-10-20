-- Creates (or re-passwords) the login the API and worker use: techshop_app, a member of the
-- techshop_rw group from migration 00024. Run after migrations, as the table owner:
--   make db-users APP_DB_PASSWORD=<16+ random characters>
-- The password is passed in as a psql variable; it is never stored in this repo.
\set ON_ERROR_STOP on

select format('create role techshop_app login password %L in role techshop_rw', :'app_password')
where not exists (select 1 from pg_roles where rolname = 'techshop_app') \gexec

select format('alter role techshop_app with login password %L', :'app_password')
where exists (select 1 from pg_roles where rolname = 'techshop_app') \gexec

grant techshop_rw to techshop_app;
