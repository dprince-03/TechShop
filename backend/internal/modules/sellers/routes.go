package sellers

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/modules/payments"
	"github.com/dprince-03/techshop/backend/internal/platform/auth"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
)

var sellerAud = []string{auth.AudSeller}

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	s.listingRoutes(r)
	s.orderRoutes(r)
	s.reviewRoutes(r)
	g := r.Public
	sel := func(o httpx.Op) httpx.Op { o.Tag = "Seller Centre"; o.Auth = sellerAud; return o }

	httpx.Register(g, sel(httpx.Op{ID: "sellerRegisterShop", Method: http.MethodPost, Path: "/api/v1/seller/shops", Status: 201, Summary: "Open a shop (waits for verification)"}),
		func(ctx context.Context, in *struct{ Body RegisterInput }) (*struct{ Body Shop }, error) {
			shop, err := s.RegisterShop(ctx, httpx.MustPrincipal(ctx).UserID, in.Body)
			return &struct{ Body Shop }{shop}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerMyShops", Method: http.MethodGet, Path: "/api/v1/seller/shops", Summary: "Shops you belong to, with your role"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []Shop }, error) {
			shops, err := s.MyShops(ctx, httpx.MustPrincipal(ctx).UserID)
			return &struct{ Body []Shop }{shops}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerOnboarding", Method: http.MethodGet, Path: "/api/v1/seller/shops/{shopId}/onboarding", Summary: "Verification and payout setup, with the next step"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
		}) (*struct{ Body Onboarding }, error) {
			o, err := s.Onboarding(ctx, in.ShopID)
			return &struct{ Body Onboarding }{o}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerSubmitKyc", Method: http.MethodPost, Path: "/api/v1/seller/shops/{shopId}/kyc", Status: 201, Summary: "Submit identity and business verification (owner)"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			Body   KycInput
		}) (*struct{ Body Kyc }, error) {
			k, err := s.SubmitKyc(ctx, in.ShopID, in.Body)
			return &struct{ Body Kyc }{k}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerBanks", Method: http.MethodGet, Path: "/api/v1/seller/banks", Summary: "Banks payouts can go to"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []payments.Bank }, error) {
			banks, err := s.banks.Banks(ctx)
			if err != nil {
				s.d.Logger.Warn("bank list failed", "error", err)
				return nil, httpx.E(503, "bank_unavailable", "The bank list isn't available right now. Try again shortly.")
			}
			return &struct{ Body []payments.Bank }{banks}, nil
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerAddBankAccount", Method: http.MethodPost, Path: "/api/v1/seller/shops/{shopId}/bank-accounts", Status: 201, Summary: "Add a payout account (checked with the bank; owner or finance)"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			Body   BankInput
		}) (*struct{ Body BankAccount }, error) {
			b, err := s.AddBankAccount(ctx, in.ShopID, in.Body)
			return &struct{ Body BankAccount }{b}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerSetDefaultBank", Method: http.MethodPost, Path: "/api/v1/seller/shops/{shopId}/bank-accounts/{accountId}/default", Status: 204, Summary: "Pay out to this account"}),
		func(ctx context.Context, in *struct {
			ShopID    uuid.UUID `path:"shopId"`
			AccountID uuid.UUID `path:"accountId"`
		}) (*struct{}, error) {
			return nil, s.SetDefaultBank(ctx, in.ShopID, in.AccountID)
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerDeleteBankAccount", Method: http.MethodDelete, Path: "/api/v1/seller/shops/{shopId}/bank-accounts/{accountId}", Status: 204, Summary: "Remove an account that isn't the default"}),
		func(ctx context.Context, in *struct {
			ShopID    uuid.UUID `path:"shopId"`
			AccountID uuid.UUID `path:"accountId"`
		}) (*struct{}, error) {
			return nil, s.DeleteBank(ctx, in.ShopID, in.AccountID)
		})

	st := func(o httpx.Op) httpx.Op { o.Tag = "Staff · Vendors"; o.Auth = []string{auth.AudStaff}; return o }
	httpx.Register(g, st(httpx.Op{ID: "staffListVendors", Method: http.MethodGet, Path: "/api/v1/staff/vendors", Perm: "vendors.view", Summary: "Marketplace sellers, newest first"}),
		func(ctx context.Context, in *struct {
			Status    string `query:"status" enum:"pending_kyc,active,suspended,closed,"`
			KycStatus string `query:"kycStatus" enum:"submitted,in_review,approved,rejected,needs_more_info,"`
			Q         string `query:"q" maxLength:"60"`
			Cursor    string `query:"cursor"`
			Limit     int    `query:"limit"`
		}) (*struct{ Body httpx.List[VendorRow] }, error) {
			before, err := s.d.Cursors.Before(in.Cursor)
			if err != nil {
				return nil, httpx.BadCursor()
			}
			lim := httpx.Limit(in.Limit)
			rows, err := s.d.Q.SellersStaffList(ctx, store.SellersStaffListParams{Limit: int32(lim + 1), Status: optional(in.Status),
				KycStatus: optional(in.KycStatus), Q: optional(in.Q), Before: before})
			if err != nil {
				return nil, err
			}
			views := make([]VendorRow, len(rows))
			for i, r := range rows {
				views[i] = vendorRow(r)
			}
			return &struct{ Body httpx.List[VendorRow] }{httpx.Paginate(s.d.Cursors, views, lim, func(v VendorRow) time.Time { return v.CreatedAt })}, nil
		})
	httpx.Register(g, st(httpx.Op{ID: "staffGetVendor", Method: http.MethodGet, Path: "/api/v1/staff/vendors/{id}", Perm: "vendors.view", Summary: "A seller with KYC, bank accounts and duplicate signals"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body Vendor }, error) {
			v, err := s.Vendor(ctx, in.ID)
			return &struct{ Body Vendor }{v}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffReviewKyc", Method: http.MethodPost, Path: "/api/v1/staff/vendors/kyc/{id}/review", Perm: "vendors.kyc_review", Summary: "Approve, reject or ask for more information"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body KycDecision
		}) (*struct{ Body Kyc }, error) {
			k, err := s.ReviewKyc(ctx, in.ID, in.Body)
			return &struct{ Body Kyc }{k}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffRevealKycId", Method: http.MethodPost, Path: "/api/v1/staff/vendors/kyc/{id}/reveal", Perm: "vendors.reveal_id", Summary: "Show the full ID number (reason required; audited)"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Reason string `json:"reason" minLength:"10" maxLength:"300"`
			}
		}) (*struct {
			Body struct {
				IDNumber string `json:"idNumber"`
			}
		}, error) {
			n, err := s.RevealID(ctx, in.ID, in.Body.Reason)
			out := &struct {
				Body struct {
					IDNumber string `json:"idNumber"`
				}
			}{}
			out.Body.IDNumber = n
			return out, err
		})
}
