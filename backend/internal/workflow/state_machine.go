package workflow

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
)

type StateMachine struct {
	repo *postgres.WorkflowRepository
}

func NewStateMachine(repo *postgres.WorkflowRepository) *StateMachine {
	return &StateMachine{repo: repo}
}

func (sm *StateMachine) ValidateTransition(ctx context.Context, agencyID uuid.UUID, fromStatus, toStatus, role string) error {
	role = normalizeWorkflowRole(role)
	transition, err := sm.repo.GetTransition(ctx, agencyID, fromStatus, toStatus)
	if err != nil {
		return err
	}
	if transition == nil {
		return fmt.Errorf("invalid transition from %s to %s", fromStatus, toStatus)
	}
	if transition.RequiredRole != role && !hasElevatedRole(role, transition.RequiredRole) {
		return fmt.Errorf("role %s cannot perform transition requiring %s", role, transition.RequiredRole)
	}
	return nil
}

func normalizeWorkflowRole(role string) string {
	if role == "worker" {
		return "case_worker"
	}
	return role
}

func hasElevatedRole(actual, required string) bool {
	actual = normalizeWorkflowRole(actual)
	required = normalizeWorkflowRole(required)
	if actual == "admin" {
		return true
	}
	if actual == "supervisor" && (required == "case_worker" || required == "citizen") {
		return true
	}
	return false
}

func (sm *StateMachine) AvailableTransitions(ctx context.Context, agencyID uuid.UUID, fromStatus, role string) ([]string, error) {
	role = normalizeWorkflowRole(role)
	transitions, err := sm.repo.ListTransitions(ctx, agencyID)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	var statuses []string
	for _, t := range transitions {
		if t.FromStatus != fromStatus {
			continue
		}
		if t.RequiredRole != role && !hasElevatedRole(role, t.RequiredRole) {
			continue
		}
		if _, ok := seen[t.ToStatus]; ok {
			continue
		}
		seen[t.ToStatus] = struct{}{}
		statuses = append(statuses, t.ToStatus)
	}
	return statuses, nil
}

func (sm *StateMachine) Transition(ctx context.Context, db *postgres.DB, agencyID, userID uuid.UUID, input domain.StatusTransitionInput) (*domain.WorkflowEvent, error) {
	var event *domain.WorkflowEvent
	err := postgres.WithTenant(ctx, db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		caseRepo := postgres.NewCaseRepository(db)
		c, err := caseRepo.GetByID(ctx, tx, input.CaseID)
		if err != nil {
			return err
		}
		if c == nil {
			return fmt.Errorf("case not found")
		}

		if err := sm.ValidateTransition(ctx, agencyID, c.Status, input.ToStatus, normalizeWorkflowRole(input.Role)); err != nil {
			return err
		}

		if err := caseRepo.UpdateStatus(ctx, tx, input.CaseID, input.ToStatus); err != nil {
			return err
		}

		actorID := input.ActorID
		event = &domain.WorkflowEvent{
			AgencyID:   agencyID,
			CaseID:     input.CaseID,
			FromStatus: c.Status,
			ToStatus:   input.ToStatus,
			ActorID:    &actorID,
			Reason:     input.Reason,
		}
		return sm.repo.RecordEvent(ctx, tx, event)
	})
	if err != nil {
		return nil, err
	}
	return event, nil
}
