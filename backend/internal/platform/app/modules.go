package app

import (
	"context"

	"github.com/dprince-03/techshop/backend/internal/modules/aftersales"
	"github.com/dprince-03/techshop/backend/internal/modules/catalog"
	"github.com/dprince-03/techshop/backend/internal/modules/content"
	"github.com/dprince-03/techshop/backend/internal/modules/files"
	"github.com/dprince-03/techshop/backend/internal/modules/identity"
	"github.com/dprince-03/techshop/backend/internal/modules/inventory"
	"github.com/dprince-03/techshop/backend/internal/modules/ledger"
	"github.com/dprince-03/techshop/backend/internal/modules/logistics"
	"github.com/dprince-03/techshop/backend/internal/modules/marketing"
	"github.com/dprince-03/techshop/backend/internal/modules/notify"
	"github.com/dprince-03/techshop/backend/internal/modules/payments"
	"github.com/dprince-03/techshop/backend/internal/modules/reviews"
	"github.com/dprince-03/techshop/backend/internal/modules/sales"
	"github.com/dprince-03/techshop/backend/internal/modules/sellers"
	"github.com/dprince-03/techshop/backend/internal/modules/tracking"
	"github.com/dprince-03/techshop/backend/internal/modules/wms"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
)

// Registry exposes modules the seed and other modules need by type.
type Registry struct {
	Notify    *notify.Service
	Identity  *identity.Service
	Catalog   *catalog.Service
	Inventory *inventory.Service
	Ledger    *ledger.Service
	Marketing *marketing.Service
	Sales     *sales.Service
	Payments  *payments.Service
	Logistics *logistics.Service
	WMS       *wms.Service
	Sellers   *sellers.Service
	Returns   *aftersales.Service
}

// Mods is filled by buildModules (dependencies flow one way).
var Mods Registry

func buildModules(d *kit.Deps) []kit.Module {
	n := notify.New(d)
	id := identity.New(d, n)
	cat := catalog.New(d)
	inv := inventory.New(d)
	led := ledger.New(d)
	mk := marketing.New(d)
	mk.Send = func(ctx context.Context, tx *uow.Tx, m marketing.Message) error {
		u := m.UserID
		return n.Send(ctx, tx, notify.Msg{UserID: &u, To: m.To, Channel: m.Channel, Category: notify.CatDeals, Template: m.Template, Data: m.Data, Key: m.Key,
			CampaignID: m.CampaignID, JourneyEnrollmentID: m.JourneyEnrollmentID, TemplateVersionID: m.TemplateVersionID})
	}
	sl := sales.New(d, inv, mk, led, n)
	pay := payments.New(d, sl, led, n) // also plugs itself into sales as its PaymentPort
	lg := logistics.New(d, sl, n)      // also plugs zone pricing into sales
	wh := wms.New(d, sl, lg)
	rv := reviews.New(d)
	as := aftersales.New(d, pay, inv)          // refunds through payments, restock through inventory
	sel := sellers.New(d, pay, n, cat, sl, rv) // payments answers bank name enquiries; catalog owns listing rules
	Mods = Registry{Notify: n, Identity: id, Catalog: cat, Inventory: inv, Ledger: led, Marketing: mk, Sales: sl, Payments: pay, Logistics: lg, WMS: wh, Sellers: sel, Returns: as}
	return []kit.Module{n, id, files.New(d), cat, inv, content.New(d), tracking.New(d), led, mk, sl, pay, lg, wh, rv, sel, as}
}
