package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha3"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"hash"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// opayStandIn is a local stand-in for OPay Cashier (shapes from doc.opaycheckout.com): create
// takes the public key; status and refunds must carry the HMAC-SHA512 of the exact body.
func opayStandIn(t *testing.T) *OPay {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if r.Header.Get("MerchantId") != "256612345678901" {
			t.Errorf("MerchantId header missing on %s", r.URL.Path)
		}
		auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		signed := func() bool {
			m := hmac.New(sha512.New, []byte("OPAYPRV_secret"))
			m.Write(raw)
			return auth == hex.EncodeToString(m.Sum(nil))
		}
		var in map[string]any
		_ = json.Unmarshal(raw, &in)
		ok := func(data string) {
			_, _ = w.Write([]byte(`{"code":"00000","message":"SUCCESSFUL","data":` + data + `}`))
		}
		switch r.URL.Path {
		case "/api/v1/international/cashier/create":
			amt, _ := in["amount"].(map[string]any)
			if auth != "OPAYPUB_key" || in["country"] != "NG" || amt["total"] != float64(125_050) || in["callbackUrl"] != "https://api.example/api/v1/webhooks/opay" {
				t.Errorf("create: auth %q body %v", auth, in)
			}
			ok(`{"reference":"ts_pay_2","orderNo":"211009140896553163","cashierUrl":"https://sandboxcashier.opaycheckout.com/x","status":"INITIAL"}`)
		case "/api/v1/international/cashier/status":
			if !signed() {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if in["reference"] != "ts_pay_2" {
				_, _ = w.Write([]byte(`{"code":"02006","message":"order not found"}`))
				return
			}
			ok(`{"reference":"ts_pay_2","status":"SUCCESS","amount":{"total":125050,"currency":"NGN"}}`)
		case "/api/v1/international/payment/refund/create":
			if !signed() || in["originalReference"] != "ts_pay_2" || in["reference"] != "refund-uuid-2" || in["refundWay"] != "Original" {
				t.Errorf("refund: %v", in)
			}
			ok(`{"reference":"refund-uuid-2","status":"PENDING"}`)
		case "/api/v1/international/payment/refund/status":
			if !signed() {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ok(`{"reference":"refund-uuid-2","status":"SUCCESS"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return NewOPay("256612345678901", "OPAYPUB_key", "OPAYPRV_secret", srv.URL, "https://api.example/api/v1/webhooks/opay")
}

func TestOPayCheckoutVerifyRefund(t *testing.T) {
	o, ctx := opayStandIn(t), context.Background()
	res, err := o.Initialize(ctx, InitRequest{Reference: "ts_pay_2", AmountKobo: 125_050, Email: "a@b.ng", CallbackURL: "http://x"})
	if err != nil || res.CheckoutURL == "" || res.Reference != "ts_pay_2" {
		t.Fatalf("initialise: %+v %v", res, err)
	}
	if v, err := o.Verify(ctx, "ts_pay_2"); err != nil || v.Status != "success" || v.AmountKobo != 125_050 || v.Currency != "NGN" {
		t.Fatalf("verify: %+v %v", v, err)
	}
	if v, _ := o.Verify(ctx, "nope"); v.Status != "not_found" {
		t.Fatalf("unknown reference: %+v", v)
	}
	if rf, err := o.Refund(ctx, "ts_pay_2", 4_950, "refund-uuid-2"); err != nil || rf.Status != "pending" || rf.Reference != "refund-uuid-2" {
		t.Fatalf("refund: %+v %v", rf, err)
	}
	if st, err := o.RefundStatus(ctx, "refund-uuid-2"); err != nil || st.Status != "processed" {
		t.Fatalf("refund status: %+v %v", st, err)
	}
}

func TestOPayCallbackSignature(t *testing.T) {
	o := NewOPay("m", "pub", "OPAYPRV_secret", "http://x", "")
	payload := `{"amount":"125050","channel":"Web","country":"NG","currency":"NGN","displayedFailure":"","fee":"737","feeCurrency":"NGN","instrumentType":"BankCard","reference":"ts_pay_2","refunded":false,"status":"SUCCESS","timestamp":"2026-10-21T16:20:46Z","token":"220507145660712931829","transactionId":"220507145660712931829","updated_at":"2026-10-21T16:20:46Z"}`
	// The string OPay's callback-signature page signs, written out by hand.
	sign := `{Amount:"125050",Currency:"NGN",Reference:"ts_pay_2",Refunded:f,Status:"SUCCESS",Timestamp:"2026-10-21T16:20:46Z",Token:"220507145660712931829",TransactionID:"220507145660712931829"}`
	m := hmac.New(func() hash.Hash { return sha3.New512() }, []byte("OPAYPRV_secret"))
	m.Write([]byte(sign))
	sig := hex.EncodeToString(m.Sum(nil))
	body := []byte(`{"payload":` + payload + `,"sha512":"` + sig + `","type":"transaction-status"}`)
	if !o.VerifySignature(body, "") {
		t.Fatal("valid callback refused")
	}
	tampered := strings.Replace(string(body), `"amount":"125050"`, `"amount":"999999"`, 1)
	if o.VerifySignature([]byte(tampered), "") || o.VerifySignature([]byte(`{"payload":`+payload+`,"type":"transaction-status"}`), "") {
		t.Fatal("tampered or unsigned callback accepted")
	}
	ev, err := o.ParseWebhook(body)
	if err != nil || ev.Type != "charge.success" || ev.Reference != "ts_pay_2" {
		t.Fatalf("parse: %+v %v", ev, err)
	}
}
