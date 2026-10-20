// Package rbac defines the staff permission catalogue (module.action) and checks permissions
// live (cached 60 s, invalidated instantly by pg_notify('authz')). Plan: identity-access.md §7.
package rbac

// Permission describes one staff action. Sensitive ones need a recent MFA step-up.
type Permission struct {
	Key         string
	Module      string
	Description string
	Sensitive   bool
}

// Catalogue is the single source of permission keys; cmd/seed writes it to identity.permissions.
var Catalogue = []Permission{
	{"orders.view", "orders", "View orders", false},
	{"orders.edit", "orders", "Edit orders and fulfilments", false},
	{"orders.cancel", "orders", "Cancel orders or lines", false},
	{"customers.view", "customers", "View customers (masked contact)", false},
	{"customers.reveal_contact", "customers", "Reveal full customer contact details", true},
	{"support.view", "support", "View support tickets", false},
	{"support.reply", "support", "Reply to tickets", false},
	{"support.assign", "support", "Assign tickets", false},
	{"returns.create", "returns", "Create returns for customers", false},
	{"returns.manage", "returns", "Approve, inspect and resolve returns", false},
	{"refunds.request", "refunds", "Request refunds", false},
	{"refunds.approve_small", "refunds", "Approve refunds up to the threshold", true},
	{"refunds.approve", "refunds", "Approve refunds of any size", true},
	{"catalog.view", "catalog", "View the catalogue", false},
	{"catalog.edit", "catalog", "Create and edit products, variants and listings", false},
	{"catalog.moderate", "catalog", "Moderate seller listings and suggestions", false},
	{"search.manage", "search", "Manage synonyms, redirects and pins", false},
	{"content.edit", "content", "Edit CMS pages, banners and help", false},
	{"content.publish", "content", "Publish content", false},
	{"inventory.view", "inventory", "View stock", false},
	{"inventory.adjust", "inventory", "Adjust stock and approve count variances", true},
	{"fulfilment.pick", "fulfilment", "Pick and pack", false},
	{"fulfilment.manage", "fulfilment", "Release waves, manifests and carriers", false},
	{"purchasing.view", "purchasing", "View suppliers and purchase orders", false},
	{"purchasing.manage", "purchasing", "Create purchase orders and receive goods", false},
	{"purchasing.approve", "purchasing", "Approve large purchase orders", true},
	{"dispatch.view", "dispatch", "View dispatch", false},
	{"dispatch.assign", "dispatch", "Assign riders", false},
	{"riders.manage_devices", "dispatch", "Approve and revoke rider devices", false},
	{"logistics.manage", "logistics", "Manage zones, rates and riders", false},
	{"finance.view", "finance", "View finance", false},
	{"payouts.prepare", "finance", "Prepare payout batches", false},
	{"payouts.approve", "finance", "Approve payout batches", true},
	{"ledger.adjust", "finance", "Post adjustments and reversals", true},
	{"finance.period_close", "finance", "Close accounting periods", true},
	{"reconciliation.resolve", "finance", "Run reconciliation and resolve exceptions", false},
	{"disputes.decide", "finance", "Contest or accept chargebacks", true},
	{"risk.view", "risk", "View risk cases", false},
	{"risk.resolve", "risk", "Clear or block risk cases", false},
	{"risk.rules", "risk", "Publish risk rule sets", true},
	{"devices.block", "risk", "Manage the device blocklist", false},
	{"vendors.view", "vendors", "View sellers", false},
	{"vendors.kyc_review", "vendors", "Review seller KYC", false},
	{"vendors.reveal_id", "vendors", "Reveal a seller's full ID number", true},
	{"vendors.enforce", "vendors", "Apply seller enforcement", false},
	{"marketing.view", "marketing", "View marketing", false},
	{"marketing.edit", "marketing", "Create promotions, coupons and campaigns", false},
	{"marketing.segments", "marketing", "Build audiences", false},
	{"marketing.approve", "marketing", "Approve campaigns", true},
	{"marketing.journeys", "marketing", "Manage journeys", false},
	{"notify.templates", "notify", "Manage message templates and suppressions", false},
	{"business.view", "business", "View businesses and quotes", false},
	{"business.quotes", "business", "Price and send quotes", false},
	{"business.credit_request", "business", "Request credit for a business", false},
	{"business.credit_approve", "business", "Approve credit limits", true},
	{"pos.sell", "pos", "Sell at the till", false},
	{"pos.refund", "pos", "Refund at the till", true},
	{"pos.close_shift", "pos", "Close till shifts", false},
	{"cars.view", "cars", "View car listings and bookings", false},
	{"cars.manage", "cars", "Manage car listings, inspections and documents", false},
	{"aftersales.manage", "aftersales", "Warranty claims, repairs and trade-ins", false},
	{"hr.view", "hr", "View staff", false},
	{"hr.manage", "hr", "Add and exit staff", false},
	{"admin.roles", "admin", "Change roles and grant them", true},
	{"admin.users", "admin", "Manage user accounts", false},
	{"admin.mfa_reset", "admin", "Reset another person's MFA", true},
	{"admin.audit_view", "admin", "View audit and security logs", false},
	{"it.flags", "it", "Feature flags and job views", false},
	{"analytics.view", "analytics", "View analytics", false},
	{"recs.manage", "recs", "Recommendation overrides and model versions", false},
}

var sensitive = func() map[string]bool {
	m := map[string]bool{}
	for _, p := range Catalogue {
		if p.Sensitive {
			m[p.Key] = true
		}
	}
	return m
}()

// Keys returns every permission key.
func Keys() []string {
	out := make([]string, len(Catalogue))
	for i, p := range Catalogue {
		out[i] = p.Key
	}
	return out
}

// RoleTemplates are the PROPOSED staff roles (identity-access.md §7.2). Only "admin" is seeded
// outside development; the others are dev data until the owner approves the list.
var RoleTemplates = map[string]struct {
	Name  string
	Perms []string
}{
	"ops_manager":       {"Operations manager", []string{"orders.view", "orders.edit", "orders.cancel", "dispatch.view", "inventory.view", "support.view", "analytics.view", "customers.view"}},
	"support_agent":     {"Customer support agent", []string{"support.view", "support.reply", "orders.view", "returns.create", "customers.view", "refunds.request"}},
	"support_lead":      {"Support lead", []string{"support.view", "support.reply", "support.assign", "orders.view", "returns.create", "returns.manage", "customers.view", "customers.reveal_contact", "refunds.request"}},
	"catalog_editor":    {"Catalogue editor", []string{"catalog.view", "catalog.edit", "catalog.moderate", "search.manage", "content.edit"}},
	"warehouse_lead":    {"Warehouse lead", []string{"inventory.view", "inventory.adjust", "fulfilment.pick", "fulfilment.manage", "purchasing.view", "purchasing.manage", "returns.manage"}},
	"dispatcher":        {"Dispatcher", []string{"dispatch.view", "dispatch.assign", "riders.manage_devices", "orders.view", "logistics.manage"}},
	"finance_officer":   {"Finance officer", []string{"finance.view", "refunds.request", "refunds.approve_small", "payouts.prepare", "reconciliation.resolve", "orders.view"}},
	"finance_manager":   {"Finance manager", []string{"finance.view", "refunds.request", "refunds.approve_small", "refunds.approve", "payouts.approve", "ledger.adjust", "finance.period_close", "reconciliation.resolve", "disputes.decide", "business.credit_approve", "purchasing.approve", "orders.view"}},
	"risk_analyst":      {"Risk analyst", []string{"risk.view", "risk.resolve", "devices.block", "orders.view"}},
	"kyc_reviewer":      {"KYC reviewer", []string{"vendors.view", "vendors.kyc_review", "vendors.reveal_id", "vendors.enforce"}},
	"marketing_exec":    {"Marketing executive", []string{"marketing.view", "marketing.edit", "marketing.segments", "notify.templates"}},
	"marketing_manager": {"Marketing manager", []string{"marketing.view", "marketing.edit", "marketing.segments", "marketing.approve", "marketing.journeys", "notify.templates"}},
	"content_editor":    {"Content editor", []string{"content.edit", "content.publish"}},
	"b2b_manager":       {"B2B account manager", []string{"business.view", "business.quotes", "business.credit_request", "orders.view"}},
	"store_manager":     {"Store manager", []string{"pos.sell", "pos.refund", "pos.close_shift", "inventory.view"}},
	"cashier":           {"Store cashier", []string{"pos.sell"}},
	"aftersales":        {"After-sales technician", []string{"aftersales.manage", "returns.manage", "orders.view"}},
	"car_sales":         {"Car sales", []string{"cars.view", "cars.manage"}},
	"hr_officer":        {"HR officer", []string{"hr.view", "hr.manage"}},
	"auditor":           {"Auditor (read-only)", []string{"orders.view", "finance.view", "inventory.view", "vendors.view", "risk.view", "admin.audit_view", "analytics.view", "catalog.view"}},
}
