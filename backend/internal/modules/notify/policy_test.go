package notify

import (
	"testing"
	"time"
)

func TestQuietHoursOverMidnight(t *testing.T) {
	at := func(h, m int) time.Time { return time.Date(2026, 10, 5, h, m, 0, 0, lagos) }
	if quietUntil(at(14, 5), "21:00", "08:00") != nil {
		t.Fatal("14:05 is not quiet")
	}
	u := quietUntil(at(22, 30), "21:00", "08:00")
	if u == nil || u.Hour() != 8 || u.Day() != 6 {
		t.Fatalf("22:30 should be held until 08:00 next day, got %v", u)
	}
	u = quietUntil(at(6, 0), "21:00", "08:00")
	if u == nil || u.Hour() != 8 || u.Day() != 5 {
		t.Fatalf("06:00 should be held until 08:00 same day, got %v", u)
	}
}

func TestRenderTemplates(t *testing.T) {
	s, err := Render("otp", "sms", map[string]any{"code": "530218"})
	if err != nil || s == "" || !contains(s, "530218") {
		t.Fatalf("otp sms: %q %v", s, err)
	}
	if _, err := Render("otp", "push", nil); err == nil {
		t.Fatal("missing channel must error")
	}
}

func contains(s, sub string) bool { return indexOf(s, sub) >= 0 }
