package domain

import (
	"time"

	"github.com/google/uuid"
)

type Agency struct {
	ID           uuid.UUID `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Type         string    `json:"type,omitempty"`
	Jurisdiction string    `json:"jurisdiction,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type Program struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
}

type AgencyProgram struct {
	ID        uuid.UUID `json:"id"`
	AgencyID  uuid.UUID `json:"agency_id"`
	ProgramID uuid.UUID `json:"program_id"`
	IsEnabled bool      `json:"is_enabled"`
	Program   *Program  `json:"program,omitempty"`
}
