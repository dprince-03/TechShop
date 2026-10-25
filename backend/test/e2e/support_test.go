//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// A customer's ticket: priority from the topic, staff reply and internal note, waiting on the
// customer, the customer's reply reopens it, and nobody else can read it.
func TestSupportCustomerTicket(t *testing.T) {
	t.Parallel()
	cust := newCustomer(t, "Ngozi")
	expectCode(t, post(t, "/api/v1/me/tickets", cust, map[string]any{"topic": "refund", "subject": "Refund", "message": "x", "orderNumber": "TS-1"}), 422, "unknown_order", "someone else's order")
	r := post(t, "/api/v1/me/tickets", cust, map[string]any{"topic": "refund", "subject": "Refund not received", "message": "I returned my speaker last week and haven't been refunded."})
	expect(t, r, 201, "open ticket")
	number := r.str("number")
	if !strings.HasPrefix(number, "TKT-") || r.str("status") != "open" || len(r.get("messages").([]any)) != 1 {
		t.Fatalf("new ticket: %s", r.Raw)
	}

	agent := staffLogin(t, "support_agent")
	st := get(t, "/api/v1/staff/support/tickets/"+number, agent)
	expect(t, st, 200, "staff view")
	if st.str("priority") != "high" || st.str("requesterKind") != "customer" || st.str("requester") != "Ngozi C." || st.get("slaDueAt") == nil {
		t.Fatalf("refund tickets start high with a 4 h target: %s", st.Raw)
	}
	if q := get(t, "/api/v1/staff/support/tickets?status=active&q="+number, agent); len(q.List) != 1 {
		t.Fatalf("queue search: %s", q.Raw)
	}

	note := post(t, "/api/v1/staff/support/tickets/"+number+"/messages", agent, map[string]any{"body": "Checked: refund RF is waiting on finance approval.", "internal": true})
	expect(t, note, 200, "internal note")
	if note.str("status") != "open" || note.get("firstReplyAt") != nil || note.str("assignee") == "" {
		t.Fatalf("a note doesn't count as a reply but assigns the ticket: %s", note.Raw)
	}
	rep := post(t, "/api/v1/staff/support/tickets/"+number+"/messages", agent, map[string]any{"body": "Sorry about the wait. Your refund is approved today."})
	expect(t, rep, 200, "reply")
	if rep.str("status") != "pending_customer" || rep.get("firstReplyAt") == nil {
		t.Fatalf("a reply waits on the customer: %s", rep.Raw)
	}

	list := get(t, "/api/v1/me/tickets", cust)
	if len(list.List) != 1 || list.List[0].(map[string]any)["unread"] != true {
		t.Fatalf("customer sees an unread reply: %s", list.Raw)
	}
	mine := get(t, "/api/v1/me/tickets/"+number, cust)
	if strings.Contains(string(mine.Raw), "finance approval") || len(mine.get("messages").([]any)) != 2 {
		t.Fatalf("customers never see internal notes: %s", mine.Raw)
	}
	if got := get(t, "/api/v1/me/tickets", cust).List[0].(map[string]any)["unread"]; got != false {
		t.Fatal("opening the ticket marks it read")
	}
	expect(t, get(t, "/api/v1/me/tickets/"+number, newCustomer(t, "Nosy")), 404, "someone else's ticket")

	expect(t, post(t, "/api/v1/staff/support/tickets/"+number+"/status", agent, map[string]any{"status": "resolved"}), 200, "resolve")
	back := post(t, "/api/v1/me/tickets/"+number+"/messages", cust, map[string]any{"message": "It still hasn't arrived."})
	expect(t, back, 200, "customer reply")
	if back.str("status") != "open" {
		t.Fatalf("a customer reply reopens a resolved ticket: %s", back.Raw)
	}
	expect(t, post(t, "/api/v1/staff/support/tickets/"+number+"/status", agent, map[string]any{"status": "closed"}), 200, "close")
	expectCode(t, post(t, "/api/v1/me/tickets/"+number+"/messages", cust, map[string]any{"message": "Hello?"}), 409, "ticket_closed", "reply on closed")
	expect(t, post(t, "/api/v1/staff/support/tickets/"+number+"/messages", staffLogin(t, "warehouse_lead"), map[string]any{"body": "x"}), 403, "warehouse can't answer tickets")
}

// The public contact form: needs a way to reply, honeypot looks successful but files nothing,
// replies go back by email, and it is rate limited per address.
func TestSupportContactForm(t *testing.T) {
	t.Parallel()
	email := fmt.Sprintf("visitor%d@example.com", time.Now().UnixNano())
	form := func(extra map[string]any) map[string]any {
		m := map[string]any{"site": "corporate", "name": "Tunde Bakare", "email": email, "topic": "business", "subject": "Bulk laptops", "message": "We need 40 laptops for a school."}
		for k, v := range extra {
			m[k] = v
		}
		return m
	}
	expectCode(t, post(t, "/api/v1/contact", "", form(map[string]any{"email": ""})), 422, "contact_required", "no email or phone")
	expectCode(t, post(t, "/api/v1/contact", "", form(map[string]any{"email": "", "phone": "12345"})), 422, "phone_invalid", "bad phone")
	expectCode(t, post(t, "/api/v1/contact", "", form(map[string]any{"email": "tunde@"})), 422, "email_invalid", "bad email")
	bot := post(t, "/api/v1/contact", "", form(map[string]any{"website": "http://spam.example"}))
	expect(t, bot, 202, "honeypot")
	if bot.str("ticketNumber") != "" {
		t.Fatalf("a bot gets no ticket: %s", bot.Raw)
	}
	r := post(t, "/api/v1/contact", "", form(map[string]any{"orderNumber": "TS-10001"}))
	expect(t, r, 202, "contact")
	number := r.str("ticketNumber")

	agent := staffLogin(t, "support_lead")
	st := get(t, "/api/v1/staff/support/tickets/"+number, agent)
	if st.str("requesterKind") != "visitor" || st.str("requester") != "Tunde Bakare" || st.str("contact") == email || !strings.Contains(st.str("contact"), "***@example.com") ||
		!strings.Contains(string(st.Raw), "Order: TS-10001") {
		t.Fatalf("visitor ticket with masked contact and the order mentioned: %s", st.Raw)
	}
	agents := get(t, "/api/v1/staff/support/agents", agent)
	if len(agents.List) == 0 {
		t.Fatalf("agents: %s", agents.Raw)
	}
	as := post(t, "/api/v1/staff/support/tickets/"+number+"/assign", agent, map[string]any{"agentId": agents.List[0].(map[string]any)["id"]})
	expect(t, as, 200, "assign")
	expectCode(t, post(t, "/api/v1/staff/support/tickets/"+number+"/assign", agent, map[string]any{"agentId": newUUID()}), 422, "not_an_agent", "assign to a stranger")
	expect(t, post(t, "/api/v1/staff/support/tickets/"+number+"/assign", staffLogin(t, "support_agent"), map[string]any{}), 403, "agents can't reassign")
	expect(t, post(t, "/api/v1/staff/support/tickets/"+number+"/messages", agent, map[string]any{"body": "Thanks Tunde, our B2B team will call you today.", "status": "resolved"}), 200, "reply to visitor")

	for i := 0; i < 4; i++ { // 5 a day per address, the first one above included
		post(t, "/api/v1/contact", "", form(nil))
	}
	expect(t, post(t, "/api/v1/contact", "", form(nil)), 429, "per-address limit")
	if s := get(t, "/api/v1/staff/support/summary", agent); s.num("open") < 1 {
		t.Fatalf("summary: %s", s.Raw)
	}
}

// A seller's ticket stays inside the shop.
func TestSupportSellerTicket(t *testing.T) {
	t.Parallel()
	seller, shopID := approvedShop(t, "Kemi", "Ola")
	base := "/api/v1/seller/shops/" + shopID + "/tickets"
	r := post(t, base, seller, map[string]any{"topic": "selling", "subject": "Listing stuck in review", "message": "My laptop listing has been in review for 3 days."})
	expect(t, r, 201, "seller ticket")
	number := r.str("number")
	st := get(t, "/api/v1/staff/support/tickets/"+number, staffLogin(t, "support_agent"))
	if st.str("requesterKind") != "seller" || st.str("site") != "seller" {
		t.Fatalf("seller ticket in the queue: %s", st.Raw)
	}
	expect(t, post(t, "/api/v1/staff/support/tickets/"+number+"/messages", staffLogin(t, "support_agent"), map[string]any{"body": "Approved now, sorry for the delay."}), 200, "reply")
	if l := get(t, base, seller); len(l.List) != 1 || l.List[0].(map[string]any)["unread"] != true {
		t.Fatalf("shop tickets: %s", l.Raw)
	}
	expect(t, post(t, base+"/"+number+"/messages", seller, map[string]any{"message": "Thanks!"}), 200, "seller reply")
	other, otherShop := approvedShop(t, "Sola", "Ade")
	expect(t, get(t, "/api/v1/seller/shops/"+otherShop+"/tickets/"+number, other), 404, "another shop's ticket")
	expect(t, get(t, base+"/"+number, other), 404, "not a member")
	expect(t, get(t, "/api/v1/me/tickets/"+number, newCustomer(t, "Peek")), 404, "customer can't read shop tickets")
}
