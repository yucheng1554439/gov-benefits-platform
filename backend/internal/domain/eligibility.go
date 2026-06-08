package domain

import (
	"time"

	"github.com/google/uuid"
)

type EligibilityRule struct {
	ID        uuid.UUID `json:"id"`
	AgencyID  uuid.UUID `json:"agency_id"`
	ProgramID uuid.UUID `json:"program_id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
}

type EligibilityRuleVersion struct {
	ID            uuid.UUID      `json:"id"`
	RuleID        uuid.UUID      `json:"rule_id"`
	Version       int            `json:"version"`
	Conditions    map[string]any `json:"conditions"`
	Actions       map[string]any `json:"actions,omitempty"`
	EffectiveFrom time.Time      `json:"effective_from"`
	EffectiveTo   *time.Time     `json:"effective_to,omitempty"`
}

type EligibilityEvaluation struct {
	ID              uuid.UUID `json:"id"`
	CaseID          uuid.UUID `json:"case_id"`
	VersionID       uuid.UUID `json:"version_id"`
	IsEligible      bool      `json:"is_eligible"`
	EvaluationTrace []any     `json:"evaluation_trace,omitempty"`
	EvaluatedAt     time.Time `json:"evaluated_at"`
}

type EligibilityRuleDetail struct {
	EligibilityRule
	ProgramName   string     `json:"program_name"`
	ProgramCode   string     `json:"program_code"`
	Version       int        `json:"version"`
	EffectiveFrom time.Time  `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`
	Conditions    map[string]any `json:"conditions,omitempty"`
}
