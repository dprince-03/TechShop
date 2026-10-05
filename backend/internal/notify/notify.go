// Package notify is the single send pipeline for every message (SMS, email, push, in-app):
// Send writes the notifications row and a River job in the caller's transaction; the worker
// checks consent, preferences, suppressions, caps and quiet hours at send time, renders the
// template and calls a provider. Plan: docs/messaging-marketing.md §3.
package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/realtime"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/pkg/crypto"
)

// Categories and their lanes (messaging-marketing.md §2).
const (
	CatSecurity        = "security"
	CatDelivery        = "delivery"
	CatOrders          = "orders"
	CatAccount         = "account"
	CatSeller          = "seller"
	CatB2B             = "b2b"
	CatDeals           = "deals"
	CatRecommendations = "recommendations"
	CatReminders       = "reminders"
)

func priorityFor(category string) string {
	switch category {
	case CatSecurity, CatDelivery:
		return "critical"
	case CatDeals, CatRecommendations, CatReminders:
		return "marketing"
	}
	return "transactional"
}

func queueFor(priority string) string {
	switch priority {
	case "critical":
		return jobs.QueueCritical
	case "marketing":
		return jobs.QueueMarketing
	}
	return jobs.QueueTransactional
}

// Msg is one message to one recipient on one channel.
type Msg struct {
	UserID              *uuid.UUID
	To                  string // phone (E.164) or email; push resolves tokens from UserID
	Channel             string // sms, email, push, in_app
	Category            string
	Template            string
	Data                map[string]any
	Key                 string // idempotency key, e.g. "order:TS-10482:paid:sms"
	CampaignID          *uuid.UUID
	JourneyEnrollmentID *uuid.UUID
	TemplateVersionID   *uuid.UUID
	ScheduledFor        *time.Time
}

// SendArgs is the River job that delivers one notification.
type SendArgs struct {
	NotificationID uuid.UUID `json:"notificationId"`
}

// Kind identifies the job.
func (SendArgs) Kind() string { return "notify.send" }

// Service sends messages.
type Service struct {
	d         *kit.Deps
	SMS       SMSProvider
	Email     EmailProvider
	Push      PushProvider
	hashKey   string
	quietFrom string
	quietTo   string
}

// New builds the notify service with providers chosen from config (log providers by default).
func New(d *kit.Deps) *Service {
	return &Service{d: d, SMS: LogSMS{d.Logger}, Email: LogEmail{d.Logger}, Push: LogPush{d.Logger},
		hashKey: d.Cfg.BlindIndexKey, quietFrom: d.Cfg.QuietHoursStart, quietTo: d.Cfg.QuietHoursEnd}
}

// AddressHash is the blind index of an address (for suppressions without storing raw addresses twice).
func (s *Service) AddressHash(addr string) []byte { return crypto.HMAC(s.hashKey, addr) }

// Send records the message and enqueues delivery in the caller's transaction.
func (s *Service) Send(ctx context.Context, tx *uow.Tx, m Msg) error {
	if m.Key == "" {
		m.Key = uuid.NewString()
	}
	priority := priorityFor(m.Category)
	payload := map[string]any{"to": m.To, "data": m.Data}
	b, _ := json.Marshal(payload)
	var toHash []byte
	if m.To != "" {
		toHash = s.AddressHash(m.To)
	}
	n, err := tx.Q.MessagingInsertNotification(ctx, store.MessagingInsertNotificationParams{
		UserID: m.UserID, Channel: m.Channel, Category: m.Category, Priority: priority, Template: m.Template,
		TemplateVersionID: m.TemplateVersionID, Locale: "en", Payload: b, IdempotencyKey: m.Key,
		CampaignID: m.CampaignID, JourneyEnrollmentID: m.JourneyEnrollmentID, ToHash: toHash, ScheduledFor: m.ScheduledFor,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // duplicate idempotency key: already sent or queued
	}
	if err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}
	opts := &river.InsertOpts{Queue: queueFor(priority), MaxAttempts: 8}
	if m.ScheduledFor != nil {
		opts.ScheduledAt = *m.ScheduledFor
	}
	tx.Enqueue(SendArgs{NotificationID: n.ID}, opts)
	return nil
}

// SendWorker delivers notifications.
type SendWorker struct {
	river.WorkerDefaults[SendArgs]
	s *Service
}

// Work runs the policy checks and sends.
func (w *SendWorker) Work(ctx context.Context, job *river.Job[SendArgs]) error {
	return w.s.deliver(ctx, job.Args.NotificationID)
}

func (s *Service) deliver(ctx context.Context, id uuid.UUID) error {
	q := s.d.Q
	n, err := q.MessagingGetNotification(ctx, id)
	if err != nil {
		return err
	}
	if n.Status != "queued" {
		return nil // already handled (retry after success)
	}
	var p struct {
		To   string         `json:"to"`
		Data map[string]any `json:"data"`
	}
	_ = json.Unmarshal(n.Payload, &p)

	if reason, until := s.policy(ctx, n, p.To); reason != "" {
		if reason == "quiet_hours" && until != nil {
			if err := q.MessagingReschedule(ctx, store.MessagingRescheduleParams{ID: n.ID, ScheduledFor: until}); err != nil {
				return err
			}
			return river.JobSnooze(time.Until(*until))
		}
		return q.MessagingMarkSkipped(ctx, store.MessagingMarkSkippedParams{ID: n.ID, SkipReason: &reason})
	}

	body, err := s.render(ctx, n, p.Data)
	if err != nil {
		msg := err.Error()
		return q.MessagingMarkFailed(ctx, store.MessagingMarkFailedParams{ID: n.ID, Error: &msg})
	}

	var provider, providerID string
	switch n.Channel {
	case "sms":
		provider, providerID, err = s.SMS.SendSMS(ctx, p.To, body, n.Priority == "marketing")
	case "email":
		provider, providerID, err = s.Email.SendEmail(ctx, p.To, subjectFor(n.Template), body)
	case "push":
		provider, providerID, err = s.sendPush(ctx, n, body)
		if errors.Is(err, errNoPushToken) && n.Category == CatOrders {
			return s.fallbackToSMS(ctx, n, p.Data)
		}
	case "in_app":
		provider, providerID = "inbox", n.ID.String()
		if n.UserID != nil {
			s.publishInbox(ctx, *n.UserID, n.ID, body)
		}
	default:
		err = fmt.Errorf("unsupported channel %s", n.Channel)
	}
	if err != nil {
		if errors.Is(err, errNoPushToken) {
			r := "invalid_address"
			return q.MessagingMarkSkipped(ctx, store.MessagingMarkSkippedParams{ID: n.ID, SkipReason: &r})
		}
		return err // River retries with backoff
	}
	extra, _ := json.Marshal(map[string]string{"body": body})
	cost := int64(0)
	if n.Channel == "sms" {
		cost = 500 // ₦5 placeholder per page (messaging-marketing.md §3.2)
	}
	return q.MessagingMarkSent(ctx, store.MessagingMarkSentParams{ID: n.ID, Provider: &provider, ProviderMessageID: &providerID, CostKobo: &cost, Extra: extra})
}

func subjectFor(template string) string {
	switch template {
	case "otp":
		return "Your TechShop code"
	case "order_confirmed":
		return "Your TechShop order is confirmed"
	case "staff_invite":
		return "Your TechShop Staff portal invite"
	}
	return "TechShop"
}

var errNoPushToken = errors.New("no active push token")

func (s *Service) sendPush(ctx context.Context, n store.MessagingNotification, body string) (string, string, error) {
	if n.UserID == nil {
		return "", "", errNoPushToken
	}
	tokens, err := s.d.Q.MessagingActivePushTokens(ctx, store.MessagingActivePushTokensParams{UserID: *n.UserID})
	if err != nil {
		return "", "", err
	}
	if len(tokens) == 0 {
		return "", "", errNoPushToken
	}
	var lastID, provider string
	for _, t := range tokens {
		provider, lastID, err = s.Push.SendPush(ctx, t.Token, "TechShop", body)
		if err != nil {
			return "", "", err
		}
	}
	return provider, lastID, nil
}

// fallbackToSMS sends an order update by SMS when the customer has no push token.
func (s *Service) fallbackToSMS(ctx context.Context, n store.MessagingNotification, data map[string]any) error {
	if n.UserID == nil || !HasChannel(n.Template, "sms") {
		r := "invalid_address"
		return s.d.Q.MessagingMarkSkipped(ctx, store.MessagingMarkSkippedParams{ID: n.ID, SkipReason: &r})
	}
	u, err := s.d.Q.IdentityGetUser(ctx, *n.UserID)
	if err != nil || u.Phone == nil {
		r := "invalid_address"
		return s.d.Q.MessagingMarkSkipped(ctx, store.MessagingMarkSkippedParams{ID: n.ID, SkipReason: &r})
	}
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		r := "invalid_address"
		if err := tx.Q.MessagingMarkSkipped(ctx, store.MessagingMarkSkippedParams{ID: n.ID, SkipReason: &r}); err != nil {
			return err
		}
		return s.Send(ctx, tx, Msg{UserID: n.UserID, To: *u.Phone, Channel: "sms", Category: n.Category, Template: n.Template, Data: data, Key: n.IdempotencyKey + ":sms-fallback"})
	})
}

func (s *Service) publishInbox(ctx context.Context, userID, id uuid.UUID, body string) {
	data, _ := json.Marshal(map[string]any{"id": id, "body": body})
	msg, _ := json.Marshal(realtime.Message{Topic: "user:" + userID.String(), Event: "notification", Data: data})
	if _, err := s.d.Pool.Exec(ctx, "select pg_notify($1, $2)", realtime.ChannelRealtime, string(msg)); err != nil {
		s.d.Logger.Warn("inbox realtime notify failed", slog.Any("error", err))
	}
}

func (s *Service) render(ctx context.Context, n store.MessagingNotification, data map[string]any) (string, error) {
	if n.TemplateVersionID != nil {
		v, err := s.d.Q.MessagingGetTemplateVersion(ctx, *n.TemplateVersionID)
		if err != nil {
			return "", err
		}
		return renderBlocks(v.Blocks, data), nil
	}
	return Render(n.Template, n.Channel, data)
}

// renderBlocks turns block JSON into plain text with merge tags (fallback "there" for names).
func renderBlocks(blocks json.RawMessage, data map[string]any) string {
	var bs []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	_ = json.Unmarshal(blocks, &bs)
	out := ""
	for _, b := range bs {
		if b.Text != "" {
			out += b.Text + "\n"
		}
	}
	name, _ := data["firstName"].(string)
	if name == "" {
		name = "there"
	}
	link, _ := data["link"].(string)
	coupon, _ := data["coupon"].(string)
	return replaceAll(out, map[string]string{"{{first_name}}": name, "{{coupon}}": coupon, "{{link}}": link})
}

func replaceAll(s string, m map[string]string) string {
	for k, v := range m {
		for {
			i := indexOf(s, k)
			if i < 0 {
				break
			}
			s = s[:i] + v + s[i+len(k):]
		}
	}
	return s
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// Jobs registers the send worker and housekeeping.
func (s *Service) Jobs(reg *jobs.Registry) {
	river.AddWorker(reg.Workers, &SendWorker{s: s})
	river.AddWorker(reg.Workers, &TrimWorker{s: s})
	reg.Every(24*time.Hour, func() (river.JobArgs, *river.InsertOpts) { return TrimArgs{}, nil })
}

// TrimArgs trims message bodies older than 90 days.
type TrimArgs struct{}

// Kind identifies the job.
func (TrimArgs) Kind() string { return "notify.trim_bodies" }

// TrimWorker removes old bodies (retention: messaging-marketing.md §5).
type TrimWorker struct {
	river.WorkerDefaults[TrimArgs]
	s *Service
}

// Work runs the trim.
func (w *TrimWorker) Work(ctx context.Context, _ *river.Job[TrimArgs]) error {
	_, err := w.s.d.Q.MessagingTrimBodies(ctx)
	return err
}
