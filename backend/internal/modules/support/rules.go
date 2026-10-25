package support

import (
	"strings"
	"time"
)

// Ticket rules (docs/plan.md 2026-10-23). Pure, so they're unit-tested without a database.

// Topics a ticket can be about. Payment and refund problems mean someone's money is stuck, so
// they start as high priority.
var Topics = []string{"order", "delivery", "payment", "refund", "return", "warranty", "trade_in", "account", "selling", "business", "other"}

var highTopics = map[string]bool{"payment": true, "refund": true}

// priorityFor is a new ticket's starting priority.
func priorityFor(topic string) string {
	if highTopics[topic] {
		return "high"
	}
	return "normal"
}

// firstReplyWithin is the reply target: 4 hours for high or urgent, 24 hours otherwise
// (ecommerce email benchmarks average 12–24 h; hardware sellers reply in under 2 h).
func firstReplyWithin(priority string) time.Duration {
	if priority == "high" || priority == "urgent" {
		return 4 * time.Hour
	}
	return 24 * time.Hour
}

// slaDue is when the first reply is due.
func slaDue(priority string, from time.Time) time.Time { return from.Add(firstReplyWithin(priority)) }

// honeypotTripped reports a filled hidden field: people never see it, bots fill every field.
func honeypotTripped(v string) bool { return strings.TrimSpace(v) != "" }

// cleanBody trims a message and refuses empty ones.
func cleanBody(s string) (string, bool) {
	s = strings.TrimSpace(s)
	return s, s != ""
}

// excerpt shortens a reply for an SMS.
func excerpt(s string, n int) string {
	r := []rune(strings.Join(strings.Fields(s), " "))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n-1]) + "…"
}
