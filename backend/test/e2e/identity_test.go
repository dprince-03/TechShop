//go:build e2e

package e2e

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func randomPhone() string { return fmt.Sprintf("+23481%08d", rand.Intn(100000000)) }

func TestHealth(t *testing.T) {
	r := get(t, "/readyz", "")
	expect(t, r, 200, "readyz")
	if r.str("database") != "up" {
		t.Fatal("database not up")
	}
}

func TestOTPSignupRefreshReuse(t *testing.T) {
	expectCode(t, post(t, "/api/v1/auth/otp/request", "", map[string]any{"phone": "12345"}), 422, "invalid_phone", "bad phone")

	phone := randomPhone()
	r := post(t, "/api/v1/auth/otp/request", "", map[string]any{"phone": phone})
	expect(t, r, 200, "otp request")
	challenge := r.str("challengeId")
	// Second request within a minute is limited (1/min per number).
	expectCode(t, post(t, "/api/v1/auth/otp/request", "", map[string]any{"phone": phone}), 429, "otp_rate_limited", "otp per-minute limit")

	code := latestCode(t, phone, "")
	wrong := "000000"
	if code == wrong {
		wrong = "111111"
	}
	bad := post(t, "/api/v1/auth/otp/verify", "", map[string]any{"challengeId": challenge, "code": wrong, "audience": "market"})
	expectCode(t, bad, 422, "code_invalid", "wrong code")

	ok := post(t, "/api/v1/auth/otp/verify", "", map[string]any{"challengeId": challenge, "code": code, "audience": "market"})
	expect(t, ok, 200, "verify")
	if ok.str("status") != "needs_profile" {
		t.Fatalf("new number should need a profile: %s", ok.Raw)
	}
	// The code is single use.
	expectCode(t, post(t, "/api/v1/auth/otp/verify", "", map[string]any{"challengeId": challenge, "code": code, "audience": "market"}), 422, "code_expired", "code reuse")

	s := post(t, "/api/v1/auth/otp/signup", "", map[string]any{"signupToken": ok.str("signupToken"), "firstName": "Test", "lastName": "Buyer", "audience": "market", "deviceName": "e2e"})
	expect(t, s, 200, "signup")
	access, refresh := s.str("accessToken"), s.str("refreshToken")

	me := get(t, "/api/v1/me", access)
	expect(t, me, 200, "me")
	if me.str("phone") != phone || me.str("firstName") != "Test" {
		t.Fatalf("me: %s", me.Raw)
	}
	expect(t, get(t, "/api/v1/me", ""), 401, "me without token")
	expect(t, get(t, "/api/v1/staff/admin/roles", access), 401, "customer token on staff route")

	// Rotate, then replay the old refresh token after the grace window → whole family revoked.
	r1 := post(t, "/api/v1/auth/refresh", "", map[string]any{"refreshToken": refresh})
	expect(t, r1, 200, "refresh")
	newRefresh := r1.str("refreshToken")
	time.Sleep(11 * time.Second)
	expectCode(t, post(t, "/api/v1/auth/refresh", "", map[string]any{"refreshToken": refresh}), 401, "unauthenticated", "reuse detected")
	expectCode(t, post(t, "/api/v1/auth/refresh", "", map[string]any{"refreshToken": newRefresh}), 401, "unauthenticated", "family revoked")
	events := get(t, "/api/v1/me/security-events", r1.str("accessToken"))
	if events.Status == 200 {
		found := false
		for _, e := range events.List {
			if e.(map[string]any)["kind"] == "refresh_reuse_detected" {
				found = true
			}
		}
		if !found {
			t.Fatalf("reuse event missing: %s", events.Raw)
		}
	}
}

func TestAddressesSessionsAndTOTP(t *testing.T) {
	tok := customerLogin(t, "customer_b")
	a := post(t, "/api/v1/me/addresses", tok, map[string]any{"recipientName": "Halima Bello", "phone": "0803 123 4570", "line1": "14 Admiralty Way", "city": "Lekki", "stateCode": "LA"})
	expect(t, a, 201, "create address")
	expectCode(t, post(t, "/api/v1/me/addresses", tok, map[string]any{"recipientName": "Xy Z", "phone": "123", "line1": "abc", "city": "Ikeja", "stateCode": "LA"}), 422, "invalid_phone", "address phone")
	expectCode(t, post(t, "/api/v1/me/addresses", tok, map[string]any{"recipientName": "X Y", "phone": "08031234570", "line1": "abc", "city": "Ikeja", "stateCode": "XX"}), 422, "invalid_state", "address state")
	list := get(t, "/api/v1/me/addresses", tok)
	expect(t, list, 200, "list addresses")
	defaults := 0
	for _, x := range list.List {
		if x.(map[string]any)["isDefault"] == true {
			defaults++
		}
	}
	if len(list.List) < 1 || defaults != 1 {
		t.Fatalf("want exactly one default address, got %d of %d", defaults, len(list.List))
	}
	sess := get(t, "/api/v1/me/sessions", tok)
	expect(t, sess, 200, "sessions")

	// Authenticator: setup, confirm, then a sensitive call needs a recent step-up.
	setup := post(t, "/api/v1/me/mfa/totp/setup", tok, nil)
	expect(t, setup, 200, "totp setup")
	secret := setup.str("secret")
	conf := post(t, "/api/v1/me/mfa/totp/confirm", tok, map[string]any{"code": totpCode(t, secret, 0)})
	expect(t, conf, 200, "totp confirm")
	if n := len(conf.get("recoveryCodes").([]any)); n != 10 {
		t.Fatalf("want 10 recovery codes, got %d", n)
	}
	expectCode(t, call(t, req{method: "DELETE", path: "/api/v1/me/mfa/totp", token: tok}), 401, "step_up_required", "remove totp without step-up")
	time.Sleep(31 * time.Second) // next TOTP step (codes can't be reused)
	su := post(t, "/api/v1/auth/mfa/step-up", tok, map[string]any{"code": totpCode(t, secret, 0)})
	expect(t, su, 200, "step-up")
	expect(t, call(t, req{method: "DELETE", path: "/api/v1/me/mfa/totp", token: su.str("accessToken")}), 204, "remove totp after step-up")
}

func TestStaffAdminFlows(t *testing.T) {
	admin := staffLogin(t, "admin")
	roles := get(t, "/api/v1/staff/admin/roles", admin)
	expect(t, roles, 200, "roles")

	_, _, _, _, agentID := account(t, "support_agent")
	_, _, _, _, adminID := account(t, "admin")
	// Granting roles is sensitive (★): the fresh staff login has mfa_at = now, so it passes.
	expect(t, post(t, "/api/v1/staff/admin/members/"+agentID+"/roles", admin, map[string]any{"role": "auditor"}), 204, "grant role")
	expectCode(t, post(t, "/api/v1/staff/admin/members/"+adminID+"/roles", admin, map[string]any{"role": "auditor"}), 403, "forbidden", "self-grant blocked")

	// Role removal takes effect on the very next request.
	agent := staffLogin(t, "support_agent")
	expect(t, get(t, "/api/v1/staff/admin/audit-log", agent), 200, "auditor can view audit log")
	expect(t, call(t, req{method: "DELETE", path: "/api/v1/staff/admin/members/" + agentID + "/roles/auditor", token: admin}), 204, "revoke role")
	time.Sleep(300 * time.Millisecond) // NOTIFY → cache invalidation
	expect(t, get(t, "/api/v1/staff/admin/audit-log", agent), 403, "revoked immediately")

	// Customers can't use staff routes; staff without a permission get 403.
	cust := customerLogin(t, "customer")
	expect(t, get(t, "/api/v1/staff/hr/members", cust), 401, "customer on staff route")
	expect(t, get(t, "/api/v1/staff/hr/members", agent), 403, "missing permission")

	// HR adds a staff member; the invite works once.
	hr := staffLogin(t, "hr_officer")
	email := fmt.Sprintf("new%d@dev.techshop.ng", time.Now().UnixNano())
	c := post(t, "/api/v1/staff/hr/members", hr, map[string]any{"email": email, "firstName": "New", "lastName": "Hire", "employeeNo": fmt.Sprint(time.Now().UnixNano()), "department": "Ops", "jobTitle": "Agent"})
	expect(t, c, 201, "create staff")
	link := c.str("inviteUrl")
	token := link[len(link)-43:]
	acc := post(t, "/api/v1/auth/invite/accept", "", map[string]any{"token": token, "password": "a long staff password"})
	expect(t, acc, 200, "accept invite")
	expectCode(t, post(t, "/api/v1/auth/invite/accept", "", map[string]any{"token": token, "password": "a long staff password"}), 422, "invite_invalid", "invite reuse")
	cf := post(t, "/api/v1/auth/invite/confirm", "", map[string]any{"enrollToken": acc.str("enrollToken"), "code": totpCode(t, acc.str("secret"), 0), "audience": "staff"})
	expect(t, cf, 200, "confirm invite")
	newToken := cf.str("tokens", "accessToken")
	expect(t, get(t, "/api/v1/me", newToken), 200, "new staff signed in")

	// Exiting removes sessions at once.
	expect(t, post(t, "/api/v1/staff/hr/members/"+cf.str("tokens", "userId")+"/exit", hr, nil), 204, "exit staff")
	time.Sleep(300 * time.Millisecond)
	expect(t, get(t, "/api/v1/me", newToken), 401, "exited staff signed out")
}
