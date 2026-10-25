package support

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/pkg/validate"
)

// TicketMessage is one message in a ticket.
type TicketMessage struct {
	ID         uuid.UUID `json:"id"`
	AuthorKind string    `json:"authorKind" enum:"customer,seller,staff,system"`
	Author     string    `json:"author" doc:"Staff show as 'TechShop support · first name'"`
	Body       string    `json:"body"`
	Internal   bool      `json:"internal" doc:"Staff only: internal notes never reach customers"`
	CreatedAt  time.Time `json:"createdAt"`
}

// TicketRow is a line of a ticket list.
type TicketRow struct {
	Number        string    `json:"number"`
	Topic         string    `json:"topic"`
	Subject       string    `json:"subject"`
	Status        string    `json:"status" enum:"open,pending_customer,resolved,closed"`
	OrderNumber   *string   `json:"orderNumber"`
	Unread        bool      `json:"unread" doc:"A reply you haven't opened yet"`
	CreatedAt     time.Time `json:"createdAt"`
	LastMessageAt time.Time `json:"lastMessageAt"`
}

// Ticket is a ticket as its customer or seller sees it.
type Ticket struct {
	TicketRow
	Messages []TicketMessage `json:"messages"`
}

// StaffTicketRow is a line of the staff queue.
type StaffTicketRow struct {
	Number        string     `json:"number"`
	Topic         string     `json:"topic"`
	Subject       string     `json:"subject"`
	Status        string     `json:"status" enum:"open,pending_customer,resolved,closed"`
	Priority      string     `json:"priority" enum:"low,normal,high,urgent"`
	Channel       string     `json:"channel"`
	Site          *string    `json:"site"`
	Requester     string     `json:"requester" doc:"Masked name, shop name, or the contact form's name"`
	RequesterKind string     `json:"requesterKind" enum:"customer,seller,visitor"`
	OrderNumber   *string    `json:"orderNumber"`
	Assignee      *string    `json:"assignee"`
	AssigneeID    *uuid.UUID `json:"assigneeId"`
	SLADueAt      *time.Time `json:"slaDueAt"`
	Overdue       bool       `json:"overdue" doc:"No first reply yet and past the reply target"`
	FirstReplyAt  *time.Time `json:"firstReplyAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	LastMessageAt time.Time  `json:"lastMessageAt"`
}

// StaffTicket is a ticket with every message, internal notes included.
type StaffTicket struct {
	StaffTicketRow
	Contact  string          `json:"contact" doc:"Masked phone or email, for contact-form tickets"`
	Messages []TicketMessage `json:"messages"`
}

// SupportSummary is the support page's headline numbers.
type SupportSummary struct {
	Open                 int32 `json:"open"`
	Urgent               int32 `json:"urgent" doc:"Open and high or urgent"`
	Overdue              int32 `json:"overdue" doc:"Open, no first reply, past target"`
	Waiting              int32 `json:"waiting" doc:"Waiting on the customer"`
	ResolvedToday        int32 `json:"resolvedToday"`
	AvgFirstReplyMinutes int32 `json:"avgFirstReplyMinutes" doc:"Last 7 days"`
}

// SupportAgent is someone tickets can be assigned to.
type SupportAgent struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func masked(first, last *string) string {
	if first == nil {
		return ""
	}
	n := *first
	if last != nil && *last != "" {
		n += " " + (*last)[:1] + "."
	}
	return n
}

func maskEmail(e string) string {
	at := strings.IndexByte(e, '@')
	if at <= 0 {
		return "***"
	}
	keep := min(2, at)
	return e[:keep] + "***" + e[at:]
}

func messages(rows []store.SupportMessagesRow, staffView bool) []TicketMessage {
	out := make([]TicketMessage, 0, len(rows))
	for _, m := range rows {
		if m.IsInternal && !staffView {
			continue
		}
		author := ""
		switch m.AuthorKind {
		case "staff":
			author = "TechShop support"
			if m.FirstName != nil {
				author += " · " + *m.FirstName
			}
			if staffView && m.LastName != nil {
				author = masked(m.FirstName, m.LastName)
			}
		case "system":
			author = "TechShop"
		default:
			author = masked(m.FirstName, m.LastName)
			if author == "" {
				author = "You"
			}
		}
		out = append(out, TicketMessage{ID: m.ID, AuthorKind: m.AuthorKind, Author: author, Body: m.Body, Internal: m.IsInternal, CreatedAt: m.CreatedAt})
	}
	return out
}

func row(t store.SupportSupportTicket, order *string) TicketRow {
	subject := ""
	if t.Subject != nil {
		subject = *t.Subject
	}
	return TicketRow{Number: t.TicketNumber, Topic: t.Topic, Subject: subject, Status: t.Status, OrderNumber: order, Unread: t.CustomerUnread,
		CreatedAt: t.CreatedAt, LastMessageAt: t.LastMessageAt}
}

// CustomerTicket loads one of the caller's tickets and marks the reply as read.
func (s *Service) CustomerTicket(ctx context.Context, number string) (Ticket, error) {
	userID := httpx.MustPrincipal(ctx).UserID
	t, err := s.d.Q.SupportTicket(ctx, number)
	if err != nil || t.CustomerUserID == nil || *t.CustomerUserID != userID || t.SellerID != nil {
		return Ticket{}, httpx.NotFound("Ticket")
	}
	return s.requesterView(ctx, s.d.Q, t)
}

// ShopTicket loads one of a shop's tickets.
func (s *Service) ShopTicket(ctx context.Context, shopID uuid.UUID, number string) (Ticket, error) {
	var out Ticket
	err := s.sellers.InShop(ctx, shopID, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
		t, err := tx.Q.SupportTicket(ctx, number)
		if err != nil || t.SellerID == nil || *t.SellerID != shopID {
			return httpx.NotFound("Ticket")
		}
		out, err = s.requesterView(ctx, tx.Q, t)
		return err
	})
	return out, err
}

func (s *Service) requesterView(ctx context.Context, q *store.Queries, t store.SupportTicketRow) (Ticket, error) {
	ms, err := q.SupportMessages(ctx, t.ID)
	if err != nil {
		return Ticket{}, err
	}
	base := store.SupportSupportTicket{TicketNumber: t.TicketNumber, Topic: t.Topic, Subject: t.Subject, Status: t.Status, CustomerUnread: t.CustomerUnread,
		CreatedAt: t.CreatedAt, LastMessageAt: t.LastMessageAt}
	out := Ticket{TicketRow: row(base, t.OrderNumber), Messages: messages(ms, false)}
	if t.CustomerUnread {
		if err := q.SupportMarkRead(ctx, t.ID); err != nil {
			return Ticket{}, err
		}
	}
	return out, nil
}

type staffRowSource struct {
	t                                      store.SupportSupportTicket
	order, shop                            *string
	custFirst, custLast, asgFirst, asgLast *string
}

func staffRow(x staffRowSource) StaffTicketRow {
	t := x.t
	subject := ""
	if t.Subject != nil {
		subject = *t.Subject
	}
	r := StaffTicketRow{Number: t.TicketNumber, Topic: t.Topic, Subject: subject, Status: t.Status, Priority: t.Priority, Channel: t.Channel, Site: t.Site,
		OrderNumber: x.order, AssigneeID: t.AssignedTo, SLADueAt: t.SlaDueAt, FirstReplyAt: t.FirstResponseAt, CreatedAt: t.CreatedAt, LastMessageAt: t.LastMessageAt,
		Overdue: t.Status == "open" && t.FirstResponseAt == nil && t.SlaDueAt != nil && time.Now().After(*t.SlaDueAt)}
	if n := masked(x.asgFirst, x.asgLast); n != "" {
		r.Assignee = &n
	}
	switch {
	case t.SellerID != nil:
		r.RequesterKind = "seller"
		if x.shop != nil {
			r.Requester = *x.shop
		}
	case t.CustomerUserID != nil:
		r.RequesterKind, r.Requester = "customer", masked(x.custFirst, x.custLast)
	default:
		r.RequesterKind = "visitor"
		if t.ContactName != nil {
			r.Requester = *t.ContactName
		}
	}
	return r
}

// StaffTicket loads a ticket with every message for staff.
func (s *Service) StaffTicket(ctx context.Context, number string) (StaffTicket, error) {
	t, err := s.d.Q.SupportTicket(ctx, number)
	if notFound(err) {
		return StaffTicket{}, httpx.NotFound("Ticket")
	}
	if err != nil {
		return StaffTicket{}, err
	}
	ms, err := s.d.Q.SupportMessages(ctx, t.ID)
	if err != nil {
		return StaffTicket{}, err
	}
	base := store.SupportSupportTicket{ID: t.ID, TicketNumber: t.TicketNumber, CustomerUserID: t.CustomerUserID, SellerID: t.SellerID, Site: t.Site, ContactName: t.ContactName,
		ContactEmail: t.ContactEmail, ContactPhone: t.ContactPhone, Channel: t.Channel, Topic: t.Topic, Priority: t.Priority, Status: t.Status, AssignedTo: t.AssignedTo,
		FirstResponseAt: t.FirstResponseAt, CreatedAt: t.CreatedAt, Subject: t.Subject, SlaDueAt: t.SlaDueAt, LastMessageAt: t.LastMessageAt}
	out := StaffTicket{StaffTicketRow: staffRow(staffRowSource{t: base, order: t.OrderNumber, shop: t.ShopName, custFirst: t.CustomerFirstName, custLast: t.CustomerLastName,
		asgFirst: t.AssigneeFirstName, asgLast: t.AssigneeLastName}), Messages: messages(ms, true)}
	switch {
	case t.ContactPhone != nil:
		out.Contact = validate.MaskPhone(*t.ContactPhone)
	case t.ContactEmail != nil:
		out.Contact = maskEmail(*t.ContactEmail)
	}
	return out, nil
}
