package aftersales

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// TradeInDevice is a catalogue model with what we'd pay for it in each condition.
type TradeInDevice struct {
	VariantID   uuid.UUID        `json:"variantId"`
	Brand       string           `json:"brand"`
	ProductName string           `json:"productName"`
	VariantName string           `json:"variantName"`
	Estimates   map[string]int64 `json:"estimates" doc:"Kobo per condition: like_new, good, fair, faulty"`
}

// TradeInRow is a line of a trade-in list.
type TradeInRow struct {
	Reference         string     `json:"reference"`
	DeviceType        string     `json:"deviceType"`
	Device            string     `json:"device" doc:"Brand, model and storage"`
	DeclaredCondition string     `json:"declaredCondition"`
	Estimate          *int64     `json:"estimate" doc:"Kobo"`
	Offer             *int64     `json:"offer" doc:"Kobo, after inspection"`
	Status            string     `json:"status" enum:"submitted,estimated,awaiting_inspection,inspected,offer_sent,accepted,declined,paid,cancelled,blocked"`
	PayoutMethod      *string    `json:"payoutMethod" enum:"bank,store_credit"`
	Customer          string     `json:"customer" doc:"Staff lists only"`
	OfferExpiresAt    *time.Time `json:"offerExpiresAt"`
	CreatedAt         time.Time  `json:"createdAt"`
}

// TradeIn is one trade-in in full.
type TradeIn struct {
	TradeInRow
	InspectedCondition *string    `json:"inspectedCondition"`
	InspectionNotes    *string    `json:"inspectionNotes"`
	IMEI               *string    `json:"imei"`
	Serial             *string    `json:"serial"`
	Store              *string    `json:"store"`
	ReceivedAt         *time.Time `json:"receivedAt"`
	AccountLast4       *string    `json:"accountLast4"`
	AccountName        *string    `json:"accountName"`
	PaidAt             *time.Time `json:"paidAt"`
	PayoutReference    *string    `json:"payoutReference"`
	DropOff            []Store    `json:"dropOff" doc:"Where to bring the device while it's estimated"`
}

func tradeInRow(t store.AftersalesTradeIn, first, last string) TradeInRow {
	device := t.Brand + " " + t.Model
	if t.Storage != nil {
		device += " " + *t.Storage
	}
	cust := first
	if last != "" {
		cust += " " + last[:1] + "."
	}
	return TradeInRow{Reference: t.Reference, DeviceType: t.DeviceType, Device: device, DeclaredCondition: t.DeclaredCondition, Estimate: t.EstimateKobo, Offer: t.OfferKobo,
		Status: t.Status, PayoutMethod: t.PayoutMethod, Customer: cust, OfferExpiresAt: t.OfferExpiresAt, CreatedAt: t.CreatedAt}
}

func (s *Service) tradeInView(ctx context.Context, ref string) (TradeIn, error) {
	r, err := s.d.Q.TradeIn(ctx, ref)
	if errors.Is(err, pgx.ErrNoRows) {
		return TradeIn{}, httpx.NotFound("Trade-in")
	}
	if err != nil {
		return TradeIn{}, err
	}
	t := store.AftersalesTradeIn{Reference: r.Reference, DeviceType: r.DeviceType, Brand: r.Brand, Model: r.Model, Storage: r.Storage, DeclaredCondition: r.DeclaredCondition,
		EstimateKobo: r.EstimateKobo, OfferKobo: r.OfferKobo, Status: r.Status, PayoutMethod: r.PayoutMethod, OfferExpiresAt: r.OfferExpiresAt, CreatedAt: r.CreatedAt}
	out := TradeIn{TradeInRow: tradeInRow(t, r.FirstName, r.LastName), InspectedCondition: r.InspectedCondition, InspectionNotes: r.InspectionNotes, IMEI: r.Imei,
		Serial: r.SerialNumber, Store: r.StoreName, ReceivedAt: r.ReceivedAt, AccountLast4: r.AccountNumberLast4, AccountName: r.AccountName, PaidAt: r.PaidAt,
		PayoutReference: r.PayoutReference, DropOff: []Store{}}
	if p := httpx.PrincipalFrom(ctx); p == nil || p.Audience != "staff" {
		out.Customer = ""
	}
	if r.Status == "estimated" {
		stores, err := s.d.Q.WarrantyStores(ctx)
		if err != nil {
			return TradeIn{}, err
		}
		for _, st := range stores {
			out.DropOff = append(out.DropOff, Store{ID: st.ID, Name: st.Name, Address: st.Address, City: st.City, State: st.StateCode})
		}
	}
	return out, nil
}

func (s *Service) customerTradeIn(ctx context.Context, userID uuid.UUID, ref string) (TradeIn, error) {
	r, err := s.d.Q.TradeIn(ctx, ref)
	if err != nil || r.UserID != userID {
		return TradeIn{}, httpx.NotFound("Trade-in")
	}
	return s.tradeInView(ctx, ref)
}

// TradeInSummary is the staff page's headline numbers.
type TradeInSummary struct {
	NeedsEstimate      int32 `json:"needsEstimate"`
	AwaitingInspection int32 `json:"awaitingInspection"`
	OffersOut          int32 `json:"offersOut"`
	PayoutsDue         int32 `json:"payoutsDue" doc:"Accepted, to be paid by bank transfer"`
	PaidThisMonth      int32 `json:"paidThisMonth"`
	AvgPaid            int64 `json:"avgPaid" doc:"Kobo, last 90 days"`
}
