package notify

import (
	"bytes"
	"fmt"
	"text/template"
)

// Transactional templates live in code and are reviewed like code (messaging-marketing.md §3.3).
// Marketing content comes from template_versions instead.
var codeTemplates = map[string]map[string]string{
	"otp": {
		"sms":   "Your TechShop code is {{.code}}. It expires in 10 minutes. Never share it — TechShop will never call to ask for it.",
		"email": "Your TechShop sign-in code is {{.code}}. It expires in 10 minutes.",
	},
	"new_device": {
		"push":  "New sign-in to your TechShop account from {{.device}}. Not you? Secure your account.",
		"email": "New sign-in to your TechShop account from {{.device}}. If this wasn't you, change your password and sign out other devices.",
	},
	"session_reuse": {
		"push":  "We signed out a device because an old sign-in token was reused. Check your recent activity.",
		"email": "We signed out a device because an old sign-in token was reused, which can mean it was stolen.",
	},
	"staff_invite":     {"email": "You've been invited to the TechShop Staff portal. Set your password and authenticator: {{.link}} (expires in 72 hours)."},
	"password_changed": {"email": "Your TechShop password was changed. All other devices were signed out."},
	"order_confirmed": {
		"push":  "Order {{.order}} is confirmed. We'll tell you when it ships.",
		"sms":   "TechShop: order {{.order}} is confirmed. Total {{.total}}.",
		"email": "Thanks {{.firstName}}! Order {{.order}} is confirmed. Total {{.total}}.",
	},
	"order_shipped": {
		"push": "{{.order}} is on its way.",
		"sms":  "TechShop: {{.order}} is on its way. Track it in the app.",
	},
	"delivery_code": {
		"sms":  "Your TechShop delivery code for {{.order}} is {{.code}}. Give it to the rider only when you have your parcel.",
		"push": "Your delivery code for {{.order}} is {{.code}}.",
	},
	"order_delivered":   {"push": "{{.order}} was delivered. Enjoy! How was it? Leave a review."},
	"order_cancelled":   {"push": "{{.order}} was cancelled: {{.reason}}. Any payment will be refunded.", "sms": "TechShop: {{.order}} was cancelled. Any payment will be refunded."},
	"refund_processed":  {"push": "We've refunded {{.amount}} for {{.order}}.", "email": "We've refunded {{.amount}} for {{.order}}. It can take 3-5 working days to reach you."},
	"payment_partial":   {"sms": "TechShop: we received {{.received}} for {{.order}}. Please send the remaining {{.remaining}} to the same account."},
	"seller_new_order":  {"push": "New order {{.order}}: accept it within 24 hours.", "email": "You have a new order {{.order}}. Accept it within 24 hours in Seller Centre."},
	"seller_kyc_result": {"email": "Your TechShop seller application was {{.result}}. {{.note}}"},
	"payout_sent":       {"email": "We've sent {{.amount}} to your bank account ending {{.last4}}.", "push": "Payout of {{.amount}} sent."},
	"payout_failed":     {"email": "We couldn't pay {{.amount}} to your bank account: {{.reason}}. Please update your bank details."},
	"quote_ready":       {"email": "Your quote {{.ref}} is ready. It's valid until {{.validUntil}}."},
	"invoice_due":       {"email": "Invoice {{.ref}} for {{.amount}} is due on {{.due}}."},
	"ticket_reply":      {"email": "There's a new reply on your TechShop ticket {{.ref}}.", "push": "New reply on ticket {{.ref}}."},
	"return_update":     {"push": "Return {{.ref}}: {{.status}}.", "email": "Your return {{.ref}} is now {{.status}}."},
	"rider_new_job":     {"push": "New delivery job {{.order}} in {{.zone}}."},
	"viewing_confirmed": {"sms": "TechShop: your viewing of {{.car}} is booked for {{.date}} ({{.slot}})."},
	"campaign":          {"email": "{{.body}}", "sms": "{{.body}} Reply STOP to opt out.", "push": "{{.body}}"},
	"journey":           {"email": "{{.body}}", "push": "{{.body}}", "sms": "{{.body}} Reply STOP to opt out."},
	"back_in_stock":     {"push": "{{.product}} is back in stock.", "email": "Good news: {{.product}} is back in stock."},
}

// Render fills a code template for a channel.
func Render(key, channel string, data map[string]any) (string, error) {
	byChannel, ok := codeTemplates[key]
	if !ok {
		return "", fmt.Errorf("notify: unknown template %q", key)
	}
	src, ok := byChannel[channel]
	if !ok {
		return "", fmt.Errorf("notify: template %q has no %s version", key, channel)
	}
	t, err := template.New(key).Option("missingkey=zero").Parse(src)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

// HasChannel reports whether a code template exists for a channel.
func HasChannel(key, channel string) bool { _, ok := codeTemplates[key][channel]; return ok }
