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
	Refund          int64        `json:"refund" doc:"Kobo, total for the items"`
	CreatedAt       time.Time    `json:"createdAt"`
}

func returnView(r store.AftersalesGetReturnRow, items []store.AftersalesItemsRow) Return {
	out := Return{RMA: r.RmaNumber, OrderNumber: r.OrderNumber, Reason: r.Reason, Details: r.Details, Method: r.ReturnMethod, Status: r.Status,
		AutoApproved: r.AutoApproved, RejectionReason: r.RejectionReason, CreatedAt: r.CreatedAt, Items: make([]ReturnItem, len(items))}
	for i, it := range items {
		refund := lineRefund(it.LineTotalKobo, it.LineQuantity, it.Quantity)
		out.Items[i] = ReturnItem{ID: it.ID, OrderLineID: it.OrderLineID, ProductName: it.ProductName, VariantName: it.VariantName,
			Quantity: it.Quantity, Refund: refund, Serialised: it.DeviceUnitID != nil, Grade: it.Grade, Disposition: it.Disposition}
		out.Refund += refund
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
