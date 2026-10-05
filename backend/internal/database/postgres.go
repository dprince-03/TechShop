// Package database wires TechShop's config into the reusable Postgres client and runs migrations.
package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dprince-03/techshop/backend/internal/config"
	"github.com/dprince-03/techshop/backend/pkg/database/postgres"
)

// NewPool opens TechShop's Postgres pool from config (see pkg/database/postgres).
func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	return postgres.NewPool(ctx, postgres.Options{
		URL:              cfg.DatabaseURL,
		MaxConns:         cfg.DBMaxConns,
		MinConns:         cfg.DBMinConns,
		StatementTimeout: cfg.DBStatementTimeout,
	})
}
