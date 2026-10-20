package rbac

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/store"
)

const cacheTTL = 60 * time.Second

type permEntry struct {
	keys map[string]bool
	exp  time.Time
}

type sessionEntry struct {
	active bool
	exp    time.Time
}

// Checker answers permission and session questions with a short cache. Role changes, staff
// exits and session revocations call pg_notify, and every API instance drops its entries.
type Checker struct {
	q        *store.Queries
	mu       sync.Mutex
	perms    map[uuid.UUID]permEntry
	sessions map[uuid.UUID]sessionEntry
}

// NewChecker creates a checker.
func NewChecker(q *store.Queries) *Checker {
	return &Checker{q: q, perms: map[uuid.UUID]permEntry{}, sessions: map[uuid.UUID]sessionEntry{}}
}

// HasPermission reports whether userID currently holds perm. "admin.*" holders don't get
// everything implicitly: the admin role is seeded with every permission explicitly.
func (c *Checker) HasPermission(ctx context.Context, userID uuid.UUID, perm string) (bool, error) {
	c.mu.Lock()
	e, ok := c.perms[userID]
	c.mu.Unlock()
	if !ok || time.Now().After(e.exp) {
		keys, err := c.q.IdentityUserPermissionKeys(ctx, userID)
		if err != nil {
			return false, err
		}
		e = permEntry{keys: map[string]bool{}, exp: time.Now().Add(cacheTTL)}
		for _, k := range keys {
			e.keys[k] = true
		}
		c.mu.Lock()
		c.perms[userID] = e
		c.mu.Unlock()
	}
	return e.keys[perm], nil
}

// Permissions lists a user's permission keys (for /me and the staff portal menu).
func (c *Checker) Permissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return c.q.IdentityUserPermissionKeys(ctx, userID)
}

// IsSensitive reports whether perm needs a step-up.
func (c *Checker) IsSensitive(perm string) bool { return sensitive[perm] }

// SessionActive reports whether a session is still active.
func (c *Checker) SessionActive(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	c.mu.Lock()
	e, ok := c.sessions[sessionID]
	c.mu.Unlock()
	if ok && time.Now().Before(e.exp) {
		return e.active, nil
	}
	active, err := c.q.IdentitySessionActive(ctx, sessionID)
	if err != nil {
		return false, err
	}
	c.mu.Lock()
	c.sessions[sessionID] = sessionEntry{active: active, exp: time.Now().Add(cacheTTL)}
	c.mu.Unlock()
	return active, nil
}

// InvalidateUser drops cached permissions for a user (payload of pg_notify('authz')).
func (c *Checker) InvalidateUser(payload string) {
	var msg struct {
		UserID uuid.UUID `json:"userId"`
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if json.Unmarshal([]byte(payload), &msg) == nil && msg.UserID != uuid.Nil {
		delete(c.perms, msg.UserID)
		return
	}
	c.perms = map[uuid.UUID]permEntry{}
}

// InvalidateSessions drops the session cache (payload of pg_notify('sessions')).
func (c *Checker) InvalidateSessions(string) {
	c.mu.Lock()
	c.sessions = map[uuid.UUID]sessionEntry{}
	c.mu.Unlock()
}
