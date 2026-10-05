// Package idempotency makes retried requests run once (docs/backend.md §2.3): a replay with the
// same key and body returns the stored response; a concurrent duplicate gets 409; the same key
// with a different body gets 422.
package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Store wraps the idempotency_keys table.
type Store struct {
	q   *store.Queries
	ttl time.Duration
}

// New creates a store whose keys expire after 24 hours.
func New(q *store.Queries) *Store { return &Store{q: q, ttl: 24 * time.Hour} }

// Do runs fn once per (scope, key). The response is stored as JSON and replayed on retries.
func Do[T any](ctx context.Context, s *Store, scope, key string, userID *uuid.UUID, method, path string, request any, fn func() (*T, error)) (*T, bool, error) {
	if key == "" || len(key) > 128 {
		return nil, false, httpx.Invalid("idempotency_key_required", "Send an Idempotency-Key header (1–128 characters) with this request.")
	}
	reqBytes, _ := json.Marshal(request)
	hash := sha256.Sum256(reqBytes)

	_, err := s.q.PlatformInsertIdempotencyKey(ctx, store.PlatformInsertIdempotencyKeyParams{
		Scope: scope, Key: key, UserID: userID, Method: method, Path: path, RequestHash: hash[:], ExpiresAt: time.Now().Add(s.ttl),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		existing, gerr := s.q.PlatformGetIdempotencyKey(ctx, store.PlatformGetIdempotencyKeyParams{Scope: scope, Key: key})
		if gerr != nil {
			return nil, false, gerr
		}
		if string(existing.RequestHash) != string(hash[:]) {
			return nil, false, httpx.Invalid("idempotency_key_reused", "This Idempotency-Key was already used with a different request.")
		}
		if existing.CompletedAt == nil {
			return nil, false, httpx.Conflict("request_in_progress", "The same request is still being processed.")
		}
		var out T
		if err := json.Unmarshal(existing.ResponseBody, &out); err != nil {
			return nil, false, err
		}
		return &out, true, nil
	}
	if err != nil {
		return nil, false, err
	}
	out, ferr := fn()
	if ferr != nil {
		_ = s.q.PlatformReleaseIdempotencyKey(context.WithoutCancel(ctx), store.PlatformReleaseIdempotencyKeyParams{Scope: scope, Key: key})
		return nil, false, ferr
	}
	body, _ := json.Marshal(out)
	code := int16(200)
	if err := s.q.PlatformCompleteIdempotencyKey(ctx, store.PlatformCompleteIdempotencyKeyParams{Scope: scope, Key: key, ResponseCode: &code, ResponseBody: body}); err != nil {
		return nil, false, err
	}
	return out, false, nil
}
