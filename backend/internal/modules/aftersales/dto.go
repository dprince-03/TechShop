package aftersales

import (
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/store"
)

// ReturnItem is one item in a return, with its inspection once done.
type ReturnItem struct {
	ID          uuid.UUID `json:"id"`
	OrderLineID uuid.UUID `json:"orderLineId"`
	ProductName string    `json:"productName"`
	VariantName string    `json:"variantName"`
	Quantity    int32     `json:"quantity"`
	Refund      int64     `json:"refund" doc:"Kobo: the price paid for these units"`
	Serialised  bool      `json:"serialised" doc:"The inspector must scan the IMEI or serial"`
	Grade       *string   `json:"grade"`
	Disposition *string   `json:"disposition"`
}

// Return is a return request (RMA) as the customer and staff see it.
type Return struct {
	RMA             string       `json:"rma"`
	OrderNumber     string       `json:"orderNumber"`
	Reason          string       `json:"reason"`
	Details         *string      `json:"details"`
	Method          *string      `json:"method"`
	Status          string       `json:"status" enum:"requested,approved,rejected,awaiting_pickup,received,inspected,refunded,replaced,closed"`
	AutoApproved    bool         `json:"autoApproved"`
	RejectionReason *string      `json:"rejectionReason"`
	Items           []ReturnItem `json:"items"`
	Refund          int64        `json:"refund" doc:"Kobo: the items, plus delivery when a whole parcel comes back for an our-fault reason"`
	Pickup          *Pickup      `json:"pickup" doc:"Rider collection, for pickup returns in a delivery zone"`
	Dispute         *Dispute     `json:"dispute" doc:"A seller's dispute of the inspection, if any"`
	CreatedAt       time.Time    `json:"createdAt"`
}

// Pickup is the rider job collecting a return.
type Pickup struct {
	Status     string     `json:"status" enum:"unassigned,assigned,picked_up,en_route,delivered,failed,returned,cancelled"`
	Assigned   bool       `json:"assigned"`
	HandedInAt *time.Time `json:"handedInAt"`
}

// Dispute is a seller's challenge to an inspection.
type Dispute struct {
	Reason    string     `json:"reason"`
	At        time.Time  `json:"at"`
	Decision  *string    `json:"decision" enum:"upheld,rejected"`
	Note      *string    `json:"note"`
	DecidedAt *time.Time `json:"decidedAt"`
}

func returnView(r store.AftersalesGetReturnRow, items []store.AftersalesItemsRow) Return {
	var dispute *Dispute
	if r.SellerDisputedAt != nil && r.SellerDisputeReason != nil {
		dispute = &Dispute{Reason: *r.SellerDisputeReason, At: *r.SellerDisputedAt, Decision: r.DisputeDecision, Note: r.DisputeNote, DecidedAt: r.DisputeDecidedAt}
	}
	out := Return{Dispute: dispute, Refund: refundTotal(r.Reason, items), RMA: r.RmaNumber, OrderNumber: r.OrderNumber, Reason: r.Reason, Details: r.Details, Method: r.ReturnMethod, Status: r.Status,
		AutoApproved: r.AutoApproved, RejectionReason: r.RejectionReason, CreatedAt: r.CreatedAt, Items: make([]ReturnItem, len(items))}
	for i, it := range items {
		refund := lineRefund(it.LineTotalKobo, it.LineQuantity, it.Quantity)
		out.Items[i] = ReturnItem{ID: it.ID, OrderLineID: it.OrderLineID, ProductName: it.ProductName, VariantName: it.VariantName,
			Quantity: it.Quantity, Refund: refund, Serialised: it.DeviceUnitID != nil, Grade: it.Grade, Disposition: it.Disposition}
	}
	return out
}

// ReturnRow is a line of a returns list.
type ReturnRow struct {
	RMA          string    `json:"rma"`
	OrderNumber  string    `json:"orderNumber"`
	Reason       string    `json:"reason"`
	Status       string    `json:"status"`
	AutoApproved bool      `json:"autoApproved"`
	CreatedAt    time.Time `json:"createdAt"`
}
