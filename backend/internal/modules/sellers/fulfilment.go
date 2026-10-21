package sellers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Seller-shipped orders in Seller Centre (M4 slice 3, orders-fulfilment.md §2.1, §4): the queue,
// accept, pack and cancel. Deadlines are enforced by the sales module's SLA job. Hand-over is a
// drop-off at a TechShop hub (docs/plan.md 2026-10-21): packing gives the parcel a label, hub
// staff scan it in and TechShop delivers it. Sellers never deliver, so they see only the
// recipient's first name and city, not the full address or phone.

var orderRoles = []string{"owner", "manager", "staff"}

// OrderLine is one item a seller must ship.
type OrderLine struct {
	ProductName string     `json:"productName"`
	VariantName string     `json:"variantName"`
	SKU         string     `json:"sku"`
	Condition   string     `json:"condition"`
	Quantity    int32      `json:"quantity"`
	UnitPrice   int64      `json:"unitPrice" doc:"Kobo"`
	LineTotal   int64      `json:"lineTotal" doc:"Kobo"`
	CancelledAt *time.Time `json:"cancelledAt"`
}

// SellerOrder is a fulfilment as its seller sees it.
type SellerOrder struct {
	ID              uuid.UUID   `json:"id"`
	OrderNumber     string      `json:"orderNumber"`
	Status          string      `json:"status"`
	PlacedAt        time.Time   `json:"placedAt"`
	AcceptBy        *time.Time  `json:"acceptBy" doc:"Accept before this or the order is cancelled"`
	PackBy          *time.Time  `json:"packBy" doc:"Pack before this or the order is cancelled"`
	HandoverBy      *time.Time  `json:"handoverBy" doc:"Drop the packed parcel at a TechShop hub before this or the order is cancelled"`
	Label           *string     `json:"label" doc:"Packed orders: the code to write on the parcel; hub staff scan it"`
	RecipientFirst  string      `json:"recipientFirstName"`
	City            string      `json:"city"`
	StateCode       string      `json:"stateCode"`
	CancelledReason *string     `json:"cancelledReason"`
	CreatedAt       time.Time   `json:"createdAt"`
	Lines           []OrderLine `json:"lines,omitempty"`
}

func sellerOrder(f store.SalesFulfilment, number string, shipTo []byte, placed time.Time) SellerOrder {
	var a struct {
		RecipientName string `json:"recipientName"`
		City          string `json:"city"`
		StateCode     string `json:"stateCode"`
	}
	_ = json.Unmarshal(shipTo, &a)
	first, _, _ := strings.Cut(strings.TrimSpace(a.RecipientName), " ")
	return SellerOrder{ID: f.ID, OrderNumber: number, Status: f.Status, PlacedAt: placed, AcceptBy: f.AcceptBy, PackBy: f.PackBy, HandoverBy: f.HandoverBy,
		RecipientFirst: first, City: a.City, StateCode: a.StateCode, CancelledReason: f.CancelledReason, CreatedAt: f.CreatedAt}
}

func asFulfilment(r store.SalesGetSellerFulfilmentRow) store.SalesFulfilment {
	return store.SalesFulfilment{ID: r.ID, OrderID: r.OrderID, SellerID: r.SellerID, FulfilledBy: r.FulfilledBy, Method: r.Method,
		WarehouseID: r.WarehouseID, CollectionStoreID: r.CollectionStoreID, Status: r.Status, DeliveryFeeKobo: r.DeliveryFeeKobo,
		AcceptBy: r.AcceptBy, PackBy: r.PackBy, HandoverBy: r.HandoverBy, DeliveredAt: r.DeliveredAt, CancelledReason: r.CancelledReason, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

// withOrder runs fn on one of the shop's own seller-shipped fulfilments and returns its view.
func (s *Service) withOrder(ctx context.Context, shopID, id uuid.UUID, fn func(tx *uow.Tx, f store.SalesFulfilment) error) (SellerOrder, error) {
	var out SellerOrder
	err := s.inShop(ctx, shopID, orderRoles, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
		r, err := tx.Q.SalesGetSellerFulfilment(ctx, store.SalesGetSellerFulfilmentParams{ID: id, SellerID: shopID})
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && r.FulfilledBy != "seller") {
			return httpx.NotFound("Order")
		}
		if err != nil {
			return err
		}
		if fn != nil {
			if err := fn(tx, asFulfilment(r)); err != nil {
				return err
			}
			if r, err = tx.Q.SalesGetSellerFulfilment(ctx, store.SalesGetSellerFulfilmentParams{ID: id, SellerID: shopID}); err != nil {
				return err
			}
		}
		out = sellerOrder(asFulfilment(r), r.OrderNumber, r.ShipTo, r.PlacedAt)
		parcels, err := tx.Q.FulfilmentParcelsForFulfilment(ctx, id)
		if err != nil {
			return err
		}
		if len(parcels) > 0 {
			out.Label = &parcels[0].LabelCode
		}
		lines, err := tx.Q.SalesFulfilmentLines(ctx, id)
		out.Lines = make([]OrderLine, len(lines))
		for i, l := range lines {
			out.Lines[i] = OrderLine{ProductName: l.ProductName, VariantName: l.VariantName, SKU: l.Sku, Condition: l.Condition,
				Quantity: l.Quantity, UnitPrice: l.UnitPriceKobo, LineTotal: l.LineTotalKobo, CancelledAt: l.CancelledAt}
		}
		return err
	})
	return out, err
}

// DropoffPoint is a TechShop warehouse or hub that receives seller parcels.
type DropoffPoint struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	StateCode string    `json:"stateCode"`
}

func (s *Service) orderRoutes(r *kit.Routes) {
	g := r.Public
	sel := func(o httpx.Op) httpx.Op { o.Tag = "Seller Centre · Orders"; o.Auth = sellerAud; return o }

	httpx.Register(g, sel(httpx.Op{ID: "sellerDropoffPoints", Method: http.MethodGet, Path: "/api/v1/seller/dropoff-points", Summary: "Where to drop packed parcels"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []DropoffPoint }, error) {
			rows, err := s.d.Q.FulfilmentDropoffPoints(ctx)
			out := make([]DropoffPoint, len(rows))
			for i, w := range rows {
				out[i] = DropoffPoint{ID: w.ID, Code: w.Code, Name: w.Name, Address: w.Address, City: w.City, StateCode: w.StateCode}
			}
			return &struct{ Body []DropoffPoint }{out}, err
		})
	base := "/api/v1/seller/shops/{shopId}/orders"

	httpx.Register(g, sel(httpx.Op{ID: "sellerListOrders", Method: http.MethodGet, Path: base, Summary: "Orders to ship, newest first"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			Status string    `query:"status" enum:"pending,accepted,packed,handed_over,in_transit,delivered,failed,returned,cancelled,"`
			Cursor string    `query:"cursor"`
			Limit  int       `query:"limit"`
		}) (*struct{ Body httpx.List[SellerOrder] }, error) {
			before, err := s.d.Cursors.Before(in.Cursor)
			if err != nil {
				return nil, httpx.BadCursor()
			}
			lim := httpx.Limit(in.Limit)
			views := []SellerOrder{}
			err = s.inShop(ctx, in.ShopID, orderRoles, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
				rows, err := tx.Q.SalesSellerFulfilmentsPage(ctx, store.SalesSellerFulfilmentsPageParams{SellerID: in.ShopID, Status: optional(in.Status), Before: before, Lim: int32(lim + 1)})
				for _, r := range rows {
					f := store.SalesFulfilment{ID: r.ID, Status: r.Status, AcceptBy: r.AcceptBy, PackBy: r.PackBy, HandoverBy: r.HandoverBy, CancelledReason: r.CancelledReason, CreatedAt: r.CreatedAt}
					views = append(views, sellerOrder(f, r.OrderNumber, r.ShipTo, r.PlacedAt))
				}
				return err
			})
			if err != nil {
				return nil, err
			}
			return &struct{ Body httpx.List[SellerOrder] }{httpx.Paginate(s.d.Cursors, views, lim, func(v SellerOrder) time.Time { return v.CreatedAt })}, nil
		})

	httpx.Register(g, sel(httpx.Op{ID: "sellerGetOrder", Method: http.MethodGet, Path: base + "/{orderId}", Summary: "An order with its items"}),
		func(ctx context.Context, in *struct {
			ShopID  uuid.UUID `path:"shopId"`
			OrderID uuid.UUID `path:"orderId" doc:"The fulfilment ID"`
		}) (*struct{ Body SellerOrder }, error) {
			o, err := s.withOrder(ctx, in.ShopID, in.OrderID, nil)
			return &struct{ Body SellerOrder }{o}, err
		})

	httpx.Register(g, sel(httpx.Op{ID: "sellerAcceptOrder", Method: http.MethodPost, Path: base + "/{orderId}/accept", Summary: "Accept (starts the packing deadline)"}),
		func(ctx context.Context, in *struct {
			ShopID  uuid.UUID `path:"shopId"`
			OrderID uuid.UUID `path:"orderId"`
		}) (*struct{ Body SellerOrder }, error) {
			uid := httpx.MustPrincipal(ctx).UserID
			o, err := s.withOrder(ctx, in.ShopID, in.OrderID, func(tx *uow.Tx, f store.SalesFulfilment) error {
				if f.Status != "pending" {
					return httpx.Conflict("bad_transition", "Only new orders can be accepted.")
				}
				_, err := s.sales.AcceptFulfilment(ctx, tx, f, uid)
				return err
			})
			return &struct{ Body SellerOrder }{o}, err
		})

	httpx.Register(g, sel(httpx.Op{ID: "sellerPackOrder", Method: http.MethodPost, Path: base + "/{orderId}/pack", Summary: "Mark as packed: returns the parcel label and the drop-off deadline"}),
		func(ctx context.Context, in *struct {
			ShopID  uuid.UUID `path:"shopId"`
			OrderID uuid.UUID `path:"orderId"`
		}) (*struct{ Body SellerOrder }, error) {
			uid := httpx.MustPrincipal(ctx).UserID
			o, err := s.withOrder(ctx, in.ShopID, in.OrderID, func(tx *uow.Tx, f store.SalesFulfilment) error {
				if f.Status != "accepted" {
					return httpx.Conflict("bad_transition", "Accept the order before packing it.")
				}
				p, err := s.wms.SellerParcel(ctx, tx, f.ID)
				if err != nil {
					return err
				}
				_, err = s.sales.PackSellerFulfilment(ctx, tx, f.ID, uid, "Packed by seller as "+p.LabelCode)
				return err
			})
			return &struct{ Body SellerOrder }{o}, err
		})

	httpx.Register(g, sel(httpx.Op{ID: "sellerCancelOrder", Method: http.MethodPost, Path: base + "/{orderId}/cancel", Summary: "Cancel before packing (refunds the customer; counts against the shop)"}),
		func(ctx context.Context, in *struct {
			ShopID  uuid.UUID `path:"shopId"`
			OrderID uuid.UUID `path:"orderId"`
			Body    struct {
				Reason string `json:"reason" enum:"out_of_stock,damaged,price_error,other"`
				Note   string `json:"note,omitempty" maxLength:"300"`
			}
		}) (*struct{ Body SellerOrder }, error) {
			uid := httpx.MustPrincipal(ctx).UserID
			reason := "Seller cancelled: " + strings.ReplaceAll(in.Body.Reason, "_", " ")
			if note := strings.TrimSpace(in.Body.Note); note != "" {
				reason += " (" + note + ")"
			}
			o, err := s.withOrder(ctx, in.ShopID, in.OrderID, func(tx *uow.Tx, f store.SalesFulfilment) error {
				if err := s.sales.CancelFulfilment(ctx, tx, f.ID, &uid, reason, "seller_cancel"); err != nil {
					return err
				}
				return tx.Audit("seller.order_cancelled", "fulfilment", f.ID.String(), map[string]string{"reason": in.Body.Reason})
			})
			return &struct{ Body SellerOrder }{o}, err
		})
}
