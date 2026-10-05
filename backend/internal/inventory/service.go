// Package inventory is the single code path for stock: every change writes a stock movement and
// updates the level in the same transaction. It also reserves stock race-free for orders,
// tracks serialised device units (IMEI/serial, blocklist checked) and runs transfers and counts.
// Plan: docs/orders-fulfilment.md §3, §5, §9.
package inventory

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/pkg/validate"
)

// Service is the inventory module.
type Service struct{ d *kit.Deps }

// New creates the module.
func New(d *kit.Deps) *Service { return &Service{d: d} }

// Name implements kit.Module.
func (s *Service) Name() string { return "inventory" }

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}

// Move describes one stock change.
type Move struct {
	WarehouseID   uuid.UUID
	VariantID     uuid.UUID
	Condition     string
	OwnerSellerID uuid.UUID
	Delta         int32
	Reason        string
	DeviceUnitID  *uuid.UUID
	RefType       string
	RefID         *uuid.UUID
	Actor         *uuid.UUID
}

// Apply is THE way stock changes: the level and an append-only movement in one transaction.
// Going below zero (or below reserved) fails with a 409.
func (s *Service) Apply(ctx context.Context, tx *uow.Tx, m Move) (store.InventoryInventoryLevel, error) {
	lvl, err := tx.Q.InventoryApplyMovement(ctx, store.InventoryApplyMovementParams{WarehouseID: m.WarehouseID, VariantID: m.VariantID, Condition: m.Condition, OwnerSellerID: m.OwnerSellerID, Delta: m.Delta})
	if err != nil {
		if p := httpx.DB(err, "stock level"); p != nil {
			var prob *httpx.Problem
			if errors.As(p, &prob) && prob.Status == 422 {
				return lvl, httpx.Conflict("insufficient_stock", "Not enough stock for this change.")
			}
			return lvl, p
		}
	}
	var refType *string
	if m.RefType != "" {
		refType = &m.RefType
	}
	if _, err := tx.Q.InventoryInsertMovement(ctx, store.InventoryInsertMovementParams{WarehouseID: m.WarehouseID, VariantID: m.VariantID, Condition: m.Condition,
		OwnerSellerID: m.OwnerSellerID, QuantityDelta: m.Delta, Reason: m.Reason, DeviceUnitID: m.DeviceUnitID, ReferenceType: refType, ReferenceID: m.RefID, CreatedBy: m.Actor}); err != nil {
		return lvl, err
	}
	tx.Emit("variant", m.VariantID, "stock.changed", map[string]any{"variantId": m.VariantID, "warehouseId": m.WarehouseID, "delta": m.Delta, "reason": m.Reason})
	return lvl, nil
}

// ReserveLine is one order line to hold stock for.
type ReserveLine struct {
	OrderID, OrderLineID, ListingID, VariantID, OwnerSellerID uuid.UUID
	Condition                                                 string
	Quantity                                                  int32
	FulfilledBy                                               string
	PreferredWarehouse                                        *uuid.UUID
}

// Reserve holds stock for every line or fails with 409 out_of_stock. Lines must be sorted by the
// caller (listing id) so concurrent checkouts lock rows in the same order (no deadlocks).
func (s *Service) Reserve(ctx context.Context, tx *uow.Tx, lines []ReserveLine, expires time.Time) error {
	for _, l := range lines {
		var wh *uuid.UUID
		if l.FulfilledBy == "seller" {
			if _, err := tx.Q.InventoryReserveSellerStock(ctx, store.InventoryReserveSellerStockParams{Qty: l.Quantity, ListingID: l.ListingID}); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return httpx.Conflict("out_of_stock", "An item in your cart just sold out.")
				}
				return err
			}
		} else {
			w, err := tx.Q.InventoryReserveWarehouse(ctx, store.InventoryReserveWarehouseParams{Qty: l.Quantity, VariantID: l.VariantID, Condition: l.Condition,
				OwnerSellerID: l.OwnerSellerID, PreferredWarehouse: l.PreferredWarehouse})
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return httpx.Conflict("out_of_stock", "An item in your cart just sold out.")
				}
				return err
			}
			wh = &w
		}
		if _, err := tx.Q.InventoryInsertReservation(ctx, store.InventoryInsertReservationParams{OrderID: l.OrderID, OrderLineID: l.OrderLineID, WarehouseID: wh,
			ListingID: l.ListingID, VariantID: l.VariantID, Condition: l.Condition, OwnerSellerID: l.OwnerSellerID, Quantity: l.Quantity, ExpiresAt: expires}); err != nil {
			return err
		}
	}
	return nil
}

// Commit turns an order's reservations into sales (warehouse stock leaves; seller stock was
// already decremented when reserved). Called in the payment transaction.
func (s *Service) Commit(ctx context.Context, tx *uow.Tx, orderID uuid.UUID) error {
	rs, err := tx.Q.InventoryOrderReservations(ctx, orderID)
	if err != nil {
		return err
	}
	for _, r := range rs {
		if r.WarehouseID != nil {
			if err := tx.Q.InventoryCommitWarehouseReservation(ctx, store.InventoryCommitWarehouseReservationParams{Qty: r.Quantity, WarehouseID: *r.WarehouseID,
				VariantID: r.VariantID, Condition: r.Condition, OwnerSellerID: r.OwnerSellerID}); err != nil {
				return err
			}
			ref := "order"
			if _, err := tx.Q.InventoryInsertMovement(ctx, store.InventoryInsertMovementParams{WarehouseID: *r.WarehouseID, VariantID: r.VariantID, Condition: r.Condition,
				OwnerSellerID: r.OwnerSellerID, QuantityDelta: -r.Quantity, Reason: "sale", ReferenceType: &ref, ReferenceID: &orderID}); err != nil {
				return err
			}
			tx.Emit("variant", r.VariantID, "stock.changed", map[string]any{"variantId": r.VariantID, "reason": "sale"})
		}
		if err := tx.Q.InventorySetReservationStatus(ctx, store.InventorySetReservationStatusParams{ID: r.ID, Status: "committed"}); err != nil {
			return err
		}
	}
	return nil
}

// Release returns an order's reserved stock exactly as it was taken.
func (s *Service) Release(ctx context.Context, tx *uow.Tx, orderID uuid.UUID) error {
	rs, err := tx.Q.InventoryOrderReservations(ctx, orderID)
	if err != nil {
		return err
	}
	for _, r := range rs {
		if r.WarehouseID != nil {
			err = tx.Q.InventoryReleaseWarehouseReservation(ctx, store.InventoryReleaseWarehouseReservationParams{Qty: r.Quantity, WarehouseID: *r.WarehouseID,
				VariantID: r.VariantID, Condition: r.Condition, OwnerSellerID: r.OwnerSellerID})
		} else {
			err = tx.Q.InventoryReleaseSellerStock(ctx, store.InventoryReleaseSellerStockParams{Qty: r.Quantity, ListingID: r.ListingID})
		}
		if err != nil {
			return err
		}
		if err := tx.Q.InventorySetReservationStatus(ctx, store.InventorySetReservationStatusParams{ID: r.ID, Status: "released"}); err != nil {
			return err
		}
		tx.Emit("variant", r.VariantID, "stock.changed", map[string]any{"variantId": r.VariantID, "reason": "released"})
	}
	return nil
}

// ReceiveUnit registers a serialised device: IMEI must pass Luhn and must not be blocklisted
// (blocked devices never enter stock). It also books the +1 movement.
func (s *Service) ReceiveUnit(ctx context.Context, tx *uow.Tx, variantID, warehouseID, owner uuid.UUID, condition, imei, serial, via string, cost *int64, refType string, refID *uuid.UUID, actor *uuid.UUID) (store.InventoryDeviceUnit, error) {
	var imeiPtr, serialPtr *string
	if imei != "" {
		if !validate.IMEI(imei) {
			return store.InventoryDeviceUnit{}, httpx.Invalid("invalid_imei", "IMEI "+imei+" fails the check digit; re-scan it from the device (*#06#).")
		}
		blocked, err := tx.Q.InventoryIsBlocklisted(ctx, imei)
		if err != nil {
			return store.InventoryDeviceUnit{}, err
		}
		if blocked {
			tx.Emit("device", uuid.New(), "device.blocked_at_intake", map[string]any{"imei": imei, "via": via})
			return store.InventoryDeviceUnit{}, httpx.Invalid("device_blocked", "This device is on the stolen/blocked list. Quarantine it; a risk case was opened.")
		}
		imeiPtr = &imei
	}
	if serial != "" {
		serialPtr = &serial
	}
	u, err := tx.Q.InventoryCreateDeviceUnit(ctx, store.InventoryCreateDeviceUnitParams{VariantID: variantID, Condition: condition, Imei: imeiPtr, SerialNumber: serialPtr,
		WarehouseID: &warehouseID, OwnerSellerID: owner, CostKobo: cost, Status: "in_stock", AcquiredVia: via})
	if err != nil {
		return u, httpx.DB(err, "device unit")
	}
	reason := map[string]string{"purchase": "purchase_receipt", "trade_in": "trade_in", "return": "return", "consignment": "purchase_receipt"}[via]
	_, err = s.Apply(ctx, tx, Move{WarehouseID: warehouseID, VariantID: variantID, Condition: condition, OwnerSellerID: owner, Delta: 1, Reason: reason, DeviceUnitID: &u.ID, RefType: refType, RefID: refID, Actor: actor})
	return u, err
}

// ExpireArgs releases reservations of unpaid orders (payments verifies first; see orders).
type integrityArgs struct{}

func (integrityArgs) Kind() string { return "inventory.integrity_check" }

type integrityWorker struct {
	river.WorkerDefaults[integrityArgs]
	s *Service
}

// Work logs any level whose on_hand differs from its movements (should never happen).
func (w *integrityWorker) Work(ctx context.Context, _ *river.Job[integrityArgs]) error {
	rows, err := w.s.d.Q.InventoryIntegrityDrift(ctx)
	if err != nil {
		return err
	}
	for _, r := range rows {
		w.s.d.Logger.Error("inventory drift", slog.String("variant", r.VariantID.String()), slog.Int("onHand", int(r.OnHand)), slog.Int("movements", int(r.Movements)))
	}
	return nil
}

// Jobs implements kit.Module.
func (s *Service) Jobs(reg *jobs.Registry) {
	river.AddWorker(reg.Workers, &integrityWorker{s: s})
	reg.Every(24*time.Hour, func() (river.JobArgs, *river.InsertOpts) { return integrityArgs{}, nil })
}
