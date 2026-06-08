package domain

import (
	"time"

	"github.com/google/uuid"
)

type WorkflowTransition struct {
	ID           uuid.UUID  `json:"id"`
	AgencyID     *uuid.UUID `json:"agency_id,omitempty"`
	FromStatus   string     `json:"from_status"`
	ToStatus     string     `json:"to_status"`
	RequiredRole string     `json:"required_role"`
}

type WorkflowEvent struct {
	ID         uuid.UUID  `json:"id"`
	AgencyID   uuid.UUID  `json:"agency_id"`
	CaseID     uuid.UUID  `json:"case_id"`
	FromStatus string     `json:"from_status,omitempty"`
	ToStatus   string     `json:"to_status"`
	ActorID    *uuid.UUID `json:"actor_id,omitempty"`
	ActorName  string     `json:"actor_name,omitempty"`
	Reason     string     `json:"reason,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type StatusTransitionInput struct {
	CaseID   uuid.UUID
	ToStatus string
	ActorID  uuid.UUID
	Role     string
	Reason   string
}
