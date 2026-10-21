package payments

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// monnifyStandIn is a local stand-in for Monnify (shapes from the API reference as used by the
// Laravel, Go and PayZephyr integrations): login with Basic auth, then Bearer calls.
func monnifyStandIn(t *testing.T) (*Monnify, *atomic.Int32) {
	t.Helper()
	logins := &atomic.Int32{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok := func(body string) {
			_, _ = w.Write([]byte(`{"requestSuccessful":true,"responseMessage":"success","responseCode":"0","responseBody":` + body + `}`))
		}
		if r.URL.Path == "/api/v1/auth/login" {
			if r.Header.Get("Authorization") != "Basic "+base64.StdEncoding.EncodeToString([]byte("MK_TEST_KEY:secret-key")) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			logins.Add(1)
			ok(`{"accessToken":"jwt-1","expiresIn":3599}`)
			return
		}
		if r.Header.Get("Authorization") != "Bearer jwt-1" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		switch r.URL.Path {
		case "/api/v1/merchant/transactions/init-transaction":
			// Naira with kobo as decimals; transfer only; our contract.
			if in["amount"] != 1250.5 || in["contractCode"] != "1234567890" || in["paymentReference"] != "ts_pay_1" {
				t.Errorf("init body: %v", in)
			}
			ok(`{"transactionReference":"MNFY|20|20261021|000001","paymentReference":"ts_pay_1","checkoutUrl":"https://sandbox.sdk.monnify.com/checkout/MNFY|20|20261021|000001"}`)
		case "/api/v1/merchant/bank-transfer/init-payment":
			if in["transactionReference"] != "MNFY|20|20261021|000001" {
				t.Errorf("init-payment body: %v", in)
			}
			ok(`{"accountNumber":"5000123456","accountName":"TechShop Checkout","bankName":"Wema bank","accountDurationSeconds":2400}`)
		case "/api/v2/merchant/transactions/query":
			if r.URL.Query().Get("transactionReference") != "MNFY|20|20261021|000001" {
				_, _ = w.Write([]byte(`{"requestSuccessful":false,"responseMessage":"Could not find transaction","responseCode":"99"}`))
				return
			}
			ok(`{"paymentStatus":"OVERPAID","amountPaid":1300.00,"settlementAmount":1286.00,"paidOn":"21/10/2026 5:14:58 PM","paymentMethod":"ACCOUNT_TRANSFER","currency":"NGN"}`)
		case "/api/v1/refunds/initiate-refund":
			if in["refundReference"] != "refund-uuid-1" || in["refundAmount"] != 49.5 {
				t.Errorf("refund body: %v", in)
			}
			ok(`{"refundReference":"refund-uuid-1","refundStatus":"PENDING"}`)
		case "/api/v1/refunds/refund-uuid-1":
			ok(`{"refundReference":"refund-uuid-1","refundStatus":"COMPLETED"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return NewMonnify("MK_TEST_KEY", "secret-key", "1234567890", srv.URL), logins
}

func TestMonnifyCollectVerifyRefund(t *testing.T) {
	m, logins := monnifyStandIn(t)
	ctx := context.Background()
	res, err := m.Initialize(ctx, InitRequest{Reference: "ts_pay_1", AmountKobo: 125_050, Email: "a@b.ng", Transfer: true, CallbackURL: "http://x"})
	if err != nil || res.Reference != "MNFY|20|20261021|000001" || res.AccountNumber != "5000123456" || res.BankName != "Wema bank" {
		t.Fatalf("initialise: %+v %v", res, err)
	}
	v, err := m.Verify(ctx, res.Reference)
	if err != nil || v.Status != "success" || v.AmountKobo != 130_000 || v.FeesKobo != 1_400 || v.Channel != "bank_transfer" || v.PaidAt == nil {
		t.Fatalf("verify: %+v %v", v, err)
	}
	if want := time.Date(2026, 10, 21, 16, 14, 58, 0, time.UTC); !v.PaidAt.Equal(want) {
		t.Fatalf("paidOn is Lagos time: %v", v.PaidAt)
	}
	if v, _ := m.Verify(ctx, "MNFY|missing"); v.Status != "not_found" {
		t.Fatalf("unknown reference: %+v", v)
	}
	rf, err := m.Refund(ctx, res.Reference, 4_950, "refund-uuid-1")
	if err != nil || rf.Status != "pending" || rf.Reference != "refund-uuid-1" {
		t.Fatalf("refund: %+v %v", rf, err)
	}
	if st, err := m.RefundStatus(ctx, "refund-uuid-1"); err != nil || st.Status != "processed" {
		t.Fatalf("refund status: %+v %v", st, err)
	}
	if n := logins.Load(); n != 1 {
		t.Fatalf("the access token should be reused, logged in %d times", n)
	}
}

func TestMonnifyWebhook(t *testing.T) {
	m := NewMonnify("k", "secret-key", "c", "http://x")
	body := []byte(`{"eventType":"SUCCESSFUL_TRANSACTION","eventData":{"transactionReference":"MNFY|20|20261021|000001","paymentReference":"ts_pay_1","paidOn":"21/10/2026 5:14:58 PM","amountPaid":1250.50,"paymentStatus":"PAID"}}`)
	sig := hmacSHA512Hex([]byte("secret-key"), body)
	if !m.VerifySignature(body, sig) || !m.VerifySignature(body, strings.ToUpper(sig)) {
		t.Fatal("valid signature refused")
	}
	if m.VerifySignature(body, "") || m.VerifySignature(append(body, ' '), sig) {
		t.Fatal("missing or mismatched signature accepted")
	}
	ev, err := m.ParseWebhook(body)
	if err != nil || ev.Type != "charge.success" || ev.Reference != "MNFY|20|20261021|000001" || ev.ID == "" {
		t.Fatalf("parse: %+v %v", ev, err)
	}
	ref, _ := m.ParseWebhook([]byte(`{"eventType":"SUCCESSFUL_REFUND","eventData":{"transactionReference":"MNFY|1","refundReference":"refund-uuid-1"}}`))
	if ref.Type != "refund.processed" || ref.ID == ev.ID {
		t.Fatalf("refund event: %+v", ref)
	}
}

func TestMonnifyStatusMapping(t *testing.T) {
	for in, want := range map[string]string{"PAID": "success", "PARTIALLY_PAID": "success", "OVERPAID": "success", "PENDING": "pending",
		"EXPIRED": "abandoned", "CANCELLED": "abandoned", "FAILED": "failed", "REVERSED": "reversed"} {
		if got := monnifyStatus(in); got != want {
			t.Errorf("%s → %s, want %s", in, got, want)
		}
	}
	if naira(125_050) != "1250.50" || naira(7) != "0.07" || kobo(1250.5) != 125_050 {
		t.Fatal("naira/kobo conversion")
	}
}
