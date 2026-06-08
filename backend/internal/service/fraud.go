package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/events"
	"github.com/govbenefits/platform/internal/fraud"
	"github.com/govbenefits/platform/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
)

type FraudService struct {
	db       *postgres.DB
	fraud    *postgres.FraudRepository
	caseRepo *postgres.CaseRepository
	appRepo  *postgres.ApplicationRepository
	detector *fraud.Detector
	bus      *events.Bus
}

func NewFraudService(db *postgres.DB, fraudRepo *postgres.FraudRepository, caseRepo *postgres.CaseRepository, appRepo *postgres.ApplicationRepository, detector *fraud.Detector, bus *events.Bus) *FraudService {
	return &FraudService{db: db, fraud: fraudRepo, caseRepo: caseRepo, appRepo: appRepo, detector: detector, bus: bus}
}

func (s *FraudService) ScanCase(ctx context.Context, agencyID, userID, caseID uuid.UUID) ([]domain.FraudFlag, error) {
	var c *domain.Case
	var app *domain.Application
	var existing []domain.FraudFlag

	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		c, err = s.caseRepo.GetByID(ctx, tx, caseID)
		if err != nil || c == nil {
			return err
		}
		app, _ = s.appRepo.GetByCaseID(ctx, tx, caseID)
		existing, _ = s.fraud.ListByCase(ctx, tx, caseID)
		return nil
	})
	if err != nil || c == nil {
		return nil, err
	}

	result := s.detector.CheckCase(c, app, len(existing))
	var created []domain.FraudFlag

	err = postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		for i := range result.Flags {
			flag := result.Flags[i]
			if err := s.fraud.CreateFlag(ctx, tx, &flag); err != nil {
				return err
			}
			created = append(created, flag)
			actorID := userID
			s.bus.Publish(ctx, events.NewEvent(events.EventFraudFlagged, agencyID, flag.ID, &actorID, map[string]any{
				"case_id": caseID.String(),
				"type":    flag.FlagType,
			}))
		}
		return nil
	})
	return created, err
}

func (s *FraudService) ListFlags(ctx context.Context, agencyID, userID, caseID uuid.UUID) ([]domain.FraudFlag, error) {
	var flags []domain.FraudFlag
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		flags, err = s.fraud.ListByCase(ctx, tx, caseID)
		return err
	})
	return flags, err
}

func (s *FraudService) ReviewFlag(ctx context.Context, agencyID, userID, flagID uuid.UUID, disposition, notes string) error {
	return postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		review := &domain.FraudReview{
			FraudFlagID: flagID,
			ReviewerID:  userID,
			Disposition: disposition,
			Notes:       notes,
		}
		if err := s.fraud.CreateReview(ctx, tx, review); err != nil {
			return err
		}
		return s.fraud.UpdateFlagStatus(ctx, tx, flagID, "reviewed")
	})
}
