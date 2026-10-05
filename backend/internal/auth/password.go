package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters (stored per hash so they can be raised later).
const (
	argonMemoryKiB = 64 * 1024
	argonTime      = 3
	argonThreads   = 1
	argonKeyLen    = 32
)

// commonPasswords is a small local breached-password list (the full k-anonymity range check
// is an owner decision; docs/identity-access.md §16).
var commonPasswords = map[string]bool{"password123": true, "1234567890": true, "qwertyuiop": true, "password1234": true,
	"iloveyou123": true, "techshop123": true, "adminadmin": true, "0123456789": true, "abcdefghij": true}

// ErrWeakPassword is returned for passwords under 10 characters or on the breached list.
var ErrWeakPassword = errors.New("password must be at least 10 characters and not a common password")

// CheckPasswordPolicy enforces length ≥ 10 and the breached list (no composition rules).
func CheckPasswordPolicy(pw string) error {
	if len(pw) < 10 || len(pw) > 128 || commonPasswords[strings.ToLower(pw)] {
		return ErrWeakPassword
	}
	return nil
}

// HashPassword returns an encoded argon2id hash.
func HashPassword(pw string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(pw), salt, argonTime, argonMemoryKiB, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argonMemoryKiB, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

// VerifyPassword compares pw with an encoded hash in constant time.
func VerifyPassword(pw, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var m, t uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[4])
	want, err2 := base64.RawStdEncoding.DecodeString(parts[5])
	if err1 != nil || err2 != nil {
		return false
	}
	got := argon2.IDKey([]byte(pw), salt, t, m, p, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

// DummyVerify burns the same time as a real check, so "no such user" and "wrong password"
// take the same time (no account enumeration).
func DummyVerify(pw string) {
	_ = argon2.IDKey([]byte(pw), []byte("constant-salt-16"), argonTime, argonMemoryKiB, argonThreads, argonKeyLen)
}
