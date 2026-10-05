// Package postgres opens PostgreSQL connection pools (pgx). It knows nothing about TechShop:
// callers pass plain Options, so it can be reused by any service.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Options configures a pool. Zero values fall back to pgx defaults.
type Options struct {
	URL              string        // e.g. postgres://user:pass@postgres:5432/db?sslmode=…
	MaxConns         int32         // upper bound on open connections
	MinConns         int32         // connections kept warm
	StatementTimeout time.Duration // kills runaway queries (0 = no limit)
	PingTimeout      time.Duration // how long to wait for the first ping (default 5s)
}

// NewPool opens a pool, applies safety settings (statement and idle-transaction timeouts,
// UTC) and checks the connection with a ping before returning.
func NewPool(ctx context.Context, o Options) (*pgxpool.Pool, error) {
	pc, err := pgxpool.ParseConfig(o.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	if o.MaxConns > 0 {
		pc.MaxConns = o.MaxConns
	}
	if o.MinConns > 0 {
		pc.MinConns = o.MinConns
	}
	pc.MaxConnIdleTime = 5 * time.Minute
	if o.StatementTimeout > 0 {
		pc.ConnConfig.RuntimeParams["statement_timeout"] = fmt.Sprint(o.StatementTimeout.Milliseconds())
	}
	pc.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = "60000"
	pc.ConnConfig.RuntimeParams["timezone"] = "UTC"

	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	pingTimeout := o.PingTimeout
	if pingTimeout <= 0 {
		pingTimeout = 5 * time.Second
	}
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}
