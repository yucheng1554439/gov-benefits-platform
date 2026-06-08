package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/jackc/pgx/v5"
)

type LetterRepository struct {
	db *DB
}

func NewLetterRepository(db *DB) *LetterRepository {
	return &LetterRepository{db: db}
}

func (r *LetterRepository) GetTemplate(ctx context.Context, agencyID uuid.UUID, letterType string) (*domain.LetterTemplate, error) {
	var t domain.LetterTemplate
	var mergeFields []byte
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, agency_id, letter_type, name, body_template, merge_fields, is_active
		FROM letter_templates WHERE agency_id = $1 AND letter_type = $2 AND is_active = true LIMIT 1
	`, agencyID, letterType).Scan(&t.ID, &t.AgencyID, &t.LetterType, &t.Name, &t.BodyTemplate, &mergeFields, &t.IsActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var fields []string
	_ = json.Unmarshal(mergeFields, &fields)
	t.MergeFields = fields
	return &t, nil
}

func (r *LetterRepository) SaveGenerated(ctx context.Context, tx pgx.Tx, letter *domain.GeneratedLetter) error {
	return tx.QueryRow(ctx, `
		INSERT INTO generated_letters (agency_id, case_id, template_id, letter_type, file_key, generated_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, generated_at
	`, letter.AgencyID, letter.CaseID, letter.TemplateID, letter.LetterType, letter.FileKey, letter.GeneratedBy).
		Scan(&letter.ID, &letter.GeneratedAt)
}

func (r *LetterRepository) ListByCase(ctx context.Context, tx pgx.Tx, caseID uuid.UUID) ([]domain.GeneratedLetter, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, agency_id, case_id, template_id, letter_type, COALESCE(file_key,''), generated_by, generated_at
		FROM generated_letters WHERE case_id = $1 ORDER BY generated_at DESC
	`, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var letters []domain.GeneratedLetter
	for rows.Next() {
		var l domain.GeneratedLetter
		if err := rows.Scan(&l.ID, &l.AgencyID, &l.CaseID, &l.TemplateID, &l.LetterType, &l.FileKey, &l.GeneratedBy, &l.GeneratedAt); err != nil {
			return nil, err
		}
		letters = append(letters, l)
	}
	return letters, rows.Err()
}

func (r *LetterRepository) GetByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.GeneratedLetter, error) {
	var l domain.GeneratedLetter
	err := tx.QueryRow(ctx, `
		SELECT id, agency_id, case_id, template_id, letter_type, COALESCE(file_key,''), generated_by, generated_at
		FROM generated_letters WHERE id = $1
	`, id).Scan(&l.ID, &l.AgencyID, &l.CaseID, &l.TemplateID, &l.LetterType, &l.FileKey, &l.GeneratedBy, &l.GeneratedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}
