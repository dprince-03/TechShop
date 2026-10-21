package database

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dprince-03/techshop/backend/internal/platform/config"
)

// ErrPrivilegedRole means the API or worker is connected as a superuser or as the owner of the
// tables. Such a login can alter or drop tables and bypasses row-level security, so it is refused
// outside development (docs/database/RLS-POLICIES.md).
var ErrPrivilegedRole = errors.New("database login is a superuser or owns the tables; " +
	"connect DATABASE_URL as a member of techshop_rw and use DATABASE_MIGRATE_URL for migrations")

// CheckAppRole refuses a privileged login outside development and warns about it locally.
func CheckAppRole(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config, logger *slog.Logger) error {
	var privileged bool
	err := pool.QueryRow(ctx, `
		select r.rolsuper or r.rolbypassrls or exists (
			select 1 from pg_tables t where t.tableowner = current_user and t.schemaname in ('identity', 'sales', 'finance'))
		from pg_roles r where r.rolname = current_user`).Scan(&privileged)
	if err != nil {
		return err
	}
	if !privileged {
		return nil
	}
	if cfg.IsDevelopment() || cfg.Env == "test" {
		logger.Warn("database login is privileged; fine locally, refused in staging and production",
			slog.String("fix", "make db-users, then point DATABASE_URL at techshop_app"))
		return nil
	}
	return ErrPrivilegedRole
}
