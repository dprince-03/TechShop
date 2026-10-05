// Package ratelimit counts hits in fixed windows stored in Postgres (identity.auth_rate_limits),
// so limits hold across API instances without Redis. Used for OTP sends and sign-in attempts.
package ratelimit

import (
	"context"
	"time"

	"github.com/dprince-03/techshop/backend/internal/store"
)

// Rule is a limit: at most Max hits per Window for a key prefix.
type Rule struct {
	Name   string
	Max    int32
	Window time.Duration
}

// Limiter checks and records hits.
type Limiter struct{ q *store.Queries }

// New creates a limiter.
func New(q *store.Queries) *Limiter { return &Limiter{q: q} }

// Allow records a hit for key under rule and reports whether it is within the limit.
func (l *Limiter) Allow(ctx context.Context, rule Rule, key string) (bool, error) {
	start := time.Now().Truncate(rule.Window)
	n, err := l.q.PlatformHitRateLimit(ctx, store.PlatformHitRateLimitParams{Key: rule.Name + ":" + key, WindowStart: start})
	if err != nil {
		return false, err
	}
	return n <= rule.Max, nil
}

// Count returns hits for key since `since` (without recording one).
func (l *Limiter) Count(ctx context.Context, rule Rule, key string, since time.Time) (int32, error) {
	return l.q.PlatformCountRateLimit(ctx, store.PlatformCountRateLimitParams{Key: rule.Name + ":" + key, WindowStart: since})
}
