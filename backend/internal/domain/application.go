package domain

import (
	"time"

	"github.com/google/uuid"
)

type Application struct {
	ID               uuid.UUID      `json:"id"`
	AgencyID         uuid.UUID      `json:"agency_id"`
	CaseID           uuid.UUID      `json:"case_id"`
	HouseholdSize    int            `json:"household_size"`
	AnnualIncome     float64        `json:"annual_income"`
	EmploymentStatus string         `json:"employment_status,omitempty"`
	FormData         map[string]any `json:"form_data,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
}

type CreateApplicationInput struct {
	AgencyID         uuid.UUID
	CitizenID        uuid.UUID
	ProgramID        uuid.UUID
	HouseholdSize    int
	AnnualIncome     float64
	EmploymentStatus string
	FormData         map[string]any
	ZipCode          string
	CensusTract      string
	Priority         string
}
