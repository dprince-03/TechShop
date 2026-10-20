package notify

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/pkg/validate"
)

// SMSProvider sends SMS. marketing=true uses the promotional route (DND numbers filtered);
// otherwise the transactional (DND-capable) route.
type SMSProvider interface {
	SendSMS(ctx context.Context, to, body string, marketing bool) (provider, messageID string, err error)
}

// EmailProvider sends email (transactional from mail.techshop.ng; marketing from news.techshop.ng).
type EmailProvider interface {
	SendEmail(ctx context.Context, to, subject, body string) (provider, messageID string, err error)
}

// PushProvider sends a push notification to one device token.
type PushProvider interface {
	SendPush(ctx context.Context, token, title, body string) (provider, messageID string, err error)
}

// LogSMS logs instead of sending (development; refused in production by config).
type LogSMS struct{ Logger *slog.Logger }

// SendSMS logs the message with a masked number.
func (l LogSMS) SendSMS(_ context.Context, to, body string, marketing bool) (string, string, error) {
	l.Logger.Info("sms (log provider)", slog.String("to", validate.MaskPhone(to)), slog.Bool("marketing", marketing), slog.Int("chars", len(body)))
	return "log", "log-" + uuid.NewString(), nil
}

// LogEmail logs instead of sending.
type LogEmail struct{ Logger *slog.Logger }

// SendEmail logs the message.
func (l LogEmail) SendEmail(_ context.Context, _ string, subject, _ string) (string, string, error) {
	l.Logger.Info("email (log provider)", slog.String("subject", subject))
	return "log", "log-" + uuid.NewString(), nil
}

// LogPush logs instead of sending.
type LogPush struct{ Logger *slog.Logger }

// SendPush logs the message.
func (l LogPush) SendPush(_ context.Context, _, title, _ string) (string, string, error) {
	l.Logger.Info("push (log provider)", slog.String("title", title))
	return "log", "log-" + uuid.NewString(), nil
}
