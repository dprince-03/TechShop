package identity

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/notify"
	"github.com/dprince-03/techshop/backend/internal/rbac"
	"github.com/dprince-03/techshop/backend/internal/realtime"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/pkg/crypto"
)

// NewStaffInput adds a staff member (HR).
type NewStaffInput struct {
	Email      string `json:"email" format:"email"`
	FirstName  string `json:"firstName" minLength:"1" maxLength:"80"`
	LastName   string `json:"lastName" minLength:"1" maxLength:"80"`
	EmployeeNo string `json:"employeeNo" minLength:"1" maxLength:"40"`
	Department string `json:"department" minLength:"1" maxLength:"80"`
	JobTitle   string `json:"jobTitle" minLength:"1" maxLength:"80"`
}

// NewStaffResult returns the invite link only in development (it is emailed otherwise).
type NewStaffResult struct {
	UserID    uuid.UUID `json:"userId"`
	InviteURL string    `json:"inviteUrl,omitempty"`
}

// CreateStaff creates the user, the employee record and a 72-hour single-use invite.
func (s *Service) CreateStaff(ctx context.Context, in NewStaffInput) (*NewStaffResult, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	token := crypto.RandomToken(32)
	link := "http://staff.techshop.localhost/invite?token=" + token
	var res NewStaffResult
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		u, err := tx.Q.IdentityGetUserByEmail(ctx, &email)
		if err != nil {
			u, err = tx.Q.IdentityCreateUser(ctx, store.IdentityCreateUserParams{Email: &email, FirstName: in.FirstName, LastName: in.LastName})
			if err != nil {
				return httpx.DB(err, "user")
			}
		}
		today := time.Now()
		if _, err := tx.Q.IdentityCreateStaffMember(ctx, store.IdentityCreateStaffMemberParams{UserID: u.ID, EmployeeNo: in.EmployeeNo, Department: in.Department, JobTitle: in.JobTitle, HiredOn: &today}); err != nil {
			return httpx.DB(err, "staff member")
		}
		p := httpx.MustPrincipal(ctx)
		if _, err := tx.Q.IdentityCreateInvite(ctx, store.IdentityCreateInviteParams{StaffUserID: u.ID, TokenHash: crypto.SHA256(token), CreatedBy: p.UserID, ExpiresAt: time.Now().Add(72 * time.Hour)}); err != nil {
			return err
		}
		res.UserID = u.ID
		if err := tx.Audit("staff.created", "staff_member", u.ID.String(), map[string]any{"employeeNo": in.EmployeeNo, "department": in.Department}); err != nil {
			return err
		}
		return s.notify.Send(ctx, tx, notify.Msg{UserID: &u.ID, To: email, Channel: "email", Category: notify.CatAccount, Template: "staff_invite", Data: map[string]any{"link": link}, Key: "invite:" + u.ID.String() + ":" + token[:8]})
	})
	if err == nil && s.d.Cfg.IsDevelopment() {
		res.InviteURL = link
	}
	return &res, err
}

// InviteStart is returned when a staff invite is accepted: the authenticator setup.
type InviteStart struct {
	EnrollToken string `json:"enrollToken" doc:"Send with the first authenticator code to finish"`
	TOTPSetup
}

// AcceptInvite sets the password and starts authenticator enrolment.
func (s *Service) AcceptInvite(ctx context.Context, token, password string) (*InviteStart, error) {
	if err := auth.CheckPasswordPolicy(password); err != nil {
		return nil, httpx.Invalid("weak_password", err.Error())
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	var userID uuid.UUID
	var email string
	err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		inv, err := tx.Q.IdentityGetInvite(ctx, crypto.SHA256(token))
		if err != nil {
			return httpx.Invalid("invite_invalid", "This invite link is invalid, used or expired. Ask HR for a new one.")
		}
		if err := tx.Q.IdentityUseInvite(ctx, inv.ID); err != nil {
			return err
		}
		if err := tx.Q.IdentitySetPassword(ctx, store.IdentitySetPasswordParams{ID: inv.StaffUserID, PasswordHash: &hash}); err != nil {
			return err
		}
		u, err := tx.Q.IdentityGetUser(ctx, inv.StaffUserID)
		if err != nil {
			return err
		}
		userID = u.ID
		if u.Email != nil {
			email = *u.Email
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	setup, err := s.SetupTOTP(ctx, userID, email)
	if err != nil {
		return nil, err
	}
	return &InviteStart{EnrollToken: s.mfaToken(userID, "enroll"), TOTPSetup: *setup}, nil
}

// ConfirmInvite finishes enrolment with the first authenticator code and signs in.
func (s *Service) ConfirmInvite(ctx context.Context, enrollToken, code string, dev Device) (*TokenPair, []string, error) {
	userID, aud, ok := s.parseMFAToken(enrollToken)
	if !ok || aud != "enroll" {
		// Enrol tokens get a longer window: re-issue isn't needed for the demo; ask to restart.
		return nil, nil, httpx.Unauthenticated("This setup step has expired. Use your invite link again.")
	}
	codes, err := s.ConfirmTOTP(ctx, userID, code)
	if err != nil {
		return nil, nil, err
	}
	dev.Audience = auth.AudStaff
	var tp *TokenPair
	err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		now := time.Now()
		tp, err = s.createSession(ctx, tx, userID, dev, []string{"pwd", "totp"}, &now)
		return err
	})
	return tp, codes, err
}

// GrantRole gives a staff member a role. Nobody grants roles to themselves.
func (s *Service) GrantRole(ctx context.Context, target uuid.UUID, roleKey string) error {
	p := httpx.MustPrincipal(ctx)
	if p.UserID == target {
		return httpx.Forbidden("You can't change your own roles; another admin must.")
	}
	role, err := s.d.Q.IdentityGetRoleByKey(ctx, roleKey)
	if err != nil {
		return httpx.DB(err, "role")
	}
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if err := tx.Q.IdentityGrantRole(ctx, store.IdentityGrantRoleParams{UserID: target, RoleID: role.ID, GrantedBy: &p.UserID}); err != nil {
			return httpx.DB(err, "staff role")
		}
		s.event(ctx, tx.Q, &target, "role_granted", map[string]any{"role": roleKey, "by": p.UserID})
		tx.Notify(realtime.ChannelAuthz, map[string]any{"userId": target})
		return tx.Audit("staff.role_granted", "staff_member", target.String(), map[string]any{"role": roleKey})
	})
}

// RevokeRole removes a role; it takes effect on the target's very next request.
func (s *Service) RevokeRole(ctx context.Context, target uuid.UUID, roleKey string) error {
	p := httpx.MustPrincipal(ctx)
	if p.UserID == target {
		return httpx.Forbidden("You can't change your own roles; another admin must.")
	}
	role, err := s.d.Q.IdentityGetRoleByKey(ctx, roleKey)
	if err != nil {
		return httpx.DB(err, "role")
	}
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		n, err := tx.Q.IdentityRevokeRole(ctx, store.IdentityRevokeRoleParams{UserID: target, RoleID: role.ID})
		if err != nil {
			return err
		}
		if n == 0 {
			return httpx.NotFound("role assignment")
		}
		s.event(ctx, tx.Q, &target, "role_revoked", map[string]any{"role": roleKey, "by": p.UserID})
		tx.Notify(realtime.ChannelAuthz, map[string]any{"userId": target})
		return tx.Audit("staff.role_revoked", "staff_member", target.String(), map[string]any{"role": roleKey})
	})
}

// ExitStaff removes every role, revokes every session and push token in one transaction.
func (s *Service) ExitStaff(ctx context.Context, target uuid.UUID) error {
	p := httpx.MustPrincipal(ctx)
	if p.UserID == target {
		return httpx.Forbidden("You can't exit yourself.")
	}
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if _, err := tx.Q.IdentityGetStaffMember(ctx, target); err != nil {
			return httpx.DB(err, "staff member")
		}
		if err := tx.Q.IdentitySetStaffStatus(ctx, store.IdentitySetStaffStatusParams{UserID: target, Status: "exited"}); err != nil {
			return err
		}
		roles, err := tx.Q.IdentityRevokeAllRoles(ctx, target)
		if err != nil {
			return err
		}
		if _, err := tx.Q.IdentityRevokeAllSessions(ctx, store.IdentityRevokeAllSessionsParams{UserID: target, RevokedReason: strOrNil("staff_exited")}); err != nil {
			return err
		}
		if err := tx.Q.MessagingRevokeUserPushTokens(ctx, target); err != nil {
			return err
		}
		s.event(ctx, tx.Q, &target, "staff_exited", map[string]any{"rolesRemoved": roles, "by": p.UserID})
		tx.Notify(realtime.ChannelAuthz, map[string]any{"userId": target})
		tx.Notify(realtime.ChannelSessions, map[string]any{"userId": target})
		return tx.Audit("staff.exited", "staff_member", target.String(), map[string]any{"rolesRemoved": roles})
	})
}

// ResetMFA clears another person's authenticator and signs them out everywhere.
func (s *Service) ResetMFA(ctx context.Context, target uuid.UUID) error {
	p := httpx.MustPrincipal(ctx)
	if p.UserID == target {
		return httpx.Forbidden("You can't reset your own authenticator.")
	}
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if err := tx.Q.IdentityDeleteTOTP(ctx, target); err != nil {
			return err
		}
		if err := tx.Q.IdentityDeleteRecoveryCodes(ctx, target); err != nil {
			return err
		}
		if _, err := tx.Q.IdentityRevokeAllSessions(ctx, store.IdentityRevokeAllSessionsParams{UserID: target, RevokedReason: strOrNil("admin")}); err != nil {
			return err
		}
		s.event(ctx, tx.Q, &target, "mfa_removed", map[string]any{"by": p.UserID, "reason": "admin_reset"})
		tx.Notify(realtime.ChannelSessions, map[string]any{"userId": target})
		return tx.Audit("staff.mfa_reset", "user", target.String(), nil)
	})
}

// SetRolePermissions replaces a role's permissions (only known catalogue keys).
func (s *Service) SetRolePermissions(ctx context.Context, roleKey, name string, perms []string) error {
	known := map[string]bool{}
	for _, k := range rbac.Keys() {
		known[k] = true
	}
	for _, k := range perms {
		if !known[k] {
			return httpx.Invalid("unknown_permission", "Unknown permission "+k)
		}
	}
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		role, err := tx.Q.IdentityUpsertRole(ctx, store.IdentityUpsertRoleParams{Key: roleKey, Name: name})
		if err != nil {
			return httpx.DB(err, "role")
		}
		if role.Key == "admin" && !contains(perms, "admin.roles") {
			return httpx.Forbidden("The admin role must keep admin.roles.")
		}
		if err := tx.Q.IdentityClearRolePermissions(ctx, role.ID); err != nil {
			return err
		}
		if err := tx.Q.IdentityAddRolePermissions(ctx, store.IdentityAddRolePermissionsParams{RoleID: role.ID, Keys: perms}); err != nil {
			return err
		}
		users, err := tx.Q.IdentityUsersWithRole(ctx, role.ID)
		if err != nil {
			return err
		}
		for _, u := range users {
			tx.Notify(realtime.ChannelAuthz, map[string]any{"userId": u})
		}
		return tx.Audit("role.permissions_set", "role", roleKey, map[string]any{"permissions": perms})
	})
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// Jobs implements kit.Module: deletion processing after the cancel window.
func (s *Service) Jobs(reg *jobs.Registry) {
	river.AddWorker(reg.Workers, &privacyWorker{s: s})
	reg.Every(time.Hour, func() (river.JobArgs, *river.InsertOpts) {
		return privacyArgs{}, &river.InsertOpts{Queue: jobs.QueueDefault}
	})
}

type privacyArgs struct{}

func (privacyArgs) Kind() string { return "identity.process_privacy_deletions" }

type privacyWorker struct {
	river.WorkerDefaults[privacyArgs]
	s *Service
}

// Work anonymises accounts whose deletion window has passed (orders and ledger are kept without
// personal data; identity-access.md §11).
func (w *privacyWorker) Work(ctx context.Context, _ *river.Job[privacyArgs]) error {
	due, err := w.s.d.Q.IdentityDuePrivacyDeletions(ctx)
	if err != nil {
		return err
	}
	for _, r := range due {
		err := w.s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
			if err := tx.Q.IdentityAnonymiseUser(ctx, r.UserID); err != nil {
				return err
			}
			if _, err := tx.Q.IdentityRevokeAllSessions(ctx, store.IdentityRevokeAllSessionsParams{UserID: r.UserID, RevokedReason: strOrNil("admin")}); err != nil {
				return err
			}
			if err := tx.Q.IdentityDeleteTOTP(ctx, r.UserID); err != nil {
				return err
			}
			if err := tx.Q.MessagingRevokeUserPushTokens(ctx, r.UserID); err != nil {
				return err
			}
			return tx.Q.IdentitySetPrivacyRequestStatus(ctx, store.IdentitySetPrivacyRequestStatusParams{ID: r.ID, Status: "completed"})
		})
		if err != nil {
			return err
		}
	}
	return nil
}
