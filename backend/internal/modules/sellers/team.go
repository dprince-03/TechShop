package sellers

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/modules/notify"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/pkg/crypto"
	"github.com/dprince-03/techshop/backend/pkg/validate"
)

// Seller teams (identity-access.md §7.3; docs/plan.md 2026-10-21 decision 4). Owners and
// managers invite people by phone or email; the link is single use and expires after 7 days
// (re-sending issues a new link, like Shopify's staff invites). A manager can't invite, change
// or remove an owner, and a shop always keeps at least one owner. Finance members must have an
// authenticator before they join or use money screens.

// inviteTTL is how long an invite link works.
const inviteTTL = 7 * 24 * time.Hour

// teamRoles may see and manage the team.
var teamRoles = []string{"owner", "manager"}

// errAuthenticatorRequired: finance members must enrol an authenticator first.
func errAuthenticatorRequired(who string) error {
	return httpx.E(403, "authenticator_required", who+" set up an authenticator app first (Account › Security). Finance access needs it.")
}

// TeamMember is one person on a shop's team.
type TeamMember struct {
	UserID           uuid.UUID `json:"userId"`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	Role             string    `json:"role" enum:"owner,manager,staff,finance"`
	Contact          string    `json:"contact" doc:"Masked phone, or email"`
	HasAuthenticator bool      `json:"hasAuthenticator"`
	JoinedAt         time.Time `json:"joinedAt"`
}

// TeamInvite is an open invitation.
type TeamInvite struct {
	ID        uuid.UUID `json:"id"`
	Contact   string    `json:"contact" doc:"Masked phone, or email"`
	Role      string    `json:"role" enum:"owner,manager,staff,finance"`
	InvitedBy string    `json:"invitedBy"`
	ExpiresAt time.Time `json:"expiresAt"`
	Expired   bool      `json:"expired"`
	CreatedAt time.Time `json:"createdAt"`
	Link      string    `json:"link,omitempty" doc:"Development only; otherwise the link is only sent to the invitee"`
}

// ShopTeam is a shop's members and open invites.
type ShopTeam struct {
	Members []TeamMember `json:"members"`
	Invites []TeamInvite `json:"invites"`
}

// InviteInput invites one person by phone or email.
type InviteInput struct {
	Phone string `json:"phone,omitempty" maxLength:"20" doc:"Nigerian mobile number; or give an email"`
	Email string `json:"email,omitempty" maxLength:"254" format:"email"`
	Role  string `json:"role" enum:"owner,manager,staff,finance"`
}

// InvitePreview is what an invitee sees before accepting.
type InvitePreview struct {
	Shop      string    `json:"shop"`
	Role      string    `json:"role" enum:"owner,manager,staff,finance"`
	ExpiresAt time.Time `json:"expiresAt"`
	SentTo    string    `json:"sentTo" doc:"Masked: accept from the account with this phone or email"`
}

func maskContact(phone, email *string) string {
	if phone != nil {
		return validate.MaskPhone(*phone)
	}
	if email != nil {
		return *email
	}
	return ""
}

// target normalises the invitee's phone or email (exactly one).
func (in InviteInput) target() (phone, email *string, err error) {
	p, e := strings.TrimSpace(in.Phone), strings.ToLower(strings.TrimSpace(in.Email))
	switch {
	case p != "" && e != "", p == "" && e == "":
		return nil, nil, httpx.Invalid("phone_or_email", "Give a phone number or an email address (one of them).")
	case p != "":
		n := validate.NormalisePhone(p)
		if n == "" {
			return nil, nil, httpx.Invalid("invalid_phone", "Enter a Nigerian mobile number, for example 0803 123 4567.")
		}
		return &n, nil, nil
	default:
		if _, err := mail.ParseAddress(e); err != nil {
			return nil, nil, httpx.Invalid("invalid_email", "Enter a valid email address.")
		}
		return nil, &e, nil
	}
}

// Team lists members and open invites.
func (s *Service) Team(ctx context.Context, shopID uuid.UUID) (ShopTeam, error) {
	var out ShopTeam
	err := s.inShop(ctx, shopID, teamRoles, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
		var err error
		out, err = s.team(ctx, tx.Q, shopID)
		return err
	})
	return out, err
}

func (s *Service) team(ctx context.Context, q *store.Queries, shopID uuid.UUID) (ShopTeam, error) {
	out := ShopTeam{Members: []TeamMember{}, Invites: []TeamInvite{}}
	ms, err := q.SellersTeam(ctx, shopID)
	if err != nil {
		return out, err
	}
	for _, m := range ms {
		out.Members = append(out.Members, TeamMember{UserID: m.UserID, FirstName: m.FirstName, LastName: m.LastName, Role: m.Role,
			Contact: maskContact(m.Phone, m.Email), HasAuthenticator: m.HasAuthenticator, JoinedAt: m.CreatedAt})
	}
	invs, err := q.SellersOpenInvites(ctx, shopID)
	for _, i := range invs {
		out.Invites = append(out.Invites, TeamInvite{ID: i.ID, Contact: maskContact(i.Phone, i.Email), Role: i.Role, InvitedBy: i.InvitedByFirstName,
			ExpiresAt: i.ExpiresAt, Expired: time.Now().After(i.ExpiresAt), CreatedAt: i.CreatedAt})
	}
	return out, err
}

// InviteMember sends a 7-day, single-use invite link by text or email.
func (s *Service) InviteMember(ctx context.Context, shopID uuid.UUID, in InviteInput) (TeamInvite, error) {
	phone, email, err := in.target()
	if err != nil {
		return TeamInvite{}, err
	}
	actor := httpx.MustPrincipal(ctx).UserID
	token := crypto.RandomToken(32)
	var out TeamInvite
	err = s.inShop(ctx, shopID, teamRoles, func(tx *uow.Tx, shop store.SellersSeller, role string) error {
		if in.Role == "owner" && role != "owner" {
			return httpx.Forbidden("Only an owner can invite another owner.")
		}
		if err := s.notMemberYet(ctx, tx.Q, shopID, phone, email); err != nil {
			return err
		}
		inv, err := tx.Q.SellersCreateInvite(ctx, store.SellersCreateInviteParams{SellerID: shopID, Phone: phone, Email: email, Role: in.Role,
			TokenHash: crypto.SHA256(token), InvitedBy: actor, ExpiresAt: time.Now().Add(inviteTTL)})
		if err != nil {
			if strings.Contains(err.Error(), "seller_invites_one_open") {
				return httpx.Conflict("already_invited", "This person already has an open invite. Re-send it instead.")
			}
			return httpx.DB(err, "invite")
		}
		if err := s.sendInvite(ctx, tx, shop, inv, token, actor); err != nil {
			return err
		}
		out = TeamInvite{ID: inv.ID, Contact: maskContact(phone, email), Role: inv.Role, ExpiresAt: inv.ExpiresAt, CreatedAt: inv.CreatedAt}
		return tx.Audit("seller.member_invited", "seller", shopID.String(), map[string]any{"invite": inv.ID, "role": in.Role})
	})
	if err == nil && s.d.Cfg.IsDevelopment() {
		out.Link = s.inviteLink(token)
	}
	return out, err
}

// notMemberYet refuses an invite for someone already on the team.
func (s *Service) notMemberYet(ctx context.Context, q *store.Queries, shopID uuid.UUID, phone, email *string) error {
	var u store.IdentityUser
	var err error
	if phone != nil {
		u, err = q.IdentityGetUserByPhone(ctx, phone)
	} else {
		u, err = q.IdentityGetUserByEmail(ctx, email)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := q.SellersMemberRole(ctx, store.SellersMemberRoleParams{SellerID: shopID, UserID: u.ID}); err == nil {
		return httpx.Conflict("already_member", "This person is already on the team.")
	}
	return nil
}

func (s *Service) inviteLink(token string) string {
	return strings.TrimRight(s.d.Cfg.SellerAppURL, "/") + "/invite/" + token
}

func (s *Service) sendInvite(ctx context.Context, tx *uow.Tx, shop store.SellersSeller, inv store.SellersSellerInvite, token string, actor uuid.UUID) error {
	inviter, err := tx.Q.IdentityGetUser(ctx, actor)
	if err != nil {
		return err
	}
	m := notify.Msg{Category: notify.CatSeller, Template: "seller_invite", Key: "seller-invite:" + inv.ID.String() + ":" + token[:8],
		Data: map[string]any{"inviter": inviter.FirstName, "shop": shop.DisplayName, "role": inv.Role, "link": s.inviteLink(token)}}
	if inv.Phone != nil {
		m.Channel, m.To = "sms", *inv.Phone
	} else {
		m.Channel, m.To = "email", *inv.Email
	}
	return s.notify.Send(ctx, tx, m)
}

// ResendInvite issues a new link with a fresh 7 days; the old link stops working.
func (s *Service) ResendInvite(ctx context.Context, shopID, inviteID uuid.UUID) (TeamInvite, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	token := crypto.RandomToken(32)
	var out TeamInvite
	err := s.inShop(ctx, shopID, teamRoles, func(tx *uow.Tx, shop store.SellersSeller, role string) error {
		inv, err := tx.Q.SellersGetInvite(ctx, store.SellersGetInviteParams{ID: inviteID, SellerID: shopID})
		if err != nil {
			return httpx.NotFound("Invite")
		}
		if inv.Role == "owner" && role != "owner" {
			return httpx.Forbidden("Only an owner can re-send an owner invite.")
		}
		if inv, err = tx.Q.SellersRenewInvite(ctx, store.SellersRenewInviteParams{ID: inv.ID, TokenHash: crypto.SHA256(token), ExpiresAt: time.Now().Add(inviteTTL)}); err != nil {
			return httpx.Conflict("invite_closed", "This invite was already used or cancelled.")
		}
		if err := s.sendInvite(ctx, tx, shop, inv, token, actor); err != nil {
			return err
		}
		out = TeamInvite{ID: inv.ID, Contact: maskContact(inv.Phone, inv.Email), Role: inv.Role, ExpiresAt: inv.ExpiresAt, CreatedAt: inv.CreatedAt}
		return tx.Audit("seller.invite_resent", "seller", shopID.String(), map[string]any{"invite": inv.ID})
	})
	if err == nil && s.d.Cfg.IsDevelopment() {
		out.Link = s.inviteLink(token)
	}
	return out, err
}

// RevokeInvite cancels an open invite.
func (s *Service) RevokeInvite(ctx context.Context, shopID, inviteID uuid.UUID) error {
	return s.inShop(ctx, shopID, teamRoles, func(tx *uow.Tx, _ store.SellersSeller, role string) error {
		inv, err := tx.Q.SellersGetInvite(ctx, store.SellersGetInviteParams{ID: inviteID, SellerID: shopID})
		if err != nil {
			return httpx.NotFound("Invite")
		}
		if inv.Role == "owner" && role != "owner" {
			return httpx.Forbidden("Only an owner can cancel an owner invite.")
		}
		n, err := tx.Q.SellersRevokeInvite(ctx, store.SellersRevokeInviteParams{ID: inviteID, SellerID: shopID})
		if err != nil || n == 0 {
			return errors.Join(err, httpx.Conflict("invite_closed", "This invite was already used or cancelled."))
		}
		return tx.Audit("seller.invite_revoked", "seller", shopID.String(), map[string]any{"invite": inviteID})
	})
}

// openInvite finds a usable invite by its link token.
func openInvite(ctx context.Context, q *store.Queries, token string) (store.SellersInviteByTokenRow, error) {
	inv, err := q.SellersInviteByToken(ctx, crypto.SHA256(token))
	if err != nil || inv.AcceptedAt != nil || inv.RevokedAt != nil || time.Now().After(inv.ExpiresAt) || inv.ShopStatus == "closed" {
		return inv, httpx.Invalid("invite_invalid", "This invite link is used, cancelled or expired. Ask the shop to send a new one.")
	}
	return inv, nil
}

// matchesInvitee: the invite is accepted from the account it was sent to, so a forwarded or
// leaked link can't add a stranger.
func matchesInvitee(inv store.SellersInviteByTokenRow, u store.IdentityUser) bool {
	if inv.Phone != nil {
		return u.Phone != nil && *u.Phone == *inv.Phone
	}
	return inv.Email != nil && u.Email != nil && strings.EqualFold(*u.Email, *inv.Email)
}

// PreviewInvite shows the shop and role to the signed-in invitee.
func (s *Service) PreviewInvite(ctx context.Context, token string) (InvitePreview, error) {
	inv, err := openInvite(ctx, s.d.Q, token)
	if err != nil {
		return InvitePreview{}, err
	}
	return InvitePreview{Shop: inv.DisplayName, Role: inv.Role, ExpiresAt: inv.ExpiresAt, SentTo: maskContact(inv.Phone, inv.Email)}, nil
}

// AcceptInvite adds the signed-in user to the shop with the invited role.
func (s *Service) AcceptInvite(ctx context.Context, token string) (Shop, error) {
	userID := httpx.MustPrincipal(ctx).UserID
	var shop Shop
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		inv, err := openInvite(ctx, tx.Q, token)
		if err != nil {
			return err
		}
		u, err := tx.Q.IdentityGetUser(ctx, userID)
		if err != nil {
			return err
		}
		if !matchesInvitee(inv, u) {
			return httpx.Forbidden("This invite was sent to " + maskContact(inv.Phone, inv.Email) + ". Sign in with that account to accept it.")
		}
		if inv.Role == "finance" {
			ok, err := tx.Q.SellersHasAuthenticator(ctx, userID)
			if err != nil {
				return err
			}
			if !ok {
				return errAuthenticatorRequired("Finance members must")
			}
		}
		if _, err := tx.Q.SellersMemberRole(ctx, store.SellersMemberRoleParams{SellerID: inv.SellerID, UserID: userID}); err == nil {
			return httpx.Conflict("already_member", "You're already on this shop's team.")
		}
		if n, err := tx.Q.SellersAcceptInvite(ctx, store.SellersAcceptInviteParams{ID: inv.ID, AcceptedBy: &userID}); err != nil || n == 0 {
			return errors.Join(err, httpx.Invalid("invite_invalid", "This invite link is used, cancelled or expired. Ask the shop to send a new one."))
		}
		if err := tx.Q.SellersAddMember(ctx, store.SellersAddMemberParams{SellerID: inv.SellerID, UserID: userID, Role: inv.Role}); err != nil {
			return httpx.DB(err, "member")
		}
		s2, err := tx.Q.SellersGet(ctx, inv.SellerID)
		if err != nil {
			return err
		}
		shop = Shop{ID: s2.ID, Type: s2.Type, DisplayName: s2.DisplayName, Slug: s2.Slug, Status: s2.Status, StateCode: s2.StateCode, City: s2.City,
			Role: inv.Role, CreatedAt: s2.CreatedAt}
		return tx.Audit("seller.member_joined", "seller", inv.SellerID.String(), map[string]any{"invite": inv.ID, "role": inv.Role, "user": userID})
	})
	return shop, err
}

// SetMemberRole changes a member's role. Managers can't touch owners or make anyone an owner,
// the last owner can't be demoted, and finance needs an authenticator.
func (s *Service) SetMemberRole(ctx context.Context, shopID, userID uuid.UUID, newRole string) (ShopTeam, error) {
	var out ShopTeam
	err := s.inShop(ctx, shopID, teamRoles, func(tx *uow.Tx, _ store.SellersSeller, role string) error {
		current, err := tx.Q.SellersMemberRole(ctx, store.SellersMemberRoleParams{SellerID: shopID, UserID: userID})
		if err != nil {
			return httpx.NotFound("Member")
		}
		if role != "owner" && (current == "owner" || newRole == "owner") {
			return httpx.Forbidden("Only an owner can change an owner, or make someone an owner.")
		}
		if current == "owner" && newRole != "owner" {
			if err := lastOwnerGuard(ctx, tx.Q, shopID); err != nil {
				return err
			}
		}
		if newRole == "finance" {
			ok, err := tx.Q.SellersHasAuthenticator(ctx, userID)
			if err != nil {
				return err
			}
			if !ok {
				return errAuthenticatorRequired("They must")
			}
		}
		if _, err := tx.Q.SellersSetMemberRole(ctx, store.SellersSetMemberRoleParams{SellerID: shopID, UserID: userID, Role: newRole}); err != nil {
			return err
		}
		if err := tx.Audit("seller.member_role_changed", "seller", shopID.String(), map[string]any{"user": userID, "from": current, "to": newRole}); err != nil {
			return err
		}
		out, err = s.team(ctx, tx.Q, shopID)
		return err
	})
	return out, err
}

// RemoveMember takes someone off the team (members may also leave themselves, unless they're
// the last owner).
func (s *Service) RemoveMember(ctx context.Context, shopID, userID uuid.UUID) error {
	self := httpx.MustPrincipal(ctx).UserID == userID
	roles := teamRoles
	if self {
		roles = []string{"owner", "manager", "staff", "finance"}
	}
	return s.inShop(ctx, shopID, roles, func(tx *uow.Tx, _ store.SellersSeller, role string) error {
		current, err := tx.Q.SellersMemberRole(ctx, store.SellersMemberRoleParams{SellerID: shopID, UserID: userID})
		if err != nil {
			return httpx.NotFound("Member")
		}
		if current == "owner" && role != "owner" {
			return httpx.Forbidden("A manager can't remove an owner.")
		}
		if current == "owner" {
			if err := lastOwnerGuard(ctx, tx.Q, shopID); err != nil {
				return err
			}
		}
		if _, err := tx.Q.SellersRemoveMember(ctx, store.SellersRemoveMemberParams{SellerID: shopID, UserID: userID}); err != nil {
			return err
		}
		return tx.Audit("seller.member_removed", "seller", shopID.String(), map[string]any{"user": userID, "role": current, "self": self})
	})
}

func lastOwnerGuard(ctx context.Context, q *store.Queries, shopID uuid.UUID) error {
	n, err := q.SellersOwnerCount(ctx, shopID)
	if err != nil {
		return err
	}
	if n <= 1 {
		return httpx.Conflict("last_owner", "A shop needs at least one owner. Make someone else an owner first.")
	}
	return nil
}
