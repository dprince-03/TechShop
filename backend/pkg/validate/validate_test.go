package validate

import "testing"

// Shared vectors: the same cases are used by @techshop/api-client normaliseNigerianPhone.
func TestNormalisePhone(t *testing.T) {
	cases := map[string]string{
		"0803 123 4567":     "+2348031234567",
		"+234 803 123 4567": "+2348031234567",
		"2348031234567":     "+2348031234567",
		"8031234567":        "+2348031234567",
		"09012345678":       "+2349012345678",
		"0703-123-4567":     "+2347031234567",
		"12345":             "",
		"06031234567":       "",
		"+44 7700 900123":   "",
	}
	for in, want := range cases {
		if got := NormalisePhone(in); got != want {
			t.Errorf("NormalisePhone(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIMEI(t *testing.T) {
	if !IMEI("490154203237518") || !IMEI("356789104452212") {
		t.Fatal("valid IMEIs rejected")
	}
	if IMEI("356789104452211") || IMEI("12345") || IMEI("49015420323751a") {
		t.Fatal("invalid IMEI accepted")
	}
}

func TestVINAndState(t *testing.T) {
	if !VIN("1HGCM82633A004352") || VIN("1HGCM82633A00435O") {
		t.Fatal("VIN check wrong")
	}
	if !StateCode("LA") || StateCode("XX") {
		t.Fatal("state code check wrong")
	}
}

func TestNUBANLength(t *testing.T) {
	if NUBAN("058", "12345") {
		t.Fatal("short account accepted")
	}
}
