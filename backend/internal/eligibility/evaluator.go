package eligibility

import (
	"fmt"
	"strings"
)

type Evaluator struct{}

func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

type EvaluationResult struct {
	Eligible bool
	Trace    []map[string]any
}

func (e *Evaluator) Evaluate(conditions map[string]any, data map[string]any) (*EvaluationResult, error) {
	result, trace, err := e.evalNode(conditions, data)
	if err != nil {
		return nil, err
	}
	return &EvaluationResult{Eligible: result, Trace: trace}, nil
}

func (e *Evaluator) evalNode(node map[string]any, data map[string]any) (bool, []map[string]any, error) {
	if op, ok := node["operator"].(string); ok {
		return e.evalLogical(op, node, data)
	}
	if field, ok := node["field"].(string); ok {
		pass, err := e.evalRule(field, node, data)
		trace := []map[string]any{{
			"field":  field,
			"op":     node["op"],
			"value":  node["value"],
			"actual": data[field],
			"pass":   pass,
		}}
		return pass, trace, err
	}
	if rules, ok := node["rules"].([]any); ok && len(rules) > 0 {
		return e.evalLogical("AND", node, data)
	}
	return false, nil, fmt.Errorf("invalid condition node")
}

func (e *Evaluator) evalLogical(operator string, node map[string]any, data map[string]any) (bool, []map[string]any, error) {
	rules, _ := node["rules"].([]any)
	if len(rules) == 0 {
		return true, nil, nil
	}

	var allTrace []map[string]any
	op := strings.ToUpper(operator)
	if opVal, ok := node["operator"].(string); ok {
		op = strings.ToUpper(opVal)
	}

	switch op {
	case "AND":
		for _, r := range rules {
			rule, ok := r.(map[string]any)
			if !ok {
				return false, allTrace, fmt.Errorf("invalid rule in AND")
			}
			pass, trace, err := e.evalNode(rule, data)
			allTrace = append(allTrace, trace...)
			if err != nil {
				return false, allTrace, err
			}
			if !pass {
				return false, allTrace, nil
			}
		}
		return true, allTrace, nil
	case "OR":
		for _, r := range rules {
			rule, ok := r.(map[string]any)
			if !ok {
				return false, allTrace, fmt.Errorf("invalid rule in OR")
			}
			pass, trace, err := e.evalNode(rule, data)
			allTrace = append(allTrace, trace...)
			if err != nil {
				return false, allTrace, err
			}
			if pass {
				return true, allTrace, nil
			}
		}
		return false, allTrace, nil
	default:
		return false, allTrace, fmt.Errorf("unknown operator: %s", op)
	}
}

func (e *Evaluator) evalRule(field string, node map[string]any, data map[string]any) (bool, error) {
	op, _ := node["op"].(string)
	expected := node["value"]
	actual := data[field]

	switch op {
	case "eq":
		return compareEqual(actual, expected), nil
	case "neq":
		return !compareEqual(actual, expected), nil
	case "lt":
		return toFloat(actual) < toFloat(expected), nil
	case "lte":
		return toFloat(actual) <= toFloat(expected), nil
	case "gt":
		return toFloat(actual) > toFloat(expected), nil
	case "gte":
		return toFloat(actual) >= toFloat(expected), nil
	default:
		return false, fmt.Errorf("unknown op: %s", op)
	}
}

func compareEqual(a, b any) bool {
	return toFloat(a) == toFloat(b)
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
	case int32:
		return float64(n)
	default:
		return 0
	}
}
