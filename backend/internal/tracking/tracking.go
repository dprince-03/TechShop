// Package tracking accepts batched behaviour events from the web and apps, only when the visitor
// has consented to analytics (or personalisation for recommendation events). It links a guest's
// anonymous id to their account at sign-in. Plan: docs/recommendations.md (tracking, NDPA).
package tracking

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Event is one behaviour event.
type Event struct {
	Type         string          `json:"type" enum:"product_view,list_impression,rec_impression,rec_click,search,search_click,add_to_cart,remove_from_cart,save,unsave,begin_checkout,purchase,car_view,car_enquiry,share"`
	OccurredAt   time.Time       `json:"occurredAt"`
	Surface      string          `json:"surface,omitempty"`
	ProductID    *uuid.UUID      `json:"productId,omitempty"`
	ListingID    *uuid.UUID      `json:"listingId,omitempty"`
	CategoryID   *uuid.UUID      `json:"categoryId,omitempty"`
	Query        string          `json:"query,omitempty" maxLength:"120"`
	ResultsCount *int32          `json:"resultsCount,omitempty"`
	Position     *int16          `json:"position,omitempty"`
	RecRequestID string          `json:"recRequestId,omitempty"`
	Strategy     string          `json:"strategy,omitempty"`
	Properties   json.RawMessage `json:"properties,omitempty"`
}

// Service is the tracking module.
type Service struct{ d *kit.Deps }

// New creates the module.
func New(d *kit.Deps) *Service { return &Service{d: d} }

// Name implements kit.Module.
func (s *Service) Name() string { return "tracking" }

// Jobs implements kit.Module.
func (s *Service) Jobs(*jobs.Registry) {}

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	httpx.Register(r.Public, httpx.Op{ID: "trackEvents", Method: http.MethodPost, Path: "/api/v1/events", Tag: "Tracking", Auth: []string{auth.AudMarket, auth.AudWholesale, auth.AudCustomerApp}, Optional: true, MaxBody: 256 << 10,
		Summary: "Batched behaviour events (dropped unless the visitor consented)"},
		func(ctx context.Context, in *struct {
			Body struct {
				AnonymousID uuid.UUID `json:"anonymousId"`
				SessionID   string    `json:"sessionId,omitempty" maxLength:"64"`
				App         string    `json:"app" enum:"market,wholesale,customer_app"`
				DeviceClass string    `json:"deviceClass,omitempty" enum:"mobile,tablet,desktop,"`
				StateCode   string    `json:"stateCode,omitempty" pattern:"^([A-Z]{2})?$"`
				Events      []Event   `json:"events" minItems:"1" maxItems:"100"`
			}
		}) (*struct {
			Body struct {
				Accepted int `json:"accepted"`
				Dropped  int `json:"dropped" doc:"Events dropped for lack of consent"`
			}
		}, error) {
			var user *uuid.UUID
			if p := httpx.PrincipalFrom(ctx); p != nil {
				user = &p.UserID
				_ = s.d.Q.PersonalisationLinkIdentity(ctx, store.PersonalisationLinkIdentityParams{AnonymousID: in.Body.AnonymousID, UserID: p.UserID})
			}
			anon := in.Body.AnonymousID
			analytics, _ := s.d.Q.IdentityHasConsent(ctx, store.IdentityHasConsentParams{UserID: user, AnonymousID: &anon, Purpose: "analytics"})
			personal, _ := s.d.Q.IdentityHasConsent(ctx, store.IdentityHasConsentParams{UserID: user, AnonymousID: &anon, Purpose: "personalisation"})
			rows := []store.PersonalisationInsertEventParams{}
			dropped := 0
			for _, e := range in.Body.Events {
				recEvent := e.Type == "rec_impression" || e.Type == "rec_click"
				if !analytics || (recEvent && !personal) {
					dropped++
					continue
				}
				props := e.Properties
				if len(props) == 0 {
					props = []byte("{}")
				}
				rows = append(rows, store.PersonalisationInsertEventParams{OccurredAt: e.OccurredAt, AnonymousID: anon, UserID: user, SessionID: strOrNil(in.Body.SessionID),
					EventType: e.Type, App: in.Body.App, Surface: strOrNil(e.Surface), ProductID: e.ProductID, ListingID: e.ListingID, CategoryID: e.CategoryID,
					QueryNorm: strOrNil(e.Query), ResultsCount: e.ResultsCount, Position: e.Position, RecRequestID: strOrNil(e.RecRequestID), Strategy: strOrNil(e.Strategy),
					StateCode: strOrNil(in.Body.StateCode), DeviceClass: strOrNil(in.Body.DeviceClass), Properties: props})
			}
			if len(rows) > 0 {
				if _, err := s.d.Q.PersonalisationInsertEvent(ctx, rows); err != nil {
					return nil, httpx.DB(err, "event")
				}
			}
			out := &struct {
				Body struct {
					Accepted int `json:"accepted"`
					Dropped  int `json:"dropped" doc:"Events dropped for lack of consent"`
				}
			}{}
			out.Body.Accepted, out.Body.Dropped = len(rows), dropped
			return out, nil
		})
}

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
