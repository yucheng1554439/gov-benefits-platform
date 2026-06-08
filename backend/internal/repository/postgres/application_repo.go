package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/jackc/pgx/v5"
)

type ApplicationRepository struct {
	db *DB
}

func NewApplicationRepository(db *DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) Create(ctx context.Context, tx pgx.Tx, app *domain.Application) error {
	return tx.QueryRow(ctx, `
		INSERT INTO applications (agency_id, case_id, household_size, annual_income, employment_status, form_data)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, app.AgencyID, app.CaseID, app.HouseholdSize, app.AnnualIncome, app.EmploymentStatus, encodeJSON(app.FormData)).
		Scan(&app.ID, &app.CreatedAt)
}

func (r *ApplicationRepository) GetByCaseID(ctx context.Context, tx pgx.Tx, caseID uuid.UUID) (*domain.Application, error) {
	var app domain.Application
	var formData []byte
	err := tx.QueryRow(ctx, `
		SELECT id, agency_id, case_id, household_size, annual_income, COALESCE(employment_status,''), form_data, created_at
		FROM applications WHERE case_id = $1
	`, caseID).Scan(&app.ID, &app.AgencyID, &app.CaseID, &app.HouseholdSize, &app.AnnualIncome, &app.EmploymentStatus, &formData, &app.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	app.FormData = decodeJSONMap(formData)
	return &app, nil
}
