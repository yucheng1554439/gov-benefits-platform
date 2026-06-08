package benefit_test

import (
	"testing"

	"github.com/govbenefits/platform/internal/benefit"
)

func TestCalculator_FoodAssistance(t *testing.T) {
	c := benefit.NewCalculator()
	formula := map[string]any{
		"baseBenefit":         350.0,
		"householdMultiplier": 1.4,
		"maxBenefit":          1200.0,
	}
	result, err := c.Calculate(formula, 3, 22000)
	if err != nil {
		t.Fatal(err)
	}
	if result.Amount <= 0 || result.Amount > 1200 {
		t.Fatalf("unexpected amount: %v", result.Amount)
	}
}

func TestCalculator_EmergencyFlat(t *testing.T) {
	c := benefit.NewCalculator()
	formula := map[string]any{
		"baseBenefit": 500.0,
		"maxBenefit":  500.0,
	}
	result, err := c.Calculate(formula, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Amount != 500 {
		t.Fatalf("expected 500, got %v", result.Amount)
	}
}
