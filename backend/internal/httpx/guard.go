package httpx

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
)

// SessionChecker reports whether a session is still active (cached; invalidated by NOTIFY).
type SessionChecker interface {
	SessionActive(ctx context.Context, sessionID uuid.UUID) (bool, error)
}

// PermissionChecker reports whether a user holds a permission and whether it is sensitive.
type PermissionChecker interface {
	HasPermission(ctx context.Context, userID uuid.UUID, perm string) (bool, error)
	IsSensitive(perm string) bool
}

// Guard builds per-route middleware: authentication, audience, permission and step-up.
type Guard struct {
	API          huma.API
	Keys         *auth.Keyring
	Sessions     SessionChecker
	Perms        PermissionChecker
	StepUpMaxAge time.Duration
}

// Authenticate requires a valid bearer token for one of the audiences.
func (g *Guard) Authenticate(audiences ...string) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		raw := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
		if raw == "" || raw == ctx.Header("Authorization") {
			_ = huma.WriteErr(g.API, ctx, http.StatusUnauthorized, "Sign in to continue.")
			return
		}
		claims, err := g.Keys.Verify(raw, audiences...)
		if err != nil {
			_ = huma.WriteErr(g.API, ctx, http.StatusUnauthorized, "Your session is invalid or has expired.")
			return
		}
		uid, err := claims.UserID()
		if err != nil {
			_ = huma.WriteErr(g.API, ctx, http.StatusUnauthorized, "Invalid token subject.")
			return
		}
		p := &Principal{UserID: uid, SessionID: claims.SessionID, Audience: claims.Aud(), AMR: claims.AMR,
			AuthTime: time.Unix(claims.AuthTime, 0), Scopes: claims.Scopes}
		if claims.MFAAt > 0 {
			t := time.Unix(claims.MFAAt, 0)
			p.MFAAt = &t
		}
		if claims.Aud() != auth.AudInternal && g.Sessions != nil {
			ok, err := g.Sessions.SessionActive(ctx.Context(), claims.SessionID)
			if err != nil || !ok {
				_ = huma.WriteErr(g.API, ctx, http.StatusUnauthorized, "This session has been signed out.")
				return
			}
		}
		ctx.SetHeader("Cache-Control", "no-store")
		next(huma.WithContext(ctx, WithPrincipal(ctx.Context(), p)))
	}
}

// Optional attaches the principal when a valid token is present, but allows anonymous calls.
func (g *Guard) Optional(audiences ...string) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		raw := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
		if raw != "" && raw != ctx.Header("Authorization") {
			if claims, err := g.Keys.Verify(raw, audiences...); err == nil {
				if uid, err := claims.UserID(); err == nil {
					p := &Principal{UserID: uid, SessionID: claims.SessionID, Audience: claims.Aud(), AMR: claims.AMR, AuthTime: time.Unix(claims.AuthTime, 0)}
					next(huma.WithContext(ctx, WithPrincipal(ctx.Context(), p)))
					return
				}
			}
		}
		next(ctx)
	}
}

// Require checks a staff permission; sensitive permissions also need a recent MFA (step-up).
func (g *Guard) Require(perm string) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		p := PrincipalFrom(ctx.Context())
		if p == nil {
			_ = huma.WriteErr(g.API, ctx, http.StatusUnauthorized, "Sign in to continue.")
			return
		}
		ok, err := g.Perms.HasPermission(ctx.Context(), p.UserID, perm)
		if err != nil {
			_ = huma.WriteErr(g.API, ctx, http.StatusInternalServerError, "Could not check permissions.")
			return
		}
		if !ok {
			_ = huma.WriteErr(g.API, ctx, http.StatusForbidden, "You don't have permission "+perm+".")
			return
		}
		if g.Perms.IsSensitive(perm) && !g.recentMFA(p) {
			writeProblem(g.API, ctx, StepUpRequired())
			return
		}
		next(ctx)
	}
}

// RequireScope checks a service-token scope (internal API).
func (g *Guard) RequireScope(scope string) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		p := PrincipalFrom(ctx.Context())
		if p != nil {
			for _, s := range p.Scopes {
				if s == scope {
					next(ctx)
					return
				}
			}
		}
		_ = huma.WriteErr(g.API, ctx, http.StatusForbidden, "Missing scope "+scope+".")
	}
}

// RecentMFA reports whether the principal passed MFA within the step-up window.
func (g *Guard) recentMFA(p *Principal) bool {
	return p.MFAAt != nil && time.Since(*p.MFAAt) <= g.StepUpMaxAge
}

// RequireStepUp is for sensitive actions that aren't tied to one permission (e.g. MFA removal).
func (g *Guard) RequireStepUp() func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		p := PrincipalFrom(ctx.Context())
		if p == nil || !g.recentMFA(p) {
			writeProblem(g.API, ctx, StepUpRequired())
			return
		}
		next(ctx)
	}
}

func writeProblem(api huma.API, ctx huma.Context, p *Problem) {
	ctx.SetHeader("Content-Type", "application/problem+json")
	ctx.SetStatus(p.Status)
	b, _ := jsonMarshal(p)
	_, _ = ctx.BodyWriter().Write(b)
}

// GinPrincipal authenticates a raw Gin route (SSE streams). Browsers can't set headers on
// EventSource, so the BFF proxies with the header; apps use an SSE client that sets it.
func (g *Guard) GinPrincipal(authorization string, audiences ...string) (*Principal, bool) {
	raw := strings.TrimPrefix(authorization, "Bearer ")
	if raw == "" || raw == authorization {
		return nil, false
	}
	claims, err := g.Keys.Verify(raw, audiences...)
	if err != nil {
		return nil, false
	}
	uid, err := claims.UserID()
	if err != nil {
		return nil, false
	}
	return &Principal{UserID: uid, SessionID: claims.SessionID, Audience: claims.Aud(), AMR: claims.AMR, AuthTime: time.Unix(claims.AuthTime, 0)}, true
}
