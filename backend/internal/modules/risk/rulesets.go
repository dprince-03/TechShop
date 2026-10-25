package risk

// The rules editor: a draft copied from the active rule set, weights, switches and thresholds
// edited, a test bench that replays the last 30 days of decisions against the draft, then
// publish (a sensitive permission with a fresh MFA). Plan: trust-safety.md §3 "Test bench".

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/platform/auth"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// RuleView is one rule as the editor shows it.
type RuleView struct {
	ID        uuid.UUID `json:"id"`
	Key       string    `json:"key"`
	Label     string    `json:"label"`
	Condition Condition `json:"condition"`
	Weight    int32     `json:"weight"`
	Action    string    `json:"action" enum:"score,hold,block"`
	Enabled   bool      `json:"enabled"`
}

// RuleSetView is a versioned rule set.
type RuleSetView struct {
	ID          uuid.UUID  `json:"id"`
	Checkpoint  string     `json:"checkpoint"`
	Version     int32      `json:"version"`
	Status      string     `json:"status" enum:"draft,active,retired"`
	Thresholds  Thresholds `json:"thresholds"`
	Notes       *string    `json:"notes"`
	PublishedAt *time.Time `json:"publishedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	Rules       []RuleView `json:"rules"`
}

func (s *Service) ruleSetView(ctx context.Context, q *store.Queries, rs store.RiskRiskRuleSet) (RuleSetView, error) {
	out := RuleSetView{ID: rs.ID, Checkpoint: rs.Checkpoint, Version: rs.Version, Status: rs.Status, Notes: rs.Notes, PublishedAt: rs.PublishedAt, CreatedAt: rs.CreatedAt, Rules: []RuleView{}}
	_ = json.Unmarshal(rs.Thresholds, &out.Thresholds)
	rows, err := q.RiskRules(ctx, rs.ID)
	if err != nil {
		return out, err
	}
	for _, r := range rows {
		c, _ := ParseCondition(r.Condition)
		out.Rules = append(out.Rules, RuleView{ID: r.ID, Key: r.Key, Label: r.Label, Condition: c, Weight: r.Weight, Action: r.Action, Enabled: r.Enabled})
	}
	return out, nil
}

// Draft copies the active rule set of a checkpoint into a new draft version.
func (s *Service) Draft(ctx context.Context, checkpoint, notes string) (RuleSetView, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	var out RuleSetView
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if n, err := tx.Q.RiskDraftFor(ctx, checkpoint); err != nil {
			return err
		} else if n > 0 {
			return httpx.Conflict("draft_exists", "There's already a draft for this checkpoint. Edit or discard it first.")
		}
		active, err := tx.Q.RiskActiveRuleSet(ctx, checkpoint)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.NotFound("Active rule set")
		}
		if err != nil {
			return err
		}
		v, err := tx.Q.RiskNextVersion(ctx, checkpoint)
		if err != nil {
			return err
		}
		d, err := tx.Q.RiskCreateRuleSet(ctx, store.RiskCreateRuleSetParams{Checkpoint: checkpoint, Version: v, Thresholds: active.Thresholds, CreatedBy: &actor, Notes: optional(notes)})
		if err != nil {
			return httpx.DB(err, "rule set")
		}
		if err := tx.Q.RiskCopyRules(ctx, store.RiskCopyRulesParams{Target: d.ID, Source: active.ID}); err != nil {
			return err
		}
		out, err = s.ruleSetView(ctx, tx.Q, d)
		return err
	})
	return out, err
}

func (s *Service) onDraft(ctx context.Context, id uuid.UUID, fn func(*uow.Tx, store.RiskRiskRuleSet) error) (RuleSetView, error) {
	var out RuleSetView
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		rs, err := tx.Q.RiskRuleSetForUpdate(ctx, id)
		if err != nil {
			return httpx.NotFound("Rule set")
		}
		if rs.Status != "draft" {
			return httpx.Conflict("not_a_draft", "Only drafts can be changed. Start a new draft from the active rule set.")
		}
		if err := fn(tx, rs); err != nil {
			return err
		}
		rs, err = tx.Q.RiskRuleSet(ctx, id)
		if err != nil {
			return err
		}
		out, err = s.ruleSetView(ctx, tx.Q, rs)
		return err
	})
	return out, err
}

// UpdateRule changes a draft rule's weight or switches it on or off.
func (s *Service) UpdateRule(ctx context.Context, setID, ruleID uuid.UUID, weight int32, enabled bool) (RuleSetView, error) {
	return s.onDraft(ctx, setID, func(tx *uow.Tx, _ store.RiskRiskRuleSet) error {
		if _, err := tx.Q.RiskUpdateRule(ctx, store.RiskUpdateRuleParams{ID: ruleID, RuleSetID: setID, Weight: weight, Enabled: enabled}); errors.Is(err, pgx.ErrNoRows) {
			return httpx.NotFound("Rule")
		} else if err != nil {
			return httpx.DB(err, "rule")
		}
		return nil
	})
}

// SetThresholds changes a draft's score bands.
func (s *Service) SetThresholds(ctx context.Context, setID uuid.UUID, t Thresholds) (RuleSetView, error) {
	if !t.Valid() {
		return RuleSetView{}, httpx.Invalid("thresholds_invalid", "Bands must rise: step up < hold < block, between 1 and 100.")
	}
	return s.onDraft(ctx, setID, func(tx *uow.Tx, _ store.RiskRiskRuleSet) error {
		raw, _ := json.Marshal(t)
		return tx.Q.RiskSetThresholds(ctx, store.RiskSetThresholdsParams{ID: setID, Thresholds: raw})
	})
}

// BenchResult is what a draft would have done over the last 30 days.
type BenchResult struct {
	Decisions    int            `json:"decisions" doc:"Decisions replayed (last 30 days, up to 5,000)"`
	Changed      int            `json:"changed" doc:"Decisions whose outcome would differ"`
	Before       map[string]int `json:"before" doc:"Outcomes as decided"`
	After        map[string]int `json:"after" doc:"Outcomes under the draft"`
	FraudCaught  int            `json:"fraudCaught" doc:"Confirmed fraud the draft would hold or block"`
	FraudMissed  int            `json:"fraudMissed" doc:"Confirmed fraud the draft would let through"`
	LegitStopped int            `json:"legitStopped" doc:"Cleared (legitimate) cases the draft would still hold or block"`
}

// Bench replays recent decisions of the draft's checkpoint against it.
func (s *Service) Bench(ctx context.Context, setID uuid.UUID) (BenchResult, error) {
	rs, err := s.d.Q.RiskRuleSet(ctx, setID)
	if err != nil {
		return BenchResult{}, httpx.NotFound("Rule set")
	}
	rules, t, err := loadRules(ctx, s.d.Q, rs)
	if err != nil {
		return BenchResult{}, httpx.Invalid("rule_set_invalid", err.Error())
	}
	rows, err := s.d.Q.RiskRecentDecisions(ctx, store.RiskRecentDecisionsParams{Checkpoint: rs.Checkpoint, CreatedAt: time.Now().Add(-30 * 24 * time.Hour)})
	if err != nil {
		return BenchResult{}, err
	}
	out := BenchResult{Before: map[string]int{}, After: map[string]int{}}
	stops := func(o string) bool { return o == "hold" || o == "block" || o == "step_up" }
	for _, r := range rows {
		var why reasons
		if json.Unmarshal(r.Reasons, &why) != nil || why.Features == nil {
			continue // a fallback decision has no features to replay
		}
		after := Score(rules, t, why.Features).Outcome
		out.Decisions++
		out.Before[r.Outcome]++
		out.After[after]++
		if after != r.Outcome {
			out.Changed++
		}
		switch {
		case r.Label != nil && *r.Label == "fraud" && stops(after):
			out.FraudCaught++
		case r.Label != nil && *r.Label == "fraud":
			out.FraudMissed++
		case r.Label != nil && *r.Label == "legit" && stops(after):
			out.LegitStopped++
		}
	}
	return out, nil
}

// Publish makes a draft the checkpoint's active rule set and retires the old one.
func (s *Service) Publish(ctx context.Context, setID uuid.UUID) (RuleSetView, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	return s.onDraft(ctx, setID, func(tx *uow.Tx, rs store.RiskRiskRuleSet) error {
		if _, _, err := loadRules(ctx, tx.Q, rs); err != nil {
			return httpx.Invalid("rule_set_invalid", err.Error())
		}
		if err := tx.Q.RiskRetireActive(ctx, rs.Checkpoint); err != nil {
			return err
		}
		if err := tx.Q.RiskPublish(ctx, store.RiskPublishParams{ID: rs.ID, PublishedBy: &actor}); err != nil {
			return err
		}
		return tx.Audit("risk.rules_published", "risk_rule_set", rs.ID.String(), map[string]any{"checkpoint": rs.Checkpoint, "version": rs.Version})
	})
}

// Discard deletes a draft.
func (s *Service) Discard(ctx context.Context, setID uuid.UUID) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		rs, err := tx.Q.RiskRuleSetForUpdate(ctx, setID)
		if err != nil || rs.Status != "draft" {
			return httpx.NotFound("Draft")
		}
		return tx.Q.RiskDeleteDraft(ctx, setID)
	})
}

func (s *Service) ruleRoutes(r *kit.Routes) {
	g := r.Public
	st := func(perm string, o httpx.Op) httpx.Op {
		o.Tag = "Staff · Risk"
		o.Auth = []string{auth.AudStaff}
		o.Perm = perm
		return o
	}
	type setOut struct{ Body RuleSetView }
	httpx.Register(g, st("risk.view", httpx.Op{ID: "staffRiskRuleSets", Method: http.MethodGet, Path: "/api/v1/staff/risk/rule-sets", Summary: "Active rule sets and drafts per checkpoint"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []RuleSetView }, error) {
			rows, err := s.d.Q.RiskRuleSets(ctx)
			if err != nil {
				return nil, err
			}
			out := make([]RuleSetView, 0, len(rows))
			for _, rs := range rows {
				v, err := s.ruleSetView(ctx, s.d.Q, rs)
				if err != nil {
					return nil, err
				}
				out = append(out, v)
			}
			return &struct{ Body []RuleSetView }{out}, nil
		})
	httpx.Register(g, st("risk.rules", httpx.Op{ID: "staffRiskDraft", Method: http.MethodPost, Path: "/api/v1/staff/risk/rule-sets/drafts", Status: 201, Summary: "Start a draft from a checkpoint's active rules"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Checkpoint string `json:"checkpoint" enum:"checkout,payout,trade_in,return"`
				Notes      string `json:"notes,omitempty" maxLength:"500" doc:"Why the change"`
			}
		}) (*setOut, error) {
			out, err := s.Draft(ctx, in.Body.Checkpoint, strings.TrimSpace(in.Body.Notes))
			return &setOut{out}, err
		})
	httpx.Register(g, st("risk.rules", httpx.Op{ID: "staffRiskUpdateRule", Method: http.MethodPost, Path: "/api/v1/staff/risk/rule-sets/{id}/rules/{ruleId}", Summary: "Change a draft rule's weight or switch it"}),
		func(ctx context.Context, in *struct {
			ID     uuid.UUID `path:"id"`
			RuleID uuid.UUID `path:"ruleId"`
			Body   struct {
				Weight  int32 `json:"weight" minimum:"-100" maximum:"100"`
				Enabled bool  `json:"enabled"`
			}
		}) (*setOut, error) {
			out, err := s.UpdateRule(ctx, in.ID, in.RuleID, in.Body.Weight, in.Body.Enabled)
			return &setOut{out}, err
		})
	httpx.Register(g, st("risk.rules", httpx.Op{ID: "staffRiskThresholds", Method: http.MethodPost, Path: "/api/v1/staff/risk/rule-sets/{id}/thresholds", Summary: "Change a draft's score bands"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body Thresholds
		}) (*setOut, error) {
			out, err := s.SetThresholds(ctx, in.ID, in.Body)
			return &setOut{out}, err
		})
	httpx.Register(g, st("risk.rules", httpx.Op{ID: "staffRiskBench", Method: http.MethodPost, Path: "/api/v1/staff/risk/rule-sets/{id}/test", Summary: "Replay the last 30 days of decisions against this rule set"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body BenchResult }, error) {
			out, err := s.Bench(ctx, in.ID)
			return &struct{ Body BenchResult }{out}, err
		})
	httpx.Register(g, st("risk.rules", httpx.Op{ID: "staffRiskPublish", Method: http.MethodPost, Path: "/api/v1/staff/risk/rule-sets/{id}/publish", StepUp: true, Summary: "Make a draft live (fresh MFA)"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*setOut, error) {
			out, err := s.Publish(ctx, in.ID)
			return &setOut{out}, err
		})
	httpx.Register(g, st("risk.rules", httpx.Op{ID: "staffRiskDiscard", Method: http.MethodPost, Path: "/api/v1/staff/risk/rule-sets/{id}/discard", Status: 204, Summary: "Delete a draft"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{}, error) {
			return nil, s.Discard(ctx, in.ID)
		})
}
