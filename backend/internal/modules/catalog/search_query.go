package catalog

// This file holds the ONE sanctioned dynamic query in the codebase (docs/backend.md §4.2):
// catalogue search with optional filters and facets over the read model. Every value is passed
// as a bind parameter; only fixed fragments from this file are concatenated.

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Understood is what query understanding extracted from the raw text.
type Understood struct {
	Text      string   `json:"text" doc:"Search terms after synonyms"`
	Condition string   `json:"condition,omitempty"`
	MinPrice  int64    `json:"minPrice,omitempty" doc:"Kobo"`
	MaxPrice  int64    `json:"maxPrice,omitempty" doc:"Kobo"`
	Notes     []string `json:"notes,omitempty" doc:"Human-readable rewrites, e.g. tokunbo → UK-used"`
}

var (
	reUnder   = regexp.MustCompile(`\b(?:under|below|less than|max)\s+₦?(\d+(?:\.\d+)?)\s*(k|m)?\b`)
	reBetween = regexp.MustCompile(`\b(?:between\s+)?₦?(\d+(?:\.\d+)?)\s*(k|m)?\s*(?:-|to|and)\s*₦?(\d+(?:\.\d+)?)\s*(k|m)?\b`)
	reGB      = regexp.MustCompile(`(\d+)\s?gb\b`)
	stopWords = map[string]bool{"for": true, "the": true, "a": true, "with": true, "and": true, "of": true}
)

func moneyOf(num, unit string) int64 {
	f, _ := strconv.ParseFloat(num, 64)
	switch unit {
	case "k":
		f *= 1e3
	case "m":
		f *= 1e6
	}
	return int64(math.Round(f * 100))
}

// synonymCache holds active synonyms (reloaded on NOTIFY 'search' and every 5 minutes).
type synonymCache struct {
	mu     sync.RWMutex
	byTerm map[string]string
	loaded time.Time
	reload func(ctx context.Context) (map[string]string, error)
}

func (c *synonymCache) get(ctx context.Context) map[string]string {
	c.mu.RLock()
	m, fresh := c.byTerm, time.Since(c.loaded) < 5*time.Minute
	c.mu.RUnlock()
	if m != nil && fresh {
		return m
	}
	nm, err := c.reload(ctx)
	if err != nil {
		return m
	}
	c.mu.Lock()
	c.byTerm, c.loaded = nm, time.Now()
	c.mu.Unlock()
	return nm
}

func (c *synonymCache) invalidate(string) {
	c.mu.Lock()
	c.loaded = time.Time{}
	c.mu.Unlock()
}

// understand normalises a query: price intents, condition words, storage, synonyms.
func understand(raw string, syn map[string]string) Understood {
	q := " " + strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(raw, ",", ""), "₦", "")) + " "
	var u Understood
	if m := reBetween.FindStringSubmatch(q); m != nil {
		u.MinPrice, u.MaxPrice = moneyOf(m[1], m[2]), moneyOf(m[3], m[4])
		q = strings.Replace(q, m[0], " ", 1)
		u.Notes = append(u.Notes, fmt.Sprintf("price ₦%d–₦%d", u.MinPrice/100, u.MaxPrice/100))
	} else if m := reUnder.FindStringSubmatch(q); m != nil {
		u.MaxPrice = moneyOf(m[1], m[2])
		q = strings.Replace(q, m[0], " ", 1)
		u.Notes = append(u.Notes, fmt.Sprintf("price ≤ ₦%d", u.MaxPrice/100))
	}
	q = reGB.ReplaceAllString(q, "${1}gb")
	var terms []string
	for _, t := range strings.Fields(q) {
		if to, ok := syn[t]; ok {
			u.Notes = append(u.Notes, fmt.Sprintf("“%s” → “%s”", t, to))
			t = to
		}
		for _, w := range strings.Fields(t) {
			switch {
			case w == "uk_used" || w == "uk-used" || w == "tokunbo":
				u.Condition = "uk_used"
			case w == "refurbished":
				u.Condition = "refurbished"
			case w == "new" || w == "brand-new":
				u.Condition = "new"
			case stopWords[w]:
			default:
				terms = append(terms, w)
			}
		}
	}
	// "uk used" as two words
	joined := " " + strings.Join(terms, " ") + " "
	if strings.Contains(joined, " uk used ") {
		u.Condition = "uk_used"
		joined = strings.Replace(joined, " uk used ", " ", 1)
	}
	u.Text = strings.TrimSpace(joined)
	return u
}

// SearchParams are the filters of /search and category pages.
type SearchParams struct {
	Q          string
	Category   string
	Brands     []string
	Conditions []string
	SellerType string
	MinPrice   int64
	MaxPrice   int64
	InStock    bool
	Sort       string
	Offset     int
	Limit      int
	Deals      bool
}

// ProductSummary is a product card.
type ProductSummary struct {
	ID          uuid.UUID `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	BrandName   *string   `json:"brandName,omitempty"`
	Price       *int64    `json:"price,omitempty" doc:"Lowest offer price in kobo"`
	CompareAt   *int64    `json:"compareAt,omitempty" doc:"Real previous price in kobo (FCCPA-checked)"`
	PromoPrice  *int64    `json:"promoPrice,omitempty"`
	DiscountPct *int16    `json:"discountPct,omitempty"`
	OfferCount  int32     `json:"offerCount"`
	Conditions  []string  `json:"conditions"`
	SellerTypes []string  `json:"sellerTypes"`
	InStock     bool      `json:"inStock"`
	StockBand   *string   `json:"stockBand,omitempty" enum:"out,low,in"`
	RatingAvg   *float64  `json:"ratingAvg,omitempty"`
	RatingCount int32     `json:"ratingCount"`
	Score       float64   `json:"score,omitempty" doc:"Ranking score (explain mode)"`
}

// Facet is one filter value with its count.
type Facet struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

// SearchResult is a page of products plus facets.
type SearchResult struct {
	Items      []ProductSummary   `json:"items"`
	Total      int                `json:"total"`
	Facets     map[string][]Facet `json:"facets"`
	Understood Understood         `json:"understood"`
	DidYouMean string             `json:"didYouMean,omitempty"`
	Redirect   string             `json:"redirect,omitempty"`
	NextOffset int                `json:"-"`
	LatencyMs  int                `json:"latencyMs"`
}

// rankExpr: text relevance × (1 + sales + Bayesian rating) × availability (search-catalogue.md §4.1).
const rankExpr = `(case when $1 = '' then 1 else ts_rank_cd(s.search_vector, websearch_to_tsquery('simple', $1)) + 0.0001 end)
  * (1 + 0.25 * ln(1 + s.sales_30d) / ln(1501) + 0.15 * ((coalesce(s.rating_avg, 4.2) * s.rating_count + 4.2 * 20) / (s.rating_count + 20) - 3) / 2)
  * (case s.stock_band when 'in' then 1.0 when 'low' then 0.9 else 0.2 end)`

func (s *Service) search(ctx context.Context, p SearchParams, u Understood, fuzzy bool) (*SearchResult, error) {
	args := []any{u.Text}
	where := []string{"s.channel = 'retail'", "$1::text is not null"}
	add := func(cond string, v any) {
		args = append(args, v)
		where = append(where, strings.ReplaceAll(cond, "?", fmt.Sprintf("$%d", len(args))))
	}
	if u.Text != "" {
		if fuzzy {
			add("(similarity(s.title, ?) > 0.3 or s.title ilike '%' || ? || '%')", u.Text)
		} else {
			where = append(where, "s.search_vector @@ websearch_to_tsquery('simple', $1)")
		}
	}
	if p.Category != "" {
		add("exists (select 1 from catalog.categories c where c.slug = ? and c.id = any(s.category_ids))", p.Category)
	}
	if len(p.Brands) > 0 {
		add("exists (select 1 from catalog.brands b where b.id = s.brand_id and b.slug = any(?))", p.Brands)
	}
	conds := p.Conditions
	if u.Condition != "" && len(conds) == 0 {
		conds = []string{u.Condition}
	}
	if len(conds) > 0 {
		add("s.conditions && ?::text[]", conds)
	}
	if p.SellerType != "" {
		add("? = any(s.seller_types)", p.SellerType)
	}
	minP, maxP := p.MinPrice, p.MaxPrice
	if minP == 0 {
		minP = u.MinPrice
	}
	if maxP == 0 {
		maxP = u.MaxPrice
	}
	if minP > 0 {
		add("s.min_price_kobo >= ?", minP)
	}
	if maxP > 0 {
		add("s.min_price_kobo <= ?", maxP)
	}
	if p.InStock {
		where = append(where, "s.in_stock")
	}
	if p.Deals {
		where = append(where, "(s.discount_pct > 0 or s.promo_price_kobo is not null)")
	}
	w := strings.Join(where, " and ")
	order := "score desc, s.sales_30d desc"
	switch p.Sort {
	case "price_asc":
		order = "s.min_price_kobo asc nulls last"
	case "price_desc":
		order = "s.min_price_kobo desc nulls last"
	case "rating":
		order = "s.rating_avg desc nulls last, s.rating_count desc"
	case "discount":
		order = "s.discount_pct desc nulls last"
	case "best_selling":
		order = "s.sales_30d desc"
	}
	limit := p.Limit
	args = append(args, limit+1, p.Offset)
	sql := fmt.Sprintf(`select s.product_id, s.slug, s.title, b.name, s.min_price_kobo, s.compare_at_kobo, s.promo_price_kobo, s.discount_pct,
	  s.offer_count, s.conditions, s.seller_types, s.in_stock, s.stock_band, s.rating_avg::float8, s.rating_count, %s as score
	  from catalog.product_offer_summary s left join catalog.brands b on b.id = s.brand_id
	  where %s order by %s limit $%d offset $%d`, rankExpr, w, order, len(args)-1, len(args))
	rows, err := s.d.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer rows.Close()
	res := &SearchResult{Items: []ProductSummary{}, Facets: map[string][]Facet{}, Understood: u}
	for rows.Next() {
		var it ProductSummary
		if err := rows.Scan(&it.ID, &it.Slug, &it.Name, &it.BrandName, &it.Price, &it.CompareAt, &it.PromoPrice, &it.DiscountPct, &it.OfferCount,
			&it.Conditions, &it.SellerTypes, &it.InStock, &it.StockBand, &it.RatingAvg, &it.RatingCount, &it.Score); err != nil {
			return nil, err
		}
		res.Items = append(res.Items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(res.Items) > limit {
		res.Items = res.Items[:limit]
		res.NextOffset = p.Offset + limit
	}
	// Facets and total over the same filters (without paging).
	fargs := args[:len(args)-2]
	facetSQL := fmt.Sprintf(`with m as (select s.* from catalog.product_offer_summary s where %s)
	  select 'total', '', '', count(*)::int from m
	  union all select 'brand', b.slug, b.name, count(*)::int from m join catalog.brands b on b.id = m.brand_id group by b.slug, b.name
	  union all select 'condition', c, c, count(*)::int from m, unnest(m.conditions) c group by c
	  union all select 'sellerType', t, t, count(*)::int from m, unnest(m.seller_types) t group by t
	  union all select 'price', bucket, bucket, count(*)::int from (select case when min_price_kobo < 10000000 then 'under_100k'
	      when min_price_kobo < 50000000 then '100k_500k' when min_price_kobo < 150000000 then '500k_1_5m' else 'over_1_5m' end as bucket from m) x group by bucket`, w)
	frows, err := s.d.Pool.Query(ctx, facetSQL, fargs...)
	if err != nil {
		return nil, fmt.Errorf("facets: %w", err)
	}
	defer frows.Close()
	for frows.Next() {
		var kind, value, label string
		var n int
		if err := frows.Scan(&kind, &value, &label, &n); err != nil {
			return nil, err
		}
		if kind == "total" {
			res.Total = n
			continue
		}
		res.Facets[kind] = append(res.Facets[kind], Facet{Value: value, Label: label, Count: n})
	}
	return res, frows.Err()
}

// didYouMean finds the closest title word for a zero-result query (trigram similarity).
func (s *Service) didYouMean(ctx context.Context, text string) string {
	var best string
	err := s.d.Pool.QueryRow(ctx, `select w from (
	    select distinct lower(regexp_split_to_table(title, '\s+')) as w from catalog.product_offer_summary
	    union select lower(name) from catalog.categories union select lower(name) from catalog.brands) words
	  where similarity(w, $1) > 0.3 order by similarity(w, $1) desc limit 1`, text).Scan(&best)
	if err != nil {
		return ""
	}
	return best
}
