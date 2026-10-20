package auth

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// NewTOTPSecret creates a TOTP secret and the otpauth:// URL shown as a QR code.
func NewTOTPSecret(issuer, account string) (secret, url string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{Issuer: issuer, AccountName: account, Period: 30, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// CheckTOTP validates a code within ±1 step and returns the time step it matched, so callers
// can store it and refuse the same code twice (user_mfa_factors.last_used_step).
func CheckTOTP(secret, code string, now time.Time) (step int64, ok bool) {
	for _, skew := range []int64{0, -1, 1} {
		t := now.Add(time.Duration(skew*30) * time.Second)
		valid, err := totp.ValidateCustom(code, secret, t, totp.ValidateOpts{Period: 30, Skew: 0, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
		if err == nil && valid {
			return t.Unix() / 30, true
		}
	}
	return 0, false
}

// TOTPCode returns the current code for a secret (dev seed and tests only).
func TOTPCode(secret string, now time.Time) (string, error) {
	return totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{Period: 30, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
}
