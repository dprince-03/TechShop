//go:build e2e

// Package e2e makes real HTTP calls against a running API (API_URL) seeded with cmd/seed
// (SEED_FILE). Run it in a container on the same Docker network as the API.
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

var (
	baseURL = os.Getenv("API_URL")
	client  = &http.Client{Timeout: 30 * time.Second}
	seed    struct {
		Accounts []struct {
			Role       string `json:"role"`
			UserID     string `json:"userId"`
			Email      string `json:"email"`
			Phone      string `json:"phone"`
			Password   string `json:"password"`
			TOTPSecret string `json:"totpSecret"`
		} `json:"accounts"`
		Extra map[string]string `json:"extra"`
	}
)

func TestMain(m *testing.M) {
	b, err := os.ReadFile(os.Getenv("SEED_FILE"))
	if err == nil {
		_ = json.Unmarshal(b, &seed)
	}
	os.Exit(m.Run())
}

// resp is a decoded response.
type resp struct {
	Status int
	Body   map[string]any
	List   []any
	Raw    []byte
	Header http.Header
}

func (r resp) str(path ...string) string {
	v := r.get(path...)
	if s, ok := v.(string); ok {
		return s
	}
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func (r resp) get(path ...string) any {
	var cur any = r.Body
	if r.Body == nil && r.List != nil {
		cur = r.List
	}
	for _, p := range path {
		switch c := cur.(type) {
		case map[string]any:
			cur = c[p]
		case []any:
			var i int
			fmt.Sscan(p, &i)
			if i >= len(c) {
				return nil
			}
			cur = c[i]
		default:
			return nil
		}
	}
	return cur
}

func (r resp) num(path ...string) float64 {
	f, _ := r.get(path...).(float64)
	return f
}

type req struct {
	method, path, token string
	body                any
	raw                 string // sent verbatim instead of body (e.g. a signed webhook)
	headers             map[string]string
}

func call(t *testing.T, r req) resp {
	t.Helper()
	var rd io.Reader
	if r.raw != "" {
		rd = strings.NewReader(r.raw)
	} else if r.body != nil {
		b, _ := json.Marshal(r.body)
		rd = bytes.NewReader(b)
	}
	hr, err := http.NewRequest(r.method, baseURL+r.path, rd)
	if err != nil {
		t.Fatal(err)
	}
	hr.Header.Set("Content-Type", "application/json")
	if r.token != "" {
		hr.Header.Set("Authorization", "Bearer "+r.token)
	}
	for k, v := range r.headers {
		hr.Header.Set(k, v)
	}
	res, err := client.Do(hr)
	if err != nil {
		t.Fatalf("%s %s: %v", r.method, r.path, err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	out := resp{Status: res.StatusCode, Raw: raw, Header: res.Header}
	if len(raw) > 0 && raw[0] == '[' {
		_ = json.Unmarshal(raw, &out.List)
	} else {
		_ = json.Unmarshal(raw, &out.Body)
	}
	return out
}

func get(t *testing.T, path, token string) resp {
	return call(t, req{method: "GET", path: path, token: token})
}
func post(t *testing.T, path, token string, body any) resp {
	return call(t, req{method: "POST", path: path, token: token, body: body})
}

func expect(t *testing.T, r resp, status int, what string) {
	t.Helper()
	if r.Status != status {
		t.Fatalf("%s: want %d, got %d: %s", what, status, r.Status, truncate(string(r.Raw)))
	}
}

func expectCode(t *testing.T, r resp, status int, code, what string) {
	t.Helper()
	expect(t, r, status, what)
	if r.str("code") != code {
		t.Fatalf("%s: want code %q, got %q (%s)", what, code, r.str("code"), truncate(string(r.Raw)))
	}
}

func truncate(s string) string {
	if len(s) > 600 {
		return s[:600] + "…"
	}
	return s
}

var sixDigits = regexp.MustCompile(`\b(\d{6})\b`)

// latestCode waits for the worker to send the newest code to a phone and returns it.
func latestCode(t *testing.T, phone, notBefore string) string {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		r := get(t, "/dev/messages?to="+urlEncode(phone), "")
		for _, m := range r.List {
			mm := m.(map[string]any)
			payload, _ := mm["payload"].(map[string]any)
			body, _ := payload["body"].(string)
			if body == "" {
				continue
			}
			// Only the newest message counts: wait until it carries a code we haven't seen.
			if c := sixDigits.FindStringSubmatch(body); c != nil {
				if c[1] != notBefore {
					return c[1]
				}
				break
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
	t.Fatalf("no code delivered to %s", phone)
	return ""
}

func urlEncode(s string) string { return strings.ReplaceAll(s, "+", "%2B") }

func account(t *testing.T, role string) (email, phone, password, secret, id string) {
	t.Helper()
	for _, a := range seed.Accounts {
		if a.Role == role {
			return a.Email, a.Phone, a.Password, a.TOTPSecret, a.UserID
		}
	}
	t.Fatalf("seed account %q missing", role)
	return
}

// staffTokens caches staff sessions: each TOTP code works once per 30-second step, and a
// fresh login also satisfies step-up (10 minutes), so tokens are reused for 8 minutes.
var (
	staffMu     sync.Mutex
	staffTokens = map[string]struct {
		token string
		at    time.Time
	}{}
)

// staffLogin signs a seeded staff member in with password + TOTP (cached per role).
func staffLogin(t *testing.T, role string) string {
	t.Helper()
	staffMu.Lock()
	defer staffMu.Unlock()
	if c, ok := staffTokens[role]; ok && time.Since(c.at) < 8*time.Minute {
		return c.token
	}
	tok := staffLoginFresh(t, role)
	staffTokens[role] = struct {
		token string
		at    time.Time
	}{tok, time.Now()}
	return tok
}

func staffLoginFresh(t *testing.T, role string) string {
	t.Helper()
	email, _, pw, secret, _ := account(t, role)
	r := post(t, "/api/v1/auth/password", "", map[string]any{"identifier": email, "password": pw, "audience": "staff"})
	expect(t, r, 200, "staff password "+role)
	code := totpCode(t, secret, 0)
	r2 := post(t, "/api/v1/auth/mfa/totp", "", map[string]any{"mfaToken": r.str("mfaToken"), "code": code, "audience": "staff"})
	if r2.Status != 200 && r2.str("code") == "code_reused" {
		time.Sleep(31 * time.Second)
		return staffLoginFresh(t, role)
	}
	expect(t, r2, 200, "staff totp "+role)
	return r2.str("accessToken")
}

func totpCode(t *testing.T, secret string, offset time.Duration) string {
	c, err := totp.GenerateCode(secret, time.Now().Add(offset))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// customerLogin signs a seeded customer in with password on the market audience.
func customerLogin(t *testing.T, role string) string {
	t.Helper()
	email, _, pw, _, _ := account(t, role)
	r := post(t, "/api/v1/auth/password", "", map[string]any{"identifier": email, "password": pw, "audience": "market"})
	expect(t, r, 200, "customer login "+role)
	return r.str("tokens", "accessToken")
}

func newUUID() string { return uuid.NewString() }

// latestCodeContaining waits for the newest message to a phone whose text contains `about`
// (e.g. "delivery code") and returns the 6-digit code in it.
func latestCodeContaining(t *testing.T, phone, about string) string {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		r := get(t, "/dev/messages?to="+urlEncode(phone), "")
		for _, m := range r.List {
			payload, _ := m.(map[string]any)["payload"].(map[string]any)
			body, _ := payload["body"].(string)
			if strings.Contains(body, about) {
				if c := sixDigits.FindStringSubmatch(body); c != nil {
					return c[1]
				}
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
	t.Fatalf("no %q message delivered to %s", about, phone)
	return ""
}

// peekCode returns the newest 6-digit code already sent to a phone ("" if none), so a test can
// wait for a fresh one.
func peekCode(t *testing.T, phone string) string {
	t.Helper()
	r := get(t, "/dev/messages?to="+urlEncode(phone), "")
	for _, m := range r.List {
		payload, _ := m.(map[string]any)["payload"].(map[string]any)
		if body, _ := payload["body"].(string); body != "" {
			if c := sixDigits.FindStringSubmatch(body); c != nil {
				return c[1]
			}
		}
	}
	return ""
}
