package marketing

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestRulesSQLBindsEveryValue(t *testing.T) {
	min, days := 2, 30
	cat := uuid.New()
	r := Rules{States: []string{"LA'; drop table x;--"}, MinOrders: &min, NotOrderedWithinDays: &days, BoughtCategoryID: &cat, Consent: "sms"}
	q, args := r.SQL()
	if strings.Contains(q, "drop table") {
		t.Fatal("rule values must never be inlined")
	}
	if len(args) != 5 {
		t.Fatalf("want 5 bound args, got %d: %v", len(args), args)
	}
	for _, want := range []string{"u.status = 'active'", "staff_members", "$1::text[]", "$5"} {
		if !strings.Contains(q, want) {
			t.Errorf("query missing %q", want)
		}
	}
}

func TestBucketsAreStableAndSpread(t *testing.T) {
	c := uuid.New()
	u := uuid.New()
	if bucketOf(c, u) != bucketOf(c, u) {
		t.Fatal("bucket must be stable")
	}
	var holdout int
	for i := 0; i < 10000; i++ {
		if bucketOf(c, uuid.New()) < 5 {
			holdout++
		}
	}
	if holdout < 380 || holdout > 620 {
		t.Fatalf("5%% holdout drew %d of 10000", holdout)
	}
}

func TestPickByShare(t *testing.T) {
	shares := []share{{"A", 50}, {"B", 50}}
	if pick(shares, 10) != "A" || pick(shares, 60) != "B" || pick(shares, 100) != "B" {
		t.Fatal("pick by cumulative share is wrong")
	}
}

func TestPromotionApplies(t *testing.T) {
	cat, parent, l := uuid.New(), uuid.New(), uuid.New()
	if !(Promotion{}).Applies(l, nil) {
		t.Fatal("no targets = sitewide")
	}
	if !(Promotion{CategoryIDs: []uuid.UUID{parent}}).Applies(l, []uuid.UUID{cat, parent}) {
		t.Fatal("a parent category target covers its children")
	}
	if (Promotion{ListingIDs: []uuid.UUID{uuid.New()}}).Applies(l, []uuid.UUID{cat}) {
		t.Fatal("other listing must not match")
	}
}
