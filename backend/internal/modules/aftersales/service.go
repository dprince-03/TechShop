// Package aftersales runs returns (RMA) — request, approval, receiving, inspection with the
// serial check, restocking and the refund — and, later, warranty, repairs and trade-ins.
// Plan: orders-fulfilment.md §7; refunds go through payments' dual control (payments-finance.md §7).
//
// Scope now: items TechShop shipped. Seller-shipped items need the seller hand-over decision
// (docs/plan.md, 2026-10-21) and change-of-mind returns need the owner's policy, so both are
// refused with an explanation.
package aftersales

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/modules/inventory"
	"github.com/dprince-03/techshop/backend/internal/modules/payments"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/jobs"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/outbox"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Service is the after-sales module.
type Service struct {
	d   *kit.Deps
	pay *payments.Service
	inv *inventory.Service
}

// New creates the module.
func New(d *kit.Deps, pay *payments.Service, inv *inventory.Service) *Service {
	return &Service{d: d, pay: pay, inv: inv}
}

// Name implements kit.Module.
func (s *Service) Name() string { return "aftersales" }

// Jobs implements kit.Module.
func (s *Service) Jobs(*jobs.Registry) {}

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}

// ReturnItemInput is one order line (and how many) to send back.
type ReturnItemInput struct {
	OrderLineID uuid.UUID `json:"orderLineId"`
	Quantity    int32     `json:"quantity" minimum:"1"`
}

// ReturnInput is a customer's return request.
type ReturnInput struct {
	Items    []ReturnItemInput `json:"items" minItems:"1" maxItems:"20"`
	Reason   string            `json:"reason" enum:"faulty,damaged,not_as_described,wrong_item,change_of_mind"`
	Details  string            `json:"details,omitempty" maxLength:"1000"`
	Method   string            `json:"method" enum:"pickup,store_dropoff" doc:"How the item comes back (change of mind: store drop-off only)"`
	Unopened bool              `json:"unopened,omitempty" doc:"Change of mind only: the customer confirms the item is unopened with seals intact"`
}

// RequestReturn records a return for items of the caller's own delivered order. Problem
// returns within 7 days from customers in good standing are approved straight away.
func (s *Service) RequestReturn(ctx context.Context, userID uuid.UUID, orderNumber string, in ReturnInput) (Return, error) {
	changeOfMind := in.Reason == "change_of_mind"
	if changeOfMind && in.Method != "store_dropoff" {
		return Return{}, httpx.Invalid("dropoff_only", "Change-of-mind returns are brought to a TechShop store.")
	}
	if changeOfMind && !in.Unopened {
		return Return{}, httpx.Invalid("must_be_unopened", "Change-of-mind returns must be unopened with the seals intact.")
	}
	var out Return
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		o, err := tx.Q.SalesGetOrderByNumber(ctx, orderNumber)
		if err != nil || o.CustomerUserID == nil || *o.CustomerUserID != userID {
			return httpx.NotFound("Order")
		}
		lines, err := tx.Q.AftersalesReturnableLines(ctx, o.ID)
		if err != nil {
			return err
		}
		byID := map[uuid.UUID]store.AftersalesReturnableLinesRow{}
		for _, l := range lines {
			byID[l.ID] = l
		}
		now := time.Now()
		seen := map[uuid.UUID]bool{}
		for _, it := range in.Items {
			l, ok := byID[it.OrderLineID]
			switch {
			case !ok || seen[it.OrderLineID]:
				return httpx.Invalid("unknown_item", "One of the items isn’t on this order (or is listed twice).")
			case l.CancelledAt != nil || l.FulfilmentStatus != "delivered":
				return httpx.Invalid("not_delivered", "Only delivered items can be returned.")
			case l.FulfilledBy != "techshop":
				return httpx.Conflict("seller_return_unavailable", "Returns for items shipped by a marketplace seller aren’t open yet. Contact support and we’ll help.")
			case !withinWindow(l.DeliveredAt, now):
				return httpx.Invalid("return_window_closed", "The 7-day return window has passed. If it’s faulty, make a warranty claim.")
			case changeOfMind && !changeOfMindAllowed(l.CategorySlug, l.ParentCategorySlug, in.Unopened):
				return httpx.Invalid("not_returnable", l.ProductName+" can’t be returned on a change of mind (hygiene). If it’s faulty, choose that reason.")
			case it.Quantity > l.Quantity-l.AlreadyReturned:
				return httpx.Invalid("too_many", fmt.Sprintf("You can return at most %d of %s.", max(l.Quantity-l.AlreadyReturned, 0), l.ProductName))
			}
			seen[it.OrderLineID] = true
		}
		st, err := tx.Q.AftersalesStanding(ctx, userID)
		if err != nil {
			return err
		}
		// Only our-fault returns are auto-approved (orders-fulfilment.md §7); change of mind waits for support.
		auto := ourFault[in.Reason] && goodStanding(int(st.Returns90d), int(st.Orders90d), int(st.OpenCases))
		status := "requested"
		if auto {
			status = "approved"
			if in.Method == "pickup" {
				status = "awaiting_pickup"
			}
		}
		r, err := tx.Q.AftersalesCreateReturn(ctx, store.AftersalesCreateReturnParams{OrderID: o.ID, RequestedBy: userID, Reason: in.Reason,
			Details: optional(in.Details), ReturnMethod: &in.Method, Status: status, AutoApproved: auto})
		if err != nil {
			return httpx.DB(err, "return")
		}
		for _, it := range in.Items {
			if err := tx.Q.AftersalesAddItem(ctx, store.AftersalesAddItemParams{ReturnRequestID: r.ID, OrderLineID: it.OrderLineID, Quantity: it.Quantity}); err != nil {
				return err
			}
		}
		tx.Emit("return", r.ID, "return.requested", map[string]any{"returnId": r.ID, "rma": r.RmaNumber, "orderId": o.ID, "autoApproved": auto})
		if err := tx.Audit("return.requested", "return", r.ID.String(), map[string]any{"rma": r.RmaNumber, "reason": in.Reason, "autoApproved": auto}); err != nil {
			return err
		}
		out, err = s.view(ctx, tx.Q, r.RmaNumber)
		return err
	})
	return out, err
}

// Decide approves or rejects a return waiting for support.
func (s *Service) Decide(ctx context.Context, rma, decision, reason string) (Return, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	var out Return
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		r, err := tx.Q.AftersalesGetReturnForUpdate(ctx, rma)
		if err != nil {
			return httpx.NotFound("Return")
		}
		// A change-of-mind item that arrives opened can still be rejected at receipt (it goes back).
		lateReject := decision == "reject" && r.Reason == "change_of_mind" && r.Status == "received"
		if r.Status != "requested" && !lateReject {
			return httpx.Conflict("bad_transition", "This return was already "+humanise(r.Status)+".")
		}
		status, why := "approved", (*string)(nil)
		if decision == "reject" {
			if strings.TrimSpace(reason) == "" {
				return httpx.Invalid("reason_required", "Tell the customer why.")
			}
			status, why = "rejected", &reason
		} else if r.ReturnMethod != nil && *r.ReturnMethod == "pickup" {
			status = "awaiting_pickup"
		}
		if _, err := tx.Q.AftersalesDecide(ctx, store.AftersalesDecideParams{ID: r.ID, Status: status, DecidedBy: &actor, RejectionReason: why}); err != nil {
			return err
		}
		if err := tx.Audit("return.decided", "return", r.ID.String(), map[string]string{"decision": status}); err != nil {
			return err
		}
		out, err = s.view(ctx, tx.Q, rma)
		return err
	})
	return out, err
}

// Receive records that the item arrived at a warehouse or store.
func (s *Service) Receive(ctx context.Context, rma string) (Return, error) {
	var out Return
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		r, err := tx.Q.AftersalesGetReturnForUpdate(ctx, rma)
		if err != nil {
			return httpx.NotFound("Return")
		}
		if r.Status != "approved" && r.Status != "awaiting_pickup" {
			return httpx.Conflict("bad_transition", "Only approved returns can be received.")
		}
		if _, err := tx.Q.AftersalesSetStatus(ctx, store.AftersalesSetStatusParams{ID: r.ID, Status: "received"}); err != nil {
			return err
		}
		if err := tx.Audit("return.received", "return", r.ID.String(), nil); err != nil {
			return err
		}
		out, err = s.view(ctx, tx.Q, rma)
		return err
	})
	return out, err
}

// InspectInput is the result of checking one returned item.
type InspectInput struct {
	WarehouseID   uuid.UUID `json:"warehouseId" doc:"Where restocked items go"`
	ScannedSerial string    `json:"scannedSerial,omitempty" maxLength:"40" doc:"IMEI or serial on the item (required for serialised devices)"`
	Grade         string    `json:"grade" enum:"A,B,C,faulty"`
	Disposition   string    `json:"disposition" enum:"restock_new,restock_open_box,repair,return_to_supplier,write_off"`
	Notes         string    `json:"notes,omitempty" maxLength:"500"`
}

// Inspect records the check of one item. A device whose IMEI or serial doesn't match the unit
// sold rejects the whole return and opens a risk case (swap fraud). Restocked items go back into
// stock in the stated condition.
func (s *Service) Inspect(ctx context.Context, rma string, itemID uuid.UUID, in InspectInput) (Return, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	var out Return
	swapped := false
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		r, err := tx.Q.AftersalesGetReturnForUpdate(ctx, rma)
		if err != nil {
			return httpx.NotFound("Return")
		}
		if r.Status != "received" {
			return httpx.Conflict("bad_transition", "Receive the return before inspecting it.")
		}
		items, err := tx.Q.AftersalesItems(ctx, r.ID)
		if err != nil {
			return err
		}
		var item *store.AftersalesItemsRow
		for i := range items {
			if items[i].ID == itemID {
				item = &items[i]
			}
		}
		if item == nil {
			return httpx.NotFound("Return item")
		}
		if item.Grade != nil {
			return httpx.Conflict("already_inspected", "This item was already inspected.")
		}
		var match *bool
		if item.DeviceUnitID != nil {
			unit, err := tx.Q.InventoryGetDeviceUnit(ctx, *item.DeviceUnitID)
			if err != nil {
				return err
			}
			ok := serialMatches(in.ScannedSerial, unit.Imei, unit.SerialNumber)
			match = &ok
			if !ok {
				swapped = true // commit the rejection and the risk case, then refuse below
				return s.swapFraud(ctx, tx, r, item, in.ScannedSerial)
			}
		}
		if _, err := tx.Q.AftersalesInsertInspection(ctx, store.AftersalesInsertInspectionParams{ReturnItemID: item.ID, DeviceUnitID: item.DeviceUnitID,
			ScannedSerial: optional(in.ScannedSerial), ImeiMatch: match, Grade: in.Grade, Disposition: in.Disposition, Notes: optional(in.Notes), InspectedBy: actor}); err != nil {
			return httpx.DB(err, "inspection")
		}
		if cond, restock := stockCondition(in.Disposition); restock {
			if _, err := s.inv.Apply(ctx, tx, inventory.Move{WarehouseID: in.WarehouseID, VariantID: item.VariantID, Condition: cond, OwnerSellerID: item.SellerID,
				Delta: item.Quantity, Reason: "return", DeviceUnitID: item.DeviceUnitID, RefType: "return", RefID: &r.ID, Actor: &actor}); err != nil {
				return err
			}
			if item.DeviceUnitID != nil {
				if err := tx.Q.AftersalesSetUnitCondition(ctx, store.AftersalesSetUnitConditionParams{ID: *item.DeviceUnitID, Condition: cond, Status: "in_stock", WarehouseID: &in.WarehouseID}); err != nil {
					return err
				}
			}
		}
		if allInspected(items, item.ID) {
			if _, err := tx.Q.AftersalesSetStatus(ctx, store.AftersalesSetStatusParams{ID: r.ID, Status: "inspected"}); err != nil {
				return err
			}
		}
		if err := tx.Audit("return.inspected", "return", r.ID.String(), map[string]any{"item": item.ID, "grade": in.Grade, "disposition": in.Disposition}); err != nil {
			return err
		}
		out, err = s.view(ctx, tx.Q, rma)
		return err
	})
	if err == nil && swapped {
		return Return{}, httpx.Conflict("serial_mismatch", "This isn’t the unit we sold (IMEI or serial doesn’t match). The return is rejected and a risk case was opened.")
	}
	return out, err
}

func allInspected(items []store.AftersalesItemsRow, justDone uuid.UUID) bool {
	for _, it := range items {
		if it.ID != justDone && it.Grade == nil {
			return false
		}
	}
	return true
}

// swapFraud rejects the return and opens a risk case when the item isn't the unit sold. It
// returns nil so both are committed; Inspect then answers with the refusal.
func (s *Service) swapFraud(ctx context.Context, tx *uow.Tx, r store.AftersalesReturnRequest, item *store.AftersalesItemsRow, scanned string) error {
	reason := "The IMEI or serial doesn’t match the unit sold."
	if _, err := tx.Q.AftersalesDecide(ctx, store.AftersalesDecideParams{ID: r.ID, Status: "rejected", DecidedBy: &httpx.MustPrincipal(ctx).UserID, RejectionReason: &reason}); err != nil {
		return err
	}
	score := int16(80)
	money := lineRefund(item.LineTotalKobo, item.LineQuantity, item.Quantity)
	if _, err := tx.Q.AftersalesOpenRiskCase(ctx, store.AftersalesOpenRiskCaseParams{Kind: "return", SubjectType: "user", SubjectID: r.RequestedBy,
		Reason: fmt.Sprintf("Return %s: scanned %q doesn't match the unit sold (possible swap)", r.RmaNumber, scanned), Score: &score, MoneyAtRiskKobo: &money}); err != nil {
		return err
	}
	return tx.Audit("return.serial_mismatch", "return", r.ID.String(), map[string]string{"scanned": scanned})
}

// Refund raises the refund for an inspected return: the price paid for each returned item, plus
// the delivery fee when a whole parcel comes back for an our-fault reason. No restocking fee.
func (s *Service) Refund(ctx context.Context, rma string) (Return, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	var out Return
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		r, err := tx.Q.AftersalesGetReturnForUpdate(ctx, rma)
		if err != nil {
			return httpx.NotFound("Return")
		}
		if r.Status != "inspected" {
			return httpx.Conflict("bad_transition", "Inspect every item before refunding.")
		}
		items, err := tx.Q.AftersalesItems(ctx, r.ID)
		if err != nil {
			return err
		}
		amount := refundTotal(r.Reason, items)
		if _, err := s.pay.RequestReturnRefund(ctx, tx, r.OrderID, r.ID, amount, "Return "+r.RmaNumber, &actor); err != nil {
			return err
		}
		resolution := "refund"
		if _, err := tx.Q.AftersalesSetStatus(ctx, store.AftersalesSetStatusParams{ID: r.ID, Status: "refunded", Resolution: &resolution}); err != nil {
			return err
		}
		if err := tx.Audit("return.refund_requested", "return", r.ID.String(), map[string]any{"amountKobo": amount}); err != nil {
			return err
		}
		out, err = s.view(ctx, tx.Q, rma)
		return err
	})
	return out, err
}

func (s *Service) view(ctx context.Context, q *store.Queries, rma string) (Return, error) {
	r, err := q.AftersalesGetReturn(ctx, rma)
	if errors.Is(err, pgx.ErrNoRows) {
		return Return{}, httpx.NotFound("Return")
	}
	if err != nil {
		return Return{}, err
	}
	items, err := q.AftersalesItems(ctx, r.ID)
	return returnView(r, items), err
}

func optional(s string) *string {
	if s = strings.TrimSpace(s); s == "" {
		return nil
	}
	return &s
}

func humanise(s string) string { return strings.ReplaceAll(s, "_", " ") }

// refundTotal is the money back for a return (see Refund).
func refundTotal(reason string, items []store.AftersalesItemsRow) int64 {
	var amount int64
	units := map[uuid.UUID]int32{}
	fees := map[uuid.UUID]store.AftersalesItemsRow{}
	for _, it := range items {
		amount += lineRefund(it.LineTotalKobo, it.LineQuantity, it.Quantity)
		units[it.FulfilmentID] += it.Quantity
		fees[it.FulfilmentID] = it
	}
	for fid, it := range fees {
		amount += deliveryRefund(reason, it.DeliveryFeeKobo, units[fid], it.FulfilmentUnits)
	}
	return amount
}
