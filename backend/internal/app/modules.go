package app

import (
	"context"

	"github.com/dprince-03/techshop/backend/internal/catalog"
	"github.com/dprince-03/techshop/backend/internal/content"
	"github.com/dprince-03/techshop/backend/internal/files"
	"github.com/dprince-03/techshop/backend/internal/identity"
	"github.com/dprince-03/techshop/backend/internal/inventory"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/ledger"
	"github.com/dprince-03/techshop/backend/internal/logistics"
	"github.com/dprince-03/techshop/backend/internal/marketing"
	"github.com/dprince-03/techshop/backend/internal/notify"
	"github.com/dprince-03/techshop/backend/internal/payments"
	"github.com/dprince-03/techshop/backend/internal/sales"
	"github.com/dprince-03/techshop/backend/internal/tracking"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/internal/wms"
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
	Mods = Registry{Notify: n, Identity: id, Catalog: cat, Inventory: inv, Ledger: led, Marketing: mk, Sales: sl, Payments: pay, Logistics: lg, WMS: wh}
	return []kit.Module{n, id, files.New(d), cat, inv, content.New(d), tracking.New(d), led, mk, sl, pay, lg, wh}
}
