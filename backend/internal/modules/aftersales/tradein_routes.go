package aftersales

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/modules/payments"
	"github.com/dprince-03/techshop/backend/internal/platform/auth"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/jobs"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
)

type tradeInOut struct{ Body TradeIn }

func (s *Service) tradeInRoutes(r *kit.Routes) {
	g := r.Public

	httpx.Register(g, httpx.Op{ID: "tradeInDevices", Method: http.MethodGet, Path: "/api/v1/trade-ins/devices", Tag: "Trade-ins",
		Summary: "Models we can price instantly, with an estimate per condition (final value after inspection)"},
		func(ctx context.Context, in *struct {
			Type string `query:"type" enum:"phone,laptop,tablet,console,smartwatch" required:"true"`
		}) (*struct{ Body []TradeInDevice }, error) {
			slugs := tradeInCategories[in.Type]
			out := []TradeInDevice{}
			if len(slugs) == 0 {
				return &struct{ Body []TradeInDevice }{out}, nil
			}
			rows, err := s.d.Q.TradeInDevices(ctx, slugs)
			for _, d := range rows {
				est := map[string]int64{}
				for c := range tradeInFactor {
					est[c] = tradeInEstimate(d.NewPriceKobo, c)
				}
				out = append(out, TradeInDevice{VariantID: d.VariantID, Brand: d.Brand, ProductName: d.ProductName, VariantName: d.VariantName, Estimates: est})
			}
			return &struct{ Body []TradeInDevice }{out}, err
		})

	me := func(o httpx.Op) httpx.Op { o.Tag = "Trade-ins"; o.Auth = shoppers; return o }
	httpx.Register(g, me(httpx.Op{ID: "tradeInBanks", Method: http.MethodGet, Path: "/api/v1/me/banks", Summary: "Banks a trade-in can be paid to"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []payments.Bank }, error) {
			banks, err := s.pay.Banks(ctx)
			if err != nil {
				s.d.Logger.Warn("bank list failed", "error", err)
				return nil, httpx.E(503, "bank_unavailable", "The bank list isn't available right now. Try again shortly.")
			}
			return &struct{ Body []payments.Bank }{banks}, nil
		})
	httpx.Register(g, me(httpx.Op{ID: "startTradeIn", Method: http.MethodPost, Path: "/api/v1/me/trade-ins", Status: 201, Summary: "Get an estimate, then bring the device to a store"}),
		func(ctx context.Context, in *struct{ Body TradeInInput }) (*tradeInOut, error) {
			out, err := s.CreateTradeIn(ctx, httpx.MustPrincipal(ctx).UserID, in.Body)
			return &tradeInOut{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "myTradeIns", Method: http.MethodGet, Path: "/api/v1/me/trade-ins", Summary: "Your trade-ins, newest first"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []TradeInRow }, error) {
			rows, err := s.d.Q.TradeInsForUser(ctx, httpx.MustPrincipal(ctx).UserID)
			out := make([]TradeInRow, len(rows))
			for i, t := range rows {
				out[i] = tradeInRow(t, "", "")
			}
			return &struct{ Body []TradeInRow }{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "myTradeIn", Method: http.MethodGet, Path: "/api/v1/me/trade-ins/{ref}", Summary: "One of your trade-ins"}),
		func(ctx context.Context, in *struct {
			Ref string `path:"ref"`
		}) (*tradeInOut, error) {
			out, err := s.customerTradeIn(ctx, httpx.MustPrincipal(ctx).UserID, in.Ref)
			return &tradeInOut{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "cancelTradeIn", Method: http.MethodPost, Path: "/api/v1/me/trade-ins/{ref}/cancel", Summary: "Withdraw before handing the device in"}),
		func(ctx context.Context, in *struct {
			Ref string `path:"ref"`
		}) (*tradeInOut, error) {
			out, err := s.CancelTradeIn(ctx, httpx.MustPrincipal(ctx).UserID, in.Ref)
			return &tradeInOut{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "acceptTradeIn", Method: http.MethodPost, Path: "/api/v1/me/trade-ins/{ref}/accept", Summary: "Take the offer: store credit at once, or a bank transfer"}),
		func(ctx context.Context, in *struct {
			Ref  string `path:"ref"`
			Body AcceptInput
		}) (*tradeInOut, error) {
			out, err := s.AcceptTradeIn(ctx, httpx.MustPrincipal(ctx).UserID, in.Ref, in.Body)
			return &tradeInOut{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "declineTradeIn", Method: http.MethodPost, Path: "/api/v1/me/trade-ins/{ref}/decline", Summary: "Turn the offer down and collect the device"}),
		func(ctx context.Context, in *struct {
			Ref string `path:"ref"`
		}) (*tradeInOut, error) {
			out, err := s.DeclineTradeIn(ctx, httpx.MustPrincipal(ctx).UserID, in.Ref)
			return &tradeInOut{out}, err
		})

	st := func(perm string, o httpx.Op) httpx.Op {
		o.Tag = "Staff · Trade-ins"
		o.Auth = []string{auth.AudStaff}
		o.Perm = perm
		return o
	}
	httpx.Register(g, st("aftersales.manage", httpx.Op{ID: "staffTradeInSummary", Method: http.MethodGet, Path: "/api/v1/staff/trade-ins/summary", Summary: "Trade-in counts"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body TradeInSummary }, error) {
			r, err := s.d.Q.TradeInSummary(ctx)
			return &struct{ Body TradeInSummary }{TradeInSummary{NeedsEstimate: r.NeedsEstimate, AwaitingInspection: r.AwaitingInspection, OffersOut: r.OffersOut,
				PayoutsDue: r.PayoutsDue, PaidThisMonth: r.PaidThisMonth, AvgPaid: r.AvgPaidKobo}}, err
		})
	httpx.Register(g, st("aftersales.manage", httpx.Op{ID: "staffTradeIns", Method: http.MethodGet, Path: "/api/v1/staff/trade-ins", Summary: "Trade-ins: work waiting first"}),
		func(ctx context.Context, in *struct {
			Status string `query:"status" enum:"active,payout_due,submitted,estimated,awaiting_inspection,offer_sent,accepted,declined,paid,cancelled,blocked,"`
			Q      string `query:"q" maxLength:"40" doc:"Reference or IMEI"`
		}) (*struct{ Body []TradeInRow }, error) {
			rows, err := s.d.Q.TradeInStaffList(ctx, store.TradeInStaffListParams{Status: optional(in.Status), Q: optional(in.Q)})
			out := make([]TradeInRow, len(rows))
			for i, r := range rows {
				out[i] = tradeInRow(store.AftersalesTradeIn{Reference: r.Reference, DeviceType: r.DeviceType, Brand: r.Brand, Model: r.Model, Storage: r.Storage,
					DeclaredCondition: r.DeclaredCondition, EstimateKobo: r.EstimateKobo, OfferKobo: r.OfferKobo, Status: r.Status, PayoutMethod: r.PayoutMethod,
					OfferExpiresAt: r.OfferExpiresAt, CreatedAt: r.CreatedAt}, r.FirstName, r.LastName)
			}
			return &struct{ Body []TradeInRow }{out}, err
		})
	httpx.Register(g, st("aftersales.manage", httpx.Op{ID: "staffTradeIn", Method: http.MethodGet, Path: "/api/v1/staff/trade-ins/{ref}", Summary: "One trade-in"}),
		func(ctx context.Context, in *struct {
			Ref string `path:"ref"`
		}) (*tradeInOut, error) {
			out, err := s.tradeInView(ctx, in.Ref)
			return &tradeInOut{out}, err
		})
	httpx.Register(g, st("aftersales.manage", httpx.Op{ID: "staffEstimateTradeIn", Method: http.MethodPost, Path: "/api/v1/staff/trade-ins/{ref}/estimate", Summary: "Price a device we couldn't price automatically"}),
		func(ctx context.Context, in *struct {
			Ref  string `path:"ref"`
			Body struct {
				Estimate  int64      `json:"estimate" minimum:"0" doc:"Kobo"`
				VariantID *uuid.UUID `json:"variantId,omitempty"`
			}
		}) (*tradeInOut, error) {
			out, err := s.EstimateTradeIn(ctx, in.Ref, in.Body.Estimate, in.Body.VariantID)
			return &tradeInOut{out}, err
		})
	httpx.Register(g, st("aftersales.manage", httpx.Op{ID: "staffReceiveTradeIn", Method: http.MethodPost, Path: "/api/v1/staff/trade-ins/{ref}/receive", Summary: "Take the device at a store (IMEI and risk check)"}),
		func(ctx context.Context, in *struct {
			Ref  string `path:"ref"`
			Body ReceiveInput
		}) (*tradeInOut, error) {
			out, err := s.ReceiveTradeIn(ctx, in.Ref, in.Body)
			return &tradeInOut{out}, err
		})
	httpx.Register(g, st("aftersales.manage", httpx.Op{ID: "staffInspectTradeIn", Method: http.MethodPost, Path: "/api/v1/staff/trade-ins/{ref}/inspection", Summary: "Grade the device and send the offer (valid 7 days)"}),
		func(ctx context.Context, in *struct {
			Ref  string `path:"ref"`
			Body InspectTradeInInput
		}) (*tradeInOut, error) {
			out, err := s.InspectTradeIn(ctx, in.Ref, in.Body)
			return &tradeInOut{out}, err
		})
	httpx.Register(g, st("refunds.approve", httpx.Op{ID: "staffPayTradeIn", Method: http.MethodPost, Path: "/api/v1/staff/trade-ins/{ref}/paid", StepUp: true,
		Summary: "Record the bank transfer for an accepted trade-in (finance; not whoever priced it)"}),
		func(ctx context.Context, in *struct {
			Ref  string `path:"ref"`
			Body struct {
				Reference string `json:"reference" maxLength:"80" doc:"The bank transfer reference"`
			}
		}) (*tradeInOut, error) {
			out, err := s.MarkTradeInPaid(ctx, in.Ref, in.Body.Reference)
			return &tradeInOut{out}, err
		})
}

type expireOffersArgs struct{}

// Kind identifies the job.
func (expireOffersArgs) Kind() string { return "aftersales.expire_trade_in_offers" }

type expireOffersWorker struct {
	river.WorkerDefaults[expireOffersArgs]
	s *Service
}

func (w *expireOffersWorker) Work(ctx context.Context, _ *river.Job[expireOffersArgs]) error {
	return w.s.ExpireTradeInOffers(ctx)
}

// Jobs implements kit.Module: trade-in offers lapse after 7 days (checked hourly).
func (s *Service) Jobs(reg *jobs.Registry) {
	river.AddWorker(reg.Workers, &expireOffersWorker{s: s})
	reg.Every(time.Hour, func() (river.JobArgs, *river.InsertOpts) { return expireOffersArgs{}, nil })
}
