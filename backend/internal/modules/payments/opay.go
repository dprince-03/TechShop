package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha3"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strings"
	"time"
)

// ---- OPay Cashier (Express Checkout) live adapter.
//
// Sources (docs/plan.md 2026-10-21): doc.opaycheckout.com — payment authentication, cashier
// create/status, refund create/status, callback signature (v1). Amounts are integers in kobo.
// Cashier create authenticates with "Bearer {public key}"; every other call with "Bearer
// {hex HMAC-SHA512 of the exact JSON body, keyed with the secret key}"; both send MerchantId.
// Callbacks carry {payload, sha512, type}; sha512 is a hex HMAC-SHA3-512 (secret key) of a
// fixed string built from the payload (OPay's v1 callback-signature page). OPay's own pages
// disagree on that algorithm, which is why a callback only prompts a signed status query and
// the payment sweep re-checks anything a rejected callback missed. Re-confirm in the sandbox.

// OPay is the live OPay adapter.
type OPay struct {
	merchantID, publicKey, secret, base, callbackURL string
	client                                           *http.Client
}

// NewOPay creates the live adapter; base is https://testapi.opaycheckout.com or https://liveapi.opaycheckout.com.
func NewOPay(merchantID, publicKey, secret, base, callbackURL string) *OPay {
	return &OPay{merchantID: merchantID, publicKey: publicKey, secret: secret, base: strings.TrimRight(base, "/"), callbackURL: callbackURL,
		client: &http.Client{Timeout: 20 * time.Second}}
}

// Name implements Provider.
func (o *OPay) Name() string { return "opay" }

type opayEnvelope struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (e opayEnvelope) ok() bool { return e.Code == "00000" }

// sign is the request signature: hex HMAC-SHA512 of the body bytes as sent.
func (o *OPay) sign(body []byte) string {
	m := hmac.New(sha512.New, []byte(o.secret))
	m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}

// call posts a JSON body. The body is a map, so encoding/json writes its keys in sorted order,
// as OPay's signature examples expect; the signed bytes are exactly the bytes sent.
func (o *OPay) call(ctx context.Context, path string, in map[string]any, publicKey bool, out any) (opayEnvelope, error) {
	var env opayEnvelope
	body, err := json.Marshal(in)
	if err != nil {
		return env, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.base+path, bytes.NewReader(body))
	if err != nil {
		return env, err
	}
	auth := o.sign(body)
	if publicKey {
		auth = o.publicKey
	}
	req.Header.Set("Authorization", "Bearer "+auth)
	req.Header.Set("MerchantId", o.merchantID)
	req.Header.Set("Content-Type", "application/json")
	res, err := o.client.Do(req)
	if err != nil {
		return env, err
	}
	defer func() { _ = res.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 500 {
		return env, fmt.Errorf("opay %s: status %d", path, res.StatusCode)
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return env, fmt.Errorf("opay %s: status %d, unreadable body", path, res.StatusCode)
	}
	if env.ok() && out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return env, fmt.Errorf("opay %s: %w", path, err)
		}
	}
	return env, nil
}

type opayAmount struct {
	Total    int64  `json:"total"`
	Currency string `json:"currency"`
}

// Initialize implements Provider (POST /api/v1/international/cashier/create): OPay's hosted
// checkout, where the customer picks card, OPay wallet, transfer or USSD.
func (o *OPay) Initialize(ctx context.Context, r InitRequest) (InitResult, error) {
	minutes := 30
	if !r.ExpiresAt.IsZero() {
		minutes = max(5, int(time.Until(r.ExpiresAt).Minutes()))
	}
	in := map[string]any{"country": "NG", "reference": r.Reference, "amount": opayAmount{Total: r.AmountKobo, Currency: "NGN"},
		"product":   map[string]string{"name": "TechShop order", "description": "TechShop order " + r.Reference},
		"returnUrl": r.CallbackURL, "cancelUrl": r.CallbackURL, "callbackUrl": o.callbackURL, "expireAt": minutes,
		"userInfo": map[string]string{"userEmail": r.Email, "userMobile": r.Phone}}
	var d struct {
		Reference  string `json:"reference"`
		CashierURL string `json:"cashierUrl"`
	}
	env, err := o.call(ctx, "/api/v1/international/cashier/create", in, true, &d)
	if err != nil {
		return InitResult{}, err
	}
	if !env.ok() || d.CashierURL == "" {
		return InitResult{}, fmt.Errorf("opay initialise: %s %s", env.Code, env.Message)
	}
	return InitResult{CheckoutURL: d.CashierURL, Reference: r.Reference}, nil
}

func opayStatus(s string) string {
	switch strings.ToUpper(s) {
	case "SUCCESS":
		return "success"
	case "FAIL":
		return "failed"
	case "CLOSE":
		return "abandoned"
	default: // INITIAL, PENDING
		return "pending"
	}
}

// Verify implements Provider (POST /api/v1/international/cashier/status, signed).
func (o *OPay) Verify(ctx context.Context, reference string) (VerifyResult, error) {
	var d struct {
		Status string     `json:"status"`
		Amount opayAmount `json:"amount"`
	}
	env, err := o.call(ctx, "/api/v1/international/cashier/status", map[string]any{"country": "NG", "reference": reference}, false, &d)
	if err != nil {
		return VerifyResult{}, err
	}
	if !env.ok() {
		return VerifyResult{Status: "not_found"}, nil
	}
	return VerifyResult{Status: opayStatus(d.Status), AmountKobo: d.Amount.Total, Currency: d.Amount.Currency}, nil
}

func opayRefundStatus(s string) string {
	switch strings.ToUpper(s) {
	case "SUCCESS":
		return "processed"
	case "FAIL", "CLOSE":
		return "failed"
	default:
		return "pending"
	}
}

// Refund implements Provider (POST /api/v1/international/payment/refund/create, signed) to
// the original method. OPay sends card, transfer and USSD refunds to a bank account instead;
// it refuses those here and finance refunds them as store credit or by bank transfer.
func (o *OPay) Refund(ctx context.Context, reference string, amountKobo int64, refundRef string) (RefundResult, error) {
	var d struct {
		Status string `json:"status"`
	}
	env, err := o.call(ctx, "/api/v1/international/payment/refund/create", map[string]any{"country": "NG", "reference": refundRef, "originalReference": reference,
		"amount": opayAmount{Total: amountKobo, Currency: "NGN"}, "refundWay": "Original", "refundReason": "TechShop refund", "callbackUrl": o.callbackURL}, false, &d)
	if err != nil {
		return RefundResult{}, err
	}
	if !env.ok() {
		return RefundResult{Status: "failed", Reason: "OPay: " + env.Message}, nil
	}
	return RefundResult{Reference: refundRef, Status: opayRefundStatus(d.Status)}, nil
}

// RefundStatus implements Provider (POST /api/v1/international/payment/refund/status, signed).
func (o *OPay) RefundStatus(ctx context.Context, refundRef string) (RefundResult, error) {
	var d struct {
		Status string `json:"status"`
	}
	env, err := o.call(ctx, "/api/v1/international/payment/refund/status", map[string]any{"country": "NG", "reference": refundRef}, false, &d)
	if err != nil {
		return RefundResult{}, err
	}
	if !env.ok() {
		return RefundResult{}, errors.New("opay refund status: " + env.Message)
	}
	return RefundResult{Reference: refundRef, Status: opayRefundStatus(d.Status)}, nil
}

// SignatureHeader implements Provider: OPay signs inside the body (the sha512 field).
func (o *OPay) SignatureHeader() string { return "" }

// opayCallback is the Cashier transaction-status callback.
type opayCallback struct {
	Payload struct {
		Amount        string `json:"amount"`
		Currency      string `json:"currency"`
		Reference     string `json:"reference"`
		Refunded      bool   `json:"refunded"`
		Status        string `json:"status"`
		Timestamp     string `json:"timestamp"`
		Token         string `json:"token"`
		TransactionID string `json:"transactionId"`
		UpdatedAt     string `json:"updated_at"`
	} `json:"payload"`
	SHA512 string `json:"sha512"`
	Type   string `json:"type"`
}

// opayCallbackString is the string OPay's v1 callback-signature page signs.
func opayCallbackString(c opayCallback) string {
	p := c.Payload
	refunded := "f"
	if p.Refunded {
		refunded = "t"
	}
	return fmt.Sprintf(`{Amount:"%s",Currency:"%s",Reference:"%s",Refunded:%s,Status:"%s",Timestamp:"%s",Token:"%s",TransactionID:"%s"}`,
		p.Amount, p.Currency, p.Reference, refunded, p.Status, p.Timestamp, p.Token, p.TransactionID)
}

func hmacSHA3512Hex(key []byte, msg string) string {
	m := hmac.New(func() hash.Hash { return sha3.New512() }, key)
	m.Write([]byte(msg))
	return hex.EncodeToString(m.Sum(nil))
}

// VerifySignature implements Provider: the body's sha512 field against HMAC-SHA3-512 of the
// callback string (the header value is unused).
func (o *OPay) VerifySignature(body []byte, _ string) bool {
	var c opayCallback
	if json.Unmarshal(body, &c) != nil || c.SHA512 == "" {
		return false
	}
	return hmac.Equal([]byte(hmacSHA3512Hex([]byte(o.secret), opayCallbackString(c))), []byte(strings.ToLower(c.SHA512)))
}

// ParseWebhook implements Provider.
func (o *OPay) ParseWebhook(body []byte) (WebhookEvent, error) {
	var c opayCallback
	if err := json.Unmarshal(body, &c); err != nil {
		return WebhookEvent{}, err
	}
	if c.Payload.Reference == "" {
		return WebhookEvent{}, errors.New("opay callback: missing reference")
	}
	t := "charge." + strings.ToLower(c.Payload.Status)
	if c.Payload.Refunded || strings.Contains(strings.ToLower(c.Type), "refund") {
		t = "refund." + strings.ToLower(c.Payload.Status)
	}
	id := strings.Join([]string{"opay", c.Type, c.Payload.TransactionID, c.Payload.Status, c.Payload.UpdatedAt}, ":")
	return WebhookEvent{ID: id, Type: t, Reference: c.Payload.Reference}, nil
}
