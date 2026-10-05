package notify

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Customer-facing audiences that have an inbox.
var userAuds = []string{auth.AudMarket, auth.AudWholesale, auth.AudCustomerApp, auth.AudSeller, auth.AudLogistics, auth.AudStaff}

// Name implements kit.Module.
func (s *Service) Name() string { return "notify" }

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}

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

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	httpx.Register(g, httpx.Op{ID: "listNotifications", Method: http.MethodGet, Path: "/api/v1/me/notifications", Tag: "Notifications", Auth: userAuds, Summary: "In-app inbox"},
		func(ctx context.Context, in *struct {
			Cursor string `query:"cursor"`
			Limit  int    `query:"limit"`
		}) (*inboxOut, error) {
			p := httpx.MustPrincipal(ctx)
			var cur struct{ Before time.Time }
			if err := s.d.Cursors.Decode(in.Cursor, &cur); err != nil {
				return nil, httpx.BadCursor()
			}
			limit := httpx.Limit(in.Limit)
			var before *time.Time
			if !cur.Before.IsZero() {
				before = &cur.Before
			}
			rows, err := s.d.Q.MessagingListInbox(ctx, store.MessagingListInboxParams{UserID: &p.UserID, Limit: int32(limit + 1), Before: before})
			if err != nil {
				return nil, err
			}
			out := &inboxOut{}
			out.Body.Data = []NotificationDTO{}
			for i, n := range rows {
				if i == limit {
					out.Body.Page = httpx.Page{HasMore: true, NextCursor: s.d.Cursors.Encode(struct{ Before time.Time }{rows[i-1].CreatedAt})}
					break
				}
				out.Body.Data = append(out.Body.Data, toDTO(n))
			}
			out.Body.Unread, _ = s.d.Q.MessagingUnreadCount(ctx, &p.UserID)
			return out, nil
		})

	httpx.Register(g, httpx.Op{ID: "readNotification", Method: http.MethodPost, Path: "/api/v1/me/notifications/{id}/read", Tag: "Notifications", Auth: userAuds, Status: 204, Summary: "Mark one as read"},
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{}, error) {
			p := httpx.MustPrincipal(ctx)
			n, err := s.d.Q.MessagingMarkRead(ctx, store.MessagingMarkReadParams{ID: in.ID, UserID: &p.UserID})
			if err != nil {
				return nil, err
			}
			if n == 0 {
				return nil, httpx.NotFound("notification")
			}
			return nil, nil
		})

	httpx.Register(g, httpx.Op{ID: "readAllNotifications", Method: http.MethodPost, Path: "/api/v1/me/notifications/read-all", Tag: "Notifications", Auth: userAuds, Status: 204, Summary: "Mark all as read"},
		func(ctx context.Context, _ *struct{}) (*struct{}, error) {
			p := httpx.MustPrincipal(ctx)
			_, err := s.d.Q.MessagingMarkAllRead(ctx, &p.UserID)
			return nil, err
		})

	httpx.Register(g, httpx.Op{ID: "getNotificationPreferences", Method: http.MethodGet, Path: "/api/v1/me/notification-preferences", Tag: "Notifications", Auth: userAuds, Summary: "Preferences and quiet hours"},
		func(ctx context.Context, _ *struct{}) (*struct{ Body prefsBody }, error) {
			p := httpx.MustPrincipal(ctx)
			return s.prefs(ctx, p.UserID)
		})

	httpx.Register(g, httpx.Op{ID: "setNotificationPreferences", Method: http.MethodPut, Path: "/api/v1/me/notification-preferences", Tag: "Notifications", Auth: userAuds, Summary: "Update preferences (security and delivery messages can't be turned off)"},
		func(ctx context.Context, in *struct{ Body prefsBody }) (*struct{ Body prefsBody }, error) {
			p := httpx.MustPrincipal(ctx)
			for _, pr := range in.Body.Preferences {
				if err := s.d.Q.MessagingSetPreference(ctx, store.MessagingSetPreferenceParams{UserID: p.UserID, Category: pr.Category, Channel: pr.Channel, Enabled: pr.Enabled}); err != nil {
					return nil, err
				}
			}
			if qh := in.Body.QuietHours; qh != nil {
				st, err1 := parseClock(qh.Start)
				en, err2 := parseClock(qh.End)
				if err1 != nil || err2 != nil {
					return nil, httpx.Invalid("bad_time", "Quiet hours must be HH:MM.")
				}
				if _, err := s.d.Q.MessagingUpsertSettings(ctx, store.MessagingUpsertSettingsParams{UserID: p.UserID, QuietStart: st, QuietEnd: en, QuietHours: qh.Enabled}); err != nil {
					return nil, err
				}
			}
			return s.prefs(ctx, p.UserID)
		})

	httpx.Register(g, httpx.Op{ID: "registerPushToken", Method: http.MethodPost, Path: "/api/v1/me/push-tokens", Tag: "Notifications", Auth: []string{auth.AudCustomerApp, auth.AudLogistics}, Status: 204, Summary: "Register an Expo push token"},
		func(ctx context.Context, in *struct {
			Body struct {
				Token      string `json:"token" minLength:"10" maxLength:"300"`
				App        string `json:"app" enum:"customer,logistics"`
				Platform   string `json:"platform" enum:"ios,android"`
				AppVersion string `json:"appVersion,omitempty"`
				Locale     string `json:"locale,omitempty"`
			}
		}) (*struct{}, error) {
			p := httpx.MustPrincipal(ctx)
			_, err := s.d.Q.MessagingUpsertPushToken(ctx, store.MessagingUpsertPushTokenParams{UserID: p.UserID, App: in.Body.App, Platform: in.Body.Platform,
				Token: in.Body.Token, AppVersion: strOrNil(in.Body.AppVersion), Locale: strOrNil(in.Body.Locale)})
			return nil, err
		})

	httpx.Register(g, httpx.Op{ID: "deletePushToken", Method: http.MethodDelete, Path: "/api/v1/me/push-tokens/{token}", Tag: "Notifications", Auth: []string{auth.AudCustomerApp, auth.AudLogistics}, Status: 204, Summary: "Remove a push token (sign out)"},
		func(ctx context.Context, in *struct {
			Token string `path:"token"`
		}) (*struct{}, error) {
			return nil, s.d.Q.MessagingRevokePushToken(ctx, in.Token)
		})

	// Provider delivery reports (signature checked by each real adapter; the log provider has none).
	httpx.Register(g, httpx.Op{ID: "messagingWebhook", Method: http.MethodPost, Path: "/api/v1/webhooks/messaging/{provider}", Tag: "Webhooks", Status: 204, Summary: "Delivery reports, bounces and complaints"},
		func(ctx context.Context, in *struct {
			Provider string `path:"provider" enum:"log,termii,email,expo"`
			Body     struct {
				MessageID string `json:"messageId"`
				Status    string `json:"status" enum:"delivered,bounced,complained,failed"`
				Address   string `json:"address,omitempty"`
			}
		}) (*struct{}, error) {
			id, err := s.d.Q.MessagingMarkDeliveredByProvider(ctx, store.MessagingMarkDeliveredByProviderParams{Provider: &in.Provider, ProviderMessageID: &in.Body.MessageID})
			if err != nil {
				return nil, httpx.DB(err, "message")
			}
			if err := s.d.Q.MessagingInsertEvent(ctx, store.MessagingInsertEventParams{NotificationID: id, Kind: in.Body.Status, Meta: []byte(`{}`)}); err != nil {
				return nil, err
			}
			if (in.Body.Status == "bounced" || in.Body.Status == "complained") && in.Body.Address != "" {
				reason := map[string]string{"bounced": "hard_bounce", "complained": "complaint"}[in.Body.Status]
				ch := "email"
				if in.Provider == "termii" {
					ch = "sms"
				}
				return nil, s.d.Q.MessagingSuppress(ctx, store.MessagingSuppressParams{Channel: ch, AddressHash: s.AddressHash(in.Body.Address), Reason: reason})
			}
			return nil, nil
		})

	// Staff: suppressions and templates.
	httpx.Register(g, httpx.Op{ID: "staffListSuppressions", Method: http.MethodGet, Path: "/api/v1/staff/notify/suppressions", Tag: "Staff · Messaging", Auth: []string{auth.AudStaff}, Perm: "notify.templates", Summary: "Suppression list"},
		func(ctx context.Context, _ *struct{}) (*struct {
			Body []store.MessagingMessageSuppression
		}, error) {
			rows, err := s.d.Q.MessagingListSuppressions(ctx, 200)
			return &struct {
				Body []store.MessagingMessageSuppression
			}{rows}, err
		})
	httpx.Register(g, httpx.Op{ID: "staffListTemplates", Method: http.MethodGet, Path: "/api/v1/staff/notify/templates", Tag: "Staff · Messaging", Auth: []string{auth.AudStaff}, Perm: "notify.templates", Summary: "Marketing templates"},
		func(ctx context.Context, _ *struct{}) (*struct {
			Body []store.MessagingListTemplatesRow
		}, error) {
			rows, err := s.d.Q.MessagingListTemplates(ctx)
			return &struct {
				Body []store.MessagingListTemplatesRow
			}{rows}, err
		})

	// Realtime inbox stream (SSE).
	r.Gin.GET("/api/v1/me/notifications/stream", func(c *gin.Context) {
		p, ok := r.Public.GinPrincipal(c.GetHeader("Authorization"), userAuds...)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httpx.Unauthenticated("Sign in to continue."))
			return
		}
		n, _ := s.d.Q.MessagingUnreadCount(c.Request.Context(), &p.UserID)
		s.d.Hub.ServeSSE(c, "user:"+p.UserID.String(), map[string]any{"unread": n})
	})

	// Development helper: read recent messages to an address (e.g. to pick up an OTP).
	if s.d.Cfg.IsDevelopment() {
		r.Gin.GET("/dev/messages", func(c *gin.Context) {
			rows, err := s.d.Q.MessagingDevRecent(c.Request.Context(), []byte(c.Query("to")))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, rows)
		})
	}
}

func (s *Service) prefs(ctx context.Context, user uuid.UUID) (*struct{ Body prefsBody }, error) {
	rows, err := s.d.Q.MessagingGetPreferences(ctx, user)
	if err != nil {
		return nil, err
	}
	out := &struct{ Body prefsBody }{}
	for _, r := range rows {
		out.Body.Preferences = append(out.Body.Preferences, struct {
			Category string `json:"category" enum:"orders,seller,deals,recommendations,reminders"`
			Channel  string `json:"channel" enum:"sms,email,push,whatsapp"`
			Enabled  bool   `json:"enabled"`
		}{r.Category, r.Channel, r.Enabled})
	}
	qh := &struct {
		Enabled bool   `json:"enabled"`
		Start   string `json:"start" pattern:"^[0-2][0-9]:[0-5][0-9]$" example:"21:00"`
		End     string `json:"end" pattern:"^[0-2][0-9]:[0-5][0-9]$" example:"08:00"`
	}{true, s.quietFrom, s.quietTo}
	if st, err := s.d.Q.MessagingGetSettings(ctx, user); err == nil {
		qh.Enabled, qh.Start, qh.End = st.QuietHours, clock(st.QuietStart.Microseconds), clock(st.QuietEnd.Microseconds)
	}
	out.Body.QuietHours = qh
	return out, nil
}

func toDTO(n store.MessagingNotification) NotificationDTO {
	var p struct {
		Body string `json:"body"`
		Data any    `json:"data"`
	}
	_ = jsonUnmarshal(n.Payload, &p)
	return NotificationDTO{ID: n.ID, Template: n.Template, Category: n.Category, Body: p.Body, Data: p.Data, ReadAt: n.ReadAt, CreatedAt: n.CreatedAt}
}

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Jobs is implemented in notify.go (Service.Jobs); this assertion keeps the interface honest.
var _ kit.Module = (*Service)(nil)
var _ = jobs.QueueCritical
