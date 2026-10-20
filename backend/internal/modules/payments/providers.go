package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Provider is the port every payment provider adapter implements. Adapters never touch the
// database; the payments service owns all state.
type Provider interface {
	Name() string
	// Initialize opens a hosted checkout for an amount and returns where to send the customer.
	Initialize(ctx context.Context, r InitRequest) (InitResult, error)
	// Verify asks the provider for the truth about a reference (webhooks are never trusted alone).
	Verify(ctx context.Context, reference string) (VerifyResult, error)
	// Refund returns money for a captured reference.
	Refund(ctx context.Context, reference string, amountKobo int64) (RefundResult, error)
	// SignatureHeader names the webhook signature header; VerifySignature checks it over the raw body.
	SignatureHeader() string
	VerifySignature(body []byte, signature string) bool
	// ParseWebhook extracts the event identity and the reference it concerns.
	ParseWebhook(body []byte) (WebhookEvent, error)
}

// InitRequest describes a charge.
type InitRequest struct {
	Reference   string
	AmountKobo  int64
	Email       string
	Phone       string
	CallbackURL string
	Transfer    bool // offer a one-time transfer account (dynamic virtual account)
	ExpiresAt   time.Time
}

// InitResult is the provider's answer.
type InitResult struct {
	CheckoutURL string
	Reference   string
	// Transfer account details when the provider issued one.
	AccountNumber string
	BankName      string
}

// VerifyResult is the provider's view of a charge.
type VerifyResult struct {
	Status     string // success, failed, pending, abandoned, not_found
	AmountKobo int64
	Currency   string
	Channel    string
	FeesKobo   int64
	PaidAt     *time.Time
	Raw        string
}

// RefundResult is the provider's answer to a refund.
type RefundResult struct {
	Reference string
	Status    string // processed, pending, failed
	Reason    string
}

// WebhookEvent identifies one webhook.
type WebhookEvent struct {
	ID        string // provider event id (dedupe key)
	Type      string // e.g. charge.success, charge.failed, refund.processed
	Reference string
}

func hmacSHA512Hex(key, body []byte) string {
	m := hmac.New(sha512.New, key)
	m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}

// ---- Fake adapter (default; PAYMENTS_FAKE=true). It behaves like a provider: checkout URLs
// point at the dev payment page, webhooks are signed with a per-provider dev secret, and Verify
// answers from the last signed event the "provider" sent (read through lookup).

// FakeState returns the latest outcome the fake provider recorded for a reference.
type FakeState func(ctx context.Context, provider, reference string) (VerifyResult, error)

// Fake is a stand-in provider for local development and tests.
type Fake struct {
	name    string
	secret  []byte
	baseURL string
	state   FakeState
}

// NewFake creates a fake provider.
func NewFake(name string, secret []byte, baseURL string, state FakeState) *Fake {
	return &Fake{name: name, secret: secret, baseURL: baseURL, state: state}
}

// Name implements Provider.
func (f *Fake) Name() string { return f.name }

// Initialize implements Provider.
func (f *Fake) Initialize(_ context.Context, r InitRequest) (InitResult, error) {
	res := InitResult{CheckoutURL: f.baseURL + "/dev/pay/" + f.name + "/" + r.Reference, Reference: r.Reference}
	if r.Transfer {
		// Deterministic 10-digit NUBAN-shaped number from the reference (dev only).
		sum := sha512.Sum512([]byte(r.Reference))
		n := uint64(0)
		for _, b := range sum[:8] {
			n = n<<8 | uint64(b)
		}
		res.AccountNumber = fmt.Sprintf("99%08d", n%100000000)
		res.BankName = "Moniepoint MFB (test)"
	}
	return res, nil
}

// Verify implements Provider.
func (f *Fake) Verify(ctx context.Context, reference string) (VerifyResult, error) {
	return f.state(ctx, f.name, reference)
}

// Refund implements Provider: always processed immediately.
func (f *Fake) Refund(_ context.Context, reference string, _ int64) (RefundResult, error) {
	return RefundResult{Reference: "rf_" + reference + "_" + strconv.FormatInt(time.Now().UnixNano()%1e6, 10), Status: "processed"}, nil
}

// SignatureHeader implements Provider (the real header names; OPay's scheme is confirmed when its live adapter is built).
func (f *Fake) SignatureHeader() string {
	switch f.name {
	case "moniepoint":
		return "monnify-signature"
	case "opay":
		return "x-opay-signature"
	}
	return "x-paystack-signature"
}

// VerifySignature implements Provider (HMAC-SHA512 over the raw body, constant-time compare).
func (f *Fake) VerifySignature(body []byte, signature string) bool {
	return hmac.Equal([]byte(hmacSHA512Hex(f.secret, body)), []byte(signature))
}

// Sign produces the signature the fake provider sends with a webhook body.
func (f *Fake) Sign(body []byte) string { return hmacSHA512Hex(f.secret, body) }

// ParseWebhook implements Provider. Fake events use the Paystack shape for every provider.
func (f *Fake) ParseWebhook(body []byte) (WebhookEvent, error) { return parsePaystackShape(body) }

// paystackEvent is the Paystack webhook shape: {"event": "charge.success", "data": {...}}.
type paystackEvent struct {
	Event string `json:"event"`
	Data  struct {
		ID        json.Number `json:"id"`
		Reference string      `json:"reference"`
		Status    string      `json:"status"`
		Amount    int64       `json:"amount"`
		Currency  string      `json:"currency"`
		Channel   string      `json:"channel"`
		Fees      int64       `json:"fees"`
		PaidAt    *time.Time  `json:"paid_at"`
	} `json:"data"`
}

func parsePaystackShape(body []byte) (WebhookEvent, error) {
	var e paystackEvent
	if err := json.Unmarshal(body, &e); err != nil {
		return WebhookEvent{}, err
	}
	if e.Event == "" || e.Data.Reference == "" {
		return WebhookEvent{}, errors.New("webhook: missing event or reference")
	}
	// Paystack has no event id: event type + transaction id + status is unique per delivery.
	return WebhookEvent{ID: e.Event + ":" + e.Data.ID.String() + ":" + e.Data.Reference, Type: e.Event, Reference: e.Data.Reference}, nil
}

// ---- Paystack live adapter (https://paystack.com/docs/api). Used only when PAYMENTS_FAKE=false.
// Endpoints and the x-paystack-signature HMAC-SHA512 scheme follow Paystack's public API docs;
// they must be re-confirmed against the sandbox before go-live (logged in docs/log.md).

// Paystack is the live Paystack adapter.
type Paystack struct {
	secret string
	base   string
	client *http.Client
}

// NewPaystack creates the live adapter.
func NewPaystack(secret string) *Paystack {
	return &Paystack{secret: secret, base: "https://api.paystack.co", client: &http.Client{Timeout: 15 * time.Second}}
}

// Name implements Provider.
func (p *Paystack) Name() string { return "paystack" }

func (p *Paystack) call(ctx context.Context, method, path string, body any, out any) error {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, p.base+path, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.secret)
	req.Header.Set("Content-Type", "application/json")
	res, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 500 {
		return fmt.Errorf("paystack %s: status %d", path, res.StatusCode)
	}
	return json.Unmarshal(raw, out)
}

// Initialize implements Provider.
func (p *Paystack) Initialize(ctx context.Context, r InitRequest) (InitResult, error) {
	var out struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AuthorizationURL string `json:"authorization_url"`
			Reference        string `json:"reference"`
		} `json:"data"`
	}
	body := map[string]any{"email": r.Email, "amount": r.AmountKobo, "reference": r.Reference, "callback_url": r.CallbackURL, "currency": "NGN"}
	if err := p.call(ctx, http.MethodPost, "/transaction/initialize", body, &out); err != nil {
		return InitResult{}, err
	}
	if !out.Status {
		return InitResult{}, fmt.Errorf("paystack initialize: %s", out.Message)
	}
	return InitResult{CheckoutURL: out.Data.AuthorizationURL, Reference: out.Data.Reference}, nil
}

// Verify implements Provider.
func (p *Paystack) Verify(ctx context.Context, reference string) (VerifyResult, error) {
	var out struct {
		Status bool `json:"status"`
		Data   struct {
			Status   string     `json:"status"`
			Amount   int64      `json:"amount"`
			Currency string     `json:"currency"`
			Channel  string     `json:"channel"`
			Fees     int64      `json:"fees"`
			PaidAt   *time.Time `json:"paid_at"`
		} `json:"data"`
	}
	if err := p.call(ctx, http.MethodGet, "/transaction/verify/"+reference, nil, &out); err != nil {
		return VerifyResult{}, err
	}
	if !out.Status {
		return VerifyResult{Status: "not_found"}, nil
	}
	return VerifyResult{Status: out.Data.Status, AmountKobo: out.Data.Amount, Currency: out.Data.Currency, Channel: out.Data.Channel, FeesKobo: out.Data.Fees, PaidAt: out.Data.PaidAt}, nil
}

// Refund implements Provider.
func (p *Paystack) Refund(ctx context.Context, reference string, amountKobo int64) (RefundResult, error) {
	var out struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			ID     json.Number `json:"id"`
			Status string      `json:"status"`
		} `json:"data"`
	}
	if err := p.call(ctx, http.MethodPost, "/refund", map[string]any{"transaction": reference, "amount": amountKobo}, &out); err != nil {
		return RefundResult{}, err
	}
	if !out.Status {
		return RefundResult{Status: "failed", Reason: out.Message}, nil
	}
	st := "pending"
	if out.Data.Status == "processed" {
		st = "processed"
	}
	return RefundResult{Reference: out.Data.ID.String(), Status: st}, nil
}

// SignatureHeader implements Provider.
func (p *Paystack) SignatureHeader() string { return "x-paystack-signature" }

// VerifySignature implements Provider.
func (p *Paystack) VerifySignature(body []byte, signature string) bool {
	return hmac.Equal([]byte(hmacSHA512Hex([]byte(p.secret), body)), []byte(signature))
}

// ParseWebhook implements Provider.
func (p *Paystack) ParseWebhook(body []byte) (WebhookEvent, error) { return parsePaystackShape(body) }
