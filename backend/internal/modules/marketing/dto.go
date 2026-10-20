package marketing

import (
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/store"
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
