package logistics

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

var (
	riderAud    = []string{auth.AudLogistics}
	dispatchAud = []string{auth.AudStaff, auth.AudLogistics} // staff web and the dispatcher mode of the app
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

// JobDetail is the dispatcher's view of a job with its event trail.
type JobDetail struct {
	RiderJob
	RiderID *uuid.UUID                     `json:"riderId"`
	ZoneID  uuid.UUID                      `json:"zoneId"`
	Proof   *string                        `json:"proofMethod"`
	Events  []store.LogisticsDeliveryEvent `json:"events"`
}

// rider resolves the signed-in rider (403 for anyone else on the logistics audience).
func (s *Service) rider(ctx context.Context) (store.LogisticsRider, error) {
	p := httpx.MustPrincipal(ctx)
	r, err := s.d.Q.LogisticsGetRider(ctx, p.UserID)
	if err != nil {
		return r, httpx.Forbidden("This account isn't set up as a rider.")
	}
	if r.Status == "suspended" {
		return r, httpx.Forbidden("Your rider account is suspended. Contact dispatch.")
	}
	return r, nil
}

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	rid := func(o httpx.Op) httpx.Op { o.Tag = "Rider"; o.Auth = riderAud; return o }
	disp := func(o httpx.Op) httpx.Op { o.Tag = "Dispatch"; o.Auth = dispatchAud; return o }

	// ---- Public delivery estimate (product pages, cart).
	httpx.Register(g, httpx.Op{ID: "deliveryEstimate", Method: http.MethodGet, Path: "/api/v1/delivery/estimate", Tag: "Catalogue", Summary: "Delivery fee and days for an address"},
		func(ctx context.Context, in *struct {
			State       string `query:"state" pattern:"^[A-Z]{2}$" required:"true"`
			LGA         string `query:"lga" maxLength:"60"`
			WeightGrams int32  `query:"weightGrams" minimum:"0" maximum:"100000"`
			FulfilledBy string `query:"fulfilledBy" enum:"techshop,seller," default:"techshop"`
		}) (*struct {
			Body struct {
				Covered    bool    `json:"covered" doc:"False: no own-rider zone; a carrier or flat interstate fee applies"`
				Zone       *string `json:"zone,omitempty"`
				FeeKobo    int64   `json:"feeKobo"`
				EtaMinDays int16   `json:"etaMinDays"`
				EtaMaxDays int16   `json:"etaMaxDays"`
			}
		}, error) {
			out := &struct {
				Body struct {
					Covered    bool    `json:"covered" doc:"False: no own-rider zone; a carrier or flat interstate fee applies"`
					Zone       *string `json:"zone,omitempty"`
					FeeKobo    int64   `json:"feeKobo"`
					EtaMinDays int16   `json:"etaMinDays"`
					EtaMaxDays int16   `json:"etaMaxDays"`
				}
			}{}
			fb := in.FulfilledBy
			if fb == "" {
				fb = "techshop"
			}
			rate, z, ok := s.estimate(ctx, s.d.Q, fb, in.State, in.LGA, in.WeightGrams)
			if ok {
				out.Body.Covered, out.Body.Zone, out.Body.FeeKobo = true, &z.Name, rate.FeeKobo
				out.Body.EtaMinDays, out.Body.EtaMaxDays = rate.EtaMinDays, rate.EtaMaxDays
				return out, nil
			}
			// Not covered by own riders: the flat fallback used at checkout.
			out.Body.FeeKobo, out.Body.EtaMinDays, out.Body.EtaMaxDays = s.d.Cfg.DeliveryFeeInterstateKobo, 2, 5
			return out, nil
		})

	// ---- Rider: devices, shift, jobs, actions, offline sync, location
	httpx.Register(g, rid(httpx.Op{ID: "riderRegisterDevice", Method: http.MethodPost, Path: "/api/v1/rider/devices", Status: 201,
		Summary: "Register this phone (dispatch must approve it before a shift)"}),
		func(ctx context.Context, in *struct {
			Body struct {
				DeviceID string `json:"deviceId" minLength:"8" maxLength:"128"`
				Platform string `json:"platform" enum:"ios,android"`
			}
		}) (*struct{ Body store.LogisticsRiderDevice }, error) {
			rd, err := s.rider(ctx)
			if err != nil {
				return nil, err
			}
			d, err := s.d.Q.LogisticsRegisterDevice(ctx, store.LogisticsRegisterDeviceParams{RiderID: rd.UserID, DeviceID: in.Body.DeviceID, Platform: in.Body.Platform})
			return &struct{ Body store.LogisticsRiderDevice }{d}, err
		})
	httpx.Register(g, rid(httpx.Op{ID: "riderStartShift", Method: http.MethodPost, Path: "/api/v1/rider/shift/start", Status: 201,
		Summary: "Start a shift (approved device and location consent required)"}),
		func(ctx context.Context, in *struct {
			Body struct {
				LocationConsent bool     `json:"locationConsent" doc:"The rider accepted the location disclosure"`
				Latitude        *float64 `json:"latitude,omitempty" minimum:"-90" maximum:"90"`
				Longitude       *float64 `json:"longitude,omitempty" minimum:"-180" maximum:"180"`
			}
		}) (*struct{ Body store.LogisticsRiderShift }, error) {
			rd, err := s.rider(ctx)
			if err != nil {
				return nil, err
			}
			if !in.Body.LocationConsent {
				return nil, httpx.Invalid("consent_required", "Location sharing during shifts is required to deliver.")
			}
			p := httpx.MustPrincipal(ctx)
			var out store.LogisticsRiderShift
			err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				sess, err := tx.Q.IdentityGetSession(ctx, p.SessionID)
				if err != nil || sess.DeviceID == nil {
					return httpx.Forbidden("Sign in from the rider app on a registered phone.")
				}
				ok, err := tx.Q.LogisticsDeviceApproved(ctx, store.LogisticsDeviceApprovedParams{RiderID: rd.UserID, DeviceID: *sess.DeviceID})
				if err != nil {
					return err
				}
				if !ok {
					return httpx.E(403, "device_not_approved", "This phone isn't approved yet. Ask dispatch to approve it.")
				}
				if open, err := tx.Q.LogisticsOpenShift(ctx, rd.UserID); err == nil {
					out = open
					return nil // already on shift: idempotent
				}
				out, err = tx.Q.LogisticsStartShift(ctx, store.LogisticsStartShiftParams{RiderID: rd.UserID, DeviceSessionID: &p.SessionID,
					StartLatitude: in.Body.Latitude, StartLongitude: in.Body.Longitude})
				if err != nil {
					return httpx.DB(err, "shift")
				}
				if err := tx.Q.LogisticsSetRiderStatus(ctx, store.LogisticsSetRiderStatusParams{UserID: rd.UserID, Status: "available"}); err != nil {
					return err
				}
				tx.Notify("realtime", map[string]any{"topic": "dispatch", "event": "rider", "data": map[string]any{"riderId": rd.UserID, "status": "available"}})
				return nil
			})
			return &struct{ Body store.LogisticsRiderShift }{out}, err
		})
	httpx.Register(g, rid(httpx.Op{ID: "riderEndShift", Method: http.MethodPost, Path: "/api/v1/rider/shift/end", Summary: "End the shift (location sharing stops)"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body store.LogisticsRiderShift }, error) {
			rd, err := s.rider(ctx)
			if err != nil {
				return nil, err
			}
			var out store.LogisticsRiderShift
			err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var err error
				out, err = tx.Q.LogisticsEndShift(ctx, rd.UserID)
				if err != nil {
					return httpx.Conflict("not_on_shift", "You aren't on shift.")
				}
				tx.Notify("realtime", map[string]any{"topic": "dispatch", "event": "rider", "data": map[string]any{"riderId": rd.UserID, "status": "off_shift"}})
				return tx.Q.LogisticsSetRiderStatus(ctx, store.LogisticsSetRiderStatusParams{UserID: rd.UserID, Status: "off_shift"})
			})
			return &struct{ Body store.LogisticsRiderShift }{out}, err
		})
	httpx.Register(g, rid(httpx.Op{ID: "riderJobs", Method: http.MethodGet, Path: "/api/v1/rider/jobs", Summary: "My jobs: to pick up, to deliver, done today"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []RiderJob }, error) {
			rd, err := s.rider(ctx)
			if err != nil {
				return nil, err
			}
			rows, err := s.d.Q.LogisticsRiderJobs(ctx, &rd.UserID)
			out := make([]RiderJob, 0, len(rows))
			for _, j := range rows {
				out = append(out, riderJob(j.ID, j.Kind, j.Status, j.OrderNumber, j.RecipientName, j.ShipTo, j.DropoffLatitude, j.DropoffLongitude,
					j.PickupWarehouseName, j.PickupWarehouseAddress, j.WindowStart, j.WindowEnd, j.Attempts, j.RouteSequence))
			}
			return &struct{ Body []RiderJob }{out}, err
		})
	httpx.Register(g, rid(httpx.Op{ID: "riderAct", Method: http.MethodPost, Path: "/api/v1/rider/jobs/{id}/actions", Summary: "Pick up, start, fail or deliver a job"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body Action
		}) (*struct{ Body Result }, error) {
			rd, err := s.rider(ctx)
			if err != nil {
				return nil, err
			}
			in.Body.JobID = in.ID
			res, err := s.Act(ctx, rd.UserID, in.Body)
			return &struct{ Body Result }{res}, err
		})
	httpx.Register(g, rid(httpx.Op{ID: "riderSyncEvents", Method: http.MethodPost, Path: "/api/v1/rider/events/batch", MaxBody: 512 << 10,
		Summary: "Offline sync: apply queued actions once each (per clientEventId)"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Events []Action `json:"events" minItems:"1" maxItems:"100"`
			}
		}) (*struct {
			Body struct {
				Results []Result `json:"results"`
			}
		}, error) {
			rd, err := s.rider(ctx)
			if err != nil {
				return nil, err
			}
			res, err := s.Sync(ctx, rd.UserID, in.Body.Events)
			out := &struct {
				Body struct {
					Results []Result `json:"results"`
				}
			}{}
			out.Body.Results = res
			return out, err
		})
	httpx.Register(g, rid(httpx.Op{ID: "riderLocation", Method: http.MethodPost, Path: "/api/v1/rider/location", Status: 204, Summary: "Upload location fixes (on shift only)"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Points []Point `json:"points" minItems:"1" maxItems:"120"`
			}
		}) (*struct{}, error) {
			rd, err := s.rider(ctx)
			if err != nil {
				return nil, err
			}
			return &struct{}{}, s.RecordLocation(ctx, rd.UserID, in.Body.Points)
		})

	// ---- Dispatch
	httpx.Register(g, disp(httpx.Op{ID: "dispatchBoard", Method: http.MethodGet, Path: "/api/v1/dispatch/board", Perm: "dispatch.view", Summary: "Key figures and jobs by status"}),
		func(ctx context.Context, in *struct {
			Status string `query:"status" enum:"unassigned,assigned,picked_up,en_route,delivered,failed,returned,cancelled,"`
			ZoneID string `query:"zoneId"`
		}) (*struct {
			Body struct {
				Counts store.LogisticsBoardCountsRow `json:"counts"`
				Jobs   []store.LogisticsBoardRow     `json:"jobs"`
			}
		}, error) {
			out := &struct {
				Body struct {
					Counts store.LogisticsBoardCountsRow `json:"counts"`
					Jobs   []store.LogisticsBoardRow     `json:"jobs"`
				}
			}{}
			p := store.LogisticsBoardParams{}
			if in.Status != "" {
				p.Status = &in.Status
			}
			if id, err := uuid.Parse(in.ZoneID); err == nil {
				p.ZoneID = &id
			}
			var err error
			if out.Body.Counts, err = s.d.Q.LogisticsBoardCounts(ctx); err != nil {
				return nil, err
			}
			out.Body.Jobs, err = s.d.Q.LogisticsBoard(ctx, p)
			for i := range out.Body.Jobs {
				out.Body.Jobs[i].OtpHash = nil // never leaves the server
			}
			if out.Body.Jobs == nil {
				out.Body.Jobs = []store.LogisticsBoardRow{}
			}
			return out, err
		})
	httpx.Register(g, disp(httpx.Op{ID: "dispatchJob", Method: http.MethodGet, Path: "/api/v1/dispatch/jobs/{id}", Perm: "dispatch.view", Summary: "Job with its event trail"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body JobDetail }, error) {
			v, err := s.d.Q.LogisticsJobView(ctx, in.ID)
			if err != nil {
				return nil, httpx.NotFound("Job")
			}
			d := JobDetail{RiderJob: riderJob(v.ID, v.Kind, v.Status, v.OrderNumber, v.RecipientName, v.ShipTo, v.DropoffLatitude, v.DropoffLongitude,
				v.PickupWarehouseName, v.PickupWarehouseAddress, v.WindowStart, v.WindowEnd, v.Attempts, v.RouteSequence), RiderID: v.RiderID, ZoneID: v.ZoneID, Proof: v.ProofMethod}
			d.Events, err = s.d.Q.LogisticsJobEvents(ctx, in.ID)
			return &struct{ Body JobDetail }{d}, err
		})
	httpx.Register(g, disp(httpx.Op{ID: "dispatchAssign", Method: http.MethodPost, Path: "/api/v1/dispatch/jobs/{id}/assign", Perm: "dispatch.assign", Summary: "Assign or reassign a rider"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				RiderID       uuid.UUID `json:"riderId"`
				RouteSequence *int16    `json:"routeSequence,omitempty" minimum:"1"`
			}
		}) (*struct{ Body store.LogisticsDeliveryJob }, error) {
			j, err := s.Assign(ctx, in.ID, in.Body.RiderID, httpx.MustPrincipal(ctx).UserID, in.Body.RouteSequence)
			j.OtpHash = nil
			return &struct{ Body store.LogisticsDeliveryJob }{j}, err
		})
	httpx.Register(g, disp(httpx.Op{ID: "dispatchUnassign", Method: http.MethodPost, Path: "/api/v1/dispatch/jobs/{id}/unassign", Perm: "dispatch.assign", Summary: "Back to the board"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body store.LogisticsDeliveryJob }, error) {
			j, err := s.Unassign(ctx, in.ID, httpx.MustPrincipal(ctx).UserID)
			j.OtpHash = nil
			return &struct{ Body store.LogisticsDeliveryJob }{j}, err
		})
	httpx.Register(g, disp(httpx.Op{ID: "dispatchReturn", Method: http.MethodPost, Path: "/api/v1/dispatch/jobs/{id}/return", Perm: "dispatch.assign", Status: 204,
		Summary: "Return a failed parcel to the hub"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Note string `json:"note" minLength:"3" maxLength:"300"`
			}
		}) (*struct{}, error) {
			return &struct{}{}, s.ReturnToHub(ctx, in.ID, httpx.MustPrincipal(ctx).UserID, in.Body.Note)
		})
	httpx.Register(g, disp(httpx.Op{ID: "dispatchRiders", Method: http.MethodGet, Path: "/api/v1/dispatch/riders", Perm: "dispatch.view", Summary: "Riders with shift, load and last position"}),
		func(ctx context.Context, _ *struct{}) (*struct {
			Body []store.LogisticsListRidersRow
		}, error) {
			rows, err := s.d.Q.LogisticsListRiders(ctx)
			if rows == nil {
				rows = []store.LogisticsListRidersRow{}
			}
			return &struct {
				Body []store.LogisticsListRidersRow
			}{rows}, err
		})
	httpx.Register(g, disp(httpx.Op{ID: "dispatchCreateRider", Method: http.MethodPost, Path: "/api/v1/dispatch/riders", Perm: "logistics.manage", Status: 201,
		Summary: "Make a staff member a rider"}),
		func(ctx context.Context, in *struct {
			Body struct {
				UserID          uuid.UUID   `json:"userId"`
				VehicleType     string      `json:"vehicleType" enum:"bike,car,van"`
				PlateNumber     string      `json:"plateNumber,omitempty" maxLength:"20"`
				HomeWarehouseID uuid.UUID   `json:"homeWarehouseId"`
				ZoneIDs         []uuid.UUID `json:"zoneIds,omitempty"`
			}
		}) (*struct{ Body store.LogisticsRider }, error) {
			var out store.LogisticsRider
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				if _, err := tx.Q.IdentityGetStaffMember(ctx, in.Body.UserID); err != nil {
					return httpx.Invalid("not_staff", "Riders must first be added as staff by HR.")
				}
				var plate *string
				if in.Body.PlateNumber != "" {
					plate = &in.Body.PlateNumber
				}
				var err error
				out, err = tx.Q.LogisticsCreateRider(ctx, store.LogisticsCreateRiderParams{UserID: in.Body.UserID, VehicleType: in.Body.VehicleType, PlateNumber: plate,
					HomeWarehouseID: in.Body.HomeWarehouseID})
				if err != nil {
					return httpx.DB(err, "rider")
				}
				for _, z := range in.Body.ZoneIDs {
					if err := tx.Q.LogisticsAddRiderZone(ctx, store.LogisticsAddRiderZoneParams{RiderID: out.UserID, ZoneID: z}); err != nil {
						return httpx.DB(err, "rider zone")
					}
				}
				return tx.Audit("rider.created", "rider", out.UserID.String(), in.Body)
			})
			return &struct{ Body store.LogisticsRider }{out}, err
		})
	httpx.Register(g, disp(httpx.Op{ID: "dispatchSetRiderZones", Method: http.MethodPut, Path: "/api/v1/dispatch/riders/{id}/zones", Perm: "logistics.manage", Status: 204,
		Summary: "Set the zones a rider covers"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				ZoneIDs []uuid.UUID `json:"zoneIds"`
			}
		}) (*struct{}, error) {
			return &struct{}{}, s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				if err := tx.Q.LogisticsSetRiderZones(ctx, in.ID); err != nil {
					return err
				}
				for _, z := range in.Body.ZoneIDs {
					if err := tx.Q.LogisticsAddRiderZone(ctx, store.LogisticsAddRiderZoneParams{RiderID: in.ID, ZoneID: z}); err != nil {
						return httpx.DB(err, "rider zone")
					}
				}
				return tx.Audit("rider.zones_set", "rider", in.ID.String(), in.Body)
			})
		})
	httpx.Register(g, disp(httpx.Op{ID: "dispatchSetRiderStatus", Method: http.MethodPost, Path: "/api/v1/dispatch/riders/{id}/{action}", Perm: "logistics.manage", Status: 204,
		Summary: "Suspend or reinstate a rider"}),
		func(ctx context.Context, in *struct {
			ID     uuid.UUID `path:"id"`
			Action string    `path:"action" enum:"suspend,reinstate"`
		}) (*struct{}, error) {
			status := map[string]string{"suspend": "suspended", "reinstate": "off_shift"}[in.Action]
			return &struct{}{}, s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				if _, err := tx.Q.LogisticsGetRider(ctx, in.ID); err != nil {
					return httpx.NotFound("Rider")
				}
				if in.Action == "suspend" {
					_, _ = tx.Q.LogisticsEndShift(ctx, in.ID)
				}
				if err := tx.Q.LogisticsSetRiderStatus(ctx, store.LogisticsSetRiderStatusParams{UserID: in.ID, Status: status}); err != nil {
					return err
				}
				return tx.Audit("rider."+in.Action, "rider", in.ID.String(), nil)
			})
		})
	httpx.Register(g, disp(httpx.Op{ID: "dispatchDevices", Method: http.MethodGet, Path: "/api/v1/dispatch/devices", Perm: "riders.manage_devices", Summary: "Rider phones (pending first)"}),
		func(ctx context.Context, in *struct {
			Pending bool `query:"pending"`
		}) (*struct {
			Body []store.LogisticsListDevicesRow
		}, error) {
			p := in.Pending
			rows, err := s.d.Q.LogisticsListDevices(ctx, &p)
			if rows == nil {
				rows = []store.LogisticsListDevicesRow{}
			}
			return &struct {
				Body []store.LogisticsListDevicesRow
			}{rows}, err
		})
	httpx.Register(g, disp(httpx.Op{ID: "dispatchDeviceAction", Method: http.MethodPost, Path: "/api/v1/dispatch/devices/{id}/{action}", Perm: "riders.manage_devices",
		Summary: "Approve or revoke a rider's phone (riders can't approve their own)"}),
		func(ctx context.Context, in *struct {
			ID     uuid.UUID `path:"id"`
			Action string    `path:"action" enum:"approve,revoke"`
		}) (*struct{ Body store.LogisticsRiderDevice }, error) {
			actor := httpx.MustPrincipal(ctx).UserID
			var out store.LogisticsRiderDevice
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var err error
				if in.Action == "approve" {
					out, err = tx.Q.LogisticsApproveDevice(ctx, store.LogisticsApproveDeviceParams{ID: in.ID, ApprovedBy: &actor})
				} else {
					out, err = tx.Q.LogisticsRevokeDevice(ctx, in.ID)
				}
				if err != nil {
					if p := httpx.DB(err, "device"); p != nil && strings.Contains(err.Error(), "check") {
						return httpx.E(403, "separation_of_duties", "Riders can't approve their own phone.")
					}
					return httpx.NotFound("Device")
				}
				return tx.Audit("rider_device."+in.Action, "rider_device", in.ID.String(), nil)
			})
			return &struct{ Body store.LogisticsRiderDevice }{out}, err
		})
	httpx.Register(g, disp(httpx.Op{ID: "listZones", Method: http.MethodGet, Path: "/api/v1/dispatch/zones", Perm: "dispatch.view", Summary: "Delivery zones"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.LogisticsDeliveryZone }, error) {
			rows, err := s.d.Q.LogisticsListZones(ctx)
			if rows == nil {
				rows = []store.LogisticsDeliveryZone{}
			}
			return &struct{ Body []store.LogisticsDeliveryZone }{rows}, err
		})
	type zoneIn struct {
		Name      string   `json:"name" minLength:"2" maxLength:"80"`
		StateCode string   `json:"stateCode" pattern:"^[A-Z]{2}$"`
		LGAs      []string `json:"lgas" doc:"Empty = the whole state"`
		IsActive  bool     `json:"isActive"`
	}
	httpx.Register(g, disp(httpx.Op{ID: "createZone", Method: http.MethodPost, Path: "/api/v1/dispatch/zones", Perm: "logistics.manage", Status: 201, Summary: "Add a delivery zone"}),
		func(ctx context.Context, in *struct{ Body zoneIn }) (*struct{ Body store.LogisticsDeliveryZone }, error) {
			var out store.LogisticsDeliveryZone
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var err error
				lgas := in.Body.LGAs
				if lgas == nil {
					lgas = []string{}
				}
				out, err = tx.Q.LogisticsCreateZone(ctx, store.LogisticsCreateZoneParams{Name: in.Body.Name, StateCode: in.Body.StateCode, Lgas: lgas, IsActive: in.Body.IsActive})
				if err != nil {
					return httpx.DB(err, "zone")
				}
				return tx.Audit("zone.created", "zone", out.ID.String(), in.Body)
			})
			return &struct{ Body store.LogisticsDeliveryZone }{out}, err
		})
	httpx.Register(g, disp(httpx.Op{ID: "updateZone", Method: http.MethodPut, Path: "/api/v1/dispatch/zones/{id}", Perm: "logistics.manage", Summary: "Edit a delivery zone"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body zoneIn
		}) (*struct{ Body store.LogisticsDeliveryZone }, error) {
			var out store.LogisticsDeliveryZone
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				lgas := in.Body.LGAs
				if lgas == nil {
					lgas = []string{}
				}
				var err error
				out, err = tx.Q.LogisticsUpdateZone(ctx, store.LogisticsUpdateZoneParams{ID: in.ID, Name: in.Body.Name, Lgas: lgas, IsActive: in.Body.IsActive})
				if err != nil {
					return httpx.NotFound("Zone")
				}
				return tx.Audit("zone.updated", "zone", in.ID.String(), in.Body)
			})
			return &struct{ Body store.LogisticsDeliveryZone }{out}, err
		})
	httpx.Register(g, disp(httpx.Op{ID: "zoneRates", Method: http.MethodGet, Path: "/api/v1/dispatch/zones/{id}/rates", Perm: "dispatch.view", Summary: "A zone's rate bands"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body []store.LogisticsDeliveryRate }, error) {
			rows, err := s.d.Q.LogisticsZoneRates(ctx, in.ID)
			if rows == nil {
				rows = []store.LogisticsDeliveryRate{}
			}
			return &struct{ Body []store.LogisticsDeliveryRate }{rows}, err
		})
	httpx.Register(g, disp(httpx.Op{ID: "createZoneRate", Method: http.MethodPost, Path: "/api/v1/dispatch/zones/{id}/rates", Perm: "logistics.manage", Status: 201,
		Summary: "Add a rate band (new versions take effect from validFrom)"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				FulfilledBy    string    `json:"fulfilledBy" enum:"techshop,seller"`
				MaxWeightGrams int32     `json:"maxWeightGrams" minimum:"1"`
				FeeKobo        int64     `json:"feeKobo" minimum:"0"`
				EtaMinDays     int16     `json:"etaMinDays" minimum:"0" maximum:"30"`
				EtaMaxDays     int16     `json:"etaMaxDays" minimum:"0" maximum:"30"`
				ValidFrom      time.Time `json:"validFrom,omitempty" format:"date"`
			}
		}) (*struct{ Body store.LogisticsDeliveryRate }, error) {
			var out store.LogisticsDeliveryRate
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				from := in.Body.ValidFrom
				if from.IsZero() {
					from = time.Now()
				}
				var err error
				out, err = tx.Q.LogisticsCreateRate(ctx, store.LogisticsCreateRateParams{ZoneID: in.ID, FulfilledBy: in.Body.FulfilledBy, MaxWeightGrams: in.Body.MaxWeightGrams,
					FeeKobo: in.Body.FeeKobo, EtaMinDays: in.Body.EtaMinDays, EtaMaxDays: in.Body.EtaMaxDays, ValidFrom: from})
				if err != nil {
					return httpx.DB(err, "rate")
				}
				return tx.Audit("zone.rate_added", "zone", in.ID.String(), in.Body)
			})
			return &struct{ Body store.LogisticsDeliveryRate }{out}, err
		})

	// Live dispatch feed and the rider's own job feed (SSE).
	r.Gin.GET("/api/v1/dispatch/stream", func(c *gin.Context) {
		p, ok := r.Public.GinPrincipal(c.GetHeader("Authorization"), dispatchAud...)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httpx.Unauthenticated("Sign in to continue."))
			return
		}
		if allowed, err := s.d.RBAC.HasPermission(c.Request.Context(), p.UserID, "dispatch.view"); err != nil || !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, httpx.Forbidden("Missing permission dispatch.view."))
			return
		}
		counts, _ := s.d.Q.LogisticsBoardCounts(c.Request.Context())
		s.d.Hub.ServeSSE(c, "dispatch", counts)
	})
	r.Gin.GET("/api/v1/rider/stream", func(c *gin.Context) {
		p, ok := r.Public.GinPrincipal(c.GetHeader("Authorization"), riderAud...)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httpx.Unauthenticated("Sign in to continue."))
			return
		}
		if _, err := s.d.Q.LogisticsGetRider(c.Request.Context(), p.UserID); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, httpx.Forbidden("This account isn't set up as a rider."))
			return
		}
		s.d.Hub.ServeSSE(c, "rider:"+p.UserID.String(), map[string]any{"riderId": p.UserID})
	})
}
