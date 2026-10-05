package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/realtime"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

var shopperAuds = []string{auth.AudMarket, auth.AudWholesale, auth.AudCustomerApp}
var staff = []string{auth.AudStaff}

// Offer is one seller's offer with its buy-box score.
type Offer struct {
	ListingID      uuid.UUID `json:"listingId"`
	VariantID      uuid.UUID `json:"variantId"`
	VariantName    string    `json:"variantName"`
	SellerName     string    `json:"sellerName"`
	SellerSlug     string    `json:"sellerSlug"`
	SellerType     string    `json:"sellerType"`
	SellerRating   *float64  `json:"sellerRating,omitempty"`
	Condition      string    `json:"condition"`
	Grade          *string   `json:"grade,omitempty"`
	Price          int64     `json:"price" doc:"Kobo"`
	CompareAt      *int64    `json:"compareAt,omitempty"`
	DeliveryFee    int64     `json:"deliveryFee" doc:"To the requested state, kobo"`
	WarrantyMonths int16     `json:"warrantyMonths"`
	FulfilledBy    string    `json:"fulfilledBy"`
	Score          float64   `json:"score" doc:"Buy-box score (same formula for TechShop and vendors)"`
	BuyBox         bool      `json:"buyBox"`
}

// ProductDetail is the product page.
type ProductDetail struct {
	ID          uuid.UUID                             `json:"id"`
	Slug        string                                `json:"slug"`
	Name        string                                `json:"name"`
	Description *string                               `json:"description,omitempty"`
	Brand       *string                               `json:"brand,omitempty"`
	Category    struct{ Slug, Name string }           `json:"category"`
	Variants    []store.CatalogProductVariant         `json:"variants"`
	Attributes  []store.CatalogListAttributeValuesRow `json:"attributes"`
	Images      []Image                               `json:"images"`
	Offers      []Offer                               `json:"offers" doc:"Buy box first"`
	RatingAvg   *float64                              `json:"ratingAvg,omitempty"`
	RatingCount int32                                 `json:"ratingCount"`
	InStock     bool                                  `json:"inStock"`
	FitsWith    []store.CatalogListCompatibleRow      `json:"accessoriesThatFit"`
}

// Image is a product photo URL.
type Image struct {
	URL      string `json:"url"`
	Alt      string `json:"alt"`
	Position int32  `json:"position"`
}

// buyBox scores offers (search-catalogue.md §4.2) and sorts the winner first.
func (s *Service) buyBox(rows []store.CatalogListOffersRow, shipTo string) []Offer {
	offers := make([]Offer, 0, len(rows))
	minLanded := int64(0)
	for _, r := range rows {
		fee := s.d.Cfg.DeliveryFeeVendorKobo
		if r.SellerType == "first_party" {
			fee = s.d.Cfg.DeliveryFeeTechShopKobo
		}
		if r.SellerState != nil && shipTo != "" && *r.SellerState != shipTo {
			fee = s.d.Cfg.DeliveryFeeInterstateKobo
		}
		o := Offer{ListingID: r.ID, VariantID: r.VariantID, VariantName: r.VariantName, SellerName: r.SellerName, SellerSlug: r.SellerSlug, SellerType: r.SellerType,
			SellerRating: r.SellerRating, Condition: r.Condition, Grade: r.Grade, Price: r.PriceKobo, CompareAt: r.CompareAtKobo, DeliveryFee: fee,
			WarrantyMonths: r.WarrantyMonths, FulfilledBy: r.FulfilledBy}
		landed := o.Price + fee
		if minLanded == 0 || landed < minLanded {
			minLanded = landed
		}
		offers = append(offers, o)
	}
	for i := range offers {
		o := &offers[i]
		rating := 4.0
		if o.SellerRating != nil {
			rating = *o.SellerRating
		}
		o.Score = float64(minLanded)/float64(o.Price+o.DeliveryFee) + 0.20*rating/5 + 0.15*0.9 + 0.10*0.97 + 0.05*float64(o.WarrantyMonths)/12
		if o.FulfilledBy == "techshop" {
			o.Score += 0.05 // fulfilled by TechShop (applies to vendors too); no hidden first-party boost
		}
	}
	sort.SliceStable(offers, func(i, j int) bool { return offers[i].Score > offers[j].Score })
	if len(offers) > 0 {
		offers[0].BuyBox = true
	}
	return offers
}

type searchIn struct {
	Q          string `query:"q" maxLength:"120"`
	Category   string `query:"category"`
	Brand      string `query:"brand" doc:"Comma-separated brand slugs"`
	Condition  string `query:"condition" doc:"Comma-separated: new,uk_used,refurbished,open_box"`
	SellerType string `query:"sellerType" enum:"first_party,business,individual,"`
	MinPrice   int64  `query:"minPrice" doc:"Kobo"`
	MaxPrice   int64  `query:"maxPrice" doc:"Kobo"`
	InStock    bool   `query:"inStock"`
	Sort       string `query:"sort" enum:"relevance,price_asc,price_desc,rating,discount,best_selling,"`
	Cursor     string `query:"cursor"`
	Limit      int    `query:"limit"`
	Explain    bool   `query:"explain" doc:"Include ranking scores"`
}

type searchOut struct {
	Body struct {
		*SearchResult
		Page httpx.Page `json:"page"`
	}
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (s *Service) runSearch(ctx context.Context, in *searchIn, deals bool, forced string) (*searchOut, error) {
	start := time.Now()
	var cur struct{ Offset int }
	if err := s.d.Cursors.Decode(in.Cursor, &cur); err != nil {
		return nil, httpx.BadCursor()
	}
	qNorm := strings.ToLower(strings.TrimSpace(in.Q))
	out := &searchOut{}
	if qNorm != "" && cur.Offset == 0 {
		if r, err := s.d.Q.SearchGetRedirect(ctx, qNorm); err == nil {
			out.Body.SearchResult = &SearchResult{Items: []ProductSummary{}, Redirect: r.Url, Facets: map[string][]Facet{}}
			return out, nil
		}
	}
	p := SearchParams{Q: in.Q, Category: in.Category, Brands: splitCSV(in.Brand), Conditions: splitCSV(in.Condition), SellerType: in.SellerType,
		MinPrice: in.MinPrice, MaxPrice: in.MaxPrice, InStock: in.InStock, Sort: in.Sort, Offset: cur.Offset, Limit: httpx.Limit(in.Limit), Deals: deals}
	if forced != "" {
		p.Sort = forced
	}
	u := understand(in.Q, s.syn.get(ctx))
	res, err := s.search(ctx, p, u, false)
	if err != nil {
		return nil, err
	}
	if res.Total == 0 && u.Text != "" {
		if dym := s.didYouMean(ctx, u.Text); dym != "" && dym != u.Text {
			u2 := u
			u2.Text = dym
			if r2, err := s.search(ctx, p, u2, false); err == nil && r2.Total > 0 {
				r2.DidYouMean = dym
				res = r2
			}
		}
		if res.Total == 0 {
			if r3, err := s.search(ctx, p, u, true); err == nil && r3.Total > 0 {
				res = r3
			}
		}
	}
	if !in.Explain {
		for i := range res.Items {
			res.Items[i].Score = 0
		}
	}
	res.LatencyMs = int(time.Since(start).Milliseconds())
	out.Body.SearchResult = res
	if res.NextOffset > 0 {
		out.Body.Page = httpx.Page{HasMore: true, NextCursor: s.d.Cursors.Encode(struct{ Offset int }{res.NextOffset})}
	}
	if qNorm != "" && cur.Offset == 0 {
		var user *uuid.UUID
		if pr := httpx.PrincipalFrom(ctx); pr != nil {
			user = &pr.UserID
		}
		filters, _ := json.Marshal(map[string]any{"category": in.Category, "brand": in.Brand, "condition": in.Condition})
		lat := int32(res.LatencyMs)
		var corrected *string
		if res.DidYouMean != "" {
			corrected = &res.DidYouMean
		}
		_ = s.d.Q.SearchLogQuery(ctx, store.SearchLogQueryParams{UserID: user, QueryNorm: qNorm, Filters: filters, ResultsCount: int32(res.Total), CorrectedTo: corrected, LatencyMs: &lat, Channel: "retail"})
	}
	return out, nil
}

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	pub := func(op httpx.Op) httpx.Op { op.Tag = "Catalogue"; op.Auth = shopperAuds; op.Optional = true; return op }

	httpx.Register(g, pub(httpx.Op{ID: "listCategories", Method: http.MethodGet, Path: "/api/v1/categories", Summary: "Category tree"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.CatalogCategory }, error) {
			rows, err := s.d.Q.CatalogListCategories(ctx)
			return &struct{ Body []store.CatalogCategory }{rows}, err
		})
	httpx.Register(g, pub(httpx.Op{ID: "listBrands", Method: http.MethodGet, Path: "/api/v1/brands", Summary: "Brands"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.CatalogBrand }, error) {
			rows, err := s.d.Q.CatalogListBrands(ctx)
			return &struct{ Body []store.CatalogBrand }{rows}, err
		})
	httpx.Register(g, pub(httpx.Op{ID: "search", Method: http.MethodGet, Path: "/api/v1/search", Summary: "Search with synonyms, price intents, typo fallback and facets"}),
		func(ctx context.Context, in *searchIn) (*searchOut, error) { return s.runSearch(ctx, in, false, "") })
	httpx.Register(g, pub(httpx.Op{ID: "categoryProducts", Method: http.MethodGet, Path: "/api/v1/categories/{slug}/products", Summary: "Products in a category (same engine as search)"}),
		func(ctx context.Context, in *struct {
			Slug string `path:"slug"`
			searchIn
		}) (*searchOut, error) {
			in.Category = in.Slug
			return s.runSearch(ctx, &in.searchIn, false, "")
		})
	httpx.Register(g, pub(httpx.Op{ID: "deals", Method: http.MethodGet, Path: "/api/v1/deals", Summary: "Discounted products, biggest discount first"}),
		func(ctx context.Context, in *searchIn) (*searchOut, error) {
			return s.runSearch(ctx, in, true, "discount")
		})
	httpx.Register(g, pub(httpx.Op{ID: "bestSellers", Method: http.MethodGet, Path: "/api/v1/best-sellers", Summary: "Best sellers (last 30 days)"}),
		func(ctx context.Context, in *searchIn) (*searchOut, error) {
			return s.runSearch(ctx, in, false, "best_selling")
		})

	httpx.Register(g, pub(httpx.Op{ID: "searchSuggest", Method: http.MethodGet, Path: "/api/v1/search/suggest", Summary: "Autocomplete: queries, categories, products"}),
		func(ctx context.Context, in *struct {
			Q string `query:"q" minLength:"1" maxLength:"60"`
		}) (*struct {
			Body struct {
				Queries    []string                            `json:"queries"`
				Categories []store.CatalogSuggestCategoriesRow `json:"categories"`
				Products   []store.CatalogSuggestProductsRow   `json:"products"`
			}
		}, error) {
			q := strings.ToLower(strings.TrimSpace(in.Q))
			out := &struct {
				Body struct {
					Queries    []string                            `json:"queries"`
					Categories []store.CatalogSuggestCategoriesRow `json:"categories"`
					Products   []store.CatalogSuggestProductsRow   `json:"products"`
				}
			}{}
			prefix := q
			if len(prefix) > 4 {
				prefix = prefix[:4]
			}
			out.Body.Queries, _ = s.d.Q.CatalogSuggestQueries(ctx, prefix)
			out.Body.Categories, _ = s.d.Q.CatalogSuggestCategories(ctx, &q)
			out.Body.Products, _ = s.d.Q.CatalogSuggestProducts(ctx, &q)
			if out.Body.Queries == nil {
				out.Body.Queries = []string{}
			}
			return out, nil
		})

	httpx.Register(g, pub(httpx.Op{ID: "getProduct", Method: http.MethodGet, Path: "/api/v1/products/{slug}", Summary: "Product page with offers and the buy box"}),
		func(ctx context.Context, in *struct {
			Slug   string `path:"slug"`
			ShipTo string `query:"shipTo" pattern:"^([A-Z]{2})?$" doc:"State code for delivery fees, e.g. LA"`
		}) (*struct{ Body *ProductDetail }, error) {
			d, err := s.productDetail(ctx, in.Slug, in.ShipTo)
			return &struct{ Body *ProductDetail }{d}, err
		})

	// Saved items and stock alerts.
	me := func(op httpx.Op) httpx.Op { op.Tag = "Saved"; op.Auth = shopperAuds; return op }
	httpx.Register(g, me(httpx.Op{ID: "listSaved", Method: http.MethodGet, Path: "/api/v1/me/saved", Summary: "Saved products"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.CatalogListSavedRow }, error) {
			rows, err := s.d.Q.CatalogListSaved(ctx, httpx.MustPrincipal(ctx).UserID)
			return &struct{ Body []store.CatalogListSavedRow }{rows}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "toggleSaved", Method: http.MethodPost, Path: "/api/v1/me/saved/{productId}", Summary: "Save or unsave a product"}),
		func(ctx context.Context, in *struct {
			ProductID uuid.UUID `path:"productId"`
		}) (*struct {
			Body struct {
				Saved bool `json:"saved"`
			}
		}, error) {
			p := httpx.MustPrincipal(ctx)
			out := &struct {
				Body struct {
					Saved bool `json:"saved"`
				}
			}{}
			n, err := s.d.Q.CatalogDeleteSaved(ctx, store.CatalogDeleteSavedParams{UserID: p.UserID, ProductID: in.ProductID})
			if err != nil {
				return nil, err
			}
			if n == 0 {
				if err := s.d.Q.CatalogInsertSaved(ctx, store.CatalogInsertSavedParams{UserID: p.UserID, ProductID: in.ProductID}); err != nil {
					return nil, httpx.DB(err, "product")
				}
				out.Body.Saved = true
			}
			return out, nil
		})
	httpx.Register(g, me(httpx.Op{ID: "createStockAlert", Method: http.MethodPost, Path: "/api/v1/me/stock-alerts/{productId}", Status: 204, Summary: "Notify me when back in stock"}),
		func(ctx context.Context, in *struct {
			ProductID uuid.UUID `path:"productId"`
		}) (*struct{}, error) {
			return nil, httpx.DB(s.d.Q.CatalogAddStockAlert(ctx, store.CatalogAddStockAlertParams{UserID: httpx.MustPrincipal(ctx).UserID, ProductID: in.ProductID}), "product")
		})

	s.staffRoutes(g)
}

func (s *Service) productDetail(ctx context.Context, slug, shipTo string) (*ProductDetail, error) {
	p, err := s.d.Q.CatalogGetProductBySlug(ctx, slug)
	if err != nil || p.Status != "active" {
		return nil, httpx.NotFound("product")
	}
	d := &ProductDetail{ID: p.ID, Slug: p.Slug, Name: p.Name, Description: p.Description, Brand: p.BrandName}
	d.Category.Slug, d.Category.Name = p.CategorySlug, p.CategoryName
	if d.Variants, err = s.d.Q.CatalogListVariants(ctx, p.ID); err != nil {
		return nil, err
	}
	if d.Attributes, err = s.d.Q.CatalogListAttributeValues(ctx, p.ID); err != nil {
		return nil, err
	}
	imgs, err := s.d.Q.CatalogListImages(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	d.Images = []Image{}
	for _, im := range imgs {
		d.Images = append(d.Images, Image{URL: s.d.Storage.PublicURL(im.StorageKey), Alt: im.Alt, Position: im.Position})
	}
	offers, err := s.d.Q.CatalogListOffers(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	d.Offers = s.buyBox(offers, shipTo)
	if sum, err := s.d.Q.CatalogSummaryByProduct(ctx, store.CatalogSummaryByProductParams{ProductID: p.ID, Channel: "retail"}); err == nil {
		d.RatingAvg, d.RatingCount, d.InStock = sum.RatingAvg, sum.RatingCount, sum.InStock
	}
	d.FitsWith, _ = s.d.Q.CatalogListCompatible(ctx, &p.ID)
	if d.FitsWith == nil {
		d.FitsWith = []store.CatalogListCompatibleRow{}
	}
	return d, nil
}

// ProductInput creates a product with its first variants and attribute values (staff).
type ProductInput struct {
	CategorySlug string `json:"categorySlug"`
	BrandName    string `json:"brandName,omitempty"`
	Slug         string `json:"slug" pattern:"^[a-z0-9]+(-[a-z0-9]+)*$"`
	Name         string `json:"name" minLength:"2" maxLength:"160"`
	Model        string `json:"model,omitempty"`
	Description  string `json:"description,omitempty" maxLength:"5000"`
	Status       string `json:"status,omitempty" enum:"draft,active,"`
	Variants     []struct {
		SKU          string            `json:"sku" minLength:"2"`
		Name         string            `json:"name" minLength:"1"`
		AxisValues   map[string]string `json:"axisValues,omitempty"`
		GTIN         string            `json:"gtin,omitempty"`
		WeightGrams  *int32            `json:"weightGrams,omitempty"`
		IsSerialised bool              `json:"isSerialised,omitempty"`
	} `json:"variants" minItems:"1"`
	Attributes map[string]any `json:"attributes,omitempty" doc:"key → value (text, number or boolean)"`
}

// CreateProduct is used by staff and by the dev seed.
func (s *Service) CreateProduct(ctx context.Context, tx *uow.Tx, in ProductInput, actor *uuid.UUID) (store.CatalogProduct, []store.CatalogProductVariant, error) {
	cat, err := tx.Q.CatalogGetCategoryBySlug(ctx, in.CategorySlug)
	if err != nil {
		return store.CatalogProduct{}, nil, httpx.Invalid("unknown_category", "Unknown category "+in.CategorySlug)
	}
	var brandID *uuid.UUID
	if in.BrandName != "" {
		b, err := tx.Q.CatalogUpsertBrand(ctx, store.CatalogUpsertBrandParams{Slug: slugify(in.BrandName), Name: in.BrandName})
		if err != nil {
			return store.CatalogProduct{}, nil, err
		}
		brandID = &b.ID
	}
	status := in.Status
	if status == "" {
		status = "draft"
	}
	p, err := tx.Q.CatalogCreateProduct(ctx, store.CatalogCreateProductParams{CategoryID: cat.ID, BrandID: brandID, Slug: in.Slug, Name: in.Name,
		Model: strOrNil(in.Model), Description: strOrNil(in.Description), Status: status, CreatedBy: actor})
	if err != nil {
		return p, nil, httpx.DB(err, "product")
	}
	var variants []store.CatalogProductVariant
	for _, v := range in.Variants {
		axis, _ := json.Marshal(v.AxisValues)
		if v.AxisValues == nil {
			axis = []byte("{}")
		}
		row, err := tx.Q.CatalogCreateVariant(ctx, store.CatalogCreateVariantParams{ProductID: p.ID, Sku: v.SKU, Name: v.Name, AxisValues: axis,
			Gtin: strOrNil(v.GTIN), WeightGrams: v.WeightGrams, IsSerialised: v.IsSerialised})
		if err != nil {
			return p, nil, httpx.DB(err, "variant")
		}
		variants = append(variants, row)
	}
	if len(in.Attributes) > 0 {
		defs, err := tx.Q.CatalogListAttributeDefinitions(ctx, cat.ID)
		if err != nil {
			return p, nil, err
		}
		byKey := map[string]store.CatalogAttributeDefinition{}
		for _, d := range defs {
			byKey[d.Key] = d
		}
		for k, val := range in.Attributes {
			def, ok := byKey[k]
			if !ok {
				return p, nil, httpx.Invalid("unknown_attribute", "Unknown attribute "+k+" for this category")
			}
			prm := store.CatalogSetAttributeValueParams{ProductID: p.ID, AttributeID: def.ID}
			switch v := val.(type) {
			case float64:
				prm.ValueNumber = &v
			case bool:
				prm.ValueBool = &v
			default:
				str := strings.TrimSpace(toString(v))
				prm.ValueText = &str
			}
			if err := tx.Q.CatalogSetAttributeValue(ctx, prm); err != nil {
				return p, nil, err
			}
		}
	}
	tx.Emit("product", p.ID, "product.changed", map[string]any{"productId": p.ID})
	return p, variants, nil
}

func toString(v any) string {
	b, _ := json.Marshal(v)
	return strings.Trim(string(b), `"`)
}

func slugify(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// Slugify is exported for other modules and the seed.
func Slugify(s string) string { return slugify(s) }

func (s *Service) staffRoutes(g *httpx.Guard) {
	st := func(op httpx.Op) httpx.Op { op.Tag = "Staff · Catalogue"; op.Auth = staff; return op }

	httpx.Register(g, st(httpx.Op{ID: "staffCreateCategory", Method: http.MethodPost, Path: "/api/v1/staff/catalog/categories", Perm: "catalog.edit", Status: 201, Summary: "Create or update a category"}),
		func(ctx context.Context, in *struct {
			Body struct {
				ParentSlug     string `json:"parentSlug,omitempty"`
				Slug           string `json:"slug" pattern:"^[a-z0-9]+(-[a-z0-9]+)*$"`
				Name           string `json:"name" minLength:"2"`
				Description    string `json:"description,omitempty"`
				Kind           string `json:"kind,omitempty" enum:"product,vehicle,"`
				RepurchaseDays *int32 `json:"repurchaseDays,omitempty" minimum:"1"`
				Position       int32  `json:"position,omitempty"`
			}
		}) (*struct{ Body store.CatalogCategory }, error) {
			var parent *uuid.UUID
			if in.Body.ParentSlug != "" {
				pc, err := s.d.Q.CatalogGetCategoryBySlug(ctx, in.Body.ParentSlug)
				if err != nil {
					return nil, httpx.Invalid("unknown_category", "Unknown parent category")
				}
				parent = &pc.ID
			}
			kind := in.Body.Kind
			if kind == "" {
				kind = "product"
			}
			var out store.CatalogCategory
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var err error
				out, err = tx.Q.CatalogCreateCategory(ctx, store.CatalogCreateCategoryParams{ParentID: parent, Slug: in.Body.Slug, Name: in.Body.Name,
					Description: strOrNil(in.Body.Description), Kind: kind, RepurchaseDays: in.Body.RepurchaseDays, Position: in.Body.Position})
				if err != nil {
					return httpx.DB(err, "category")
				}
				return tx.Audit("category.saved", "category", out.ID.String(), in.Body)
			})
			return &struct{ Body store.CatalogCategory }{out}, err
		})

	httpx.Register(g, st(httpx.Op{ID: "staffCreateAttribute", Method: http.MethodPost, Path: "/api/v1/staff/catalog/categories/{slug}/attributes", Perm: "catalog.edit", Status: 201, Summary: "Define a spec field for a category"}),
		func(ctx context.Context, in *struct {
			Slug string `path:"slug"`
			Body struct {
				Key           string   `json:"key" pattern:"^[a-z][a-z0-9_]*$"`
				Label         string   `json:"label"`
				DataType      string   `json:"dataType" enum:"text,number,boolean,enum"`
				Unit          string   `json:"unit,omitempty"`
				Options       []string `json:"options,omitempty"`
				IsRequired    bool     `json:"isRequired,omitempty"`
				IsFilterable  bool     `json:"isFilterable,omitempty"`
				IsVariantAxis bool     `json:"isVariantAxis,omitempty"`
				Position      int32    `json:"position,omitempty"`
			}
		}) (*struct {
			Body store.CatalogAttributeDefinition
		}, error) {
			cat, err := s.d.Q.CatalogGetCategoryBySlug(ctx, in.Slug)
			if err != nil {
				return nil, httpx.NotFound("category")
			}
			var opts json.RawMessage
			if in.Body.Options != nil {
				opts, _ = json.Marshal(in.Body.Options)
			}
			row, err := s.d.Q.CatalogUpsertAttributeDefinition(ctx, store.CatalogUpsertAttributeDefinitionParams{CategoryID: cat.ID, Key: in.Body.Key, Label: in.Body.Label,
				DataType: in.Body.DataType, Unit: strOrNil(in.Body.Unit), Options: opts, IsRequired: in.Body.IsRequired, IsFilterable: in.Body.IsFilterable,
				IsVariantAxis: in.Body.IsVariantAxis, Position: in.Body.Position})
			return &struct {
				Body store.CatalogAttributeDefinition
			}{row}, httpx.DB(err, "attribute")
		})

	httpx.Register(g, st(httpx.Op{ID: "staffCreateProduct", Method: http.MethodPost, Path: "/api/v1/staff/catalog/products", Perm: "catalog.edit", Status: 201, Summary: "Create a product with variants and specs"}),
		func(ctx context.Context, in *struct{ Body ProductInput }) (*struct {
			Body struct {
				Product  store.CatalogProduct          `json:"product"`
				Variants []store.CatalogProductVariant `json:"variants"`
			}
		}, error) {
			p := httpx.MustPrincipal(ctx)
			out := &struct {
				Body struct {
					Product  store.CatalogProduct          `json:"product"`
					Variants []store.CatalogProductVariant `json:"variants"`
				}
			}{}
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var err error
				out.Body.Product, out.Body.Variants, err = s.CreateProduct(ctx, tx, in.Body, &p.UserID)
				if err != nil {
					return err
				}
				return tx.Audit("product.created", "product", out.Body.Product.ID.String(), map[string]any{"slug": in.Body.Slug})
			})
			return out, err
		})

	httpx.Register(g, st(httpx.Op{ID: "staffUpdateProduct", Method: http.MethodPatch, Path: "/api/v1/staff/catalog/products/{id}", Perm: "catalog.edit", Summary: "Edit product content or status"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Name        *string `json:"name,omitempty"`
				Description *string `json:"description,omitempty"`
				Status      *string `json:"status,omitempty" enum:"draft,in_review,active,archived"`
			}
		}) (*struct{ Body store.CatalogProduct }, error) {
			var out store.CatalogProduct
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var err error
				out, err = tx.Q.CatalogUpdateProduct(ctx, store.CatalogUpdateProductParams{ID: in.ID, Name: in.Body.Name, Description: in.Body.Description, Status: in.Body.Status})
				if err != nil {
					return httpx.DB(err, "product")
				}
				tx.Emit("product", out.ID, "product.changed", map[string]any{"productId": out.ID})
				return tx.Audit("product.updated", "product", in.ID.String(), in.Body)
			})
			return &struct{ Body store.CatalogProduct }{out}, err
		})

	httpx.Register(g, st(httpx.Op{ID: "staffAddProductImage", Method: http.MethodPost, Path: "/api/v1/staff/catalog/products/{id}/images", Perm: "catalog.edit", Status: 201, Summary: "Attach an uploaded image"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				FileID    uuid.UUID  `json:"fileId"`
				VariantID *uuid.UUID `json:"variantId,omitempty"`
				Alt       string     `json:"alt" minLength:"3"`
				Position  int32      `json:"position,omitempty"`
			}
		}) (*struct{ Body store.CatalogProductImage }, error) {
			row, err := s.d.Q.CatalogAddImage(ctx, store.CatalogAddImageParams{ProductID: in.ID, VariantID: in.Body.VariantID, FileID: in.Body.FileID, Alt: in.Body.Alt, Position: in.Body.Position})
			return &struct{ Body store.CatalogProductImage }{row}, httpx.DB(err, "image")
		})

	httpx.Register(g, st(httpx.Op{ID: "staffCreateListing", Method: http.MethodPost, Path: "/api/v1/staff/catalog/listings", Perm: "catalog.edit", Status: 201, Summary: "Create a TechShop (first-party) offer"}),
		func(ctx context.Context, in *struct{ Body ListingInput }) (*struct{ Body store.CatalogListing }, error) {
			p := httpx.MustPrincipal(ctx)
			var out store.CatalogListing
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				fp, err := tx.Q.SellersGetFirstParty(ctx)
				if err != nil {
					return err
				}
				out, _, err = s.CreateListing(ctx, tx, fp.ID, "first_party", in.Body, &p.UserID)
				if err != nil {
					return err
				}
				return tx.Audit("listing.created", "listing", out.ID.String(), in.Body)
			})
			return &struct{ Body store.CatalogListing }{out}, err
		})

	httpx.Register(g, st(httpx.Op{ID: "staffSetListingPrice", Method: http.MethodPatch, Path: "/api/v1/staff/catalog/listings/{id}/price", Perm: "catalog.edit", Summary: "Change a price (FCCPA guard on “was” prices)"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Price     int64  `json:"price" minimum:"1"`
				CompareAt *int64 `json:"compareAt,omitempty"`
			}
		}) (*struct{ Body store.CatalogListing }, error) {
			p := httpx.MustPrincipal(ctx)
			var out store.CatalogListing
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var err error
				out, err = s.UpdatePrice(ctx, tx, in.ID, in.Body.Price, in.Body.CompareAt, &p.UserID)
				if err != nil {
					return err
				}
				return tx.Audit("listing.price_changed", "listing", in.ID.String(), in.Body)
			})
			return &struct{ Body store.CatalogListing }{out}, err
		})

	httpx.Register(g, st(httpx.Op{ID: "staffModerationQueue", Method: http.MethodGet, Path: "/api/v1/staff/catalog/moderation", Perm: "catalog.moderate", Summary: "Listings waiting for review (highest risk first)"}),
		func(ctx context.Context, _ *struct{}) (*struct {
			Body []store.CatalogModerationQueueRow
		}, error) {
			rows, err := s.d.Q.CatalogModerationQueue(ctx, 100)
			return &struct {
				Body []store.CatalogModerationQueueRow
			}{rows}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffModerateListing", Method: http.MethodPost, Path: "/api/v1/staff/catalog/listings/{id}/decision", Perm: "catalog.moderate", Summary: "Approve, request changes, reject or suspend a listing"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Decision string   `json:"decision" enum:"approved,changes_requested,rejected,suspended"`
				Reasons  []string `json:"reasons,omitempty"`
				Note     string   `json:"note,omitempty" maxLength:"1000"`
			}
		}) (*struct{ Body *store.CatalogListing }, error) {
			if in.Body.Reasons == nil {
				in.Body.Reasons = []string{}
			}
			l, err := s.Moderate(ctx, in.ID, in.Body.Decision, in.Body.Reasons, in.Body.Note)
			return &struct{ Body *store.CatalogListing }{l}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffListSuggestions", Method: http.MethodGet, Path: "/api/v1/staff/catalog/suggestions", Perm: "catalog.moderate", Summary: "Seller-proposed products and edits"}),
		func(ctx context.Context, _ *struct{}) (*struct {
			Body []store.CatalogProductSuggestion
		}, error) {
			rows, err := s.d.Q.CatalogListSuggestions(ctx, 100)
			return &struct {
				Body []store.CatalogProductSuggestion
			}{rows}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffDecideSuggestion", Method: http.MethodPost, Path: "/api/v1/staff/catalog/suggestions/{id}/decision", Perm: "catalog.moderate", Summary: "Accept or reject a suggestion"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Decision string `json:"decision" enum:"accepted,rejected"`
			}
		}) (*struct {
			Body store.CatalogProductSuggestion
		}, error) {
			p := httpx.MustPrincipal(ctx)
			row, err := s.d.Q.CatalogDecideSuggestion(ctx, store.CatalogDecideSuggestionParams{ID: in.ID, Status: in.Body.Decision, ReviewedBy: &p.UserID})
			return &struct {
				Body store.CatalogProductSuggestion
			}{row}, httpx.DB(err, "pending suggestion")
		})
	httpx.Register(g, st(httpx.Op{ID: "staffReindex", Method: http.MethodPost, Path: "/api/v1/staff/catalog/reindex", Perm: "catalog.edit", Summary: "Rebuild the search read model now"}),
		func(ctx context.Context, _ *struct{}) (*struct {
			Body struct {
				Products int `json:"products"`
			}
		}, error) {
			n, err := s.RebuildAll(ctx)
			out := &struct {
				Body struct {
					Products int `json:"products"`
				}
			}{}
			out.Body.Products = n
			return out, err
		})

	// Search merchandising (search-catalogue.md §5).
	sm := func(op httpx.Op) httpx.Op {
		op.Tag = "Staff · Search"
		op.Auth = staff
		op.Perm = "search.manage"
		return op
	}
	httpx.Register(g, sm(httpx.Op{ID: "staffListSynonyms", Method: http.MethodGet, Path: "/api/v1/staff/search/synonyms", Summary: "Active synonyms"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.SearchSearchSynonym }, error) {
			rows, err := s.d.Q.SearchListSynonyms(ctx)
			return &struct{ Body []store.SearchSearchSynonym }{rows}, err
		})
	httpx.Register(g, sm(httpx.Op{ID: "staffCreateSynonym", Method: http.MethodPost, Path: "/api/v1/staff/search/synonyms", Status: 201, Summary: "Add a synonym (live immediately)"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Terms  []string `json:"terms" minItems:"1"`
				Target string   `json:"target" minLength:"1"`
				Kind   string   `json:"kind" enum:"one_way,two_way"`
			}
		}) (*struct{ Body store.SearchSearchSynonym }, error) {
			p := httpx.MustPrincipal(ctx)
			var out store.SearchSearchSynonym
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				for i := range in.Body.Terms {
					in.Body.Terms[i] = strings.ToLower(strings.TrimSpace(in.Body.Terms[i]))
				}
				var err error
				out, err = tx.Q.SearchCreateSynonym(ctx, store.SearchCreateSynonymParams{Terms: in.Body.Terms, Target: strings.ToLower(in.Body.Target), Kind: in.Body.Kind, CreatedBy: &p.UserID})
				if err != nil {
					return err
				}
				tx.Notify(realtime.ChannelSearch, map[string]string{"changed": "synonyms"})
				return tx.Audit("search.synonym_created", "synonym", out.ID.String(), in.Body)
			})
			return &struct{ Body store.SearchSearchSynonym }{out}, err
		})
	httpx.Register(g, sm(httpx.Op{ID: "staffDeleteSynonym", Method: http.MethodDelete, Path: "/api/v1/staff/search/synonyms/{id}", Status: 204, Summary: "Deactivate a synonym"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{}, error) {
			return nil, s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				if _, err := tx.Q.SearchDeactivateSynonym(ctx, in.ID); err != nil {
					return err
				}
				tx.Notify(realtime.ChannelSearch, map[string]string{"changed": "synonyms"})
				return tx.Audit("search.synonym_removed", "synonym", in.ID.String(), nil)
			})
		})
	httpx.Register(g, sm(httpx.Op{ID: "staffSetRedirect", Method: http.MethodPut, Path: "/api/v1/staff/search/redirects/{query}", Summary: "Send a query straight to a page"}),
		func(ctx context.Context, in *struct {
			Query string `path:"query"`
			Body  struct {
				URL       string     `json:"url" pattern:"^/"`
				ExpiresAt *time.Time `json:"expiresAt,omitempty"`
			}
		}) (*struct{ Body store.SearchSearchRedirect }, error) {
			p := httpx.MustPrincipal(ctx)
			row, err := s.d.Q.SearchUpsertRedirect(ctx, store.SearchUpsertRedirectParams{QueryNorm: strings.ToLower(in.Query), Url: in.Body.URL, CreatedBy: &p.UserID, ExpiresAt: in.Body.ExpiresAt})
			return &struct{ Body store.SearchSearchRedirect }{row}, err
		})
	httpx.Register(g, sm(httpx.Op{ID: "staffCreatePin", Method: http.MethodPost, Path: "/api/v1/staff/search/pins", Status: 201, Summary: "Pin or bury a product for a query (time-boxed)"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Query     string    `json:"query"`
				ProductID uuid.UUID `json:"productId"`
				Action    string    `json:"action" enum:"pin,bury"`
				Position  *int16    `json:"position,omitempty"`
				EndsAt    time.Time `json:"endsAt"`
			}
		}) (*struct{ Body store.SearchSearchPin }, error) {
			p := httpx.MustPrincipal(ctx)
			row, err := s.d.Q.SearchCreatePin(ctx, store.SearchCreatePinParams{QueryNorm: strings.ToLower(in.Body.Query), ProductID: in.Body.ProductID, Action: in.Body.Action,
				Position: in.Body.Position, StartsAt: time.Now(), EndsAt: in.Body.EndsAt, CreatedBy: p.UserID})
			return &struct{ Body store.SearchSearchPin }{row}, httpx.DB(err, "pin")
		})
	httpx.Register(g, sm(httpx.Op{ID: "staffSearchInsights", Method: http.MethodGet, Path: "/api/v1/staff/search/insights", Summary: "Top and zero-result queries (7 days)"}),
		func(ctx context.Context, _ *struct{}) (*struct {
			Body struct {
				Stats       store.SearchStatsRow         `json:"stats"`
				ZeroResults []store.SearchZeroResultsRow `json:"zeroResults"`
				Top         []store.SearchTopQueriesRow  `json:"top"`
				Gaps        []store.SearchCatalogueGap   `json:"gaps"`
			}
		}, error) {
			out := &struct {
				Body struct {
					Stats       store.SearchStatsRow         `json:"stats"`
					ZeroResults []store.SearchZeroResultsRow `json:"zeroResults"`
					Top         []store.SearchTopQueriesRow  `json:"top"`
					Gaps        []store.SearchCatalogueGap   `json:"gaps"`
				}
			}{}
			var err error
			if out.Body.Stats, err = s.d.Q.SearchStats(ctx); err != nil {
				return nil, err
			}
			out.Body.ZeroResults, _ = s.d.Q.SearchZeroResults(ctx, 50)
			out.Body.Top, _ = s.d.Q.SearchTopQueries(ctx, 50)
			out.Body.Gaps, _ = s.d.Q.SearchListGaps(ctx)
			return out, nil
		})
	httpx.Register(g, sm(httpx.Op{ID: "staffMarkGap", Method: http.MethodPut, Path: "/api/v1/staff/search/gaps/{query}", Summary: "Record a catalogue gap for purchasing"}),
		func(ctx context.Context, in *struct {
			Query string `path:"query"`
			Body  struct {
				Status   string `json:"status" enum:"new,sourcing,wont_stock,resolved"`
				Searches int32  `json:"searches30d,omitempty"`
				Note     string `json:"note,omitempty"`
			}
		}) (*struct{ Body store.SearchCatalogueGap }, error) {
			p := httpx.MustPrincipal(ctx)
			row, err := s.d.Q.SearchUpsertGap(ctx, store.SearchUpsertGapParams{QueryNorm: strings.ToLower(in.Query), Searches30d: in.Body.Searches, Status: in.Body.Status, OwnerID: &p.UserID, Note: strOrNil(in.Body.Note)})
			return &struct{ Body store.SearchCatalogueGap }{row}, err
		})
}
