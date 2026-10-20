package payments

import (
	"github.com/google/uuid"
)

// PaymentStatus is what the checkout return page polls.
type PaymentStatus struct {
	PaymentID   uuid.UUID `json:"paymentId"`
	Status      string    `json:"status"`
	OrderNumber string    `json:"orderNumber"`
	OrderStatus string    `json:"orderStatus"`
}
