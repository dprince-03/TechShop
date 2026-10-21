package sellers

import "testing"

func TestValidIDNumber(t *testing.T) {
	cases := []struct {
		idType, number string
		ok             bool
	}{
		{"nin", "12345678901", true},
		{"nin", "1234567890", false},  // 10 digits
		{"nin", "1234567890A", false}, // letters
		{"passport", "A12345678", true},
		{"drivers_licence", "ABC12", false}, // too short
	}
	for _, c := range cases {
		if got := validIDNumber(c.idType, c.number); got != c.ok {
			t.Errorf("validIDNumber(%q, %q) = %v, want %v", c.idType, c.number, got, c.ok)
		}
	}
	if normaliseID(" 1234-5678 901 ") != "12345678901" {
		t.Fatal("normaliseID should strip spaces and dashes")
	}
}

func TestSlugify(t *testing.T) {
	if got := slugify("  Ikeja Gadget Hub!! "); got != "ikeja-gadget-hub" {
		t.Fatalf("slugify = %q", got)
	}
	if got := slugify("★★★"); got != "" {
		t.Fatalf("slugify of symbols = %q, want empty", got)
	}
}

func TestNameMatch(t *testing.T) {
	cases := []struct{ bank, expected, want string }{
		{"OKAFOR ADAEZE CHIOMA", "Adaeze Okafor", NameMatch},               // reordered + middle name
		{"IKEJA GADGET HUB LTD", "Ikeja Gadget Hub", NameMatch},            // company suffix ignored
		{"ADAEZE NGOZI OKAFOR", "Adaeze Chioma Okafor-Bello", NamePartial}, // two shared words
		{"BELLO MUSA", "Adaeze Okafor", NameMismatch},
		{"", "Adaeze Okafor", NameUnknown},
	}
	for _, c := range cases {
		if got := nameMatch(c.bank, c.expected); got != c.want {
			t.Errorf("nameMatch(%q, %q) = %s, want %s", c.bank, c.expected, got, c.want)
		}
	}
}
