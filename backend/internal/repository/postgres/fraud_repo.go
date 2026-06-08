package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/jackc/pgx/v5"
)

type FraudRepository struct {
	db *DB
}

func NewFraudRepository(db *DB) *FraudRepository {
	return &FraudRepository{db: db}
}

func (r *FraudRepository) CreateFlag(ctx context.Context, tx pgx.Tx, flag *domain.FraudFlag) error {
	return tx.QueryRow(ctx, `
		INSERT INTO fraud_flags (agency_id, case_id, flag_type, severity, evidence, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, flag.AgencyID, flag.CaseID, flag.FlagType, flag.Severity, encodeJSON(flag.Evidence), flag.Status).
		Scan(&flag.ID, &flag.CreatedAt)
}

func (r *FraudRepository) ListByCase(ctx context.Context, tx pgx.Tx, caseID uuid.UUID) ([]domain.FraudFlag, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, agency_id, case_id, flag_type, severity, evidence, status, created_at
		FROM fraud_flags WHERE case_id = $1 ORDER BY created_at DESC
	`, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []domain.FraudFlag
	for rows.Next() {
		var f domain.FraudFlag
		var evidence []byte
		if err := rows.Scan(&f.ID, &f.AgencyID, &f.CaseID, &f.FlagType, &f.Severity, &evidence, &f.Status, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.Evidence = decodeJSONMap(evidence)
		flags = append(flags, f)
	}
	return flags, rows.Err()
}

func (r *FraudRepository) CreateReview(ctx context.Context, tx pgx.Tx, review *domain.FraudReview) error {
	return tx.QueryRow(ctx, `
		INSERT INTO fraud_reviews (fraud_flag_id, reviewer_id, disposition, notes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, reviewed_at
	`, review.FraudFlagID, review.ReviewerID, review.Disposition, review.Notes).
		Scan(&review.ID, &review.ReviewedAt)
}

func (r *FraudRepository) UpdateFlagStatus(ctx context.Context, tx pgx.Tx, flagID uuid.UUID, status string) error {
	_, err := tx.Exec(ctx, `UPDATE fraud_flags SET status = $2 WHERE id = $1`, flagID, status)
	return err
}

func (r *FraudRepository) CountOpenFlags(ctx context.Context, agencyID uuid.UUID) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM fraud_flags WHERE agency_id = $1 AND status = 'open'
	`, agencyID).Scan(&count)
	return count, err
}
