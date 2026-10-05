// Package crypto provides envelope encryption (AES-256-GCM with associated data), blind-index
// HMACs for duplicate detection, token hashing and secure random values.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
)

// Sealer encrypts and decrypts small secrets (NIN, bank account numbers, TOTP secrets).
// Associated data binds ciphertext to its table, column and row so it can't be moved.
type Sealer interface {
	Seal(plaintext []byte, aad string) ([]byte, error)
	Open(ciphertext []byte, aad string) ([]byte, error)
}

// LocalSealer uses a key-encryption key from config (development only; production uses a KMS
// adapter that wraps per-purpose data keys stored in platform.encryption_keys).
type LocalSealer struct{ aead cipher.AEAD }

// NewLocalSealer builds a sealer from a base64 32-byte key (any longer input is hashed to 32 bytes).
func NewLocalSealer(keyB64 string) (*LocalSealer, error) {
	raw, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil || len(raw) < 16 {
		sum := sha256.Sum256([]byte(keyB64))
		raw = sum[:]
	} else if len(raw) != 32 {
		sum := sha256.Sum256(raw)
		raw = sum[:]
	}
	block, err := aes.NewCipher(raw)
	if err != nil {
		return nil, fmt.Errorf("aes: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	return &LocalSealer{aead: aead}, nil
}

// Seal returns nonce || ciphertext.
func (s *LocalSealer) Seal(plaintext []byte, aad string) ([]byte, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return s.aead.Seal(nonce, nonce, plaintext, []byte(aad)), nil
}

// Open reverses Seal; it fails if the ciphertext was tampered with or moved to another row.
func (s *LocalSealer) Open(ciphertext []byte, aad string) ([]byte, error) {
	n := s.aead.NonceSize()
	if len(ciphertext) < n {
		return nil, errors.New("crypto: ciphertext too short")
	}
	return s.aead.Open(nil, ciphertext[:n], ciphertext[n:], []byte(aad))
}

// HMAC returns HMAC-SHA256(key, msg). Used for OTP storage and blind indexes.
func HMAC(key, msg string) []byte {
	m := hmac.New(sha256.New, []byte(key))
	m.Write([]byte(msg))
	return m.Sum(nil)
}

// SHA256 hashes opaque tokens (refresh tokens, invite links) before storage.
func SHA256(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// RandomToken returns a URL-safe random token with n bytes of entropy.
func RandomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// RandomDigits returns an n-digit numeric code (e.g. a 6-digit OTP) with uniform digits.
func RandomDigits(n int) string {
	out := make([]byte, n)
	for i := range out {
		v, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			panic(err)
		}
		out[i] = byte('0' + v.Int64())
	}
	return string(out)
}

// Equal compares two byte slices in constant time.
func Equal(a, b []byte) bool { return hmac.Equal(a, b) }
