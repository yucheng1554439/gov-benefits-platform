package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/benefit"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/events"
	"github.com/govbenefits/platform/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
)

var ErrNoBenefitRuleVersion = fmt.Errorf("unable to calculate benefit because no active rule version exists")

type BenefitService struct {
	db       *postgres.DB
	benefit  *postgres.BenefitRepository
	appRepo  *postgres.ApplicationRepository
	calc     *benefit.Calculator
	bus      *events.Bus
}

func NewBenefitService(db *postgres.DB, benefitRepo *postgres.BenefitRepository, appRepo *postgres.ApplicationRepository, calc *benefit.Calculator, bus *events.Bus) *BenefitService {
	return &BenefitService{db: db, benefit: benefitRepo, appRepo: appRepo, calc: calc, bus: bus}
}

func (s *BenefitService) Calculate(ctx context.Context, agencyID, userID, caseID, programID uuid.UUID) (*domain.BenefitCalculation, error) {
	version, err := s.benefit.GetActiveVersion(ctx, agencyID, programID)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, ErrNoBenefitRuleVersion
	}

	var app *domain.Application
	_ = postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		app, _ = s.appRepo.GetByCaseID(ctx, tx, caseID)
		return nil
	})

	householdSize := 1
	annualIncome := float64(0)
	if app != nil {
		householdSize = app.HouseholdSize
		annualIncome = app.AnnualIncome
	}

	result, err := s.calc.Calculate(version.Formula, householdSize, annualIncome)
	if err != nil {
		return nil, err
	}

	calc := &domain.BenefitCalculation{
		CaseID:           caseID,
		VersionID:        version.ID,
		RuleVersion:      version.Version,
		CalculatedAmount: result.Amount,
	}
	for _, t := range result.Trace {
		calc.CalculationTrace = append(calc.CalculationTrace, t)
	}

	err = postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		return s.benefit.SaveCalculation(ctx, tx, calc)
	})
	if err != nil {
		return nil, err
	}

	actorID := userID
	s.bus.Publish(ctx, events.NewEvent(events.EventBenefitCalculated, agencyID, caseID, &actorID, map[string]any{
		"amount": calc.CalculatedAmount,
	}))
	return calc, nil
}

func (s *BenefitService) GetLatest(ctx context.Context, agencyID, userID, caseID uuid.UUID) (*domain.BenefitCalculation, error) {
	var calc *domain.BenefitCalculation
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		calc, err = s.benefit.GetLatestCalculation(ctx, tx, caseID)
		return err
	})
	return calc, err
}
