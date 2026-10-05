package seed

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/catalog"
	"github.com/dprince-03/techshop/backend/internal/inventory"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

//go:embed fixtures.json
var fixturesJSON []byte

type fxMoney struct {
	Amount int64 `json:"amount"`
}
type fixtures struct {
	Categories []struct {
		Slug, Name, Description string
		Subcategories           []struct{ Slug, Name string } `json:"subcategories"`
	} `json:"categories"`
	Products []struct {
		ID, Name, Brand, Category, Subcategory, KeySpec, Condition, Stock, Slug string
		Price                                                                   fxMoney  `json:"price"`
		CompareAt                                                               *fxMoney `json:"compareAtPrice"`
		Rating                                                                  struct {
			Average float64 `json:"average"`
			Count   int32   `json:"count"`
		} `json:"rating"`
		Seller struct{ ID, Name, Type string } `json:"seller"`
	} `json:"products"`
	Wholesale []struct {
		Slug  string `json:"slug"`
		Tiers []struct {
			MinQuantity int32   `json:"minQuantity"`
			UnitPrice   fxMoney `json:"unitPrice"`
		} `json:"tiers"`
	} `json:"wholesaleProducts"`
}

// Hooks lets later milestones add their own dev data (zones, riders, cars…).
var Hooks []func(ctx context.Context, d *kit.Deps, out *Output, ids map[string]uuid.UUID) error

// Catalog and Inventory are set by cmd/seed from the composition root.
var Catalog *catalog.Service
var Inventory *inventory.Service

var vendorSellers = map[string]struct {
	slug, state, city string
	rating            float64
}{
	"s-ikeja":    {"ikeja-gadget-hub", "LA", "Ikeja", 4.6},
	"s-abuja":    {"wuse-tech-store", "FC", "Wuse II", 4.4},
	"s-kano":     {"sabon-gari-mobile", "KN", "Kano", 4.3},
	"s-autolane": {"lekki-auto-lane", "LA", "Lekki", 4.5},
}

func seedCommerce(ctx context.Context, d *kit.Deps, out *Output,
	user func(email, phone, first, last string) (store.IdentityUser, error),
	staff func(u store.IdentityUser, no, dept, title string, roles ...string) (string, error), password string) error {
	var fx fixtures
	if err := json.Unmarshal(fixturesJSON, &fx); err != nil {
		return fmt.Errorf("fixtures: %w", err)
	}
	q := d.Q
	ids := map[string]uuid.UUID{}

	// Vendor shops, each owned by a dev seller account; the first is the seller test account.
	sellerUser, err := user("seller@dev.techshop.ng", "+2348031234568", "Adeola", "Bello")
	if err != nil {
		return err
	}
	out.Accounts = append(out.Accounts, Account{Role: "seller_owner", UserID: sellerUser.ID, Email: "seller@dev.techshop.ng", Phone: "+2348031234568", Password: password, Audience: "seller"})
	fp, err := q.SellersGetFirstParty(ctx)
	if err != nil {
		return err
	}
	ids["s-techshop"] = fp.ID
	for fxID, v := range vendorSellers {
		owner := sellerUser
		if fxID != "s-ikeja" {
			owner, err = user(strings.TrimPrefix(fxID, "s-")+"@dev.techshop.ng", "", title(strings.TrimPrefix(fxID, "s-")), "Owner")
			if err != nil {
				return err
			}
		}
		name := map[string]string{"s-ikeja": "Ikeja Gadget Hub", "s-abuja": "Wuse Tech Store", "s-kano": "Sabon Gari Mobile", "s-autolane": "Lekki Auto Lane"}[fxID]
		st, city, rating := v.state, v.city, v.rating
		sel, err := q.SellersCreate(ctx, store.SellersCreateParams{Type: "business", DisplayName: name, Slug: v.slug, OwnerUserID: &owner.ID, Status: "active", StateCode: &st, City: &city, RatingAvg: &rating, RatingCount: 120})
		if err != nil {
			return fmt.Errorf("seller %s: %w", name, err)
		}
		if err := q.SellersAddMember(ctx, store.SellersAddMemberParams{SellerID: sel.ID, UserID: owner.ID, Role: "owner"}); err != nil {
			return err
		}
		ids[fxID] = sel.ID
	}
	out.Extra["sellerIkejaId"] = ids["s-ikeja"].String()

	// Warehouses.
	ikj, err := q.InventoryCreateWarehouse(ctx, store.InventoryCreateWarehouseParams{Code: "IKJ", Name: "Ikeja warehouse", Kind: "warehouse", Address: "14 Oba Akran Ave", City: "Ikeja", StateCode: "LA"})
	if err != nil {
		return err
	}
	if _, err := q.InventoryCreateWarehouse(ctx, store.InventoryCreateWarehouseParams{Code: "WUSE", Name: "Wuse hub", Kind: "hub", Address: "3 Adetokunbo Ademola Cres", City: "Abuja", StateCode: "FC"}); err != nil {
		return err
	}
	if _, err := q.InventoryCreateWarehouse(ctx, store.InventoryCreateWarehouseParams{Code: "VI-STORE", Name: "Victoria Island store", Kind: "store", Address: "1 Adeola Odeku St", City: "Victoria Island", StateCode: "LA"}); err != nil {
		return err
	}
	ids["wh-ikj"] = ikj.ID
	out.Extra["warehouseIkejaId"] = ikj.ID.String()

	// Categories (+ filterable attributes for phones and laptops).
	for i, c := range fx.Categories {
		kind := "product"
		if c.Slug == "cars" {
			kind = "vehicle"
		}
		var rep *int32
		if c.Slug == "phones" || c.Slug == "laptops" {
			v := int32(365)
			rep = &v
		}
		parent, err := q.CatalogCreateCategory(ctx, store.CatalogCreateCategoryParams{Slug: c.Slug, Name: c.Name, Description: &c.Description, Kind: kind, RepurchaseDays: rep, Position: int32(i)})
		if err != nil {
			return err
		}
		for j, sc := range c.Subcategories {
			if _, err := q.CatalogCreateCategory(ctx, store.CatalogCreateCategoryParams{ParentID: &parent.ID, Slug: c.Slug + "-" + sc.Slug, Name: sc.Name, Kind: kind, Position: int32(j)}); err != nil {
				return err
			}
		}
		if c.Slug == "phones" || c.Slug == "laptops" {
			for k, a := range []struct{ key, label, unit string }{{"ram_gb", "RAM", "GB"}, {"storage_gb", "Storage", "GB"}, {"key_spec", "Key specs", ""}} {
				dt := "number"
				if a.key == "key_spec" {
					dt = "text"
				}
				unit := a.unit
				if _, err := q.CatalogUpsertAttributeDefinition(ctx, store.CatalogUpsertAttributeDefinitionParams{CategoryID: parent.ID, Key: a.key, Label: a.label, DataType: dt,
					Unit: &unit, IsFilterable: a.key != "key_spec", Position: int32(k)}); err != nil {
					return err
				}
			}
		}
	}

	// Products, variants, listings and stock (same names and prices as the web sample data).
	tiers := map[string][]struct {
		MinQuantity int32   `json:"minQuantity"`
		UnitPrice   fxMoney `json:"unitPrice"`
	}{}
	for _, w := range fx.Wholesale {
		tiers[w.Slug] = w.Tiers
	}
	ramRe, storRe := regexpMust(`(\d+)GB RAM`), regexpMust(`(\d+)\s?(GB|TB)\b`)
	for _, p := range fx.Products {
		if p.Category == "cars" {
			continue
		}
		var lid uuid.UUID
		err := d.Runner.Run(ctx, func(tx *uow.Tx) error {
			existing, err := tx.Q.CatalogGetProductBySlugForUpdate(ctx, p.Slug)
			var variantID uuid.UUID
			if err == nil {
				vs, _ := tx.Q.CatalogListVariants(ctx, existing.ID)
				if len(vs) > 0 {
					variantID = vs[0].ID
				}
				ids[p.ID] = existing.ID
				ids[p.ID+":variant"] = variantID
				return nil
			}
			attrs := map[string]any{}
			if p.Category == "phones" || p.Category == "laptops" {
				attrs["key_spec"] = p.KeySpec
				if m := ramRe.FindStringSubmatch(p.KeySpec); m != nil {
					attrs["ram_gb"] = atof(m[1])
				}
				for _, m := range storRe.FindAllStringSubmatch(p.KeySpec, -1) {
					if !strings.Contains(p.KeySpec, m[1]+"GB RAM") {
						v := atof(m[1])
						if m[2] == "TB" {
							v *= 1024
						}
						attrs["storage_gb"] = v
					}
				}
			}
			cat := p.Category
			serial := p.Category == "phones" || p.Category == "laptops" || p.Category == "gaming"
			in := catalog.ProductInput{CategorySlug: cat, BrandName: p.Brand, Slug: p.Slug, Name: p.Name, Description: p.KeySpec, Status: "active", Attributes: attrs}
			in.Variants = append(in.Variants, struct {
				SKU          string            `json:"sku" minLength:"2"`
				Name         string            `json:"name" minLength:"1"`
				AxisValues   map[string]string `json:"axisValues,omitempty"`
				GTIN         string            `json:"gtin,omitempty"`
				WeightGrams  *int32            `json:"weightGrams,omitempty"`
				IsSerialised bool              `json:"isSerialised,omitempty"`
			}{SKU: strings.ToUpper(p.ID) + "-" + strings.ToUpper(catalog.Slugify(p.KeySpec))[:min(12, len(catalog.Slugify(p.KeySpec)))], Name: p.KeySpec, IsSerialised: serial, WeightGrams: ptr(int32(800))})
			prod, variants, err := Catalog.CreateProduct(ctx, tx, in, nil)
			if err != nil {
				return fmt.Errorf("product %s: %w", p.Slug, err)
			}
			ids[p.ID] = prod.ID
			variantID = variants[0].ID
			ids[p.ID+":variant"] = variantID
			sellerID := ids[p.Seller.ID]
			sellerType := "business"
			fulfilled := "seller"
			stock := ptr(int32(25))
			if p.Seller.Type == "techshop" {
				sellerType, fulfilled, stock = "first_party", "techshop", nil
			}
			cond := strings.ReplaceAll(p.Condition, "-", "_")
			var notes string
			var battery *int16
			if cond != "new" {
				notes = p.KeySpec
				battery = ptr(int16(89))
			}
			if p.Stock == "low-stock" && stock != nil {
				stock = ptr(int32(2))
			}
			li := catalog.ListingInput{VariantID: variantID, Condition: cond, Price: p.Price.Amount, FulfilledBy: fulfilled, SellerStock: stock, WarrantyMonths: 12, ConditionNotes: notes, BatteryHealthPct: battery}
			// Seeded vendor listings go live directly (dev data; real ones pass moderation).
			l, _, err := Catalog.CreateListing(ctx, tx, sellerID, "first_party", li, nil)
			if err != nil {
				return fmt.Errorf("listing %s: %w", p.Slug, err)
			}
			lid = l.ID
			_ = sellerType
			if p.CompareAt != nil {
				// Real price history: it sold at the higher price first, so the "was" price is genuine.
				if _, err := tx.Q.CatalogUpdateListingPrice(ctx, store.CatalogUpdateListingPriceParams{ID: l.ID, PriceKobo: p.CompareAt.Amount}); err != nil {
					return err
				}
				if err := tx.Q.CatalogInsertPriceHistory(ctx, store.CatalogInsertPriceHistoryParams{ListingID: l.ID, PriceKobo: p.CompareAt.Amount}); err != nil {
					return err
				}
				if _, err := Catalog.UpdatePrice(ctx, tx, l.ID, p.Price.Amount, &p.CompareAt.Amount, nil); err != nil {
					return err
				}
			}
			for _, t := range tiers[p.Slug] {
				if err := tx.Q.CatalogInsertTier(ctx, store.CatalogInsertTierParams{ListingID: l.ID, MinQuantity: t.MinQuantity, UnitPriceKobo: t.UnitPrice.Amount}); err != nil {
					return err
				}
			}
			if fulfilled == "techshop" {
				qty := int32(30)
				if p.Stock == "low-stock" {
					qty = 2
				}
				if _, err := Inventory.Apply(ctx, tx, inventory.Move{WarehouseID: ikj.ID, VariantID: variantID, Condition: cond, OwnerSellerID: fp.ID, Delta: qty, Reason: "purchase_receipt", RefType: "seed"}); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		if lid != uuid.Nil {
			ids[p.ID+":listing"] = lid
		}
	}
	// Ids for tests and tools: every fixture product with its first listing and variant.
	for _, p := range fx.Products {
		k := p.ID
		out.Extra["product_"+k] = ids[k].String()
		if id, ok := ids[k+":listing"]; ok {
			out.Extra["listing_"+k] = id.String()
		}
		if id, ok := ids[k+":variant"]; ok {
			out.Extra["variant_"+k] = id.String()
		}
	}

	// Search synonyms (search-catalogue.md §4.1) and a redirect.
	for _, sy := range [][3]string{{"tokunbo", "uk_used", "one_way"}, {"iphone", "orbit one", "one_way"}, {"ps5", "play 5", "one_way"}, {"playstation", "play 5", "one_way"},
		{"airpods", "buds", "one_way"}, {"earpiece", "buds", "one_way"}, {"powerbank", "power bank", "one_way"}, {"pc", "laptop", "one_way"}, {"cctv", "camera", "one_way"}} {
		_, _ = q.SearchCreateSynonym(ctx, store.SearchCreateSynonymParams{Terms: []string{sy[0]}, Target: sy[1], Kind: sy[2]})
	}
	_, _ = q.SearchUpsertRedirect(ctx, store.SearchUpsertRedirectParams{QueryNorm: "cars", Url: "/c/cars"})

	// CMS: one published page and help article.
	pg, err := q.ContentUpsertPage(ctx, store.ContentUpsertPageParams{Site: "market", Slug: "delivery", Title: "Delivery", Body: []byte(`[{"type":"paragraph","text":"Lagos: next day. Other states: 2–4 working days."}]`)})
	if err == nil {
		_, _ = q.ContentSetPageStatus(ctx, store.ContentSetPageStatusParams{ID: pg.ID, Status: "published"})
	}
	_, _ = q.ContentUpsertHelp(ctx, store.ContentUpsertHelpParams{Site: "market", Topic: "Orders", Slug: "track-order", Title: "How do I track my order?", Body: []byte(`[{"type":"paragraph","text":"Use Track order with your order number."}]`), Status: "published"})

	if _, err := Catalog.RebuildAll(ctx); err != nil {
		return fmt.Errorf("rebuild read model: %w", err)
	}
	for _, h := range Hooks {
		if err := h(ctx, d, out, ids); err != nil {
			return err
		}
	}
	return nil
}

func ptr[T any](v T) *T { return &v }

func atof(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

// title capitalises the first letter (dev seed names like "abuja" → "Abuja").
func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func regexpMust(p string) *regexp.Regexp { return regexp.MustCompile(p) }
