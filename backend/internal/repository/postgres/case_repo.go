package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/jackc/pgx/v5"
)

type CaseRepository struct {
	db *DB
}

func NewCaseRepository(db *DB) *CaseRepository {
	return &CaseRepository{db: db}
}

func (r *CaseRepository) Create(ctx context.Context, tx pgx.Tx, c *domain.Case) error {
	query := `
		INSERT INTO cases (agency_id, case_number, citizen_id, program_id, status, priority, zip_code, census_tract)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, submitted_at, created_at, updated_at
	`
	return tx.QueryRow(ctx, query,
		c.AgencyID, c.CaseNumber, c.CitizenID, c.ProgramID, c.Status, c.Priority, c.ZipCode, c.CensusTract,
	).Scan(&c.ID, &c.SubmittedAt, &c.CreatedAt, &c.UpdatedAt)
}

func (r *CaseRepository) GetByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Case, error) {
	var c domain.Case
	var closedAt *time.Time
	err := tx.QueryRow(ctx, `
		SELECT id, agency_id, case_number, citizen_id, program_id, status, priority,
		       COALESCE(zip_code,''), COALESCE(census_tract,''), submitted_at, closed_at, created_at, updated_at
		FROM cases WHERE id = $1
	`, id).Scan(
		&c.ID, &c.AgencyID, &c.CaseNumber, &c.CitizenID, &c.ProgramID, &c.Status, &c.Priority,
		&c.ZipCode, &c.CensusTract, &c.SubmittedAt, &closedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	c.ClosedAt = closedAt
	return &c, nil
}

func (r *CaseRepository) List(ctx context.Context, tx pgx.Tx, filter domain.CaseListFilter) ([]domain.Case, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, agency_id, case_number, citizen_id, program_id, status, priority,
		       COALESCE(zip_code,''), COALESCE(census_tract,''), submitted_at, closed_at, created_at, updated_at
		FROM cases WHERE agency_id = $1
	`
	args := []any{filter.AgencyID}
	argIdx := 2

	if filter.CitizenID != nil {
		query += fmt.Sprintf(" AND citizen_id = $%d", argIdx)
		args = append(args, *filter.CitizenID)
		argIdx++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, filter.Offset)

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []domain.Case
	for rows.Next() {
		var c domain.Case
		var closedAt *time.Time
		if err := rows.Scan(
			&c.ID, &c.AgencyID, &c.CaseNumber, &c.CitizenID, &c.ProgramID, &c.Status, &c.Priority,
			&c.ZipCode, &c.CensusTract, &c.SubmittedAt, &closedAt, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		c.ClosedAt = closedAt
		cases = append(cases, c)
	}
	return cases, rows.Err()
}

func (r *CaseRepository) UpdateStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string) error {
	closeCase := status == "closed" || status == "appeal_denied"
	_, err := tx.Exec(ctx, `
		UPDATE cases SET status = $2, updated_at = NOW(),
			closed_at = CASE WHEN $3 THEN NOW() ELSE closed_at END
		WHERE id = $1
	`, id, status, closeCase)
	return err
}

func (r *CaseRepository) Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, priority, zipCode, censusTract string) error {
	_, err := tx.Exec(ctx, `
		UPDATE cases SET priority = COALESCE(NULLIF($2,''), priority),
			zip_code = COALESCE(NULLIF($3,''), zip_code),
			census_tract = COALESCE(NULLIF($4,''), census_tract),
			updated_at = NOW()
		WHERE id = $1
	`, id, priority, zipCode, censusTract)
	return err
}

func (r *CaseRepository) NextCaseNumber(ctx context.Context, agencyID uuid.UUID) (string, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM cases WHERE agency_id = $1`, agencyID).Scan(&count)
	if err != nil {
		return "", err
	}
	year := time.Now().Year()
	return fmt.Sprintf("CASE-%d-%06d", year, count+1), nil
}

func (r *CaseRepository) AssignWorker(ctx context.Context, tx pgx.Tx, caseID, workerID uuid.UUID) error {
	_, err := tx.Exec(ctx, `UPDATE case_assignments SET is_active = false WHERE case_id = $1`, caseID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO case_assignments (case_id, worker_id, is_active) VALUES ($1, $2, true)
	`, caseID, workerID)
	return err
}
