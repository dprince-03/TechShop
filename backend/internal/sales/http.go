package sales

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/idempotency"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/ratelimit"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/pkg/validate"
)

var (
	buyers     = []string{auth.AudMarket, auth.AudCustomerApp, auth.AudWholesale}
	trackLimit = ratelimit.Rule{Name: "track", Max: 20, Window: 10 * time.Minute}
)

// channelFor maps the caller's audience to the order channel.
func channelFor(aud string) string {
	switch aud {
	case auth.AudCustomerApp:
		return "app"
	case auth.AudWholesale:
		return "wholesale"
	}
	return "market"
}

// FulfilmentView is one seller's part of an order with its lines.
type FulfilmentView struct {
	store.SalesOrderFulfilmentsRow
	Lines []store.SalesOrderLine `json:"lines"`
}

// PaymentSummary is what customers see about payments (no provider internals).
type PaymentSummary struct {
	ID         uuid.UUID  `json:"id"`
	Provider   string     `json:"provider"`
	Channel    *string    `json:"channel"`
	Status     string     `json:"status"`
	AmountKobo int64      `json:"amountKobo"`
	PaidAt     *time.Time `json:"paidAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// OrderDetail is an order with fulfilments, timeline and payments.
type OrderDetail struct {
	store.SalesOrder
	Fulfilments []FulfilmentView                `json:"fulfilments"`
	History     []store.SalesOrderStatusHistory `json:"history"`
	Payments    []PaymentSummary                `json:"payments"`
	Refunds     []store.PaymentsRefund          `json:"refunds,omitempty"`
}

// Detail loads everything about an order.
func (s *Service) Detail(ctx context.Context, o store.SalesOrder, withRefunds bool) (OrderDetail, error) {
	q := s.d.Q
	d := OrderDetail{SalesOrder: o, Fulfilments: []FulfilmentView{}, Payments: []PaymentSummary{}}
	fs, err := q.SalesOrderFulfilments(ctx, o.ID)
	if err != nil {
		return d, err
	}
	lines, err := q.SalesOrderLines(ctx, o.ID)
	if err != nil {
		return d, err
	}
	for _, f := range fs {
		v := FulfilmentView{SalesOrderFulfilmentsRow: f, Lines: []store.SalesOrderLine{}}
		for _, l := range lines {
			if l.FulfilmentID == f.ID {
				v.Lines = append(v.Lines, l)
			}
		}
		d.Fulfilments = append(d.Fulfilments, v)
	}
	if d.History, err = q.SalesOrderHistory(ctx, o.ID); err != nil {
		return d, err
	}
	pays, err := q.PaymentsForOrder(ctx, &o.ID)
	if err != nil {
		return d, err
	}
	for _, p := range pays {
		d.Payments = append(d.Payments, PaymentSummary{p.ID, p.Provider, p.Channel, p.Status, p.AmountKobo, p.PaidAt, p.CreatedAt})
	}
	if withRefunds {
		d.Refunds, err = q.PaymentsOrderRefunds(ctx, o.ID)
	}
	return d, err
}

type cartOut struct {
	CartToken string `header:"X-Cart-Token"`
	Body      CartView
}

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	cartOp := func(o httpx.Op) httpx.Op { o.Tag = "Cart & checkout"; o.Auth = buyers; o.Optional = true; return o }
	buyOp := func(o httpx.Op) httpx.Op { o.Tag = "Cart & checkout"; o.Auth = buyers; return o }
	meOp := func(o httpx.Op) httpx.Op { o.Tag = "Orders"; o.Auth = buyers; return o }
	staffOp := func(o httpx.Op) httpx.Op { o.Tag = "Staff · Orders"; o.Auth = []string{auth.AudStaff}; return o }

	ref := func(ctx context.Context, token string) (CartRef, string) {
		if p := httpx.PrincipalFrom(ctx); p != nil {
			return CartRef{UserID: &p.UserID}, channelFor(p.Audience)
		}
		return CartRef{Token: token}, "market"
	}
	view := func(ctx context.Context, cr CartRef, ch, coupon string) (*cartOut, error) {
		v, err := s.ViewCart(ctx, cr, ch, coupon)
		if err != nil {
			return nil, err
		}
		if cr.UserID == nil && cr.Token != "" {
			v.Token = cr.Token
		}
		return &cartOut{CartToken: v.Token, Body: v}, nil
	}

	httpx.Register(g, cartOp(httpx.Op{ID: "getCart", Method: http.MethodGet, Path: "/api/v1/cart", Summary: "The cart with live prices, promotions, delivery and problems"}),
		func(ctx context.Context, in *struct {
			Token  string `header:"X-Cart-Token"`
			Coupon string `query:"coupon" maxLength:"40"`
		}) (*cartOut, error) {
			cr, ch := ref(ctx, in.Token)
			return view(ctx, cr, ch, in.Coupon)
		})
	httpx.Register(g, cartOp(httpx.Op{ID: "addCartItem", Method: http.MethodPost, Path: "/api/v1/cart/items", Summary: "Add a listing (guests get an X-Cart-Token)"}),
		func(ctx context.Context, in *struct {
			Token string `header:"X-Cart-Token"`
			Body  struct {
				ListingID uuid.UUID `json:"listingId"`
				Quantity  int32     `json:"quantity" minimum:"1" maximum:"50"`
				Replace   bool      `json:"replace,omitempty" doc:"Set the quantity instead of adding to it"`
			}
		}) (*cartOut, error) {
			cr, ch := ref(ctx, in.Token)
			if err := s.AddItem(ctx, &cr, in.Body.ListingID, in.Body.Quantity, in.Body.Replace); err != nil {
				return nil, err
			}
			return view(ctx, cr, ch, "")
		})
	httpx.Register(g, cartOp(httpx.Op{ID: "updateCartItem", Method: http.MethodPatch, Path: "/api/v1/cart/items/{id}", Summary: "Change a line's quantity (0 removes it)"}),
		func(ctx context.Context, in *struct {
			Token string    `header:"X-Cart-Token"`
			ID    uuid.UUID `path:"id"`
			Body  struct {
				Quantity int32 `json:"quantity" minimum:"0" maximum:"50"`
			}
		}) (*cartOut, error) {
			cr, ch := ref(ctx, in.Token)
			if err := s.SetQuantity(ctx, cr, in.ID, in.Body.Quantity); err != nil {
				return nil, err
			}
			return view(ctx, cr, ch, "")
		})
	httpx.Register(g, cartOp(httpx.Op{ID: "removeCartItem", Method: http.MethodDelete, Path: "/api/v1/cart/items/{id}", Summary: "Remove a line"}),
		func(ctx context.Context, in *struct {
			Token string    `header:"X-Cart-Token"`
			ID    uuid.UUID `path:"id"`
		}) (*cartOut, error) {
			cr, ch := ref(ctx, in.Token)
			if err := s.SetQuantity(ctx, cr, in.ID, 0); err != nil {
				return nil, err
			}
			return view(ctx, cr, ch, "")
		})
	httpx.Register(g, buyOp(httpx.Op{ID: "mergeCart", Method: http.MethodPost, Path: "/api/v1/cart/merge", Summary: "Merge a guest cart into the signed-in cart"}),
		func(ctx context.Context, in *struct {
			Body struct {
				CartToken string `json:"cartToken" minLength:"10" maxLength:"64"`
			}
		}) (*cartOut, error) {
			p := httpx.MustPrincipal(ctx)
			if err := s.Merge(ctx, p.UserID, in.Body.CartToken); err != nil {
				return nil, err
			}
			return view(ctx, CartRef{UserID: &p.UserID}, channelFor(p.Audience), "")
		})

	httpx.Register(g, buyOp(httpx.Op{ID: "paymentMethods", Method: http.MethodGet, Path: "/api/v1/checkout/payment-methods", Summary: "Payment providers available right now"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []string }, error) {
			return &struct{ Body []string }{s.Pay.Providers(ctx)}, nil
		})
	httpx.Register(g, buyOp(httpx.Op{ID: "checkoutPreview", Method: http.MethodPost, Path: "/api/v1/checkout/preview", Summary: "Totals per seller and the priceHash to confirm"}),
		func(ctx context.Context, in *struct {
			Body struct {
				AddressID  *uuid.UUID `json:"addressId,omitempty"`
				CouponCode string     `json:"couponCode,omitempty" maxLength:"40"`
			}
		}) (*struct{ Body *Quote }, error) {
			p := httpx.MustPrincipal(ctx)
			qt, err := s.Preview(ctx, p.UserID, channelFor(p.Audience), in.Body.AddressID, in.Body.CouponCode)
			return &struct{ Body *Quote }{qt}, err
		})
	httpx.Register(g, buyOp(httpx.Op{ID: "placeOrder", Method: http.MethodPost, Path: "/api/v1/orders", Status: 201,
		Summary: "Place the order (idempotent): reserves stock and opens the payment"}),
		func(ctx context.Context, in *struct {
			Key  string `header:"Idempotency-Key" required:"true"`
			Body struct {
				AddressID  uuid.UUID `json:"addressId"`
				CouponCode string    `json:"couponCode,omitempty" maxLength:"40"`
				PriceHash  string    `json:"priceHash" minLength:"8" maxLength:"64"`
				Provider   string    `json:"provider" enum:"paystack,opay,moniepoint"`
			}
		}) (*struct {
			Replayed bool `header:"Idempotent-Replayed"`
			Body     *Placed
		}, error) {
			p := httpx.MustPrincipal(ctx)
			out, replayed, err := idempotency.Do(ctx, s.d.Idem, "orders.create", in.Key, &p.UserID, http.MethodPost, "/api/v1/orders", in.Body, func() (*Placed, error) {
				return s.PlaceOrder(ctx, CheckoutInput{UserID: p.UserID, Channel: channelFor(p.Audience), AddressID: in.Body.AddressID,
					CouponCode: in.Body.CouponCode, PriceHash: in.Body.PriceHash, Provider: in.Body.Provider})
			})
			if err != nil {
				return nil, err
			}
			return &struct {
				Replayed bool `header:"Idempotent-Replayed"`
				Body     *Placed
			}{replayed, out}, nil
		})

	// ---- Customer orders
	httpx.Register(g, meOp(httpx.Op{ID: "listMyOrders", Method: http.MethodGet, Path: "/api/v1/me/orders", Summary: "My orders, newest first"}),
		func(ctx context.Context, in *struct {
			Cursor string `query:"cursor"`
			Limit  int    `query:"limit"`
		}) (*struct {
			Body httpx.List[store.SalesOrder]
		}, error) {
			p := httpx.MustPrincipal(ctx)
			before, err := s.d.Cursors.Before(in.Cursor)
			if err != nil {
				return nil, err
			}
			lim := httpx.Limit(in.Limit)
			rows, err := s.d.Q.SalesListCustomerOrders(ctx, store.SalesListCustomerOrdersParams{CustomerUserID: &p.UserID, Before: before, Limit: int32(lim + 1)})
			if err != nil {
				return nil, err
			}
			return &struct {
				Body httpx.List[store.SalesOrder]
			}{httpx.Paginate(s.d.Cursors, rows, lim, func(o store.SalesOrder) time.Time { return o.PlacedAt })}, nil
		})
	mine := func(ctx context.Context, number string) (store.SalesOrder, error) {
		p := httpx.MustPrincipal(ctx)
		o, err := s.d.Q.SalesCustomerOrderByNumber(ctx, store.SalesCustomerOrderByNumberParams{OrderNumber: number, CustomerUserID: &p.UserID})
		if err != nil {
			return o, httpx.NotFound("Order")
		}
		return o, nil
	}
	httpx.Register(g, meOp(httpx.Op{ID: "getMyOrder", Method: http.MethodGet, Path: "/api/v1/me/orders/{number}", Summary: "Order detail: fulfilments, timeline, payments, refunds"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
		}) (*struct{ Body OrderDetail }, error) {
			o, err := mine(ctx, in.Number)
			if err != nil {
				return nil, err
			}
			d, err := s.Detail(ctx, o, true)
			return &struct{ Body OrderDetail }{d}, err
		})
	type reasonBody struct {
		Reason string `json:"reason" minLength:"3" maxLength:"300"`
	}
	httpx.Register(g, meOp(httpx.Op{ID: "cancelMyOrder", Method: http.MethodPost, Path: "/api/v1/me/orders/{number}/cancel", Summary: "Cancel before it's prepared (paid orders are refunded)"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
			Body   reasonBody
		}) (*struct{ Body OrderDetail }, error) {
			o, err := mine(ctx, in.Number)
			if err != nil {
				return nil, err
			}
			if err := s.CancelOrder(ctx, o, httpx.MustPrincipal(ctx).UserID, in.Body.Reason); err != nil {
				return nil, err
			}
			o, _ = s.d.Q.SalesGetOrder(ctx, o.ID)
			d, err := s.Detail(ctx, o, true)
			return &struct{ Body OrderDetail }{d}, err
		})
	httpx.Register(g, meOp(httpx.Op{ID: "cancelMyOrderLine", Method: http.MethodPost, Path: "/api/v1/me/orders/{number}/lines/{lineId}/cancel", Summary: "Cancel one item before it's prepared"}),
		func(ctx context.Context, in *struct {
			Number string    `path:"number"`
			LineID uuid.UUID `path:"lineId"`
			Body   reasonBody
		}) (*struct{ Body OrderDetail }, error) {
			o, err := mine(ctx, in.Number)
			if err != nil {
				return nil, err
			}
			if err := s.CancelLine(ctx, o, in.LineID, httpx.MustPrincipal(ctx).UserID, in.Body.Reason); err != nil {
				return nil, err
			}
			o, _ = s.d.Q.SalesGetOrder(ctx, o.ID)
			d, err := s.Detail(ctx, o, true)
			return &struct{ Body OrderDetail }{d}, err
		})

	// Realtime order status (SSE): snapshot first, then status events.
	r.Gin.GET("/api/v1/me/orders/:number/stream", func(c *gin.Context) {
		p, ok := r.Public.GinPrincipal(c.GetHeader("Authorization"), buyers...)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httpx.Unauthenticated("Sign in to continue."))
			return
		}
		o, err := s.d.Q.SalesCustomerOrderByNumber(c.Request.Context(), store.SalesCustomerOrderByNumberParams{OrderNumber: c.Param("number"), CustomerUserID: &p.UserID})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, httpx.NotFound("Order"))
			return
		}
		s.d.Hub.ServeSSE(c, "order:"+o.ID.String(), map[string]any{"status": o.Status, "orderNumber": o.OrderNumber})
	})

	// Guest tracking: order number + the phone it was placed with (rate limited per IP).
	httpx.Register(g, httpx.Op{ID: "trackOrder", Method: http.MethodGet, Path: "/api/v1/track/{number}", Tag: "Orders", Summary: "Track an order with its number and phone"},
		func(ctx context.Context, in *struct {
			Number string `path:"number" maxLength:"20"`
			Phone  string `query:"phone" required:"true"`
		}) (*struct{ Body Tracking }, error) {
			ip := httpx.MetaFrom(ctx).IP
			if ok, err := s.d.Limiter.Allow(ctx, trackLimit, ip); err == nil && !ok {
				return nil, httpx.E(429, "rate_limited", "Too many tracking requests. Try again later.")
			}
			phone := validate.NormalisePhone(in.Phone)
			o, err := s.d.Q.SalesGetOrderByNumber(ctx, in.Number)
			if err != nil || phone == "" || o.ContactPhone == nil || *o.ContactPhone != phone {
				return nil, httpx.NotFound("Order") // same answer whether the number or the phone is wrong
			}
			t, err := s.tracking(ctx, o)
			return &struct{ Body Tracking }{t}, err
		})

	// ---- Staff orders
	httpx.Register(g, staffOp(httpx.Op{ID: "staffListOrders", Method: http.MethodGet, Path: "/api/v1/staff/orders", Perm: "orders.view", Summary: "Search orders"}),
		func(ctx context.Context, in *struct {
			Status string `query:"status"`
			Q      string `query:"q" maxLength:"80"`
			Cursor string `query:"cursor"`
			Limit  int    `query:"limit"`
		}) (*struct {
			Body httpx.List[store.SalesStaffListOrdersRow]
		}, error) {
			before, err := s.d.Cursors.Before(in.Cursor)
			if err != nil {
				return nil, err
			}
			lim := httpx.Limit(in.Limit)
			p := store.SalesStaffListOrdersParams{Limit: int32(lim + 1), Before: before}
			if in.Status != "" {
				p.Status = &in.Status
			}
			if in.Q != "" {
				p.Q = &in.Q
			}
			rows, err := s.d.Q.SalesStaffListOrders(ctx, p)
			if err != nil {
				return nil, err
			}
			return &struct {
				Body httpx.List[store.SalesStaffListOrdersRow]
			}{httpx.Paginate(s.d.Cursors, rows, lim, func(o store.SalesStaffListOrdersRow) time.Time { return o.PlacedAt })}, nil
		})
	byNumber := func(ctx context.Context, number string) (store.SalesOrder, error) {
		o, err := s.d.Q.SalesGetOrderByNumber(ctx, number)
		if errors.Is(err, pgx.ErrNoRows) {
			return o, httpx.NotFound("Order")
		}
		return o, err
	}
	httpx.Register(g, staffOp(httpx.Op{ID: "staffGetOrder", Method: http.MethodGet, Path: "/api/v1/staff/orders/{number}", Perm: "orders.view",
		Summary: "Order timeline, fulfilments, payments and refunds (contact masked)"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
		}) (*struct{ Body OrderDetail }, error) {
			o, err := byNumber(ctx, in.Number)
			if err != nil {
				return nil, err
			}
			d, err := s.Detail(ctx, o, true)
			maskContact(&d.SalesOrder)
			return &struct{ Body OrderDetail }{d}, err
		})
	httpx.Register(g, staffOp(httpx.Op{ID: "staffCancelOrder", Method: http.MethodPost, Path: "/api/v1/staff/orders/{number}/cancel", Perm: "orders.cancel", Summary: "Cancel an order before it's prepared"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
			Body   reasonBody
		}) (*struct{ Body OrderDetail }, error) {
			o, err := byNumber(ctx, in.Number)
			if err != nil {
				return nil, err
			}
			if err := s.CancelOrder(ctx, o, httpx.MustPrincipal(ctx).UserID, in.Body.Reason); err != nil {
				return nil, err
			}
			o, _ = s.d.Q.SalesGetOrder(ctx, o.ID)
			d, err := s.Detail(ctx, o, true)
			maskContact(&d.SalesOrder)
			return &struct{ Body OrderDetail }{d}, err
		})
	httpx.Register(g, staffOp(httpx.Op{ID: "staffCancelOrderLine", Method: http.MethodPost, Path: "/api/v1/staff/orders/{number}/lines/{lineId}/cancel", Perm: "orders.cancel",
		Summary: "Cancel one line (e.g. seller out of stock)"}),
		func(ctx context.Context, in *struct {
			Number string    `path:"number"`
			LineID uuid.UUID `path:"lineId"`
			Body   reasonBody
		}) (*struct{ Body OrderDetail }, error) {
			o, err := byNumber(ctx, in.Number)
			if err != nil {
				return nil, err
			}
			if err := s.CancelLine(ctx, o, in.LineID, httpx.MustPrincipal(ctx).UserID, in.Body.Reason); err != nil {
				return nil, err
			}
			o, _ = s.d.Q.SalesGetOrder(ctx, o.ID)
			d, err := s.Detail(ctx, o, true)
			maskContact(&d.SalesOrder)
			return &struct{ Body OrderDetail }{d}, err
		})
}

// Tracking is the public view of an order (no address, no prices).
type Tracking struct {
	OrderNumber string    `json:"orderNumber"`
	Status      string    `json:"status"`
	PlacedAt    time.Time `json:"placedAt"`
	Parcels     []struct {
		Seller      string     `json:"seller"`
		Status      string     `json:"status"`
		DeliveredAt *time.Time `json:"deliveredAt"`
		Items       int        `json:"items"`
	} `json:"parcels"`
	History []struct {
		Status string    `json:"status"`
		At     time.Time `json:"at"`
	} `json:"history"`
}

func (s *Service) tracking(ctx context.Context, o store.SalesOrder) (Tracking, error) {
	d, err := s.Detail(ctx, o, false)
	if err != nil {
		return Tracking{}, err
	}
	t := Tracking{OrderNumber: o.OrderNumber, Status: o.Status, PlacedAt: o.PlacedAt}
	for _, f := range d.Fulfilments {
		t.Parcels = append(t.Parcels, struct {
			Seller      string     `json:"seller"`
			Status      string     `json:"status"`
			DeliveredAt *time.Time `json:"deliveredAt"`
			Items       int        `json:"items"`
		}{f.SellerName, f.Status, f.DeliveredAt, len(f.Lines)})
	}
	for _, h := range d.History {
		t.History = append(t.History, struct {
			Status string    `json:"status"`
			At     time.Time `json:"at"`
		}{h.ToStatus, h.CreatedAt})
	}
	return t, nil
}

// maskContact hides the customer's phone in staff views (customers.reveal_contact reveals it).
func maskContact(o *store.SalesOrder) {
	if o.ContactPhone != nil {
		m := validate.MaskPhone(*o.ContactPhone)
		o.ContactPhone = &m
	}
	var ship map[string]any
	if json.Unmarshal(o.ShipTo, &ship) == nil {
		if p, ok := ship["phone"].(string); ok {
			ship["phone"] = validate.MaskPhone(p)
		}
		o.ShipTo, _ = json.Marshal(ship)
	}
}
