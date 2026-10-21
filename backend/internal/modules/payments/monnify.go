package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ---- Monnify (Moniepoint) live adapter: pay by bank transfer into a one-time account.
//
// Sources (docs/plan.md 2026-10-21): developers.monnify.com (API reference, webhooks; the site
// blocks automated reading, so shapes were cross-checked against henryejemuta/laravel-monnify,
// jiyamathias/monnify-go and ken-de-nigerian/payzephyr). Amounts are in naira with kobo as
// decimals (every SDK sends naira); Bearer tokens come from POST /api/v1/auth/login with Basic
// apiKey:secretKey; webhooks carry monnify-signature = hex HMAC-SHA512 of the raw body with the
// secret key. Re-confirm against the sandbox before go-live.

// Monnify is the live Monnify adapter.
type Monnify struct {
	apiKey, secret, contract, base string
	client                         *http.Client

	mu      sync.Mutex
	token   string
	expires time.Time
}

// NewMonnify creates the live adapter; base is https://sandbox.monnify.com or https://api.monnify.com.
func NewMonnify(apiKey, secret, contract, base string) *Monnify {
	return &Monnify{apiKey: apiKey, secret: secret, contract: contract, base: strings.TrimRight(base, "/"), client: &http.Client{Timeout: 20 * time.Second}}
}

// Name implements Provider. The provider is stored as "moniepoint" (Monnify's parent company).
func (m *Monnify) Name() string { return "moniepoint" }

// monnifyEnvelope is every Monnify response: {requestSuccessful, responseMessage, responseCode, responseBody}.
type monnifyEnvelope struct {
	RequestSuccessful bool            `json:"requestSuccessful"`
	ResponseMessage   string          `json:"responseMessage"`
	ResponseCode      string          `json:"responseCode"`
	ResponseBody      json.RawMessage `json:"responseBody"`
}

// accessToken logs in once and reuses the token until a minute before it expires.
func (m *Monnify) accessToken(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.token != "" && time.Now().Before(m.expires) {
		return m.token, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.base+"/api/v1/auth/login", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(m.apiKey+":"+m.secret)))
	var env monnifyEnvelope
	if err := m.do(req, &env); err != nil {
		return "", err
	}
	var body struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int64  `json:"expiresIn"`
	}
	if !env.RequestSuccessful || json.Unmarshal(env.ResponseBody, &body) != nil || body.AccessToken == "" {
		return "", fmt.Errorf("monnify login: %s", env.ResponseMessage)
	}
	m.token, m.expires = body.AccessToken, time.Now().Add(time.Duration(max(body.ExpiresIn-60, 30))*time.Second)
	return m.token, nil
}

func (m *Monnify) do(req *http.Request, out *monnifyEnvelope) error {
	res, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 500 {
		return fmt.Errorf("monnify %s: status %d", req.URL.Path, res.StatusCode)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("monnify %s: status %d, unreadable body", req.URL.Path, res.StatusCode)
	}
	return nil
}

// call sends an authenticated request and decodes responseBody into body (nil = ignore).
func (m *Monnify) call(ctx context.Context, method, path string, in any, body any) (monnifyEnvelope, error) {
	var env monnifyEnvelope
	tok, err := m.accessToken(ctx)
	if err != nil {
		return env, err
	}
	var rd io.Reader
	if in != nil {
		b, _ := json.Marshal(in)
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, m.base+path, rd)
	if err != nil {
		return env, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	if err := m.do(req, &env); err != nil {
		return env, err
	}
	if env.RequestSuccessful && body != nil && len(env.ResponseBody) > 0 {
		if err := json.Unmarshal(env.ResponseBody, body); err != nil {
			return env, fmt.Errorf("monnify %s: %w", path, err)
		}
	}
	return env, nil
}

// naira turns kobo into Monnify's decimal naira amount.
func naira(kobo int64) json.Number { return json.Number(fmt.Sprintf("%d.%02d", kobo/100, kobo%100)) }

// kobo turns Monnify's naira amount back into kobo.
func kobo(n float64) int64 { return int64(math.Round(n * 100)) }

// Initialize implements Provider: a transaction limited to bank transfer, then its one-time
// account (bank-transfer/init-payment). The stored reference is Monnify's transactionReference,
// which status, refunds and webhooks all use.
func (m *Monnify) Initialize(ctx context.Context, r InitRequest) (InitResult, error) {
	var tx struct {
		TransactionReference string `json:"transactionReference"`
		CheckoutURL          string `json:"checkoutUrl"`
	}
	in := map[string]any{"amount": naira(r.AmountKobo), "customerName": "TechShop customer", "customerEmail": r.Email, "paymentReference": r.Reference,
		"paymentDescription": "TechShop order", "currencyCode": "NGN", "contractCode": m.contract, "redirectUrl": r.CallbackURL,
		"paymentMethods": []string{"ACCOUNT_TRANSFER"}}
	env, err := m.call(ctx, http.MethodPost, "/api/v1/merchant/transactions/init-transaction", in, &tx)
	if err != nil {
		return InitResult{}, err
	}
	if !env.RequestSuccessful || tx.TransactionReference == "" {
		return InitResult{}, fmt.Errorf("monnify initialise: %s", env.ResponseMessage)
	}
	res := InitResult{CheckoutURL: tx.CheckoutURL, Reference: tx.TransactionReference}
	if !r.Transfer {
		return res, nil
	}
	var acct struct {
		AccountNumber string `json:"accountNumber"`
		BankName      string `json:"bankName"`
	}
	env, err = m.call(ctx, http.MethodPost, "/api/v1/merchant/bank-transfer/init-payment", map[string]any{"transactionReference": tx.TransactionReference}, &acct)
	if err != nil {
		return InitResult{}, err
	}
	if !env.RequestSuccessful || acct.AccountNumber == "" {
		return InitResult{}, fmt.Errorf("monnify transfer account: %s", env.ResponseMessage)
	}
	res.AccountNumber, res.BankName = acct.AccountNumber, acct.BankName
	return res, nil
}

// monnifyStatus maps paymentStatus onto ours. Partly and over-paid transfers count as
// "success" with the amount received; the payments service decides what that means.
func monnifyStatus(s string) string {
	switch strings.ToUpper(s) {
	case "PAID", "OVERPAID", "PARTIALLY_PAID":
		return "success"
	case "FAILED":
		return "failed"
	case "REVERSED":
		return "reversed"
	case "EXPIRED", "ABANDONED", "CANCELLED":
		return "abandoned"
	default:
		return "pending"
	}
}

// monnifyChannel maps paymentMethod onto our channel names.
func monnifyChannel(s string) string {
	switch strings.ToUpper(s) {
	case "ACCOUNT_TRANSFER":
		return "bank_transfer"
	case "CARD":
		return "card"
	case "USSD":
		return "ussd"
	}
	return strings.ToLower(s)
}

// monnifyTime reads paidOn, which Monnify writes in more than one layout (Lagos time).
func monnifyTime(s string) *time.Time {
	for _, layout := range []string{"02/01/2006 3:04:05 PM", "2006-01-02 15:04:05.0", "2006-01-02 15:04:05", time.RFC3339} {
		if t, err := time.ParseInLocation(layout, strings.TrimSpace(s), time.FixedZone("WAT", 3600)); err == nil {
			return &t
		}
	}
	return nil
}

// Verify implements Provider (GET /api/v2/merchant/transactions/query?transactionReference=).
func (m *Monnify) Verify(ctx context.Context, reference string) (VerifyResult, error) {
	var b struct {
		PaymentStatus    string  `json:"paymentStatus"`
		AmountPaid       float64 `json:"amountPaid"`
		SettlementAmount float64 `json:"settlementAmount"`
		PaidOn           string  `json:"paidOn"`
		PaymentMethod    string  `json:"paymentMethod"`
		Currency         string  `json:"currency"`
		CurrencyCode     string  `json:"currencyCode"`
	}
	env, err := m.call(ctx, http.MethodGet, "/api/v2/merchant/transactions/query?transactionReference="+url.QueryEscape(reference), nil, &b)
	if err != nil {
		return VerifyResult{}, err
	}
	if !env.RequestSuccessful {
		return VerifyResult{Status: "not_found"}, nil
	}
	cur := b.Currency
	if cur == "" {
		cur = b.CurrencyCode
	}
	v := VerifyResult{Status: monnifyStatus(b.PaymentStatus), AmountKobo: kobo(b.AmountPaid), Currency: cur, Channel: monnifyChannel(b.PaymentMethod)}
	if b.SettlementAmount > 0 && b.AmountPaid >= b.SettlementAmount {
		v.FeesKobo = kobo(b.AmountPaid) - kobo(b.SettlementAmount)
	}
	if b.PaidOn != "" {
		v.PaidAt = monnifyTime(b.PaidOn)
	}
	return v, nil
}

// monnifyRefundStatus maps refundStatus onto ours.
func monnifyRefundStatus(s string) string {
	switch strings.ToUpper(s) {
	case "COMPLETED", "SUCCESSFUL":
		return "processed"
	case "FAILED":
		return "failed"
	default: // PENDING, IN_PROGRESS
		return "pending"
	}
}

// Refund implements Provider (POST /api/v1/refunds/initiate-refund). Our refund ID is the
// refundReference, so a retried request can't refund twice.
func (m *Monnify) Refund(ctx context.Context, reference string, amountKobo int64, refundRef string) (RefundResult, error) {
	var b struct {
		RefundReference string `json:"refundReference"`
		RefundStatus    string `json:"refundStatus"`
	}
	env, err := m.call(ctx, http.MethodPost, "/api/v1/refunds/initiate-refund", map[string]any{"transactionReference": reference, "refundReference": refundRef,
		"refundAmount": naira(amountKobo), "refundReason": "TechShop refund", "customerNote": "TechShop refund"}, &b)
	if err != nil {
		return RefundResult{}, err
	}
	if !env.RequestSuccessful {
		return RefundResult{Status: "failed", Reason: env.ResponseMessage}, nil
	}
	return RefundResult{Reference: refundRef, Status: monnifyRefundStatus(b.RefundStatus)}, nil
}

// RefundStatus implements Provider (GET /api/v1/refunds/{refundReference}).
func (m *Monnify) RefundStatus(ctx context.Context, refundRef string) (RefundResult, error) {
	var b struct {
		RefundStatus string `json:"refundStatus"`
	}
	env, err := m.call(ctx, http.MethodGet, "/api/v1/refunds/"+url.PathEscape(refundRef), nil, &b)
	if err != nil {
		return RefundResult{}, err
	}
	if !env.RequestSuccessful {
		return RefundResult{}, errors.New("monnify refund status: " + env.ResponseMessage)
	}
	return RefundResult{Reference: refundRef, Status: monnifyRefundStatus(b.RefundStatus)}, nil
}

// SignatureHeader implements Provider.
func (m *Monnify) SignatureHeader() string { return "monnify-signature" }

// VerifySignature implements Provider: hex HMAC-SHA512 of the raw body with the secret key.
func (m *Monnify) VerifySignature(body []byte, signature string) bool {
	return signature != "" && hmac.Equal([]byte(hmacSHA512Hex([]byte(m.secret), body)), []byte(strings.ToLower(signature)))
}

// ParseWebhook implements Provider: {eventType, eventData{transactionReference, ...}}. Event
// types are mapped onto the shared names (charge.*, refund.*) the payments service acts on.
func (m *Monnify) ParseWebhook(body []byte) (WebhookEvent, error) {
	var e struct {
		EventType string `json:"eventType"`
		EventData struct {
			TransactionReference string `json:"transactionReference"`
			RefundReference      string `json:"refundReference"`
			PaidOn               string `json:"paidOn"`
			CompletedOn          string `json:"completedOn"`
		} `json:"eventData"`
	}
	if err := json.Unmarshal(body, &e); err != nil {
		return WebhookEvent{}, err
	}
	if e.EventType == "" || e.EventData.TransactionReference == "" {
		return WebhookEvent{}, errors.New("monnify webhook: missing eventType or transactionReference")
	}
	types := map[string]string{"SUCCESSFUL_TRANSACTION": "charge.success", "FAILED_TRANSACTION": "charge.failed",
		"SUCCESSFUL_REFUND": "refund.processed", "FAILED_REFUND": "refund.failed"}
	t, ok := types[e.EventType]
	if !ok {
		t = "monnify." + strings.ToLower(e.EventType)
	}
	id := strings.Join([]string{e.EventType, e.EventData.TransactionReference, e.EventData.RefundReference, e.EventData.PaidOn, e.EventData.CompletedOn}, ":")
	return WebhookEvent{ID: id, Type: t, Reference: e.EventData.TransactionReference}, nil
}
