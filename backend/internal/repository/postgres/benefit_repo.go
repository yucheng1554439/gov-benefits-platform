package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/jackc/pgx/v5"
)

type BenefitRepository struct {
	db *DB
}

func NewBenefitRepository(db *DB) *BenefitRepository {
	return &BenefitRepository{db: db}
}

func (r *BenefitRepository) GetActiveVersion(ctx context.Context, agencyID, programID uuid.UUID) (*domain.BenefitCalculationVersion, error) {
	var v domain.BenefitCalculationVersion
	var formula []byte
	var effectiveTo *time.Time
	err := r.db.Pool.QueryRow(ctx, `
		SELECT bcv.id, bcv.rule_id, bcv.version, bcv.formula, bcv.effective_from, bcv.effective_to
		FROM benefit_calculation_versions bcv
		JOIN benefit_calculation_rules bcr ON bcr.id = bcv.rule_id
		WHERE bcr.agency_id = $1 AND bcr.program_id = $2 AND bcr.is_active = true
		  AND bcv.effective_from <= CURRENT_DATE
		  AND (bcv.effective_to IS NULL OR bcv.effective_to >= CURRENT_DATE)
		ORDER BY bcv.version DESC LIMIT 1
	`, agencyID, programID).Scan(&v.ID, &v.RuleID, &v.Version, &formula, &v.EffectiveFrom, &effectiveTo)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	v.Formula = decodeJSONMap(formula)
	v.EffectiveTo = effectiveTo
	return &v, nil
}

func (r *BenefitRepository) SaveCalculation(ctx context.Context, tx pgx.Tx, calc *domain.BenefitCalculation) error {
	return tx.QueryRow(ctx, `
		INSERT INTO benefit_calculations (case_id, version_id, calculated_amount, approved_amount, calculation_trace)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, calculated_at
	`, calc.CaseID, calc.VersionID, calc.CalculatedAmount, calc.ApprovedAmount, encodeJSON(calc.CalculationTrace)).
		Scan(&calc.ID, &calc.CalculatedAt)
}

func (r *BenefitRepository) GetLatestCalculation(ctx context.Context, tx pgx.Tx, caseID uuid.UUID) (*domain.BenefitCalculation, error) {
	var c domain.BenefitCalculation
	var trace []byte
	err := tx.QueryRow(ctx, `
		SELECT bc.id, bc.case_id, bc.version_id, bc.calculated_amount, bc.approved_amount, bc.calculation_trace, bc.calculated_at, bcv.version
		FROM benefit_calculations bc
		JOIN benefit_calculation_versions bcv ON bcv.id = bc.version_id
		WHERE bc.case_id = $1 ORDER BY bc.calculated_at DESC LIMIT 1
	`, caseID).Scan(&c.ID, &c.CaseID, &c.VersionID, &c.CalculatedAmount, &c.ApprovedAmount, &trace, &c.CalculatedAt, &c.RuleVersion)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	c.CalculationTrace = decodeJSONSlice(trace)
	return &c, nil
}
