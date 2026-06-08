package domain

import (
	"time"

	"github.com/google/uuid"
)

type BenefitCalculationRule struct {
	ID        uuid.UUID `json:"id"`
	AgencyID  uuid.UUID `json:"agency_id"`
	ProgramID uuid.UUID `json:"program_id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
}

type BenefitCalculationVersion struct {
	ID            uuid.UUID      `json:"id"`
	RuleID        uuid.UUID      `json:"rule_id"`
	Version       int            `json:"version"`
	Formula       map[string]any `json:"formula"`
	EffectiveFrom time.Time      `json:"effective_from"`
	EffectiveTo   *time.Time     `json:"effective_to,omitempty"`
}

type BenefitCalculation struct {
	ID               uuid.UUID `json:"id"`
	CaseID           uuid.UUID `json:"case_id"`
	VersionID        uuid.UUID `json:"version_id"`
	RuleVersion      int       `json:"rule_version,omitempty"`
	CalculatedAmount float64   `json:"calculated_amount"`
	ApprovedAmount   *float64  `json:"approved_amount,omitempty"`
	CalculationTrace []any     `json:"calculation_trace,omitempty"`
	CalculatedAt     time.Time `json:"calculated_at"`
}
