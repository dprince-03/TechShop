package marketing

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

// StepInput is one journey step.
//   - wait: {"hours": 2}
//   - condition: {"if": "cart_active" | "not_ordered" | "consent_sms" | "consent_email" | "consent_push"} — false exits
//   - branch: {"if": …, "goto": 3} — true jumps to that position
//   - send: {"channel": "push", "body": "…{{first_name}} {{coupon}} {{link}}…"}
//   - issue_coupon: {"couponId": "…", "expiresDays": 3, "prefix": "CART"} — puts the code in the context
//   - exit
type StepInput struct {
	Kind   string         `json:"kind" enum:"wait,condition,branch,send,issue_coupon,exit"`
	Config map[string]any `json:"config"`
}

// JourneyInput creates a journey.
type JourneyInput struct {
	Name        string      `json:"name" minLength:"3" maxLength:"120"`
	Trigger     string      `json:"trigger" enum:"signed_up,cart_abandoned,order_paid,order_delivered"`
	EntryRules  Rules       `json:"entryRules"`
	ReentryDays *int32      `json:"reentryDays,omitempty" minimum:"0" maximum:"365"`
	GoalEvent   string      `json:"goalEvent,omitempty" enum:"order_placed,"`
	Steps       []StepInput `json:"steps" minItems:"1" maxItems:"20"`
}

// CreateJourney stores a draft journey; send steps' bodies become template versions.
func (s *Service) CreateJourney(ctx context.Context, in JourneyInput, actor uuid.UUID) (store.MarketingJourney, error) {
	var j store.MarketingJourney
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		rules, _ := json.Marshal(in.EntryRules)
		var goal *string
		if in.GoalEvent != "" {
			goal = &in.GoalEvent
		}
		var err error
		j, err = tx.Q.MarketingCreateJourney(ctx, store.MarketingCreateJourneyParams{Name: in.Name, Trigger: in.Trigger, EntryRules: rules, ReentryDays: in.ReentryDays, GoalEvent: goal, CreatedBy: actor})
		if err != nil {
			return httpx.DB(err, "journey")
		}
		for i, st := range in.Steps {
			cfg := st.Config
			if cfg == nil {
				cfg = map[string]any{}
			}
			if err := validateStep(st.Kind, cfg, len(in.Steps)); err != nil {
				return err
			}
			if st.Kind == "send" {
				if body, _ := cfg["body"].(string); body != "" {
					ch, _ := cfg["channel"].(string)
					id, err := s.textTemplate(ctx, tx, "journey:"+j.ID.String()+":"+itoa(i), ch, "", body, actor)
					if err != nil {
						return err
					}
					cfg["templateVersionId"] = id.String()
				}
			}
			b, _ := json.Marshal(cfg)
			if _, err := tx.Q.MarketingAddJourneyStep(ctx, store.MarketingAddJourneyStepParams{JourneyID: j.ID, Position: int32(i), Kind: st.Kind, Config: b}); err != nil {
				return httpx.DB(err, "journey step")
			}
		}
		return tx.Audit("journey.created", "journey", j.ID.String(), map[string]any{"name": in.Name, "trigger": in.Trigger})
	})
	return j, err
}

func validateStep(kind string, cfg map[string]any, n int) error {
	bad := func(msg string) error { return httpx.Invalid("bad_step", msg) }
	switch kind {
	case "wait":
		if h, ok := cfg["hours"].(float64); !ok || h <= 0 || h > 24*60 {
			return bad("A wait step needs hours between 1 and 1440.")
		}
	case "condition", "branch":
		c, _ := cfg["if"].(string)
		if !containsStr([]string{"cart_active", "not_ordered", "consent_sms", "consent_email", "consent_push"}, c) {
			return bad("Unknown condition " + c + ".")
		}
		if kind == "branch" {
			if g, ok := cfg["goto"].(float64); !ok || int(g) < 0 || int(g) >= n {
				return bad("A branch needs goto: a step position in this journey.")
			}
		}
	case "send":
		ch, _ := cfg["channel"].(string)
		if !containsStr([]string{"sms", "email", "push"}, ch) {
			return bad("A send step needs channel sms, email or push.")
		}
		if b, _ := cfg["body"].(string); b == "" {
			return bad("A send step needs a body.")
		}
	case "issue_coupon":
		if _, err := uuid.Parse(str(cfg["couponId"])); err != nil {
			return bad("issue_coupon needs couponId.")
		}
	}
	return nil
}

// Enroll puts a person into every active journey with this trigger whose entry rules match,
// respecting the re-entry window. Runs in the caller's (relay) transaction.
func (s *Service) Enroll(ctx context.Context, tx *uow.Tx, trigger string, userID uuid.UUID, context map[string]any) error {
	js, err := tx.Q.MarketingActiveJourneysByTrigger(ctx, trigger)
	if err != nil {
		return err
	}
	for _, j := range js {
		if j.ReentryDays != nil {
			recent, err := tx.Q.MarketingRecentlyEnrolled(ctx, store.MarketingRecentlyEnrolledParams{JourneyID: j.ID, UserID: userID,
				CreatedAt: time.Now().AddDate(0, 0, -int(*j.ReentryDays))})
			if err != nil {
				return err
			}
			if recent {
				continue
			}
		}
		var rules Rules
		_ = json.Unmarshal(j.EntryRules, &rules)
		ok, err := s.matches(ctx, tx, rules, userID)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		b, _ := json.Marshal(context)
		if _, err := tx.Q.MarketingEnroll(ctx, store.MarketingEnrollParams{JourneyID: j.ID, UserID: userID, Context: b}); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}
	return nil
}

// matches checks one person against entry rules (the same compiled allowlist as segments).
func (s *Service) matches(ctx context.Context, tx *uow.Tx, r Rules, userID uuid.UUID) (bool, error) {
	q, args := r.SQL()
	args = append(args, userID)
	var ok bool
	err := tx.PgTx.QueryRow(ctx, "select exists ("+q+" and u.id = $"+itoa(len(args))+")", args...).Scan(&ok)
	return ok, err
}

// RunDue advances every enrollment whose next step is due (a worker runs this every minute;
// SKIP LOCKED lets several workers share the load).
func (s *Service) RunDue(ctx context.Context) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		due, err := tx.Q.MarketingDueEnrollments(ctx)
		if err != nil {
			return err
		}
		for _, e := range due {
			if err := s.advance(ctx, tx, e); err != nil {
				return err
			}
		}
		return nil
	})
}

// advance runs steps from the current one until a wait (reschedules) or the end.
func (s *Service) advance(ctx context.Context, tx *uow.Tx, e store.MarketingDueEnrollmentsRow) error {
	steps, err := tx.Q.MarketingJourneySteps(ctx, e.JourneyID)
	if err != nil {
		return err
	}
	var jctx map[string]any
	_ = json.Unmarshal(e.Context, &jctx)
	if jctx == nil {
		jctx = map[string]any{}
	}
	pos := int(e.CurrentStep)
	for guard := 0; guard < 50; guard++ {
		if pos >= len(steps) {
			return tx.Q.MarketingFinishEnrollment(ctx, store.MarketingFinishEnrollmentParams{ID: e.ID, Status: "completed"})
		}
		st := steps[pos]
		var cfg map[string]any
		_ = json.Unmarshal(st.Config, &cfg)
		switch st.Kind {
		case "wait":
			hours, _ := cfg["hours"].(float64)
			// The wait at pos is "done" once we've scheduled past it: resume at pos+1.
			next := time.Now().Add(time.Duration(hours * float64(time.Hour)))
			return tx.Q.MarketingAdvanceEnrollment(ctx, store.MarketingAdvanceEnrollmentParams{ID: e.ID, CurrentStep: int32(pos + 1), NextRunAt: &next})
		case "condition":
			ok, err := s.condition(ctx, tx, str(cfg["if"]), e)
			if err != nil {
				return err
			}
			if !ok {
				reason := "condition " + str(cfg["if"]) + " not met"
				return tx.Q.MarketingFinishEnrollment(ctx, store.MarketingFinishEnrollmentParams{ID: e.ID, Status: "exited", ExitReason: &reason})
			}
		case "branch":
			ok, err := s.condition(ctx, tx, str(cfg["if"]), e)
			if err != nil {
				return err
			}
			if ok {
				g, _ := cfg["goto"].(float64)
				if int(g) <= pos {
					return errors.New("journey: branch must jump forward")
				}
				pos = int(g)
				continue
			}
		case "issue_coupon":
			code, err := s.issueCode(ctx, tx, cfg, e.UserID)
			if err != nil {
				return err
			}
			jctx["coupon"] = code // kept in the context so later send steps can use it
			b, _ := json.Marshal(jctx)
			if err := tx.Q.MarketingSetEnrollmentContext(ctx, store.MarketingSetEnrollmentContextParams{ID: e.ID, Context: b}); err != nil {
				return err
			}
		case "send":
			if err := s.journeySend(ctx, tx, e, cfg, jctx, pos); err != nil {
				return err
			}
		case "exit":
			reason := "exit step"
			return tx.Q.MarketingFinishEnrollment(ctx, store.MarketingFinishEnrollmentParams{ID: e.ID, Status: "completed", ExitReason: &reason})
		}
		pos++
	}
	return errors.New("journey: too many steps in one run")
}

func (s *Service) condition(ctx context.Context, tx *uow.Tx, cond string, e store.MarketingDueEnrollmentsRow) (bool, error) {
	switch cond {
	case "cart_active":
		return tx.Q.MarketingCartStillActive(ctx, &e.UserID)
	case "not_ordered":
		ordered, err := tx.Q.MarketingOrderedSince(ctx, store.MarketingOrderedSinceParams{CustomerUserID: &e.UserID, PlacedAt: e.CreatedAt})
		return !ordered, err
	case "consent_sms", "consent_email", "consent_push":
		return tx.Q.IdentityHasConsent(ctx, store.IdentityHasConsentParams{UserID: &e.UserID, Purpose: "marketing_" + cond[len("consent_"):]})
	}
	return false, nil
}

func (s *Service) journeySend(ctx context.Context, tx *uow.Tx, e store.MarketingDueEnrollmentsRow, cfg, jctx map[string]any, pos int) error {
	u, err := tx.Q.MarketingUserContact(ctx, e.UserID)
	if err != nil {
		return nil // account gone or suspended: nothing to send
	}
	ch := str(cfg["channel"])
	to := ""
	switch ch {
	case "sms":
		to = derefStr(u.Phone)
	case "email":
		to = derefStr(u.Email)
	}
	if ch != "push" && to == "" {
		return nil
	}
	var tv *uuid.UUID
	if id, err := uuid.Parse(str(cfg["templateVersionId"])); err == nil {
		tv = &id
	}
	data := map[string]any{"firstName": u.FirstName}
	for k, v := range jctx {
		data[k] = v
	}
	eid := e.ID
	return s.Send(ctx, tx, Message{UserID: e.UserID, To: to, Channel: ch, Template: "journey", Data: data, TemplateVersionID: tv, JourneyEnrollmentID: &eid,
		Key: "journey:" + e.ID.String() + ":" + itoa(pos)})
}

// issueCode creates a unique single-use code for one person (a leaked code works once).
func (s *Service) issueCode(ctx context.Context, tx *uow.Tx, cfg map[string]any, userID uuid.UUID) (string, error) {
	couponID, err := uuid.Parse(str(cfg["couponId"]))
	if err != nil {
		return "", err
	}
	prefix := str(cfg["prefix"])
	if prefix == "" {
		prefix = "TS"
	}
	var exp *time.Time
	if d, ok := cfg["expiresDays"].(float64); ok && d > 0 {
		t := time.Now().Add(time.Duration(d * 24 * float64(time.Hour)))
		exp = &t
	}
	for attempt := 0; attempt < 5; attempt++ {
		code := prefix + "-" + randomCode(6)
		var made store.MarketingCouponCode
		err := tx.Savepoint(ctx, func(sub *uow.Tx) error {
			var err error
			made, err = sub.Q.MarketingCreateCouponCode(ctx, store.MarketingCreateCouponCodeParams{CouponID: couponID, Code: code, IssuedToUserID: &userID, ExpiresAt: exp})
			return err
		})
		if err == nil {
			return made.Code, nil
		}
	}
	return "", errors.New("journey: could not issue a unique coupon code")
}

// randomCode avoids look-alike characters (0/O, 1/I/L).
func randomCode(n int) string {
	const alphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func itoa(i int) string { return strconv.Itoa(i) }
