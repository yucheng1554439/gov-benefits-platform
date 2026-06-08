package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
)

type ReportRepository struct {
	db *DB
}

func NewReportRepository(db *DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Create(ctx context.Context, report *domain.Report) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO reports (agency_id, report_type, status, params, requested_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, report.AgencyID, report.ReportType, report.Status, encodeJSON(report.Params), report.RequestedBy).
		Scan(&report.ID, &report.CreatedAt)
}

func (r *ReportRepository) UpdateCompleted(ctx context.Context, id uuid.UUID, fileKey string) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE reports SET status = 'completed', file_key = $2, completed_at = NOW() WHERE id = $1
	`, id, fileKey)
	return err
}

func (r *ReportRepository) List(ctx context.Context, agencyID uuid.UUID) ([]domain.Report, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, agency_id, report_type, status, COALESCE(file_key,''), params, requested_by, created_at, completed_at
		FROM reports WHERE agency_id = $1 ORDER BY created_at DESC LIMIT 50
	`, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []domain.Report
	for rows.Next() {
		var rep domain.Report
		var params []byte
		var completedAt *time.Time
		if err := rows.Scan(&rep.ID, &rep.AgencyID, &rep.ReportType, &rep.Status, &rep.FileKey, &params, &rep.RequestedBy, &rep.CreatedAt, &completedAt); err != nil {
			return nil, err
		}
		rep.Params = decodeJSONMap(params)
		rep.CompletedAt = completedAt
		reports = append(reports, rep)
	}
	return reports, rows.Err()
}

func (r *ReportRepository) CaseStatusCounts(ctx context.Context, agencyID uuid.UUID) (map[string]int, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT status, COUNT(*) FROM cases WHERE agency_id = $1 GROUP BY status
	`, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
	}
	return counts, rows.Err()
}

func (r *ReportRepository) CasesByZip(ctx context.Context, agencyID uuid.UUID) (map[string]int, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT COALESCE(zip_code,'unknown'), COUNT(*) FROM cases WHERE agency_id = $1 GROUP BY zip_code
	`, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var zip string
		var count int
		if err := rows.Scan(&zip, &count); err != nil {
			return nil, err
		}
		counts[zip] = count
	}
	return counts, rows.Err()
}
