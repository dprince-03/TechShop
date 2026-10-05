// Package logistics runs last-mile delivery (docs/mobile.md §4): zones and rates, riders and
// their approved devices, shifts, delivery jobs with a delivery code (OTP) or photo proof,
// idempotent offline event sync, live location and the dispatch board.
package logistics

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/notify"
	"github.com/dprince-03/techshop/backend/internal/sales"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/pkg/crypto"
)

const (
	otpTTL         = 24 * time.Hour // the delivery code is valid for a day after pickup
	maxOTPAttempts = 5
	maxAttempts    = 5 // delivery attempts before the parcel goes back to the hub
)

// Service is the logistics module.
type Service struct {
	d      *kit.Deps
	sales  *sales.Service
	notify *notify.Service
}

// New creates the logistics service and plugs zone pricing into sales.
func New(d *kit.Deps, sl *sales.Service, n *notify.Service) *Service {
	s := &Service{d: d, sales: sl, notify: n}
	sl.Delivery = s
	return s
}

// Name implements kit.Module.
func (s *Service) Name() string { return "logistics" }

// ---- Zones and rates

// Quote implements sales.DeliveryQuoter: the zone's rate band for the parcel weight.
func (s *Service) Quote(ctx context.Context, q *store.Queries, fulfilledBy, state, lga string, weight int32) (int64, bool) {
	r, _, ok := s.estimate(ctx, q, fulfilledBy, state, lga, weight)
	return r.FeeKobo, ok
}

func (s *Service) estimate(ctx context.Context, q *store.Queries, fulfilledBy, state, lga string, weight int32) (store.LogisticsDeliveryRate, store.LogisticsDeliveryZone, bool) {
	z, err := q.LogisticsResolveZone(ctx, store.LogisticsResolveZoneParams{StateCode: state, Lga: lga})
	if err != nil {
		return store.LogisticsDeliveryRate{}, z, false
	}
	if weight <= 0 {
		weight = 1
	}
	r, err := q.LogisticsRate(ctx, store.LogisticsRateParams{ZoneID: z.ID, FulfilledBy: fulfilledBy, Weight: weight})
	if err != nil {
		return r, z, false
	}
	return r, z, true
}

// ---- Delivery jobs

// shipTo is the address snapshot on an order.
type shipTo struct {
	RecipientName string   `json:"recipientName"`
	Phone         string   `json:"phone"`
	Line1         string   `json:"line1"`
	Line2         *string  `json:"line2"`
	Landmark      *string  `json:"landmark"`
	City          string   `json:"city"`
	LGA           *string  `json:"lga"`
	StateCode     string   `json:"stateCode"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
}

// Covers reports whether own riders deliver to an order's address (otherwise a carrier ships it).
func (s *Service) Covers(ctx context.Context, q *store.Queries, shipToJSON []byte) bool {
	var a shipTo
	if json.Unmarshal(shipToJSON, &a) != nil {
		return false
	}
	_, err := q.LogisticsResolveZone(ctx, store.LogisticsResolveZoneParams{StateCode: a.StateCode, Lga: deref(a.LGA)})
	return err == nil
}

// CreateDeliveryJob opens an unassigned delivery job for a packed fulfilment (called by the
// warehouse in the packing transaction).
func (s *Service) CreateDeliveryJob(ctx context.Context, tx *uow.Tx, fulfilmentID uuid.UUID) (store.LogisticsDeliveryJob, error) {
	if j, err := tx.Q.LogisticsJobForFulfilment(ctx, &fulfilmentID); err == nil {
		return j, nil // already has one
	}
	info, err := tx.Q.LogisticsFulfilmentDeliveryInfo(ctx, fulfilmentID)
	if err != nil {
		return store.LogisticsDeliveryJob{}, httpx.NotFound("Fulfilment")
	}
	var a shipTo
	if err := json.Unmarshal(info.ShipTo, &a); err != nil {
		return store.LogisticsDeliveryJob{}, err
	}
	z, err := tx.Q.LogisticsResolveZone(ctx, store.LogisticsResolveZoneParams{StateCode: a.StateCode, Lga: deref(a.LGA)})
	if err != nil {
		return store.LogisticsDeliveryJob{}, httpx.Conflict("no_delivery_zone", "No delivery zone covers this address; ship it with a carrier.")
	}
	wh := info.WarehouseID
	if wh == nil {
		wh = info.ReservedWarehouseID
	}
	var pickup []byte
	if wh == nil {
		pickup, _ = json.Marshal(map[string]any{"note": "collect from seller"})
	}
	first := strings.Fields(a.RecipientName)
	name := a.RecipientName
	if len(first) > 0 {
		name = first[0] // riders see the first name only
	}
	start := time.Now()
	end := start.Add(8 * time.Hour)
	j, err := tx.Q.LogisticsCreateJob(ctx, store.LogisticsCreateJobParams{Kind: "delivery", FulfilmentID: &fulfilmentID, ZoneID: z.ID, PickupWarehouseID: wh,
		PickupAddress: pickup, DropoffLatitude: a.Latitude, DropoffLongitude: a.Longitude, WindowStart: &start, WindowEnd: &end, RecipientName: &name})
	if err != nil {
		return j, httpx.DB(err, "delivery job")
	}
	tx.Notify("realtime", map[string]any{"topic": "dispatch", "event": "job", "data": map[string]any{"jobId": j.ID, "status": j.Status, "zoneId": z.ID}})
	return j, nil
}

// Assign gives a job to a rider on shift.
func (s *Service) Assign(ctx context.Context, jobID, riderID, actor uuid.UUID, sequence *int16) (store.LogisticsDeliveryJob, error) {
	var out store.LogisticsDeliveryJob
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		r, err := tx.Q.LogisticsGetRider(ctx, riderID)
		if err != nil {
			return httpx.Invalid("not_a_rider", "That person isn't a rider.")
		}
		if r.Status == "suspended" {
			return httpx.Conflict("rider_suspended", "This rider is suspended.")
		}
		if _, err := tx.Q.LogisticsOpenShift(ctx, riderID); err != nil {
			return httpx.Conflict("rider_off_shift", "This rider isn't on shift.")
		}
		j, err := tx.Q.LogisticsGetJobForUpdate(ctx, jobID)
		if err != nil {
			return httpx.NotFound("Job")
		}
		if j.Status == "failed" && j.Attempts >= maxAttempts {
			return httpx.Conflict("too_many_attempts", "This parcel has failed too often; return it to the hub.")
		}
		out, err = tx.Q.LogisticsAssignJob(ctx, store.LogisticsAssignJobParams{ID: jobID, RiderID: &riderID, AssignedBy: &actor, RouteSequence: sequence})
		if err != nil {
			return httpx.Conflict("not_assignable", "Only unassigned, assigned or failed jobs can be (re)assigned.")
		}
		if _, err := tx.Q.LogisticsInsertEvent(ctx, store.LogisticsInsertEventParams{DeliveryJobID: jobID, Kind: "assigned", CreatedBy: &actor, OccurredAt: time.Now()}); err != nil {
			return err
		}
		if j.RiderID != nil && *j.RiderID != riderID {
			tx.Notify("realtime", map[string]any{"topic": "rider:" + j.RiderID.String(), "event": "job_unassigned", "data": map[string]any{"jobId": jobID}})
		}
		zone := ""
		if z, err := tx.Q.LogisticsListZones(ctx); err == nil {
			for _, x := range z {
				if x.ID == out.ZoneID {
					zone = x.Name
				}
			}
		}
		number := ""
		if v, err := tx.Q.LogisticsJobView(ctx, jobID); err == nil && v.OrderNumber != nil {
			number = *v.OrderNumber
		}
		tx.Notify("realtime", map[string]any{"topic": "rider:" + riderID.String(), "event": "job_assigned", "data": map[string]any{"jobId": jobID}})
		tx.Notify("realtime", map[string]any{"topic": "dispatch", "event": "job", "data": map[string]any{"jobId": jobID, "status": "assigned", "riderId": riderID}})
		if err := s.notify.Send(ctx, tx, notify.Msg{UserID: &riderID, Channel: "push", Category: notify.CatDelivery, Template: "rider_new_job",
			Data: map[string]any{"order": number, "zone": zone}, Key: "job:" + jobID.String() + ":assigned:" + riderID.String() + ":" + time.Now().Format(time.RFC3339)}); err != nil {
			return err
		}
		return tx.Audit("dispatch.assigned", "delivery_job", jobID.String(), map[string]any{"rider": riderID})
	})
	return out, err
}

// Unassign takes an assigned (not yet picked up) job back to the board.
func (s *Service) Unassign(ctx context.Context, jobID, actor uuid.UUID) (store.LogisticsDeliveryJob, error) {
	var out store.LogisticsDeliveryJob
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		j, err := tx.Q.LogisticsGetJobForUpdate(ctx, jobID)
		if err != nil {
			return httpx.NotFound("Job")
		}
		out, err = tx.Q.LogisticsUnassignJob(ctx, jobID)
		if err != nil {
			return httpx.Conflict("not_unassignable", "Only assigned jobs that haven't been picked up can be unassigned.")
		}
		if j.RiderID != nil {
			tx.Notify("realtime", map[string]any{"topic": "rider:" + j.RiderID.String(), "event": "job_unassigned", "data": map[string]any{"jobId": jobID}})
		}
		return tx.Audit("dispatch.unassigned", "delivery_job", jobID.String(), nil)
	})
	return out, err
}

// ReturnToHub closes a failed job: the parcel goes back to the warehouse.
func (s *Service) ReturnToHub(ctx context.Context, jobID, actor uuid.UUID, note string) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		j, err := tx.Q.LogisticsGetJobForUpdate(ctx, jobID)
		if err != nil {
			return httpx.NotFound("Job")
		}
		if j.Status != "failed" {
			return httpx.Conflict("bad_transition", "Only failed deliveries can be returned to the hub.")
		}
		if _, err := tx.Q.LogisticsSetJobStatus(ctx, store.LogisticsSetJobStatusParams{ID: jobID, Status: "returned"}); err != nil {
			return err
		}
		if _, err := tx.Q.LogisticsInsertEvent(ctx, store.LogisticsInsertEventParams{DeliveryJobID: jobID, Kind: "returned", Note: &note, CreatedBy: &actor, OccurredAt: time.Now()}); err != nil {
			return err
		}
		if j.FulfilmentID != nil {
			if _, err := s.sales.Move(ctx, tx, *j.FulfilmentID, "returned", &actor, note); err != nil {
				return err
			}
		}
		return tx.Audit("dispatch.returned", "delivery_job", jobID.String(), map[string]any{"note": note})
	})
}

// ---- Rider actions (online calls and offline sync share this path)

// Action is one rider action on a job.
type Action struct {
	ClientEventID uuid.UUID  `json:"clientEventId" doc:"Generated on the device; makes the action idempotent"`
	JobID         uuid.UUID  `json:"jobId,omitempty" doc:"Required in batch sync; taken from the path for single actions"`
	Kind          string     `json:"kind" enum:"picked_up,en_route,attempt_failed,delivered,note"`
	OccurredAt    time.Time  `json:"occurredAt"`
	Latitude      *float64   `json:"latitude,omitempty" minimum:"-90" maximum:"90"`
	Longitude     *float64   `json:"longitude,omitempty" minimum:"-180" maximum:"180"`
	AccuracyM     *float32   `json:"accuracyM,omitempty"`
	LabelCode     string     `json:"labelCode,omitempty" maxLength:"40" doc:"picked_up: the scanned parcel label"`
	Reason        string     `json:"reason,omitempty" enum:"customer_unreachable,customer_refused,wrong_address,rescheduled,unsafe,other,"`
	Note          string     `json:"note,omitempty" maxLength:"500"`
	OTP           string     `json:"otp,omitempty" pattern:"^[0-9]{6}$" doc:"delivered: the customer's delivery code"`
	PhotoFileID   *uuid.UUID `json:"photoFileId,omitempty" doc:"delivered: photo proof (required when there's no code)"`
	RecipientName string     `json:"recipientName,omitempty" maxLength:"80"`
	Offline       bool       `json:"offline,omitempty" doc:"Recorded without signal (photo proof becomes photo_offline)"`
}

// Result is the outcome of one synced action.
type Result struct {
	ClientEventID uuid.UUID `json:"clientEventId"`
	Status        string    `json:"status" enum:"accepted,duplicate,rejected"`
	Code          string    `json:"code,omitempty"`
	Detail        string    `json:"detail,omitempty"`
}

// wrongOTP is returned for a wrong delivery code. The action is rolled back, but the failed
// attempt must still count, so the caller records it in the outer transaction.
type wrongOTP struct{ jobID uuid.UUID }

func (e *wrongOTP) Error() string { return "wrong delivery code" }

// applyOne runs one action in a savepoint and turns failures into a per-item result.
func (s *Service) applyOne(ctx context.Context, tx *uow.Tx, riderID uuid.UUID, a Action) (Result, error) {
	res := Result{ClientEventID: a.ClientEventID, Status: "accepted"}
	err := tx.Savepoint(ctx, func(sub *uow.Tx) error { return s.apply(ctx, sub, riderID, a, &res) })
	var w *wrongOTP
	if errors.As(err, &w) {
		n, ferr := tx.Q.LogisticsOTPFailed(ctx, w.jobID)
		if ferr != nil {
			return res, ferr
		}
		err = httpx.Invalid("otp_wrong", fmt.Sprintf("That code is wrong. %d tries left; check it with the customer.", max(0, maxOTPAttempts-int(n))))
	}
	var p *httpx.Problem
	switch {
	case err == nil:
		return res, nil
	case errors.As(err, &p):
		res.Status, res.Code, res.Detail = "rejected", p.Code, p.Detail
		return res, p
	default:
		return res, err
	}
}

// Sync applies a batch of offline actions, each once per clientEventId. A failing item is
// rolled back alone (savepoint) and reported; the rest still apply.
func (s *Service) Sync(ctx context.Context, riderID uuid.UUID, actions []Action) ([]Result, error) {
	out := make([]Result, 0, len(actions))
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		for _, a := range actions {
			res, err := s.applyOne(ctx, tx, riderID, a)
			var p *httpx.Problem
			if err != nil && !errors.As(err, &p) {
				return err
			}
			out = append(out, res)
		}
		return nil
	})
	return out, err
}

// Act applies one online action. A rejected action still commits what must persist (a wrong
// code's attempt count) and then returns its problem.
func (s *Service) Act(ctx context.Context, riderID uuid.UUID, a Action) (Result, error) {
	var res Result
	var rejected error
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		var err error
		res, err = s.applyOne(ctx, tx, riderID, a)
		var p *httpx.Problem
		if errors.As(err, &p) {
			rejected = p
			return nil
		}
		return err
	})
	if err != nil {
		return res, err
	}
	return res, rejected
}

func (s *Service) apply(ctx context.Context, tx *uow.Tx, riderID uuid.UUID, a Action, res *Result) error {
	if a.JobID == uuid.Nil || a.ClientEventID == uuid.Nil {
		return httpx.Invalid("bad_event", "Each action needs a jobId and a clientEventId.")
	}
	j, err := tx.Q.LogisticsGetJobForUpdate(ctx, a.JobID)
	if err != nil {
		return httpx.NotFound("Job")
	}
	if j.RiderID == nil || *j.RiderID != riderID {
		return httpx.Conflict("job_reassigned", "This job is no longer yours. Hand the parcel back at the hub.")
	}
	kind := a.Kind
	if kind == "delivered" && a.OTP == "" && a.PhotoFileID == nil {
		return httpx.Invalid("proof_required", "Enter the customer's delivery code or take a photo.")
	}
	if kind == "attempt_failed" && a.Reason == "" {
		return httpx.Invalid("reason_required", "Choose why the delivery failed.")
	}
	occurred := a.OccurredAt
	if occurred.IsZero() || occurred.After(time.Now().Add(5*time.Minute)) {
		occurred = time.Now()
	}
	var reason, note *string
	if a.Reason != "" {
		reason = &a.Reason
	}
	if a.Note != "" {
		note = &a.Note
	}
	cid := a.ClientEventID
	_, err = tx.Q.LogisticsInsertEvent(ctx, store.LogisticsInsertEventParams{DeliveryJobID: j.ID, ClientEventID: &cid, Kind: kind, Reason: reason,
		Latitude: a.Latitude, Longitude: a.Longitude, AccuracyM: a.AccuracyM, Note: note, FileID: a.PhotoFileID, CreatedBy: &riderID, OccurredAt: occurred})
	if errors.Is(err, pgx.ErrNoRows) {
		res.Status = "duplicate" // already applied when first synced
		return nil
	}
	if err != nil {
		return httpx.DB(err, "delivery event")
	}
	switch kind {
	case "note":
		return nil
	case "picked_up":
		return s.pickedUp(ctx, tx, j, a, riderID)
	case "en_route":
		if j.Status != "picked_up" && j.Status != "failed" {
			return httpx.Conflict("bad_transition", "Pick the parcel up first.")
		}
		if j.Status == "failed" && j.FulfilmentID != nil {
			if _, err := s.sales.Move(ctx, tx, *j.FulfilmentID, "in_transit", &riderID, "re-attempt"); err != nil {
				return err
			}
		}
		_, err := tx.Q.LogisticsSetJobStatus(ctx, store.LogisticsSetJobStatusParams{ID: j.ID, Status: "en_route"})
		s.publishJob(tx, j.ID, "en_route", riderID)
		return err
	case "attempt_failed":
		if j.Status != "picked_up" && j.Status != "en_route" {
			return httpx.Conflict("bad_transition", "Only a parcel on its way can fail delivery.")
		}
		if _, err := tx.Q.LogisticsSetJobStatus(ctx, store.LogisticsSetJobStatusParams{ID: j.ID, Status: "failed"}); err != nil {
			return err
		}
		if j.FulfilmentID != nil {
			if _, err := s.sales.Move(ctx, tx, *j.FulfilmentID, "failed", &riderID, "Delivery attempt failed: "+a.Reason); err != nil {
				return err
			}
		}
		s.publishJob(tx, j.ID, "failed", riderID)
		return s.refreshRiderStatus(ctx, tx, riderID)
	case "delivered":
		return s.deliver(ctx, tx, j, a, riderID)
	}
	return httpx.Invalid("bad_kind", "Unknown action.")
}

// pickedUp: the scanned label must belong to this job's parcel; the customer gets their delivery code.
func (s *Service) pickedUp(ctx context.Context, tx *uow.Tx, j store.LogisticsDeliveryJob, a Action, riderID uuid.UUID) error {
	if j.Status != "assigned" {
		return httpx.Conflict("bad_transition", "This job isn't waiting for pickup.")
	}
	if j.FulfilmentID != nil {
		parcels, err := tx.Q.FulfilmentParcelsForFulfilment(ctx, *j.FulfilmentID)
		if err != nil {
			return err
		}
		match := false
		for _, p := range parcels {
			if strings.EqualFold(p.LabelCode, strings.TrimSpace(a.LabelCode)) {
				match = true
				if err := tx.Q.FulfilmentMarkParcelScanned(ctx, p.ID); err != nil {
					return err
				}
			}
		}
		if !match {
			return httpx.Conflict("label_mismatch", "That label doesn't belong to this job. Check the parcel.")
		}
		if _, err := s.sales.Move(ctx, tx, *j.FulfilmentID, "in_transit", &riderID, "Picked up by rider"); err != nil {
			return err
		}
	}
	if _, err := tx.Q.LogisticsSetJobStatus(ctx, store.LogisticsSetJobStatusParams{ID: j.ID, Status: "picked_up"}); err != nil {
		return err
	}
	if err := tx.Q.LogisticsSetRiderStatus(ctx, store.LogisticsSetRiderStatusParams{UserID: riderID, Status: "on_delivery"}); err != nil {
		return err
	}
	s.publishJob(tx, j.ID, "picked_up", riderID)
	return s.issueOTP(ctx, tx, j.ID)
}

func (s *Service) otpHash(jobID uuid.UUID, code string) []byte {
	return crypto.HMAC(s.d.Cfg.OTPHMACKey, "delivery|"+jobID.String()+"|"+code)
}

// issueOTP sends the customer a 6-digit delivery code (hashed at rest, valid for a day).
func (s *Service) issueOTP(ctx context.Context, tx *uow.Tx, jobID uuid.UUID) error {
	v, err := tx.Q.LogisticsJobView(ctx, jobID)
	if err != nil || v.CustomerUserID == nil {
		return err
	}
	code := crypto.RandomDigits(6)
	exp := time.Now().Add(otpTTL)
	if err := tx.Q.LogisticsSetJobOTP(ctx, store.LogisticsSetJobOTPParams{ID: jobID, OtpHash: s.otpHash(jobID, code), OtpExpiresAt: &exp}); err != nil {
		return err
	}
	data := map[string]any{"order": deref(v.OrderNumber), "code": code}
	if err := s.notify.Send(ctx, tx, notify.Msg{UserID: v.CustomerUserID, To: deref(v.ContactPhone), Channel: "sms", Category: notify.CatDelivery, Template: "delivery_code",
		Data: data, Key: "job:" + jobID.String() + ":otp:sms:" + exp.Format(time.RFC3339Nano)}); err != nil {
		return err
	}
	return s.notify.Send(ctx, tx, notify.Msg{UserID: v.CustomerUserID, Channel: "push", Category: notify.CatDelivery, Template: "delivery_code",
		Data: data, Key: "job:" + jobID.String() + ":otp:push:" + exp.Format(time.RFC3339Nano)})
}

// deliver closes a job with a verified code, or with photo proof (offline-safe).
func (s *Service) deliver(ctx context.Context, tx *uow.Tx, j store.LogisticsDeliveryJob, a Action, riderID uuid.UUID) error {
	if j.Status != "picked_up" && j.Status != "en_route" {
		return httpx.Conflict("bad_transition", "Only a parcel on its way can be delivered.")
	}
	p := store.LogisticsSetJobStatusParams{ID: j.ID, Status: "delivered"}
	if a.RecipientName != "" {
		p.RecipientName = &a.RecipientName
	}
	if a.OTP != "" {
		switch {
		case j.OtpHash == nil || j.OtpExpiresAt == nil || time.Now().After(*j.OtpExpiresAt):
			return httpx.Conflict("otp_expired", "The delivery code has expired. Ask dispatch to resend it, or use photo proof.")
		case j.OtpAttempts >= maxOTPAttempts:
			return httpx.Conflict("otp_locked", "Too many wrong codes. Use photo proof.")
		case !hmac.Equal(j.OtpHash, s.otpHash(j.ID, a.OTP)):
			return &wrongOTP{jobID: j.ID}
		}
		if err := tx.Q.LogisticsOTPVerified(ctx, j.ID); err != nil {
			return err
		}
		p.ProofMethod = ptr("otp")
	} else {
		method := "photo"
		if a.Offline {
			method = "photo_offline"
		}
		p.ProofMethod, p.ProofFileID = &method, a.PhotoFileID
	}
	if _, err := tx.Q.LogisticsSetJobStatus(ctx, p); err != nil {
		return httpx.DB(err, "delivery job")
	}
	if j.FulfilmentID != nil {
		if j.Status == "failed" {
			if _, err := s.sales.Move(ctx, tx, *j.FulfilmentID, "in_transit", &riderID, "re-attempt"); err != nil {
				return err
			}
		}
		if _, err := s.sales.Move(ctx, tx, *j.FulfilmentID, "delivered", &riderID, "Delivered ("+*p.ProofMethod+")"); err != nil {
			return err
		}
	}
	s.publishJob(tx, j.ID, "delivered", riderID)
	return s.refreshRiderStatus(ctx, tx, riderID)
}

// refreshRiderStatus: a rider with no parcel in hand is available again.
func (s *Service) refreshRiderStatus(ctx context.Context, tx *uow.Tx, riderID uuid.UUID) error {
	jobs, err := tx.Q.LogisticsRiderJobs(ctx, &riderID)
	if err != nil {
		return err
	}
	for _, j := range jobs {
		if j.Status == "picked_up" || j.Status == "en_route" {
			return nil
		}
	}
	return tx.Q.LogisticsSetRiderStatus(ctx, store.LogisticsSetRiderStatusParams{UserID: riderID, Status: "available"})
}

func (s *Service) publishJob(tx *uow.Tx, jobID uuid.UUID, status string, riderID uuid.UUID) {
	tx.Notify("realtime", map[string]any{"topic": "dispatch", "event": "job", "data": map[string]any{"jobId": jobID, "status": status, "riderId": riderID}})
}

// ---- Location

// Point is one location fix from the rider's phone.
type Point struct {
	Latitude   float64   `json:"latitude" minimum:"-90" maximum:"90"`
	Longitude  float64   `json:"longitude" minimum:"-180" maximum:"180"`
	AccuracyM  *float32  `json:"accuracyM,omitempty"`
	Speed      *float32  `json:"speed,omitempty"`
	Heading    *float32  `json:"heading,omitempty"`
	RecordedAt time.Time `json:"recordedAt"`
}

// RecordLocation stores a batch of fixes (only on shift). Customers see the rider only while
// their own parcel is en route; dispatch sees every rider on shift.
func (s *Service) RecordLocation(ctx context.Context, riderID uuid.UUID, pts []Point) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if _, err := tx.Q.LogisticsOpenShift(ctx, riderID); err != nil {
			return httpx.Conflict("off_shift", "Location is only shared during a shift.")
		}
		var last *Point
		for i := range pts {
			p := pts[i]
			if p.RecordedAt.After(time.Now().Add(time.Minute)) || p.RecordedAt.Before(time.Now().Add(-24*time.Hour)) {
				continue // clock nonsense: ignore the point
			}
			if err := tx.Q.LogisticsInsertPing(ctx, store.LogisticsInsertPingParams{RiderID: riderID, Latitude: p.Latitude, Longitude: p.Longitude,
				AccuracyM: p.AccuracyM, Speed: p.Speed, Heading: p.Heading, RecordedAt: p.RecordedAt}); err != nil {
				return err
			}
			if last == nil || p.RecordedAt.After(last.RecordedAt) {
				last = &pts[i]
			}
		}
		if last == nil {
			return nil
		}
		enRoute, err := tx.Q.LogisticsEnRouteJobs(ctx, &riderID)
		if err != nil {
			return err
		}
		var jobID *uuid.UUID
		if len(enRoute) > 0 {
			jobID = &enRoute[0].ID
		}
		if err := tx.Q.LogisticsUpsertLocation(ctx, store.LogisticsUpsertLocationParams{RiderID: riderID, Latitude: last.Latitude, Longitude: last.Longitude,
			AccuracyM: last.AccuracyM, JobID: jobID, RecordedAt: last.RecordedAt}); err != nil {
			return err
		}
		pos := map[string]any{"riderId": riderID, "latitude": last.Latitude, "longitude": last.Longitude, "at": last.RecordedAt}
		tx.Notify("realtime", map[string]any{"topic": "dispatch", "event": "rider_location", "data": pos})
		for _, j := range enRoute {
			// A sampled trail point per batch, and the live dot for that customer only.
			if _, err := tx.Q.LogisticsInsertEvent(ctx, store.LogisticsInsertEventParams{DeliveryJobID: j.ID, Kind: "location", Latitude: &last.Latitude,
				Longitude: &last.Longitude, AccuracyM: last.AccuracyM, CreatedBy: &riderID, OccurredAt: last.RecordedAt}); err != nil {
				return err
			}
			tx.Notify("realtime", map[string]any{"topic": "order:" + j.OrderID.String(), "event": "rider_location",
				"data": map[string]any{"latitude": last.Latitude, "longitude": last.Longitude, "at": last.RecordedAt}})
		}
		return nil
	})
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func ptr[T any](v T) *T { return &v }
