// Package risk is the trust-and-safety engine: versioned rule sets scored at checkpoints
// (checkout, payout, trade-in, return), every decision logged with the features it saw, one case
// queue for reviews, and the stolen-device blocklist. Plan: docs/trust-safety.md §3–5.
//
// Other modules call Evaluate inside their own transaction and act on the outcome. When an
// analyst clears or blocks a case, "risk.case_resolved" tells the owning module to release or
// stop what it held.
package risk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/jobs"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/outbox"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/pkg/validate"
)

// Service is the risk module.
type Service struct{ d *kit.Deps }

// New creates the module.
func New(d *kit.Deps) *Service { return &Service{d: d} }

// Name implements kit.Module.
func (s *Service) Name() string { return "risk" }

// Jobs implements kit.Module.
func (s *Service) Jobs(*jobs.Registry) {}

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}

// fallbackHoldFrom is the amount from which an engine failure holds instead of allowing
// (trust-safety.md §3: allow + flag for low value, hold for high value; docs/plan.md 2026-10-23).
const fallbackHoldFrom = 20_000_000 // ₦200,000

// Check is what a checkpoint asks the engine.
type Check struct {
	Checkpoint  string
	SubjectType string // order, user, seller, trade_in, return, payout
	SubjectID   uuid.UUID
	UserID      *uuid.UUID
	Features    Features
	MoneyKobo   int64
	CaseKind    string // the case kind if a hold or block opens one
	CaseTitle   string // e.g. "Order TS-10482"
	NoCase      bool   // the caller handles non-allow outcomes itself (returns wait for support)
}

// Decision is the engine's answer.
type Decision struct {
	ID      int64
	Score   int
	Outcome string // allow, step_up, hold, block, fallback_allow, fallback_hold
	Fired   []FiredRule
	CaseID  *uuid.UUID
	CaseRef string
}

// Held reports whether the checkpoint must stop and wait for review.
func (d Decision) Held() bool {
	return d.Outcome == "hold" || d.Outcome == "fallback_hold"
}

// reasons is the stored shape of risk_decisions.reasons: what fired and what the engine saw.
type reasons struct {
	Fired    []FiredRule `json:"fired"`
	Features Features    `json:"features"`
	Error    string      `json:"error,omitempty"`
}

// Evaluate scores a checkpoint with its active rule set, logs the decision and opens a case for a
// hold or block. If the rule set can't be loaded or read, it falls back by amount and says so.
func (s *Service) Evaluate(ctx context.Context, tx *uow.Tx, c Check) (Decision, error) {
	start := time.Now()
	var res Result
	var version *int32
	var engineErr error
	if err := tx.Savepoint(ctx, func(sp *uow.Tx) error {
		set, err := sp.Q.RiskActiveRuleSet(ctx, c.Checkpoint)
		if err != nil {
			return err
		}
		rules, t, err := loadRules(ctx, sp.Q, set)
		if err != nil {
			return err
		}
		v := set.Version
		version = &v
		res = Score(rules, t, c.Features)
		return nil
	}); err != nil {
		engineErr = err
		res = Result{Outcome: "fallback_allow"}
		if c.MoneyKobo >= fallbackHoldFrom {
			res.Outcome = "fallback_hold"
		}
		s.d.Logger.Error("risk engine fell back", "checkpoint", c.Checkpoint, "subject", c.SubjectID, "err", err)
	}
	why := reasons{Fired: res.Fired, Features: c.Features}
	if why.Fired == nil {
		why.Fired = []FiredRule{}
	}
	if engineErr != nil {
		why.Error = "rule set unavailable"
	}
	raw, err := json.Marshal(why)
	if err != nil {
		return Decision{}, err
	}
	latency := int32(time.Since(start).Milliseconds())
	id, err := tx.Q.RiskInsertDecision(ctx, store.RiskInsertDecisionParams{Checkpoint: c.Checkpoint, SubjectType: c.SubjectType, SubjectID: c.SubjectID,
		UserID: c.UserID, Score: int16(res.Score), Outcome: res.Outcome, Reasons: raw, RuleSetVersion: version, LatencyMs: &latency})
	if err != nil {
		return Decision{}, err
	}
	d := Decision{ID: id, Score: res.Score, Outcome: res.Outcome, Fired: res.Fired}
	if c.NoCase || (res.Outcome != "hold" && res.Outcome != "block" && res.Outcome != "fallback_hold") {
		return d, nil
	}
	cs, err := s.OpenCase(ctx, tx, NewCase{Kind: c.CaseKind, SubjectType: c.SubjectType, SubjectID: c.SubjectID, Reason: caseReason(c.CaseTitle, res),
		Score: res.Score, MoneyKobo: c.MoneyKobo, DecisionID: &id, Block: res.Outcome == "block"})
	if err != nil {
		return Decision{}, err
	}
	d.CaseID, d.CaseRef = &cs.ID, cs.Reference
	return d, nil
}

func caseReason(title string, res Result) string {
	labels := make([]string, 0, len(res.Fired))
	for _, f := range res.Fired {
		labels = append(labels, f.Label)
	}
	why := strings.Join(labels, "; ")
	if why == "" {
		why = "Risk engine unavailable; held for a manual check"
	}
	if title == "" {
		return why
	}
	return title + ": " + why
}

func loadRules(ctx context.Context, q *store.Queries, set store.RiskRiskRuleSet) ([]Rule, Thresholds, error) {
	var t Thresholds
	if err := json.Unmarshal(set.Thresholds, &t); err != nil || !t.Valid() {
		return nil, t, fmt.Errorf("rule set %s v%d: bad thresholds", set.Checkpoint, set.Version)
	}
	rows, err := q.RiskRules(ctx, set.ID)
	if err != nil {
		return nil, t, err
	}
	rules := make([]Rule, 0, len(rows))
	for _, r := range rows {
		cond, err := ParseCondition(r.Condition)
		if err != nil {
			return nil, t, fmt.Errorf("rule %s: %w", r.Key, err)
		}
		rules = append(rules, Rule{Key: r.Key, Label: r.Label, Condition: cond, Weight: int(r.Weight), Action: r.Action, Enabled: r.Enabled})
	}
	return rules, t, nil
}

// NewCase opens a case directly (from the engine, or a module that found fraud itself).
type NewCase struct {
	Kind        string
	SubjectType string
	SubjectID   uuid.UUID
	Reason      string
	Score       int
	MoneyKobo   int64
	DecisionID  *int64
	Block       bool // blocks get a day; holds have a customer waiting, so 2 hours
}

// OpenCase adds a case to the queue with its SLA.
func (s *Service) OpenCase(ctx context.Context, tx *uow.Tx, n NewCase) (store.RiskRiskCase, error) {
	due := time.Now().Add(2 * time.Hour)
	if n.Block {
		due = time.Now().Add(24 * time.Hour)
	}
	score, money := int16(n.Score), n.MoneyKobo
	cs, err := tx.Q.RiskOpenCase(ctx, store.RiskOpenCaseParams{Kind: n.Kind, SubjectType: n.SubjectType, SubjectID: n.SubjectID, Reason: n.Reason,
		Score: &score, MoneyAtRiskKobo: &money, DecisionID: n.DecisionID, SlaDueAt: &due})
	if err != nil {
		return cs, err
	}
	if err := tx.Q.RiskCaseEvent(ctx, store.RiskCaseEventParams{CaseID: cs.ID, Action: "opened", Note: &n.Reason}); err != nil {
		return cs, err
	}
	return cs, nil
}

// OpenCases counts unresolved cases about a subject (a feature for many rules).
func (s *Service) OpenCases(ctx context.Context, q *store.Queries, subjectType string, id uuid.UUID) (int, error) {
	n, err := q.RiskOpenCasesForSubject(ctx, store.RiskOpenCasesForSubjectParams{SubjectType: subjectType, SubjectID: id})
	return int(n), err
}

// Claim takes a case for review.
func (s *Service) Claim(ctx context.Context, ref string) (RiskCase, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	return s.onCase(ctx, ref, func(tx *uow.Tx, c store.RiskRiskCase) error {
		// In review can be taken over (an analyst away, or a manager stepping in).
		if c.Status != "open" && c.Status != "escalated" && c.Status != "in_review" {
			return httpx.Conflict("bad_transition", "This case is already "+humanise(c.Status)+".")
		}
		if _, err := tx.Q.RiskClaimCase(ctx, store.RiskClaimCaseParams{ID: c.ID, AssignedTo: &actor}); err != nil {
			return err
		}
		return s.event(ctx, tx, c.ID, actor, "claimed", "")
	})
}

// Note adds a note to a case's timeline.
func (s *Service) Note(ctx context.Context, ref, note string) (RiskCase, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	if strings.TrimSpace(note) == "" {
		return RiskCase{}, httpx.Invalid("note_required", "Write a note.")
	}
	return s.onCase(ctx, ref, func(tx *uow.Tx, c store.RiskRiskCase) error {
		return s.event(ctx, tx, c.ID, actor, "note", note)
	})
}

// Escalate sends a case to a manager.
func (s *Service) Escalate(ctx context.Context, ref, note string) (RiskCase, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	if strings.TrimSpace(note) == "" {
		return RiskCase{}, httpx.Invalid("note_required", "Say why it needs a manager.")
	}
	return s.onCase(ctx, ref, func(tx *uow.Tx, c store.RiskRiskCase) error {
		if c.Status != "open" && c.Status != "in_review" {
			return httpx.Conflict("bad_transition", "Only open cases can be escalated.")
		}
		if _, err := tx.Q.RiskEscalateCase(ctx, c.ID); err != nil {
			return err
		}
		return s.event(ctx, tx, c.ID, actor, "escalated", note)
	})
}

// Resolve clears or blocks a case. The label feeds rule tuning; the owning module is told so it
// can release or stop what it held.
func (s *Service) Resolve(ctx context.Context, ref, decision, label, note string) (RiskCase, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	if strings.TrimSpace(note) == "" {
		return RiskCase{}, httpx.Invalid("note_required", "Record why, for the audit trail.")
	}
	status, action := "cleared", "cleared"
	switch {
	case decision == "block":
		status, action, label = "blocked", "blocked", "fraud"
	case label != "legit" && label != "unclear":
		label = "legit"
	}
	return s.onCase(ctx, ref, func(tx *uow.Tx, c store.RiskRiskCase) error {
		if c.Status == "cleared" || c.Status == "blocked" {
			return httpx.Conflict("bad_transition", "This case is already "+c.Status+".")
		}
		if _, err := tx.Q.RiskResolveCase(ctx, store.RiskResolveCaseParams{ID: c.ID, Status: status, Label: &label, ResolvedBy: &actor, ResolutionNote: &note}); err != nil {
			return err
		}
		if err := s.event(ctx, tx, c.ID, actor, action, note); err != nil {
			return err
		}
		tx.Emit("risk_case", c.ID, "risk.case_resolved", map[string]any{"caseId": c.ID, "reference": c.Reference, "kind": c.Kind,
			"subjectType": c.SubjectType, "subjectId": c.SubjectID, "outcome": status})
		return tx.Audit("risk.case_"+status, "risk_case", c.ID.String(), map[string]string{"label": label})
	})
}

func (s *Service) event(ctx context.Context, tx *uow.Tx, caseID, actor uuid.UUID, action, note string) error {
	var n *string
	if strings.TrimSpace(note) != "" {
		n = &note
	}
	return tx.Q.RiskCaseEvent(ctx, store.RiskCaseEventParams{CaseID: caseID, ActorID: &actor, Action: action, Note: n})
}

func (s *Service) onCase(ctx context.Context, ref string, fn func(*uow.Tx, store.RiskRiskCase) error) (RiskCase, error) {
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		c, err := tx.Q.RiskCaseForUpdate(ctx, ref)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.NotFound("Case")
		}
		if err != nil {
			return err
		}
		return fn(tx, c)
	})
	if err != nil {
		return RiskCase{}, err
	}
	return s.CaseView(ctx, ref)
}

// IMEIStatus is what intake desks need to know about a device.
type IMEIStatus struct {
	IMEI    string  `json:"imei"`
	LuhnOK  bool    `json:"luhnOk" doc:"The check digit is right (a typo or a fake IMEI fails)"`
	Blocked bool    `json:"blocked"`
	Reason  *string `json:"reason" doc:"Why it's blocked: reported_stolen, fraud or counterfeit"`
}

// CheckIMEI checks a device against the Luhn rule and the blocklist.
func CheckIMEI(ctx context.Context, q *store.Queries, imei string) (IMEIStatus, error) {
	imei = strings.ReplaceAll(strings.TrimSpace(imei), " ", "")
	st := IMEIStatus{IMEI: imei, LuhnOK: validate.IMEI(imei)}
	reason, err := q.RiskDeviceBlocked(ctx, imei)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return st, nil
	case err != nil:
		return st, err
	}
	st.Blocked, st.Reason = true, &reason
	return st, nil
}

// BlockDeviceInput adds an IMEI to the blocklist.
type BlockDeviceInput struct {
	IMEI      string `json:"imei" pattern:"^[0-9]{15}$"`
	Reason    string `json:"reason" enum:"reported_stolen,fraud,counterfeit"`
	Source    string `json:"source" minLength:"2" maxLength:"80" doc:"Who reported it, e.g. customer report, police extract, case RK-301"`
	Reference string `json:"reference,omitempty" maxLength:"80" doc:"Police extract number, case or ticket"`
}

// Block adds a device to the blocklist. Only Luhn-valid IMEIs are accepted, so a typo can't
// block somebody else's phone.
func (s *Service) Block(ctx context.Context, in BlockDeviceInput) (BlockedDevice, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	if !validate.IMEI(in.IMEI) {
		return BlockedDevice{}, httpx.Invalid("imei_invalid", "That IMEI fails the check digit. Check it against *#06# or the box.")
	}
	var out BlockedDevice
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		row, err := tx.Q.RiskBlockDevice(ctx, store.RiskBlockDeviceParams{Imei: in.IMEI, Reason: in.Reason, Source: in.Source, Reference: optional(in.Reference), CreatedBy: &actor})
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.Conflict("already_blocked", "That IMEI is already on the blocklist.")
		}
		if err != nil {
			return err
		}
		out = BlockedDevice{IMEI: row.Imei, Reason: row.Reason, Source: row.Source, Reference: row.Reference, CreatedAt: row.CreatedAt}
		return tx.Audit("device.blocked", "device", in.IMEI, map[string]string{"reason": in.Reason, "source": in.Source})
	})
	return out, err
}

// Unblock removes a device from the blocklist (a recovered phone, or a mistaken report).
func (s *Service) Unblock(ctx context.Context, imei, note string) error {
	if strings.TrimSpace(note) == "" {
		return httpx.Invalid("note_required", "Say why it's being removed.")
	}
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if _, err := tx.Q.RiskUnblockDevice(ctx, imei); errors.Is(err, pgx.ErrNoRows) {
			return httpx.NotFound("Blocked device")
		} else if err != nil {
			return err
		}
		return tx.Audit("device.unblocked", "device", imei, map[string]string{"note": note})
	})
}

func optional(s string) *string {
	if s = strings.TrimSpace(s); s == "" {
		return nil
	}
	return &s
}

func humanise(s string) string { return strings.ReplaceAll(s, "_", " ") }
