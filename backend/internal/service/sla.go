package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/repository/postgres"
	"github.com/govbenefits/platform/internal/sla"
)

type SLAService struct {
	slaRepo *postgres.SLARepository
	calc    *sla.Calculator
}

func NewSLAService(slaRepo *postgres.SLARepository, calc *sla.Calculator) *SLAService {
	return &SLAService{slaRepo: slaRepo, calc: calc}
}

func (s *SLAService) GetTracking(ctx context.Context, caseID uuid.UUID) (*domain.CaseSLATracking, error) {
	return s.slaRepo.GetTracking(ctx, caseID)
}

func (s *SLAService) ListBreached(ctx context.Context, agencyID uuid.UUID) ([]domain.CaseSLATracking, error) {
	return s.slaRepo.ListBreached(ctx, agencyID)
}

func (s *SLAService) RefreshStatus(ctx context.Context, caseID uuid.UUID) (*domain.CaseSLATracking, error) {
	if err := s.slaRepo.UpdateTrackingStatus(ctx, caseID); err != nil {
		return nil, err
	}
	return s.slaRepo.GetTracking(ctx, caseID)
}

func (s *SLAService) ComputeDueDate(ctx context.Context, agencyID, programID uuid.UUID, start time.Time) (*time.Time, error) {
	policy, err := s.slaRepo.GetPolicy(ctx, agencyID, programID)
	if err != nil || policy == nil {
		return nil, err
	}
	due := s.calc.ComputeDueDate(policy, start)
	return &due, nil
}
