package domain

import (
	"time"

	"github.com/google/uuid"
)

type Appeal struct {
	ID          uuid.UUID  `json:"id"`
	AgencyID    uuid.UUID  `json:"agency_id"`
	CaseID      uuid.UUID  `json:"case_id"`
	CitizenID   uuid.UUID  `json:"citizen_id"`
	Status      string     `json:"status"`
	Grounds     string     `json:"grounds"`
	FiledAt     time.Time  `json:"filed_at"`
	HearingDate *time.Time `json:"hearing_date,omitempty"`
	CaseNumber  string     `json:"case_number,omitempty"`
	ProgramName string     `json:"program_name,omitempty"`
	CitizenName string     `json:"citizen_name,omitempty"`
	CaseStatus  string     `json:"case_status,omitempty"`
}

type AppealDecision struct {
	ID         uuid.UUID `json:"id"`
	AppealID   uuid.UUID `json:"appeal_id"`
	ReviewerID uuid.UUID `json:"reviewer_id"`
	Decision   string    `json:"decision"`
	Rationale  string    `json:"rationale,omitempty"`
	DecidedAt  time.Time `json:"decided_at"`
}
