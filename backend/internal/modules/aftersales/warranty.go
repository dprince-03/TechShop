package aftersales

// Warranty claims and repairs. TechShop handles every claim on items bought here, whoever gives
// the warranty: the customer drops the device at a store, it's checked against the unit sold and
// the stolen-device list, repaired in the workshop, replaced or refunded, and collected.
// Plan: orders-fulfilment.md §7; docs/plan.md 2026-10-23.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/modules/inventory"
	"github.com/dprince-03/techshop/backend/internal/modules/notify"
	"github.com/dprince-03/techshop/backend/internal/modules/risk"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// ClaimInput is a customer's warranty claim.
type ClaimInput struct {
	OrderLineID  uuid.UUID `json:"orderLineId"`
	IMEIOrSerial string    `json:"imeiOrSerial" minLength:"3" maxLength:"40" doc:"The device's IMEI (dial *#06#) or the serial number on the box"`
	Fault        string    `json:"fault" minLength:"5" maxLength:"1000" doc:"What's wrong, and when it started"`
}

// CreateClaim records a claim for one of the caller's items still under warranty.
func (s *Service) CreateClaim(ctx context.Context, userID uuid.UUID, in ClaimInput) (Claim, error) {
	var number string
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		l, err := tx.Q.WarrantyLine(ctx, in.OrderLineID)
		if err != nil || l.CustomerUserID == nil || *l.CustomerUserID != userID {
			return httpx.NotFound("Order item")
		}
		c, err := s.newClaim(ctx, tx, l, in.IMEIOrSerial, in.Fault)
		number = c.ClaimNumber
		return err
	})
	if err != nil {
		return Claim{}, err
	}
	return s.claimView(ctx, number)
}

// newClaim checks cover and the serial, then records the claim.
func (s *Service) newClaim(ctx context.Context, tx *uow.Tx, l store.WarrantyLineRow, serial, fault string) (store.AftersalesWarrantyClaim, error) {
	var c store.AftersalesWarrantyClaim
	switch {
	case l.CancelledAt != nil || l.FulfilmentStatus != "delivered" || l.DeliveredAt == nil:
		return c, httpx.Invalid("not_delivered", "Only delivered items can be claimed under warranty.")
	case !underWarranty(l.DeliveredAt, l.WarrantyMonths, l.WarrantyProvider, time.Now()):
		if l.WarrantyMonths == nil || *l.WarrantyMonths == 0 || l.WarrantyProvider == nil || *l.WarrantyProvider == "none" {
			return c, httpx.Invalid("no_warranty", l.ProductName+" was sold without a warranty. Contact support if it's faulty.")
		}
		return c, httpx.Invalid("warranty_expired", fmt.Sprintf("The warranty on %s ended on %s.", l.ProductName, warrantyEnds(*l.DeliveredAt, *l.WarrantyMonths).Format("02/01/2006")))
	}
	serial = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(serial), " ", ""))
	if l.DeviceUnitID != nil && !serialMatches(serial, l.Imei, l.SerialNumber) {
		return c, httpx.Invalid("serial_mismatch", "That IMEI or serial isn't the one we sold you. Check *#06# or the label on the box.")
	}
	ends := warrantyEnds(*l.DeliveredAt, *l.WarrantyMonths)
	c, err := tx.Q.WarrantyCreateClaim(ctx, store.WarrantyCreateClaimParams{CustomerUserID: *l.CustomerUserID, OrderLineID: &l.ID, DeviceUnitID: l.DeviceUnitID,
		ImeiOrSerial: serial, FaultDescription: strings.TrimSpace(fault), WarrantyProvider: l.WarrantyProvider, WarrantyEndsOn: &ends})
	if err != nil {
		var dup = httpx.DB(err, "warranty claim")
		if p, ok := dup.(*httpx.Problem); ok && p.Status == 409 {
			return c, httpx.Conflict("claim_open", "There's already an open claim for this item.")
		}
		return c, dup
	}
	tx.Emit("warranty_claim", c.ID, "warranty.claimed", map[string]any{"claimId": c.ID, "number": c.ClaimNumber})
	return c, tx.Audit("warranty.claimed", "warranty_claim", c.ID.String(), map[string]string{"number": c.ClaimNumber})
}

// WalkInInput is a claim staff take at the counter.
type WalkInInput struct {
	IMEIOrSerial string     `json:"imeiOrSerial" minLength:"3" maxLength:"40"`
	OrderLineID  *uuid.UUID `json:"orderLineId,omitempty" doc:"Needed when the item has no tracked serial (accessories)"`
	Fault        string     `json:"fault" minLength:"5" maxLength:"1000"`
	StoreID      uuid.UUID  `json:"storeId" doc:"Where the device was handed in"`
}

// WalkIn records and receives a claim brought to a store: the device is found by its IMEI or
// serial (or the order line), then checked in at once.
func (s *Service) WalkIn(ctx context.Context, in WalkInInput) (Claim, error) {
	serial := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(in.IMEIOrSerial), " ", ""))
	lineID := in.OrderLineID
	if lineID == nil {
		id, err := s.d.Q.WarrantySoldUnit(ctx, serial)
		if errors.Is(err, pgx.ErrNoRows) {
			return Claim{}, httpx.Invalid("unknown_device", "We can't find a sale for that IMEI or serial. Pick the order item instead.")
		}
		if err != nil {
			return Claim{}, err
		}
		lineID = &id
	}
	var number string
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		l, err := tx.Q.WarrantyLine(ctx, *lineID)
		if err != nil || l.CustomerUserID == nil {
			return httpx.NotFound("Order item")
		}
		c, err := s.newClaim(ctx, tx, l, serial, in.Fault)
		number = c.ClaimNumber
		return err
	})
	if err != nil {
		return Claim{}, err
	}
	return s.ReceiveClaim(ctx, number, in.StoreID, serial)
}

// ReceiveClaim checks the device in at a store. The serial must match the claim, and a device on
// the stolen-device list is refused: the claim is rejected and a risk case opened.
func (s *Service) ReceiveClaim(ctx context.Context, number string, storeID uuid.UUID, scanned string) (Claim, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	blocked := false
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		c, err := tx.Q.WarrantyClaimForUpdate(ctx, number)
		if err != nil {
			return httpx.NotFound("Warranty claim")
		}
		if c.Status != "new" {
			return httpx.Conflict("bad_transition", "This claim was already "+humanise(c.Status)+".")
		}
		got := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(scanned), " ", ""))
		if got != c.ImeiOrSerial {
			return httpx.Invalid("serial_mismatch", "The device in hand doesn't match the claim's IMEI or serial.")
		}
		st, err := risk.CheckIMEI(ctx, tx.Q, got)
		if err != nil {
			return err
		}
		if st.Blocked {
			blocked = true // commit the rejection and the case, then refuse below
			reason := "This device is on the stolen-device list, so we can't accept it."
			if _, err := tx.Q.WarrantyReject(ctx, store.WarrantyRejectParams{ID: c.ID, RejectionReason: &reason}); err != nil {
				return err
			}
			_, err := s.risk.OpenCase(ctx, tx, risk.NewCase{Kind: "device", SubjectType: "user", SubjectID: c.CustomerUserID, Score: 90, Block: true,
				Reason: fmt.Sprintf("Warranty %s: IMEI %s is blocklisted (%s)", c.ClaimNumber, got, *st.Reason)})
			return err
		}
		if _, err := tx.Q.WarrantyReceive(ctx, store.WarrantyReceiveParams{ID: c.ID, StoreID: &storeID, ReceivedBy: &actor}); err != nil {
			return httpx.DB(err, "store")
		}
		if c.DeviceUnitID != nil {
			if err := tx.Q.WarrantySetUnit(ctx, store.WarrantySetUnitParams{ID: *c.DeviceUnitID, Status: "in_repair", WarehouseID: &storeID}); err != nil {
				return err
			}
		}
		if err := s.tellClaim(ctx, tx, c, "received at the store"); err != nil {
			return err
		}
		return tx.Audit("warranty.received", "warranty_claim", c.ID.String(), map[string]any{"store": storeID})
	})
	if err == nil && blocked {
		return Claim{}, httpx.Conflict("device_blocked", "This device is on the stolen-device list. The claim is rejected and a risk case was opened; keep the device and call your manager.")
	}
	if err != nil {
		return Claim{}, err
	}
	return s.claimView(ctx, number)
}

// RejectClaim refuses a claim the warranty doesn't cover (physical or liquid damage, tampering).
func (s *Service) RejectClaim(ctx context.Context, number, reason string) (Claim, error) {
	if strings.TrimSpace(reason) == "" {
		return Claim{}, httpx.Invalid("reason_required", "Tell the customer why it isn't covered.")
	}
	return s.onClaim(ctx, number, func(tx *uow.Tx, c store.AftersalesWarrantyClaim) error {
		if c.Status != "new" && c.Status != "received" && c.Status != "in_repair" {
			return httpx.Conflict("bad_transition", "This claim was already "+humanise(c.Status)+".")
		}
		if _, err := tx.Q.WarrantyReject(ctx, store.WarrantyRejectParams{ID: c.ID, RejectionReason: &reason}); err != nil {
			return err
		}
		return s.tellClaim(ctx, tx, c, "not covered: "+reason)
	})
}

// RepairInput opens a workshop job for a received claim.
type RepairInput struct {
	Diagnosis    string     `json:"diagnosis,omitempty" maxLength:"1000"`
	TechnicianID *uuid.UUID `json:"technicianId,omitempty"`
}

// StartRepair sends a received claim to the workshop.
func (s *Service) StartRepair(ctx context.Context, number string, in RepairInput) (Claim, error) {
	return s.onClaim(ctx, number, func(tx *uow.Tx, c store.AftersalesWarrantyClaim) error {
		if c.Status != "received" && c.Status != "in_repair" {
			return httpx.Conflict("bad_transition", "Receive the device before repairing it.")
		}
		if open, err := tx.Q.WarrantyOpenRepairs(ctx, &c.ID); err != nil {
			return err
		} else if open > 0 {
			return httpx.Conflict("repair_open", "This claim already has a repair in progress.")
		}
		if _, err := tx.Q.WarrantyCreateRepair(ctx, store.WarrantyCreateRepairParams{WarrantyClaimID: &c.ID, DeviceUnitID: c.DeviceUnitID, TechnicianID: in.TechnicianID,
			Diagnosis: optional(in.Diagnosis)}); err != nil {
			return httpx.DB(err, "technician")
		}
		if c.Status == "received" {
			if _, err := tx.Q.WarrantySetStatus(ctx, store.WarrantySetStatusParams{ID: c.ID, Status: "in_repair"}); err != nil {
				return err
			}
			return s.tellClaim(ctx, tx, c, "in repair")
		}
		return nil
	})
}

// RepairUpdate moves a repair job along.
type RepairUpdate struct {
	Status       string     `json:"status" enum:"diagnosing,awaiting_parts,repairing,done,unrepairable"`
	Diagnosis    string     `json:"diagnosis,omitempty" maxLength:"1000"`
	Notes        string     `json:"notes,omitempty" maxLength:"1000"`
	TechnicianID *uuid.UUID `json:"technicianId,omitempty"`
}

// UpdateRepair records workshop progress. A finished warranty repair makes the claim ready for
// collection (an unrepairable one waits for a replacement or refund). A finished refurbishment
// puts the device into sellable stock as refurbished; an unrepairable one is written off.
func (s *Service) UpdateRepair(ctx context.Context, jobID uuid.UUID, in RepairUpdate) (RepairView, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		j, err := tx.Q.WarrantyRepairForUpdate(ctx, jobID)
		if err != nil {
			return httpx.NotFound("Repair job")
		}
		if !repairCanMove(j.Status, in.Status) {
			return httpx.Conflict("bad_transition", "A repair can't go from "+humanise(j.Status)+" to "+humanise(in.Status)+".")
		}
		if (in.Status == "done" || in.Status == "unrepairable") && strings.TrimSpace(in.Diagnosis) == "" && j.Diagnosis == nil {
			return httpx.Invalid("diagnosis_required", "Record what was wrong before finishing the job.")
		}
		if _, err := tx.Q.WarrantyUpdateRepair(ctx, store.WarrantyUpdateRepairParams{ID: j.ID, Status: in.Status, Diagnosis: optional(in.Diagnosis),
			Notes: optional(in.Notes), TechnicianID: in.TechnicianID}); err != nil {
			return httpx.DB(err, "technician")
		}
		if j.Purpose == "refurbish" {
			return s.finishRefurb(ctx, tx, j.ID, j.DeviceUnitID, in.Status, actor)
		}
		if in.Status != "done" || j.ClaimNumber == nil {
			return nil
		}
		c, err := tx.Q.WarrantyClaimForUpdate(ctx, *j.ClaimNumber)
		if err != nil {
			return err
		}
		note := "Repaired"
		if _, err := tx.Q.WarrantyReady(ctx, store.WarrantyReadyParams{ID: c.ID, Resolution: ptr("repaired"), ResolutionNote: &note}); err != nil {
			return err
		}
		return s.tellClaim(ctx, tx, c, "repaired and ready to collect")
	})
	if err != nil {
		return RepairView{}, err
	}
	return s.repairView(ctx, jobID)
}

// finishRefurb stocks a refurbished device, or writes off one that can't be fixed.
func (s *Service) finishRefurb(ctx context.Context, tx *uow.Tx, jobID uuid.UUID, unitID *uuid.UUID, status string, actor uuid.UUID) error {
	if unitID == nil || (status != "done" && status != "unrepairable") {
		return nil
	}
	u, err := tx.Q.WarrantyUnit(ctx, *unitID)
	if err != nil {
		return err
	}
	if status == "unrepairable" {
		return tx.Q.WarrantySetUnit(ctx, store.WarrantySetUnitParams{ID: u.ID, Status: "written_off", WarehouseID: u.WarehouseID})
	}
	if u.WarehouseID == nil {
		return httpx.Conflict("no_warehouse", "The device has no location; move it to a warehouse first.")
	}
	if _, err := s.inv.Apply(ctx, tx, inventory.Move{WarehouseID: *u.WarehouseID, VariantID: u.VariantID, Condition: "refurbished", OwnerSellerID: u.OwnerSellerID,
		Delta: 1, Reason: "trade_in", DeviceUnitID: &u.ID, RefType: "repair_job", RefID: &jobID, Actor: &actor}); err != nil {
		return err
	}
	return tx.Q.WarrantySetUnit(ctx, store.WarrantySetUnitParams{ID: u.ID, Status: "in_stock", WarehouseID: u.WarehouseID})
}

// ResolveInput settles a claim that can't be (or wasn't) repaired.
type ResolveInput struct {
	Resolution        string     `json:"resolution" enum:"replaced,refunded"`
	ReplacementSerial string     `json:"replacementSerial,omitempty" maxLength:"40" doc:"Replaced, tracked devices: the IMEI or serial of the new unit"`
	WarehouseID       *uuid.UUID `json:"warehouseId,omitempty" doc:"Replaced, untracked items: where the replacement comes from"`
	Note              string     `json:"note,omitempty" maxLength:"500"`
}

// ResolveClaim replaces the item from stock (ready for collection) or raises a refund for one
// unit through the normal refund approval (the faulty device stays with TechShop).
func (s *Service) ResolveClaim(ctx context.Context, number string, in ResolveInput) (Claim, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	return s.onClaim(ctx, number, func(tx *uow.Tx, c store.AftersalesWarrantyClaim) error {
		if c.Status != "received" && c.Status != "in_repair" {
			return httpx.Conflict("bad_transition", "Only devices checked in can be replaced or refunded.")
		}
		if open, err := tx.Q.WarrantyOpenRepairs(ctx, &c.ID); err != nil {
			return err
		} else if open > 0 {
			return httpx.Conflict("repair_open", "Finish or close the repair job first.")
		}
		v, err := tx.Q.WarrantyClaim(ctx, number)
		if err != nil {
			return err
		}
		if in.Resolution == "refunded" {
			amount := lineRefund(*v.LineTotalKobo, *v.LineQuantity, 1)
			alloc := refundAllocation([]refundShare{{FirstParty: v.SellerType != nil && *v.SellerType == "first_party", SellerID: derefID(v.SellerID), LineTotal: *v.LineTotalKobo,
				VAT: derefInt(v.VatKobo), SellerEarning: v.SellerEarningKobo, CommissionBps: derefInt32(v.CommissionBps), LineQty: *v.LineQuantity, Qty: 1, CommissionPosted: true}}, 0)
			r, err := s.pay.RequestItemRefund(ctx, tx, *v.OrderID, amount, "Warranty "+number, alloc, &actor)
			if err != nil {
				return err
			}
			if _, err := tx.Q.WarrantyClose(ctx, store.WarrantyCloseParams{ID: c.ID, ResolutionNote: optional(in.Note), RefundID: &r.ID}); err != nil {
				return err
			}
			return s.tellClaim(ctx, tx, c, "refund on its way (finance approves it first)")
		}
		replacement, err := s.takeReplacement(ctx, tx, c, v, in, actor)
		if err != nil {
			return err
		}
		if c.DeviceUnitID != nil { // the faulty unit stays with us
			if err := tx.Q.WarrantySetUnit(ctx, store.WarrantySetUnitParams{ID: *c.DeviceUnitID, Status: "returned", WarehouseID: c.StoreID}); err != nil {
				return err
			}
		}
		if _, err := tx.Q.WarrantyReady(ctx, store.WarrantyReadyParams{ID: c.ID, Resolution: ptr("replaced"), ResolutionNote: optional(in.Note), ReplacementUnitID: replacement}); err != nil {
			return err
		}
		return s.tellClaim(ctx, tx, c, "replacement ready to collect")
	})
}

// takeReplacement takes the new item out of stock: the scanned unit for tracked devices, one
// unit from a warehouse otherwise.
func (s *Service) takeReplacement(ctx context.Context, tx *uow.Tx, c store.AftersalesWarrantyClaim, v store.WarrantyClaimRow, in ResolveInput, actor uuid.UUID) (*uuid.UUID, error) {
	if v.VariantID == nil {
		return nil, httpx.Invalid("no_variant", "This claim has no product to replace.")
	}
	if c.DeviceUnitID != nil {
		if strings.TrimSpace(in.ReplacementSerial) == "" {
			return nil, httpx.Invalid("replacement_required", "Scan the IMEI or serial of the replacement unit.")
		}
		u, err := tx.Q.WarrantyReplacementUnit(ctx, store.WarrantyReplacementUnitParams{Serial: strings.TrimSpace(in.ReplacementSerial), VariantID: *v.VariantID})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpx.Invalid("replacement_unavailable", "That unit isn't in stock, or isn't the same model.")
		}
		if err != nil {
			return nil, err
		}
		if u.WarehouseID == nil {
			return nil, httpx.Invalid("replacement_unavailable", "That unit isn't in a warehouse.")
		}
		if _, err := s.inv.Apply(ctx, tx, inventory.Move{WarehouseID: *u.WarehouseID, VariantID: u.VariantID, Condition: u.Condition, OwnerSellerID: u.OwnerSellerID,
			Delta: -1, Reason: "warranty_replacement", DeviceUnitID: &u.ID, RefType: "warranty_claim", RefID: &c.ID, Actor: &actor}); err != nil {
			return nil, err
		}
		if err := tx.Q.WarrantySetUnit(ctx, store.WarrantySetUnitParams{ID: u.ID, Status: "sold", WarehouseID: u.WarehouseID}); err != nil {
			return nil, err
		}
		return &u.ID, nil
	}
	if in.WarehouseID == nil {
		return nil, httpx.Invalid("warehouse_required", "Choose where the replacement comes from.")
	}
	line, err := tx.Q.WarrantyLine(ctx, *c.OrderLineID)
	if err != nil {
		return nil, err
	}
	if _, err := s.inv.Apply(ctx, tx, inventory.Move{WarehouseID: *in.WarehouseID, VariantID: *v.VariantID, Condition: "new", OwnerSellerID: line.SellerID,
		Delta: -1, Reason: "warranty_replacement", RefType: "warranty_claim", RefID: &c.ID, Actor: &actor}); err != nil {
		return nil, err
	}
	return nil, nil
}

// CollectClaim records that the customer took the device (or its replacement) home.
func (s *Service) CollectClaim(ctx context.Context, number string) (Claim, error) {
	return s.onClaim(ctx, number, func(tx *uow.Tx, c store.AftersalesWarrantyClaim) error {
		if _, err := tx.Q.WarrantyCollected(ctx, c.ID); errors.Is(err, pgx.ErrNoRows) {
			return httpx.Conflict("bad_transition", "Only claims ready for collection can be collected.")
		} else if err != nil {
			return err
		}
		if c.DeviceUnitID != nil && (c.Resolution == nil || *c.Resolution != "replaced") {
			if err := tx.Q.WarrantySetUnit(ctx, store.WarrantySetUnitParams{ID: *c.DeviceUnitID, Status: "sold"}); err != nil {
				return err
			}
		}
		return tx.Audit("warranty.collected", "warranty_claim", c.ID.String(), nil)
	})
}

func (s *Service) onClaim(ctx context.Context, number string, fn func(*uow.Tx, store.AftersalesWarrantyClaim) error) (Claim, error) {
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		c, err := tx.Q.WarrantyClaimForUpdate(ctx, number)
		if err != nil {
			return httpx.NotFound("Warranty claim")
		}
		return fn(tx, c)
	})
	if err != nil {
		return Claim{}, err
	}
	return s.claimView(ctx, number)
}

// tellClaim tells the customer what happened to their claim.
func (s *Service) tellClaim(ctx context.Context, tx *uow.Tx, c store.AftersalesWarrantyClaim, what string) error {
	data := map[string]any{"ref": c.ClaimNumber, "status": what}
	for _, ch := range []string{"push", "in_app"} {
		if err := s.notify.Send(ctx, tx, notify.Msg{UserID: &c.CustomerUserID, Channel: ch, Category: notify.CatOrders, Template: "warranty_update", Data: data,
			Key: "warranty:" + c.ClaimNumber + ":" + what + ":" + ch}); err != nil {
			return err
		}
	}
	return nil
}

func ptr(s string) *string { return &s }

func derefID(p *uuid.UUID) uuid.UUID {
	if p == nil {
		return uuid.Nil
	}
	return *p
}

func derefInt(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}
