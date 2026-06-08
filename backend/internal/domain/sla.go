package domain

import (
	"time"

	"github.com/google/uuid"
)

type SLAPolicy struct {
	ID                  uuid.UUID `json:"id"`
	AgencyID            uuid.UUID `json:"agency_id"`
	ProgramID           uuid.UUID `json:"program_id"`
	TargetDays          int       `json:"target_days"`
	WarningThresholdPct int       `json:"warning_threshold_pct"`
	BusinessDaysOnly    bool      `json:"business_days_only"`
}

type CaseSLATracking struct {
	ID          uuid.UUID  `json:"id"`
	CaseID      uuid.UUID  `json:"case_id"`
	SLAPolicyID uuid.UUID  `json:"sla_policy_id"`
	DueAt       time.Time  `json:"due_at"`
	Status      string     `json:"status"`
	ElapsedDays int        `json:"elapsed_days"`
	BreachedAt  *time.Time `json:"breached_at,omitempty"`
}
