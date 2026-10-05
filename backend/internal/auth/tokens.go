// Package auth issues and verifies access tokens (EdDSA/Ed25519 JWTs with key rotation by kid),
// hashes passwords with argon2id and checks TOTP codes. Plan: docs/identity-access.md §4–§5.
package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Audiences: one per client. A token for one audience is rejected on another route group.
const (
	AudMarket      = "market"
	AudWholesale   = "wholesale"
	AudSeller      = "seller"
	AudStaff       = "staff"
	AudCustomerApp = "customer_app"
	AudLogistics   = "logistics"
	AudInternal    = "internal"
)

// ValidAudience reports whether aud is a known client audience.
func ValidAudience(aud string) bool {
	switch aud {
	case AudMarket, AudWholesale, AudSeller, AudStaff, AudCustomerApp, AudLogistics:
		return true
	}
	return false
}

// Claims carried by an access token. It deliberately holds no permissions, so a revoked role
// takes effect on the next request (permissions are checked live).
type Claims struct {
	SessionID uuid.UUID `json:"sid"`
	AMR       []string  `json:"amr"`
	AuthTime  int64     `json:"auth_time"`
	MFAAt     int64     `json:"mfa_at,omitempty"`
	Scopes    []string  `json:"scp,omitempty"` // service tokens only
	jwt.RegisteredClaims
}

// Keyring holds Ed25519 signing keys; the first key signs, all keys verify (rotation by kid).
type Keyring struct {
	issuer   string
	activeID string
	keys     map[string]ed25519.PrivateKey
}

// NewKeyring parses "kid:base64seed,kid2:base64seed". An empty spec in development creates a
// random ephemeral key (tokens don't survive restarts).
func NewKeyring(spec, issuer string, allowEphemeral bool) (*Keyring, error) {
	kr := &Keyring{issuer: issuer, keys: map[string]ed25519.PrivateKey{}}
	if strings.TrimSpace(spec) == "" {
		if !allowEphemeral {
			return nil, errors.New("auth: no signing keys configured")
		}
		_, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}
		kr.keys["dev-ephemeral"] = priv
		kr.activeID = "dev-ephemeral"
		return kr, nil
	}
	for i, part := range strings.Split(spec, ",") {
		kid, b64, ok := strings.Cut(strings.TrimSpace(part), ":")
		if !ok || kid == "" {
			return nil, fmt.Errorf("auth: bad key entry %d (want kid:base64seed)", i)
		}
		seed, err := base64.StdEncoding.DecodeString(b64)
		if err != nil || len(seed) != ed25519.SeedSize {
			return nil, fmt.Errorf("auth: key %q must be a base64 32-byte Ed25519 seed", kid)
		}
		kr.keys[kid] = ed25519.NewKeyFromSeed(seed)
		if i == 0 {
			kr.activeID = kid
		}
	}
	return kr, nil
}

// IssueInput describes the token to sign.
type IssueInput struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	Audience  string
	AMR       []string
	AuthTime  time.Time
	MFAAt     *time.Time
	Scopes    []string
	TTL       time.Duration
}

// Issue signs an access token.
func (k *Keyring) Issue(in IssueInput) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(in.TTL)
	c := Claims{
		SessionID: in.SessionID,
		AMR:       in.AMR,
		AuthTime:  in.AuthTime.Unix(),
		Scopes:    in.Scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    k.issuer,
			Subject:   in.UserID.String(),
			Audience:  jwt.ClaimStrings{in.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        uuid.NewString(),
		},
	}
	if in.MFAAt != nil {
		c.MFAAt = in.MFAAt.Unix()
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodEdDSA, c)
	tok.Header["kid"] = k.activeID
	s, err := tok.SignedString(k.keys[k.activeID])
	return s, exp, err
}

// Verify checks signature, issuer, expiry and that the audience matches one of allowed.
func (k *Keyring) Verify(token string, allowed ...string) (*Claims, error) {
	c := &Claims{}
	_, err := jwt.ParseWithClaims(token, c, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		priv, ok := k.keys[kid]
		if !ok {
			return nil, errors.New("unknown key id")
		}
		return priv.Public(), nil
	}, jwt.WithValidMethods([]string{"EdDSA"}), jwt.WithIssuer(k.issuer), jwt.WithLeeway(5*time.Second))
	if err != nil {
		return nil, err
	}
	if len(allowed) > 0 {
		ok := false
		for _, a := range allowed {
			for _, ca := range c.Audience {
				if a == ca {
					ok = true
				}
			}
		}
		if !ok {
			return nil, errors.New("audience not allowed for this route")
		}
	}
	return c, nil
}

// UserID returns the subject as a UUID.
func (c *Claims) UserID() (uuid.UUID, error) { return uuid.Parse(c.Subject) }

// Audience returns the token's single audience.
func (c *Claims) Aud() string {
	if len(c.Audience) == 0 {
		return ""
	}
	return c.Audience[0]
}
