package config

import (
	"strings"
	"testing"
)

const strong = "Zq8vN2kP0wR7tY4uI1oA6sD3fG5hJ9lX"

// prod returns a staging config that passes validation; tests break one rule at a time.
func prod() *Config {
	return &Config{
		Env: "staging", OTPHMACKey: strong, CursorHMACKey: strong + "1", TokenHMACKey: strong + "2", BlindIndexKey: strong + "3",
		JWTSigningKeys: "k1:abc", KMSProvider: "aws", SMSProvider: "termii", EmailProvider: "ses",
		PaymentsEnabledProviders: []string{"paystack"}, PaystackSecretKey: "sk_test_" + strong,
		DatabaseURL: "postgres://techshop:" + strong + "@db:5432/techshop",
	}
}

func TestValidStagingConfigPasses(t *testing.T) {
	if err := prod().validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
}

func TestOutsideDevelopmentRefusesWeakSettings(t *testing.T) {
	cases := map[string]func(*Config){
		"OTP_HMAC_KEY is required":  func(c *Config) { c.OTPHMACKey = "" },
		"at least 32 characters":    func(c *Config) { c.TokenHMACKey = "short" },
		"placeholder":               func(c *Config) { c.CursorHMACKey = "change-me-change-me-change-me-change-me" },
		"JWT_SIGNING_KEYS":          func(c *Config) { c.JWTSigningKeys = "" },
		"KMS_PROVIDER=local":        func(c *Config) { c.KMSProvider = "local" },
		"PAYMENTS_FAKE":             func(c *Config) { c.PaymentsFake = true },
		"cannot be 'log'":           func(c *Config) { c.SMSProvider = "log" },
		"default or short password": func(c *Config) { c.DatabaseURL = "postgres://techshop:techshop@db/techshop" },
		"PAYSTACK_SECRET_KEY":       func(c *Config) { c.PaystackSecretKey = "" },
		"has no live adapter":       func(c *Config) { c.PaymentsEnabledProviders = []string{"paystack", "opay"} },
		"APP_ENV must be":           func(c *Config) { c.Env = "prod" },
	}
	for want, breakIt := range cases {
		c := prod()
		breakIt(c)
		err := c.validate()
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("want error containing %q, got %v", want, err)
		}
	}
}

func TestDevelopmentFillsMissingSecrets(t *testing.T) {
	c := &Config{Env: "development"}
	if err := c.validate(); err != nil {
		t.Fatal(err)
	}
	if len(c.OTPHMACKey) < 32 || c.LocalKEK == "" || c.TokenHMACKey == c.CursorHMACKey {
		t.Fatal("development must get distinct random secrets")
	}
}
