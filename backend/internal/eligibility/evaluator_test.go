package eligibility_test

import (
	"testing"

	"github.com/govbenefits/platform/internal/eligibility"
)

func TestEvaluator_ANDRules(t *testing.T) {
	e := eligibility.NewEvaluator()
	conditions := map[string]any{
		"operator": "AND",
		"rules": []any{
			map[string]any{"field": "annual_income", "op": "lt", "value": 35000.0},
			map[string]any{"field": "household_size", "op": "gte", "value": 1.0},
		},
	}
	data := map[string]any{
		"annual_income":  28000.0,
		"household_size": 3.0,
	}
	result, err := e.Evaluate(conditions, data)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Eligible {
		t.Fatal("expected eligible")
	}
}

func TestEvaluator_FailsIncome(t *testing.T) {
	e := eligibility.NewEvaluator()
	conditions := map[string]any{
		"operator": "AND",
		"rules": []any{
			map[string]any{"field": "annual_income", "op": "lt", "value": 35000.0},
		},
	}
	data := map[string]any{"annual_income": 50000.0}
	result, err := e.Evaluate(conditions, data)
	if err != nil {
		t.Fatal(err)
	}
	if result.Eligible {
		t.Fatal("expected not eligible")
	}
}
