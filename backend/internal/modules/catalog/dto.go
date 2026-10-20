package catalog

import (
	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/store"
)

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
