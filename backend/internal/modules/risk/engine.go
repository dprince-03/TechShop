package risk

import (
	"encoding/json"
	"fmt"
	"sort"
)

// The rule engine is pure (no database), so rule sets can be scored live, replayed on the test
// bench and unit-tested the same way. Plan: trust-safety.md §3.

// Features are the numeric facts a checkpoint gives the engine (booleans are 0 or 1).
type Features map[string]float64

// Cond is one comparison: feature, operator, value.
type Cond struct {
	F  string  `json:"f"`
	Op string  `json:"op"`
	V  float64 `json:"v"`
}

// Condition holds when every comparison holds.
type Condition struct {
	All []Cond `json:"all"`
}

// Rule is a scored or deterministic rule.
type Rule struct {
	Key       string
	Label     string
	Condition Condition
	Weight    int
	Action    string // score, hold, block
	Enabled   bool
}

// Thresholds turn a score into an outcome.
type Thresholds struct {
	Step  int `json:"step"`
	Hold  int `json:"hold"`
	Block int `json:"block"`
}

// DefaultThresholds are trust-safety.md §3: < 40 allow, 40–69 step up, 70–89 hold, ≥ 90 block.
var DefaultThresholds = Thresholds{Step: 40, Hold: 70, Block: 90}

// FiredRule is a rule that matched, as stored in the decision's reasons.
type FiredRule struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Weight int    `json:"weight"`
	Action string `json:"action"`
}

// Result is the engine's verdict.
type Result struct {
	Score   int
	Outcome string // allow, step_up, hold, block
	Fired   []FiredRule
}

// ParseCondition reads a stored condition.
func ParseCondition(raw []byte) (Condition, error) {
	var c Condition
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, err
	}
	for _, x := range c.All {
		if !validOp(x.Op) {
			return c, fmt.Errorf("unknown operator %q", x.Op)
		}
	}
	return c, nil
}

func validOp(op string) bool {
	switch op {
	case "lt", "lte", "gt", "gte", "eq":
		return true
	}
	return false
}

// Holds reports whether a condition matches. A feature the checkpoint didn't supply never
// matches, so a rule can't fire on missing data.
func (c Condition) Holds(f Features) bool {
	if len(c.All) == 0 {
		return false
	}
	for _, x := range c.All {
		v, ok := f[x.F]
		if !ok || !compare(v, x.Op, x.V) {
			return false
		}
	}
	return true
}

func compare(a float64, op string, b float64) bool {
	switch op {
	case "lt":
		return a < b
	case "lte":
		return a <= b
	case "gt":
		return a > b
	case "gte":
		return a >= b
	case "eq":
		return a == b
	}
	return false
}

// Score runs the rules over the features. Weights add up (clamped to 0–100); a matching hold or
// block rule sets that outcome whatever the score, and block beats hold.
func Score(rules []Rule, t Thresholds, f Features) Result {
	var res Result
	sum, forced := 0, ""
	for _, r := range rules {
		if !r.Enabled || !r.Condition.Holds(f) {
			continue
		}
		res.Fired = append(res.Fired, FiredRule{Key: r.Key, Label: r.Label, Weight: r.Weight, Action: r.Action})
		switch r.Action {
		case "block":
			forced = "block"
		case "hold":
			if forced != "block" {
				forced = "hold"
			}
		default:
			sum += r.Weight
		}
	}
	res.Score = min(max(sum, 0), 100)
	switch {
	case forced == "block" || res.Score >= t.Block:
		res.Outcome = "block"
		if forced == "block" {
			res.Score = 100
		}
	case forced == "hold" || res.Score >= t.Hold:
		res.Outcome = "hold"
		res.Score = max(res.Score, t.Hold)
	case res.Score >= t.Step:
		res.Outcome = "step_up"
	default:
		res.Outcome = "allow"
	}
	sort.SliceStable(res.Fired, func(i, j int) bool { return res.Fired[i].Weight > res.Fired[j].Weight })
	return res
}

// Valid reports whether thresholds are in order and inside 1–100.
func (t Thresholds) Valid() bool {
	return t.Step >= 1 && t.Step < t.Hold && t.Hold < t.Block && t.Block <= 100
}

// LuhnOK checks a 15-digit IMEI's check digit (the Luhn algorithm, as GSMA defines for IMEIs).
func LuhnOK(imei string) bool {
	if len(imei) != 15 {
		return false
	}
	sum := 0
	for i := range 15 {
		c := imei[14-i]
		if c < '0' || c > '9' {
			return false
		}
		d := int(c - '0')
		if i%2 == 1 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return sum%10 == 0
}
