//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"
)

// The stolen-device blocklist: only valid IMEIs, no duplicates, blocked phones can't enter stock,
// and only devices.block may change it.
func TestDeviceBlocklist(t *testing.T) {
	t.Parallel()
	ra := staffLogin(t, "risk_analyst")
	imei := luhnIMEI(fmt.Sprintf("35%012d", (time.Now().UnixNano()+31)%1e12))
	bad := imei[:14] + fmt.Sprint((int(imei[14]-'0')+1)%10)

	expectCode(t, post(t, "/api/v1/staff/risk/devices", ra, map[string]any{"imei": bad, "reason": "reported_stolen", "source": "Customer report"}), 422, "imei_invalid", "typo IMEI")
	expect(t, post(t, "/api/v1/staff/risk/devices", staffLogin(t, "support_agent"), map[string]any{"imei": imei, "reason": "reported_stolen", "source": "x"}), 403, "support can't block")
	b := post(t, "/api/v1/staff/risk/devices", ra, map[string]any{"imei": imei, "reason": "reported_stolen", "source": "Customer report", "reference": "Police extract 4471"})
	expect(t, b, 201, "block")
	expectCode(t, post(t, "/api/v1/staff/risk/devices", ra, map[string]any{"imei": imei, "reason": "fraud", "source": "again"}), 409, "already_blocked", "block twice")

	c := get(t, "/api/v1/staff/risk/devices/check?imei="+imei, ra)
	if c.get("blocked") != true || c.get("luhnOk") != true || c.str("reason") != "reported_stolen" {
		t.Fatalf("check blocked IMEI: %s", c.Raw)
	}
	if c := get(t, "/api/v1/staff/risk/devices/check?imei="+bad, ra); c.get("luhnOk") != false || c.get("blocked") != false {
		t.Fatalf("check typo IMEI: %s", c.Raw)
	}
	if l := get(t, "/api/v1/staff/risk/devices?imei="+imei, ra); len(l.List) != 1 {
		t.Fatalf("blocklist search: %s", l.Raw)
	}
	recv := post(t, "/api/v1/staff/inventory/device-units", staffLogin(t, "warehouse_lead"), map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"],
		"variantId": seed.Extra["variant_p18"], "condition": "new", "units": []any{map[string]any{"imei": imei}}})
	if recv.Status < 400 {
		t.Fatalf("a blocklisted IMEI must not enter stock: %d %s", recv.Status, recv.Raw)
	}
	if s := get(t, "/api/v1/staff/risk/summary", ra); s.num("blockedDevices") < 1 {
		t.Fatalf("summary: %s", s.Raw)
	}

	expectCode(t, post(t, "/api/v1/staff/risk/devices/"+imei+"/unblock", ra, map[string]any{"note": ""}), 422, "note_required", "unblock without a reason")
	expect(t, post(t, "/api/v1/staff/risk/devices/"+imei+"/unblock", ra, map[string]any{"note": "Phone recovered and returned to owner"}), 204, "unblock")
	if c := get(t, "/api/v1/staff/risk/devices/check?imei="+imei, ra); c.get("blocked") != false {
		t.Fatalf("unblocked IMEI still blocked: %s", c.Raw)
	}
	expect(t, get(t, "/api/v1/staff/risk/cases", staffLogin(t, "support_agent")), 403, "support can't see risk cases")
}
