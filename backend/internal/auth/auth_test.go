package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/uuid"
)

func seed() string {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	return base64.StdEncoding.EncodeToString(priv.Seed())
}

func TestTokensRotateAndCheckAudience(t *testing.T) {
	oldKey, newKey := seed(), seed()
	oldRing, _ := NewKeyring("k1:"+oldKey, "iss", false)
	tok, _, err := oldRing.Issue(IssueInput{UserID: uuid.New(), SessionID: uuid.New(), Audience: AudMarket, TTL: time.Minute, AuthTime: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	// After rotation the new key signs, the old one still verifies.
	ring, _ := NewKeyring("k2:"+newKey+",k1:"+oldKey, "iss", false)
	if _, err := ring.Verify(tok, AudMarket); err != nil {
		t.Fatalf("old token should verify after rotation: %v", err)
	}
	if _, err := ring.Verify(tok, AudStaff); err == nil {
		t.Fatal("market token must be rejected on staff routes")
	}
	other, _ := NewKeyring("k1:"+seed(), "iss", false)
	if _, err := other.Verify(tok); err == nil {
		t.Fatal("token signed with a different key must fail")
	}
}

func TestPassword(t *testing.T) {
	h, err := HashPassword("correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword("correct horse battery", h) || VerifyPassword("wrong horse battery", h) {
		t.Fatal("verify mismatch")
	}
	if CheckPasswordPolicy("short") == nil || CheckPasswordPolicy("password123") == nil || CheckPasswordPolicy("a long enough phrase") != nil {
		t.Fatal("policy mismatch")
	}
}

func TestTOTPRejectsWrongCode(t *testing.T) {
	secret, _, err := NewTOTPSecret("TechShop", "ada@example.com")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	code, _ := TOTPCode(secret, now)
	if _, ok := CheckTOTP(secret, code, now); !ok {
		t.Fatal("current code rejected")
	}
	if _, ok := CheckTOTP(secret, "000000", now); ok && code != "000000" {
		t.Fatal("wrong code accepted")
	}
}
