package sellers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/modules/catalog"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Seller Centre listings (M4 slice 2, search-catalogue.md §3). The catalogue module owns the
// listing rules (automatic checks, FCCPA "was" price guard, moderation); this file scopes them
// to the caller's shop. Marketplace sellers ship their own stock for now (fulfilled_by =
// seller): TechShop-fulfilled vendor stock needs inbound receiving, which isn't built yet.

// Roles that manage listings (identity-access.md §7.3: finance only sees money).
var listingRoles = []string{"owner", "manager", "staff"}

// Listing is a shop's offer as Seller Centre shows it (no internal risk score).
type Listing struct {
	ID               uuid.UUID  `json:"id"`
	VariantID        uuid.UUID  `json:"variantId"`
	ProductName      string     `json:"productName"`
	ProductSlug      string     `json:"productSlug"`
	VariantName      string     `json:"variantName"`
	SKU              string     `json:"sku"`
	Condition        string     `json:"condition"`
	Grade            *string    `json:"grade"`
	Price            int64      `json:"price" doc:"Kobo"`
	CompareAt        *int64     `json:"compareAt" doc:"Kobo"`
	SellerStock      *int32     `json:"sellerStock"`
	HandlingDays     int16      `json:"handlingDays"`
	WarrantyProvider string     `json:"warrantyProvider"`
	WarrantyMonths   int16      `json:"warrantyMonths"`
	ConditionNotes   *string    `json:"conditionNotes"`
	BatteryHealthPct *int16     `json:"batteryHealthPct"`
	Status           string     `json:"status" enum:"draft,auto_check,in_review,changes_requested,active,paused,suspended,rejected"`
	PublishedAt      *time.Time `json:"publishedAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// NewListing is a seller's offer on a catalogue variant. Sellers ship their own stock, so there
// is no fulfilment choice and stock is required.
type NewListing struct {
	VariantID        uuid.UUID `json:"variantId"`
	Condition        string    `json:"condition" enum:"new,uk_used,refurbished,open_box"`
	Grade            string    `json:"grade,omitempty" enum:"A,B,C,"`
	Price            int64     `json:"price" minimum:"1" doc:"Kobo"`
	CompareAt        *int64    `json:"compareAt,omitempty" doc:"Kobo. A new listing has no price history, so this is refused (FCCPA)"`
	SellerStock      int32     `json:"sellerStock" minimum:"0" maximum:"100000"`
	HandlingDays     int16     `json:"handlingDays,omitempty" minimum:"0" maximum:"10" doc:"Days to hand over to delivery (default 1)"`
	WarrantyProvider string    `json:"warrantyProvider,omitempty" enum:"manufacturer,techshop,seller,none,"`
	WarrantyMonths   int16     `json:"warrantyMonths,omitempty" minimum:"0" maximum:"60"`
	ConditionNotes   string    `json:"conditionNotes,omitempty" maxLength:"500" doc:"Required unless new"`
	BatteryHealthPct *int16    `json:"batteryHealthPct,omitempty" minimum:"1" maximum:"100"`
}

func (n NewListing) catalogInput() catalog.ListingInput {
	stock := n.SellerStock
	return catalog.ListingInput{VariantID: n.VariantID, Condition: n.Condition, Grade: n.Grade, Price: n.Price, CompareAt: n.CompareAt,
		FulfilledBy: "seller", SellerStock: &stock, HandlingDays: n.HandlingDays, WarrantyProvider: n.WarrantyProvider,
		WarrantyMonths: n.WarrantyMonths, ConditionNotes: n.ConditionNotes, BatteryHealthPct: n.BatteryHealthPct}
}

// ListingReview is a moderation decision the seller can read (who decided stays internal).
type ListingReview struct {
	Decision  string    `json:"decision" enum:"auto_passed,auto_flagged,approved,changes_requested,rejected,suspended"`
	Reasons   []string  `json:"reasons"`
	Note      *string   `json:"note"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListingDetail is a listing with its moderation history.
type ListingDetail struct {
	Listing
	Reviews []ListingReview `json:"reviews"`
}

// ListingResult is a created or resubmitted listing with the automatic check's outcome.
type ListingResult struct {
	Listing Listing             `json:"listing"`
	Check   catalog.CheckResult `json:"check"`
}

func listingView(l store.CatalogListing, product, slug, variant, sku string) Listing {
	return Listing{ID: l.ID, VariantID: l.VariantID, ProductName: product, ProductSlug: slug, VariantName: variant, SKU: sku,
		Condition: l.Condition, Grade: l.Grade, Price: l.PriceKobo, CompareAt: l.CompareAtKobo, SellerStock: l.SellerStock,
		HandlingDays: l.HandlingDays, WarrantyProvider: l.WarrantyProvider, WarrantyMonths: l.WarrantyMonths, ConditionNotes: l.ConditionNotes,
		BatteryHealthPct: l.BatteryHealthPct, Status: l.Status, PublishedAt: l.PublishedAt, CreatedAt: l.CreatedAt, UpdatedAt: l.UpdatedAt}
}

// ownListing loads a listing and checks it belongs to the shop. Row-level security lets a
// seller read other shops' *active* listings (they're public), so this check is required.
func ownListing(ctx context.Context, q *store.Queries, shopID, listingID uuid.UUID) (store.CatalogGetListingRow, store.CatalogListing, error) {
	r, err := q.CatalogGetListing(ctx, listingID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && r.SellerID != shopID) {
		return r, store.CatalogListing{}, httpx.NotFound("Listing")
	}
	if err != nil {
		return r, store.CatalogListing{}, err
	}
	return r, store.CatalogListing{ID: r.ID, SellerID: r.SellerID, VariantID: r.VariantID, Condition: r.Condition, Grade: r.Grade,
		PriceKobo: r.PriceKobo, CompareAtKobo: r.CompareAtKobo, Currency: r.Currency, FulfilledBy: r.FulfilledBy, SellerStock: r.SellerStock,
		MinOrderQuantity: r.MinOrderQuantity, HandlingDays: r.HandlingDays, WarrantyProvider: r.WarrantyProvider, WarrantyMonths: r.WarrantyMonths,
		ConditionNotes: r.ConditionNotes, BatteryHealthPct: r.BatteryHealthPct, Status: r.Status, RiskScore: r.RiskScore,
		PublishedAt: r.PublishedAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}, nil
}

// withListing runs fn on one of the shop's listings and returns the updated view.
func (s *Service) withListing(ctx context.Context, shopID, listingID uuid.UUID, fn func(tx *uow.Tx, shop store.SellersSeller, l store.CatalogListing) (store.CatalogListing, error)) (Listing, error) {
	var out Listing
	err := s.inShop(ctx, shopID, listingRoles, func(tx *uow.Tx, shop store.SellersSeller, _ string) error {
		row, l, err := ownListing(ctx, tx.Q, shopID, listingID)
		if err != nil {
			return err
		}
		updated, err := fn(tx, shop, l)
		if err != nil {
			return err
		}
		out = listingView(updated, row.ProductName, row.ProductSlug, row.VariantName, row.Sku)
		return tx.Audit("seller.listing_updated", "listing", listingID.String(), map[string]string{"status": updated.Status})
	})
	return out, err
}

func (s *Service) listingRoutes(r *kit.Routes) {
	g := r.Public
	sel := func(o httpx.Op) httpx.Op { o.Tag = "Seller Centre · Listings"; o.Auth = sellerAud; return o }
	type idPath struct {
		ShopID    uuid.UUID `path:"shopId"`
		ListingID uuid.UUID `path:"listingId"`
	}
	base := "/api/v1/seller/shops/{shopId}/listings"

	httpx.Register(g, sel(httpx.Op{ID: "sellerCreateListing", Method: http.MethodPost, Path: base, Status: 201, Summary: "Offer a catalogue variant (automatic checks decide: live or review)"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			Body   NewListing
		}) (*struct{ Body ListingResult }, error) {
			var out ListingResult
			err := s.inShop(ctx, in.ShopID, listingRoles, func(tx *uow.Tx, shop store.SellersSeller, _ string) error {
				if shop.Status != "active" {
					return httpx.Conflict("shop_not_active", "Your shop must be verified before you can list products.")
				}
				uid := httpx.MustPrincipal(ctx).UserID
				l, chk, err := s.catalog.CreateListing(ctx, tx, shop.ID, shop.Type, in.Body.catalogInput(), &uid)
				if err != nil {
					return err
				}
				row, _, err := ownListing(ctx, tx.Q, shop.ID, l.ID)
				if err != nil {
					return err
				}
				out = ListingResult{Listing: listingView(l, row.ProductName, row.ProductSlug, row.VariantName, row.Sku), Check: chk}
				return tx.Audit("seller.listing_created", "listing", l.ID.String(), map[string]any{"status": l.Status, "reasons": chk.Reasons})
			})
			return &struct{ Body ListingResult }{out}, err
		})

	httpx.Register(g, sel(httpx.Op{ID: "sellerListListings", Method: http.MethodGet, Path: base, Summary: "Your listings, newest first"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			Status string    `query:"status" enum:"draft,in_review,changes_requested,active,paused,suspended,rejected,"`
			Cursor string    `query:"cursor"`
			Limit  int       `query:"limit"`
		}) (*struct{ Body httpx.List[Listing] }, error) {
			before, err := s.d.Cursors.Before(in.Cursor)
			if err != nil {
				return nil, httpx.BadCursor()
			}
			lim := httpx.Limit(in.Limit)
			var views []Listing
			err = s.inShop(ctx, in.ShopID, anyRole, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
				rows, err := tx.Q.CatalogSellerListingsPage(ctx, store.CatalogSellerListingsPageParams{SellerID: in.ShopID, Status: optional(in.Status), Before: before, Lim: int32(lim + 1)})
				views = make([]Listing, len(rows))
				for i, r := range rows {
					views[i] = listingView(store.CatalogListing{ID: r.ID, SellerID: r.SellerID, VariantID: r.VariantID, Condition: r.Condition, Grade: r.Grade,
						PriceKobo: r.PriceKobo, CompareAtKobo: r.CompareAtKobo, SellerStock: r.SellerStock, HandlingDays: r.HandlingDays,
						WarrantyProvider: r.WarrantyProvider, WarrantyMonths: r.WarrantyMonths, ConditionNotes: r.ConditionNotes,
						BatteryHealthPct: r.BatteryHealthPct, Status: r.Status, PublishedAt: r.PublishedAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt},
						r.ProductName, r.ProductSlug, r.VariantName, r.Sku)
				}
				return err
			})
			if err != nil {
				return nil, err
			}
			return &struct{ Body httpx.List[Listing] }{httpx.Paginate(s.d.Cursors, views, lim, func(l Listing) time.Time { return l.CreatedAt })}, nil
		})

	httpx.Register(g, sel(httpx.Op{ID: "sellerGetListing", Method: http.MethodGet, Path: base + "/{listingId}", Summary: "A listing with its moderation history"}),
		func(ctx context.Context, in *idPath) (*struct{ Body ListingDetail }, error) {
			var out ListingDetail
			err := s.inShop(ctx, in.ShopID, anyRole, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
				row, l, err := ownListing(ctx, tx.Q, in.ShopID, in.ListingID)
				if err != nil {
					return err
				}
				reviews, err := tx.Q.CatalogListListingReviews(ctx, l.ID)
				out = ListingDetail{Listing: listingView(l, row.ProductName, row.ProductSlug, row.VariantName, row.Sku), Reviews: make([]ListingReview, len(reviews))}
				for i, rv := range reviews {
					out.Reviews[i] = ListingReview{Decision: rv.Decision, Reasons: rv.Reasons, Note: rv.Note, CreatedAt: rv.CreatedAt}
				}
				return err
			})
			return &struct{ Body ListingDetail }{out}, err
		})

	httpx.Register(g, sel(httpx.Op{ID: "sellerSetListingPrice", Method: http.MethodPatch, Path: base + "/{listingId}/price", Summary: "Change the price (a deep cut sends a live listing back to review)"}),
		func(ctx context.Context, in *struct {
			ShopID    uuid.UUID `path:"shopId"`
			ListingID uuid.UUID `path:"listingId"`
			Body      struct {
				Price     int64  `json:"price" minimum:"1" doc:"Kobo"`
				CompareAt *int64 `json:"compareAt,omitempty" doc:"Kobo; at most the 30-day maximum price (FCCPA)"`
			}
		}) (*struct{ Body Listing }, error) {
			uid := httpx.MustPrincipal(ctx).UserID
			l, err := s.withListing(ctx, in.ShopID, in.ListingID, func(tx *uow.Tx, _ store.SellersSeller, l store.CatalogListing) (store.CatalogListing, error) {
				if l.Status == "suspended" || l.Status == "rejected" {
					return l, httpx.Conflict("bad_transition", "This listing can't be changed.")
				}
				return s.catalog.RepriceMarketplace(ctx, tx, l, in.Body.Price, in.Body.CompareAt, &uid)
			})
			return &struct{ Body Listing }{l}, err
		})

	httpx.Register(g, sel(httpx.Op{ID: "sellerSetListingStock", Method: http.MethodPut, Path: base + "/{listingId}/stock", Summary: "Set how many you have in stock"}),
		func(ctx context.Context, in *struct {
			ShopID    uuid.UUID `path:"shopId"`
			ListingID uuid.UUID `path:"listingId"`
			Body      struct {
				Stock int32 `json:"stock" minimum:"0" maximum:"100000"`
			}
		}) (*struct{ Body Listing }, error) {
			l, err := s.withListing(ctx, in.ShopID, in.ListingID, func(tx *uow.Tx, _ store.SellersSeller, l store.CatalogListing) (store.CatalogListing, error) {
				out, err := tx.Q.CatalogSetListingStock(ctx, store.CatalogSetListingStockParams{ID: l.ID, SellerStock: &in.Body.Stock})
				if err == nil {
					tx.Emit("listing", l.ID, "listing.changed", map[string]any{"listingId": l.ID, "stock": in.Body.Stock})
				}
				return out, err
			})
			return &struct{ Body Listing }{l}, err
		})

	for _, act := range []struct {
		id, verb string
		paused   bool
	}{{"sellerPauseListing", "pause", true}, {"sellerResumeListing", "resume", false}} {
		httpx.Register(g, sel(httpx.Op{ID: act.id, Method: http.MethodPost, Path: base + "/{listingId}/" + act.verb, Summary: map[bool]string{true: "Take a live listing off sale", false: "Put a paused listing back on sale"}[act.paused]}),
			func(ctx context.Context, in *idPath) (*struct{ Body Listing }, error) {
				l, err := s.withListing(ctx, in.ShopID, in.ListingID, func(tx *uow.Tx, _ store.SellersSeller, l store.CatalogListing) (store.CatalogListing, error) {
					return s.catalog.SetPaused(ctx, tx, l, act.paused)
				})
				return &struct{ Body Listing }{l}, err
			})
	}

	httpx.Register(g, sel(httpx.Op{ID: "sellerResubmitListing", Method: http.MethodPut, Path: base + "/{listingId}", Summary: "Fix a draft or a listing with changes requested, and resubmit"}),
		func(ctx context.Context, in *struct {
			ShopID    uuid.UUID `path:"shopId"`
			ListingID uuid.UUID `path:"listingId"`
			Body      catalog.ListingEdit
		}) (*struct{ Body ListingResult }, error) {
			var chk catalog.CheckResult
			l, err := s.withListing(ctx, in.ShopID, in.ListingID, func(tx *uow.Tx, shop store.SellersSeller, l store.CatalogListing) (store.CatalogListing, error) {
				out, c, err := s.catalog.EditAndResubmit(ctx, tx, l, shop.Type, in.Body)
				chk = c
				return out, err
			})
			return &struct{ Body ListingResult }{ListingResult{Listing: l, Check: chk}}, err
		})
}
