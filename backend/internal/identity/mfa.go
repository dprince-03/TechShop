package identity

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/realtime"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/pkg/crypto"
	"github.com/dprince-03/techshop/backend/pkg/validate"
)

// PasswordResult is a session, or the next step (MFA or new-device OTP).
type PasswordResult struct {
	Status      string     `json:"status" enum:"signed_in,mfa_required,otp_required"`
	Tokens      *TokenPair `json:"tokens,omitempty"`
	MFAToken    string     `json:"mfaToken,omitempty" doc:"Valid 2 minutes; only for the next step"`
	ChallengeID *uuid.UUID `json:"challengeId,omitempty" doc:"Set when a new-device code was sent by SMS"`
}

// mfaToken binds a pending sign-in to a user and audience for 2 minutes.
func (s *Service) mfaToken(user uuid.UUID, aud string) string {
	body := fmt.Sprintf("%s|%s|%d", user, aud, time.Now().Add(2*time.Minute).Unix())
	return body + "|" + fmt.Sprintf("%x", crypto.HMAC(s.d.Cfg.TokenHMACKey, "mfa:"+body))
}

func (s *Service) parseMFAToken(tok string) (uuid.UUID, string, bool) {
	parts := strings.Split(tok, "|")
	if len(parts) != 4 {
		return uuid.Nil, "", false
	}
	body := strings.Join(parts[:3], "|")
	if !crypto.Equal([]byte(fmt.Sprintf("%x", crypto.HMAC(s.d.Cfg.TokenHMACKey, "mfa:"+body))), []byte(parts[3])) {
		return uuid.Nil, "", false
	}
	var exp int64
	if _, err := fmt.Sscan(parts[2], &exp); err != nil || time.Now().Unix() > exp {
		return uuid.Nil, "", false
	}
	id, err := uuid.Parse(parts[0])
	return id, parts[1], err == nil
}

// PasswordLogin signs in with email or phone and password. Staff always continue to TOTP;
// sellers on a new device continue to an SMS code. Errors are identical for unknown users and
// wrong passwords; repeated failures add a growing delay (no hard lockout).
func (s *Service) PasswordLogin(ctx context.Context, identifier, password string, dev Device) (*PasswordResult, error) {
	ident := strings.TrimSpace(strings.ToLower(identifier))
	var u store.IdentityUser
	var err error
	if strings.Contains(ident, "@") {
		u, err = s.d.Q.IdentityGetUserByEmail(ctx, &ident)
	} else {
		phone := validate.NormalisePhone(ident)
		u, err = s.d.Q.IdentityGetUserByPhone(ctx, &phone)
	}
	fails, _ := s.d.Limiter.Count(ctx, loginFails, ident, time.Now().Add(-15*time.Minute))
	if fails >= 3 {
		delay := time.Duration(1<<min(int(fails)-3, 5)) * time.Second
		time.Sleep(min(delay, 30*time.Second))
	}
	wrong := httpx.Unauthenticated("Wrong email, phone or password.")
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && u.PasswordHash == nil) {
		auth.DummyVerify(password)
		_, _ = s.d.Limiter.Allow(ctx, loginFails, ident)
		s.event(ctx, s.d.Q, nil, "login_failed", map[string]any{"reason": "unknown_or_no_password"})
		return nil, wrong
	}
	if err != nil {
		return nil, err
	}
	if !auth.VerifyPassword(password, *u.PasswordHash) {
		_, _ = s.d.Limiter.Allow(ctx, loginFails, ident)
		s.event(ctx, s.d.Q, &u.ID, "login_failed", map[string]any{"reason": "wrong_password", "aud": dev.Audience})
		return nil, wrong
	}
	if u.Status != "active" {
		return nil, httpx.Forbidden("This account is suspended. Contact support.")
	}
	if dev.Audience == auth.AudStaff {
		sm, err := s.d.Q.IdentityGetStaffMember(ctx, u.ID)
		if err != nil || sm.Status != "active" {
			s.event(ctx, s.d.Q, &u.ID, "login_failed", map[string]any{"reason": "not_active_staff"})
			return nil, wrong
		}
		f, err := s.d.Q.IdentityGetTOTPFactor(ctx, u.ID)
		if err != nil || f.ConfirmedAt == nil {
			return nil, httpx.E(403, "mfa_enrollment_required", "Set up your authenticator from your invite link before signing in.")
		}
		return &PasswordResult{Status: "mfa_required", MFAToken: s.mfaToken(u.ID, dev.Audience)}, nil
	}
	if f, err := s.d.Q.IdentityGetTOTPFactor(ctx, u.ID); err == nil && f.ConfirmedAt != nil {
		return &PasswordResult{Status: "mfa_required", MFAToken: s.mfaToken(u.ID, dev.Audience)}, nil
	}
	if dev.Audience == auth.AudSeller && u.Phone != nil && !s.knownDevice(ctx, u.ID, dev.DeviceID) {
		challenge, err := s.RequestOTP(ctx, *u.Phone, "new_device")
		if err != nil {
			return nil, err
		}
		return &PasswordResult{Status: "otp_required", MFAToken: s.mfaToken(u.ID, dev.Audience), ChallengeID: &challenge}, nil
	}
	var tp *TokenPair
	err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		tp, err = s.createSession(ctx, tx, u.ID, dev, []string{"pwd"}, nil)
		return err
	})
	return &PasswordResult{Status: "signed_in", Tokens: tp}, err
}

func (s *Service) knownDevice(ctx context.Context, user uuid.UUID, deviceID string) bool {
	if deviceID == "" {
		return false
	}
	known, err := s.d.Q.IdentityHasDeviceSession(ctx, store.IdentityHasDeviceSessionParams{UserID: user, DeviceID: &deviceID})
	return err == nil && known
}

// VerifyTOTP completes a password sign-in with an authenticator code (no code reuse).
func (s *Service) VerifyTOTP(ctx context.Context, mfaToken, code string, dev Device) (*TokenPair, error) {
	userID, aud, ok := s.parseMFAToken(mfaToken)
	if !ok || aud != dev.Audience {
		return nil, httpx.Unauthenticated("This sign-in step has expired. Start again.")
	}
	var tp *TokenPair
	err := s.runKeep(ctx, func(tx *uow.Tx) error {
		if err := s.checkTOTP(ctx, tx, userID, code); err != nil {
			return err
		}
		now := time.Now()
		var err error
		tp, err = s.createSession(ctx, tx, userID, dev, []string{"pwd", "totp"}, &now)
		return err
	})
	return tp, err
}

// VerifyRecoveryCode completes MFA with a single-use recovery code (admins are alerted).
func (s *Service) VerifyRecoveryCode(ctx context.Context, mfaToken, code string, dev Device) (*TokenPair, error) {
	userID, aud, ok := s.parseMFAToken(mfaToken)
	if !ok || aud != dev.Audience {
		return nil, httpx.Unauthenticated("This sign-in step has expired. Start again.")
	}
	var tp *TokenPair
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		n, err := tx.Q.IdentityUseRecoveryCode(ctx, store.IdentityUseRecoveryCodeParams{UserID: userID, CodeHash: crypto.HMAC(s.d.Cfg.OTPHMACKey, "recovery:"+strings.ToLower(strings.TrimSpace(code)))})
		if err != nil {
			return err
		}
		if n == 0 {
			return httpx.Invalid("code_invalid", "That recovery code isn't valid or was already used.")
		}
		now := time.Now()
		s.event(ctx, tx.Q, &userID, "step_up_succeeded", map[string]any{"method": "recovery_code"})
		tp, err = s.createSession(ctx, tx, userID, dev, []string{"pwd", "recovery"}, &now)
		return err
	})
	return tp, err
}

// VerifyNewDeviceOTP completes a seller sign-in on a new device.
func (s *Service) VerifyNewDeviceOTP(ctx context.Context, mfaToken string, challenge uuid.UUID, code string, dev Device) (*TokenPair, error) {
	userID, aud, ok := s.parseMFAToken(mfaToken)
	if !ok || aud != dev.Audience {
		return nil, httpx.Unauthenticated("This sign-in step has expired. Start again.")
	}
	var tp *TokenPair
	err := s.runKeep(ctx, func(tx *uow.Tx) error {
		c, err := s.checkCode(ctx, tx, challenge, code)
		if err != nil {
			return err
		}
		if c.UserID == nil || *c.UserID != userID {
			return httpx.Invalid("code_invalid", "That code isn't for this sign-in.")
		}
		tp, err = s.createSession(ctx, tx, userID, dev, []string{"pwd", "otp"}, nil)
		return err
	})
	return tp, err
}

func (s *Service) checkTOTP(ctx context.Context, tx *uow.Tx, userID uuid.UUID, code string) error {
	f, err := tx.Q.IdentityGetTOTPFactor(ctx, userID)
	if err != nil || f.ConfirmedAt == nil {
		return httpx.Invalid("mfa_not_enrolled", "No authenticator is set up for this account.")
	}
	secret, err := s.d.Sealer.Open(f.SecretEncrypted, "identity.user_mfa_factors:"+userID.String())
	if err != nil {
		return err
	}
	step, ok := auth.CheckTOTP(string(secret), strings.TrimSpace(code), time.Now())
	if !ok {
		s.event(ctx, tx.Q, &userID, "mfa_failed", nil)
		return &errCommit{httpx.Invalid("code_invalid", "That authenticator code isn't right.")}
	}
	n, err := tx.Q.IdentityUseTOTPStep(ctx, store.IdentityUseTOTPStepParams{ID: f.ID, LastUsedStep: &step})
	if err != nil {
		return err
	}
	if n == 0 {
		return &errCommit{httpx.Invalid("code_reused", "That code was already used. Wait for the next one.")}
	}
	return nil
}

// StepUp re-confirms the authenticator for the current session and returns fresh tokens
// carrying mfa_at = now (sensitive actions accept it for STEP_UP_MAX_AGE).
func (s *Service) StepUp(ctx context.Context, p *httpx.Principal, code string) (*TokenPair, error) {
	var tp *TokenPair
	err := s.runKeep(ctx, func(tx *uow.Tx) error {
		if err := s.checkTOTP(ctx, tx, p.UserID, code); err != nil {
			return err
		}
		sess, err := tx.Q.IdentityGetSession(ctx, p.SessionID)
		if err != nil {
			return err
		}
		if err := tx.Q.IdentitySetSessionMFA(ctx, sess.FamilyID); err != nil {
			return err
		}
		s.event(ctx, tx.Q, &p.UserID, "step_up_succeeded", nil)
		now := time.Now()
		sess.MfaAt = &now
		access, exp, err := s.d.Keys.Issue(auth.IssueInput{UserID: p.UserID, SessionID: sess.ID, Audience: sess.Aud, AMR: append(sess.Amr, "totp"),
			AuthTime: p.AuthTime, MFAAt: &now, TTL: s.accessTTL(sess.Aud)})
		if err != nil {
			return err
		}
		tp = &TokenPair{AccessToken: access, AccessExpiresAt: exp, SessionID: sess.ID, UserID: p.UserID, RefreshExpiresAt: sess.ExpiresAt}
		return nil
	})
	return tp, err
}

// TOTPSetup is shown once: the secret and the otpauth:// URL for the QR code.
type TOTPSetup struct {
	Secret     string `json:"secret"`
	OTPAuthURL string `json:"otpauthUrl"`
}

// SetupTOTP stores a new, unconfirmed secret (encrypted, bound to the user row).
func (s *Service) SetupTOTP(ctx context.Context, userID uuid.UUID, account string) (*TOTPSetup, error) {
	secret, url, err := auth.NewTOTPSecret("TechShop", account)
	if err != nil {
		return nil, err
	}
	sealed, err := s.d.Sealer.Seal([]byte(secret), "identity.user_mfa_factors:"+userID.String())
	if err != nil {
		return nil, err
	}
	if _, err := s.d.Q.IdentityUpsertPendingTOTP(ctx, store.IdentityUpsertPendingTOTPParams{UserID: userID, SecretEncrypted: sealed}); err != nil {
		return nil, err
	}
	return &TOTPSetup{Secret: secret, OTPAuthURL: url}, nil
}

// ConfirmTOTP turns the factor on and returns 10 recovery codes (shown once, stored hashed).
func (s *Service) ConfirmTOTP(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	var codes []string
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		f, err := tx.Q.IdentityGetTOTPFactor(ctx, userID)
		if err != nil {
			return httpx.Invalid("mfa_not_started", "Start authenticator setup first.")
		}
		secret, err := s.d.Sealer.Open(f.SecretEncrypted, "identity.user_mfa_factors:"+userID.String())
		if err != nil {
			return err
		}
		step, ok := auth.CheckTOTP(string(secret), strings.TrimSpace(code), time.Now())
		if !ok {
			return httpx.Invalid("code_invalid", "That code doesn't match. Use the current code from your authenticator.")
		}
		if err := tx.Q.IdentityConfirmTOTP(ctx, store.IdentityConfirmTOTPParams{ID: f.ID, LastUsedStep: &step}); err != nil {
			return err
		}
		codes, err = s.newRecoveryCodes(ctx, tx, userID)
		if err != nil {
			return err
		}
		s.event(ctx, tx.Q, &userID, "mfa_enrolled", map[string]any{"kind": "totp"})
		return nil
	})
	return codes, err
}

func (s *Service) newRecoveryCodes(ctx context.Context, tx *uow.Tx, userID uuid.UUID) ([]string, error) {
	if err := tx.Q.IdentityDeleteRecoveryCodes(ctx, userID); err != nil {
		return nil, err
	}
	codes := make([]string, 10)
	for i := range codes {
		raw := strings.ToLower(crypto.RandomToken(6))
		codes[i] = raw[:4] + "-" + raw[4:8]
		if err := tx.Q.IdentityInsertRecoveryCode(ctx, store.IdentityInsertRecoveryCodeParams{UserID: userID, CodeHash: crypto.HMAC(s.d.Cfg.OTPHMACKey, "recovery:"+codes[i])}); err != nil {
			return nil, err
		}
	}
	return codes, nil
}

// RemoveTOTP turns MFA off (needs a recent step-up; staff can't remove it).
func (s *Service) RemoveTOTP(ctx context.Context, userID uuid.UUID) error {
	if _, err := s.d.Q.IdentityGetStaffMember(ctx, userID); err == nil {
		return httpx.Forbidden("Staff accounts must keep an authenticator. Ask an admin to reset it instead.")
	}
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if err := tx.Q.IdentityDeleteTOTP(ctx, userID); err != nil {
			return err
		}
		if err := tx.Q.IdentityDeleteRecoveryCodes(ctx, userID); err != nil {
			return err
		}
		s.event(ctx, tx.Q, &userID, "mfa_removed", nil)
		tx.Notify(realtime.ChannelSessions, map[string]any{"userId": userID})
		return nil
	})
}
