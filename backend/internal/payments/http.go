package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/idempotency"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/ledger"
	"github.com/dprince-03/techshop/backend/internal/sales"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

var buyers = []string{auth.AudMarket, auth.AudCustomerApp, auth.AudWholesale}

// PaymentStatus is what the checkout return page polls.
type PaymentStatus struct {
	PaymentID   uuid.UUID `json:"paymentId"`
	Status      string    `json:"status"`
	OrderNumber string    `json:"orderNumber"`
	OrderStatus string    `json:"orderStatus"`
}

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	me := func(o httpx.Op) httpx.Op { o.Tag = "Payments"; o.Auth = buyers; return o }
	staff := func(o httpx.Op) httpx.Op { o.Tag = "Staff · Finance"; o.Auth = []string{auth.AudStaff}; return o }

	// Provider webhooks: signature over the raw body first, then stored once, processed in a job.
	httpx.Register(g, httpx.Op{ID: "paymentWebhook", Method: http.MethodPost, Path: "/api/v1/webhooks/{provider}", Tag: "Webhooks", MaxBody: 256 << 10,
		Summary: "Payment provider webhook (signed)"},
		func(ctx context.Context, in *struct {
			Provider    string `path:"provider" enum:"paystack,opay,moniepoint"`
			PaystackSig string `header:"x-paystack-signature"`
			MonnifySig  string `header:"monnify-signature"`
			OPaySig     string `header:"x-opay-signature"`
			RawBody     []byte
		}) (*struct {
			Body struct {
				Status string `json:"status"`
			}
		}, error) {
			sig := map[string]string{"paystack": in.PaystackSig, "moniepoint": in.MonnifySig, "opay": in.OPaySig}[in.Provider]
			dup, err := s.ReceiveWebhook(ctx, in.Provider, in.RawBody, sig)
			if err != nil {
				return nil, err
			}
			out := &struct {
				Body struct {
					Status string `json:"status"`
				}
			}{}
			out.Body.Status = "ok"
			if dup {
				out.Body.Status = "duplicate"
			}
			return out, nil
		})

	// ---- Customer
	httpx.Register(g, me(httpx.Op{ID: "retryPayment", Method: http.MethodPost, Path: "/api/v1/me/orders/{number}/payments", Status: 201,
		Summary: "Pay again for an unpaid order (e.g. another method)"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
			Key    string `header:"Idempotency-Key" required:"true"`
			Body   struct {
				Provider string `json:"provider" enum:"paystack,opay,moniepoint"`
			}
		}) (*struct{ Body *sales.PaymentAction }, error) {
			p := httpx.MustPrincipal(ctx)
			out, _, err := idempotency.Do(ctx, s.d.Idem, "payments.retry", in.Key, &p.UserID, http.MethodPost, "/api/v1/me/orders/"+in.Number+"/payments", in.Body,
				func() (*sales.PaymentAction, error) {
					a, err := s.Retry(ctx, p.UserID, in.Number, in.Body.Provider)
					return &a, err
				})
			return &struct{ Body *sales.PaymentAction }{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "getMyPayment", Method: http.MethodGet, Path: "/api/v1/me/payments/{id}", Summary: "Payment and order status (poll after checkout)"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body PaymentStatus }, error) {
			p := httpx.MustPrincipal(ctx)
			pay, err := s.d.Q.PaymentsGet(ctx, in.ID)
			if err != nil || pay.OrderID == nil {
				return nil, httpx.NotFound("Payment")
			}
			o, err := s.d.Q.SalesGetOrder(ctx, *pay.OrderID)
			if err != nil || o.CustomerUserID == nil || *o.CustomerUserID != p.UserID {
				return nil, httpx.NotFound("Payment")
			}
			return &struct{ Body PaymentStatus }{PaymentStatus{pay.ID, pay.Status, o.OrderNumber, o.Status}}, nil
		})
	httpx.Register(g, me(httpx.Op{ID: "listMyRefunds", Method: http.MethodGet, Path: "/api/v1/me/refunds", Summary: "My refunds"}),
		func(ctx context.Context, _ *struct{}) (*struct {
			Body []store.PaymentsCustomerRefundsRow
		}, error) {
			rows, err := s.d.Q.PaymentsCustomerRefunds(ctx, &httpx.MustPrincipal(ctx).UserID)
			if rows == nil {
				rows = []store.PaymentsCustomerRefundsRow{}
			}
			return &struct {
				Body []store.PaymentsCustomerRefundsRow
			}{rows}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "getMyStoreCredit", Method: http.MethodGet, Path: "/api/v1/me/store-credit", Summary: "Store credit balance"}),
		func(ctx context.Context, _ *struct{}) (*struct {
			Body struct {
				BalanceKobo int64 `json:"balanceKobo"`
			}
		}, error) {
			bal, err := s.d.Q.FinanceAccountBalance(ctx, ledger.StoreCredit(httpx.MustPrincipal(ctx).UserID))
			out := &struct {
				Body struct {
					BalanceKobo int64 `json:"balanceKobo"`
				}
			}{}
			out.Body.BalanceKobo = -bal // a liability: credits are negative
			return out, err
		})

	// ---- Staff: refunds (dual control)
	httpx.Register(g, staff(httpx.Op{ID: "staffListRefunds", Method: http.MethodGet, Path: "/api/v1/staff/finance/refunds", Perm: "finance.view", Summary: "Refunds (filter by status)"}),
		func(ctx context.Context, in *struct {
			Status string `query:"status" enum:"requested,approved,processing,succeeded,failed,rejected,"`
		}) (*struct {
			Body []store.PaymentsListRefundsRow
		}, error) {
			p := store.PaymentsListRefundsParams{Limit: 200}
			if in.Status != "" {
				p.Status = &in.Status
			}
			rows, err := s.d.Q.PaymentsListRefunds(ctx, p)
			if rows == nil {
				rows = []store.PaymentsListRefundsRow{}
			}
			return &struct {
				Body []store.PaymentsListRefundsRow
			}{rows}, err
		})
	httpx.Register(g, staff(httpx.Op{ID: "staffRequestRefund", Method: http.MethodPost, Path: "/api/v1/staff/finance/refunds", Perm: "refunds.request", Status: 201,
		Summary: "Request a refund (someone else approves it)"}),
		func(ctx context.Context, in *struct {
			Key  string `header:"Idempotency-Key" required:"true"`
			Body struct {
				OrderNumber string `json:"orderNumber" minLength:"3"`
				AmountKobo  int64  `json:"amountKobo" minimum:"1"`
				Reason      string `json:"reason" minLength:"5" maxLength:"500"`
				Method      string `json:"method,omitempty" enum:"original,bank_transfer,store_credit" default:"original"`
			}
		}) (*struct{ Body *store.PaymentsRefund }, error) {
			actor := httpx.MustPrincipal(ctx).UserID
			out, _, err := idempotency.Do(ctx, s.d.Idem, "refunds.request", in.Key, &actor, http.MethodPost, "/api/v1/staff/finance/refunds", in.Body,
				func() (*store.PaymentsRefund, error) {
					var r store.PaymentsRefund
					err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
						o, err := tx.Q.SalesGetOrderByNumber(ctx, in.Body.OrderNumber)
						if err != nil {
							return httpx.NotFound("Order")
						}
						r, err = s.RequestRefund(ctx, tx, o.ID, in.Body.AmountKobo, in.Body.Reason, in.Body.Method, &actor)
						return err
					})
					return &r, err
				})
			return &struct{ Body *store.PaymentsRefund }{out}, err
		})
	httpx.Register(g, staff(httpx.Op{ID: "staffApproveRefund", Method: http.MethodPost, Path: "/api/v1/staff/finance/refunds/{id}/approve", Perm: "refunds.approve_small", StepUp: true,
		Summary: "Approve a refund (not your own; above the threshold needs refunds.approve)"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body store.PaymentsRefund }, error) {
			r, err := s.Approve(ctx, in.ID, httpx.MustPrincipal(ctx).UserID)
			return &struct{ Body store.PaymentsRefund }{r}, err
		})
	httpx.Register(g, staff(httpx.Op{ID: "staffRejectRefund", Method: http.MethodPost, Path: "/api/v1/staff/finance/refunds/{id}/reject", Perm: "refunds.approve_small", Summary: "Reject a refund request"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Note string `json:"note" minLength:"3" maxLength:"500"`
			}
		}) (*struct{ Body store.PaymentsRefund }, error) {
			r, err := s.Reject(ctx, in.ID, httpx.MustPrincipal(ctx).UserID, in.Body.Note)
			return &struct{ Body store.PaymentsRefund }{r}, err
		})

	// ---- Staff: payments, providers, reconciliation
	httpx.Register(g, staff(httpx.Op{ID: "staffOrderPayments", Method: http.MethodGet, Path: "/api/v1/staff/finance/orders/{number}/payments", Perm: "finance.view", Summary: "Payments for an order"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
		}) (*struct{ Body []store.PaymentsPayment }, error) {
			o, err := s.d.Q.SalesGetOrderByNumber(ctx, in.Number)
			if err != nil {
				return nil, httpx.NotFound("Order")
			}
			rows, err := s.d.Q.PaymentsForOrder(ctx, &o.ID)
			return &struct{ Body []store.PaymentsPayment }{rows}, err
		})
	httpx.Register(g, staff(httpx.Op{ID: "staffVerifyPayment", Method: http.MethodPost, Path: "/api/v1/staff/finance/payments/{id}/verify", Perm: "finance.view",
		Summary: "Re-verify a payment with its provider now"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body store.PaymentsPayment }, error) {
			if err := s.Reconcile(ctx, in.ID); err != nil {
				return nil, err
			}
			p, err := s.d.Q.PaymentsGet(ctx, in.ID)
			return &struct{ Body store.PaymentsPayment }{p}, err
		})
	httpx.Register(g, staff(httpx.Op{ID: "staffProviderHealth", Method: http.MethodGet, Path: "/api/v1/staff/finance/providers", Perm: "finance.view", Summary: "Provider circuit breakers"}),
		func(ctx context.Context, _ *struct{}) (*struct {
			Body struct {
				Available []string                       `json:"available"`
				Health    []store.PaymentsProviderHealth `json:"health"`
				Fake      bool                           `json:"fake"`
			}
		}, error) {
			out := &struct {
				Body struct {
					Available []string                       `json:"available"`
					Health    []store.PaymentsProviderHealth `json:"health"`
					Fake      bool                           `json:"fake"`
				}
			}{}
			out.Body.Available, out.Body.Fake = s.Providers(ctx), s.d.Cfg.PaymentsFake
			var err error
			out.Body.Health, err = s.d.Q.PaymentsGetHealth(ctx)
			return out, err
		})
	httpx.Register(g, staff(httpx.Op{ID: "staffListExceptions", Method: http.MethodGet, Path: "/api/v1/staff/finance/reconciliation/exceptions", Perm: "finance.view",
		Summary: "Reconciliation exceptions"}),
		func(ctx context.Context, in *struct {
			Status string `query:"status" enum:"open,resolved,escalated,"`
		}) (*struct {
			Body []store.PaymentsReconciliationException
		}, error) {
			p := store.PaymentsListExceptionsParams{Limit: 200}
			if in.Status != "" {
				p.Status = &in.Status
			}
			rows, err := s.d.Q.PaymentsListExceptions(ctx, p)
			if rows == nil {
				rows = []store.PaymentsReconciliationException{}
			}
			return &struct {
				Body []store.PaymentsReconciliationException
			}{rows}, err
		})
	httpx.Register(g, staff(httpx.Op{ID: "staffResolveException", Method: http.MethodPost, Path: "/api/v1/staff/finance/reconciliation/exceptions/{id}/resolve",
		Perm: "reconciliation.resolve", Summary: "Resolve or escalate an exception"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Status     string `json:"status" enum:"resolved,escalated"`
				Resolution string `json:"resolution" minLength:"5" maxLength:"1000"`
			}
		}) (*struct {
			Body store.PaymentsReconciliationException
		}, error) {
			actor := httpx.MustPrincipal(ctx).UserID
			var out store.PaymentsReconciliationException
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var err error
				out, err = tx.Q.PaymentsResolveException(ctx, store.PaymentsResolveExceptionParams{ID: in.ID, Status: in.Body.Status, Resolution: &in.Body.Resolution, ResolvedBy: &actor})
				if err != nil {
					return httpx.Conflict("not_open", "This exception isn't open.")
				}
				return tx.Audit("reconciliation.resolved", "reconciliation_exception", in.ID.String(), in.Body)
			})
			return &struct {
				Body store.PaymentsReconciliationException
			}{out}, err
		})

	if s.d.Cfg.IsDevelopment() && s.d.Cfg.PaymentsFake {
		s.devRoutes(r.Gin)
	}
}

// devRoutes is the fake providers' "hosted checkout": a page with buttons that make the
// provider send a signed webhook, exactly as a real provider would (development only).
func (s *Service) devRoutes(e *gin.Engine) {
	e.GET("/dev/pay/:provider/:reference", func(c *gin.Context) {
		ref := c.Param("reference")
		p, err := s.d.Q.PaymentsGetByReference(c.Request.Context(), store.PaymentsGetByReferenceParams{Provider: c.Param("provider"), ProviderReference: &ref})
		if err != nil {
			c.String(http.StatusNotFound, "unknown reference")
			return
		}
		base := "/dev/pay/" + html.EscapeString(c.Param("provider")) + "/" + html.EscapeString(ref)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, `<!doctype html><meta charset="utf-8"><title>Fake %s checkout</title>
<body style="font-family:system-ui;max-width:28rem;margin:3rem auto"><h1>Fake %s</h1><p>Reference %s — amount %d kobo (development only).</p>
<form method="post" action="%s/success"><button>Pay successfully</button></form>
<form method="post" action="%s/failed"><button>Decline</button></form></body>`,
			html.EscapeString(p.Provider), html.EscapeString(p.Provider), html.EscapeString(ref), p.AmountKobo, base, base)
	})
	// outcome: success | failed | underpay | overpay. ?amount= overrides the paid amount (kobo).
	e.POST("/dev/pay/:provider/:reference/:outcome", func(c *gin.Context) {
		code, body := s.FakeComplete(c.Request.Context(), c.Param("provider"), c.Param("reference"), c.Param("outcome"), c.Query("amount"))
		c.JSON(code, body)
	})
}

// FakeComplete makes a fake provider send a signed charge webhook for a reference.
func (s *Service) FakeComplete(ctx context.Context, provider, ref, outcome, amountOverride string) (int, any) {
	prov, ok := s.providers[provider].(*Fake)
	if !ok {
		return http.StatusNotFound, gin.H{"error": "not a fake provider"}
	}
	p, err := s.d.Q.PaymentsGetByReference(ctx, store.PaymentsGetByReferenceParams{Provider: provider, ProviderReference: &ref})
	if err != nil {
		return http.StatusNotFound, gin.H{"error": "unknown reference"}
	}
	amount, status, event := p.AmountKobo, "success", "charge.success"
	switch outcome {
	case "failed":
		status, event = "failed", "charge.failed"
	case "underpay":
		amount = p.AmountKobo / 2
	case "overpay":
		amount = p.AmountKobo + 100000
	}
	if a, err := strconv.ParseInt(amountOverride, 10, 64); err == nil && a > 0 {
		amount = a
	}
	now := time.Now().UTC()
	payload := map[string]any{"event": event, "data": map[string]any{"id": now.UnixNano(), "reference": ref, "status": status, "amount": amount,
		"currency": "NGN", "channel": map[bool]string{true: "bank_transfer", false: "card"}[provider == "moniepoint"], "fees": amount * 15 / 1000, "paid_at": now}}
	body, _ := json.Marshal(payload)
	dup, err := s.ReceiveWebhook(ctx, provider, body, prov.Sign(body))
	if err != nil {
		return http.StatusInternalServerError, gin.H{"error": fmt.Sprint(err)}
	}
	// The exact bytes and signature are returned so tests can replay the delivery.
	return http.StatusOK, gin.H{"sent": event, "amountKobo": amount, "duplicate": dup, "raw": string(body), "signatureHeader": prov.SignatureHeader(), "signature": prov.Sign(body)}
}
