// Package identity implements sign-in (phone OTP, password, staff TOTP, step-up), sessions with
// rotating refresh tokens and reuse detection, profiles, addresses, consents, privacy requests,
// staff administration (roles, invites, exits) and service tokens.
// Plan: docs/identity-access.md.
package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/notify"
	"github.com/dprince-03/techshop/backend/internal/ratelimit"
	"github.com/dprince-03/techshop/backend/internal/realtime"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/pkg/crypto"
	"github.com/dprince-03/techshop/backend/pkg/validate"
)

// OTP limits (identity-access.md §6).
var (
	otpPerMinute = ratelimit.Rule{Name: "otp:phone:1m", Max: 1, Window: time.Minute}
	otpPerHour   = ratelimit.Rule{Name: "otp:phone:1h", Max: 5, Window: time.Hour}
	otpPerDay    = ratelimit.Rule{Name: "otp:phone:1d", Max: 10, Window: 24 * time.Hour}
	otpPerIP     = ratelimit.Rule{Name: "otp:ip:1h", Max: 20, Window: time.Hour}
	loginFails   = ratelimit.Rule{Name: "login:fail:15m", Max: 1000, Window: 15 * time.Minute}
)

const maxOTPAttempts = 5

// Session lifetimes per audience: idle and absolute (identity-access.md §2).
var lifetimes = map[string][2]time.Duration{
	auth.AudMarket:      {30 * 24 * time.Hour, 90 * 24 * time.Hour},
	auth.AudWholesale:   {30 * 24 * time.Hour, 90 * 24 * time.Hour},
	auth.AudCustomerApp: {30 * 24 * time.Hour, 90 * 24 * time.Hour},
	auth.AudSeller:      {14 * 24 * time.Hour, 30 * 24 * time.Hour},
	auth.AudStaff:       {12 * time.Hour, 12 * time.Hour},
	auth.AudLogistics:   {7 * 24 * time.Hour, 30 * 24 * time.Hour},
}

// Service holds identity logic.
type Service struct {
	d      *kit.Deps
	notify *notify.Service
}

// New creates the identity module.
func New(d *kit.Deps, n *notify.Service) *Service { return &Service{d: d, notify: n} }

// Name implements kit.Module.
func (s *Service) Name() string { return "identity" }

// TokenPair is returned by every successful sign-in.
type TokenPair struct {
	AccessToken      string    `json:"accessToken"`
	AccessExpiresAt  time.Time `json:"accessExpiresAt"`
	RefreshToken     string    `json:"refreshToken"`
	RefreshExpiresAt time.Time `json:"refreshExpiresAt"`
	SessionID        uuid.UUID `json:"sessionId"`
	UserID           uuid.UUID `json:"userId"`
}

// Device describes the client creating a session.
type Device struct {
	Audience   string `json:"audience" enum:"market,wholesale,seller,staff,customer_app,logistics"`
	Platform   string `json:"platform,omitempty" enum:"web,ios,android"`
	DeviceID   string `json:"deviceId,omitempty" maxLength:"128"`
	DeviceName string `json:"deviceName,omitempty" maxLength:"120"`
	AppVersion string `json:"appVersion,omitempty" maxLength:"40"`
}

func (s *Service) accessTTL(aud string) time.Duration {
	if aud == auth.AudStaff {
		return s.d.Cfg.StaffAccessTokenTTL
	}
	return s.d.Cfg.AccessTokenTTL
}

// otpHash binds a code to its challenge so codes can't be swapped between challenges.
func (s *Service) otpHash(challengeID uuid.UUID, code string) []byte {
	return crypto.HMAC(s.d.Cfg.OTPHMACKey, challengeID.String()+":"+code)
}

func ipOf(ctx context.Context) *netip.Addr {
	if a, err := netip.ParseAddr(httpx.MetaFrom(ctx).IP); err == nil {
		return &a
	}
	return nil
}

func uaOf(ctx context.Context) *string {
	if ua := httpx.MetaFrom(ctx).UserAgent; ua != "" {
		return &ua
	}
	return nil
}

// event records a security event (append-only; identity-access.md §10).
func (s *Service) event(ctx context.Context, q *store.Queries, user *uuid.UUID, kind string, meta map[string]any) {
	b, _ := json.Marshal(meta)
	if b == nil || string(b) == "null" {
		b = []byte("{}")
	}
	_ = q.IdentityInsertAuthEvent(ctx, store.IdentityInsertAuthEventParams{UserID: user, Kind: kind, Ip: ipOf(ctx), UserAgent: uaOf(ctx), Meta: b})
}

// RequestOTP sends a sign-in code by SMS. The response is identical whether or not the number
// has an account (no enumeration).
func (s *Service) RequestOTP(ctx context.Context, phoneRaw, purpose string) (uuid.UUID, error) {
	phone := validate.NormalisePhone(phoneRaw)
	if phone == "" {
		return uuid.Nil, httpx.Invalid("invalid_phone", "Enter a Nigerian mobile number, for example 0803 123 4567.").Field("body.phone", "not a Nigerian mobile number", phoneRaw)
	}
	for _, rule := range []ratelimit.Rule{otpPerMinute, otpPerHour, otpPerDay} {
		ok, err := s.d.Limiter.Allow(ctx, rule, phone)
		if err != nil {
			return uuid.Nil, err
		}
		if !ok {
			s.event(ctx, s.d.Q, nil, "rate_limited", map[string]any{"rule": rule.Name, "phone": validate.MaskPhone(phone)})
			return uuid.Nil, httpx.E(429, "otp_rate_limited", "Too many codes requested for this number. Please wait and try again.")
		}
	}
	if ip := httpx.MetaFrom(ctx).IP; ip != "" {
		if ok, err := s.d.Limiter.Allow(ctx, otpPerIP, ip); err == nil && !ok {
			return uuid.Nil, httpx.E(429, "otp_rate_limited", "Too many codes requested from this network. Please wait and try again.")
		}
	}
	code := crypto.RandomDigits(6)
	challenge := uuid.New()
	var userID *uuid.UUID
	if u, err := s.d.Q.IdentityGetUserByPhone(ctx, &phone); err == nil {
		userID = &u.ID
	}
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		row, err := tx.Q.IdentityCreateCode(ctx, store.IdentityCreateCodeParams{
			ID: challenge, UserID: userID, Channel: "sms", Destination: phone, Purpose: purpose,
			CodeHash: s.otpHash(challenge, code), Ip: ipOf(ctx), UserAgent: uaOf(ctx), ExpiresAt: time.Now().Add(s.d.Cfg.OTPTTL),
		})
		if err != nil {
			return err
		}
		s.event(ctx, tx.Q, userID, "otp_requested", map[string]any{"phone": validate.MaskPhone(phone), "purpose": purpose})
		return s.notify.Send(ctx, tx, notify.Msg{UserID: userID, To: phone, Channel: "sms", Category: notify.CatSecurity, Template: "otp",
			Data: map[string]any{"code": code}, Key: "otp:" + row.ID.String()})
	})
	return challenge, err
}

// checkCode verifies a challenge code (constant time, 5 attempts, single use).
func (s *Service) checkCode(ctx context.Context, tx *uow.Tx, challengeID uuid.UUID, code string) (store.IdentityVerificationCode, error) {
	c, err := tx.Q.IdentityGetCode(ctx, challengeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, httpx.Invalid("code_invalid", "That code isn't right or has expired. Request a new one.")
	}
	if err != nil {
		return c, err
	}
	if c.ConsumedAt != nil || time.Now().After(c.ExpiresAt) || c.Attempts >= maxOTPAttempts {
		return c, httpx.Invalid("code_expired", "This code no longer works. Request a new one.")
	}
	if !crypto.Equal(c.CodeHash, s.otpHash(c.ID, strings.TrimSpace(code))) {
		attempts, _ := tx.Q.IdentityBumpCodeAttempts(ctx, c.ID)
		s.event(ctx, tx.Q, c.UserID, "otp_failed", map[string]any{"attempt": attempts})
		left := maxOTPAttempts - int(attempts)
		if left <= 0 {
			return c, &errCommit{httpx.Invalid("code_expired", "Too many wrong codes. Request a new one.")}
		}
		return c, &errCommit{httpx.Invalid("code_invalid", fmt.Sprintf("That code isn't right. %d attempts left.", left))}
	}
	return c, tx.Q.IdentityConsumeCode(ctx, c.ID)
}

// errCommit is a client error whose transaction must still commit (to record the failed attempt).
type errCommit struct{ p *httpx.Problem }

func (e *errCommit) Error() string { return e.p.Error() }

// runKeep runs fn in a transaction; an errCommit error commits the work and returns the problem.
func (s *Service) runKeep(ctx context.Context, fn func(tx *uow.Tx) error) error {
	var keep *errCommit
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		err := fn(tx)
		if errors.As(err, &keep) {
			return nil
		}
		return err
	})
	if keep != nil {
		return keep.p
	}
	return err
}

// VerifyOTPResult is either a session or a request to complete the profile (new number).
type VerifyOTPResult struct {
	Status      string     `json:"status" enum:"signed_in,needs_profile"`
	Tokens      *TokenPair `json:"tokens,omitempty"`
	SignupToken string     `json:"signupToken,omitempty" doc:"Send with your name to /auth/otp/signup (valid 15 minutes)"`
}

// VerifyOTP checks the code and signs in (or asks for a name for a new number).
func (s *Service) VerifyOTP(ctx context.Context, challengeID uuid.UUID, code string, dev Device) (*VerifyOTPResult, error) {
	var res *VerifyOTPResult
	err := s.runKeep(ctx, func(tx *uow.Tx) error {
		c, err := s.checkCode(ctx, tx, challengeID, code)
		if err != nil {
			return err
		}
		u, err := tx.Q.IdentityGetUserByPhone(ctx, &c.Destination)
		if errors.Is(err, pgx.ErrNoRows) {
			res = &VerifyOTPResult{Status: "needs_profile", SignupToken: s.signupToken(c.Destination)}
			return nil
		}
		if err != nil {
			return err
		}
		if u.Status != "active" {
			return httpx.Forbidden("This account is suspended. Contact support.")
		}
		if dev.Audience == auth.AudStaff {
			return httpx.Forbidden("Staff sign in with work email, password and authenticator.")
		}
		_ = tx.Q.IdentityMarkPhoneVerified(ctx, u.ID)
		tp, err := s.createSession(ctx, tx, u.ID, dev, []string{"otp"}, nil)
		if err != nil {
			return err
		}
		res = &VerifyOTPResult{Status: "signed_in", Tokens: tp}
		return nil
	})
	return res, err
}

func (s *Service) signupToken(phone string) string {
	exp := time.Now().Add(15 * time.Minute).Unix()
	body := fmt.Sprintf("%s|%d", phone, exp)
	return body + "|" + fmt.Sprintf("%x", crypto.HMAC(s.d.Cfg.TokenHMACKey, "signup:"+body))
}

func (s *Service) parseSignupToken(tok string) (string, bool) {
	parts := strings.Split(tok, "|")
	if len(parts) != 3 {
		return "", false
	}
	body := parts[0] + "|" + parts[1]
	if fmt.Sprintf("%x", crypto.HMAC(s.d.Cfg.TokenHMACKey, "signup:"+body)) != parts[2] {
		return "", false
	}
	var exp int64
	if _, err := fmt.Sscan(parts[1], &exp); err != nil || time.Now().Unix() > exp {
		return "", false
	}
	return parts[0], true
}

// CompleteSignup creates the account for a verified phone and signs in.
func (s *Service) CompleteSignup(ctx context.Context, token, first, last string, dev Device) (*TokenPair, error) {
	phone, ok := s.parseSignupToken(token)
	if !ok {
		return nil, httpx.Invalid("signup_token_invalid", "This sign-up link has expired. Verify your number again.")
	}
	var tp *TokenPair
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		now := time.Now()
		u, err := tx.Q.IdentityCreateUser(ctx, store.IdentityCreateUserParams{Phone: &phone, FirstName: strings.TrimSpace(first), LastName: strings.TrimSpace(last), PhoneVerifiedAt: &now})
		if err != nil {
			return httpx.DB(err, "account")
		}
		tx.Emit("user", u.ID, "user.signed_up", map[string]any{"userId": u.ID})
		tp, err = s.createSession(ctx, tx, u.ID, dev, []string{"otp"}, nil)
		return err
	})
	return tp, err
}

// createSession issues a refresh token (stored as SHA-256) and an access token.
func (s *Service) createSession(ctx context.Context, tx *uow.Tx, userID uuid.UUID, dev Device, amr []string, mfaAt *time.Time) (*TokenPair, error) {
	if !auth.ValidAudience(dev.Audience) {
		return nil, httpx.Invalid("bad_audience", "Unknown client audience.")
	}
	life := lifetimes[dev.Audience]
	if dev.Platform == "" {
		dev.Platform = "web"
	}
	refresh := crypto.RandomToken(32)
	now := time.Now()
	sess, err := tx.Q.IdentityCreateSession(ctx, store.IdentityCreateSessionParams{
		UserID: userID, FamilyID: uuid.New(), RefreshTokenHash: crypto.SHA256(refresh), Aud: dev.Audience, Platform: dev.Platform,
		Amr: amr, DeviceID: strOrNil(dev.DeviceID), DeviceName: strOrNil(dev.DeviceName), AppVersion: strOrNil(dev.AppVersion),
		UserAgent: uaOf(ctx), Ip: ipOf(ctx), MfaAt: mfaAt, IdleExpiresAt: now.Add(life[0]), ExpiresAt: now.Add(life[1]),
	})
	if err != nil {
		return nil, err
	}
	_ = tx.Q.IdentityTouchLogin(ctx, userID)
	s.event(ctx, tx.Q, &userID, "login_succeeded", map[string]any{"aud": dev.Audience, "amr": amr, "device": dev.DeviceName})
	return s.issue(sess, refresh, now)
}

func (s *Service) issue(sess store.IdentityUserSession, refresh string, authTime time.Time) (*TokenPair, error) {
	access, exp, err := s.d.Keys.Issue(auth.IssueInput{UserID: sess.UserID, SessionID: sess.ID, Audience: sess.Aud, AMR: sess.Amr,
		AuthTime: authTime, MFAAt: sess.MfaAt, TTL: s.accessTTL(sess.Aud)})
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, AccessExpiresAt: exp, RefreshToken: refresh, RefreshExpiresAt: sess.ExpiresAt, SessionID: sess.ID, UserID: sess.UserID}, nil
}

// Refresh rotates a refresh token. Reusing a replaced token (outside a 10 s race window)
// revokes the whole family and alerts the user (stolen-token detection).
func (s *Service) Refresh(ctx context.Context, refresh string) (*TokenPair, error) {
	var tp *TokenPair
	var reuse bool
	err := s.runKeep(ctx, func(tx *uow.Tx) error {
		old, err := tx.Q.IdentityGetSessionByToken(ctx, crypto.SHA256(refresh))
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.Unauthenticated("Sign in again.")
		}
		if err != nil {
			return err
		}
		now := time.Now()
		if old.RevokedAt != nil || now.After(old.ExpiresAt) || now.After(old.IdleExpiresAt) {
			return httpx.Unauthenticated("Your session has ended. Sign in again.")
		}
		if old.ReplacedBy != nil && now.Sub(old.LastUsedAt) > 10*time.Second {
			if _, err := tx.Q.IdentityRevokeFamily(ctx, store.IdentityRevokeFamilyParams{FamilyID: old.FamilyID, RevokedReason: strOrNil("reuse_detected")}); err != nil {
				return err
			}
			s.event(ctx, tx.Q, &old.UserID, "refresh_reuse_detected", map[string]any{"family": old.FamilyID})
			tx.Notify(realtime.ChannelSessions, map[string]any{"userId": old.UserID})
			uid := old.UserID
			_ = s.notify.Send(ctx, tx, notify.Msg{UserID: &uid, Channel: "push", Category: notify.CatSecurity, Template: "session_reuse", Key: "reuse:" + old.FamilyID.String()})
			reuse = true
			return &errCommit{httpx.Unauthenticated("This session was signed out for your safety. Sign in again.")}
		}
		life := lifetimes[old.Aud]
		newRefresh := crypto.RandomToken(32)
		sess, err := tx.Q.IdentityCreateSession(ctx, store.IdentityCreateSessionParams{
			UserID: old.UserID, FamilyID: old.FamilyID, RefreshTokenHash: crypto.SHA256(newRefresh), Aud: old.Aud, Platform: old.Platform,
			Amr: old.Amr, DeviceID: old.DeviceID, DeviceName: old.DeviceName, AppVersion: old.AppVersion, UserAgent: uaOf(ctx), Ip: ipOf(ctx),
			MfaAt: old.MfaAt, IdleExpiresAt: minTime(now.Add(life[0]), old.ExpiresAt), ExpiresAt: old.ExpiresAt,
		})
		if err != nil {
			return err
		}
		if old.ReplacedBy == nil {
			if err := tx.Q.IdentityMarkSessionReplaced(ctx, store.IdentityMarkSessionReplacedParams{ID: old.ID, ReplacedBy: &sess.ID}); err != nil {
				return err
			}
		}
		tp, err = s.issue(sess, newRefresh, old.CreatedAt)
		return err
	})
	_ = reuse
	return tp, err
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// Logout revokes the session family of a refresh token or of the current session.
func (s *Service) Logout(ctx context.Context, p *httpx.Principal, refresh string) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if refresh != "" {
			if sess, err := tx.Q.IdentityGetSessionByToken(ctx, crypto.SHA256(refresh)); err == nil {
				_, err = tx.Q.IdentityRevokeFamily(ctx, store.IdentityRevokeFamilyParams{FamilyID: sess.FamilyID, RevokedReason: strOrNil("logout")})
				tx.Notify(realtime.ChannelSessions, map[string]any{"userId": sess.UserID})
				return err
			}
		}
		if p != nil {
			_, err := tx.Q.IdentityRevokeSession(ctx, store.IdentityRevokeSessionParams{ID: p.SessionID, UserID: p.UserID, RevokedReason: strOrNil("logout")})
			tx.Notify(realtime.ChannelSessions, map[string]any{"userId": p.UserID})
			return err
		}
		return nil
	})
}

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
