// Package config loads application configuration from the environment and refuses to start
// with missing, default or weak secrets outside local development (owner rule; .claude/CLAUDE.md §3).
package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds every setting. Each field is documented in backend/.env.example.
type Config struct {
	// App
	Env               string   `env:"APP_ENV" envDefault:"development"`
	Port              string   `env:"PORT" envDefault:"8080"`
	InternalAddr      string   `env:"INTERNAL_ADDR" envDefault:":8081"`
	AppBaseURL        string   `env:"APP_BASE_URL" envDefault:"http://localhost:8080"`
	LogLevel          string   `env:"LOG_LEVEL" envDefault:"info"`
	TrustedProxies    []string `env:"TRUSTED_PROXIES" envSeparator:"," envDefault:"127.0.0.1"`
	WorkerInProcess   bool     `env:"WORKER_IN_PROCESS" envDefault:"true"`
	WorkerConcurrency int      `env:"WORKER_CONCURRENCY" envDefault:"10"`

	// Database
	DatabaseURL        string        `env:"DATABASE_URL,required"`
	MigrateOnStartup   bool          `env:"DB_MIGRATE_ON_STARTUP" envDefault:"false"`
	DBMaxConns         int32         `env:"DB_MAX_CONNS" envDefault:"10"`
	DBMinConns         int32         `env:"DB_MIN_CONNS" envDefault:"1"`
	DBStatementTimeout time.Duration `env:"DB_STATEMENT_TIMEOUT" envDefault:"15s"`

	// HTTP
	CORSAllowedOrigins []string      `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:3001,http://localhost:3002,http://localhost:3003,http://localhost:3004,http://localhost:3005,http://techshop.localhost,http://market.techshop.localhost,http://wholesale.techshop.localhost,http://seller.techshop.localhost,http://staff.techshop.localhost"`
	ReadTimeout        time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout       time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	ShutdownTimeout    time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"10s"`

	// Auth (docs/identity-access.md §5)
	JWTSigningKeys      string        `env:"JWT_SIGNING_KEYS"` // "kid:base64(ed25519 seed),kid2:…" — first is active
	JWTIssuer           string        `env:"JWT_ISSUER" envDefault:"https://api.techshop.ng"`
	AccessTokenTTL      time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"10m"`
	StaffAccessTokenTTL time.Duration `env:"STAFF_ACCESS_TOKEN_TTL" envDefault:"5m"`
	StepUpMaxAge        time.Duration `env:"STEP_UP_MAX_AGE" envDefault:"10m"`
	OTPHMACKey          string        `env:"OTP_HMAC_KEY"`
	OTPTTL              time.Duration `env:"OTP_TTL" envDefault:"10m"`
	CursorHMACKey       string        `env:"CURSOR_HMAC_KEY"`
	TokenHMACKey        string        `env:"TOKEN_HMAC_KEY"` // signs unsubscribe and click-redirect tokens

	// Crypto (envelope encryption; docs/backend.md §5.3)
	KMSProvider   string `env:"KMS_PROVIDER" envDefault:"local"`
	LocalKEK      string `env:"LOCAL_KEK"` // base64 32 bytes; development only
	BlindIndexKey string `env:"BLIND_INDEX_KEY"`

	// Payments (docs/payments-finance.md) — placeholders marked where the owner must decide
	PaymentsFake              bool          `env:"PAYMENTS_FAKE" envDefault:"true"`
	PaymentsEnabledProviders  []string      `env:"PAYMENTS_ENABLED_PROVIDERS" envSeparator:"," envDefault:"paystack,opay,moniepoint"`
	PaystackSecretKey         string        `env:"PAYSTACK_SECRET_KEY"`
	MonnifyAPIKey             string        `env:"MONNIFY_API_KEY"`
	MonnifySecretKey          string        `env:"MONNIFY_SECRET_KEY"`
	MonnifyContractCode       string        `env:"MONNIFY_CONTRACT_CODE"`
	OPayMerchantID            string        `env:"OPAY_MERCHANT_ID"`
	OPaySecretKey             string        `env:"OPAY_SECRET_KEY"`
	PaymentCallbackURL        string        `env:"PAYMENT_CALLBACK_URL" envDefault:"http://localhost:3001/checkout/return"`
	OrderPaymentTTL           time.Duration `env:"ORDER_PAYMENT_TTL" envDefault:"30m"`
	TransferAccountTTL        time.Duration `env:"TRANSFER_ACCOUNT_TTL" envDefault:"30m"`
	ReturnWindowDays          int           `env:"RETURN_WINDOW_DAYS" envDefault:"7"`
	VATBps                    int           `env:"VAT_BPS" envDefault:"750"`                            // PLACEHOLDER until finance confirms
	DefaultCommissionBps      int           `env:"DEFAULT_COMMISSION_BPS" envDefault:"1000"`            // PLACEHOLDER (owner decision)
	RefundApprovalThreshold   int64         `env:"REFUND_MANAGER_THRESHOLD_KOBO" envDefault:"50000000"` // ₦500,000
	PayoutApprovalThreshold   int64         `env:"PAYOUT_APPROVAL_THRESHOLD_KOBO" envDefault:"5000000000"`
	PayoutMinimumKobo         int64         `env:"PAYOUT_MINIMUM_KOBO" envDefault:"500000"`              // ₦5,000
	LatePaymentPolicy         string        `env:"LATE_PAYMENT_POLICY" envDefault:"reinstate_or_refund"` // PLACEHOLDER (owner decision #57)
	DeliveryFeeTechShopKobo   int64         `env:"DELIVERY_FEE_TECHSHOP_KOBO" envDefault:"250000"`
	DeliveryFeeVendorKobo     int64         `env:"DELIVERY_FEE_VENDOR_KOBO" envDefault:"350000"`
	DeliveryFeeInterstateKobo int64         `env:"DELIVERY_FEE_INTERSTATE_KOBO" envDefault:"650000"`

	// Messaging (docs/messaging-marketing.md)
	SMSProvider     string `env:"SMS_PROVIDER" envDefault:"log"`
	TermiiAPIKey    string `env:"TERMII_API_KEY"`
	TermiiSenderID  string `env:"TERMII_SENDER_ID" envDefault:"TechShop"`
	EmailProvider   string `env:"EMAIL_PROVIDER" envDefault:"log"`
	EmailFrom       string `env:"EMAIL_FROM" envDefault:"TechShop <hello@mail.techshop.ng>"`
	PushProvider    string `env:"PUSH_PROVIDER" envDefault:"log"`
	ExpoAccessToken string `env:"EXPO_ACCESS_TOKEN"`
	QuietHoursStart string `env:"QUIET_HOURS_START" envDefault:"21:00"`
	QuietHoursEnd   string `env:"QUIET_HOURS_END" envDefault:"08:00"`

	// Storage
	StorageProvider    string `env:"STORAGE_PROVIDER" envDefault:"fake"`
	S3Endpoint         string `env:"S3_ENDPOINT"`
	S3Region           string `env:"S3_REGION" envDefault:"us-east-1"`
	S3BucketPublic     string `env:"S3_BUCKET_PUBLIC" envDefault:"techshop-public"`
	S3BucketPrivate    string `env:"S3_BUCKET_PRIVATE" envDefault:"techshop-private"`
	S3AccessKeyID      string `env:"S3_ACCESS_KEY_ID"`
	S3SecretAccessKey  string `env:"S3_SECRET_ACCESS_KEY"`
	S3ForcePathStyle   bool   `env:"S3_FORCE_PATH_STYLE" envDefault:"true"`
	PublicAssetBaseURL string `env:"PUBLIC_ASSET_BASE_URL" envDefault:"http://localhost:9000/techshop-public"`

	// Integrations (all fake by default; no live calls)
	KYCProvider     string        `env:"KYC_PROVIDER" envDefault:"manual"`
	CarrierProvider string        `env:"CARRIER_PROVIDER" envDefault:"fake"`
	RecsLiveURL     string        `env:"RECS_LIVE_URL"`
	RecsLiveTimeout time.Duration `env:"RECS_LIVE_TIMEOUT" envDefault:"80ms"`
}

// Load reads a local .env file if present, parses the environment and validates secrets.
func Load() (*Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// IsDevelopment reports whether this is a local development environment.
func (c *Config) IsDevelopment() bool { return c.Env == "development" }

// IsProduction reports whether this is the production environment.
func (c *Config) IsProduction() bool { return c.Env == "production" }

// weakPasswords are values that must never protect a non-development database.
var weakPasswords = map[string]bool{"techshop": true, "postgres": true, "password": true, "changeme": true, "secret": true, "admin": true}

// validate enforces the secret rules. In development, missing secrets are replaced with random
// per-process values (tokens then don't survive restarts — acceptable locally).
func (c *Config) validate() error {
	switch c.Env {
	case "development", "test", "staging", "production":
	default:
		return fmt.Errorf("APP_ENV must be development, test, staging or production, got %q", c.Env)
	}

	secrets := map[string]*string{
		"OTP_HMAC_KEY":    &c.OTPHMACKey,
		"CURSOR_HMAC_KEY": &c.CursorHMACKey,
		"TOKEN_HMAC_KEY":  &c.TokenHMACKey,
		"BLIND_INDEX_KEY": &c.BlindIndexKey,
	}

	if c.IsDevelopment() || c.Env == "test" {
		for _, p := range secrets {
			if *p == "" {
				*p = randomSecret()
			}
		}
		if c.LocalKEK == "" {
			c.LocalKEK = randomSecret()
		}
		return nil
	}

	var errs []error
	for name, p := range secrets {
		if err := checkStrong(name, *p); err != nil {
			errs = append(errs, err)
		}
	}
	if c.JWTSigningKeys == "" {
		errs = append(errs, errors.New("JWT_SIGNING_KEYS is required outside development"))
	}
	if c.LocalKEK != "" || c.KMSProvider == "local" {
		errs = append(errs, errors.New("LOCAL_KEK / KMS_PROVIDER=local are development-only; configure a KMS"))
	}
	if c.PaymentsFake {
		errs = append(errs, errors.New("PAYMENTS_FAKE must be false outside development"))
	}
	// Only Paystack has a live adapter so far; Monnify (Moniepoint) and OPay run as fakes until
	// their adapters are built and confirmed against the providers' sandboxes.
	for _, p := range c.PaymentsEnabledProviders {
		switch p {
		case "paystack":
			if err := checkStrong("PAYSTACK_SECRET_KEY", c.PaystackSecretKey); err != nil {
				errs = append(errs, err)
			}
		default:
			errs = append(errs, fmt.Errorf("payment provider %q has no live adapter yet; remove it from PAYMENTS_ENABLED_PROVIDERS", p))
		}
	}
	if c.SMSProvider == "log" || c.EmailProvider == "log" {
		errs = append(errs, errors.New("SMS_PROVIDER and EMAIL_PROVIDER cannot be 'log' outside development"))
	}
	if err := checkDatabasePassword(c.DatabaseURL); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// checkStrong rejects empty, short or obviously example secrets.
func checkStrong(name, value string) error {
	lower := strings.ToLower(value)
	switch {
	case value == "":
		return fmt.Errorf("%s is required outside development", name)
	case len(value) < 32:
		return fmt.Errorf("%s must be at least 32 characters", name)
	case strings.Contains(lower, "change") || strings.Contains(lower, "example") || strings.Contains(lower, "secret"):
		return fmt.Errorf("%s looks like a placeholder value", name)
	}
	return nil
}

// checkDatabasePassword refuses default or short database passwords outside development.
func checkDatabasePassword(databaseURL string) error {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("DATABASE_URL is not a valid URL: %w", err)
	}
	pw, _ := u.User.Password()
	if weakPasswords[strings.ToLower(pw)] || len(pw) < 16 {
		return errors.New("DATABASE_URL uses a default or short password (need 16+ characters)")
	}
	return nil
}

func randomSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
