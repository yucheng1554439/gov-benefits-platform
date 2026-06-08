package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/jackc/pgx/v5"
)

type DocumentRepository struct {
	db *DB
}

func NewDocumentRepository(db *DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(ctx context.Context, tx pgx.Tx, doc *domain.Document) error {
	return tx.QueryRow(ctx, `
		INSERT INTO documents (agency_id, case_id, document_type_id, file_key, original_name, mime_type, file_size)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, verification_status, uploaded_at
	`, doc.AgencyID, doc.CaseID, doc.DocumentTypeID, doc.FileKey, doc.OriginalName, doc.MimeType, doc.FileSize).
		Scan(&doc.ID, &doc.VerificationStatus, &doc.UploadedAt)
}

func (r *DocumentRepository) ListByCase(ctx context.Context, tx pgx.Tx, caseID uuid.UUID) ([]domain.Document, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, agency_id, case_id, document_type_id, file_key, COALESCE(original_name,''),
		       COALESCE(mime_type,''), file_size, verification_status, reviewed_by, reviewed_at, uploaded_at
		FROM documents WHERE case_id = $1 ORDER BY uploaded_at DESC
	`, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []domain.Document
	for rows.Next() {
		var d domain.Document
		var reviewedAt *time.Time
		if err := rows.Scan(
			&d.ID, &d.AgencyID, &d.CaseID, &d.DocumentTypeID, &d.FileKey, &d.OriginalName,
			&d.MimeType, &d.FileSize, &d.VerificationStatus, &d.ReviewedBy, &reviewedAt, &d.UploadedAt,
		); err != nil {
			return nil, err
		}
		d.ReviewedAt = reviewedAt
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (r *DocumentRepository) GetByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Document, error) {
	var d domain.Document
	var reviewedAt *time.Time
	err := tx.QueryRow(ctx, `
		SELECT id, agency_id, case_id, document_type_id, file_key, COALESCE(original_name,''),
		       COALESCE(mime_type,''), file_size, verification_status, reviewed_by, reviewed_at, uploaded_at
		FROM documents WHERE id = $1
	`, id).Scan(
		&d.ID, &d.AgencyID, &d.CaseID, &d.DocumentTypeID, &d.FileKey, &d.OriginalName,
		&d.MimeType, &d.FileSize, &d.VerificationStatus, &d.ReviewedBy, &reviewedAt, &d.UploadedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	d.ReviewedAt = reviewedAt
	return &d, nil
}

func (r *DocumentRepository) UpdateVerification(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string, reviewerID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE documents SET verification_status = $2, reviewed_by = $3, reviewed_at = NOW() WHERE id = $1
	`, id, status, reviewerID)
	return err
}
