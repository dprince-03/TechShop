package identity

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/notify"
	"github.com/dprince-03/techshop/backend/internal/realtime"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/pkg/validate"
)

// Me is the signed-in account with what it may act as.
type Me struct {
	ID             uuid.UUID       `json:"id"`
	FirstName      string          `json:"firstName"`
	LastName       string          `json:"lastName"`
	Email          *string         `json:"email,omitempty"`
	Phone          *string         `json:"phone,omitempty"`
	PhoneVerified  bool            `json:"phoneVerified"`
	MarketingOptIn bool            `json:"marketingOptIn"`
	MFAEnabled     bool            `json:"mfaEnabled"`
	IsStaff        bool            `json:"isStaff"`
	IsRider        bool            `json:"isRider"`
	Permissions    []string        `json:"permissions" doc:"Staff permissions (empty for non-staff)"`
	Sellers        json.RawMessage `json:"sellers" doc:"Seller memberships"`
	Businesses     json.RawMessage `json:"businesses" doc:"Business memberships"`
	CreatedAt      time.Time       `json:"createdAt"`
}

// GetMe returns the profile and memberships.
func (s *Service) GetMe(ctx context.Context, id uuid.UUID) (*Me, error) {
	u, err := s.d.Q.IdentityGetUser(ctx, id)
	if err != nil {
		return nil, httpx.DB(err, "account")
	}
	m, err := s.d.Q.IdentityMemberships(ctx, id)
	if err != nil {
		return nil, err
	}
	perms := []string{}
	if m.IsStaff {
		if p, err := s.d.RBAC.Permissions(ctx, id); err == nil {
			perms = p
		}
	}
	return &Me{ID: u.ID, FirstName: u.FirstName, LastName: u.LastName, Email: u.Email, Phone: u.Phone, PhoneVerified: u.PhoneVerifiedAt != nil,
		MarketingOptIn: u.MarketingOptIn, MFAEnabled: m.MfaEnabled, IsStaff: m.IsStaff, IsRider: m.IsRider, Permissions: perms,
		Sellers: m.Sellers, Businesses: m.Businesses, CreatedAt: u.CreatedAt}, nil
}

// ChangePassword checks the current password (when one exists), sets the new one and signs
// out every other device.
func (s *Service) ChangePassword(ctx context.Context, p *httpx.Principal, current, next string) error {
	if err := auth.CheckPasswordPolicy(next); err != nil {
		return httpx.Invalid("weak_password", err.Error())
	}
	u, err := s.d.Q.IdentityGetUser(ctx, p.UserID)
	if err != nil {
		return err
	}
	if u.PasswordHash != nil && !auth.VerifyPassword(current, *u.PasswordHash) {
		return httpx.Invalid("wrong_password", "Your current password is wrong.")
	}
	hash, err := auth.HashPassword(next)
	if err != nil {
		return err
	}
	sess, err := s.d.Q.IdentityGetSession(ctx, p.SessionID)
	if err != nil {
		return err
	}
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if err := tx.Q.IdentitySetPassword(ctx, store.IdentitySetPasswordParams{ID: p.UserID, PasswordHash: &hash}); err != nil {
			return err
		}
		if _, err := tx.Q.IdentityRevokeOtherSessions(ctx, store.IdentityRevokeOtherSessionsParams{UserID: p.UserID, FamilyID: sess.FamilyID}); err != nil {
			return err
		}
		s.event(ctx, tx.Q, &p.UserID, "password_changed", nil)
		tx.Notify(realtime.ChannelSessions, map[string]any{"userId": p.UserID})
		// Account-takeover protection: hold seller payouts for 24 hours (identity-access.md §6).
		tx.Emit("user", p.UserID, "user.credentials_changed", map[string]any{"userId": p.UserID, "kind": "password"})
		if u.Email != nil {
			return s.notify.Send(ctx, tx, notify.Msg{UserID: &p.UserID, To: *u.Email, Channel: "email", Category: notify.CatSecurity, Template: "password_changed", Key: "pwchg:" + uuid.NewString()})
		}
		return nil
	})
}

// SetInitialPassword lets an OTP-only customer add a password (optional for customers).
func (s *Service) SetInitialPassword(ctx context.Context, userID uuid.UUID, pw string) error {
	if err := auth.CheckPasswordPolicy(pw); err != nil {
		return httpx.Invalid("weak_password", err.Error())
	}
	h, err := auth.HashPassword(pw)
	if err != nil {
		return err
	}
	return s.d.Q.IdentitySetPassword(ctx, store.IdentitySetPasswordParams{ID: userID, PasswordHash: &h})
}

// SessionDTO is one signed-in device.
type SessionDTO struct {
	ID         uuid.UUID `json:"id"`
	Audience   string    `json:"audience"`
	Platform   string    `json:"platform"`
	DeviceName *string   `json:"deviceName,omitempty"`
	UserAgent  *string   `json:"userAgent,omitempty"`
	LastUsedAt time.Time `json:"lastUsedAt"`
	CreatedAt  time.Time `json:"createdAt"`
	Current    bool      `json:"current"`
}

// ListSessions lists signed-in devices.
func (s *Service) ListSessions(ctx context.Context, p *httpx.Principal) ([]SessionDTO, error) {
	rows, err := s.d.Q.IdentityListActiveSessions(ctx, p.UserID)
	if err != nil {
		return nil, err
	}
	cur, _ := s.d.Q.IdentityGetSession(ctx, p.SessionID)
	out := make([]SessionDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, SessionDTO{ID: r.ID, Audience: r.Aud, Platform: r.Platform, DeviceName: r.DeviceName, UserAgent: r.UserAgent,
			LastUsedAt: r.LastUsedAt, CreatedAt: r.CreatedAt, Current: r.FamilyID == cur.FamilyID})
	}
	return out, nil
}

// RevokeSession signs out one device.
func (s *Service) RevokeSession(ctx context.Context, p *httpx.Principal, id uuid.UUID) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		n, err := tx.Q.IdentityRevokeSession(ctx, store.IdentityRevokeSessionParams{ID: id, UserID: p.UserID, RevokedReason: strOrNil("logout")})
		if err != nil {
			return err
		}
		if n == 0 {
			return httpx.NotFound("session")
		}
		s.event(ctx, tx.Q, &p.UserID, "session_revoked", map[string]any{"by": "user"})
		tx.Notify(realtime.ChannelSessions, map[string]any{"userId": p.UserID})
		return nil
	})
}

// RevokeOthers signs out everywhere except this device.
func (s *Service) RevokeOthers(ctx context.Context, p *httpx.Principal) (int64, error) {
	sess, err := s.d.Q.IdentityGetSession(ctx, p.SessionID)
	if err != nil {
		return 0, err
	}
	var n int64
	err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		n, err = tx.Q.IdentityRevokeOtherSessions(ctx, store.IdentityRevokeOtherSessionsParams{UserID: p.UserID, FamilyID: sess.FamilyID})
		if err != nil {
			return err
		}
		s.event(ctx, tx.Q, &p.UserID, "session_revoked", map[string]any{"by": "user", "scope": "others", "count": n})
		tx.Notify(realtime.ChannelSessions, map[string]any{"userId": p.UserID})
		return nil
	})
	return n, err
}

// AddressInput is a delivery address.
type AddressInput struct {
	Label         string   `json:"label,omitempty" maxLength:"40"`
	RecipientName string   `json:"recipientName" minLength:"2" maxLength:"120"`
	Phone         string   `json:"phone" doc:"Nigerian mobile number"`
	Line1         string   `json:"line1" minLength:"3" maxLength:"200"`
	Line2         string   `json:"line2,omitempty" maxLength:"200"`
	Landmark      string   `json:"landmark,omitempty" maxLength:"200"`
	City          string   `json:"city" minLength:"2" maxLength:"80"`
	LGA           string   `json:"lga,omitempty" maxLength:"80"`
	StateCode     string   `json:"stateCode" pattern:"^[A-Z]{2}$" example:"LA"`
	Latitude      *float64 `json:"latitude,omitempty"`
	Longitude     *float64 `json:"longitude,omitempty"`
	IsDefault     bool     `json:"isDefault,omitempty"`
}

func (a *AddressInput) normalise() (string, error) {
	phone := validate.NormalisePhone(a.Phone)
	if phone == "" {
		return "", httpx.Invalid("invalid_phone", "Enter a Nigerian mobile number.").Field("body.phone", "not a Nigerian mobile number", a.Phone)
	}
	if !validate.StateCode(a.StateCode) {
		return "", httpx.Invalid("invalid_state", "Choose a Nigerian state.").Field("body.stateCode", "unknown state", a.StateCode)
	}
	return phone, nil
}

// CreateAddress saves an address (the first one becomes the default).
func (s *Service) CreateAddress(ctx context.Context, userID uuid.UUID, in AddressInput) (*store.IdentityAddress, error) {
	phone, err := in.normalise()
	if err != nil {
		return nil, err
	}
	var out store.IdentityAddress
	err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		existing, err := tx.Q.IdentityListAddresses(ctx, &userID)
		if err != nil {
			return err
		}
		isDefault := in.IsDefault || len(existing) == 0
		if isDefault {
			if err := tx.Q.IdentityClearDefaultAddress(ctx, &userID); err != nil {
				return err
			}
		}
		out, err = tx.Q.IdentityCreateAddress(ctx, store.IdentityCreateAddressParams{UserID: &userID, Label: strOrNil(in.Label), RecipientName: strings.TrimSpace(in.RecipientName),
			Phone: phone, Line1: in.Line1, Line2: strOrNil(in.Line2), Landmark: strOrNil(in.Landmark), City: in.City, Lga: strOrNil(in.LGA),
			StateCode: in.StateCode, Latitude: in.Latitude, Longitude: in.Longitude, IsDefault: isDefault})
		return httpx.DB(err, "address")
	})
	return &out, err
}

// UpdateAddress replaces an address the user owns.
func (s *Service) UpdateAddress(ctx context.Context, userID, id uuid.UUID, in AddressInput) (*store.IdentityAddress, error) {
	phone, err := in.normalise()
	if err != nil {
		return nil, err
	}
	var out store.IdentityAddress
	err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		out, err = tx.Q.IdentityUpdateAddress(ctx, store.IdentityUpdateAddressParams{ID: id, UserID: &userID, Label: strOrNil(in.Label), RecipientName: in.RecipientName,
			Phone: phone, Line1: in.Line1, Line2: strOrNil(in.Line2), Landmark: strOrNil(in.Landmark), City: in.City, Lga: strOrNil(in.LGA), StateCode: in.StateCode})
		if err != nil {
			return httpx.DB(err, "address")
		}
		if in.IsDefault {
			if err := tx.Q.IdentityClearDefaultAddress(ctx, &userID); err != nil {
				return err
			}
			_, err = tx.Q.IdentitySetDefaultAddress(ctx, store.IdentitySetDefaultAddressParams{ID: id, UserID: &userID})
			out.IsDefault = true
		}
		return err
	})
	return &out, err
}

// ConsentInput records consent for a signed-in user or an anonymous visitor.
type ConsentInput struct {
	AnonymousID   *uuid.UUID `json:"anonymousId,omitempty" doc:"For visitors who aren't signed in"`
	Purpose       string     `json:"purpose" enum:"analytics,personalisation,marketing_sms,marketing_email,marketing_push,marketing_whatsapp,location_tracking,terms,privacy"`
	Granted       bool       `json:"granted"`
	PolicyVersion string     `json:"policyVersion" minLength:"1" maxLength:"40"`
	Source        string     `json:"source" enum:"web,app"`
}

// RecordConsent appends a consent row (NDPA proof).
func (s *Service) RecordConsent(ctx context.Context, user *uuid.UUID, in ConsentInput) error {
	if user == nil && in.AnonymousID == nil {
		return httpx.Invalid("subject_required", "Sign in or send an anonymousId.")
	}
	return s.d.Q.IdentityInsertConsent(ctx, store.IdentityInsertConsentParams{UserID: user, AnonymousID: in.AnonymousID, Purpose: in.Purpose,
		Granted: in.Granted, PolicyVersion: in.PolicyVersion, Source: in.Source})
}

// RequestPrivacy creates an export or deletion request (deletion: 7-day cancel window).
func (s *Service) RequestPrivacy(ctx context.Context, userID uuid.UUID, kind string) (*store.IdentityPrivacyRequest, error) {
	var cancelUntil *time.Time
	if kind == "delete" {
		t := time.Now().Add(7 * 24 * time.Hour)
		cancelUntil = &t
	}
	var out store.IdentityPrivacyRequest
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		var err error
		out, err = tx.Q.IdentityCreatePrivacyRequest(ctx, store.IdentityCreatePrivacyRequestParams{UserID: userID, Kind: kind, CancelUntil: cancelUntil})
		if err != nil {
			return err
		}
		tx.Emit("user", userID, "privacy."+kind+"_requested", map[string]any{"requestId": out.ID})
		return tx.Audit("privacy."+kind+"_requested", "user", userID.String(), nil)
	})
	return &out, err
}
