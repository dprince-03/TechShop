package notify

import (
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// NotificationDTO is an inbox item.
type NotificationDTO struct {
	ID        uuid.UUID  `json:"id"`
	Template  string     `json:"template"`
	Category  string     `json:"category"`
	Body      string     `json:"body,omitempty"`
	Data      any        `json:"data,omitempty"`
	ReadAt    *time.Time `json:"readAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

type inboxOut struct {
	Body struct {
		Data   []NotificationDTO `json:"data"`
		Unread int32             `json:"unread"`
		Page   httpx.Page        `json:"page"`
	}
}

type prefsBody struct {
	Preferences []struct {
		Category string `json:"category" enum:"orders,seller,deals,recommendations,reminders"`
		Channel  string `json:"channel" enum:"sms,email,push,whatsapp"`
		Enabled  bool   `json:"enabled"`
	} `json:"preferences"`
	QuietHours *struct {
		Enabled bool   `json:"enabled"`
		Start   string `json:"start" pattern:"^[0-2][0-9]:[0-5][0-9]$" example:"21:00"`
		End     string `json:"end" pattern:"^[0-2][0-9]:[0-5][0-9]$" example:"08:00"`
	} `json:"quietHours,omitempty"`
}

func toDTO(n store.MessagingNotification) NotificationDTO {
	var p struct {
		Body string `json:"body"`
		Data any    `json:"data"`
	}
	_ = jsonUnmarshal(n.Payload, &p)
	return NotificationDTO{ID: n.ID, Template: n.Template, Category: n.Category, Body: p.Body, Data: p.Data, ReadAt: n.ReadAt, CreatedAt: n.CreatedAt}
}
