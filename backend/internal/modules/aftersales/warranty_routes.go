package aftersales

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/platform/auth"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
)

type claimOut struct{ Body Claim }

func (s *Service) warrantyRoutes(r *kit.Routes) {
	g := r.Public
	me := func(o httpx.Op) httpx.Op { o.Tag = "Warranty"; o.Auth = shoppers; return o }

	httpx.Register(g, me(httpx.Op{ID: "myWarrantyItems", Method: http.MethodGet, Path: "/api/v1/me/warranty/items", Summary: "Your delivered items still under warranty"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []CoveredItem }, error) {
			uid := httpx.MustPrincipal(ctx).UserID
			rows, err := s.d.Q.WarrantyCoveredLines(ctx, &uid)
			out := make([]CoveredItem, len(rows))
			for i, l := range rows {
				prov := ""
				if l.WarrantyProvider != nil {
					prov = *l.WarrantyProvider
				}
				var months int16
				if l.WarrantyMonths != nil {
					months = *l.WarrantyMonths
				}
				var open *string
				if l.OpenClaim != "" {
					open = &l.OpenClaim
				}
				out[i] = CoveredItem{OrderLineID: l.ID, OrderNumber: l.OrderNumber, ProductName: l.ProductName, VariantName: l.VariantName, Provider: prov, Months: months,
					DeliveredAt: *l.DeliveredAt, EndsOn: l.EndsOn, Serialised: l.Imei != nil || l.SerialNumber != nil, SerialHint: last4(l.Imei, l.SerialNumber), OpenClaim: open}
			}
			return &struct{ Body []CoveredItem }{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "makeWarrantyClaim", Method: http.MethodPost, Path: "/api/v1/me/warranty-claims", Status: 201, Summary: "Claim under warranty, then bring the device to a store"}),
		func(ctx context.Context, in *struct{ Body ClaimInput }) (*claimOut, error) {
			out, err := s.CreateClaim(ctx, httpx.MustPrincipal(ctx).UserID, in.Body)
			return &claimOut{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "myWarrantyClaims", Method: http.MethodGet, Path: "/api/v1/me/warranty-claims", Summary: "Your warranty claims, newest first"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []ClaimRow }, error) {
			rows, err := s.d.Q.WarrantyClaimsForUser(ctx, httpx.MustPrincipal(ctx).UserID)
			out := make([]ClaimRow, len(rows))
			for i, c := range rows {
				out[i] = claimRow(store.AftersalesWarrantyClaim{ClaimNumber: c.ClaimNumber, ImeiOrSerial: c.ImeiOrSerial, FaultDescription: c.FaultDescription, Status: c.Status,
					Resolution: c.Resolution, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}, c.ProductName, c.VariantName, c.OrderNumber, nil, "", "")
			}
			return &struct{ Body []ClaimRow }{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "myWarrantyClaim", Method: http.MethodGet, Path: "/api/v1/me/warranty-claims/{number}", Summary: "One of your warranty claims"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
		}) (*claimOut, error) {
			out, err := s.customerClaim(ctx, httpx.MustPrincipal(ctx).UserID, in.Number)
			return &claimOut{out}, err
		})

	st := func(o httpx.Op) httpx.Op {
		o.Tag = "Staff · Warranty"
		o.Auth = []string{auth.AudStaff}
		o.Perm = "aftersales.manage"
		return o
	}
	httpx.Register(g, st(httpx.Op{ID: "staffWarrantySummary", Method: http.MethodGet, Path: "/api/v1/staff/warranty/summary", Summary: "Claim counts and turnaround"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body WarrantySummary }, error) {
			r, err := s.d.Q.WarrantySummary(ctx)
			return &struct{ Body WarrantySummary }{WarrantySummary{Open: r.Open, AwaitingDropoff: r.AwaitingDropoff, InRepair: r.InRepair, Ready: r.Ready,
				ReplacedThisMonth: r.ReplacedThisMonth, AvgTurnaroundDays: r.AvgTurnaroundDays}}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffWarrantyStores", Method: http.MethodGet, Path: "/api/v1/staff/warranty/stores", Summary: "Stores and warehouses that take devices in"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []Store }, error) {
			rows, err := s.d.Q.WarrantyStores(ctx)
			out := make([]Store, len(rows))
			for i, st := range rows {
				out[i] = Store{ID: st.ID, Name: st.Name, Address: st.Address, City: st.City, State: st.StateCode}
			}
			return &struct{ Body []Store }{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffWarrantyClaims", Method: http.MethodGet, Path: "/api/v1/staff/warranty/claims", Summary: "Warranty claims: open ones first, oldest first"}),
		func(ctx context.Context, in *struct {
			Status string `query:"status" enum:"active,new,received,in_repair,ready,collected,rejected,closed,"`
			Q      string `query:"q" maxLength:"40" doc:"Claim or order number, or an exact IMEI or serial"`
		}) (*struct{ Body []ClaimRow }, error) {
			rows, err := s.d.Q.WarrantyStaffList(ctx, store.WarrantyStaffListParams{Status: optional(in.Status), Q: optional(in.Q)})
			out := make([]ClaimRow, len(rows))
			for i, c := range rows {
				out[i] = claimRow(store.AftersalesWarrantyClaim{ClaimNumber: c.ClaimNumber, ImeiOrSerial: c.ImeiOrSerial, FaultDescription: c.FaultDescription, Status: c.Status,
					Resolution: c.Resolution, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}, c.ProductName, c.VariantName, c.OrderNumber, c.StoreName, c.CustomerFirstName, c.CustomerLastName)
			}
			return &struct{ Body []ClaimRow }{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffWarrantyClaim", Method: http.MethodGet, Path: "/api/v1/staff/warranty/claims/{number}", Summary: "A claim with its repairs and the device's history"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
		}) (*claimOut, error) {
			out, err := s.claimView(ctx, in.Number)
			return &claimOut{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffWalkInClaim", Method: http.MethodPost, Path: "/api/v1/staff/warranty/claims", Status: 201, Summary: "Take a claim at the counter (checked in at once)"}),
		func(ctx context.Context, in *struct{ Body WalkInInput }) (*claimOut, error) {
			out, err := s.WalkIn(ctx, in.Body)
			return &claimOut{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffReceiveClaim", Method: http.MethodPost, Path: "/api/v1/staff/warranty/claims/{number}/receive", Summary: "Check the device in (IMEI must match; stolen devices are refused)"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
			Body   struct {
				StoreID       uuid.UUID `json:"storeId"`
				ScannedSerial string    `json:"scannedSerial" minLength:"3" maxLength:"40"`
			}
		}) (*claimOut, error) {
			out, err := s.ReceiveClaim(ctx, in.Number, in.Body.StoreID, in.Body.ScannedSerial)
			return &claimOut{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffRejectClaim", Method: http.MethodPost, Path: "/api/v1/staff/warranty/claims/{number}/reject", Summary: "Not covered (the customer sees the reason)"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
			Body   struct {
				Reason string `json:"reason" maxLength:"500"`
			}
		}) (*claimOut, error) {
			out, err := s.RejectClaim(ctx, in.Number, in.Body.Reason)
			return &claimOut{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffStartRepair", Method: http.MethodPost, Path: "/api/v1/staff/warranty/claims/{number}/repairs", Summary: "Send the device to the workshop"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
			Body   RepairInput
		}) (*claimOut, error) {
			out, err := s.StartRepair(ctx, in.Number, in.Body)
			return &claimOut{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffRepairQueue", Method: http.MethodGet, Path: "/api/v1/staff/warranty/repairs", Summary: "The workshop: warranty repairs and refurbishments"}),
		func(ctx context.Context, in *struct {
			Status  string `query:"status" enum:"active,queued,diagnosing,awaiting_parts,repairing,done,unrepairable,"`
			Purpose string `query:"purpose" enum:"warranty,refurbish,"`
		}) (*struct{ Body []RepairView }, error) {
			rows, err := s.d.Q.WarrantyRepairQueue(ctx, store.WarrantyRepairQueueParams{Status: optional(in.Status), Purpose: optional(in.Purpose)})
			out := make([]RepairView, len(rows))
			for i, r := range rows {
				out[i] = repairViewOf(r)
			}
			return &struct{ Body []RepairView }{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffUpdateRepair", Method: http.MethodPost, Path: "/api/v1/staff/warranty/repairs/{id}",
		Summary: "Record workshop progress (done: a claim is ready to collect, a refurbished device goes into stock)"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body RepairUpdate
		}) (*struct{ Body RepairView }, error) {
			out, err := s.UpdateRepair(ctx, in.ID, in.Body)
			return &struct{ Body RepairView }{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffResolveClaim", Method: http.MethodPost, Path: "/api/v1/staff/warranty/claims/{number}/resolution", Summary: "Replace from stock, or refund one unit (finance approves)"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
			Body   ResolveInput
		}) (*claimOut, error) {
			out, err := s.ResolveClaim(ctx, in.Number, in.Body)
			return &claimOut{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffCollectClaim", Method: http.MethodPost, Path: "/api/v1/staff/warranty/claims/{number}/collect", Summary: "The customer collected the device"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
		}) (*claimOut, error) {
			out, err := s.CollectClaim(ctx, in.Number)
			return &claimOut{out}, err
		})
}
