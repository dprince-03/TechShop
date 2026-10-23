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

// CoveredItem is a delivered item still under warranty.
type CoveredItem struct {
	OrderLineID uuid.UUID `json:"orderLineId"`
	OrderNumber string    `json:"orderNumber"`
	ProductName string    `json:"productName"`
	VariantName string    `json:"variantName"`
	Provider    string    `json:"provider" enum:"manufacturer,techshop,seller"`
	Months      int16     `json:"months"`
	DeliveredAt time.Time `json:"deliveredAt"`
	EndsOn      time.Time `json:"endsOn"`
	Serialised  bool      `json:"serialised" doc:"A tracked device: the claim needs its IMEI or serial"`
	SerialHint  *string   `json:"serialHint" doc:"Last 4 characters of the IMEI or serial we sold"`
	OpenClaim   *string   `json:"openClaim" doc:"An open claim on this item, if any"`
}

// RepairJob is a workshop job on a claim.
type RepairJob struct {
	ID         uuid.UUID  `json:"id"`
	Status     string     `json:"status" enum:"queued,diagnosing,awaiting_parts,repairing,done,unrepairable"`
	Diagnosis  *string    `json:"diagnosis"`
	Notes      *string    `json:"notes"`
	Technician *string    `json:"technician"`
	StartedAt  *time.Time `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// ClaimRow is a line of a claims list.
type ClaimRow struct {
	Number       string    `json:"number"`
	ProductName  *string   `json:"productName"`
	VariantName  *string   `json:"variantName"`
	OrderNumber  *string   `json:"orderNumber"`
	IMEIOrSerial string    `json:"imeiOrSerial"`
	Fault        string    `json:"fault"`
	Status       string    `json:"status" enum:"new,received,in_repair,ready,collected,rejected,closed"`
	Resolution   *string   `json:"resolution" enum:"repaired,replaced,refunded,not_covered"`
	Store        *string   `json:"store"`
	Customer     string    `json:"customer" doc:"Staff lists only: first name and last initial"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Store is a place to drop off or collect a device.
type Store struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Address string    `json:"address"`
	City    string    `json:"city"`
	State   string    `json:"state"`
}

// HistoryItem is an earlier claim on the same device.
type HistoryItem struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Resolution *string   `json:"resolution"`
	Fault      string    `json:"fault"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Claim is a warranty claim with its repairs.
type Claim struct {
	ClaimRow
	Provider        *string       `json:"provider"`
	WarrantyEndsOn  *time.Time    `json:"warrantyEndsOn"`
	ReceivedAt      *time.Time    `json:"receivedAt"`
	RejectionReason *string       `json:"rejectionReason"`
	ResolutionNote  *string       `json:"resolutionNote"`
	Repairs         []RepairJob   `json:"repairs"`
	DropOff         []Store       `json:"dropOff" doc:"Where to bring the device while the claim is new"`
	History         []HistoryItem `json:"history" doc:"Staff only: earlier claims on this device"`
}

func claimRow(c store.AftersalesWarrantyClaim, product, variant, order, storeName *string, first, last string) ClaimRow {
	cust := first
	if last != "" {
		cust += " " + last[:1] + "."
	}
	return ClaimRow{Number: c.ClaimNumber, ProductName: product, VariantName: variant, OrderNumber: order, IMEIOrSerial: c.ImeiOrSerial, Fault: c.FaultDescription,
		Status: c.Status, Resolution: c.Resolution, Store: storeName, Customer: cust, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
}

func (s *Service) claimView(ctx context.Context, number string) (Claim, error) {
	v, err := s.d.Q.WarrantyClaim(ctx, number)
	if errors.Is(err, pgx.ErrNoRows) {
		return Claim{}, httpx.NotFound("Warranty claim")
	}
	if err != nil {
		return Claim{}, err
	}
	c := store.AftersalesWarrantyClaim{ClaimNumber: v.ClaimNumber, ImeiOrSerial: v.ImeiOrSerial, FaultDescription: v.FaultDescription, Status: v.Status,
		Resolution: v.Resolution, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
	out := Claim{ClaimRow: claimRow(c, v.ProductName, v.VariantName, v.OrderNumber, v.StoreName, v.CustomerFirstName, v.CustomerLastName),
		Provider: v.WarrantyProvider, WarrantyEndsOn: v.WarrantyEndsOn, ReceivedAt: v.ReceivedAt, RejectionReason: v.RejectionReason, ResolutionNote: v.ResolutionNote,
		Repairs: []RepairJob{}, DropOff: []Store{}, History: []HistoryItem{}}
	jobs, err := s.d.Q.WarrantyRepairs(ctx, &v.ID)
	if err != nil {
		return Claim{}, err
	}
	for _, j := range jobs {
		var tech *string
		if j.TechnicianFirstName != nil {
			n := *j.TechnicianFirstName
			tech = &n
		}
		out.Repairs = append(out.Repairs, RepairJob{ID: j.ID, Status: j.Status, Diagnosis: j.Diagnosis, Notes: j.Notes, Technician: tech,
			StartedAt: j.StartedAt, FinishedAt: j.FinishedAt, CreatedAt: j.CreatedAt})
	}
	if v.Status == "new" {
		stores, err := s.d.Q.WarrantyStores(ctx)
		if err != nil {
			return Claim{}, err
		}
		for _, st := range stores {
			out.DropOff = append(out.DropOff, Store{ID: st.ID, Name: st.Name, Address: st.Address, City: st.City, State: st.StateCode})
		}
	}
	if p := httpx.PrincipalFrom(ctx); p != nil && p.Audience == "staff" {
		hist, err := s.d.Q.WarrantyHistory(ctx, store.WarrantyHistoryParams{ImeiOrSerial: v.ImeiOrSerial, ID: v.ID})
		if err != nil {
			return Claim{}, err
		}
		for _, h := range hist {
			out.History = append(out.History, HistoryItem{Number: h.ClaimNumber, Status: h.Status, Resolution: h.Resolution, Fault: h.FaultDescription, CreatedAt: h.CreatedAt})
		}
	} else {
		out.Customer = "" // customers don't need their own masked name
	}
	return out, nil
}

// customerClaim loads one of the caller's claims.
func (s *Service) customerClaim(ctx context.Context, userID uuid.UUID, number string) (Claim, error) {
	v, err := s.d.Q.WarrantyClaim(ctx, number)
	if err != nil || v.CustomerUserID != userID {
		return Claim{}, httpx.NotFound("Warranty claim")
	}
	return s.claimView(ctx, number)
}

// WarrantySummary is the staff page's headline numbers.
type WarrantySummary struct {
	Open              int32   `json:"open"`
	AwaitingDropoff   int32   `json:"awaitingDropoff"`
	InRepair          int32   `json:"inRepair"`
	Ready             int32   `json:"ready"`
	ReplacedThisMonth int32   `json:"replacedThisMonth"`
	AvgTurnaroundDays float64 `json:"avgTurnaroundDays" doc:"Received to ready, last 30 days"`
}

func last4(imei, serial *string) *string {
	for _, v := range []*string{imei, serial} {
		if v != nil && len(*v) >= 4 {
			x := (*v)[len(*v)-4:]
			return &x
		}
	}
	return nil
}

// RepairView is a workshop job with what it's for.
type RepairView struct {
	RepairJob
	Purpose      string  `json:"purpose" enum:"warranty,refurbish"`
	ClaimNumber  *string `json:"claimNumber"`
	TradeIn      *string `json:"tradeIn" doc:"The trade-in a refurbishment came from"`
	Device       string  `json:"device"`
	IMEIOrSerial *string `json:"imeiOrSerial"`
}

func repairViewOf(r store.WarrantyRepairQueueRow) RepairView {
	device := ""
	if r.ProductName != nil {
		device = *r.ProductName
		if r.VariantName != nil {
			device += " " + *r.VariantName
		}
	}
	serial := r.Imei
	if serial == nil {
		serial = r.SerialNumber
	}
	return RepairView{RepairJob: RepairJob{ID: r.ID, Status: r.Status, Diagnosis: r.Diagnosis, Notes: r.Notes, Technician: r.TechnicianFirstName,
		StartedAt: r.StartedAt, FinishedAt: r.FinishedAt, CreatedAt: r.CreatedAt}, Purpose: r.Purpose, ClaimNumber: r.ClaimNumber, TradeIn: r.TradeInRef,
		Device: device, IMEIOrSerial: serial}
}

func (s *Service) repairView(ctx context.Context, id uuid.UUID) (RepairView, error) {
	rows, err := s.d.Q.WarrantyRepairQueue(ctx, store.WarrantyRepairQueueParams{ID: &id})
	if err != nil {
		return RepairView{}, err
	}
	if len(rows) == 0 {
		return RepairView{}, httpx.NotFound("Repair job")
	}
	return repairViewOf(rows[0]), nil
}
