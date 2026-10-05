package crypto

import "testing"

func TestSealOpenBindsAAD(t *testing.T) {
	s, err := NewLocalSealer("dGVzdC1rZXktdGhhdC1pcy0zMi1ieXRlcy1sb25nISE=")
	if err != nil {
		t.Fatal(err)
	}
	ct, err := s.Seal([]byte("12345678901"), "sellers.seller_kyc_submissions.id_number:abc")
	if err != nil {
		t.Fatal(err)
	}
	pt, err := s.Open(ct, "sellers.seller_kyc_submissions.id_number:abc")
	if err != nil || string(pt) != "12345678901" {
		t.Fatalf("round trip failed: %v %q", err, pt)
	}
	if _, err := s.Open(ct, "sellers.seller_kyc_submissions.id_number:other-row"); err == nil {
		t.Fatal("ciphertext moved to another row must not decrypt")
	}
}

func TestRandomDigits(t *testing.T) {
	c := RandomDigits(6)
	if len(c) != 6 {
		t.Fatalf("got %q", c)
	}
	for _, r := range c {
		if r < '0' || r > '9' {
			t.Fatalf("non-digit in %q", c)
		}
	}
}
