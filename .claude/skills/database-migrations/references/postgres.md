# PostgreSQL Notes

## Row-Level Security for multi-tenancy
```sql
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON invoices
  USING (tenant_id = current_setting('app.tenant_id')::uuid);
-- app sets per transaction: SET LOCAL app.tenant_id = '<uuid>';
```
- App role must not be table owner or superuser (owners bypass RLS unless FORCE).

## updated_at trigger
```sql
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN NEW.updated_at = now(); RETURN NEW; END; $$ LANGUAGE plpgsql;
CREATE TRIGGER trg_invoices_updated_at BEFORE UPDATE ON invoices
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

## Roles
- `app_rw` (DML only), `app_migrator` (DDL), `app_ro` (read replicas/analytics). No app connection as `postgres`.

## Useful checks
- Unused indexes: `pg_stat_user_indexes WHERE idx_scan = 0`.
- Slow queries: enable `pg_stat_statements`.
- Bloat and vacuum: monitor autovacuum on high-churn tables.

## Backups
- PITR via WAL archiving (managed: RDS automated backups / Render PITR); test restores quarterly.

## Extensions commonly useful
- `pgcrypto` / `uuid-ossp` (UUIDs), `citext` (case-insensitive email), `pg_trgm` (fuzzy search), `pgaudit`.
