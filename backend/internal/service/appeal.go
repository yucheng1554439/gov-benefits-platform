package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/events"
	"github.com/govbenefits/platform/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
)

var ErrAppealAlreadyDecided = errors.New("this appeal has already been decided")

type AppealService struct {
	db      *postgres.DB
	appeals *postgres.AppealRepository
	cases   *CaseService
	bus     *events.Bus
}

func NewAppealService(db *postgres.DB, appeals *postgres.AppealRepository, cases *CaseService, bus *events.Bus) *AppealService {
	return &AppealService{db: db, appeals: appeals, cases: cases, bus: bus}
}

func (s *AppealService) File(ctx context.Context, agencyID, citizenID, caseID uuid.UUID, grounds string) (*domain.Appeal, error) {
	caseObj, err := s.cases.Get(ctx, agencyID, citizenID, caseID)
	if err != nil {
		return nil, err
	}
	if caseObj == nil {
		return nil, fmt.Errorf("case not found")
	}
	if caseObj.Status != "denied" {
		return nil, fmt.Errorf("appeals can only be filed on denied cases")
	}

	var openAppeal bool
	_ = postgres.WithTenant(ctx, s.db, agencyID, citizenID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		openAppeal, err = s.appeals.HasOpenAppeal(ctx, tx, caseID)
		return err
	})
	if openAppeal {
		return nil, fmt.Errorf("an appeal is already pending for this case")
	}

	var appeal *domain.Appeal
	err = postgres.WithTenant(ctx, s.db, agencyID, citizenID, func(ctx context.Context, tx pgx.Tx) error {
		appeal = &domain.Appeal{
			AgencyID:  agencyID,
			CaseID:    caseID,
			CitizenID: citizenID,
			Status:    "filed",
			Grounds:   grounds,
		}
		return s.appeals.Create(ctx, tx, appeal)
	})
	if err != nil {
		return nil, err
	}

	if _, err := s.cases.TransitionStatus(ctx, agencyID, citizenID, domain.StatusTransitionInput{
		CaseID: caseID, ToStatus: "appealed", Role: "citizen",
	}); err != nil {
		return nil, fmt.Errorf("appeal filed but case transition failed: %w", err)
	}

	actorID := citizenID
	s.bus.Publish(ctx, events.NewEvent(events.EventAppealFiled, agencyID, appeal.ID, &actorID, map[string]any{
		"case_id": caseID.String(),
	}))
	return appeal, nil
}

func (s *AppealService) List(ctx context.Context, agencyID, userID, caseID uuid.UUID) ([]domain.Appeal, error) {
	var appeals []domain.Appeal
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		appeals, err = s.appeals.ListByCase(ctx, tx, caseID)
		return err
	})
	return appeals, err
}

func (s *AppealService) ListAgency(ctx context.Context, agencyID, userID uuid.UUID, pendingOnly bool) ([]domain.Appeal, error) {
	var appeals []domain.Appeal
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		appeals, err = s.appeals.ListByAgency(ctx, tx, agencyID, pendingOnly)
		return err
	})
	return appeals, err
}

func (s *AppealService) Decide(ctx context.Context, agencyID, userID, appealID uuid.UUID, decision, rationale string) error {
	var appeal *domain.Appeal
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		appeal, err = s.appeals.GetByID(ctx, tx, appealID)
		if err != nil || appeal == nil {
			return fmt.Errorf("appeal not found")
		}
		if appeal.Status == "decided" {
			return ErrAppealAlreadyDecided
		}
		hasDecision, err := s.appeals.HasDecision(ctx, tx, appealID)
		if err != nil {
			return err
		}
		if hasDecision {
			_ = s.appeals.UpdateStatus(ctx, tx, appealID, "decided")
			return ErrAppealAlreadyDecided
		}

		d := &domain.AppealDecision{
			AppealID:   appealID,
			ReviewerID: userID,
			Decision:   decision,
			Rationale:  rationale,
		}
		if err := s.appeals.SaveDecision(ctx, tx, d); err != nil {
			if errors.Is(err, postgres.ErrAppealAlreadyDecided) {
				return ErrAppealAlreadyDecided
			}
			return err
		}
		return s.appeals.UpdateStatus(ctx, tx, appealID, "decided")
	})
	if err != nil {
		if errors.Is(err, ErrAppealAlreadyDecided) || errors.Is(err, postgres.ErrAppealAlreadyDecided) {
			return ErrAppealAlreadyDecided
		}
		return err
	}

	toStatus, err := appealDecisionToCaseStatus(decision)
	if err != nil {
		return err
	}

	caseObj, err := s.cases.Get(ctx, agencyID, userID, appeal.CaseID)
	if err != nil || caseObj == nil {
		return fmt.Errorf("case not found for appeal")
	}
	if caseObj.Status == "appealed" && toStatus != "appeal_review" {
		if _, err := s.cases.TransitionStatus(ctx, agencyID, userID, domain.StatusTransitionInput{
			CaseID: appeal.CaseID, ToStatus: "appeal_review", Role: "supervisor",
		}); err != nil {
			return fmt.Errorf("move case to appeal review: %w", err)
		}
	}
	if _, err := s.cases.TransitionStatus(ctx, agencyID, userID, domain.StatusTransitionInput{
		CaseID: appeal.CaseID, ToStatus: toStatus, Role: "supervisor", Reason: rationale,
	}); err != nil {
		return fmt.Errorf("decision saved but case transition failed: %w", err)
	}

	actorID := userID
	s.bus.Publish(ctx, events.NewEvent(events.EventAppealDecided, agencyID, appealID, &actorID, map[string]any{
		"case_id":  appeal.CaseID.String(),
		"decision": decision,
	}))
	return nil
}

func appealDecisionToCaseStatus(decision string) (string, error) {
	switch decision {
	case "upheld":
		return "appeal_denied", nil
	case "overturned":
		return "appeal_approved", nil
	case "remanded":
		return "appeal_review", nil
	default:
		return "", fmt.Errorf("unsupported appeal decision: %s", decision)
	}
}
