package support

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/platform/auth"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/jobs"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
)

var shoppers = []string{auth.AudMarket, auth.AudCustomerApp}

type ticketOut struct{ Body Ticket }
type staffOut struct{ Body StaffTicket }

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	type replyBody struct {
		Message string `json:"message" minLength:"1" maxLength:"5000"`
	}

	httpx.Register(g, httpx.Op{ID: "contactUs", Method: http.MethodPost, Path: "/api/v1/contact", Tag: "Support", Status: 202, MaxBody: 32 << 10,
		Summary: "Public contact form (corporate, market, wholesale): becomes a support ticket"},
		func(ctx context.Context, in *struct{ Body ContactInput }) (*struct{ Body ContactReceipt }, error) {
			out, err := s.Contact(ctx, in.Body)
			return &struct{ Body ContactReceipt }{out}, err
		})

	me := func(o httpx.Op) httpx.Op { o.Tag = "Support"; o.Auth = shoppers; return o }
	httpx.Register(g, me(httpx.Op{ID: "openTicket", Method: http.MethodPost, Path: "/api/v1/me/tickets", Status: 201, Summary: "Ask support a question"}),
		func(ctx context.Context, in *struct{ Body NewTicket }) (*ticketOut, error) {
			out, err := s.CreateForCustomer(ctx, in.Body)
			return &ticketOut{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "myTickets", Method: http.MethodGet, Path: "/api/v1/me/tickets", Summary: "Your support tickets, latest activity first"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []TicketRow }, error) {
			rows, err := s.d.Q.SupportTicketsForCustomer(ctx, &httpx.MustPrincipal(ctx).UserID)
			out := make([]TicketRow, len(rows))
			for i, t := range rows {
				out[i] = row(store.SupportSupportTicket{TicketNumber: t.TicketNumber, Topic: t.Topic, Subject: t.Subject, Status: t.Status, CustomerUnread: t.CustomerUnread,
					CreatedAt: t.CreatedAt, LastMessageAt: t.LastMessageAt}, t.OrderNumber)
			}
			return &struct{ Body []TicketRow }{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "myTicket", Method: http.MethodGet, Path: "/api/v1/me/tickets/{number}", Summary: "One of your tickets with its messages"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
		}) (*ticketOut, error) {
			out, err := s.CustomerTicket(ctx, in.Number)
			return &ticketOut{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "replyToMyTicket", Method: http.MethodPost, Path: "/api/v1/me/tickets/{number}/messages", Summary: "Reply on your ticket (reopens a resolved one)"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
			Body   replyBody
		}) (*ticketOut, error) {
			out, err := s.ReplyAsCustomer(ctx, in.Number, in.Body.Message)
			return &ticketOut{out}, err
		})

	sel := func(o httpx.Op) httpx.Op { o.Tag = "Seller Centre"; o.Auth = []string{auth.AudSeller}; return o }
	httpx.Register(g, sel(httpx.Op{ID: "sellerOpenTicket", Method: http.MethodPost, Path: "/api/v1/seller/shops/{shopId}/tickets", Status: 201, Summary: "Ask TechShop seller support"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			Body   NewTicket
		}) (*ticketOut, error) {
			out, err := s.CreateForShop(ctx, in.ShopID, in.Body)
			return &ticketOut{out}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerTickets", Method: http.MethodGet, Path: "/api/v1/seller/shops/{shopId}/tickets", Summary: "The shop's support tickets"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
		}) (*struct{ Body []TicketRow }, error) {
			var out []TicketRow
			err := s.sellers.InShop(ctx, in.ShopID, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
				rows, err := tx.Q.SupportTicketsForSeller(ctx, &in.ShopID)
				out = make([]TicketRow, len(rows))
				for i, t := range rows {
					out[i] = row(store.SupportSupportTicket{TicketNumber: t.TicketNumber, Topic: t.Topic, Subject: t.Subject, Status: t.Status, CustomerUnread: t.CustomerUnread,
						CreatedAt: t.CreatedAt, LastMessageAt: t.LastMessageAt}, t.OrderNumber)
				}
				return err
			})
			return &struct{ Body []TicketRow }{out}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerTicket", Method: http.MethodGet, Path: "/api/v1/seller/shops/{shopId}/tickets/{number}", Summary: "One of the shop's tickets"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			Number string    `path:"number"`
		}) (*ticketOut, error) {
			out, err := s.ShopTicket(ctx, in.ShopID, in.Number)
			return &ticketOut{out}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerReplyTicket", Method: http.MethodPost, Path: "/api/v1/seller/shops/{shopId}/tickets/{number}/messages", Summary: "Reply on a shop ticket"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			Number string    `path:"number"`
			Body   replyBody
		}) (*ticketOut, error) {
			out, err := s.ReplyAsShop(ctx, in.ShopID, in.Number, in.Body.Message)
			return &ticketOut{out}, err
		})

	st := func(perm string, o httpx.Op) httpx.Op {
		o.Tag = "Staff · Support"
		o.Auth = []string{auth.AudStaff}
		o.Perm = perm
		return o
	}
	httpx.Register(g, st("support.view", httpx.Op{ID: "staffSupportSummary", Method: http.MethodGet, Path: "/api/v1/staff/support/summary", Summary: "Queue counts and the average first reply"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body SupportSummary }, error) {
			r, err := s.d.Q.SupportSummary(ctx)
			return &struct{ Body SupportSummary }{SupportSummary{Open: r.Open, Urgent: r.Urgent, Overdue: r.Overdue, Waiting: r.Waiting, ResolvedToday: r.ResolvedToday, AvgFirstReplyMinutes: r.AvgFirstReplyMinutes}}, err
		})
	httpx.Register(g, st("support.view", httpx.Op{ID: "staffSupportTickets", Method: http.MethodGet, Path: "/api/v1/staff/support/tickets", Summary: "The ticket queue"}),
		func(ctx context.Context, in *struct {
			Status   string `query:"status" enum:"active,open,pending_customer,resolved,closed,"`
			Priority string `query:"priority" enum:"urgent_high,urgent,high,normal,low,"`
			Mine     bool   `query:"mine"`
			Q        string `query:"q" maxLength:"80" doc:"Ticket or order number, contact email, or words in the subject"`
		}) (*struct{ Body []StaffTicketRow }, error) {
			var assignee *uuid.UUID
			if in.Mine {
				assignee = &httpx.MustPrincipal(ctx).UserID
			}
			rows, err := s.d.Q.SupportQueue(ctx, store.SupportQueueParams{Status: ptr(in.Status), Priority: ptr(in.Priority), Assignee: assignee, Q: ptr(in.Q)})
			out := make([]StaffTicketRow, len(rows))
			for i, t := range rows {
				out[i] = staffRow(staffRowSource{t: store.SupportSupportTicket{TicketNumber: t.TicketNumber, CustomerUserID: t.CustomerUserID, SellerID: t.SellerID, Site: t.Site,
					ContactName: t.ContactName, Channel: t.Channel, Topic: t.Topic, Priority: t.Priority, Status: t.Status, AssignedTo: t.AssignedTo, FirstResponseAt: t.FirstResponseAt,
					CreatedAt: t.CreatedAt, Subject: t.Subject, SlaDueAt: t.SlaDueAt, LastMessageAt: t.LastMessageAt},
					order: t.OrderNumber, shop: t.ShopName, custFirst: t.CustomerFirstName, custLast: t.CustomerLastName, asgFirst: t.AssigneeFirstName, asgLast: t.AssigneeLastName})
			}
			return &struct{ Body []StaffTicketRow }{out}, err
		})
	httpx.Register(g, st("support.view", httpx.Op{ID: "staffSupportTicket", Method: http.MethodGet, Path: "/api/v1/staff/support/tickets/{number}", Summary: "A ticket with every message, internal notes included"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
		}) (*staffOut, error) {
			out, err := s.StaffTicket(ctx, in.Number)
			return &staffOut{out}, err
		})
	httpx.Register(g, st("support.reply", httpx.Op{ID: "staffReplyTicket", Method: http.MethodPost, Path: "/api/v1/staff/support/tickets/{number}/messages", Summary: "Reply to the requester, or add an internal note"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
			Body   StaffReply
		}) (*staffOut, error) {
			out, err := s.Reply(ctx, in.Number, in.Body)
			return &staffOut{out}, err
		})
	httpx.Register(g, st("support.reply", httpx.Op{ID: "staffClaimTicket", Method: http.MethodPost, Path: "/api/v1/staff/support/tickets/{number}/claim", Summary: "Assign a ticket to yourself"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
		}) (*staffOut, error) {
			out, err := s.Claim(ctx, in.Number)
			return &staffOut{out}, err
		})
	httpx.Register(g, st("support.reply", httpx.Op{ID: "staffTicketStatus", Method: http.MethodPost, Path: "/api/v1/staff/support/tickets/{number}/status", Summary: "Change a ticket's status or priority"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
			Body   struct {
				Status   string `json:"status" enum:"open,pending_customer,resolved,closed"`
				Priority string `json:"priority,omitempty" enum:"low,normal,high,urgent,"`
			}
		}) (*staffOut, error) {
			out, err := s.SetStatus(ctx, in.Number, in.Body.Status, in.Body.Priority)
			return &staffOut{out}, err
		})
	httpx.Register(g, st("support.assign", httpx.Op{ID: "staffAssignTicket", Method: http.MethodPost, Path: "/api/v1/staff/support/tickets/{number}/assign", Summary: "Assign a ticket to an agent (or nobody)"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
			Body   struct {
				AgentID *uuid.UUID `json:"agentId,omitempty" doc:"Leave out to unassign"`
			}
		}) (*staffOut, error) {
			out, err := s.Assign(ctx, in.Number, in.Body.AgentID)
			return &staffOut{out}, err
		})
	httpx.Register(g, st("support.assign", httpx.Op{ID: "staffSupportAgents", Method: http.MethodGet, Path: "/api/v1/staff/support/agents", Summary: "People who can answer tickets"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []SupportAgent }, error) {
			rows, err := s.d.Q.SupportAgents(ctx)
			out := make([]SupportAgent, len(rows))
			for i, a := range rows {
				out[i] = SupportAgent{ID: a.ID, Name: masked(&a.FirstName, &a.LastName)}
			}
			return &struct{ Body []SupportAgent }{out}, err
		})
}

type housekeepArgs struct{}

// Kind identifies the job.
func (housekeepArgs) Kind() string { return "support.housekeep" }

type housekeepWorker struct {
	river.WorkerDefaults[housekeepArgs]
	s *Service
}

func (w *housekeepWorker) Work(ctx context.Context, _ *river.Job[housekeepArgs]) error {
	return w.s.Housekeep(ctx)
}

// Jobs implements kit.Module: hourly auto-resolve and auto-close (backend.md §3 "support auto-close").
func (s *Service) Jobs(reg *jobs.Registry) {
	river.AddWorker(reg.Workers, &housekeepWorker{s: s})
	reg.Every(time.Hour, func() (river.JobArgs, *river.InsertOpts) { return housekeepArgs{}, nil })
}
