package payments

import (
	"context"
	"testing"
)

func TestFakeSignatureAndParse(t *testing.T) {
	f := NewFake("paystack", []byte("k"), "http://x", nil)
	body := []byte(`{"event":"charge.success","data":{"id":7,"reference":"TSPABC","status":"success","amount":1000,"currency":"NGN"}}`)
	sig := f.Sign(body)
	if !f.VerifySignature(body, sig) {
		t.Fatal("own signature must verify")
	}
	if f.VerifySignature(append(body, ' '), sig) {
		t.Fatal("a changed body must not verify")
	}
	ev, err := f.ParseWebhook(body)
	if err != nil || ev.Reference != "TSPABC" || ev.Type != "charge.success" || ev.ID == "" {
		t.Fatalf("parse = %+v, %v", ev, err)
	}
	if _, err := f.ParseWebhook([]byte(`{"event":""}`)); err == nil {
		t.Fatal("missing event must fail")
	}
}

func TestFakeTransferAccountIsNUBANShaped(t *testing.T) {
	f := NewFake("moniepoint", []byte("k"), "http://x", nil)
	r, _ := f.Initialize(context.Background(), InitRequest{Reference: "TSP1", Transfer: true})
	if len(r.AccountNumber) != 10 {
		t.Fatalf("account number %q", r.AccountNumber)
	}
}
