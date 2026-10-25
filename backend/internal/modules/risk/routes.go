package risk

import (
	"context"
	"net/http"
	"strings"

	"github.com/dprince-03/techshop/backend/internal/platform/auth"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	s.ruleRoutes(r)
	g := r.Public
	st := func(perm string, o httpx.Op) httpx.Op {
		o.Tag = "Staff · Risk"
		o.Auth = []string{auth.AudStaff}
		o.Perm = perm
		return o
	}

	httpx.Register(g, st("risk.view", httpx.Op{ID: "staffRiskSummary", Method: http.MethodGet, Path: "/api/v1/staff/risk/summary", Summary: "Case and blocklist counts"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body RiskSummary }, error) {
			c, err := s.d.Q.RiskCaseCounts(ctx)
			if err != nil {
				return nil, err
			}
			n, err := s.d.Q.RiskBlocklistCount(ctx)
			return &struct{ Body RiskSummary }{RiskSummary{Active: c.Active, Orders: c.Orders, Overdue: c.Overdue, Resolved7d: c.Resolved7d, BlockedDevices: n}}, err
		})
	httpx.Register(g, st("risk.view", httpx.Op{ID: "staffRiskCases", Method: http.MethodGet, Path: "/api/v1/staff/risk/cases", Summary: "The case queue: open work first, by score × money at risk"}),
		func(ctx context.Context, in *struct {
			Status string `query:"status" enum:"active,open,in_review,escalated,cleared,blocked,"`
			Kind   string `query:"kind" enum:"order,payment,device,seller,account,trade_in,return,review,"`
		}) (*struct{ Body []RiskCaseRow }, error) {
			rows, err := s.d.Q.RiskCases(ctx, store.RiskCasesParams{Status: optional(in.Status), Kind: optional(in.Kind)})
			out := make([]RiskCaseRow, len(rows))
			for i, r := range rows {
				out[i] = caseRow(store.RiskRiskCase{Reference: r.Reference, Kind: r.Kind, Reason: r.Reason, Score: r.Score, MoneyAtRiskKobo: r.MoneyAtRiskKobo,
					Status: r.Status, Label: r.Label, SlaDueAt: r.SlaDueAt, CreatedAt: r.CreatedAt}, r.AssigneeFirstName, r.AssigneeLastName)
			}
			return &struct{ Body []RiskCaseRow }{out}, err
		})
	httpx.Register(g, st("risk.view", httpx.Op{ID: "staffRiskCase", Method: http.MethodGet, Path: "/api/v1/staff/risk/cases/{ref}", Summary: "A case with its decision trace and timeline"}),
		func(ctx context.Context, in *struct {
			Ref string `path:"ref"`
		}) (*struct{ Body RiskCase }, error) {
			out, err := s.CaseView(ctx, in.Ref)
			return &struct{ Body RiskCase }{out}, err
		})
	type noteBody struct {
		Note string `json:"note" maxLength:"1000"`
	}
	httpx.Register(g, st("risk.resolve", httpx.Op{ID: "staffClaimRiskCase", Method: http.MethodPost, Path: "/api/v1/staff/risk/cases/{ref}/claim", Summary: "Take a case for review"}),
		func(ctx context.Context, in *struct {
			Ref string `path:"ref"`
		}) (*struct{ Body RiskCase }, error) {
			out, err := s.Claim(ctx, in.Ref)
			return &struct{ Body RiskCase }{out}, err
		})
	httpx.Register(g, st("risk.resolve", httpx.Op{ID: "staffNoteRiskCase", Method: http.MethodPost, Path: "/api/v1/staff/risk/cases/{ref}/notes", Summary: "Add a note to a case"}),
		func(ctx context.Context, in *struct {
			Ref  string `path:"ref"`
			Body noteBody
		}) (*struct{ Body RiskCase }, error) {
			out, err := s.Note(ctx, in.Ref, in.Body.Note)
			return &struct{ Body RiskCase }{out}, err
		})
	httpx.Register(g, st("risk.resolve", httpx.Op{ID: "staffEscalateRiskCase", Method: http.MethodPost, Path: "/api/v1/staff/risk/cases/{ref}/escalate", Summary: "Send a case to a manager"}),
		func(ctx context.Context, in *struct {
			Ref  string `path:"ref"`
			Body noteBody
		}) (*struct{ Body RiskCase }, error) {
			out, err := s.Escalate(ctx, in.Ref, in.Body.Note)
			return &struct{ Body RiskCase }{out}, err
		})
	httpx.Register(g, st("risk.resolve", httpx.Op{ID: "staffResolveRiskCase", Method: http.MethodPost, Path: "/api/v1/staff/risk/cases/{ref}/resolution",
		Summary: "Clear (legitimate: holds are released) or block (fraud: what was held is stopped)"}),
		func(ctx context.Context, in *struct {
			Ref  string `path:"ref"`
			Body struct {
				Decision string `json:"decision" enum:"clear,block"`
				Label    string `json:"label,omitempty" enum:"legit,unclear," doc:"For a clear: legit (default) or unclear"`
				Note     string `json:"note" maxLength:"1000"`
			}
		}) (*struct{ Body RiskCase }, error) {
			out, err := s.Resolve(ctx, in.Ref, in.Body.Decision, in.Body.Label, in.Body.Note)
			return &struct{ Body RiskCase }{out}, err
		})

	httpx.Register(g, st("risk.view", httpx.Op{ID: "staffBlockedDevices", Method: http.MethodGet, Path: "/api/v1/staff/risk/devices", Summary: "The stolen-device blocklist, newest first"}),
		func(ctx context.Context, in *struct {
			IMEI string `query:"imei"`
		}) (*struct{ Body []BlockedDevice }, error) {
			rows, err := s.d.Q.RiskBlocklist(ctx, optional(strings.ReplaceAll(in.IMEI, " ", "")))
			out := make([]BlockedDevice, len(rows))
			for i, r := range rows {
				out[i] = BlockedDevice{IMEI: r.Imei, Reason: r.Reason, Source: r.Source, Reference: r.Reference, AddedBy: r.CreatedByFirstName, CreatedAt: r.CreatedAt}
			}
			return &struct{ Body []BlockedDevice }{out}, err
		})
	httpx.Register(g, st("risk.view", httpx.Op{ID: "staffCheckIMEI", Method: http.MethodGet, Path: "/api/v1/staff/risk/devices/check", Summary: "Check an IMEI: check digit and blocklist"}),
		func(ctx context.Context, in *struct {
			IMEI string `query:"imei" required:"true" maxLength:"20"`
		}) (*struct{ Body IMEIStatus }, error) {
			out, err := CheckIMEI(ctx, s.d.Q, in.IMEI)
			return &struct{ Body IMEIStatus }{out}, err
		})
	httpx.Register(g, st("devices.block", httpx.Op{ID: "staffBlockDevice", Method: http.MethodPost, Path: "/api/v1/staff/risk/devices", Status: 201, Summary: "Add an IMEI to the blocklist"}),
		func(ctx context.Context, in *struct{ Body BlockDeviceInput }) (*struct{ Body BlockedDevice }, error) {
			out, err := s.Block(ctx, in.Body)
			return &struct{ Body BlockedDevice }{out}, err
		})
	httpx.Register(g, st("devices.block", httpx.Op{ID: "staffUnblockDevice", Method: http.MethodPost, Path: "/api/v1/staff/risk/devices/{imei}/unblock", Status: 204, Summary: "Remove an IMEI from the blocklist"}),
		func(ctx context.Context, in *struct {
			IMEI string `path:"imei"`
			Body noteBody
		}) (*struct{}, error) {
			return nil, s.Unblock(ctx, in.IMEI, in.Body.Note)
		})
}
