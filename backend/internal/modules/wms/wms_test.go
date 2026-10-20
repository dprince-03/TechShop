package wms

import "testing"

func TestFakeCarrierSignature(t *testing.T) {
	c := &FakeCarrier{secret: func(string) []byte { return []byte("k") }}
	body := []byte(`{"eventId":"1","trackingNumber":"FK1","status":"delivered"}`)
	sig := sign([]byte("k"), body)
	if !c.VerifySignature(body, sig) || c.VerifySignature(append(body, ' '), sig) || c.VerifySignature(body, "") {
		t.Fatal("carrier signature check is wrong")
	}
}
