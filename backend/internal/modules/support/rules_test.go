package support

import (
	"testing"
	"time"
)

func TestPriorityAndSLA(t *testing.T) {
	if priorityFor("refund") != "high" || priorityFor("payment") != "high" || priorityFor("delivery") != "normal" {
		t.Fatal("money topics start high, others normal")
	}
	now := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
	if got := slaDue("high", now); !got.Equal(now.Add(4 * time.Hour)) {
		t.Fatalf("high: %v", got)
	}
	if got := slaDue("urgent", now); !got.Equal(now.Add(4 * time.Hour)) {
		t.Fatalf("urgent: %v", got)
	}
	if got := slaDue("normal", now); !got.Equal(now.Add(24 * time.Hour)) {
		t.Fatalf("normal: %v", got)
	}
}

func TestHoneypotBodyExcerpt(t *testing.T) {
	if honeypotTripped("") || honeypotTripped("  ") || !honeypotTripped("http://spam") {
		t.Fatal("honeypot")
	}
	if _, ok := cleanBody("   \n"); ok {
		t.Fatal("blank body accepted")
	}
	if b, ok := cleanBody("  Where is my order?  "); !ok || b != "Where is my order?" {
		t.Fatalf("clean body: %q", b)
	}
	if got := excerpt("one  two\nthree", 50); got != "one two three" {
		t.Fatalf("excerpt short: %q", got)
	}
	if got := excerpt("abcdefghij", 5); got != "abcd…" {
		t.Fatalf("excerpt long: %q", got)
	}
}

func TestValidEmail(t *testing.T) {
	for _, ok := range []string{"ada@example.com", "a.b+c@mail.co.ng"} {
		if !validEmail(ok) {
			t.Errorf("%s should be valid", ok)
		}
	}
	for _, bad := range []string{"ada", "ada@", "ada@localhost", "Ada <ada@example.com>", "a b@example.com"} {
		if validEmail(bad) {
			t.Errorf("%q should be invalid", bad)
		}
	}
}
