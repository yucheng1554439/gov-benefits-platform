package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/assignment"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/events"
	"github.com/govbenefits/platform/internal/repository/postgres"
	"github.com/govbenefits/platform/internal/sla"
	"github.com/govbenefits/platform/internal/workflow"
	"github.com/jackc/pgx/v5"
)

type CaseService struct {
	db         *postgres.DB
	cases      *postgres.CaseRepository
	apps       *postgres.ApplicationRepository
	agencies   *postgres.AgencyRepository
	users      *postgres.UserRepository
	workflow   *workflow.StateMachine
	allocator  *assignment.Allocator
	workerRepo *postgres.WorkerRepository
	slaRepo    *postgres.SLARepository
	slaCalc    *sla.Calculator
	bus        *events.Bus
}

func NewCaseService(
	db *postgres.DB,
	cases *postgres.CaseRepository,
	apps *postgres.ApplicationRepository,
	agencies *postgres.AgencyRepository,
	users *postgres.UserRepository,
	wf *workflow.StateMachine,
	allocator *assignment.Allocator,
	workerRepo *postgres.WorkerRepository,
	slaRepo *postgres.SLARepository,
	slaCalc *sla.Calculator,
	bus *events.Bus,
) *CaseService {
	return &CaseService{
		db: db, cases: cases, apps: apps, agencies: agencies, users: users, workflow: wf,
		allocator: allocator, workerRepo: workerRepo,
		slaRepo: slaRepo, slaCalc: slaCalc, bus: bus,
	}
}

func (s *CaseService) CreateApplication(ctx context.Context, input domain.CreateApplicationInput) (*domain.Case, error) {
	caseNumber, err := s.cases.NextCaseNumber(ctx, input.AgencyID)
	if err != nil {
		return nil, err
	}

	priority := input.Priority
	if priority == "" {
		priority = "normal"
	}

	var created *domain.Case
	err = postgres.WithTenant(ctx, s.db, input.AgencyID, input.CitizenID, func(ctx context.Context, tx pgx.Tx) error {
		c := &domain.Case{
			AgencyID:    input.AgencyID,
			CaseNumber:  caseNumber,
			CitizenID:   input.CitizenID,
			ProgramID:   input.ProgramID,
			Status:      "submitted",
			Priority:    priority,
			ZipCode:     input.ZipCode,
			CensusTract: input.CensusTract,
		}
		if err := s.cases.Create(ctx, tx, c); err != nil {
			return err
		}

		app := &domain.Application{
			AgencyID:         input.AgencyID,
			CaseID:           c.ID,
			HouseholdSize:    input.HouseholdSize,
			AnnualIncome:     input.AnnualIncome,
			EmploymentStatus: input.EmploymentStatus,
			FormData:         input.FormData,
		}
		if err := s.apps.Create(ctx, tx, app); err != nil {
			return err
		}

		workerID, err := s.allocator.AssignWorker(ctx, input.AgencyID, input.ProgramID)
		if err == nil && workerID != uuid.Nil {
			_ = s.cases.AssignWorker(ctx, tx, c.ID, workerID)
			_ = s.workerRepo.IncrementCaseCount(ctx, workerID)
		}

		c.Application = app
		created = c
		return nil
	})
	if err != nil {
		return nil, err
	}

	if policy, _ := s.slaRepo.GetPolicy(ctx, input.AgencyID, input.ProgramID); policy != nil {
		dueAt := s.slaCalc.ComputeDueDate(policy, created.SubmittedAt)
		_ = s.slaRepo.CreateTracking(ctx, &domain.CaseSLATracking{
			CaseID:      created.ID,
			SLAPolicyID: policy.ID,
			DueAt:       dueAt,
			Status:      "on_track",
		})
	}

	s.bus.Publish(ctx, events.NewEvent(events.EventApplicationCreated, input.AgencyID, created.ID, &input.CitizenID, map[string]any{
		"case_number": caseNumber,
	}))
	s.bus.Publish(ctx, events.NewEvent(events.EventCaseCreated, input.AgencyID, created.ID, &input.CitizenID, map[string]any{
		"citizen_id": input.CitizenID.String(),
	}))
	return created, nil
}

func (s *CaseService) Get(ctx context.Context, agencyID, userID, caseID uuid.UUID) (*domain.Case, error) {
	var c *domain.Case
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		c, err = s.cases.GetByID(ctx, tx, caseID)
		if err != nil || c == nil {
			return err
		}
		app, err := s.apps.GetByCaseID(ctx, tx, caseID)
		if err != nil {
			return err
		}
		c.Application = app
		if c.ProgramID != uuid.Nil {
			if program, err := s.agencies.GetProgramByID(ctx, c.ProgramID); err == nil && program != nil {
				c.Program = program
			}
		}
		return nil
	})
	return c, err
}

func (s *CaseService) List(ctx context.Context, agencyID, userID uuid.UUID, filter domain.CaseListFilter) ([]domain.Case, error) {
	filter.AgencyID = agencyID
	var cases []domain.Case
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		cases, err = s.cases.List(ctx, tx, filter)
		return err
	})
	if err != nil {
		return nil, err
	}
	for i := range cases {
		if cases[i].ProgramID == uuid.Nil {
			continue
		}
		program, err := s.agencies.GetProgramByID(ctx, cases[i].ProgramID)
		if err != nil || program == nil {
			continue
		}
		cases[i].Program = program
	}
	return cases, nil
}

func (s *CaseService) Update(ctx context.Context, agencyID, userID, caseID uuid.UUID, priority, zipCode, censusTract string) error {
	return postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		return s.cases.Update(ctx, tx, caseID, priority, zipCode, censusTract)
	})
}

func (s *CaseService) TransitionStatus(ctx context.Context, agencyID, userID uuid.UUID, input domain.StatusTransitionInput) (*domain.WorkflowEvent, error) {
	input.ActorID = userID
	event, err := s.workflow.Transition(ctx, s.db, agencyID, userID, input)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"from_status": event.FromStatus,
		"to_status":   event.ToStatus,
	}
	if caseObj, _ := s.Get(ctx, agencyID, userID, input.CaseID); caseObj != nil {
		payload["citizen_id"] = caseObj.CitizenID.String()
	}
	actorID := userID
	s.bus.Publish(ctx, events.NewEvent(events.EventCaseStatusChanged, agencyID, input.CaseID, &actorID, payload))
	return event, nil
}

func (s *CaseService) GetAvailableTransitions(ctx context.Context, agencyID uuid.UUID, fromStatus, role string) ([]string, error) {
	return s.workflow.AvailableTransitions(ctx, agencyID, fromStatus, role)
}

func (s *CaseService) GetWorkflowHistory(ctx context.Context, agencyID, userID, caseID uuid.UUID) ([]domain.WorkflowEvent, error) {
	var wfEvents []domain.WorkflowEvent
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		repo := postgres.NewWorkflowRepository(s.db)
		var err error
		wfEvents, err = repo.ListEvents(ctx, tx, caseID)
		return err
	})
	if err != nil {
		return nil, err
	}

	var actorIDs []uuid.UUID
	for _, e := range wfEvents {
		if e.ActorID != nil && *e.ActorID != uuid.Nil {
			actorIDs = append(actorIDs, *e.ActorID)
		}
	}
	names, _ := s.users.GetDisplayNames(ctx, actorIDs)
	for i := range wfEvents {
		if wfEvents[i].ActorID != nil {
			if name, ok := names[*wfEvents[i].ActorID]; ok && name != "" {
				wfEvents[i].ActorName = name
			}
		}
	}
	return wfEvents, nil
}

func (s *CaseService) ValidateCaseAccess(c *domain.Case, userID uuid.UUID, roles []string, agencyRole string) error {
	isCitizen := contains(roles, "citizen")
	isStaff := contains(roles, "case_worker") || contains(roles, "supervisor") || contains(roles, "admin")
	if isCitizen && c.CitizenID == userID {
		return nil
	}
	if isStaff {
		return nil
	}
	return fmt.Errorf("access denied")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
