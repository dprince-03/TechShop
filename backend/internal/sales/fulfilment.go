package sales

import (
	"context"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/notify"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

// fulfilmentFlow lists the allowed status moves (docs/orders-fulfilment.md §2.1). Everything
// that moves a fulfilment — warehouse, riders, carriers, sellers — goes through Move.
var fulfilmentFlow = map[string][]string{
	"pending":     {"accepted", "picking", "cancelled"},
	"accepted":    {"packed", "cancelled"},
	"picking":     {"packed", "pending"},
	"packed":      {"handed_over", "in_transit"},
	"handed_over": {"in_transit"},
	"in_transit":  {"delivered", "failed"},
	"failed":      {"in_transit", "returned"},
}

// CanMove reports whether a fulfilment may go from one status to another.
func CanMove(from, to string) bool {
	for _, s := range fulfilmentFlow[from] {
		if s == to {
			return true
		}
	}
	return false
}

// Move changes a fulfilment's status, records the history, re-derives the order status and
// runs the side effects of shipping and delivery (customer messages, commission).
func (s *Service) Move(ctx context.Context, tx *uow.Tx, fulfilmentID uuid.UUID, to string, actor *uuid.UUID, note string) (store.SalesFulfilment, error) {
	f, err := tx.Q.SalesGetFulfilmentForUpdate(ctx, fulfilmentID)
	if err != nil {
		return f, httpx.NotFound("Fulfilment")
	}
	if f.Status == to {
		return f, nil // idempotent: retries and duplicate scans are harmless
	}
	if !CanMove(f.Status, to) {
		return f, httpx.Conflict("bad_transition", "This parcel can't go from "+f.Status+" to "+to+".")
	}
	var reason *string
	if to == "cancelled" {
		reason = &note
	}
	next, err := tx.Q.SalesSetFulfilmentStatus(ctx, store.SalesSetFulfilmentStatusParams{ID: f.ID, Status: to, Reason: reason})
	if err != nil {
		return next, httpx.DB(err, "fulfilment")
	}
	var notePtr *string
	if note != "" {
		notePtr = &note
	}
	if err := tx.Q.SalesInsertHistory(ctx, store.SalesInsertHistoryParams{OrderID: f.OrderID, FulfilmentID: &f.ID, FromStatus: &f.Status, ToStatus: to,
		ActorUserID: actor, Note: notePtr}); err != nil {
		return next, err
	}
	o, err := tx.Q.SalesGetOrder(ctx, f.OrderID)
	if err != nil {
		return next, err
	}
	tx.Emit("fulfilment", f.ID, "fulfilment."+to, map[string]any{"fulfilmentId": f.ID, "orderId": f.OrderID, "sellerId": f.SellerID, "from": f.Status})
	tx.Notify("realtime", map[string]any{"topic": "order:" + f.OrderID.String(), "event": "fulfilment", "data": map[string]any{"fulfilmentId": f.ID, "status": to}})

	switch to {
	case "in_transit":
		if f.Status != "failed" && o.CustomerUserID != nil {
			if err := s.notify.Send(ctx, tx, notify.Msg{UserID: o.CustomerUserID, Channel: "push", Category: "orders", Template: "order_shipped",
				Data: map[string]any{"order": o.OrderNumber}, Key: "fulfilment:" + f.ID.String() + ":shipped"}); err != nil {
				return next, err
			}
		}
	case "delivered":
		if err := s.delivered(ctx, tx, f, o); err != nil {
			return next, err
		}
	}
	status, err := s.Derive(ctx, tx, f.OrderID)
	if err != nil {
		return next, err
	}
	if status == "delivered" && o.Status != "delivered" {
		tx.Emit("order", o.ID, "order.delivered", map[string]any{"orderId": o.ID, "orderNumber": o.OrderNumber, "userId": o.CustomerUserID})
		if o.CustomerUserID != nil {
			if err := s.notify.Send(ctx, tx, notify.Msg{UserID: o.CustomerUserID, Channel: "push", Category: "orders", Template: "order_delivered",
				Data: map[string]any{"order": o.OrderNumber}, Key: "order:" + o.OrderNumber + ":delivered"}); err != nil {
				return next, err
			}
		}
	}
	return next, nil
}

// delivered posts the marketplace commission for a seller's fulfilment (TechShop's own stock
// earns no commission). Revenue for the seller's part is recognised here (§4.2 template).
func (s *Service) delivered(ctx context.Context, tx *uow.Tx, f store.SalesFulfilment, o store.SalesOrder) error {
	seller, err := tx.Q.SellersGet(ctx, f.SellerID)
	if err != nil {
		return err
	}
	if seller.Type == "first_party" {
		return nil
	}
	amount, err := tx.Q.SalesCommissionForFulfilment(ctx, f.ID)
	if err != nil {
		return err
	}
	return s.ledger.Commission(ctx, tx, f.ID, f.SellerID, amount)
}
