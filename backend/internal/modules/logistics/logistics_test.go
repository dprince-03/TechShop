package logistics

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/store"
)

func TestRiderJobHidesContact(t *testing.T) {
	ship := []byte(`{"recipientName":"Adaeze Okafor","phone":"+2348031234567","line1":"12 Allen Avenue","landmark":"Opposite the bank","city":"Ikeja","stateCode":"LA"}`)
	first := "Adaeze"
	j := riderJob(uuid.New(), "delivery", "en_route", nil, &first, ship, nil, nil, nil, nil, nil, nil, 0, nil)
	b, _ := json.Marshal(j)
	if strings.Contains(string(b), "+234") || strings.Contains(string(b), "Okafor") {
		t.Fatalf("rider view leaks contact details: %s", b)
	}
	if j.Address != "12 Allen Avenue" || j.Landmark != "Opposite the bank" || !j.CanCallCustomer {
		t.Fatalf("rider view: %+v", j)
	}
}

// The delivery code's hash proves possession of the code; it must never reach any client.
func TestDispatchJobHidesDeliveryCode(t *testing.T) {
	j := store.LogisticsDeliveryJob{ID: uuid.New(), OtpHash: []byte("secret-hash"), OtpAttempts: 2}
	for name, v := range map[string]any{
		"job":   dispatchJob(j),
		"board": boardJob(store.LogisticsBoardRow{ID: j.ID, OtpHash: j.OtpHash, OtpAttempts: 2, ZoneName: "Ikeja"}),
	} {
		b, _ := json.Marshal(v)
		if strings.Contains(string(b), "otpHash") || strings.Contains(string(b), "otpAttempts") {
			t.Fatalf("%s view exposes delivery-code fields: %s", name, b)
		}
	}
}
