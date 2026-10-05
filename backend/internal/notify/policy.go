package notify

import (
	"context"
	"time"

	"github.com/dprince-03/techshop/backend/internal/store"
)

var lagos = time.FixedZone("WAT", 3600) // Nigeria has no DST: UTC+1 all year

// marketingCaps: per person per day (messaging-marketing.md §3.1; configurable later).
var marketingCaps = map[string]int32{"sms": 1, "email": 1, "push": 2, "whatsapp": 1}

// policy decides whether a queued message may be sent now. It returns a skip reason, or
// "quiet_hours" with the time to retry.
func (s *Service) policy(ctx context.Context, n store.MessagingNotification, to string) (string, *time.Time) {
	q := s.d.Q
	if to != "" && n.Channel != "in_app" {
		if sup, err := q.MessagingIsSuppressed(ctx, store.MessagingIsSuppressedParams{Channel: n.Channel, AddressHash: s.AddressHash(to)}); err == nil && sup {
			return "suppressed", nil
		}
	}
	if n.Priority != "marketing" || n.UserID == nil {
		return "", nil // security, delivery and transactional messages always go
	}
	// Consent per channel (NDPA): marketing_sms, marketing_email, marketing_push.
	granted, err := q.IdentityHasConsent(ctx, store.IdentityHasConsentParams{UserID: n.UserID, Purpose: "marketing_" + n.Channel})
	if err != nil || !granted {
		return "no_consent", nil
	}
	prefs, _ := q.MessagingGetPreferences(ctx, *n.UserID)
	for _, p := range prefs {
		if p.Category == n.Category && p.Channel == n.Channel && !p.Enabled {
			return "preference", nil
		}
	}
	dayStart := time.Now().In(lagos).Truncate(24 * time.Hour)
	if count, err := q.MessagingCountMarketingToday(ctx, store.MessagingCountMarketingTodayParams{UserID: n.UserID, Channel: n.Channel, SentAt: &dayStart}); err == nil && count >= marketingCaps[n.Channel] {
		return "capped", nil
	}
	start, end := s.quietFrom, s.quietTo
	enabled := true
	if st, err := q.MessagingGetSettings(ctx, *n.UserID); err == nil {
		enabled = st.QuietHours
		start, end = clock(st.QuietStart.Microseconds), clock(st.QuietEnd.Microseconds)
	}
	if enabled {
		if until := quietUntil(time.Now().In(lagos), start, end); until != nil {
			return "quiet_hours", until
		}
	}
	return "", nil
}

func clock(micros int64) string {
	d := time.Duration(micros) * time.Microsecond
	return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(d).Format("15:04")
}

// quietUntil returns when quiet hours end if now is inside them (handles windows over midnight).
func quietUntil(now time.Time, start, end string) *time.Time {
	s, err1 := time.Parse("15:04", start)
	e, err2 := time.Parse("15:04", end)
	if err1 != nil || err2 != nil {
		return nil
	}
	mins := now.Hour()*60 + now.Minute()
	sm, em := s.Hour()*60+s.Minute(), e.Hour()*60+e.Minute()
	inside := (sm > em && (mins >= sm || mins < em)) || (sm < em && mins >= sm && mins < em)
	if !inside {
		return nil
	}
	endToday := time.Date(now.Year(), now.Month(), now.Day(), e.Hour(), e.Minute(), 0, 0, now.Location())
	if !endToday.After(now) {
		endToday = endToday.Add(24 * time.Hour)
	}
	return &endToday
}
