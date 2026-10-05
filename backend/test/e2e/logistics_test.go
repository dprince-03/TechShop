//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// riderLogin signs the seeded rider in with phone OTP on the logistics app, from the
// pre-approved phone.
func riderLogin(t *testing.T) string {
	t.Helper()
	_, phone, _, _, _ := account(t, "rider")
	before := peekCode(t, phone)
	r := post(t, "/api/v1/auth/otp/request", "", map[string]any{"phone": phone})
	expect(t, r, 200, "rider otp request")
	code := latestCode(t, phone, before)
	v := post(t, "/api/v1/auth/otp/verify", "", map[string]any{"challengeId": r.str("challengeId"), "code": code, "audience": "logistics",
		"deviceId": seed.Extra["riderDeviceId"], "deviceName": "Test Android"})
	expect(t, v, 200, "rider otp verify")
	tok := v.str("tokens", "accessToken")
	if tok == "" {
		t.Fatalf("rider tokens missing: %s", v.Raw)
	}
	return tok
}

// pickAndPack releases a wave in the Ikeja warehouse, scans this order's lines (serialised ones
// by IMEI) and packs the fulfilment. It returns the pack response.
func pickAndPack(t *testing.T, number string, serialFor map[string]string) resp {
	t.Helper()
	wh := staffLogin(t, "warehouse_lead")
	ops := staffLogin(t, "ops_manager")
	o := get(t, "/api/v1/staff/orders/"+number, ops)
	expect(t, o, 200, "staff order")
	f := o.get("fulfilments", "0").(map[string]any)
	fulfilmentID := f["id"].(string)
	lines := map[string]bool{}
	for _, l := range f["lines"].([]any) {
		lines[l.(map[string]any)["id"].(string)] = true
	}

	w := post(t, "/api/v1/staff/wms/waves", wh, map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"]})
	expect(t, w, 201, "release wave")
	listID := w.str("id")
	expect(t, post(t, "/api/v1/staff/wms/pick-lists/"+listID+"/claim", wh, nil), 200, "claim pick list")
	found := 0
	for _, it := range w.get("items").([]any) {
		m := it.(map[string]any)
		lineID := m["orderLineId"].(string)
		if !lines[lineID] {
			continue
		}
		found++
		body := map[string]any{"orderLineId": lineID, "quantity": m["quantity"]}
		if m["isSerialised"] == true {
			expectCode(t, post(t, "/api/v1/staff/wms/pick-lists/"+listID+"/scan", wh, body), 422, "serial_required", "serialised needs IMEI")
			body["serial"] = "359999999999999"
			expectCode(t, post(t, "/api/v1/staff/wms/pick-lists/"+listID+"/scan", wh, body), 409, "unknown_device", "unknown IMEI")
			body["serial"] = serialFor[m["sku"].(string)]
		}
		expect(t, post(t, "/api/v1/staff/wms/pick-lists/"+listID+"/scan", wh, body), 200, "scan line")
		over := map[string]any{"orderLineId": lineID, "quantity": 1}
		if m["isSerialised"] != true {
			expectCode(t, post(t, "/api/v1/staff/wms/pick-lists/"+listID+"/scan", wh, over), 409, "over_pick", "over-pick refused")
		}
	}
	if found != len(lines) {
		t.Fatalf("wave has %d of the order's %d lines", found, len(lines))
	}
	p := post(t, "/api/v1/staff/wms/pack", wh, map[string]any{"fulfilmentId": fulfilmentID, "weightGrams": 820, "boxCode": "B2"})
	expect(t, p, 201, "pack")
	return p
}

// From paid order to the doorstep: wave, pick (IMEI bound), pack, delivery job, device-approved
// rider shift, assignment, label scan, delivery code, offline sync with replay, live location,
// delivered with the code — and the ledger stays balanced.
func TestWarehouseToDoorstep(t *testing.T) {
	wh := staffLogin(t, "warehouse_lead")
	// Receive one serialised unit with a valid IMEI for the gaming handheld (p18).
	imei := luhnIMEI(fmt.Sprintf("35%012d", time.Now().UnixNano()%1e12))
	expect(t, post(t, "/api/v1/staff/inventory/device-units", wh, map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"], "variantId": seed.Extra["variant_p18"],
		"condition": "new", "units": []any{map[string]any{"imei": imei}}}), 201, "receive imei")

	cust := customerLogin(t, "customer")
	number, _ := buyAndPay(t, cust, seed.Extra["listing_p18"])

	sku := ""
	d := get(t, "/api/v1/me/orders/"+number, cust)
	for _, f := range d.get("fulfilments").([]any) {
		for _, l := range f.(map[string]any)["lines"].([]any) {
			sku = l.(map[string]any)["sku"].(string)
		}
	}
	p := pickAndPack(t, number, map[string]string{sku: imei})
	if p.str("handoff") != "rider" || p.str("jobId") == "" {
		t.Fatalf("a Lagos address should get a rider job: %s", truncate(string(p.Raw)))
	}
	jobID, label := p.str("jobId"), p.str("parcel", "labelCode")
	if s := get(t, "/api/v1/me/orders/"+number, cust).str("status"); s != "processing" {
		t.Fatalf("packed order should be processing, is %s", s)
	}

	disp := staffLogin(t, "dispatcher")
	b := get(t, "/api/v1/dispatch/board?status=unassigned", disp)
	expect(t, b, 200, "board")
	if !strings.Contains(string(b.Raw), jobID) || strings.Contains(string(b.Raw), "otpHash\":\"") {
		t.Fatalf("board: %s", truncate(string(b.Raw)))
	}
	_, _, _, _, riderID := account(t, "rider")
	_, _, _, _, custID := account(t, "customer")
	rider := riderLogin(t)
	_ = post(t, "/api/v1/rider/shift/end", rider, nil) // a clean start if an earlier run left a shift open

	// The rider must be on shift (with location consent) before getting work.
	expectCode(t, post(t, "/api/v1/dispatch/jobs/"+jobID+"/assign", disp, map[string]any{"riderId": riderID}), 409, "rider_off_shift", "assign off shift")
	expectCode(t, post(t, "/api/v1/rider/shift/start", rider, map[string]any{"locationConsent": false}), 422, "consent_required", "no consent")
	expect(t, post(t, "/api/v1/rider/shift/start", rider, map[string]any{"locationConsent": true, "latitude": 6.6018, "longitude": 3.3515}), 201, "start shift")
	expect(t, post(t, "/api/v1/dispatch/jobs/"+jobID+"/assign", disp, map[string]any{"riderId": riderID}), 200, "assign")
	// Customers aren't riders; a non-rider gets 403 on rider routes.
	expect(t, get(t, "/api/v1/rider/jobs", cust), 401, "customer token on rider app")
	expectCode(t, post(t, "/api/v1/dispatch/jobs/"+jobID+"/assign", rider, map[string]any{"riderId": custID}), 403, "forbidden", "rider can't dispatch")

	jobs := get(t, "/api/v1/rider/jobs", rider)
	expect(t, jobs, 200, "rider jobs")
	if !strings.Contains(string(jobs.Raw), jobID) || strings.Contains(string(jobs.Raw), "+234") {
		t.Fatalf("rider jobs must include the job and never the customer's number: %s", truncate(string(jobs.Raw)))
	}

	act := func(kind string, extra map[string]any) resp {
		body := map[string]any{"clientEventId": newUUID(), "kind": kind, "occurredAt": time.Now().UTC().Format(time.RFC3339), "latitude": 6.6, "longitude": 3.35}
		for k, v := range extra {
			body[k] = v
		}
		return post(t, "/api/v1/rider/jobs/"+jobID+"/actions", rider, body)
	}
	expectCode(t, act("picked_up", map[string]any{"labelCode": "TSL-WRONG1"}), 409, "label_mismatch", "wrong parcel scanned")
	expect(t, act("picked_up", map[string]any{"labelCode": label}), 200, "picked up")
	_, custPhone, _, _, _ := account(t, "customer")
	otp := latestCodeContaining(t, custPhone, "delivery code")

	// Offline sync: the same batch twice applies once.
	ev := map[string]any{"clientEventId": newUUID(), "jobId": jobID, "kind": "en_route", "occurredAt": time.Now().UTC().Format(time.RFC3339)}
	s1 := post(t, "/api/v1/rider/events/batch", rider, map[string]any{"events": []any{ev}})
	expect(t, s1, 200, "sync")
	s2 := post(t, "/api/v1/rider/events/batch", rider, map[string]any{"events": []any{ev}})
	if s1.str("results", "0", "status") != "accepted" || s2.str("results", "0", "status") != "duplicate" {
		t.Fatalf("sync replay: %s / %s", s1.Raw, s2.Raw)
	}
	expect(t, post(t, "/api/v1/rider/location", rider, map[string]any{"points": []any{
		map[string]any{"latitude": 6.59, "longitude": 3.34, "recordedAt": time.Now().Add(-30 * time.Second).UTC().Format(time.RFC3339)},
		map[string]any{"latitude": 6.595, "longitude": 3.345, "recordedAt": time.Now().UTC().Format(time.RFC3339)}}}), 204, "location")
	rs := get(t, "/api/v1/dispatch/riders", disp)
	if !strings.Contains(string(rs.Raw), "6.595") {
		t.Fatalf("dispatch should see the rider's latest position: %s", truncate(string(rs.Raw)))
	}

	// Delivery needs proof; a wrong code is refused and counted.
	expectCode(t, act("delivered", nil), 422, "proof_required", "no proof")
	wrong := "000000"
	if otp == wrong {
		wrong = "111111"
	}
	expectCode(t, act("delivered", map[string]any{"otp": wrong}), 422, "otp_wrong", "wrong delivery code")
	expect(t, act("delivered", map[string]any{"otp": otp, "recipientName": "Adaeze"}), 200, "delivered with code")
	waitOrderStatus(t, cust, number, "delivered")
	dj := get(t, "/api/v1/dispatch/jobs/"+jobID, disp)
	if dj.str("status") != "delivered" || dj.str("proofMethod") != "otp" || len(dj.get("events").([]any)) < 5 {
		t.Fatalf("job trail: %s", truncate(string(dj.Raw)))
	}
	expect(t, post(t, "/api/v1/rider/shift/end", rider, nil), 200, "end shift")
	expectCode(t, post(t, "/api/v1/rider/location", rider, map[string]any{"points": []any{
		map[string]any{"latitude": 6.6, "longitude": 3.3, "recordedAt": time.Now().UTC().Format(time.RFC3339)}}}), 409, "off_shift", "no tracking off shift")
	trialBalanced(t)
}

// Outside the delivery zones the parcel goes by carrier: booking at packing, manifest signed at
// hand-over, signed tracking events (deduplicated) move it to delivered.
func TestCarrierShipment(t *testing.T) {
	cust := customerLogin(t, "customer_b")
	a := post(t, "/api/v1/me/addresses", cust, map[string]any{"recipientName": "Halima Bello", "phone": "08031234570", "line1": "5 Murtala Mohammed Way",
		"city": "Kano", "stateCode": "KN"})
	expect(t, a, 201, "Kano address")
	clearCart(t, cust)
	addToCart(t, cust, seed.Extra["listing_p21"], 1)
	pv := preview(t, cust, a.str("id"), "")
	expect(t, pv, 200, "preview")
	o := placeOrder(t, cust, a.str("id"), pv.str("priceHash"), "paystack", newKey(), "")
	expect(t, o, 201, "order to Kano")
	number := o.str("orderNumber")
	fakePay(t, "paystack", o.str("payment", "reference"), "success")
	waitOrderStatus(t, cust, number, "paid")

	p := pickAndPack(t, number, nil)
	if p.str("handoff") != "carrier" || p.str("shipment", "trackingNumber") == "" {
		t.Fatalf("Kano should ship by carrier: %s", truncate(string(p.Raw)))
	}
	tracking, label := p.str("shipment", "trackingNumber"), p.str("parcel", "labelCode")

	wh := staffLogin(t, "warehouse_lead")
	carriers := get(t, "/api/v1/staff/wms/carriers", wh)
	carrierID := ""
	for _, c := range carriers.List {
		if c.(map[string]any)["code"] == "fake" {
			carrierID = c.(map[string]any)["id"].(string)
		}
	}
	m := post(t, "/api/v1/staff/wms/manifests", wh, map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"], "carrierId": carrierID, "labels": []string{label}})
	expect(t, m, 201, "manifest")
	expectCode(t, post(t, "/api/v1/staff/wms/manifests", wh, map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"], "carrierId": carrierID, "labels": []string{label}}),
		409, "already_manifested", "parcel already manifested")
	expect(t, post(t, "/api/v1/staff/wms/manifests/"+m.str("id")+"/sign", wh, map[string]any{"signedByName": "Courier Driver"}), 200, "sign manifest")

	send := func(status string) resp {
		ev := post(t, "/dev/carriers/fake/"+tracking+"/"+status, "", nil)
		expect(t, ev, 200, "fake carrier event")
		return call(t, req{method: "POST", path: "/api/v1/webhooks/carriers/fake", headers: map[string]string{"x-carrier-signature": ev.str("signature")}, raw: ev.str("raw")})
	}
	expectCode(t, call(t, req{method: "POST", path: "/api/v1/webhooks/carriers/fake", headers: map[string]string{"x-carrier-signature": "bad"},
		raw: `{"eventId":"x","trackingNumber":"` + tracking + `","status":"delivered"}`}), 401, "bad_signature", "forged carrier event")
	expect(t, send("in_transit"), 200, "in transit")
	waitOrderStatus(t, cust, number, "shipped")
	ev := post(t, "/dev/carriers/fake/"+tracking+"/delivered", "", nil)
	hdr := map[string]string{"x-carrier-signature": ev.str("signature")}
	expect(t, call(t, req{method: "POST", path: "/api/v1/webhooks/carriers/fake", headers: hdr, raw: ev.str("raw")}), 200, "delivered")
	dup := call(t, req{method: "POST", path: "/api/v1/webhooks/carriers/fake", headers: hdr, raw: ev.str("raw")})
	if dup.str("status") != "duplicate" {
		t.Fatalf("carrier event replay should be a duplicate: %s", dup.Raw)
	}
	waitOrderStatus(t, cust, number, "delivered")
	trialBalanced(t)
}

// Delivery estimate (public) and zone pricing at checkout.
func TestDeliveryEstimate(t *testing.T) {
	la := get(t, "/api/v1/delivery/estimate?state=LA&weightGrams=800", "")
	expect(t, la, 200, "estimate Lagos")
	if la.get("covered") != true || la.num("feeKobo") != 250000 {
		t.Fatalf("Lagos ≤2 kg band: %s", la.Raw)
	}
	heavy := get(t, "/api/v1/delivery/estimate?state=LA&weightGrams=5000&fulfilledBy=seller", "")
	if heavy.num("feeKobo") != 450000 {
		t.Fatalf("Lagos ≤10 kg seller band: %s", heavy.Raw)
	}
	kn := get(t, "/api/v1/delivery/estimate?state=KN", "")
	if kn.get("covered") != false {
		t.Fatalf("Kano isn't an own-rider zone: %s", kn.Raw)
	}
}
