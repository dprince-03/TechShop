//go:build e2e

package e2e

import (
	"strings"
	"testing"
	"time"
)

// inviteToken pulls the token out of a development invite link.
func inviteToken(t *testing.T, r resp) string {
	t.Helper()
	link := r.str("link")
	_, tok, ok := strings.Cut(link, "/invite/")
	if !ok || tok == "" {
		t.Fatalf("no invite link in %s", r.Raw)
	}
	return tok
}

// enrolAuthenticator turns on an authenticator for an account and returns its secret.
func enrolAuthenticator(t *testing.T, tok string) string {
	t.Helper()
	setup := post(t, "/api/v1/me/mfa/totp/setup", tok, nil)
	expect(t, setup, 200, "totp setup")
	expect(t, post(t, "/api/v1/me/mfa/totp/confirm", tok, map[string]any{"code": totpCode(t, setup.str("secret"), 0)}), 200, "totp confirm")
	return setup.str("secret")
}

// Seller teams (identity-access.md §7.3): invites by phone, single use and tied to the invitee's
// account; managers can't touch owners; the last owner stays; finance needs an authenticator.
func TestSellerTeamInvites(t *testing.T) {
	t.Parallel()
	owner, shopID := approvedShop(t, "Bisi", "Coker")
	team := "/api/v1/seller/shops/" + shopID + "/team"

	mgr, mgrPhone := sellerSignupPhone(t, "Tunde", "Manager")
	stranger, _ := sellerSignupPhone(t, "Kola", "Stranger")
	fin, finPhone := sellerSignupPhone(t, "Femi", "Finance")

	// Invite a manager by phone; the text carries the link.
	expectCode(t, post(t, team+"/invites", owner, map[string]any{"phone": "12345", "role": "manager"}), 422, "invalid_phone", "bad phone")
	expectCode(t, post(t, team+"/invites", owner, map[string]any{"role": "manager"}), 422, "phone_or_email", "no target")
	inv := post(t, team+"/invites", owner, map[string]any{"phone": mgrPhone, "role": "manager"})
	expect(t, inv, 201, "invite manager")
	expectCode(t, post(t, team+"/invites", owner, map[string]any{"phone": mgrPhone, "role": "staff"}), 409, "already_invited", "second open invite")
	first := inviteToken(t, inv)
	if !strings.Contains(latestSMS(t, mgrPhone, "invited you"), "/invite/"+first) {
		t.Fatal("invite text should carry the link")
	}

	// Re-sending replaces the link: the old one stops working.
	re := post(t, team+"/invites/"+inv.str("id")+"/resend", owner, nil)
	expect(t, re, 200, "resend")
	tok := inviteToken(t, re)
	expectCode(t, post(t, "/api/v1/seller/invites/accept", mgr, map[string]any{"token": first}), 422, "invite_invalid", "old link")

	// Only the invited account can accept; then it's single use.
	p := post(t, "/api/v1/seller/invites/preview", stranger, map[string]any{"token": tok})
	expect(t, p, 200, "preview")
	if p.str("role") != "manager" || strings.Contains(string(p.Raw), mgrPhone) {
		t.Fatalf("preview shows the role and a masked contact only: %s", p.Raw)
	}
	expectCode(t, post(t, "/api/v1/seller/invites/accept", stranger, map[string]any{"token": tok}), 403, "forbidden", "someone else accepts")
	acc := post(t, "/api/v1/seller/invites/accept", mgr, map[string]any{"token": tok})
	expect(t, acc, 200, "accept")
	if acc.str("id") != shopID || acc.str("role") != "manager" {
		t.Fatalf("joined shop: %s", acc.Raw)
	}
	expectCode(t, post(t, "/api/v1/seller/invites/accept", mgr, map[string]any{"token": tok}), 422, "invite_invalid", "reuse")
	expect(t, get(t, "/api/v1/seller/shops/"+shopID+"/orders", mgr), 200, "manager sees orders")
	expect(t, get(t, "/api/v1/seller/shops/"+shopID+"/orders", stranger), 404, "stranger doesn't")

	// Managers can't invite or touch owners, and the last owner can't leave.
	expectCode(t, post(t, team+"/invites", mgr, map[string]any{"email": "co-owner@example.com", "role": "owner"}), 403, "forbidden", "manager invites owner")
	tm := get(t, team, owner)
	expect(t, tm, 200, "team")
	var ownerID string
	for _, m := range tm.get("members").([]any) {
		if mm := m.(map[string]any); mm["role"] == "owner" {
			ownerID = mm["userId"].(string)
		}
	}
	if strings.Contains(string(tm.Raw), mgrPhone) {
		t.Fatalf("team list must mask phones: %s", tm.Raw)
	}
	expectCode(t, call(t, req{method: "DELETE", path: team + "/members/" + ownerID, token: mgr}), 403, "forbidden", "manager removes owner")
	expectCode(t, call(t, req{method: "DELETE", path: team + "/members/" + ownerID, token: owner}), 409, "last_owner", "last owner leaves")

	// Finance: no authenticator, no entry; with one, the payouts screens open.
	fi := post(t, team+"/invites", mgr, map[string]any{"phone": finPhone, "role": "finance"})
	expect(t, fi, 201, "invite finance")
	ft := inviteToken(t, fi)
	expectCode(t, post(t, "/api/v1/seller/invites/accept", fin, map[string]any{"token": ft}), 403, "authenticator_required", "finance without MFA")
	secret := enrolAuthenticator(t, fin)
	expect(t, post(t, "/api/v1/seller/invites/accept", fin, map[string]any{"token": ft}), 200, "finance accepts")
	expect(t, get(t, "/api/v1/seller/shops/"+shopID+"/balance", fin), 200, "finance sees balance")

	// With an authenticator on, a texted code alone no longer signs in: the app asks for it.
	before := peekCode(t, finPhone)
	var rq resp
	waitFor(t, "the per-number code limit (1 a minute) to allow another code", 75*time.Second, func() bool {
		rq = post(t, "/api/v1/auth/otp/request", "", map[string]any{"phone": finPhone})
		if rq.Status == 429 {
			time.Sleep(5 * time.Second)
			return false
		}
		return true
	})
	expect(t, rq, 200, "finance otp request")
	v := post(t, "/api/v1/auth/otp/verify", "", map[string]any{"challengeId": rq.str("challengeId"), "code": latestCode(t, finPhone, before), "audience": "seller"})
	expect(t, v, 200, "finance otp verify")
	if v.str("status") != "mfa_required" || v.str("mfaToken") == "" || v.get("tokens") != nil {
		t.Fatalf("an authenticator account must finish with its code: %s", v.Raw)
	}
	expectCode(t, post(t, "/api/v1/auth/mfa/totp", "", map[string]any{"mfaToken": v.str("mfaToken"), "code": "000000", "audience": "seller"}), 422, "code_invalid", "wrong authenticator code")
	in := post(t, "/api/v1/auth/mfa/totp", "", map[string]any{"mfaToken": v.str("mfaToken"), "code": totpCode(t, secret, 30*time.Second), "audience": "seller"})
	expect(t, in, 200, "finance signs in with the authenticator")
	expect(t, get(t, "/api/v1/seller/shops/"+shopID+"/balance", mgr), 403, "manager doesn't see money")

	// Making a member finance needs their authenticator too; staff can't manage the team.
	var mgrID string
	for _, m := range get(t, team, owner).get("members").([]any) {
		if mm := m.(map[string]any); mm["role"] == "manager" {
			mgrID = mm["userId"].(string)
		}
	}
	expectCode(t, call(t, req{method: "PUT", path: team + "/members/" + mgrID + "/role", token: owner, body: map[string]any{"role": "finance"}}), 403, "authenticator_required", "finance without MFA")
	expect(t, call(t, req{method: "PUT", path: team + "/members/" + mgrID + "/role", token: owner, body: map[string]any{"role": "staff"}}), 200, "demote to staff")
	expect(t, get(t, team, mgr), 403, "staff can't see the team")

	// An invite can be cancelled; members can leave on their own.
	x := post(t, team+"/invites", owner, map[string]any{"email": "later@example.com", "role": "staff"})
	expect(t, x, 201, "invite by email")
	expect(t, call(t, req{method: "DELETE", path: team + "/invites/" + x.str("id"), token: owner}), 204, "revoke")
	expectCode(t, post(t, "/api/v1/seller/invites/preview", stranger, map[string]any{"token": inviteToken(t, x)}), 422, "invite_invalid", "revoked link")
	expect(t, call(t, req{method: "DELETE", path: team + "/members/" + mgrID, token: mgr}), 204, "staff leaves")
	expect(t, get(t, "/api/v1/seller/shops/"+shopID+"/orders", mgr), 404, "gone")
}

// latestSMS returns the newest text to a phone containing about.
func latestSMS(t *testing.T, phone, about string) string {
	t.Helper()
	var body string
	waitFor(t, "a text about "+about, 15*time.Second, func() bool {
		for _, m := range get(t, "/dev/messages?to="+urlEncode(phone), "").List {
			payload, _ := m.(map[string]any)["payload"].(map[string]any)
			if b, _ := payload["body"].(string); strings.Contains(b, about) {
				body = b
				return true
			}
		}
		return false
	})
	return body
}
