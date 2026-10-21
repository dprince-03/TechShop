package payments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// paystackTransfersStandIn is a local stand-in for Paystack Transfers (/transferrecipient, /transfer, /transfer/verify),
// shapes from Paystack's Transfers guide. "dup-" references behave as if a first attempt timed
// out after Paystack accepted it: initiate refuses the duplicate, verify finds the transfer.
func paystackTransfersStandIn(t *testing.T) *Paystack {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk_test" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		switch {
		case r.Method == "POST" && r.URL.Path == "/transferrecipient":
			if body["type"] != "nuban" || body["currency"] != "NGN" || body["account_number"] != "0123456789" || body["bank_code"] != "058" {
				t.Errorf("recipient body: %v", body)
			}
			_, _ = w.Write([]byte(`{"status":true,"message":"Transfer recipient created successfully","data":{"recipient_code":"RCP_t0ya41mp35flk40"}}`))
		case r.Method == "POST" && r.URL.Path == "/transfer":
			ref, _ := body["reference"].(string)
			if body["source"] != "balance" || body["recipient"] != "RCP_t0ya41mp35flk40" || body["amount"] != float64(3_780_000) {
				t.Errorf("transfer body: %v", body)
			}
			switch {
			case strings.HasPrefix(ref, "dup-"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"status":false,"message":"Transfer reference already exists"}`))
			case strings.HasPrefix(ref, "poor-"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"status":false,"message":"Your balance is not enough to fulfil this request"}`))
			default:
				_, _ = w.Write([]byte(`{"status":true,"message":"Transfer has been queued","data":{"reference":"` + ref + `","status":"pending","transfer_code":"TRF_1ptvuv321ahaa7q"}}`))
			}
		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/transfer/verify/"):
			ref := strings.TrimPrefix(r.URL.Path, "/transfer/verify/")
			if strings.HasPrefix(ref, "dup-") {
				_, _ = w.Write([]byte(`{"status":true,"data":{"reference":"` + ref + `","status":"success","transfer_code":"TRF_dup"}}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"status":false,"message":"Transfer not found"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return &Paystack{secret: "sk_test", base: srv.URL, client: &http.Client{Timeout: 5 * time.Second}}
}

func TestPaystackTransfers(t *testing.T) {
	p, ctx := paystackTransfersStandIn(t), context.Background()
	code, err := p.CreateRecipient(ctx, "ADAEZE OKAFOR", "0123456789", "058")
	if err != nil || code != "RCP_t0ya41mp35flk40" {
		t.Fatalf("recipient: %q %v", code, err)
	}
	req := TransferRequest{AmountKobo: 3_780_000, Recipient: code, Reason: "TechShop payout"}

	req.Reference = "8f14e45f-ceea-467a-9575-3f6a9c1b7d21"
	if r, err := p.Initiate(ctx, req); err != nil || r.Status != "pending" || r.Code != "TRF_1ptvuv321ahaa7q" {
		t.Fatalf("queued transfer: %+v %v", r, err)
	}
	// A retry after a timeout: the duplicate is refused, but verify shows it already went.
	req.Reference = "dup-0d61f837-0cad-4f3a"
	if r, err := p.Initiate(ctx, req); err != nil || r.Status != "success" {
		t.Fatalf("duplicate reference should resolve through verify: %+v %v", r, err)
	}
	// Refused outright (and unknown to verify): failed, with the provider's reason.
	req.Reference = "poor-6512bd43-d9ca-4d"
	if r, err := p.Initiate(ctx, req); err != nil || r.Status != "failed" || !strings.Contains(r.Reason, "balance") {
		t.Fatalf("refused transfer: %+v %v", r, err)
	}
	if _, err := p.VerifyTransfer(ctx, "unknown-reference-1234"); err == nil {
		t.Fatal("verify of an unknown reference should fail")
	}
}

func TestTransferStatusMapping(t *testing.T) {
	for in, want := range map[string]string{"success": "success", "failed": "failed", "reversed": "reversed", "abandoned": "failed",
		"otp": "pending", "pending": "pending", "received": "pending"} {
		if got := transferResult(in, "", "").Status; got != want {
			t.Errorf("%s → %s, want %s", in, got, want)
		}
	}
}
