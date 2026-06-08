package domain

import (
	"time"

	"github.com/google/uuid"
)

type FraudFlag struct {
	ID        uuid.UUID      `json:"id"`
	AgencyID  uuid.UUID      `json:"agency_id"`
	CaseID    uuid.UUID      `json:"case_id"`
	FlagType  string         `json:"flag_type"`
	Severity  string         `json:"severity"`
	Evidence  map[string]any `json:"evidence,omitempty"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
}

type FraudReview struct {
	ID          uuid.UUID `json:"id"`
	FraudFlagID uuid.UUID `json:"fraud_flag_id"`
	ReviewerID  uuid.UUID `json:"reviewer_id"`
	Disposition string    `json:"disposition"`
	Notes       string    `json:"notes,omitempty"`
	ReviewedAt  time.Time `json:"reviewed_at"`
}
