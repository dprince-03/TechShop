package sellers

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
)

func (s *Service) teamRoutes(r *kit.Routes) {
	g := r.Public
	sel := func(o httpx.Op) httpx.Op { o.Tag = "Seller Centre · Team"; o.Auth = sellerAud; return o }
	base := "/api/v1/seller/shops/{shopId}/team"

	httpx.Register(g, sel(httpx.Op{ID: "sellerTeam", Method: http.MethodGet, Path: base, Summary: "Members and open invites (owners and managers)"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
		}) (*struct{ Body ShopTeam }, error) {
			t, err := s.Team(ctx, in.ShopID)
			return &struct{ Body ShopTeam }{t}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerInvite", Method: http.MethodPost, Path: base + "/invites", Status: 201,
		Summary: "Invite someone by phone or email (link valid 7 days)"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			Body   InviteInput
		}) (*struct{ Body TeamInvite }, error) {
			i, err := s.InviteMember(ctx, in.ShopID, in.Body)
			return &struct{ Body TeamInvite }{i}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerResendInvite", Method: http.MethodPost, Path: base + "/invites/{inviteId}/resend",
		Summary: "Send a new link (the old one stops working; 7 more days)"}),
		func(ctx context.Context, in *struct {
			ShopID   uuid.UUID `path:"shopId"`
			InviteID uuid.UUID `path:"inviteId"`
		}) (*struct{ Body TeamInvite }, error) {
			i, err := s.ResendInvite(ctx, in.ShopID, in.InviteID)
			return &struct{ Body TeamInvite }{i}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerRevokeInvite", Method: http.MethodDelete, Path: base + "/invites/{inviteId}", Status: 204, Summary: "Cancel an invite"}),
		func(ctx context.Context, in *struct {
			ShopID   uuid.UUID `path:"shopId"`
			InviteID uuid.UUID `path:"inviteId"`
		}) (*struct{}, error) {
			return nil, s.RevokeInvite(ctx, in.ShopID, in.InviteID)
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerSetMemberRole", Method: http.MethodPut, Path: base + "/members/{userId}/role", Summary: "Change a member's role"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			UserID uuid.UUID `path:"userId"`
			Body   struct {
				Role string `json:"role" enum:"owner,manager,staff,finance"`
			}
		}) (*struct{ Body ShopTeam }, error) {
			t, err := s.SetMemberRole(ctx, in.ShopID, in.UserID, in.Body.Role)
			return &struct{ Body ShopTeam }{t}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerRemoveMember", Method: http.MethodDelete, Path: base + "/members/{userId}", Status: 204,
		Summary: "Remove a member (or leave the shop yourself)"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			UserID uuid.UUID `path:"userId"`
		}) (*struct{}, error) {
			return nil, s.RemoveMember(ctx, in.ShopID, in.UserID)
		})

	httpx.Register(g, sel(httpx.Op{ID: "sellerPreviewInvite", Method: http.MethodPost, Path: "/api/v1/seller/invites/preview", Summary: "The shop and role an invite is for"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Token string `json:"token" minLength:"20" maxLength:"100"`
			}
		}) (*struct{ Body InvitePreview }, error) {
			p, err := s.PreviewInvite(ctx, in.Body.Token)
			return &struct{ Body InvitePreview }{p}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerAcceptInvite", Method: http.MethodPost, Path: "/api/v1/seller/invites/accept",
		Summary: "Join the shop (from the account the invite was sent to)"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Token string `json:"token" minLength:"20" maxLength:"100"`
			}
		}) (*struct{ Body Shop }, error) {
			sh, err := s.AcceptInvite(ctx, in.Body.Token)
			return &struct{ Body Shop }{sh}, err
		})
}
