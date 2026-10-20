package httpx

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ctxKey int

const (
	principalKey ctxKey = iota
	requestMetaKey
)

// Principal is the authenticated caller.
type Principal struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	Audience  string
	AMR       []string
	AuthTime  time.Time
	MFAAt     *time.Time
	Scopes    []string
}

// WithPrincipal stores the principal in a context.
func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

// PrincipalFrom returns the authenticated principal, or nil for anonymous requests.
func PrincipalFrom(ctx context.Context) *Principal {
	p, _ := ctx.Value(principalKey).(*Principal)
	return p
}

// MustPrincipal returns the principal; guarded routes always have one.
func MustPrincipal(ctx context.Context) *Principal {
	p := PrincipalFrom(ctx)
	if p == nil {
		panic("httpx: MustPrincipal on an unauthenticated route")
	}
	return p
}

// RequestMeta carries request details for audit entries.
type RequestMeta struct {
	RequestID string
	IP        string
	UserAgent string
}

// WithRequestMeta stores request metadata.
func WithRequestMeta(ctx context.Context, m RequestMeta) context.Context {
	return context.WithValue(ctx, requestMetaKey, m)
}

// MetaFrom returns request metadata (empty if absent, e.g. in jobs).
func MetaFrom(ctx context.Context) RequestMeta {
	m, _ := ctx.Value(requestMetaKey).(RequestMeta)
	return m
}
