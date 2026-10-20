package sales

import (
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/store"
)

// FulfilmentView is one seller's part of an order with its lines.
type FulfilmentView struct {
	store.SalesOrderFulfilmentsRow
	Lines []store.SalesOrderLine `json:"lines"`
}

// PaymentSummary is what customers see about payments (no provider internals).
type PaymentSummary struct {
	ID         uuid.UUID  `json:"id"`
	Provider   string     `json:"provider"`
	Channel    *string    `json:"channel"`
	Status     string     `json:"status"`
	AmountKobo int64      `json:"amountKobo"`
	PaidAt     *time.Time `json:"paidAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// OrderDetail is an order with fulfilments, timeline and payments.
type OrderDetail struct {
	store.SalesOrder
	Fulfilments []FulfilmentView                `json:"fulfilments"`
	History     []store.SalesOrderStatusHistory `json:"history"`
	Payments    []PaymentSummary                `json:"payments"`
	Refunds     []store.PaymentsRefund          `json:"refunds,omitempty"`
}

type cartOut struct {
	CartToken string `header:"X-Cart-Token"`
	Body      CartView
}

// Tracking is the public view of an order (no address, no prices).
type Tracking struct {
	OrderNumber string    `json:"orderNumber"`
	Status      string    `json:"status"`
	PlacedAt    time.Time `json:"placedAt"`
	Parcels     []struct {
		Seller      string     `json:"seller"`
		Status      string     `json:"status"`
		DeliveredAt *time.Time `json:"deliveredAt"`
		Items       int        `json:"items"`
	} `json:"parcels"`
	History []struct {
		Status string    `json:"status"`
		At     time.Time `json:"at"`
	} `json:"history"`
}
