// Package wms is warehouse work for TechShop-fulfilled orders (docs/orders-fulfilment.md §4–§5):
// pick waves, scanning (binding the exact IMEI/serial sold), packing with a weight check,
// hand-off to an own rider (delivery job) or a third-party carrier, manifests, and signed
// carrier tracking events. Carriers sit behind a port; only a fake carrier exists until the
// owner chooses carriers (open decision #9).
package wms

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/logistics"
	"github.com/dprince-03/techshop/backend/internal/sales"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/pkg/crypto"
)

// waveSize caps how many order lines one pick list holds.
const waveSize = 200

// Service is the warehouse module.
type Service struct {
	d         *kit.Deps
	sales     *sales.Service
	logistics *logistics.Service
	carrier   Carrier
}

// New creates the warehouse service.
func New(d *kit.Deps, sl *sales.Service, lg *logistics.Service) *Service {
	s := &Service{d: d, sales: sl, logistics: lg}
	s.carrier = &FakeCarrier{secret: s.carrierSecret}
	return s
}

// Name implements kit.Module.
func (s *Service) Name() string { return "wms" }

// ---- Waves and picking

// ReleaseWave puts paid lines committed in a warehouse onto a new pick list; their
// fulfilments move to picking.
func (s *Service) ReleaseWave(ctx context.Context, warehouseID, actor uuid.UUID, zone string) (store.FulfilmentPickList, int, error) {
	var pl store.FulfilmentPickList
	n := 0
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		lines, err := tx.Q.FulfilmentWaveCandidates(ctx, store.FulfilmentWaveCandidatesParams{WarehouseID: &warehouseID, Limit: waveSize})
		if err != nil {
			return err
		}
		if len(lines) == 0 {
			return httpx.Conflict("nothing_to_pick", "No paid orders are waiting in this warehouse.")
		}
		var z *string
		if zone != "" {
			z = &zone
		}
		pl, err = tx.Q.FulfilmentCreatePickList(ctx, store.FulfilmentCreatePickListParams{WarehouseID: warehouseID, Zone: z})
		if err != nil {
			return err
		}
		moved := map[uuid.UUID]bool{}
		for _, l := range lines {
			if _, err := tx.Q.FulfilmentAddPickItem(ctx, store.FulfilmentAddPickItemParams{PickListID: pl.ID, FulfilmentID: l.FulfilmentID, OrderLineID: l.OrderLineID, Quantity: l.Quantity}); err != nil {
				return httpx.DB(err, "pick item")
			}
			if !moved[l.FulfilmentID] {
				moved[l.FulfilmentID] = true
				if err := tx.Q.SalesSetFulfilmentWarehouse(ctx, store.SalesSetFulfilmentWarehouseParams{ID: l.FulfilmentID, WarehouseID: &warehouseID}); err != nil {
					return err
				}
				if _, err := s.sales.Move(ctx, tx, l.FulfilmentID, "picking", &actor, "Wave released"); err != nil {
					return err
				}
			}
			n++
		}
		return tx.Audit("wms.wave_released", "pick_list", pl.ID.String(), map[string]any{"warehouse": warehouseID, "lines": n})
	})
	return pl, n, err
}

// Claim gives an open pick list to a picker.
func (s *Service) Claim(ctx context.Context, pickListID, picker uuid.UUID) (store.FulfilmentPickList, error) {
	var out store.FulfilmentPickList
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		pl, err := tx.Q.FulfilmentGetPickListForUpdate(ctx, pickListID)
		if err != nil {
			return httpx.NotFound("Pick list")
		}
		mine := pl.Status == "picking" && pl.PickerID != nil && *pl.PickerID == picker
		if pl.Status != "open" && !mine {
			return httpx.Conflict("already_claimed", "Someone else is picking this list.")
		}
		out, err = tx.Q.FulfilmentSetPickListStatus(ctx, store.FulfilmentSetPickListStatusParams{ID: pickListID, Status: "picking", PickerID: &picker})
		return err
	})
	return out, err
}

// Scan records picked units. Serialised items need the exact IMEI or serial: the unit must be
// in stock in this warehouse for the right variant, and it is bound to the order line (sold).
func (s *Service) Scan(ctx context.Context, pickListID, orderLineID, picker uuid.UUID, qty int32, serial string) (store.FulfilmentPickListItem, bool, error) {
	var item store.FulfilmentPickListItem
	done := false
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		pl, err := tx.Q.FulfilmentGetPickListForUpdate(ctx, pickListID)
		if err != nil {
			return httpx.NotFound("Pick list")
		}
		if (pl.Status != "picking" && pl.Status != "done") || pl.PickerID == nil || *pl.PickerID != picker {
			return httpx.Conflict("not_your_list", "Claim this pick list before scanning.")
		}
		items, err := tx.Q.FulfilmentPickItems(ctx, pickListID)
		if err != nil {
			return err
		}
		var it *store.FulfilmentPickItemsRow
		for i := range items {
			if items[i].OrderLineID == orderLineID {
				it = &items[i]
			}
		}
		if it == nil {
			return httpx.NotFound("Line on this pick list")
		}
		var unitID *uuid.UUID
		if it.IsSerialised {
			if serial == "" {
				return httpx.Invalid("serial_required", "Scan the IMEI or serial number of this device.")
			}
			if qty != 1 {
				return httpx.Invalid("one_at_a_time", "Scan serialised devices one at a time.")
			}
			u, err := tx.Q.InventoryGetDeviceUnitBySerial(ctx, &serial)
			switch {
			case errors.Is(err, pgx.ErrNoRows):
				return httpx.Conflict("unknown_device", "This IMEI/serial isn't in stock. Check the box.")
			case err != nil:
				return err
			case u.Status != "in_stock" || u.WarehouseID == nil || *u.WarehouseID != pl.WarehouseID:
				return httpx.Conflict("device_unavailable", "This device isn't available stock in this warehouse.")
			case u.VariantID != it.VariantID || u.Condition != it.Condition || u.OwnerSellerID != it.SellerID:
				return httpx.Conflict("wrong_device", "Wrong item: this device is a different model or condition.")
			}
			if _, err := tx.Q.InventorySetDeviceUnitStatus(ctx, store.InventorySetDeviceUnitStatusParams{ID: u.ID, Status: "sold"}); err != nil {
				return err
			}
			if _, err := tx.Q.SalesBindDeviceUnit(ctx, store.SalesBindDeviceUnitParams{ID: orderLineID, DeviceUnitID: &u.ID}); err != nil {
				return httpx.Conflict("already_bound", "A device is already bound to this line.")
			}
			unitID = &u.ID
		}
		item, err = tx.Q.FulfilmentScanItem(ctx, store.FulfilmentScanItemParams{Qty: qty, DeviceUnitID: unitID, PickListID: pickListID, OrderLineID: orderLineID})
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.Conflict("over_pick", "That's more than the order needs.")
		}
		if err != nil {
			return err
		}
		open, err := tx.Q.FulfilmentPickListOpenItems(ctx, pickListID)
		if err != nil {
			return err
		}
		if open == 0 {
			done = true
			_, err = tx.Q.FulfilmentSetPickListStatus(ctx, store.FulfilmentSetPickListStatusParams{ID: pickListID, Status: "done"})
		}
		return err
	})
	return item, done, err
}

// ---- Packing and hand-off

// Packed is the result of packing a fulfilment.
type Packed struct {
	Parcel     store.FulfilmentParcel     `json:"parcel"`
	PackRecord store.FulfilmentPackRecord `json:"packRecord"`
	Handoff    string                     `json:"handoff" enum:"rider,carrier" doc:"Own rider (delivery job created) or a carrier (shipment booked)"`
	JobID      *uuid.UUID                 `json:"jobId,omitempty"`
	Shipment   *store.FulfilmentShipment  `json:"shipment,omitempty"`
}

// Pack closes a fully picked fulfilment into a labelled parcel. The weight is compared with
// the catalogue weight (flagged when 10% off). Addresses inside a delivery zone get a rider
// job; everywhere else a carrier shipment is booked.
func (s *Service) Pack(ctx context.Context, fulfilmentID, actor uuid.UUID, weight int32, box string) (Packed, error) {
	var out Packed
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		f, err := tx.Q.SalesGetFulfilmentForUpdate(ctx, fulfilmentID)
		if err != nil {
			return httpx.NotFound("Fulfilment")
		}
		if f.FulfilledBy != "techshop" || f.Status != "picking" {
			return httpx.Conflict("not_packable", "Only TechShop parcels that are being picked can be packed here.")
		}
		open, err := tx.Q.FulfilmentUnpickedForFulfilment(ctx, fulfilmentID)
		if err != nil {
			return err
		}
		if open > 0 {
			return httpx.Conflict("not_fully_picked", "Some items in this order haven't been picked yet.")
		}
		expected, err := tx.Q.FulfilmentExpectedWeight(ctx, fulfilmentID)
		if err != nil {
			return err
		}
		var boxCode *string
		if box != "" {
			boxCode = &box
		}
		label := "TSL-" + labelCode(8)
		out.Parcel, err = tx.Q.FulfilmentCreateParcel(ctx, store.FulfilmentCreateParcelParams{FulfilmentID: fulfilmentID, LabelCode: label, BoxCode: boxCode, WeightGrams: &weight})
		if err != nil {
			return httpx.DB(err, "parcel")
		}
		var exp *int32
		if expected > 0 {
			exp = &expected
		}
		out.PackRecord, err = tx.Q.FulfilmentCreatePackRecord(ctx, store.FulfilmentCreatePackRecordParams{FulfilmentID: fulfilmentID, ParcelID: out.Parcel.ID,
			WeightGrams: weight, ExpectedWeightGrams: exp, PackedBy: actor})
		if err != nil {
			return httpx.DB(err, "pack record")
		}
		if _, err := s.sales.Move(ctx, tx, fulfilmentID, "packed", &actor, "Packed as "+label); err != nil {
			return err
		}
		o, err := tx.Q.SalesGetOrder(ctx, f.OrderID)
		if err != nil {
			return err
		}
		if s.logistics.Covers(ctx, tx.Q, o.ShipTo) {
			j, err := s.logistics.CreateDeliveryJob(ctx, tx, fulfilmentID)
			if err != nil {
				return err
			}
			out.Handoff, out.JobID = "rider", &j.ID
		} else {
			sh, err := s.book(ctx, tx, f, o)
			if err != nil {
				return err
			}
			out.Handoff, out.Shipment = "carrier", &sh
		}
		return tx.Audit("wms.packed", "fulfilment", fulfilmentID.String(), map[string]any{"label": label, "weight": weight, "flagged": out.PackRecord.Flagged})
	})
	return out, err
}

// book creates a carrier shipment for a parcel going outside the delivery zones.
func (s *Service) book(ctx context.Context, tx *uow.Tx, f store.SalesFulfilment, o store.SalesOrder) (store.FulfilmentShipment, error) {
	c, err := tx.Q.FulfilmentGetCarrierByCode(ctx, s.carrier.Code())
	if err != nil {
		return store.FulfilmentShipment{}, httpx.Conflict("no_carrier", "No active carrier is set up for deliveries outside the zones.")
	}
	tracking, err := s.carrier.Book(ctx, o.OrderNumber)
	if err != nil {
		return store.FulfilmentShipment{}, err
	}
	if err := tx.Q.SalesSetFulfilmentMethod(ctx, store.SalesSetFulfilmentMethodParams{ID: f.ID, Method: "carrier"}); err != nil {
		return store.FulfilmentShipment{}, err
	}
	sh, err := tx.Q.FulfilmentCreateShipment(ctx, store.FulfilmentCreateShipmentParams{FulfilmentID: f.ID, CarrierID: &c.ID, TrackingNumber: tracking})
	if err != nil {
		return sh, httpx.DB(err, "shipment")
	}
	return sh, nil
}

// ---- Manifests

// CreateManifest lists the packed parcels leaving with one rider or one carrier.
func (s *Service) CreateManifest(ctx context.Context, warehouseID uuid.UUID, riderID, carrierID *uuid.UUID, labels []string, actor uuid.UUID) (store.FulfilmentManifest, error) {
	var m store.FulfilmentManifest
	if (riderID == nil) == (carrierID == nil) {
		return m, httpx.Invalid("rider_or_carrier", "A manifest goes with exactly one rider or one carrier.")
	}
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		var err error
		m, err = tx.Q.FulfilmentCreateManifest(ctx, store.FulfilmentCreateManifestParams{WarehouseID: warehouseID, RiderID: riderID, CarrierID: carrierID, CreatedBy: actor})
		if err != nil {
			return httpx.DB(err, "manifest")
		}
		for _, l := range labels {
			p, err := tx.Q.FulfilmentParcelByLabel(ctx, strings.TrimSpace(l))
			if err != nil {
				return httpx.Invalid("unknown_label", "Unknown parcel label "+l+".")
			}
			f, err := tx.Q.SalesGetFulfilment(ctx, p.FulfilmentID)
			if err != nil {
				return err
			}
			if f.Status != "packed" || f.WarehouseID == nil || *f.WarehouseID != warehouseID {
				return httpx.Conflict("parcel_not_ready", "Parcel "+l+" isn't packed in this warehouse.")
			}
			if err := tx.Q.FulfilmentAddManifestParcel(ctx, store.FulfilmentAddManifestParcelParams{ManifestID: m.ID, ParcelID: p.ID}); err != nil {
				return httpx.Conflict("already_manifested", "Parcel "+l+" is already on a manifest.")
			}
		}
		return tx.Audit("wms.manifest_created", "manifest", m.ID.String(), map[string]any{"parcels": len(labels)})
	})
	return m, err
}

// SignManifest is the hand-over: the rider or driver signs and every parcel moves to handed_over.
func (s *Service) SignManifest(ctx context.Context, manifestID, actor uuid.UUID, signedBy string) (store.FulfilmentManifest, error) {
	var m store.FulfilmentManifest
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		var err error
		m, err = tx.Q.FulfilmentSignManifest(ctx, store.FulfilmentSignManifestParams{ID: manifestID, SignedByName: &signedBy})
		if err != nil {
			return httpx.Conflict("already_signed", "This manifest is already signed (or doesn't exist).")
		}
		parcels, err := tx.Q.FulfilmentManifestParcels(ctx, manifestID)
		if err != nil {
			return err
		}
		for _, p := range parcels {
			f, err := tx.Q.SalesGetFulfilment(ctx, p.FulfilmentID)
			if err != nil {
				return err
			}
			if f.Status == "packed" {
				if _, err := s.sales.Move(ctx, tx, f.ID, "handed_over", &actor, "Manifest signed by "+signedBy); err != nil {
					return err
				}
			}
		}
		return tx.Audit("wms.manifest_signed", "manifest", manifestID.String(), map[string]any{"signedBy": signedBy})
	})
	return m, err
}

// ---- Carriers

// Carrier is the port for third-party couriers.
type Carrier interface {
	Code() string
	Book(ctx context.Context, reference string) (trackingNumber string, err error)
	VerifySignature(body []byte, signature string) bool
}

// FakeCarrier is the development courier: tracking numbers and signed webhooks, no network.
type FakeCarrier struct{ secret func(code string) []byte }

// Code implements Carrier.
func (f *FakeCarrier) Code() string { return "fake" }

// Book implements Carrier.
func (f *FakeCarrier) Book(_ context.Context, reference string) (string, error) {
	return "FK" + labelCode(10), nil
}

// VerifySignature implements Carrier (HMAC-SHA256 hex over the raw body).
func (f *FakeCarrier) VerifySignature(body []byte, signature string) bool {
	return hmac.Equal([]byte(sign(f.secret("fake"), body)), []byte(signature))
}

// labelCode is a printable, scanner-friendly code without look-alike characters.
func labelCode(n int) string {
	const alphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"
	b := []byte(crypto.RandomToken(n * 2))
	out := make([]byte, n)
	for i := range out {
		out[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(out)
}

func sign(key, body []byte) string {
	m := hmac.New(sha256.New, key)
	m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}

// carrierSecret derives the development carrier's webhook secret (never configured).
func (s *Service) carrierSecret(code string) []byte {
	return crypto.HMAC(s.d.Cfg.TokenHMACKey, "fake-carrier:"+code)
}

// CarrierEvent is one tracking update from a carrier.
type CarrierEvent struct {
	EventID        string    `json:"eventId"`
	TrackingNumber string    `json:"trackingNumber"`
	Status         string    `json:"status"` // in_transit, delivered, exception, lost, returned
	Location       string    `json:"location"`
	OccurredAt     time.Time `json:"occurredAt"`
}

// ApplyCarrierEvent stores a tracking event once and moves the shipment and fulfilment forward.
func (s *Service) ApplyCarrierEvent(ctx context.Context, code string, ev CarrierEvent) (bool, error) {
	dup := false
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		c, err := tx.Q.FulfilmentGetCarrierByCode(ctx, code)
		if err != nil {
			return httpx.NotFound("Carrier")
		}
		sh, err := tx.Q.FulfilmentShipmentByTracking(ctx, store.FulfilmentShipmentByTrackingParams{CarrierID: &c.ID, TrackingNumber: ev.TrackingNumber})
		if err != nil {
			return httpx.NotFound("Shipment")
		}
		var loc *string
		if ev.Location != "" {
			loc = &ev.Location
		}
		if _, err := tx.Q.FulfilmentInsertShipmentEvent(ctx, store.FulfilmentInsertShipmentEventParams{ShipmentID: sh.ID, CarrierEventID: &ev.EventID, Status: ev.Status,
			Location: loc, OccurredAt: ev.OccurredAt}); errors.Is(err, pgx.ErrNoRows) {
			dup = true
			return nil
		} else if err != nil {
			return err
		}
		switch ev.Status {
		case "in_transit", "delivered", "exception", "lost", "returned":
		default:
			return nil // informational
		}
		if _, err := tx.Q.FulfilmentSetShipmentStatus(ctx, store.FulfilmentSetShipmentStatusParams{ID: sh.ID, Status: ev.Status}); err != nil {
			return err
		}
		f, err := tx.Q.SalesGetFulfilment(ctx, sh.FulfilmentID)
		if err != nil {
			return err
		}
		note := "Carrier " + code + ": " + ev.Status
		switch ev.Status {
		case "in_transit":
			if f.Status == "packed" || f.Status == "handed_over" {
				_, err = s.sales.Move(ctx, tx, f.ID, "in_transit", nil, note)
			}
		case "delivered":
			if f.Status == "packed" || f.Status == "handed_over" {
				if _, err = s.sales.Move(ctx, tx, f.ID, "in_transit", nil, note); err != nil {
					return err
				}
			}
			_, err = s.sales.Move(ctx, tx, f.ID, "delivered", nil, note)
		case "returned", "lost", "exception":
			if f.Status == "in_transit" {
				_, err = s.sales.Move(ctx, tx, f.ID, "failed", nil, note)
			}
		}
		return err
	})
	return dup, err
}
