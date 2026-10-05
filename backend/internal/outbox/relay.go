// Package outbox relays committed domain events to subscribers. It wakes on pg_notify('outbox')
// with a 2-second poll fallback, locks unpublished events (FOR UPDATE SKIP LOCKED, so several
// workers are safe) and lets each subscriber enqueue jobs in the same transaction.
package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dprince-03/techshop/backend/internal/uow"
)

// Event is a published domain event.
type Event struct {
	ID            int64
	AggregateType string
	AggregateID   uuid.UUID
	Type          string
	Payload       json.RawMessage
}

// Handler reacts to an event inside the relay transaction (usually by enqueueing jobs).
type Handler func(ctx context.Context, tx *uow.Tx, ev Event) error

// Relay dispatches events by type.
type Relay struct {
	runner   *uow.Runner
	pool     *pgxpool.Pool
	logger   *slog.Logger
	handlers map[string][]Handler
}

// NewRelay creates a relay.
func NewRelay(runner *uow.Runner, pool *pgxpool.Pool, logger *slog.Logger) *Relay {
	return &Relay{runner: runner, pool: pool, logger: logger, handlers: map[string][]Handler{}}
}

// On subscribes a handler to an event type ("*" receives every event).
func (r *Relay) On(eventType string, h Handler) {
	r.handlers[eventType] = append(r.handlers[eventType], h)
}

// Run processes events until ctx is cancelled.
func (r *Relay) Run(ctx context.Context) {
	wake := make(chan struct{}, 1)
	go r.listen(ctx, wake)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		for {
			n, err := r.ProcessBatch(ctx)
			if err != nil {
				r.logger.Error("outbox batch failed", slog.Any("error", err))
				break
			}
			if n == 0 {
				break
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-wake:
		case <-ticker.C:
		}
	}
}

// ProcessBatch handles up to 100 events; it returns how many were processed.
func (r *Relay) ProcessBatch(ctx context.Context) (int, error) {
	var n int
	err := r.runner.Run(ctx, func(tx *uow.Tx) error {
		rows, err := tx.Q.PlatformLockUnpublishedEvents(ctx, 100)
		if err != nil {
			return err
		}
		ids := make([]int64, 0, len(rows))
		for _, row := range rows {
			ev := Event{ID: row.ID, AggregateType: row.AggregateType, AggregateID: row.AggregateID, Type: row.EventType, Payload: row.Payload}
			for _, h := range append(r.handlers[ev.Type], r.handlers["*"]...) {
				if err := h(ctx, tx, ev); err != nil {
					return fmt.Errorf("handler for %s: %w", ev.Type, err)
				}
			}
			ids = append(ids, row.ID)
		}
		n = len(ids)
		if n == 0 {
			return nil
		}
		return tx.Q.PlatformMarkEventsPublished(ctx, ids)
	})
	return n, err
}

func (r *Relay) listen(ctx context.Context, wake chan<- struct{}) {
	for ctx.Err() == nil {
		conn, err := r.pool.Acquire(ctx)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		if _, err := conn.Exec(ctx, "listen outbox"); err == nil {
			for {
				if _, err := conn.Conn().WaitForNotification(ctx); err != nil {
					break
				}
				select {
				case wake <- struct{}{}:
				default:
				}
			}
		}
		conn.Release()
	}
}
