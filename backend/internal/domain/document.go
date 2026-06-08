package domain

import (
	"time"

	"github.com/google/uuid"
)

type DocumentType struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`
}

type Document struct {
	ID                 uuid.UUID  `json:"id"`
	AgencyID           uuid.UUID  `json:"agency_id"`
	CaseID             uuid.UUID  `json:"case_id"`
	DocumentTypeID     *uuid.UUID `json:"document_type_id,omitempty"`
	FileKey            string     `json:"file_key"`
	OriginalName       string     `json:"original_name,omitempty"`
	MimeType           string     `json:"mime_type,omitempty"`
	FileSize           int64      `json:"file_size"`
	VerificationStatus string     `json:"verification_status"`
	ReviewedBy         *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt         *time.Time `json:"reviewed_at,omitempty"`
	UploadedAt         time.Time  `json:"uploaded_at"`
}
