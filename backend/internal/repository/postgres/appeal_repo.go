package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrAppealAlreadyDecided = errors.New("this appeal has already been decided")

type AppealRepository struct {
	db *DB
}

func NewAppealRepository(db *DB) *AppealRepository {
	return &AppealRepository{db: db}
}

func (r *AppealRepository) Create(ctx context.Context, tx pgx.Tx, appeal *domain.Appeal) error {
	return tx.QueryRow(ctx, `
		INSERT INTO appeals (agency_id, case_id, citizen_id, status, grounds)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, filed_at
	`, appeal.AgencyID, appeal.CaseID, appeal.CitizenID, appeal.Status, appeal.Grounds).
		Scan(&appeal.ID, &appeal.FiledAt)
}

func (r *AppealRepository) GetByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Appeal, error) {
	var a domain.Appeal
	var hearingDate *time.Time
	err := tx.QueryRow(ctx, `
		SELECT id, agency_id, case_id, citizen_id, status, grounds, filed_at, hearing_date
		FROM appeals WHERE id = $1
	`, id).Scan(&a.ID, &a.AgencyID, &a.CaseID, &a.CitizenID, &a.Status, &a.Grounds, &a.FiledAt, &hearingDate)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	a.HearingDate = hearingDate
	return &a, nil
}

func (r *AppealRepository) ListByCase(ctx context.Context, tx pgx.Tx, caseID uuid.UUID) ([]domain.Appeal, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, agency_id, case_id, citizen_id, status, grounds, filed_at, hearing_date
		FROM appeals WHERE case_id = $1 ORDER BY filed_at DESC
	`, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appeals []domain.Appeal
	for rows.Next() {
		var a domain.Appeal
		var hearingDate *time.Time
		if err := rows.Scan(&a.ID, &a.AgencyID, &a.CaseID, &a.CitizenID, &a.Status, &a.Grounds, &a.FiledAt, &hearingDate); err != nil {
			return nil, err
		}
		a.HearingDate = hearingDate
		appeals = append(appeals, a)
	}
	return appeals, rows.Err()
}

func (r *AppealRepository) ListByAgency(ctx context.Context, tx pgx.Tx, agencyID uuid.UUID, pendingOnly bool) ([]domain.Appeal, error) {
	query := `
		SELECT a.id, a.agency_id, a.case_id, a.citizen_id, a.status, a.grounds, a.filed_at, a.hearing_date,
		       c.case_number, COALESCE(p.name, ''), COALESCE(up.first_name || ' ' || up.last_name, ''), c.status
		FROM appeals a
		JOIN cases c ON c.id = a.case_id
		LEFT JOIN programs p ON p.id = c.program_id
		LEFT JOIN user_profiles up ON up.user_id = a.citizen_id
		WHERE a.agency_id = $1
	`
	if pendingOnly {
		query += `
		  AND a.status = 'filed'
		  AND NOT EXISTS (SELECT 1 FROM appeal_decisions ad WHERE ad.appeal_id = a.id)
		  AND c.status IN ('appealed', 'appeal_review')
		`
	}
	query += ` ORDER BY a.filed_at DESC`

	rows, err := tx.Query(ctx, query, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appeals []domain.Appeal
	for rows.Next() {
		var a domain.Appeal
		var hearingDate *time.Time
		if err := rows.Scan(&a.ID, &a.AgencyID, &a.CaseID, &a.CitizenID, &a.Status, &a.Grounds, &a.FiledAt, &hearingDate,
			&a.CaseNumber, &a.ProgramName, &a.CitizenName, &a.CaseStatus); err != nil {
			return nil, err
		}
		a.HearingDate = hearingDate
		appeals = append(appeals, a)
	}
	return appeals, rows.Err()
}

func (r *AppealRepository) HasDecision(ctx context.Context, tx pgx.Tx, appealID uuid.UUID) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM appeal_decisions WHERE appeal_id = $1)`, appealID).Scan(&exists)
	return exists, err
}

func (r *AppealRepository) HasOpenAppeal(ctx context.Context, tx pgx.Tx, caseID uuid.UUID) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM appeals a
			WHERE a.case_id = $1
			  AND a.status = 'filed'
			  AND NOT EXISTS (SELECT 1 FROM appeal_decisions ad WHERE ad.appeal_id = a.id)
		)
	`, caseID).Scan(&exists)
	return exists, err
}

func (r *AppealRepository) SaveDecision(ctx context.Context, tx pgx.Tx, decision *domain.AppealDecision) error {
	err := tx.QueryRow(ctx, `
		INSERT INTO appeal_decisions (appeal_id, reviewer_id, decision, rationale)
		VALUES ($1, $2, $3, $4)
		RETURNING id, decided_at
	`, decision.AppealID, decision.ReviewerID, decision.Decision, decision.Rationale).
		Scan(&decision.ID, &decision.DecidedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAppealAlreadyDecided
		}
		return err
	}
	return nil
}

func (r *AppealRepository) UpdateStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string) error {
	_, err := tx.Exec(ctx, `UPDATE appeals SET status = $2 WHERE id = $1`, id, status)
	return err
}
