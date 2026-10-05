package identity

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/rbac"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/pkg/crypto"
)

// Audiences of people who manage their own account.
var accountAuds = []string{auth.AudMarket, auth.AudWholesale, auth.AudCustomerApp, auth.AudSeller, auth.AudStaff, auth.AudLogistics}
var staffAud = []string{auth.AudStaff}

type tokensOut struct{ Body *TokenPair }

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	reg := func(op httpx.Op) httpx.Op {
		op.Tag = map[bool]string{true: op.Tag, false: "Auth"}[op.Tag != ""]
		return op
	}

	// ---------- Sign-in ----------
	httpx.Register(g, reg(httpx.Op{ID: "requestOtp", Method: http.MethodPost, Path: "/api/v1/auth/otp/request", Summary: "Text a 6-digit sign-in code"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Phone   string `json:"phone" example:"0803 123 4567"`
				Purpose string `json:"purpose,omitempty" enum:"login,signup" default:"login"`
			}
		}) (*struct {
			Body struct {
				ChallengeID uuid.UUID `json:"challengeId"`
				ResendAfter int       `json:"resendAfter" doc:"Seconds before another code may be requested"`
			}
		}, error) {
			purpose := in.Body.Purpose
			if purpose == "" {
				purpose = "login"
			}
			id, err := s.RequestOTP(ctx, in.Body.Phone, purpose)
			if err != nil {
				return nil, err
			}
			out := &struct {
				Body struct {
					ChallengeID uuid.UUID `json:"challengeId"`
					ResendAfter int       `json:"resendAfter" doc:"Seconds before another code may be requested"`
				}
			}{}
			out.Body.ChallengeID, out.Body.ResendAfter = id, 60
			return out, nil
		})

	httpx.Register(g, reg(httpx.Op{ID: "verifyOtp", Method: http.MethodPost, Path: "/api/v1/auth/otp/verify", Summary: "Check the code and sign in"}),
		func(ctx context.Context, in *struct {
			Body struct {
				ChallengeID uuid.UUID `json:"challengeId"`
				Code        string    `json:"code" pattern:"^[0-9]{6}$"`
				Device
			}
		}) (*struct{ Body *VerifyOTPResult }, error) {
			res, err := s.VerifyOTP(ctx, in.Body.ChallengeID, in.Body.Code, in.Body.Device)
			return &struct{ Body *VerifyOTPResult }{res}, err
		})

	httpx.Register(g, reg(httpx.Op{ID: "completeSignup", Method: http.MethodPost, Path: "/api/v1/auth/otp/signup", Summary: "Create the account for a verified number"}),
		func(ctx context.Context, in *struct {
			Body struct {
				SignupToken string `json:"signupToken"`
				FirstName   string `json:"firstName" minLength:"1" maxLength:"80"`
				LastName    string `json:"lastName" minLength:"1" maxLength:"80"`
				Device
			}
		}) (*tokensOut, error) {
			tp, err := s.CompleteSignup(ctx, in.Body.SignupToken, in.Body.FirstName, in.Body.LastName, in.Body.Device)
			return &tokensOut{tp}, err
		})

	httpx.Register(g, reg(httpx.Op{ID: "passwordLogin", Method: http.MethodPost, Path: "/api/v1/auth/password", Summary: "Sign in with email or phone and password"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Identifier string `json:"identifier" minLength:"3" maxLength:"254" doc:"Email or Nigerian phone number"`
				Password   string `json:"password" minLength:"1" maxLength:"128"`
				Device
			}
		}) (*struct{ Body *PasswordResult }, error) {
			res, err := s.PasswordLogin(ctx, in.Body.Identifier, in.Body.Password, in.Body.Device)
			return &struct{ Body *PasswordResult }{res}, err
		})

	httpx.Register(g, reg(httpx.Op{ID: "verifyTotp", Method: http.MethodPost, Path: "/api/v1/auth/mfa/totp", Summary: "Finish sign-in with an authenticator code"}),
		func(ctx context.Context, in *struct {
			Body struct {
				MFAToken string `json:"mfaToken"`
				Code     string `json:"code" pattern:"^[0-9]{6}$"`
				Device
			}
		}) (*tokensOut, error) {
			tp, err := s.VerifyTOTP(ctx, in.Body.MFAToken, in.Body.Code, in.Body.Device)
			return &tokensOut{tp}, err
		})

	httpx.Register(g, reg(httpx.Op{ID: "verifyRecoveryCode", Method: http.MethodPost, Path: "/api/v1/auth/mfa/recovery", Summary: "Finish sign-in with a recovery code"}),
		func(ctx context.Context, in *struct {
			Body struct {
				MFAToken string `json:"mfaToken"`
				Code     string `json:"code" minLength:"9" maxLength:"9"`
				Device
			}
		}) (*tokensOut, error) {
			tp, err := s.VerifyRecoveryCode(ctx, in.Body.MFAToken, in.Body.Code, in.Body.Device)
			return &tokensOut{tp}, err
		})

	httpx.Register(g, reg(httpx.Op{ID: "verifyNewDeviceOtp", Method: http.MethodPost, Path: "/api/v1/auth/mfa/otp", Summary: "Finish a seller sign-in on a new device"}),
		func(ctx context.Context, in *struct {
			Body struct {
				MFAToken    string    `json:"mfaToken"`
				ChallengeID uuid.UUID `json:"challengeId"`
				Code        string    `json:"code" pattern:"^[0-9]{6}$"`
				Device
			}
		}) (*tokensOut, error) {
			tp, err := s.VerifyNewDeviceOTP(ctx, in.Body.MFAToken, in.Body.ChallengeID, in.Body.Code, in.Body.Device)
			return &tokensOut{tp}, err
		})

	httpx.Register(g, reg(httpx.Op{ID: "stepUp", Method: http.MethodPost, Path: "/api/v1/auth/mfa/step-up", Auth: accountAuds, Summary: "Re-confirm your authenticator before a sensitive action"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Code string `json:"code" pattern:"^[0-9]{6}$"`
			}
		}) (*tokensOut, error) {
			tp, err := s.StepUp(ctx, httpx.MustPrincipal(ctx), in.Body.Code)
			return &tokensOut{tp}, err
		})

	httpx.Register(g, reg(httpx.Op{ID: "refreshTokens", Method: http.MethodPost, Path: "/api/v1/auth/refresh", Summary: "Rotate the refresh token"}),
		func(ctx context.Context, in *struct {
			Body struct {
				RefreshToken string `json:"refreshToken" minLength:"20"`
			}
		}) (*tokensOut, error) {
			tp, err := s.Refresh(ctx, in.Body.RefreshToken)
			return &tokensOut{tp}, err
		})

	httpx.Register(g, reg(httpx.Op{ID: "logout", Method: http.MethodPost, Path: "/api/v1/auth/logout", Auth: accountAuds, Optional: true, Status: 204, Summary: "Sign out this device"}),
		func(ctx context.Context, in *struct {
			Body struct {
				RefreshToken string `json:"refreshToken,omitempty"`
			}
		}) (*struct{}, error) {
			return nil, s.Logout(ctx, httpx.PrincipalFrom(ctx), in.Body.RefreshToken)
		})

	httpx.Register(g, reg(httpx.Op{ID: "acceptInvite", Method: http.MethodPost, Path: "/api/v1/auth/invite/accept", Summary: "Staff: set a password and start authenticator setup"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Token    string `json:"token" minLength:"20"`
				Password string `json:"password" minLength:"10" maxLength:"128"`
			}
		}) (*struct{ Body *InviteStart }, error) {
			res, err := s.AcceptInvite(ctx, in.Body.Token, in.Body.Password)
			return &struct{ Body *InviteStart }{res}, err
		})

	httpx.Register(g, reg(httpx.Op{ID: "confirmInvite", Method: http.MethodPost, Path: "/api/v1/auth/invite/confirm", Summary: "Staff: confirm the first authenticator code and sign in"}),
		func(ctx context.Context, in *struct {
			Body struct {
				EnrollToken string `json:"enrollToken"`
				Code        string `json:"code" pattern:"^[0-9]{6}$"`
				Device
			}
		}) (*struct {
			Body struct {
				Tokens        *TokenPair `json:"tokens"`
				RecoveryCodes []string   `json:"recoveryCodes" doc:"Shown once"`
			}
		}, error) {
			tp, codes, err := s.ConfirmInvite(ctx, in.Body.EnrollToken, in.Body.Code, in.Body.Device)
			out := &struct {
				Body struct {
					Tokens        *TokenPair `json:"tokens"`
					RecoveryCodes []string   `json:"recoveryCodes" doc:"Shown once"`
				}
			}{}
			out.Body.Tokens, out.Body.RecoveryCodes = tp, codes
			return out, err
		})

	// ---------- Account ----------
	acc := func(op httpx.Op) httpx.Op { op.Tag = "Account"; op.Auth = accountAuds; return op }
	httpx.Register(g, acc(httpx.Op{ID: "getMe", Method: http.MethodGet, Path: "/api/v1/me", Summary: "Your profile and memberships"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body *Me }, error) {
			me, err := s.GetMe(ctx, httpx.MustPrincipal(ctx).UserID)
			return &struct{ Body *Me }{me}, err
		})

	httpx.Register(g, acc(httpx.Op{ID: "updateMe", Method: http.MethodPatch, Path: "/api/v1/me", Summary: "Update your name or marketing choice"}),
		func(ctx context.Context, in *struct {
			Body struct {
				FirstName      *string `json:"firstName,omitempty" minLength:"1" maxLength:"80"`
				LastName       *string `json:"lastName,omitempty" minLength:"1" maxLength:"80"`
				MarketingOptIn *bool   `json:"marketingOptIn,omitempty"`
			}
		}) (*struct{ Body *Me }, error) {
			p := httpx.MustPrincipal(ctx)
			if _, err := s.d.Q.IdentityUpdateProfile(ctx, store.IdentityUpdateProfileParams{ID: p.UserID, FirstName: in.Body.FirstName, LastName: in.Body.LastName, MarketingOptIn: in.Body.MarketingOptIn}); err != nil {
				return nil, err
			}
			me, err := s.GetMe(ctx, p.UserID)
			return &struct{ Body *Me }{me}, err
		})

	httpx.Register(g, acc(httpx.Op{ID: "changePassword", Method: http.MethodPost, Path: "/api/v1/me/password", Status: 204, Summary: "Change or set your password (signs out other devices)"}),
		func(ctx context.Context, in *struct {
			Body struct {
				CurrentPassword string `json:"currentPassword,omitempty" maxLength:"128"`
				NewPassword     string `json:"newPassword" minLength:"10" maxLength:"128"`
			}
		}) (*struct{}, error) {
			return nil, s.ChangePassword(ctx, httpx.MustPrincipal(ctx), in.Body.CurrentPassword, in.Body.NewPassword)
		})

	httpx.Register(g, acc(httpx.Op{ID: "listSessions", Method: http.MethodGet, Path: "/api/v1/me/sessions", Summary: "Signed-in devices"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []SessionDTO }, error) {
			rows, err := s.ListSessions(ctx, httpx.MustPrincipal(ctx))
			return &struct{ Body []SessionDTO }{rows}, err
		})
	httpx.Register(g, acc(httpx.Op{ID: "revokeSession", Method: http.MethodDelete, Path: "/api/v1/me/sessions/{id}", Status: 204, Summary: "Sign out one device"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{}, error) {
			return nil, s.RevokeSession(ctx, httpx.MustPrincipal(ctx), in.ID)
		})
	httpx.Register(g, acc(httpx.Op{ID: "revokeOtherSessions", Method: http.MethodPost, Path: "/api/v1/me/sessions/revoke-others", Summary: "Sign out everywhere else"}),
		func(ctx context.Context, _ *struct{}) (*struct {
			Body struct {
				Revoked int64 `json:"revoked"`
			}
		}, error) {
			n, err := s.RevokeOthers(ctx, httpx.MustPrincipal(ctx))
			out := &struct {
				Body struct {
					Revoked int64 `json:"revoked"`
				}
			}{}
			out.Body.Revoked = n
			return out, err
		})

	httpx.Register(g, acc(httpx.Op{ID: "setupTotp", Method: http.MethodPost, Path: "/api/v1/me/mfa/totp/setup", Summary: "Start authenticator setup (secret + QR URL)"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body *TOTPSetup }, error) {
			p := httpx.MustPrincipal(ctx)
			u, err := s.d.Q.IdentityGetUser(ctx, p.UserID)
			if err != nil {
				return nil, err
			}
			account := u.FirstName
			if u.Email != nil {
				account = *u.Email
			} else if u.Phone != nil {
				account = *u.Phone
			}
			res, err := s.SetupTOTP(ctx, p.UserID, account)
			return &struct{ Body *TOTPSetup }{res}, err
		})
	httpx.Register(g, acc(httpx.Op{ID: "confirmTotp", Method: http.MethodPost, Path: "/api/v1/me/mfa/totp/confirm", Summary: "Turn on the authenticator; returns recovery codes once"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Code string `json:"code" pattern:"^[0-9]{6}$"`
			}
		}) (*struct {
			Body struct {
				RecoveryCodes []string `json:"recoveryCodes"`
			}
		}, error) {
			codes, err := s.ConfirmTOTP(ctx, httpx.MustPrincipal(ctx).UserID, in.Body.Code)
			out := &struct {
				Body struct {
					RecoveryCodes []string `json:"recoveryCodes"`
				}
			}{}
			out.Body.RecoveryCodes = codes
			return out, err
		})
	httpx.Register(g, acc(httpx.Op{ID: "removeTotp", Method: http.MethodDelete, Path: "/api/v1/me/mfa/totp", StepUp: true, Status: 204, Summary: "Turn off the authenticator (needs a recent step-up)"}),
		func(ctx context.Context, _ *struct{}) (*struct{}, error) {
			return nil, s.RemoveTOTP(ctx, httpx.MustPrincipal(ctx).UserID)
		})

	httpx.Register(g, acc(httpx.Op{ID: "listSecurityEvents", Method: http.MethodGet, Path: "/api/v1/me/security-events", Summary: "Recent sign-ins and security changes"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.IdentityAuthEvent }, error) {
			p := httpx.MustPrincipal(ctx)
			rows, err := s.d.Q.IdentityListAuthEvents(ctx, store.IdentityListAuthEventsParams{UserID: &p.UserID, Limit: 50})
			return &struct{ Body []store.IdentityAuthEvent }{rows}, err
		})

	// ---------- Addresses ----------
	addr := func(op httpx.Op) httpx.Op { op.Tag = "Addresses"; op.Auth = accountAuds; return op }
	httpx.Register(g, addr(httpx.Op{ID: "listAddresses", Method: http.MethodGet, Path: "/api/v1/me/addresses", Summary: "Saved addresses"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.IdentityAddress }, error) {
			p := httpx.MustPrincipal(ctx)
			rows, err := s.d.Q.IdentityListAddresses(ctx, &p.UserID)
			return &struct{ Body []store.IdentityAddress }{rows}, err
		})
	httpx.Register(g, addr(httpx.Op{ID: "createAddress", Method: http.MethodPost, Path: "/api/v1/me/addresses", Status: 201, Summary: "Add an address"}),
		func(ctx context.Context, in *struct{ Body AddressInput }) (*struct{ Body *store.IdentityAddress }, error) {
			a, err := s.CreateAddress(ctx, httpx.MustPrincipal(ctx).UserID, in.Body)
			return &struct{ Body *store.IdentityAddress }{a}, err
		})
	httpx.Register(g, addr(httpx.Op{ID: "updateAddress", Method: http.MethodPut, Path: "/api/v1/me/addresses/{id}", Summary: "Replace an address"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body AddressInput
		}) (*struct{ Body *store.IdentityAddress }, error) {
			a, err := s.UpdateAddress(ctx, httpx.MustPrincipal(ctx).UserID, in.ID, in.Body)
			return &struct{ Body *store.IdentityAddress }{a}, err
		})
	httpx.Register(g, addr(httpx.Op{ID: "deleteAddress", Method: http.MethodDelete, Path: "/api/v1/me/addresses/{id}", Status: 204, Summary: "Delete an address"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{}, error) {
			p := httpx.MustPrincipal(ctx)
			n, err := s.d.Q.IdentityDeleteAddress(ctx, store.IdentityDeleteAddressParams{ID: in.ID, UserID: &p.UserID})
			if err != nil {
				return nil, httpx.DB(err, "address")
			}
			if n == 0 {
				return nil, httpx.NotFound("address")
			}
			return nil, nil
		})
	httpx.Register(g, httpx.Op{ID: "listStates", Method: http.MethodGet, Path: "/api/v1/states", Tag: "Reference", Summary: "Nigerian states and the FCT"},
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.IdentityNigerianState }, error) {
			rows, err := s.d.Q.IdentityListStates(ctx)
			return &struct{ Body []store.IdentityNigerianState }{rows}, err
		})

	// ---------- Consents & privacy (NDPA) ----------
	httpx.Register(g, httpx.Op{ID: "recordConsent", Method: http.MethodPost, Path: "/api/v1/consents", Tag: "Privacy", Auth: accountAuds, Optional: true, Status: 204, Summary: "Give or withdraw consent (signed in or anonymous)"},
		func(ctx context.Context, in *struct{ Body ConsentInput }) (*struct{}, error) {
			var user *uuid.UUID
			if p := httpx.PrincipalFrom(ctx); p != nil {
				user = &p.UserID
			}
			return nil, s.RecordConsent(ctx, user, in.Body)
		})
	httpx.Register(g, httpx.Op{ID: "listConsents", Method: http.MethodGet, Path: "/api/v1/me/consents", Tag: "Privacy", Auth: accountAuds, Summary: "Your current consent choices"},
		func(ctx context.Context, _ *struct{}) (*struct {
			Body []store.IdentityLatestConsentsRow
		}, error) {
			p := httpx.MustPrincipal(ctx)
			rows, err := s.d.Q.IdentityLatestConsents(ctx, store.IdentityLatestConsentsParams{UserID: &p.UserID})
			return &struct {
				Body []store.IdentityLatestConsentsRow
			}{rows}, err
		})
	httpx.Register(g, httpx.Op{ID: "requestPrivacy", Method: http.MethodPost, Path: "/api/v1/me/privacy/{kind}", Tag: "Privacy", Auth: accountAuds, Status: 202, Summary: "Request a data export or account deletion (7-day cancel window)"},
		func(ctx context.Context, in *struct {
			Kind string `path:"kind" enum:"export,delete"`
		}) (*struct{ Body *store.IdentityPrivacyRequest }, error) {
			r, err := s.RequestPrivacy(ctx, httpx.MustPrincipal(ctx).UserID, in.Kind)
			return &struct{ Body *store.IdentityPrivacyRequest }{r}, err
		})
	httpx.Register(g, httpx.Op{ID: "cancelPrivacyRequest", Method: http.MethodDelete, Path: "/api/v1/me/privacy/requests/{id}", Tag: "Privacy", Auth: accountAuds, Status: 204, Summary: "Cancel a pending deletion"},
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{}, error) {
			p := httpx.MustPrincipal(ctx)
			n, err := s.d.Q.IdentityCancelPrivacyRequest(ctx, store.IdentityCancelPrivacyRequestParams{ID: in.ID, UserID: p.UserID})
			if err == nil && n == 0 {
				return nil, httpx.NotFound("pending request")
			}
			return nil, err
		})

	s.adminRoutes(g)
	s.internalRoutes(r.Internal)
}

func (s *Service) adminRoutes(g *httpx.Guard) {
	adm := func(op httpx.Op) httpx.Op { op.Tag = "Staff · Admin"; op.Auth = staffAud; return op }
	httpx.Register(g, adm(httpx.Op{ID: "staffListPermissions", Method: http.MethodGet, Path: "/api/v1/staff/admin/permissions", Perm: "admin.audit_view", Summary: "Permission catalogue"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.IdentityPermission }, error) {
			rows, err := s.d.Q.IdentityListPermissions(ctx)
			return &struct{ Body []store.IdentityPermission }{rows}, err
		})
	httpx.Register(g, adm(httpx.Op{ID: "staffListRoles", Method: http.MethodGet, Path: "/api/v1/staff/admin/roles", Perm: "admin.audit_view", Summary: "Roles and their permissions"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.IdentityListRolesRow }, error) {
			rows, err := s.d.Q.IdentityListRoles(ctx)
			return &struct{ Body []store.IdentityListRolesRow }{rows}, err
		})
	httpx.Register(g, adm(httpx.Op{ID: "staffSetRole", Method: http.MethodPut, Path: "/api/v1/staff/admin/roles/{key}", Perm: "admin.roles", Status: 204, Summary: "Create a role or replace its permissions"}),
		func(ctx context.Context, in *struct {
			Key  string `path:"key" pattern:"^[a-z][a-z0-9_]*$"`
			Body struct {
				Name        string   `json:"name" minLength:"2" maxLength:"80"`
				Permissions []string `json:"permissions"`
			}
		}) (*struct{}, error) {
			return nil, s.SetRolePermissions(ctx, in.Key, in.Body.Name, in.Body.Permissions)
		})
	httpx.Register(g, adm(httpx.Op{ID: "staffListMembers", Method: http.MethodGet, Path: "/api/v1/staff/hr/members", Perm: "hr.view", Summary: "Staff members with roles and MFA status"}),
		func(ctx context.Context, _ *struct{}) (*struct {
			Body []store.IdentityListStaffMembersRow
		}, error) {
			rows, err := s.d.Q.IdentityListStaffMembers(ctx)
			return &struct {
				Body []store.IdentityListStaffMembersRow
			}{rows}, err
		})
	httpx.Register(g, adm(httpx.Op{ID: "staffCreateMember", Method: http.MethodPost, Path: "/api/v1/staff/hr/members", Perm: "hr.manage", Status: 201, Summary: "Add a staff member and send an invite"}),
		func(ctx context.Context, in *struct{ Body NewStaffInput }) (*struct{ Body *NewStaffResult }, error) {
			res, err := s.CreateStaff(ctx, in.Body)
			return &struct{ Body *NewStaffResult }{res}, err
		})
	httpx.Register(g, adm(httpx.Op{ID: "staffExitMember", Method: http.MethodPost, Path: "/api/v1/staff/hr/members/{id}/exit", Perm: "hr.manage", Status: 204, Summary: "Exit a staff member (roles and sessions removed at once)"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{}, error) {
			return nil, s.ExitStaff(ctx, in.ID)
		})
	httpx.Register(g, adm(httpx.Op{ID: "staffGrantRole", Method: http.MethodPost, Path: "/api/v1/staff/admin/members/{id}/roles", Perm: "admin.roles", Status: 204, Summary: "Grant a role (never to yourself)"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Role string `json:"role"`
			}
		}) (*struct{}, error) {
			return nil, s.GrantRole(ctx, in.ID, in.Body.Role)
		})
	httpx.Register(g, adm(httpx.Op{ID: "staffRevokeRole", Method: http.MethodDelete, Path: "/api/v1/staff/admin/members/{id}/roles/{role}", Perm: "admin.roles", Status: 204, Summary: "Remove a role (effective on the next request)"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Role string    `path:"role"`
		}) (*struct{}, error) {
			return nil, s.RevokeRole(ctx, in.ID, in.Role)
		})
	httpx.Register(g, adm(httpx.Op{ID: "staffResetMfa", Method: http.MethodPost, Path: "/api/v1/staff/admin/members/{id}/reset-mfa", Perm: "admin.mfa_reset", Status: 204, Summary: "Reset someone else's authenticator"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{}, error) {
			return nil, s.ResetMFA(ctx, in.ID)
		})
	httpx.Register(g, adm(httpx.Op{ID: "staffListAuthEvents", Method: http.MethodGet, Path: "/api/v1/staff/admin/auth-events", Perm: "admin.audit_view", Summary: "Security events"}),
		func(ctx context.Context, in *struct {
			UserID string `query:"userId"`
			Kind   string `query:"kind"`
			Before int64  `query:"before"`
		}) (*struct{ Body []store.IdentityAuthEvent }, error) {
			p := store.IdentityListAuthEventsParams{Limit: 100}
			if id, err := uuid.Parse(in.UserID); err == nil {
				p.UserID = &id
			}
			if in.Kind != "" {
				p.Kind = &in.Kind
			}
			if in.Before > 0 {
				p.BeforeID = &in.Before
			}
			rows, err := s.d.Q.IdentityListAuthEvents(ctx, p)
			return &struct{ Body []store.IdentityAuthEvent }{rows}, err
		})
	httpx.Register(g, adm(httpx.Op{ID: "staffListAuditLog", Method: http.MethodGet, Path: "/api/v1/staff/admin/audit-log", Perm: "admin.audit_view", Summary: "Who changed what"}),
		func(ctx context.Context, in *struct {
			EntityType string `query:"entityType"`
			Before     int64  `query:"before"`
		}) (*struct{ Body []store.PlatformAuditLog }, error) {
			p := store.PlatformListAuditParams{Limit: 100}
			if in.EntityType != "" {
				p.EntityType = &in.EntityType
			}
			if in.Before > 0 {
				p.BeforeID = &in.Before
			}
			rows, err := s.d.Q.PlatformListAudit(ctx, p)
			return &struct{ Body []store.PlatformAuditLog }{rows}, err
		})
	httpx.Register(g, adm(httpx.Op{ID: "staffListUsers", Method: http.MethodGet, Path: "/api/v1/staff/admin/users", Perm: "admin.users", Summary: "Search accounts"}),
		func(ctx context.Context, in *struct {
			Q string `query:"q" maxLength:"80"`
		}) (*struct{ Body []userSummary }, error) {
			var q *string
			if in.Q != "" {
				q = &in.Q
			}
			rows, err := s.d.Q.IdentityListUsers(ctx, store.IdentityListUsersParams{Limit: 50, Q: q})
			out := make([]userSummary, 0, len(rows))
			for _, u := range rows {
				out = append(out, userSummary{ID: u.ID, Name: u.FirstName + " " + u.LastName, Email: u.Email, Phone: maskPtr(u.Phone), Status: u.Status, CreatedAt: u.CreatedAt})
			}
			return &struct{ Body []userSummary }{out}, err
		})
}

type userSummary struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     *string   `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty" doc:"Masked"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

func maskPtr(p *string) string {
	if p == nil || len(*p) < 8 {
		return ""
	}
	return (*p)[:7] + "***" + (*p)[len(*p)-4:]
}

// internalRoutes: client credentials for machine clients (internal listener only).
func (s *Service) internalRoutes(g *httpx.Guard) {
	httpx.Register(g, httpx.Op{ID: "serviceToken", Method: http.MethodPost, Path: "/internal/auth/token", Tag: "Internal", Summary: "Client credentials → 5-minute service token"},
		func(ctx context.Context, in *struct {
			Body struct {
				ClientID     string `json:"clientId"`
				ClientSecret string `json:"clientSecret"`
				Scope        string `json:"scope,omitempty"`
			}
		}) (*struct {
			Body struct {
				AccessToken string    `json:"accessToken"`
				ExpiresAt   time.Time `json:"expiresAt"`
			}
		}, error) {
			c, err := s.d.Q.IdentityGetServiceClient(ctx, in.Body.ClientID)
			if err != nil || !crypto.Equal(c.SecretHash, crypto.SHA256(in.Body.ClientSecret)) {
				return nil, httpx.Unauthenticated("Invalid client credentials.")
			}
			scopes := c.Scopes
			if in.Body.Scope != "" {
				scopes = nil
				for _, want := range strings.Fields(in.Body.Scope) {
					if contains(c.Scopes, want) {
						scopes = append(scopes, want)
					}
				}
			}
			tok, exp, err := s.d.Keys.Issue(auth.IssueInput{UserID: c.ID, SessionID: uuid.Nil, Audience: auth.AudInternal, Scopes: scopes, AuthTime: time.Now(), TTL: 5 * time.Minute})
			out := &struct {
				Body struct {
					AccessToken string    `json:"accessToken"`
					ExpiresAt   time.Time `json:"expiresAt"`
				}
			}{}
			out.Body.AccessToken, out.Body.ExpiresAt = tok, exp
			return out, err
		})
}

var _ kit.Module = (*Service)(nil)
var _ = rbac.Keys
