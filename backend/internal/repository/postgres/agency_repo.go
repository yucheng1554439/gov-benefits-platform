package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/jackc/pgx/v5"
)

type AgencyRepository struct {
	db *DB
}

func NewAgencyRepository(db *DB) *AgencyRepository {
	return &AgencyRepository{db: db}
}

func (r *AgencyRepository) List(ctx context.Context) ([]domain.Agency, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, code, name, COALESCE(type,''), COALESCE(jurisdiction,''), is_active, created_at
		FROM agencies WHERE is_active = true ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agencies []domain.Agency
	for rows.Next() {
		var a domain.Agency
		if err := rows.Scan(&a.ID, &a.Code, &a.Name, &a.Type, &a.Jurisdiction, &a.IsActive, &a.CreatedAt); err != nil {
			return nil, err
		}
		agencies = append(agencies, a)
	}
	return agencies, rows.Err()
}

func (r *AgencyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Agency, error) {
	var a domain.Agency
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, code, name, COALESCE(type,''), COALESCE(jurisdiction,''), is_active, created_at
		FROM agencies WHERE id = $1
	`, id).Scan(&a.ID, &a.Code, &a.Name, &a.Type, &a.Jurisdiction, &a.IsActive, &a.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *AgencyRepository) ListPrograms(ctx context.Context, agencyID uuid.UUID) ([]domain.AgencyProgram, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT ap.id, ap.agency_id, ap.program_id, ap.is_enabled, p.id, p.code, p.name, COALESCE(p.description,'')
		FROM agency_programs ap
		JOIN programs p ON p.id = ap.program_id
		WHERE ap.agency_id = $1 AND ap.is_enabled = true
		ORDER BY p.name
	`, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var programs []domain.AgencyProgram
	for rows.Next() {
		var ap domain.AgencyProgram
		var p domain.Program
		if err := rows.Scan(&ap.ID, &ap.AgencyID, &ap.ProgramID, &ap.IsEnabled, &p.ID, &p.Code, &p.Name, &p.Description); err != nil {
			return nil, err
		}
		ap.Program = &p
		programs = append(programs, ap)
	}
	return programs, rows.Err()
}

func (r *AgencyRepository) GetProgramByID(ctx context.Context, id uuid.UUID) (*domain.Program, error) {
	var p domain.Program
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, code, name, COALESCE(description,'') FROM programs WHERE id = $1
	`, id).Scan(&p.ID, &p.Code, &p.Name, &p.Description)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get program: %w", err)
	}
	return &p, nil
}

func (r *AgencyRepository) IsProgramEnabledForAgency(ctx context.Context, agencyID, programID uuid.UUID) (bool, error) {
	var enabled bool
	err := r.db.Pool.QueryRow(ctx, `
		SELECT is_enabled FROM agency_programs
		WHERE agency_id = $1 AND program_id = $2
	`, agencyID, programID).Scan(&enabled)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return enabled, nil
}

func (r *AgencyRepository) GetProgramByCode(ctx context.Context, code string) (*domain.Program, error) {
	var p domain.Program
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, code, name, COALESCE(description,'') FROM programs WHERE code = $1
	`, code).Scan(&p.ID, &p.Code, &p.Name, &p.Description)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}
