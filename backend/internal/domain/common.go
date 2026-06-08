package domain

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID `json:"id"`
	AgencyID  uuid.UUID `json:"agency_id"`
	UserID    uuid.UUID `json:"user_id"`
	Channel   string    `json:"channel"`
	EventType string    `json:"event_type"`
	Title     string    `json:"title"`
	Body      string    `json:"body,omitempty"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type AuditLog struct {
	ID            uuid.UUID      `json:"id"`
	AgencyID      *uuid.UUID     `json:"agency_id,omitempty"`
	ActorID       *uuid.UUID     `json:"actor_id,omitempty"`
	ActorName     string         `json:"actor_name,omitempty"`
	Action        string         `json:"action"`
	EntityType    string         `json:"entity_type"`
	EntityID      *uuid.UUID     `json:"entity_id,omitempty"`
	PreviousState map[string]any `json:"previous_state,omitempty"`
	NewState      map[string]any `json:"new_state,omitempty"`
	IPAddress     string         `json:"ip_address,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

type Report struct {
	ID          uuid.UUID      `json:"id"`
	AgencyID    uuid.UUID      `json:"agency_id"`
	ReportType  string         `json:"report_type"`
	Status      string         `json:"status"`
	FileKey     string         `json:"file_key,omitempty"`
	Params      map[string]any `json:"params,omitempty"`
	RequestedBy *uuid.UUID     `json:"requested_by,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

type FeatureFlag struct {
	ID         uuid.UUID      `json:"id"`
	AgencyID   uuid.UUID      `json:"agency_id"`
	FlagKey    string         `json:"flag_key"`
	IsEnabled  bool           `json:"is_enabled"`
	RolloutPct int            `json:"rollout_pct"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type RetentionPolicy struct {
	ID                uuid.UUID `json:"id"`
	AgencyID          uuid.UUID `json:"agency_id"`
	EntityType        string    `json:"entity_type"`
	RetentionYears    int       `json:"retention_years"`
	DispositionAction string    `json:"disposition_action"`
	IsActive          bool      `json:"is_active"`
}
