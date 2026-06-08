package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserProfile struct {
	ID        uuid.UUID      `json:"id"`
	UserID    uuid.UUID      `json:"user_id"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Phone     string         `json:"phone,omitempty"`
	SSNHash   string         `json:"-"`
	Address   map[string]any `json:"address,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
}

type AgencyUser struct {
	ID         uuid.UUID `json:"id"`
	AgencyID   uuid.UUID `json:"agency_id"`
	UserID     uuid.UUID `json:"user_id"`
	AgencyRole string    `json:"agency_role"`
	IsPrimary  bool      `json:"is_primary"`
}

type WorkerProfile struct {
	UserID           uuid.UUID `json:"user_id"`
	AgencyID         uuid.UUID `json:"agency_id"`
	Specializations  []string  `json:"specializations"`
	MaxActiveCases   int       `json:"max_active_cases"`
	CurrentCaseCount int       `json:"current_case_count"`
}

type AuthUser struct {
	User       User        `json:"user"`
	Profile    *UserProfile `json:"profile,omitempty"`
	Roles      []string    `json:"roles"`
	AgencyID   uuid.UUID   `json:"agency_id"`
	AgencyRole string      `json:"agency_role"`
}
