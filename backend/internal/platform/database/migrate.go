package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/dprince-03/techshop/backend/db"
	"github.com/dprince-03/techshop/backend/internal/platform/config"
	"github.com/dprince-03/techshop/backend/pkg/database/postgres"
)

// MigrateWithOwner opens a short-lived connection as the table owner (DATABASE_MIGRATE_URL, or
// DATABASE_URL when unset) and applies pending migrations. No statement timeout: index builds
// and backfills can legitimately take longer than an API request.
func MigrateWithOwner(ctx context.Context, cfg *config.Config) error {
	pool, err := postgres.NewPool(ctx, postgres.Options{URL: cfg.MigrationURL(), MaxConns: 2})
	if err != nil {
		return err
	}
	defer pool.Close()
	return Migrate(ctx, pool)
}

// Migrate applies all pending migrations embedded in the binary.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer func() { _ = sqlDB.Close() }()

	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, mustSub(db.Migrations, "migrations"))
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
