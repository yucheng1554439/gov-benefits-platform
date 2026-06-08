package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/eligibility"
	"github.com/govbenefits/platform/internal/events"
	"github.com/govbenefits/platform/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
)

var ErrNoEligibilityRuleVersion = fmt.Errorf("unable to evaluate eligibility because no active rule version exists")

type EligibilityService struct {
	db        *postgres.DB
	eligRepo  *postgres.EligibilityRepository
	appRepo   *postgres.ApplicationRepository
	evaluator *eligibility.Evaluator
	bus       *events.Bus
}

func NewEligibilityService(db *postgres.DB, eligRepo *postgres.EligibilityRepository, appRepo *postgres.ApplicationRepository, evaluator *eligibility.Evaluator, bus *events.Bus) *EligibilityService {
	return &EligibilityService{db: db, eligRepo: eligRepo, appRepo: appRepo, evaluator: evaluator, bus: bus}
}

func (s *EligibilityService) Evaluate(ctx context.Context, agencyID, userID, caseID, programID uuid.UUID) (*domain.EligibilityEvaluation, error) {
	version, err := s.eligRepo.GetActiveVersion(ctx, agencyID, programID)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, ErrNoEligibilityRuleVersion
	}

	var app *domain.Application
	_ = postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		app, _ = s.appRepo.GetByCaseID(ctx, tx, caseID)
		return nil
	})

	data := map[string]any{
		"annual_income":  float64(0),
		"household_size": float64(1),
	}
	if app != nil {
		data["annual_income"] = app.AnnualIncome
		data["household_size"] = float64(app.HouseholdSize)
	}

	result, err := s.evaluator.Evaluate(version.Conditions, data)
	if err != nil {
		return nil, err
	}

	eval := &domain.EligibilityEvaluation{
		CaseID:     caseID,
		VersionID:  version.ID,
		IsEligible: result.Eligible,
	}
	for _, t := range result.Trace {
		eval.EvaluationTrace = append(eval.EvaluationTrace, t)
	}

	err = postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		return s.eligRepo.SaveEvaluation(ctx, tx, eval)
	})
	if err != nil {
		return nil, err
	}

	actorID := userID
	s.bus.Publish(ctx, events.NewEvent(events.EventEligibilityEvaluated, agencyID, caseID, &actorID, map[string]any{
		"is_eligible": eval.IsEligible,
	}))
	return eval, nil
}

func (s *EligibilityService) GetLatest(ctx context.Context, agencyID, userID, caseID uuid.UUID) (*domain.EligibilityEvaluation, error) {
	var eval *domain.EligibilityEvaluation
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		eval, err = s.eligRepo.GetLatestEvaluation(ctx, tx, caseID)
		return err
	})
	return eval, err
}

func (s *EligibilityService) ListRules(ctx context.Context, agencyID uuid.UUID) ([]domain.EligibilityRuleDetail, error) {
	return s.eligRepo.ListRulesForAgency(ctx, agencyID)
}

type SimulateResult struct {
	IsEligible      bool  `json:"is_eligible"`
	EvaluationTrace []any `json:"evaluation_trace,omitempty"`
	RuleVersion     int   `json:"rule_version"`
}

func (s *EligibilityService) SimulateRule(ctx context.Context, agencyID, ruleID uuid.UUID, data map[string]any) (*SimulateResult, error) {
	version, err := s.eligRepo.GetActiveVersionByRuleID(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, ErrNoEligibilityRuleVersion
	}
	if data == nil {
		data = map[string]any{
			"annual_income":  float64(25000),
			"household_size": float64(2),
		}
	}
	result, err := s.evaluator.Evaluate(version.Conditions, data)
	if err != nil {
		return nil, err
	}
	trace := make([]any, len(result.Trace))
	for i, t := range result.Trace {
		trace[i] = t
	}
	return &SimulateResult{
		IsEligible:      result.Eligible,
		EvaluationTrace: trace,
		RuleVersion:     version.Version,
	}, nil
}
