package domain

import (
	"time"

	"github.com/google/uuid"
)

type LetterTemplate struct {
	ID          uuid.UUID `json:"id"`
	AgencyID    uuid.UUID `json:"agency_id"`
	LetterType  string    `json:"letter_type"`
	Name        string    `json:"name"`
	BodyTemplate string   `json:"body_template"`
	MergeFields []string  `json:"merge_fields,omitempty"`
	IsActive    bool      `json:"is_active"`
}

type GeneratedLetter struct {
	ID          uuid.UUID  `json:"id"`
	AgencyID    uuid.UUID  `json:"agency_id"`
	CaseID      uuid.UUID  `json:"case_id"`
	TemplateID  *uuid.UUID `json:"template_id,omitempty"`
	LetterType  string     `json:"letter_type"`
	FileKey     string     `json:"file_key,omitempty"`
	GeneratedBy *uuid.UUID `json:"generated_by,omitempty"`
	GeneratedAt time.Time  `json:"generated_at"`
}
