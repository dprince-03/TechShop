package reviews

import "testing"

func TestHoldReasons(t *testing.T) {
	cases := []struct {
		name, title, body string
		sig               Signals
		held              bool
	}{
		{"ordinary", "Great phone", "Battery lasts all day and the camera is sharp.", Signals{}, false},
		{"velocity", "Ok", "Fine.", Signals{ReviewsLast24h: 5}, true},
		{"duplicate", "Ok", "Fine.", Signals{SameTextByOthers: true}, true},
		{"incentivised", "Good", "I got a discount for a review, nice seller", Signals{}, true},
		{"phone number", "Call me", "Buy from me cheaper: 0803 123 4567", Signals{}, true},
		{"international phone", "Call me", "WhatsApp +234 803 123 4567", Signals{}, true},
		{"link", "Cheaper", "see www.example.com", Signals{}, true},
		{"numbers that aren't phones", "128GB", "Got 2 for 450000 naira, 256 GB model", Signals{}, false},
	}
	for _, c := range cases {
		if got := len(holdReasons(c.title, c.body, c.sig)) > 0; got != c.held {
			t.Errorf("%s: held = %v, want %v (%v)", c.name, got, c.held, holdReasons(c.title, c.body, c.sig))
		}
	}
}

func TestNormaliseAndDisplayName(t *testing.T) {
	if normalise("  Great   PHONE\n really ") != "great phone really" {
		t.Fatal("normalise")
	}
	if displayName("Adaeze", "o") != "Adaeze O." || displayName("Musa", "") != "Musa" {
		t.Fatal("displayName")
	}
}
