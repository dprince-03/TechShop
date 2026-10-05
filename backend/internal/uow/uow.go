// Package uow runs one database transaction across modules ("unit of work"). Modules share the
// same sqlc queries bound to the transaction; outbox events, audit entries and River jobs are
// written in the same commit as the data change. No network calls happen inside a transaction.
package uow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Event is a domain event written to platform.outbox_events.
type Event struct {
	AggregateType string
	AggregateID   uuid.UUID
	Type          string
	Payload       any
}

// Tx is the transaction handle given to modules.
type Tx struct {
	Q      *store.Queries
	PgTx   pgx.Tx
	ctx    context.Context
	events []Event
	jobs   []river.InsertManyParams
	notify []notification
	after  []func()
}

type notification struct{ channel, payload string }

// Emit records a domain event; it is written to the outbox in the same commit.
func (t *Tx) Emit(aggType string, aggID uuid.UUID, eventType string, payload any) {
	t.events = append(t.events, Event{AggregateType: aggType, AggregateID: aggID, Type: eventType, Payload: payload})
}

// Enqueue inserts a River job in the same transaction (it only runs if the commit succeeds).
func (t *Tx) Enqueue(args river.JobArgs, opts *river.InsertOpts) {
	t.jobs = append(t.jobs, river.InsertManyParams{Args: args, InsertOpts: opts})
}

// Notify sends pg_notify on commit (realtime fan-out, cache invalidation).
func (t *Tx) Notify(channel string, payload any) {
	b, _ := json.Marshal(payload)
	t.notify = append(t.notify, notification{channel, string(b)})
}

// AfterCommit runs f after a successful commit (e.g. clearing in-process caches).
func (t *Tx) AfterCommit(f func()) { t.after = append(t.after, f) }

// Audit writes an audit_log row in this transaction, attributed to the request's principal.
func (t *Tx) Audit(action, entityType, entityID string, changes any) error {
	var actor *uuid.UUID
	kind := "system"
	if p := httpx.PrincipalFrom(t.ctx); p != nil {
		id := p.UserID
		actor = &id
		switch p.Audience {
		case "staff":
			kind = "staff"
		case "seller":
			kind = "seller"
		case "internal":
			kind = "service"
		default:
			kind = "user"
		}
	}
	meta := httpx.MetaFrom(t.ctx)
	var ch json.RawMessage
	if changes != nil {
		ch, _ = json.Marshal(changes)
	}
	var ip *netip.Addr
	if a, err := netip.ParseAddr(meta.IP); err == nil {
		ip = &a
	}
	return t.Q.PlatformInsertAudit(t.ctx, store.PlatformInsertAuditParams{
		ActorUserID: actor, ActorKind: kind, Action: action, EntityType: entityType, EntityID: entityID,
		Changes: ch, RequestID: strPtr(meta.RequestID), Ip: ip, UserAgent: strPtr(meta.UserAgent),
	})
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Runner opens transactions.
type Runner struct {
	Pool   *pgxpool.Pool
	River  *river.Client[pgx.Tx]
	Logger *slog.Logger
}

// Run executes fn in one transaction, retrying on serialization failures and deadlocks.
func (r *Runner) Run(ctx context.Context, fn func(*Tx) error) error {
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = r.once(ctx, fn)
		var pg *pgconn.PgError
		if errors.As(err, &pg) && (pg.Code == "40001" || pg.Code == "40P01") {
			time.Sleep(time.Duration(10*(attempt+1)) * time.Millisecond)
			continue
		}
		return err
	}
	return err
}

func (r *Runner) once(ctx context.Context, fn func(*Tx) error) error {
	pgTx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = pgTx.Rollback(context.WithoutCancel(ctx)) }()

	tx := &Tx{Q: store.New(pgTx), PgTx: pgTx, ctx: ctx}
	if err := fn(tx); err != nil {
		return err
	}
	for _, e := range tx.events {
		payload, err := json.Marshal(e.Payload)
		if err != nil {
			return fmt.Errorf("marshal event %s: %w", e.Type, err)
		}
		if err := tx.Q.PlatformInsertOutboxEvent(ctx, store.PlatformInsertOutboxEventParams{
			AggregateType: e.AggregateType, AggregateID: e.AggregateID, EventType: e.Type, Payload: payload,
		}); err != nil {
			return fmt.Errorf("insert outbox event: %w", err)
		}
	}
	if len(tx.jobs) > 0 {
		if r.River == nil {
			return errors.New("uow: jobs enqueued but no River client configured")
		}
		if _, err := r.River.InsertManyTx(ctx, pgTx, dedupeJobs(tx.jobs)); err != nil {
			return fmt.Errorf("insert jobs: %w", err)
		}
	}
	for _, n := range tx.notify {
		if _, err := pgTx.Exec(ctx, "select pg_notify($1, $2)", n.channel, n.payload); err != nil {
			return fmt.Errorf("notify: %w", err)
		}
	}
	if err := pgTx.Commit(ctx); err != nil {
		return err
	}
	for _, f := range tx.after {
		f()
	}
	return nil
}

// Read runs read-only queries outside a transaction.
func (r *Runner) Read() *store.Queries { return store.New(r.Pool) }

// Savepoint runs fn in a nested transaction. If fn fails only its changes roll back and the
// outer transaction carries on; on success its events, jobs and notifications join the outer ones.
func (t *Tx) Savepoint(ctx context.Context, fn func(*Tx) error) error {
	sp, err := t.PgTx.Begin(ctx)
	if err != nil {
		return err
	}
	sub := &Tx{Q: t.Q.WithTx(sp), PgTx: sp, ctx: ctx}
	if err := fn(sub); err != nil {
		_ = sp.Rollback(ctx)
		return err
	}
	if err := sp.Commit(ctx); err != nil {
		return err
	}
	t.events = append(t.events, sub.events...)
	t.jobs = append(t.jobs, sub.jobs...)
	t.notify = append(t.notify, sub.notify...)
	t.after = append(t.after, sub.after...)
	return nil
}

// dedupeJobs drops exact duplicates (same kind and args) from one commit. River's bulk insert
// upserts unique jobs, and Postgres refuses to upsert the same row twice in one statement.
func dedupeJobs(in []river.InsertManyParams) []river.InsertManyParams {
	seen := make(map[string]bool, len(in))
	out := in[:0:0]
	for _, j := range in {
		b, err := json.Marshal(j.Args)
		key := j.Args.Kind() + "|" + string(b)
		if err == nil && seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, j)
	}
	return out
}
