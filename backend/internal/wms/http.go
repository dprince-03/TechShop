package wms

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

// PickListView is a pick list with its items (ordered by bin for a short walk).
type PickListView struct {
	store.FulfilmentPickList
	Items []store.FulfilmentPickItemsRow `json:"items"`
}

// ManifestView is a manifest with its parcels.
type ManifestView struct {
	store.FulfilmentManifest
	Parcels []store.FulfilmentManifestParcelsRow `json:"parcels"`
}

func (s *Service) pickListView(ctx context.Context, id uuid.UUID) (PickListView, error) {
	pl, err := s.d.Q.FulfilmentGetPickList(ctx, id)
	if err != nil {
		return PickListView{}, httpx.NotFound("Pick list")
	}
	items, err := s.d.Q.FulfilmentPickItems(ctx, id)
	return PickListView{pl, items}, err
}

func (s *Service) manifestView(ctx context.Context, id uuid.UUID) (ManifestView, error) {
	m, err := s.d.Q.FulfilmentGetManifest(ctx, id)
	if err != nil {
		return ManifestView{}, httpx.NotFound("Manifest")
	}
	ps, err := s.d.Q.FulfilmentManifestParcels(ctx, id)
	return ManifestView{m, ps}, err
}

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	op := func(o httpx.Op) httpx.Op { o.Tag = "Staff · Warehouse"; o.Auth = []string{auth.AudStaff}; return o }
	actor := func(ctx context.Context) uuid.UUID { return httpx.MustPrincipal(ctx).UserID }

	httpx.Register(g, op(httpx.Op{ID: "wmsReleaseWave", Method: http.MethodPost, Path: "/api/v1/staff/wms/waves", Perm: "fulfilment.manage", Status: 201,
		Summary: "Release a pick wave for paid orders in a warehouse"}),
		func(ctx context.Context, in *struct {
			Body struct {
				WarehouseID uuid.UUID `json:"warehouseId"`
				Zone        string    `json:"zone,omitempty" maxLength:"20"`
			}
		}) (*struct{ Body PickListView }, error) {
			pl, _, err := s.ReleaseWave(ctx, in.Body.WarehouseID, actor(ctx), in.Body.Zone)
			if err != nil {
				return nil, err
			}
			v, err := s.pickListView(ctx, pl.ID)
			return &struct{ Body PickListView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "wmsListPickLists", Method: http.MethodGet, Path: "/api/v1/staff/wms/pick-lists", Perm: "fulfilment.pick", Summary: "Pick lists in a warehouse"}),
		func(ctx context.Context, in *struct {
			WarehouseID uuid.UUID `query:"warehouseId" required:"true"`
			Status      string    `query:"status" enum:"open,picking,done,cancelled,"`
		}) (*struct {
			Body []store.FulfilmentListPickListsRow
		}, error) {
			p := store.FulfilmentListPickListsParams{WarehouseID: in.WarehouseID}
			if in.Status != "" {
				p.Status = &in.Status
			}
			rows, err := s.d.Q.FulfilmentListPickLists(ctx, p)
			if rows == nil {
				rows = []store.FulfilmentListPickListsRow{}
			}
			return &struct {
				Body []store.FulfilmentListPickListsRow
			}{rows}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "wmsGetPickList", Method: http.MethodGet, Path: "/api/v1/staff/wms/pick-lists/{id}", Perm: "fulfilment.pick", Summary: "Pick list items by bin"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body PickListView }, error) {
			v, err := s.pickListView(ctx, in.ID)
			return &struct{ Body PickListView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "wmsClaimPickList", Method: http.MethodPost, Path: "/api/v1/staff/wms/pick-lists/{id}/claim", Perm: "fulfilment.pick", Summary: "Start picking a list"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body PickListView }, error) {
			if _, err := s.Claim(ctx, in.ID, actor(ctx)); err != nil {
				return nil, err
			}
			v, err := s.pickListView(ctx, in.ID)
			return &struct{ Body PickListView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "wmsScan", Method: http.MethodPost, Path: "/api/v1/staff/wms/pick-lists/{id}/scan", Perm: "fulfilment.pick",
		Summary: "Scan picked units (IMEI/serial for serialised devices)"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				OrderLineID uuid.UUID `json:"orderLineId"`
				Quantity    int32     `json:"quantity" minimum:"1" maximum:"500"`
				Serial      string    `json:"serial,omitempty" maxLength:"40" doc:"IMEI or serial number"`
			}
		}) (*struct {
			Body struct {
				Item         store.FulfilmentPickListItem `json:"item"`
				ListComplete bool                         `json:"listComplete"`
			}
		}, error) {
			item, done, err := s.Scan(ctx, in.ID, in.Body.OrderLineID, actor(ctx), in.Body.Quantity, in.Body.Serial)
			out := &struct {
				Body struct {
					Item         store.FulfilmentPickListItem `json:"item"`
					ListComplete bool                         `json:"listComplete"`
				}
			}{}
			out.Body.Item, out.Body.ListComplete = item, done
			return out, err
		})
	httpx.Register(g, op(httpx.Op{ID: "wmsPack", Method: http.MethodPost, Path: "/api/v1/staff/wms/pack", Perm: "fulfilment.pick", Status: 201,
		Summary: "Pack a picked order: label, weight check, rider job or carrier booking"}),
		func(ctx context.Context, in *struct {
			Body struct {
				FulfilmentID uuid.UUID `json:"fulfilmentId"`
				WeightGrams  int32     `json:"weightGrams" minimum:"1" maximum:"100000"`
				BoxCode      string    `json:"boxCode,omitempty" maxLength:"20"`
			}
		}) (*struct{ Body Packed }, error) {
			p, err := s.Pack(ctx, in.Body.FulfilmentID, actor(ctx), in.Body.WeightGrams, in.Body.BoxCode)
			return &struct{ Body Packed }{p}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "wmsCreateManifest", Method: http.MethodPost, Path: "/api/v1/staff/wms/manifests", Perm: "fulfilment.manage", Status: 201,
		Summary: "List the parcels leaving with a rider or carrier"}),
		func(ctx context.Context, in *struct {
			Body struct {
				WarehouseID uuid.UUID  `json:"warehouseId"`
				RiderID     *uuid.UUID `json:"riderId,omitempty"`
				CarrierID   *uuid.UUID `json:"carrierId,omitempty"`
				Labels      []string   `json:"labels" minItems:"1" maxItems:"200"`
			}
		}) (*struct{ Body ManifestView }, error) {
			m, err := s.CreateManifest(ctx, in.Body.WarehouseID, in.Body.RiderID, in.Body.CarrierID, in.Body.Labels, actor(ctx))
			if err != nil {
				return nil, err
			}
			v, err := s.manifestView(ctx, m.ID)
			return &struct{ Body ManifestView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "wmsListManifests", Method: http.MethodGet, Path: "/api/v1/staff/wms/manifests", Perm: "fulfilment.pick", Summary: "Manifests in a warehouse"}),
		func(ctx context.Context, in *struct {
			WarehouseID uuid.UUID `query:"warehouseId" required:"true"`
		}) (*struct{ Body []store.FulfilmentManifest }, error) {
			rows, err := s.d.Q.FulfilmentListManifests(ctx, in.WarehouseID)
			if rows == nil {
				rows = []store.FulfilmentManifest{}
			}
			return &struct{ Body []store.FulfilmentManifest }{rows}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "wmsGetManifest", Method: http.MethodGet, Path: "/api/v1/staff/wms/manifests/{id}", Perm: "fulfilment.pick", Summary: "Manifest with parcels"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body ManifestView }, error) {
			v, err := s.manifestView(ctx, in.ID)
			return &struct{ Body ManifestView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "wmsSignManifest", Method: http.MethodPost, Path: "/api/v1/staff/wms/manifests/{id}/sign", Perm: "fulfilment.manage",
		Summary: "Hand-over: the rider or driver signs; parcels are handed over"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				SignedByName string `json:"signedByName" minLength:"2" maxLength:"80"`
			}
		}) (*struct{ Body ManifestView }, error) {
			if _, err := s.SignManifest(ctx, in.ID, actor(ctx), in.Body.SignedByName); err != nil {
				return nil, err
			}
			v, err := s.manifestView(ctx, in.ID)
			return &struct{ Body ManifestView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "wmsCarriers", Method: http.MethodGet, Path: "/api/v1/staff/wms/carriers", Perm: "fulfilment.pick", Summary: "Carriers"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.FulfilmentCarrier }, error) {
			rows, err := s.d.Q.FulfilmentListCarriers(ctx)
			if rows == nil {
				rows = []store.FulfilmentCarrier{}
			}
			return &struct{ Body []store.FulfilmentCarrier }{rows}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "wmsUpsertCarrier", Method: http.MethodPost, Path: "/api/v1/staff/wms/carriers", Perm: "fulfilment.manage", Status: 201,
		Summary: "Add or update a carrier (bookings need an adapter; only the fake carrier exists)"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Code     string `json:"code" pattern:"^[a-z0-9_]{2,20}$"`
				Name     string `json:"name" minLength:"2" maxLength:"80"`
				IsActive bool   `json:"isActive"`
			}
		}) (*struct{ Body store.FulfilmentCarrier }, error) {
			var out store.FulfilmentCarrier
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var err error
				out, err = tx.Q.FulfilmentUpsertCarrier(ctx, store.FulfilmentUpsertCarrierParams{Code: in.Body.Code, Name: in.Body.Name, IsActive: in.Body.IsActive})
				if err != nil {
					return httpx.DB(err, "carrier")
				}
				return tx.Audit("wms.carrier_saved", "carrier", out.ID.String(), in.Body)
			})
			return &struct{ Body store.FulfilmentCarrier }{out}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "wmsFulfilmentShipping", Method: http.MethodGet, Path: "/api/v1/staff/wms/fulfilments/{id}/shipping", Perm: "fulfilment.pick",
		Summary: "Parcels and carrier tracking for one fulfilment"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct {
			Body struct {
				Parcels   []store.FulfilmentParcel        `json:"parcels"`
				Shipments []store.FulfilmentShipment      `json:"shipments"`
				Events    []store.FulfilmentShipmentEvent `json:"events"`
			}
		}, error) {
			out := &struct {
				Body struct {
					Parcels   []store.FulfilmentParcel        `json:"parcels"`
					Shipments []store.FulfilmentShipment      `json:"shipments"`
					Events    []store.FulfilmentShipmentEvent `json:"events"`
				}
			}{}
			var err error
			if out.Body.Parcels, err = s.d.Q.FulfilmentParcelsForFulfilment(ctx, in.ID); err != nil {
				return nil, err
			}
			if out.Body.Shipments, err = s.d.Q.FulfilmentShipmentsForFulfilment(ctx, in.ID); err != nil {
				return nil, err
			}
			out.Body.Events = []store.FulfilmentShipmentEvent{}
			for _, sh := range out.Body.Shipments {
				ev, err := s.d.Q.FulfilmentShipmentEvents(ctx, sh.ID)
				if err != nil {
					return nil, err
				}
				out.Body.Events = append(out.Body.Events, ev...)
			}
			return out, nil
		})

	// Carrier tracking webhooks: signature over the raw body, each event stored once.
	httpx.Register(g, httpx.Op{ID: "carrierWebhook", Method: http.MethodPost, Path: "/api/v1/webhooks/carriers/{code}", Tag: "Webhooks", MaxBody: 64 << 10,
		Summary: "Carrier tracking webhook (signed)"},
		func(ctx context.Context, in *struct {
			Code      string `path:"code" pattern:"^[a-z0-9_]{2,20}$"`
			Signature string `header:"x-carrier-signature"`
			RawBody   []byte
		}) (*struct {
			Body struct {
				Status string `json:"status"`
			}
		}, error) {
			if in.Code != s.carrier.Code() || !s.carrier.VerifySignature(in.RawBody, in.Signature) {
				s.d.Logger.Warn("carrier webhook signature rejected", "carrier", in.Code)
				return nil, httpx.E(401, "bad_signature", "Invalid webhook signature.")
			}
			var ev CarrierEvent
			if err := json.Unmarshal(in.RawBody, &ev); err != nil || ev.EventID == "" || ev.TrackingNumber == "" {
				return nil, httpx.Invalid("bad_payload", "Unrecognised tracking event.")
			}
			dup, err := s.ApplyCarrierEvent(ctx, in.Code, ev)
			out := &struct {
				Body struct {
					Status string `json:"status"`
				}
			}{}
			out.Body.Status = map[bool]string{true: "duplicate", false: "ok"}[dup]
			return out, err
		})

	if s.d.Cfg.IsDevelopment() {
		// The fake carrier "sends" a signed tracking event (development only).
		r.Gin.POST("/dev/carriers/fake/:tracking/:status", func(c *gin.Context) {
			body, _ := json.Marshal(CarrierEvent{EventID: uuid.NewString(), TrackingNumber: c.Param("tracking"), Status: c.Param("status"),
				Location: "Test hub", OccurredAt: time.Now().UTC()})
			c.JSON(http.StatusOK, gin.H{"raw": string(body), "signature": sign(s.carrierSecret("fake"), body)})
		})
	}
}

// Jobs implements kit.Module.
func (s *Service) Jobs(*jobs.Registry) {}

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}
