// Package support runs help tickets: customers (market and app), sellers (Seller Centre) and
// visitors using a public contact form all land in one staff queue. Staff reply, add internal
// notes (never shown outside), assign and resolve; a daily job resolves tickets left waiting on
// the customer and closes resolved ones. Plan: docs/backend.md §5, docs/plan.md 2026-10-23.
package support

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/modules/notify"
	"github.com/dprince-03/techshop/backend/internal/modules/sellers"
	"github.com/dprince-03/techshop/backend/internal/platform/auth"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/outbox"
	"github.com/dprince-03/techshop/backend/internal/platform/ratelimit"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/pkg/validate"
)

// Service is the support module.
type Service struct {
	d       *kit.Deps
	notify  *notify.Service
	sellers *sellers.Service
}

// New creates the module.
func New(d *kit.Deps, n *notify.Service, sel *sellers.Service) *Service {
	return &Service{d: d, notify: n, sellers: sel}
}

// Name implements kit.Module.
func (s *Service) Name() string { return "support" }

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}

var (
	contactPerIP      = ratelimit.Rule{Name: "contact_ip", Max: 5, Window: time.Hour}
	contactPerAddress = ratelimit.Rule{Name: "contact_address", Max: 5, Window: 24 * time.Hour}
	ticketsPerUser    = ratelimit.Rule{Name: "ticket_user", Max: 10, Window: 24 * time.Hour}
)

// NewTicket is a customer's or seller's question.
type NewTicket struct {
	Topic       string `json:"topic" enum:"order,delivery,payment,refund,return,warranty,trade_in,account,selling,business,other"`
	Subject     string `json:"subject" minLength:"3" maxLength:"140"`
	Message     string `json:"message" minLength:"1" maxLength:"5000"`
	OrderNumber string `json:"orderNumber,omitempty" maxLength:"20" doc:"One of your orders, if it's about one"`
}

// ContactInput is the public contact form.
type ContactInput struct {
	Site        string `json:"site" enum:"corporate,market,wholesale,seller"`
	Name        string `json:"name" minLength:"2" maxLength:"80"`
	Email       string `json:"email,omitempty" maxLength:"254"`
	Phone       string `json:"phone,omitempty" maxLength:"20"`
	Topic       string `json:"topic" enum:"order,delivery,payment,refund,return,warranty,trade_in,account,selling,business,other"`
	Subject     string `json:"subject" minLength:"3" maxLength:"140"`
	Message     string `json:"message" minLength:"1" maxLength:"5000"`
	OrderNumber string `json:"orderNumber,omitempty" maxLength:"20"`
	Website     string `json:"website,omitempty" maxLength:"200" doc:"Leave empty: a hidden field that only bots fill in"`
}

// CreateForCustomer opens a ticket for a signed-in shopper.
func (s *Service) CreateForCustomer(ctx context.Context, in NewTicket) (Ticket, error) {
	p := httpx.MustPrincipal(ctx)
	body, ok := cleanBody(in.Message)
	if !ok {
		return Ticket{}, httpx.Invalid("message_required", "Tell us what's happening.")
	}
	if ok, err := s.d.Limiter.Allow(ctx, ticketsPerUser, p.UserID.String()); err == nil && !ok {
		return Ticket{}, httpx.E(429, "rate_limited", "You've opened a lot of tickets today. Reply on an open one instead.")
	}
	site, channel := "market", "web"
	if p.Audience == auth.AudCustomerApp {
		site, channel = "app", "app"
	}
	var number string
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		var orderID *uuid.UUID
		if n := strings.TrimSpace(in.OrderNumber); n != "" {
			id, err := tx.Q.SupportCustomerOrder(ctx, store.SupportCustomerOrderParams{OrderNumber: n, CustomerUserID: &p.UserID})
			if err != nil {
				return httpx.Invalid("unknown_order", "That isn't one of your orders.")
			}
			orderID = &id
		}
		t, err := s.create(ctx, tx, store.SupportCreateTicketParams{CustomerUserID: &p.UserID, OrderID: orderID, Site: &site, Channel: channel, Topic: in.Topic},
			in.Subject, body, &p.UserID, "customer")
		number = t.TicketNumber
		return err
	})
	if err != nil {
		return Ticket{}, err
	}
	return s.CustomerTicket(ctx, number)
}

// CreateForShop opens a ticket from a seller's shop.
func (s *Service) CreateForShop(ctx context.Context, shopID uuid.UUID, in NewTicket) (Ticket, error) {
	userID := httpx.MustPrincipal(ctx).UserID
	body, ok := cleanBody(in.Message)
	if !ok {
		return Ticket{}, httpx.Invalid("message_required", "Tell us what's happening.")
	}
	var number string
	site := "seller"
	err := s.sellers.InShop(ctx, shopID, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
		var orderID *uuid.UUID
		if n := strings.TrimSpace(in.OrderNumber); n != "" {
			id, err := tx.Q.SupportSellerOrder(ctx, store.SupportSellerOrderParams{OrderNumber: n, SellerID: shopID})
			if err != nil {
				return httpx.Invalid("unknown_order", "That order has nothing from your shop.")
			}
			orderID = &id
		}
		t, err := s.create(ctx, tx, store.SupportCreateTicketParams{SellerID: &shopID, OrderID: orderID, Site: &site, Channel: "web", Topic: in.Topic},
			in.Subject, body, &userID, "seller")
		number = t.TicketNumber
		return err
	})
	if err != nil {
		return Ticket{}, err
	}
	return s.ShopTicket(ctx, shopID, number)
}

// ContactReceipt is the public form's answer.
type ContactReceipt struct {
	Received     bool   `json:"received"`
	TicketNumber string `json:"ticketNumber,omitempty"`
}

// Contact files a public contact-form message as a ticket and confirms it to the sender.
func (s *Service) Contact(ctx context.Context, in ContactInput) (ContactReceipt, error) {
	if honeypotTripped(in.Website) {
		s.d.Logger.Info("contact form honeypot tripped", "site", in.Site)
		return ContactReceipt{Received: true}, nil // look successful so the bot moves on
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email != "" && !validEmail(email) {
		return ContactReceipt{}, httpx.Invalid("email_invalid", "Check the email address, e.g. name@example.com.")
	}
	phone := ""
	if strings.TrimSpace(in.Phone) != "" {
		if phone = validate.NormalisePhone(in.Phone); phone == "" {
			return ContactReceipt{}, httpx.Invalid("phone_invalid", "Enter a Nigerian mobile number, e.g. 0803 123 4567.")
		}
	}
	if email == "" && phone == "" {
		return ContactReceipt{}, httpx.Invalid("contact_required", "Give us an email address or phone number so we can reply.")
	}
	body, ok := cleanBody(in.Message)
	if !ok {
		return ContactReceipt{}, httpx.Invalid("message_required", "Tell us what's happening.")
	}
	address := email
	if address == "" {
		address = phone
	}
	if ok, err := s.d.Limiter.Allow(ctx, contactPerIP, httpx.MetaFrom(ctx).IP); err == nil && !ok {
		return ContactReceipt{}, httpx.E(429, "rate_limited", "Too many messages from here. Try again in an hour.")
	}
	if ok, err := s.d.Limiter.Allow(ctx, contactPerAddress, address); err == nil && !ok {
		return ContactReceipt{}, httpx.E(429, "rate_limited", "We've already got several messages from you today. We'll reply soon.")
	}
	if n := strings.TrimSpace(in.OrderNumber); n != "" {
		body = "Order: " + n + "\n\n" + body // not linked: a public form can't prove the order is theirs
	}
	var out ContactReceipt
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		name := strings.TrimSpace(in.Name)
		t, err := s.create(ctx, tx, store.SupportCreateTicketParams{Site: &in.Site, ContactName: &name, ContactEmail: ptr(email), ContactPhone: ptr(phone),
			Channel: "web", Topic: in.Topic}, in.Subject, body, nil, "customer")
		if err != nil {
			return err
		}
		out = ContactReceipt{Received: true, TicketNumber: t.TicketNumber}
		ch, to := "email", email
		if to == "" {
			ch, to = "sms", phone
		}
		return s.notify.Send(ctx, tx, notify.Msg{To: to, Channel: ch, Category: notify.CatAccount, Template: "ticket_received",
			Data: map[string]any{"ref": t.TicketNumber, "name": name}, Key: "ticket:" + t.TicketNumber + ":received"})
	})
	return out, err
}

func (s *Service) create(ctx context.Context, tx *uow.Tx, p store.SupportCreateTicketParams, subject, body string, author *uuid.UUID, kind string) (store.SupportSupportTicket, error) {
	p.Priority = priorityFor(p.Topic)
	due := slaDue(p.Priority, time.Now())
	p.SlaDueAt = &due
	subj := strings.TrimSpace(subject)
	p.Subject = &subj
	t, err := tx.Q.SupportCreateTicket(ctx, p)
	if err != nil {
		return t, httpx.DB(err, "ticket")
	}
	if _, err := tx.Q.SupportAddMessage(ctx, store.SupportAddMessageParams{TicketID: t.ID, AuthorUserID: author, AuthorKind: kind, Body: body}); err != nil {
		return t, err
	}
	tx.Emit("ticket", t.ID, "ticket.opened", map[string]any{"ticketId": t.ID, "number": t.TicketNumber, "topic": t.Topic, "priority": t.Priority})
	return t, nil
}

// ReplyAsCustomer adds the shopper's message; it reopens a resolved ticket.
func (s *Service) ReplyAsCustomer(ctx context.Context, number, body string) (Ticket, error) {
	userID := httpx.MustPrincipal(ctx).UserID
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		t, err := tx.Q.SupportTicketForUpdate(ctx, number)
		if err != nil || t.CustomerUserID == nil || *t.CustomerUserID != userID || t.SellerID != nil {
			return httpx.NotFound("Ticket")
		}
		return s.requesterReply(ctx, tx, t, &userID, "customer", body)
	})
	if err != nil {
		return Ticket{}, err
	}
	return s.CustomerTicket(ctx, number)
}

// ReplyAsShop adds a seller's message.
func (s *Service) ReplyAsShop(ctx context.Context, shopID uuid.UUID, number, body string) (Ticket, error) {
	userID := httpx.MustPrincipal(ctx).UserID
	err := s.sellers.InShop(ctx, shopID, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
		t, err := tx.Q.SupportTicketForUpdate(ctx, number)
		if err != nil || t.SellerID == nil || *t.SellerID != shopID {
			return httpx.NotFound("Ticket")
		}
		return s.requesterReply(ctx, tx, t, &userID, "seller", body)
	})
	if err != nil {
		return Ticket{}, err
	}
	return s.ShopTicket(ctx, shopID, number)
}

func (s *Service) requesterReply(ctx context.Context, tx *uow.Tx, t store.SupportSupportTicket, author *uuid.UUID, kind, body string) error {
	body, ok := cleanBody(body)
	if !ok {
		return httpx.Invalid("message_required", "Write a message.")
	}
	if t.Status == "closed" {
		return httpx.Conflict("ticket_closed", "This ticket is closed. Open a new one and mention "+t.TicketNumber+".")
	}
	if _, err := tx.Q.SupportAddMessage(ctx, store.SupportAddMessageParams{TicketID: t.ID, AuthorUserID: author, AuthorKind: kind, Body: body}); err != nil {
		return err
	}
	_, err := tx.Q.SupportCustomerReplied(ctx, t.ID)
	return err
}

// StaffReply is a staff message: public (the requester is told) or an internal note.
type StaffReply struct {
	Body     string `json:"body" minLength:"1" maxLength:"5000"`
	Internal bool   `json:"internal,omitempty" doc:"An internal note: never shown to the customer or seller"`
	Status   string `json:"status,omitempty" enum:"open,pending_customer,resolved," doc:"Ticket status after a public reply (default: waiting on the customer)"`
}

// Reply adds a staff message. A public reply notifies the requester and, by default, waits on
// them; an internal note changes nothing they can see.
func (s *Service) Reply(ctx context.Context, number string, in StaffReply) (StaffTicket, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	body, ok := cleanBody(in.Body)
	if !ok {
		return StaffTicket{}, httpx.Invalid("message_required", "Write a message.")
	}
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		t, err := tx.Q.SupportTicketForUpdate(ctx, number)
		if err != nil {
			return httpx.NotFound("Ticket")
		}
		if t.Status == "closed" {
			return httpx.Conflict("ticket_closed", "This ticket is closed.")
		}
		kind := "staff"
		m, err := tx.Q.SupportAddMessage(ctx, store.SupportAddMessageParams{TicketID: t.ID, AuthorUserID: &actor, AuthorKind: kind, Body: body, IsInternal: in.Internal})
		if err != nil {
			return err
		}
		if t.AssignedTo == nil { // whoever answers first owns it
			if _, err := tx.Q.SupportAssign(ctx, store.SupportAssignParams{ID: t.ID, AssignedTo: &actor}); err != nil {
				return err
			}
		}
		if in.Internal {
			return nil
		}
		status := in.Status
		if status == "" {
			status = "pending_customer"
		}
		if _, err := tx.Q.SupportStaffReplied(ctx, store.SupportStaffRepliedParams{ID: t.ID, Status: status}); err != nil {
			return err
		}
		return s.tellRequester(ctx, tx, t, m.ID, body)
	})
	if err != nil {
		return StaffTicket{}, err
	}
	return s.StaffTicket(ctx, number)
}

// tellRequester sends the "new reply" notice: in the app for account holders (plus email when
// they have one), and the reply itself by email or SMS for contact-form visitors.
func (s *Service) tellRequester(ctx context.Context, tx *uow.Tx, t store.SupportSupportTicket, msgID uuid.UUID, body string) error {
	key := "ticket:" + t.TicketNumber + ":reply:" + msgID.String() + ":"
	data := map[string]any{"ref": t.TicketNumber}
	userID := t.CustomerUserID
	if t.SellerID != nil {
		msgs, err := tx.Q.SupportPublicMessages(ctx, t.ID)
		if err != nil {
			return err
		}
		if len(msgs) > 0 {
			userID = msgs[0].AuthorUserID // the shop member who asked
		}
	}
	if userID != nil {
		for _, ch := range []string{"push", "in_app"} {
			if err := s.notify.Send(ctx, tx, notify.Msg{UserID: userID, Channel: ch, Category: notify.CatAccount, Template: "ticket_reply", Data: data, Key: key + ch}); err != nil {
				return err
			}
		}
		u, err := tx.Q.SupportUserContact(ctx, *userID)
		if err == nil && u.Email != nil {
			return s.notify.Send(ctx, tx, notify.Msg{UserID: userID, To: *u.Email, Channel: "email", Category: notify.CatAccount, Template: "ticket_reply", Data: data, Key: key + "email"})
		}
		return nil
	}
	data["body"] = body
	switch {
	case t.ContactEmail != nil:
		return s.notify.Send(ctx, tx, notify.Msg{To: *t.ContactEmail, Channel: "email", Category: notify.CatAccount, Template: "ticket_reply_contact", Data: data, Key: key + "email"})
	case t.ContactPhone != nil:
		data["body"] = excerpt(body, 300)
		return s.notify.Send(ctx, tx, notify.Msg{To: *t.ContactPhone, Channel: "sms", Category: notify.CatAccount, Template: "ticket_reply_contact", Data: data, Key: key + "sms"})
	}
	return nil
}

// Claim assigns a ticket to the caller.
func (s *Service) Claim(ctx context.Context, number string) (StaffTicket, error) {
	me := httpx.MustPrincipal(ctx).UserID
	return s.assign(ctx, number, &me)
}

// Assign gives a ticket to an agent, or unassigns it (nil).
func (s *Service) Assign(ctx context.Context, number string, to *uuid.UUID) (StaffTicket, error) {
	if to != nil {
		ok, err := s.d.Q.SupportIsAgent(ctx, *to)
		if err != nil {
			return StaffTicket{}, err
		}
		if !ok {
			return StaffTicket{}, httpx.Invalid("not_an_agent", "That person can't answer tickets.")
		}
	}
	return s.assign(ctx, number, to)
}

func (s *Service) assign(ctx context.Context, number string, to *uuid.UUID) (StaffTicket, error) {
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		t, err := tx.Q.SupportTicketForUpdate(ctx, number)
		if err != nil {
			return httpx.NotFound("Ticket")
		}
		if _, err := tx.Q.SupportAssign(ctx, store.SupportAssignParams{ID: t.ID, AssignedTo: to}); err != nil {
			return err
		}
		return tx.Audit("ticket.assigned", "ticket", t.ID.String(), map[string]any{"to": to})
	})
	if err != nil {
		return StaffTicket{}, err
	}
	return s.StaffTicket(ctx, number)
}

// SetStatus changes a ticket's status and, optionally, its priority.
func (s *Service) SetStatus(ctx context.Context, number, status, priority string) (StaffTicket, error) {
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		t, err := tx.Q.SupportTicketForUpdate(ctx, number)
		if err != nil {
			return httpx.NotFound("Ticket")
		}
		if t.Status == "closed" {
			return httpx.Conflict("ticket_closed", "This ticket is closed.")
		}
		_, err = tx.Q.SupportSetStatus(ctx, store.SupportSetStatusParams{ID: t.ID, Status: status, Priority: ptr(priority)})
		return err
	})
	if err != nil {
		return StaffTicket{}, err
	}
	return s.StaffTicket(ctx, number)
}

// Housekeep resolves tickets waiting on the customer for 7 days and closes tickets resolved for
// 4 days (Zendesk's default solved → closed is 4 days).
func (s *Service) Housekeep(ctx context.Context) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		rows, err := tx.Q.SupportAutoResolve(ctx)
		if err != nil {
			return err
		}
		for _, r := range rows {
			if _, err := tx.Q.SupportAddMessage(ctx, store.SupportAddMessageParams{TicketID: r.ID, AuthorKind: "system",
				Body: "We haven't heard back in 7 days, so we've marked this resolved. Reply within 4 days to reopen it."}); err != nil {
				return err
			}
		}
		return tx.Q.SupportAutoClose(ctx)
	})
}

func ptr(s string) *string {
	if s = strings.TrimSpace(s); s == "" {
		return nil
	}
	return &s
}

// validEmail accepts a plain address (no display name).
func validEmail(e string) bool {
	a, err := mail.ParseAddress(e)
	return err == nil && a.Address == e && strings.Contains(e[strings.LastIndexByte(e, '@')+1:], ".")
}

func notFound(err error) bool { return errors.Is(err, pgx.ErrNoRows) }
