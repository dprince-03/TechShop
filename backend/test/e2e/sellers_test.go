//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
)

// sellerSignup creates a brand-new seller login (OTP sign-up on the seller audience).
func sellerSignup(t *testing.T, first, last string) string {
	t.Helper()
	phone := randomPhone()
	r := post(t, "/api/v1/auth/otp/request", "", map[string]any{"phone": phone})
	expect(t, r, 200, "seller otp request")
	code := latestCode(t, phone, "")
	v := post(t, "/api/v1/auth/otp/verify", "", map[string]any{"challengeId": r.str("challengeId"), "code": code, "audience": "seller"})
	expect(t, v, 200, "seller otp verify")
	s := post(t, "/api/v1/auth/otp/signup", "", map[string]any{"signupToken": v.str("signupToken"), "firstName": first, "lastName": last, "audience": "seller", "deviceName": "e2e"})
	expect(t, s, 200, "seller signup")
	return s.str("accessToken")
}

// kycFile uploads a finished KYC document and returns its file ID.
func kycFile(t *testing.T, token, name string) string {
	t.Helper()
	u := post(t, "/api/v1/files/uploads", token, map[string]any{"purpose": "kyc_document", "contentType": "image/jpeg", "sizeBytes": 120000, "name": name})
	expect(t, u, 201, "kyc upload "+name)
	id := u.str("fileId")
	expect(t, post(t, "/api/v1/files/"+id+"/complete", token, map[string]any{}), 200, "complete "+name)
	return id
}

func TestSellerOnboardingAndKycReview(t *testing.T) {
	t.Parallel()
	owner := sellerSignup(t, "Adaeze", "Okafor")
	if r := get(t, "/api/v1/seller/shops", owner); r.Status != 200 || len(r.List) != 0 {
		t.Fatalf("new seller should have no shops: %d %s", r.Status, truncate(string(r.Raw)))
	}
	name := fmt.Sprintf("E2E Gadgets %s", newUUID()[:6])
	shop := post(t, "/api/v1/seller/shops", owner, map[string]any{"type": "business", "displayName": name, "stateCode": "LA", "city": "Ikeja"})
	expect(t, shop, 201, "register shop")
	shopID := shop.str("id")
	if shop.str("status") != "pending_kyc" || shop.str("role") != "owner" || !strings.HasPrefix(shop.str("slug"), "e2e-gadgets-") {
		t.Fatalf("shop: %s", shop.Raw)
	}
	expectCode(t, post(t, "/api/v1/seller/shops", owner, map[string]any{"type": "business", "displayName": "Second Shop", "stateCode": "LA", "city": "Ikeja"}), 409, "shop_exists", "one shop per owner")
	base := "/api/v1/seller/shops/" + shopID
	if o := get(t, base+"/onboarding", owner); o.str("nextStep") != "submit_kyc" {
		t.Fatalf("onboarding: %s", o.Raw)
	}

	// KYC validation: the CAC number and certificate are required for businesses; files must be
	// the seller's own KYC uploads; a NIN is 11 digits.
	idFront, selfie := kycFile(t, owner, "id.jpg"), kycFile(t, owner, "selfie.jpg")
	docs := []map[string]any{{"kind": "id_front", "fileId": idFront}, {"kind": "selfie", "fileId": selfie}}
	nin := "1234 5678 901"
	expectCode(t, post(t, base+"/kyc", owner, map[string]any{"idType": "nin", "idNumber": nin, "documents": docs}), 422, "cac_required", "business without CAC")
	expectCode(t, post(t, base+"/kyc", owner, map[string]any{"idType": "nin", "idNumber": nin, "cacRcNumber": "RC 1234567", "documents": docs}), 422, "missing_document", "no CAC certificate")
	stranger := kycFile(t, customerLogin(t, "customer"), "other.jpg")
	withStranger := append([]map[string]any{{"kind": "cac_certificate", "fileId": stranger}}, docs...)
	expectCode(t, post(t, base+"/kyc", owner, map[string]any{"idType": "nin", "idNumber": nin, "cacRcNumber": "RC 1234567", "documents": withStranger}), 422, "unknown_file", "someone else's file")
	docs = append(docs, map[string]any{"kind": "cac_certificate", "fileId": kycFile(t, owner, "cac.pdf")})
	expectCode(t, post(t, base+"/kyc", owner, map[string]any{"idType": "nin", "idNumber": "1234567", "cacRcNumber": "RC 1234567", "documents": docs}), 422, "invalid_id_number", "short NIN")
	k := post(t, base+"/kyc", owner, map[string]any{"idType": "nin", "idNumber": nin, "cacRcNumber": "RC 1234567", "documents": docs})
	expect(t, k, 201, "submit kyc")
	kycID := k.str("id")
	if k.str("status") != "submitted" || k.str("idNumberLast4") != "8901" || strings.Contains(string(k.Raw), "12345678901") {
		t.Fatalf("kyc response: %s", k.Raw)
	}
	expectCode(t, post(t, base+"/kyc", owner, map[string]any{"idType": "nin", "idNumber": nin, "cacRcNumber": "RC 1234567", "documents": docs}), 409, "kyc_in_progress", "resubmit while pending")
	if o := get(t, base+"/onboarding", owner); o.str("nextStep") != "add_bank_account" {
		t.Fatalf("onboarding after kyc: %s", o.Raw)
	}

	// Bank accounts: name enquiry (dev fake: 0000000000 not found, 9… a different holder).
	if b := get(t, "/api/v1/seller/banks", owner); b.Status != 200 || len(b.List) == 0 {
		t.Fatalf("banks: %d", b.Status)
	}
	expectCode(t, post(t, base+"/bank-accounts", owner, map[string]any{"bankCode": "058", "accountNumber": "0000000000", "accountName": name}), 422, "account_not_found", "unknown account")
	odd := post(t, base+"/bank-accounts", owner, map[string]any{"bankCode": "058", "accountNumber": "9123456789", "accountName": name})
	expect(t, odd, 201, "mismatched account")
	if odd.str("nameMatch") != "mismatch" || odd.get("isDefault") != true || odd.str("accountNumberLast4") != "6789" {
		t.Fatalf("mismatched account: %s", odd.Raw)
	}
	good := post(t, base+"/bank-accounts", owner, map[string]any{"bankCode": "044", "accountNumber": "0123456789", "accountName": name})
	expect(t, good, 201, "matching account")
	if good.str("nameMatch") != "match" || good.get("isDefault") != false || strings.Contains(string(good.Raw), "0123456789") {
		t.Fatalf("matching account: %s", good.Raw)
	}
	expectCode(t, post(t, base+"/bank-accounts", owner, map[string]any{"bankCode": "044", "accountNumber": "0123456789", "accountName": name}), 409, "account_exists", "duplicate account")
	expectCode(t, call(t, req{method: "DELETE", path: base + "/bank-accounts/" + odd.str("id"), token: owner}), 409, "cannot_delete", "delete default")
	expect(t, post(t, base+"/bank-accounts/"+good.str("id")+"/default", owner, nil), 204, "set default")
	expect(t, call(t, req{method: "DELETE", path: base + "/bank-accounts/" + odd.str("id"), token: owner}), 204, "delete old account")
	if o := get(t, base+"/onboarding", owner); o.str("nextStep") != "await_review" || len(o.get("bankAccounts").([]any)) != 1 {
		t.Fatalf("onboarding after bank: %s", o.Raw)
	}

	// Another seller can't see or touch this shop (404, not 403: existence isn't revealed).
	other := sellerSignup(t, "Musa", "Bello")
	expect(t, get(t, base+"/onboarding", other), 404, "other seller's shop")
	expect(t, post(t, base+"/bank-accounts", other, map[string]any{"bankCode": "044", "accountNumber": "0123456780", "accountName": "Musa Bello"}), 404, "other seller adds bank")
	// Other audiences can't use Seller Centre or staff routes.
	if r := get(t, "/api/v1/seller/shops", customerLogin(t, "customer")); r.Status != 401 && r.Status != 403 {
		t.Fatalf("market token on seller route: %d", r.Status)
	}
	if r := get(t, "/api/v1/staff/vendors", owner); r.Status != 401 && r.Status != 403 {
		t.Fatalf("seller token on staff route: %d", r.Status)
	}

	// Staff review. Without vendors.view: refused.
	expect(t, get(t, "/api/v1/staff/vendors", staffLogin(t, "dispatcher")), 403, "dispatcher lists vendors")
	rev := staffLogin(t, "kyc_reviewer")
	list := get(t, "/api/v1/staff/vendors?kycStatus=submitted&q="+urlEncode(name[len(name)-6:]), rev)
	expect(t, list, 200, "vendor list")
	if items, _ := list.get("items").([]any); len(items) != 1 {
		t.Fatalf("vendor list: %s", truncate(string(list.Raw)))
	}
	v := get(t, "/api/v1/staff/vendors/"+shopID, rev)
	expect(t, v, 200, "vendor detail")
	raw := string(v.Raw)
	if v.str("kyc", "idNumberLast4") != "8901" || v.str("legalHolder") != name || strings.Contains(raw, "12345678901") ||
		strings.Contains(strings.ToLower(raw), "hmac") || strings.Contains(strings.ToLower(raw), "encrypted") {
		t.Fatalf("vendor detail: %s", truncate(raw))
	}
	expectCode(t, post(t, "/api/v1/staff/vendors/kyc/"+kycID+"/review", rev, map[string]any{"decision": "needs_more_info"}), 422, "note_required", "no note")
	ok := post(t, "/api/v1/staff/vendors/kyc/"+kycID+"/review", rev, map[string]any{"decision": "approved"})
	expect(t, ok, 200, "approve")
	expectCode(t, post(t, "/api/v1/staff/vendors/kyc/"+kycID+"/review", rev, map[string]any{"decision": "rejected", "note": "Too late"}), 409, "bad_transition", "review twice")
	if o := get(t, base+"/onboarding", owner); o.str("nextStep") != "done" || o.str("shop", "status") != "active" {
		t.Fatalf("onboarding after approval: %s", o.Raw)
	}

	// Revealing the full number needs a reason (and the sensitive-permission MFA check).
	expect(t, post(t, "/api/v1/staff/vendors/kyc/"+kycID+"/reveal", rev, map[string]any{"reason": "short"}), 422, "reveal without reason")
	full := post(t, "/api/v1/staff/vendors/kyc/"+kycID+"/reveal", rev, map[string]any{"reason": "Matching NIN against the CAC director record"})
	expect(t, full, 200, "reveal")
	if full.str("idNumber") != "12345678901" {
		t.Fatalf("reveal: %s", full.Raw)
	}
}

// A rejected seller can fix and resubmit; approval needs a verified bank account.
func TestSellerKycRejectAndResubmit(t *testing.T) {
	t.Parallel()
	owner := sellerSignup(t, "Chinedu", "Eze")
	shop := post(t, "/api/v1/seller/shops", owner, map[string]any{"type": "individual", "displayName": "Chinedu Phones " + newUUID()[:4], "stateCode": "EN", "city": "Enugu"})
	expect(t, shop, 201, "register individual")
	base := "/api/v1/seller/shops/" + shop.str("id")
	docs := []map[string]any{{"kind": "id_front", "fileId": kycFile(t, owner, "id.jpg")}, {"kind": "selfie", "fileId": kycFile(t, owner, "selfie.jpg")}}
	k := post(t, base+"/kyc", owner, map[string]any{"idType": "passport", "idNumber": "A12345678", "documents": docs})
	expect(t, k, 201, "individual kyc")

	rev := staffLogin(t, "kyc_reviewer")
	expectCode(t, post(t, "/api/v1/staff/vendors/kyc/"+k.str("id")+"/review", rev, map[string]any{"decision": "approved"}), 409, "bank_account_missing", "approve without bank")
	no := post(t, "/api/v1/staff/vendors/kyc/"+k.str("id")+"/review", rev, map[string]any{"decision": "rejected", "note": "The selfie is blurred."})
	expect(t, no, 200, "reject")
	o := get(t, base+"/onboarding", owner)
	if o.str("nextStep") != "fix_kyc" || o.str("kyc", "note") != "The selfie is blurred." {
		t.Fatalf("after rejection: %s", o.Raw)
	}
	// Individuals' accounts are matched against the owner's own name.
	b := post(t, base+"/bank-accounts", owner, map[string]any{"bankCode": "058", "accountNumber": "0198765432", "accountName": "Eze Chinedu"})
	expect(t, b, 201, "individual bank")
	if b.str("nameMatch") != "match" {
		t.Fatalf("individual name match: %s", b.Raw)
	}
	docs[1] = map[string]any{"kind": "selfie", "fileId": kycFile(t, owner, "selfie2.jpg")}
	expect(t, post(t, base+"/kyc", owner, map[string]any{"idType": "passport", "idNumber": "A12345678", "documents": docs}), 201, "resubmit")
	if o := get(t, base+"/onboarding", owner); o.str("nextStep") != "await_review" {
		t.Fatalf("after resubmit: %s", o.Raw)
	}
}
