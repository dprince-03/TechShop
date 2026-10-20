package logistics

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/store"
)

// RiderJob is what a rider sees: first name, address and landmark — never the full contact.
type RiderJob struct {
	ID              uuid.UUID  `json:"id"`
	Kind            string     `json:"kind"`
	Status          string     `json:"status"`
	OrderNumber     string     `json:"orderNumber,omitempty"`
	RecipientName   string     `json:"recipientName"`
	Address         string     `json:"address"`
	Landmark        string     `json:"landmark,omitempty"`
	City            string     `json:"city"`
	StateCode       string     `json:"stateCode"`
	Latitude        *float64   `json:"latitude,omitempty"`
	Longitude       *float64   `json:"longitude,omitempty"`
	PickupName      string     `json:"pickupName,omitempty"`
	PickupAddress   string     `json:"pickupAddress,omitempty"`
	WindowStart     *time.Time `json:"windowStart,omitempty"`
	WindowEnd       *time.Time `json:"windowEnd,omitempty"`
	Attempts        int16      `json:"attempts"`
	RouteSequence   *int16     `json:"routeSequence,omitempty"`
	CanCallCustomer bool       `json:"canCallCustomer" doc:"Calls go through the masked call line; the number is never sent to the app"`
}

func riderJob(id uuid.UUID, kind, status string, number *string, recipient *string, ship []byte, lat, lng *float64, pickName, pickAddr *string,
	ws, we *time.Time, attempts int16, seq *int16) RiderJob {
	var a shipTo
	_ = json.Unmarshal(ship, &a)
	j := RiderJob{ID: id, Kind: kind, Status: status, OrderNumber: deref(number), RecipientName: deref(recipient), City: a.City, StateCode: a.StateCode,
		Latitude: lat, Longitude: lng, PickupName: deref(pickName), PickupAddress: deref(pickAddr), WindowStart: ws, WindowEnd: we, Attempts: attempts,
		RouteSequence: seq, CanCallCustomer: status == "picked_up" || status == "en_route"}
	j.Address = strings.TrimSpace(a.Line1 + " " + deref(a.Line2))
	j.Landmark = deref(a.Landmark)
	return j
}

// DispatchJob is a delivery job as dispatchers see it. It is built field by field from the
// database row so the delivery code's hash and attempt counter can never be serialised.
type DispatchJob struct {
	ID                uuid.UUID       `json:"id"`
	Kind              string          `json:"kind"`
	FulfilmentID      *uuid.UUID      `json:"fulfilmentId"`
	ReturnRequestID   *uuid.UUID      `json:"returnRequestId"`
	RiderID           *uuid.UUID      `json:"riderId"`
	ZoneID            uuid.UUID       `json:"zoneId"`
	Status            string          `json:"status"`
	PickupWarehouseID *uuid.UUID      `json:"pickupWarehouseId"`
	PickupAddress     json.RawMessage `json:"pickupAddress"`
	DropoffLatitude   *float64        `json:"dropoffLatitude"`
	DropoffLongitude  *float64        `json:"dropoffLongitude"`
	RouteSequence     *int16          `json:"routeSequence"`
	WindowStart       *time.Time      `json:"windowStart"`
	WindowEnd         *time.Time      `json:"windowEnd"`
	Attempts          int16           `json:"attempts"`
	AssignedBy        *uuid.UUID      `json:"assignedBy"`
	AssignedAt        *time.Time      `json:"assignedAt"`
	DeliveredAt       *time.Time      `json:"deliveredAt"`
	RecipientName     *string         `json:"recipientName"`
	OtpExpiresAt      *time.Time      `json:"otpExpiresAt"`
	OtpVerifiedAt     *time.Time      `json:"otpVerifiedAt"`
	ProofMethod       *string         `json:"proofMethod"`
	ProofFileID       *uuid.UUID      `json:"proofFileId"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

func dispatchJob(j store.LogisticsDeliveryJob) DispatchJob {
	return DispatchJob{ID: j.ID, Kind: j.Kind, FulfilmentID: j.FulfilmentID, ReturnRequestID: j.ReturnRequestID, RiderID: j.RiderID,
		ZoneID: j.ZoneID, Status: j.Status, PickupWarehouseID: j.PickupWarehouseID, PickupAddress: j.PickupAddress,
		DropoffLatitude: j.DropoffLatitude, DropoffLongitude: j.DropoffLongitude, RouteSequence: j.RouteSequence,
		WindowStart: j.WindowStart, WindowEnd: j.WindowEnd, Attempts: j.Attempts, AssignedBy: j.AssignedBy, AssignedAt: j.AssignedAt,
		DeliveredAt: j.DeliveredAt, RecipientName: j.RecipientName, OtpExpiresAt: j.OtpExpiresAt, OtpVerifiedAt: j.OtpVerifiedAt,
		ProofMethod: j.ProofMethod, ProofFileID: j.ProofFileID, CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt}
}

// BoardJob is a board row: the job plus the names the board shows.
type BoardJob struct {
	DispatchJob
	ZoneName       string  `json:"zoneName"`
	OrderNumber    *string `json:"orderNumber"`
	RiderFirstName *string `json:"riderFirstName"`
}

func boardJob(r store.LogisticsBoardRow) BoardJob {
	return BoardJob{DispatchJob: dispatchJob(store.LogisticsDeliveryJob{ID: r.ID, Kind: r.Kind, FulfilmentID: r.FulfilmentID,
		ReturnRequestID: r.ReturnRequestID, RiderID: r.RiderID, ZoneID: r.ZoneID, Status: r.Status, PickupWarehouseID: r.PickupWarehouseID,
		PickupAddress: r.PickupAddress, DropoffLatitude: r.DropoffLatitude, DropoffLongitude: r.DropoffLongitude, RouteSequence: r.RouteSequence,
		WindowStart: r.WindowStart, WindowEnd: r.WindowEnd, Attempts: r.Attempts, AssignedBy: r.AssignedBy, AssignedAt: r.AssignedAt,
		DeliveredAt: r.DeliveredAt, RecipientName: r.RecipientName, OtpExpiresAt: r.OtpExpiresAt, OtpVerifiedAt: r.OtpVerifiedAt,
		ProofMethod: r.ProofMethod, ProofFileID: r.ProofFileID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}),
		ZoneName: r.ZoneName, OrderNumber: r.OrderNumber, RiderFirstName: r.RiderFirstName}
}

// JobDetail is the dispatcher's view of a job with its event trail.
type JobDetail struct {
	RiderJob
	RiderID *uuid.UUID                     `json:"riderId"`
	ZoneID  uuid.UUID                      `json:"zoneId"`
	Proof   *string                        `json:"proofMethod"`
	Events  []store.LogisticsDeliveryEvent `json:"events"`
}
