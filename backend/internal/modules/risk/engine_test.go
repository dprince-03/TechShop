package risk

import "testing"

func cond(f, op string, v float64) Condition { return Condition{All: []Cond{{F: f, Op: op, V: v}}} }

func TestScoreAddsWeightsAndPicksOutcome(t *testing.T) {
	rules := []Rule{
		{Key: "new", Condition: Condition{All: []Cond{{F: "age", Op: "lt", V: 24}, {F: "total", Op: "gt", V: 500}}}, Weight: 30, Action: "score", Enabled: true},
		{Key: "cards", Condition: cond("cards", "gte", 3), Weight: 35, Action: "score", Enabled: true},
		{Key: "repeat", Condition: cond("delivered", "gte", 1), Weight: -20, Action: "score", Enabled: true},
	}
	cases := []struct {
		name    string
		f       Features
		score   int
		outcome string
	}{
		{"nothing fires", Features{"age": 100, "total": 10, "cards": 0, "delivered": 0}, 0, "allow"},
		{"one rule", Features{"age": 1, "total": 900, "cards": 0, "delivered": 0}, 30, "allow"},
		{"two rules step up", Features{"age": 1, "total": 900, "cards": 3, "delivered": 0}, 65, "step_up"},
		{"repeat customer lowers", Features{"age": 1, "total": 900, "cards": 3, "delivered": 2}, 45, "step_up"},
		{"never below zero", Features{"age": 100, "total": 1, "cards": 0, "delivered": 5}, 0, "allow"},
		{"missing feature never fires", Features{"total": 900}, 0, "allow"},
	}
	for _, c := range cases {
		r := Score(rules, DefaultThresholds, c.f)
		if r.Score != c.score || r.Outcome != c.outcome {
			t.Errorf("%s: got %d %s, want %d %s", c.name, r.Score, r.Outcome, c.score, c.outcome)
		}
	}
}

func TestScoreThresholdsAndForcedActions(t *testing.T) {
	big := []Rule{{Key: "a", Condition: cond("x", "eq", 1), Weight: 75, Action: "score", Enabled: true}}
	if r := Score(big, DefaultThresholds, Features{"x": 1}); r.Outcome != "hold" {
		t.Fatalf("75 should hold, got %s", r.Outcome)
	}
	big[0].Weight = 100
	if r := Score(big, DefaultThresholds, Features{"x": 1}); r.Outcome != "block" {
		t.Fatalf("100 should block, got %s", r.Outcome)
	}
	forced := []Rule{
		{Key: "h", Condition: cond("x", "eq", 1), Action: "hold", Enabled: true},
		{Key: "b", Condition: cond("y", "eq", 1), Action: "block", Enabled: true},
	}
	if r := Score(forced, DefaultThresholds, Features{"x": 1, "y": 0}); r.Outcome != "hold" || r.Score != 70 {
		t.Fatalf("hold rule: got %d %s", r.Score, r.Outcome)
	}
	if r := Score(forced, DefaultThresholds, Features{"x": 1, "y": 1}); r.Outcome != "block" || r.Score != 100 {
		t.Fatalf("block beats hold: got %d %s", r.Score, r.Outcome)
	}
	forced[1].Enabled = false
	if r := Score(forced, DefaultThresholds, Features{"x": 0, "y": 1}); r.Outcome != "allow" {
		t.Fatalf("disabled rule fired: %s", r.Outcome)
	}
}

func TestParseConditionRejectsUnknownOperator(t *testing.T) {
	if _, err := ParseCondition([]byte(`{"all":[{"f":"a","op":"like","v":1}]}`)); err == nil {
		t.Fatal("expected an error")
	}
	c, err := ParseCondition([]byte(`{"all":[{"f":"a","op":"gt","v":0.3}]}`))
	if err != nil || !c.Holds(Features{"a": 0.31}) || c.Holds(Features{"a": 0.3}) {
		t.Fatalf("parse or compare wrong: %v", err)
	}
}

func TestThresholdsValid(t *testing.T) {
	if !DefaultThresholds.Valid() {
		t.Fatal("defaults invalid")
	}
	for _, bad := range []Thresholds{{40, 40, 90}, {0, 70, 90}, {40, 70, 101}, {70, 40, 90}} {
		if bad.Valid() {
			t.Errorf("%v should be invalid", bad)
		}
	}
}
