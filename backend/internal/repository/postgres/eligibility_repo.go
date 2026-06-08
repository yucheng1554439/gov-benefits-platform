package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/jackc/pgx/v5"
)

type EligibilityRepository struct {
	db *DB
}

func NewEligibilityRepository(db *DB) *EligibilityRepository {
	return &EligibilityRepository{db: db}
}

func (r *EligibilityRepository) GetActiveVersion(ctx context.Context, agencyID, programID uuid.UUID) (*domain.EligibilityRuleVersion, error) {
	var v domain.EligibilityRuleVersion
	var conditions, actions []byte
	var effectiveTo *time.Time
	err := r.db.Pool.QueryRow(ctx, `
		SELECT erv.id, erv.rule_id, erv.version, erv.conditions, erv.actions, erv.effective_from, erv.effective_to
		FROM eligibility_rule_versions erv
		JOIN eligibility_rules er ON er.id = erv.rule_id
		WHERE er.agency_id = $1 AND er.program_id = $2 AND er.is_active = true
		  AND erv.effective_from <= CURRENT_DATE
		  AND (erv.effective_to IS NULL OR erv.effective_to >= CURRENT_DATE)
		ORDER BY erv.version DESC LIMIT 1
	`, agencyID, programID).Scan(&v.ID, &v.RuleID, &v.Version, &conditions, &actions, &v.EffectiveFrom, &effectiveTo)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	v.Conditions = decodeJSONMap(conditions)
	v.Actions = decodeJSONMap(actions)
	v.EffectiveTo = effectiveTo
	return &v, nil
}

func (r *EligibilityRepository) SaveEvaluation(ctx context.Context, tx pgx.Tx, eval *domain.EligibilityEvaluation) error {
	return tx.QueryRow(ctx, `
		INSERT INTO eligibility_evaluations (case_id, version_id, is_eligible, evaluation_trace)
		VALUES ($1, $2, $3, $4)
		RETURNING id, evaluated_at
	`, eval.CaseID, eval.VersionID, eval.IsEligible, encodeJSON(eval.EvaluationTrace)).
		Scan(&eval.ID, &eval.EvaluatedAt)
}

func (r *EligibilityRepository) GetLatestEvaluation(ctx context.Context, tx pgx.Tx, caseID uuid.UUID) (*domain.EligibilityEvaluation, error) {
	var e domain.EligibilityEvaluation
	var trace []byte
	err := tx.QueryRow(ctx, `
		SELECT id, case_id, version_id, is_eligible, evaluation_trace, evaluated_at
		FROM eligibility_evaluations WHERE case_id = $1 ORDER BY evaluated_at DESC LIMIT 1
	`, caseID).Scan(&e.ID, &e.CaseID, &e.VersionID, &e.IsEligible, &trace, &e.EvaluatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	e.EvaluationTrace = decodeJSONSlice(trace)
	return &e, nil
}

func (r *EligibilityRepository) ListRulesForAgency(ctx context.Context, agencyID uuid.UUID) ([]domain.EligibilityRuleDetail, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT er.id, er.agency_id, er.program_id, er.name, er.is_active,
		       p.name, p.code,
		       erv.version, erv.effective_from, erv.effective_to, erv.conditions
		FROM eligibility_rules er
		JOIN programs p ON p.id = er.program_id
		LEFT JOIN LATERAL (
			SELECT version, effective_from, effective_to, conditions
			FROM eligibility_rule_versions
			WHERE rule_id = er.id
			ORDER BY version DESC
			LIMIT 1
		) erv ON true
		WHERE er.agency_id = $1
		ORDER BY er.name
	`, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []domain.EligibilityRuleDetail
	for rows.Next() {
		var rule domain.EligibilityRuleDetail
		var conditions []byte
		var effectiveTo *time.Time
		var version *int
		var effectiveFrom *time.Time
		if err := rows.Scan(
			&rule.ID, &rule.AgencyID, &rule.ProgramID, &rule.Name, &rule.IsActive,
			&rule.ProgramName, &rule.ProgramCode,
			&version, &effectiveFrom, &effectiveTo, &conditions,
		); err != nil {
			return nil, err
		}
		if version != nil {
			rule.Version = *version
		}
		if effectiveFrom != nil {
			rule.EffectiveFrom = *effectiveFrom
		}
		rule.EffectiveTo = effectiveTo
		rule.Conditions = decodeJSONMap(conditions)
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *EligibilityRepository) GetActiveVersionByRuleID(ctx context.Context, ruleID uuid.UUID) (*domain.EligibilityRuleVersion, error) {
	var v domain.EligibilityRuleVersion
	var conditions, actions []byte
	var effectiveTo *time.Time
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, rule_id, version, conditions, actions, effective_from, effective_to
		FROM eligibility_rule_versions
		WHERE rule_id = $1
		  AND effective_from <= CURRENT_DATE
		  AND (effective_to IS NULL OR effective_to >= CURRENT_DATE)
		ORDER BY version DESC LIMIT 1
	`, ruleID).Scan(&v.ID, &v.RuleID, &v.Version, &conditions, &actions, &v.EffectiveFrom, &effectiveTo)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	v.Conditions = decodeJSONMap(conditions)
	v.Actions = decodeJSONMap(actions)
	v.EffectiveTo = effectiveTo
	return &v, nil
}
