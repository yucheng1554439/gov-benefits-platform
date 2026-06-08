package benefit

import (
	"fmt"
)

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

type CalculationResult struct {
	Amount float64
	Trace  []map[string]any
}

func (c *Calculator) Calculate(formula map[string]any, householdSize int, annualIncome float64) (*CalculationResult, error) {
	baseBenefit := toFloat(formula["baseBenefit"])
	multiplier := toFloat(formula["householdMultiplier"])
	maxBenefit := toFloat(formula["maxBenefit"])

	if baseBenefit == 0 {
		baseBenefit = 100
	}
	if multiplier == 0 {
		multiplier = 1
	}
	if maxBenefit == 0 {
		maxBenefit = 999999
	}

	raw := baseBenefit * (1 + float64(householdSize-1)*multiplier/10)
	incomeReduction := annualIncome * 0.01
	amount := raw - incomeReduction
	if amount < 0 {
		amount = 0
	}
	if amount > maxBenefit {
		amount = maxBenefit
	}

	trace := []map[string]any{
		{"step": "base_benefit", "value": baseBenefit},
		{"step": "household_size", "value": householdSize},
		{"step": "raw_amount", "value": raw},
		{"step": "income_reduction", "value": incomeReduction},
		{"step": "final_amount", "value": amount},
	}

	return &CalculationResult{Amount: amount, Trace: trace}, nil
}

func (c *Calculator) CalculateFromJSON(formula map[string]any, inputs map[string]any) (*CalculationResult, error) {
	householdSize := int(toFloat(inputs["household_size"]))
	annualIncome := toFloat(inputs["annual_income"])
	if householdSize <= 0 {
		return nil, fmt.Errorf("invalid household size")
	}
	return c.Calculate(formula, householdSize, annualIncome)
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}
