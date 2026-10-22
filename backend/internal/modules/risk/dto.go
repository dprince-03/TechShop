package risk

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// RiskCaseRow is a line of the case queue.
type RiskCaseRow struct {
	Reference   string     `json:"reference"`
	Kind        string     `json:"kind" enum:"order,payment,device,seller,account,trade_in,return,review"`
	Reason      string     `json:"reason"`
	Score       *int16     `json:"score"`
	MoneyAtRisk *int64     `json:"moneyAtRisk" doc:"Kobo"`
	Status      string     `json:"status" enum:"open,in_review,cleared,blocked,escalated"`
	Label       *string    `json:"label"`
	SLADueAt    *time.Time `json:"slaDueAt"`
	Overdue     bool       `json:"overdue"`
	Assignee    *string    `json:"assignee"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// RiskCaseEvent is one line of a case's timeline.
type RiskCaseEvent struct {
	Action    string    `json:"action" enum:"opened,claimed,note,escalated,cleared,blocked"`
	Actor     *string   `json:"actor"`
	Note      *string   `json:"note"`
	CreatedAt time.Time `json:"createdAt"`
}

// RiskTrace is the decision behind a case: rules fired with weights, and the features seen.
type RiskTrace struct {
	Checkpoint     string             `json:"checkpoint"`
	Score          int16              `json:"score"`
	Outcome        string             `json:"outcome"`
	RuleSetVersion *int32             `json:"ruleSetVersion"`
	Fired          []FiredRule        `json:"fired"`
	Features       map[string]float64 `json:"features"`
	CreatedAt      time.Time          `json:"createdAt"`
}

// RiskCase is one case with its timeline and decision trace.
type RiskCase struct {
	RiskCaseRow
	SubjectType    string          `json:"subjectType"`
	SubjectID      string          `json:"subjectId"`
	Subject        string          `json:"subject" doc:"Order or return number, trade-in reference, shop, or a masked name"`
	ResolutionNote *string         `json:"resolutionNote"`
	ResolvedAt     *time.Time      `json:"resolvedAt"`
	RiskTrace      *RiskTrace      `json:"trace"`
	Events         []RiskCaseEvent `json:"events"`
}

func name(first, last *string) *string {
	if first == nil {
		return nil
	}
	n := *first
	if last != nil && *last != "" {
		n += " " + (*last)[:1] + "."
	}
	return &n
}

func caseRow(c store.RiskRiskCase, first, last *string) RiskCaseRow {
	return RiskCaseRow{Reference: c.Reference, Kind: c.Kind, Reason: c.Reason, Score: c.Score, MoneyAtRisk: c.MoneyAtRiskKobo, Status: c.Status, Label: c.Label,
		SLADueAt: c.SlaDueAt, Overdue: c.SlaDueAt != nil && time.Now().After(*c.SlaDueAt) && (c.Status == "open" || c.Status == "in_review" || c.Status == "escalated"),
		Assignee: name(first, last), CreatedAt: c.CreatedAt}
}

// CaseView loads a case for staff.
func (s *Service) CaseView(ctx context.Context, ref string) (RiskCase, error) {
	r, err := s.d.Q.RiskCaseByRef(ctx, ref)
	if errors.Is(err, pgx.ErrNoRows) {
		return RiskCase{}, httpx.NotFound("Case")
	}
	if err != nil {
		return RiskCase{}, err
	}
	c := store.RiskRiskCase{ID: r.ID, Kind: r.Kind, SubjectType: r.SubjectType, SubjectID: r.SubjectID, Reason: r.Reason, Score: r.Score, MoneyAtRiskKobo: r.MoneyAtRiskKobo,
		DecisionID: r.DecisionID, Status: r.Status, Label: r.Label, SlaDueAt: r.SlaDueAt, AssignedTo: r.AssignedTo, ResolvedAt: r.ResolvedAt, CreatedAt: r.CreatedAt, Reference: r.Reference}
	out := RiskCase{RiskCaseRow: caseRow(c, r.AssigneeFirstName, r.AssigneeLastName), SubjectType: r.SubjectType, SubjectID: r.SubjectID.String(),
		ResolutionNote: r.ResolutionNote, ResolvedAt: r.ResolvedAt, Events: []RiskCaseEvent{}}
	if out.Subject, err = s.d.Q.RiskSubjectLabel(ctx, r.SubjectID); err != nil {
		return RiskCase{}, err
	}
	if r.DecisionID != nil {
		d, err := s.d.Q.RiskDecision(ctx, *r.DecisionID)
		if err == nil {
			var why reasons
			_ = json.Unmarshal(d.Reasons, &why)
			if why.Fired == nil {
				why.Fired = []FiredRule{}
			}
			out.RiskTrace = &RiskTrace{Checkpoint: d.Checkpoint, Score: d.Score, Outcome: d.Outcome, RuleSetVersion: d.RuleSetVersion, Fired: why.Fired, Features: why.Features, CreatedAt: d.CreatedAt}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return RiskCase{}, err
		}
	}
	evs, err := s.d.Q.RiskCaseEvents(ctx, r.ID)
	if err != nil {
		return RiskCase{}, err
	}
	for _, e := range evs {
		out.Events = append(out.Events, RiskCaseEvent{Action: e.Action, Actor: name(e.FirstName, e.LastName), Note: e.Note, CreatedAt: e.CreatedAt})
	}
	return out, nil
}

// BlockedDevice is a blocklisted IMEI.
type BlockedDevice struct {
	IMEI      string    `json:"imei"`
	Reason    string    `json:"reason" enum:"reported_stolen,fraud,counterfeit"`
	Source    string    `json:"source"`
	Reference *string   `json:"reference"`
	AddedBy   *string   `json:"addedBy"`
	CreatedAt time.Time `json:"createdAt"`
}

// Summary is the Risk page's headline numbers.
type RiskSummary struct {
	Active         int32 `json:"active" doc:"Open, in review or escalated"`
	Orders         int32 `json:"orders" doc:"Active cases about orders"`
	Overdue        int32 `json:"overdue" doc:"Active cases past their review target"`
	Resolved7d     int32 `json:"resolved7d"`
	BlockedDevices int32 `json:"blockedDevices"`
}
