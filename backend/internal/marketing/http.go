package marketing

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

// PromotionInput creates a promotion.
type PromotionInput struct {
	Name     string    `json:"name" minLength:"3" maxLength:"120"`
	Kind     string    `json:"kind" enum:"percentage,fixed_amount,free_delivery,flash_price"`
	Value    *int32    `json:"value,omitempty" minimum:"1" doc:"percentage: whole percent (≤ 90); fixed_amount: kobo"`
	FundedBy string    `json:"fundedBy,omitempty" enum:"techshop,seller" default:"techshop"`
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
	Targets  struct {
		CategoryIDs []uuid.UUID `json:"categoryIds,omitempty"`
		ListingIDs  []uuid.UUID `json:"listingIds,omitempty"`
	} `json:"targets"`
	FlashPrices []struct {
		ListingID    uuid.UUID `json:"listingId"`
		PriceKobo    int64     `json:"priceKobo" minimum:"1"`
		StockLimit   *int32    `json:"stockLimit,omitempty" minimum:"1"`
		PerUserLimit *int32    `json:"perUserLimit,omitempty" minimum:"1"`
	} `json:"flashPrices,omitempty"`
}

// PromotionView is a promotion with what it targets.
type PromotionView struct {
	store.MarketingPromotion
	Targets     []store.MarketingPromotionTarget       `json:"targets"`
	FlashPrices []store.MarketingPromotionListingPrice `json:"flashPrices"`
}

// CampaignView is a campaign with variants and results.
type CampaignView struct {
	store.MarketingCampaign
	Variants []store.MarketingCampaignVariant `json:"variants"`
	Stats    store.MarketingCampaignStatsRow  `json:"stats"`
}

// JourneyView is a journey with steps and enrollment counts.
type JourneyView struct {
	store.MarketingJourney
	Steps []store.MarketingJourneyStep `json:"steps"`
	Stats map[string]int32             `json:"stats"`
}

// CreatePromotion validates and stores a promotion. Seller-funded promotions must target only
// one seller's listings and never carry coupons (so the ledger can tell who funds a discount).
func (s *Service) CreatePromotion(ctx context.Context, in PromotionInput, actor uuid.UUID) (PromotionView, error) {
	var v PromotionView
	if !in.EndsAt.After(in.StartsAt) {
		return v, httpx.Invalid("bad_dates", "The end must be after the start.")
	}
	if in.Kind == "percentage" && (in.Value == nil || *in.Value > 90) {
		return v, httpx.Invalid("bad_value", "A percentage promotion needs a value between 1 and 90.")
	}
	if in.Kind == "fixed_amount" && in.Value == nil {
		return v, httpx.Invalid("bad_value", "A fixed-amount promotion needs a value in kobo.")
	}
	if in.Kind == "flash_price" && len(in.FlashPrices) == 0 {
		return v, httpx.Invalid("bad_value", "A flash promotion needs flash prices.")
	}
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if in.FundedBy == "seller" {
			if len(in.Targets.CategoryIDs) > 0 || len(in.Targets.ListingIDs) == 0 {
				return httpx.Invalid("seller_funded_targets", "A seller-funded promotion must target that seller's listings only.")
			}
			var seller uuid.UUID
			for i, id := range in.Targets.ListingIDs {
				l, err := tx.Q.CatalogGetListing(ctx, id)
				if err != nil {
					return httpx.Invalid("bad_listing", "Unknown listing "+id.String()+".")
				}
				if i > 0 && l.SellerID != seller {
					return httpx.Invalid("seller_funded_targets", "A seller-funded promotion can only cover one seller's listings.")
				}
				seller = l.SellerID
			}
		}
		status := "scheduled"
		if !in.StartsAt.After(time.Now()) {
			status = "live"
		}
		p, err := tx.Q.MarketingCreatePromotion(ctx, store.MarketingCreatePromotionParams{Name: in.Name, Kind: in.Kind, Value: in.Value, FundedBy: in.FundedBy,
			StartsAt: in.StartsAt, EndsAt: in.EndsAt, Status: status, CreatedBy: actor})
		if err != nil {
			return httpx.DB(err, "promotion")
		}
		for _, c := range in.Targets.CategoryIDs {
			c := c
			if _, err := tx.Q.MarketingAddTarget(ctx, store.MarketingAddTargetParams{PromotionID: p.ID, CategoryID: &c}); err != nil {
				return httpx.DB(err, "promotion target")
			}
		}
		for _, l := range in.Targets.ListingIDs {
			l := l
			if _, err := tx.Q.MarketingAddTarget(ctx, store.MarketingAddTargetParams{PromotionID: p.ID, ListingID: &l}); err != nil {
				return httpx.DB(err, "promotion target")
			}
		}
		for _, fp := range in.FlashPrices {
			l, err := tx.Q.CatalogGetListing(ctx, fp.ListingID)
			if err != nil {
				return httpx.Invalid("bad_listing", "Unknown listing "+fp.ListingID.String()+".")
			}
			if fp.PriceKobo >= l.PriceKobo {
				return httpx.Invalid("flash_not_lower", "A flash price must be below the listing's price.")
			}
			if _, err := tx.Q.MarketingSetListingPrice(ctx, store.MarketingSetListingPriceParams{PromotionID: p.ID, ListingID: fp.ListingID, PriceKobo: fp.PriceKobo,
				StockLimit: fp.StockLimit, PerUserLimit: fp.PerUserLimit}); err != nil {
				return httpx.DB(err, "flash price")
			}
		}
		v.MarketingPromotion = p
		return tx.Audit("promotion.created", "promotion", p.ID.String(), in)
	})
	if err != nil {
		return v, err
	}
	return s.promotionView(ctx, v.ID)
}

func (s *Service) promotionView(ctx context.Context, id uuid.UUID) (PromotionView, error) {
	p, err := s.d.Q.MarketingGetPromotion(ctx, id)
	if err != nil {
		return PromotionView{}, httpx.NotFound("Promotion")
	}
	v := PromotionView{MarketingPromotion: p}
	if v.Targets, err = s.d.Q.MarketingPromotionTargets(ctx, id); err != nil {
		return v, err
	}
	v.FlashPrices, err = s.d.Q.MarketingPromotionListingPrices(ctx, id)
	return v, err
}

func (s *Service) campaignView(ctx context.Context, id uuid.UUID) (CampaignView, error) {
	c, err := s.d.Q.MarketingGetCampaign(ctx, id)
	if err != nil {
		return CampaignView{}, httpx.NotFound("Campaign")
	}
	v := CampaignView{MarketingCampaign: c}
	if v.Variants, err = s.d.Q.MarketingCampaignVariants(ctx, id); err != nil {
		return v, err
	}
	v.Stats, err = s.d.Q.MarketingCampaignStats(ctx, &id)
	return v, err
}

func (s *Service) journeyView(ctx context.Context, id uuid.UUID) (JourneyView, error) {
	j, err := s.d.Q.MarketingGetJourney(ctx, id)
	if err != nil {
		return JourneyView{}, httpx.NotFound("Journey")
	}
	v := JourneyView{MarketingJourney: j, Stats: map[string]int32{}}
	if v.Steps, err = s.d.Q.MarketingJourneySteps(ctx, id); err != nil {
		return v, err
	}
	rows, err := s.d.Q.MarketingJourneyStats(ctx, id)
	for _, r := range rows {
		v.Stats[r.Status] = r.N
	}
	return v, err
}

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	op := func(o httpx.Op) httpx.Op { o.Tag = "Staff · Marketing"; o.Auth = []string{auth.AudStaff}; return o }
	actor := func(ctx context.Context) uuid.UUID { return httpx.MustPrincipal(ctx).UserID }

	// ---- Promotions and flash deals
	httpx.Register(g, op(httpx.Op{ID: "listPromotions", Method: http.MethodGet, Path: "/api/v1/staff/marketing/promotions", Perm: "marketing.view", Summary: "Promotions"}),
		func(ctx context.Context, in *struct {
			Status string `query:"status" enum:"draft,scheduled,live,ended,cancelled,"`
		}) (*struct{ Body []store.MarketingPromotion }, error) {
			p := store.MarketingListPromotionsParams{Limit: 200}
			if in.Status != "" {
				p.Status = &in.Status
			}
			rows, err := s.d.Q.MarketingListPromotions(ctx, p)
			return &struct{ Body []store.MarketingPromotion }{nonNil(rows)}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "createPromotion", Method: http.MethodPost, Path: "/api/v1/staff/marketing/promotions", Perm: "marketing.edit", Status: 201,
		Summary: "Create a promotion or flash deal (real start and end times)"}),
		func(ctx context.Context, in *struct{ Body PromotionInput }) (*struct{ Body PromotionView }, error) {
			v, err := s.CreatePromotion(ctx, in.Body, actor(ctx))
			return &struct{ Body PromotionView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "getPromotion", Method: http.MethodGet, Path: "/api/v1/staff/marketing/promotions/{id}", Perm: "marketing.view", Summary: "Promotion detail"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body PromotionView }, error) {
			v, err := s.promotionView(ctx, in.ID)
			return &struct{ Body PromotionView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "endPromotion", Method: http.MethodPost, Path: "/api/v1/staff/marketing/promotions/{id}/{action}", Perm: "marketing.edit",
		Summary: "End a live promotion or cancel a scheduled one"}),
		func(ctx context.Context, in *struct {
			ID     uuid.UUID `path:"id"`
			Action string    `path:"action" enum:"end,cancel"`
		}) (*struct{ Body PromotionView }, error) {
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				p, err := tx.Q.MarketingGetPromotion(ctx, in.ID)
				if err != nil {
					return httpx.NotFound("Promotion")
				}
				next := map[string]string{"end": "ended", "cancel": "cancelled"}[in.Action]
				if (in.Action == "end" && p.Status != "live") || (in.Action == "cancel" && p.Status != "scheduled" && p.Status != "draft") {
					return httpx.Conflict("bad_transition", "Can't "+in.Action+" a promotion that is "+p.Status+".")
				}
				if _, err := tx.Q.MarketingSetPromotionStatus(ctx, store.MarketingSetPromotionStatusParams{ID: in.ID, Status: next}); err != nil {
					return err
				}
				return tx.Audit("promotion."+in.Action, "promotion", in.ID.String(), nil)
			})
			if err != nil {
				return nil, err
			}
			v, err := s.promotionView(ctx, in.ID)
			return &struct{ Body PromotionView }{v}, err
		})

	// ---- Coupons
	httpx.Register(g, op(httpx.Op{ID: "listCoupons", Method: http.MethodGet, Path: "/api/v1/staff/marketing/coupons", Perm: "marketing.view", Summary: "Coupons"}),
		func(ctx context.Context, _ *struct{}) (*struct {
			Body []store.MarketingListCouponsRow
		}, error) {
			rows, err := s.d.Q.MarketingListCoupons(ctx, 200)
			return &struct {
				Body []store.MarketingListCouponsRow
			}{nonNil(rows)}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "createCoupon", Method: http.MethodPost, Path: "/api/v1/staff/marketing/coupons", Perm: "marketing.edit", Status: 201,
		Summary: "Create a coupon: one shared code, or unique single-use codes"}),
		func(ctx context.Context, in *struct {
			Body struct {
				PromotionID    uuid.UUID `json:"promotionId"`
				Code           string    `json:"code,omitempty" pattern:"^[A-Za-z0-9-]{4,30}$"`
				UniqueCodes    bool      `json:"uniqueCodes,omitempty"`
				MaxRedemptions *int32    `json:"maxRedemptions,omitempty" minimum:"1"`
				PerUserLimit   int32     `json:"perUserLimit,omitempty" minimum:"1" default:"1"`
				MinOrderKobo   *int64    `json:"minOrderKobo,omitempty" minimum:"0"`
			}
		}) (*struct{ Body store.MarketingCoupon }, error) {
			var out store.MarketingCoupon
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				p, err := tx.Q.MarketingGetPromotion(ctx, in.Body.PromotionID)
				if err != nil {
					return httpx.Invalid("bad_promotion", "Unknown promotion.")
				}
				if p.FundedBy == "seller" || p.Kind == "flash_price" {
					return httpx.Invalid("coupon_not_allowed", "Coupons can't be attached to seller-funded or flash promotions.")
				}
				if !in.Body.UniqueCodes && in.Body.Code == "" {
					return httpx.Invalid("code_required", "A shared coupon needs a code.")
				}
				var code *string
				if in.Body.Code != "" {
					c := strings.ToUpper(in.Body.Code)
					code = &c
				}
				out, err = tx.Q.MarketingCreateCoupon(ctx, store.MarketingCreateCouponParams{Code: code, PromotionID: p.ID, IsUniqueCodes: in.Body.UniqueCodes,
					MaxRedemptions: in.Body.MaxRedemptions, PerUserLimit: in.Body.PerUserLimit, MinOrderKobo: in.Body.MinOrderKobo})
				if err != nil {
					return httpx.DB(err, "coupon")
				}
				return tx.Audit("coupon.created", "coupon", out.ID.String(), in.Body)
			})
			return &struct{ Body store.MarketingCoupon }{out}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "generateCouponCodes", Method: http.MethodPost, Path: "/api/v1/staff/marketing/coupons/{id}/codes", Perm: "marketing.edit", Status: 201,
		Summary: "Generate unique single-use codes"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Count       int    `json:"count" minimum:"1" maximum:"5000"`
				Prefix      string `json:"prefix,omitempty" pattern:"^[A-Z]{2,8}$" default:"TS"`
				ExpiresDays int    `json:"expiresDays,omitempty" minimum:"1" maximum:"365"`
			}
		}) (*struct{ Body []string }, error) {
			var codes []string
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var exp *time.Time
				if in.Body.ExpiresDays > 0 {
					t := time.Now().AddDate(0, 0, in.Body.ExpiresDays)
					exp = &t
				}
				for len(codes) < in.Body.Count {
					code := in.Body.Prefix + "-" + randomCode(6)
					err := tx.Savepoint(ctx, func(sub *uow.Tx) error {
						_, err := sub.Q.MarketingCreateCouponCode(ctx, store.MarketingCreateCouponCodeParams{CouponID: in.ID, Code: code, ExpiresAt: exp})
						return err
					})
					if err == nil {
						codes = append(codes, code)
					} else if !strings.Contains(err.Error(), "23505") && !strings.Contains(err.Error(), "duplicate") {
						return httpx.DB(err, "coupon code")
					}
				}
				return tx.Audit("coupon.codes_generated", "coupon", in.ID.String(), map[string]any{"count": len(codes)})
			})
			return &struct{ Body []string }{codes}, err
		})

	// ---- Segments
	httpx.Register(g, op(httpx.Op{ID: "listSegments", Method: http.MethodGet, Path: "/api/v1/staff/marketing/segments", Perm: "marketing.view", Summary: "Saved audiences"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.MarketingSegment }, error) {
			rows, err := s.d.Q.MarketingListSegments(ctx)
			return &struct{ Body []store.MarketingSegment }{nonNil(rows)}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "previewSegment", Method: http.MethodPost, Path: "/api/v1/staff/marketing/segments/preview", Perm: "marketing.segments",
		Summary: "Count who matches audience rules"}),
		func(ctx context.Context, in *struct{ Body Rules }) (*struct {
			Body struct {
				Count int `json:"count"`
			}
		}, error) {
			out := &struct {
				Body struct {
					Count int `json:"count"`
				}
			}{}
			var err error
			out.Body.Count, err = s.Count(ctx, in.Body)
			return out, err
		})
	httpx.Register(g, op(httpx.Op{ID: "createSegment", Method: http.MethodPost, Path: "/api/v1/staff/marketing/segments", Perm: "marketing.segments", Status: 201, Summary: "Save an audience"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Name  string `json:"name" minLength:"3" maxLength:"120"`
				Rules Rules  `json:"rules"`
			}
		}) (*struct{ Body store.MarketingSegment }, error) {
			n, err := s.Count(ctx, in.Body.Rules)
			if err != nil {
				return nil, err
			}
			var out store.MarketingSegment
			err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				rules, _ := json.Marshal(in.Body.Rules)
				var err error
				out, err = tx.Q.MarketingCreateSegment(ctx, store.MarketingCreateSegmentParams{Name: in.Body.Name, Rules: rules, CreatedBy: actor(ctx)})
				if err != nil {
					return err
				}
				if err := tx.Q.MarketingSetSegmentCount(ctx, store.MarketingSetSegmentCountParams{ID: out.ID, LastCount: ptr32(n)}); err != nil {
					return err
				}
				c := int32(n)
				out.LastCount = &c
				return nil
			})
			return &struct{ Body store.MarketingSegment }{out}, err
		})

	// ---- Campaigns (creator ≠ approver)
	httpx.Register(g, op(httpx.Op{ID: "listCampaigns", Method: http.MethodGet, Path: "/api/v1/staff/marketing/campaigns", Perm: "marketing.view", Summary: "Campaigns"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.MarketingCampaign }, error) {
			rows, err := s.d.Q.MarketingListCampaigns(ctx, 200)
			return &struct{ Body []store.MarketingCampaign }{nonNil(rows)}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "createCampaign", Method: http.MethodPost, Path: "/api/v1/staff/marketing/campaigns", Perm: "marketing.edit", Status: 201,
		Summary: "Draft a campaign with variants (A/B) and a holdout"}),
		func(ctx context.Context, in *struct{ Body CampaignInput }) (*struct{ Body CampaignView }, error) {
			c, err := s.CreateCampaign(ctx, in.Body, actor(ctx))
			if err != nil {
				return nil, err
			}
			v, err := s.campaignView(ctx, c.ID)
			return &struct{ Body CampaignView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "getCampaign", Method: http.MethodGet, Path: "/api/v1/staff/marketing/campaigns/{id}", Perm: "marketing.view", Summary: "Campaign with variants and results"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body CampaignView }, error) {
			v, err := s.campaignView(ctx, in.ID)
			return &struct{ Body CampaignView }{v}, err
		})
	transition := func(id, action, perm string, stepUp bool) {
		httpx.Register(g, op(httpx.Op{ID: id, Method: http.MethodPost, Path: "/api/v1/staff/marketing/campaigns/{id}/" + action, Perm: perm, StepUp: stepUp,
			Summary: strings.ToUpper(action[:1]) + action[1:] + " a campaign"}),
			func(ctx context.Context, in *struct {
				ID   uuid.UUID `path:"id"`
				Body *struct {
					SendAt *time.Time `json:"sendAt,omitempty" doc:"Approve only: when to send (default now)"`
				}
			}) (*struct{ Body CampaignView }, error) {
				var at *time.Time
				if in.Body != nil {
					at = in.Body.SendAt
				}
				if _, err := s.Transition(ctx, in.ID, action, actor(ctx), at); err != nil {
					return nil, err
				}
				v, err := s.campaignView(ctx, in.ID)
				return &struct{ Body CampaignView }{v}, err
			})
	}
	transition("submitCampaign", "submit", "marketing.edit", false)
	transition("approveCampaign", "approve", "marketing.approve", true)
	transition("pauseCampaign", "pause", "marketing.edit", false)
	transition("resumeCampaign", "resume", "marketing.approve", true)
	transition("cancelCampaign", "cancel", "marketing.edit", false)

	// ---- Journeys
	httpx.Register(g, op(httpx.Op{ID: "listJourneys", Method: http.MethodGet, Path: "/api/v1/staff/marketing/journeys", Perm: "marketing.view", Summary: "Automations"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.MarketingJourney }, error) {
			rows, err := s.d.Q.MarketingListJourneys(ctx)
			return &struct{ Body []store.MarketingJourney }{nonNil(rows)}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "createJourney", Method: http.MethodPost, Path: "/api/v1/staff/marketing/journeys", Perm: "marketing.journeys", Status: 201,
		Summary: "Create a journey (draft): trigger, entry rules and steps"}),
		func(ctx context.Context, in *struct{ Body JourneyInput }) (*struct{ Body JourneyView }, error) {
			j, err := s.CreateJourney(ctx, in.Body, actor(ctx))
			if err != nil {
				return nil, err
			}
			v, err := s.journeyView(ctx, j.ID)
			return &struct{ Body JourneyView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "getJourney", Method: http.MethodGet, Path: "/api/v1/staff/marketing/journeys/{id}", Perm: "marketing.view", Summary: "Journey with steps and counts"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body JourneyView }, error) {
			v, err := s.journeyView(ctx, in.ID)
			return &struct{ Body JourneyView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "setJourneyStatus", Method: http.MethodPost, Path: "/api/v1/staff/marketing/journeys/{id}/{action}", Perm: "marketing.journeys",
		Summary: "Activate, pause or archive a journey"}),
		func(ctx context.Context, in *struct {
			ID     uuid.UUID `path:"id"`
			Action string    `path:"action" enum:"activate,pause,archive"`
		}) (*struct{ Body JourneyView }, error) {
			next := map[string]string{"activate": "active", "pause": "paused", "archive": "archived"}[in.Action]
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				if _, err := tx.Q.MarketingSetJourneyStatus(ctx, store.MarketingSetJourneyStatusParams{ID: in.ID, Status: next}); err != nil {
					return httpx.NotFound("Journey")
				}
				return tx.Audit("journey."+in.Action, "journey", in.ID.String(), nil)
			})
			if err != nil {
				return nil, err
			}
			v, err := s.journeyView(ctx, in.ID)
			return &struct{ Body JourneyView }{v}, err
		})

	// ---- Public: tracked links and one-click unsubscribe (signed with TOKEN_HMAC_KEY)
	httpx.Register(g, httpx.Op{ID: "followLink", Method: http.MethodGet, Path: "/api/v1/l/{id}", Tag: "Messaging", Status: http.StatusFound,
		Summary: "Tracked click redirect (techshop.ng targets only)"},
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			User string    `query:"u"`
			Sig  string    `query:"s"`
		}) (*struct {
			Location string `header:"Location"`
		}, error) {
			l, err := s.d.Q.MarketingGetLink(ctx, in.ID)
			if err != nil {
				return nil, httpx.NotFound("Link")
			}
			if u, err := uuid.Parse(in.User); err == nil && s.validSig(in.Sig, "link", in.ID.String(), in.User) && l.CampaignID != nil {
				_ = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
					n, err := tx.Q.MarketingCampaignNotification(ctx, store.MarketingCampaignNotificationParams{CampaignID: l.CampaignID, UserID: &u})
					if err != nil {
						return nil
					}
					if err := tx.Q.MessagingMarkClicked(ctx, n); err != nil {
						return err
					}
					return tx.Q.MessagingInsertEvent(ctx, store.MessagingInsertEventParams{NotificationID: n, Kind: "clicked", Meta: []byte(`{}`)})
				})
			}
			return &struct {
				Location string `header:"Location"`
			}{l.Url}, nil
		})
	type unsubIn struct {
		User    uuid.UUID `query:"u"`
		Channel string    `query:"c" enum:"sms,email,push,whatsapp"`
		Sig     string    `query:"s"`
	}
	type unsubOut struct {
		Body struct {
			Channel      string `json:"channel"`
			Unsubscribed bool   `json:"unsubscribed"`
		}
	}
	check := func(in *unsubIn) error {
		if !s.validSig(in.Sig, "unsub", in.User.String(), in.Channel) {
			return httpx.E(403, "bad_signature", "This unsubscribe link is invalid.")
		}
		return nil
	}
	httpx.Register(g, httpx.Op{ID: "checkUnsubscribe", Method: http.MethodGet, Path: "/api/v1/unsubscribe", Tag: "Messaging", Summary: "Validate an unsubscribe link"},
		func(ctx context.Context, in *unsubIn) (*unsubOut, error) {
			if err := check(in); err != nil {
				return nil, err
			}
			out := &unsubOut{}
			out.Body.Channel = in.Channel
			return out, nil
		})
	httpx.Register(g, httpx.Op{ID: "unsubscribe", Method: http.MethodPost, Path: "/api/v1/unsubscribe", Tag: "Messaging", Summary: "One-click unsubscribe from marketing on a channel"},
		func(ctx context.Context, in *unsubIn) (*unsubOut, error) {
			if err := check(in); err != nil {
				return nil, err
			}
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				return tx.Q.IdentityInsertConsent(ctx, store.IdentityInsertConsentParams{UserID: &in.User, Purpose: "marketing_" + in.Channel, Granted: false,
					PolicyVersion: "unsubscribe", Source: "unsubscribe_link"})
			})
			out := &unsubOut{}
			out.Body.Channel, out.Body.Unsubscribed = in.Channel, err == nil
			return out, err
		})
}

func nonNil[T any](rows []T) []T {
	if rows == nil {
		return []T{}
	}
	return rows
}

func ptr32(n int) *int32 {
	v := int32(n)
	return &v
}
