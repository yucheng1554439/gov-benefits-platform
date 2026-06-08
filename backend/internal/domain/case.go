package domain

import (
	"time"

	"github.com/google/uuid"
)

type Case struct {
	ID           uuid.UUID  `json:"id"`
	AgencyID     uuid.UUID  `json:"agency_id"`
	CaseNumber   string     `json:"case_number"`
	CitizenID    uuid.UUID  `json:"citizen_id"`
	ProgramID    uuid.UUID  `json:"program_id"`
	Status       string     `json:"status"`
	Priority     string     `json:"priority"`
	ZipCode      string     `json:"zip_code,omitempty"`
	CensusTract  string     `json:"census_tract,omitempty"`
	SubmittedAt  time.Time  `json:"submitted_at"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Program      *Program   `json:"program,omitempty"`
	Application  *Application `json:"application,omitempty"`
}

type CaseAssignment struct {
	ID         uuid.UUID `json:"id"`
	CaseID     uuid.UUID `json:"case_id"`
	WorkerID   uuid.UUID `json:"worker_id"`
	IsActive   bool      `json:"is_active"`
	AssignedAt time.Time `json:"assigned_at"`
}

type CaseNote struct {
	ID        uuid.UUID `json:"id"`
	CaseID    uuid.UUID `json:"case_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type CaseListFilter struct {
	AgencyID  uuid.UUID
	CitizenID *uuid.UUID
	Status    string
	Limit     int
	Offset    int
}
