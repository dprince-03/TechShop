package logistics

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
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
