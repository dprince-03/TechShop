package reviews

import (
	"regexp"
	"strings"
)

// Review integrity rules (trust-safety.md §6). A review that trips any of them is held for
// moderation instead of being published; the reasons are stored for moderators and never shown
// to the reviewer (so the rules can't be tuned against).

// maxReviewsPerDay is the velocity limit before reviews are held.
const maxReviewsPerDay = 5

var (
	// Starting list of incentivised-review phrases; extend as moderators see new patterns.
	incentivised = regexp.MustCompile(`(?i)\b(in exchange for (a|my|this) review|free product for|paid (me )?(to|for) (a )?review|discount for (a )?review|refund for (a )?review|review for (a )?(refund|discount|gift))\b`)
	contactInfo  = regexp.MustCompile(`(?i)(\+?234[\s-]?\d{3}[\s-]?\d{3}[\s-]?\d{4}|\b0[789][01]\d[\s-]?\d{3}[\s-]?\d{4}\b|[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}|https?://|www\.)`)
	spaces       = regexp.MustCompile(`\s+`)
)

// Signals are facts about the reviewer gathered from the database.
type Signals struct {
	ReviewsLast24h   int
	SameTextByOthers bool
}

// holdReasons lists why a review must wait for a moderator (empty = publish now).
func holdReasons(title, body string, sig Signals) []string {
	text := title + " " + body
	var out []string
	if sig.ReviewsLast24h >= maxReviewsPerDay {
		out = append(out, "velocity: many reviews from this account in 24 hours")
	}
	if sig.SameTextByOthers {
		out = append(out, "duplicate: the same text was posted by another account")
	}
	if incentivised.MatchString(text) {
		out = append(out, "incentivised: mentions a reward for the review")
	}
	if contactInfo.MatchString(text) {
		out = append(out, "contact details or links in the text")
	}
	return out
}

// normalise is how review bodies are compared for duplicates (must match the SQL in
// ReviewsSameTextByOthers: trimmed, whitespace collapsed, lower case).
func normalise(s string) string {
	return strings.ToLower(spaces.ReplaceAllString(strings.TrimSpace(s), " "))
}

// displayName shows reviewers as "Adaeze O.".
func displayName(first, lastInitial string) string {
	if lastInitial == "" {
		return first
	}
	return first + " " + strings.ToUpper(lastInitial) + "."
}
