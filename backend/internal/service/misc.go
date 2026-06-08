package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/repository/postgres"
)

type AnalyticsService struct {
	reports *postgres.ReportRepository
	fraud   *postgres.FraudRepository
}

func NewAnalyticsService(reports *postgres.ReportRepository, fraud *postgres.FraudRepository) *AnalyticsService {
	return &AnalyticsService{reports: reports, fraud: fraud}
}

type AnalyticsSummary struct {
	CaseStatusCounts map[string]int `json:"case_status_counts"`
	CasesByZip       map[string]int `json:"cases_by_zip"`
	OpenFraudFlags   int            `json:"open_fraud_flags"`
}

func (s *AnalyticsService) Summary(ctx context.Context, agencyID uuid.UUID) (*AnalyticsSummary, error) {
	statusCounts, err := s.reports.CaseStatusCounts(ctx, agencyID)
	if err != nil {
		return nil, err
	}
	zipCounts, err := s.reports.CasesByZip(ctx, agencyID)
	if err != nil {
		return nil, err
	}
	openFlags, err := s.fraud.CountOpenFlags(ctx, agencyID)
	if err != nil {
		return nil, err
	}
	return &AnalyticsSummary{
		CaseStatusCounts: statusCounts,
		CasesByZip:       zipCounts,
		OpenFraudFlags:   openFlags,
	}, nil
}

type ReportService struct {
	reports *postgres.ReportRepository
	queue   ReportQueue
}

type ReportQueue interface {
	EnqueueReport(ctx context.Context, reportID uuid.UUID) error
}

func NewReportService(reports *postgres.ReportRepository, queue ReportQueue) *ReportService {
	return &ReportService{reports: reports, queue: queue}
}

func (s *ReportService) Request(ctx context.Context, agencyID, userID uuid.UUID, reportType string, params map[string]any) (*domain.Report, error) {
	report := &domain.Report{
		AgencyID:    agencyID,
		ReportType:  reportType,
		Status:      "pending",
		Params:      params,
		RequestedBy: &userID,
	}
	if err := s.reports.Create(ctx, report); err != nil {
		return nil, err
	}
	if s.queue != nil {
		_ = s.queue.EnqueueReport(ctx, report.ID)
	}
	return report, nil
}

func (s *ReportService) List(ctx context.Context, agencyID uuid.UUID) ([]domain.Report, error) {
	return s.reports.List(ctx, agencyID)
}

func (s *ReportService) ProcessFromJob(ctx context.Context, reportID uuid.UUID) error {
	key := "reports/" + reportID.String() + ".json"
	return s.reports.UpdateCompleted(ctx, reportID, key)
}

type AuditService struct {
	audit *postgres.AuditRepository
}

func NewAuditService(audit *postgres.AuditRepository) *AuditService {
	return &AuditService{audit: audit}
}

func (s *AuditService) List(ctx context.Context, agencyID uuid.UUID, filter postgres.AuditListFilter) ([]domain.AuditLog, int, error) {
	agencyStr := agencyID.String()
	logs, err := s.audit.List(ctx, agencyStr, filter)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.audit.Count(ctx, agencyStr, filter)
	if err != nil {
		return logs, len(logs), nil
	}
	return logs, count, nil
}

type NotificationService struct {
	notifications *postgres.NotificationRepository
}

func NewNotificationService(notifications *postgres.NotificationRepository) *NotificationService {
	return &NotificationService{notifications: notifications}
}

func (s *NotificationService) List(ctx context.Context, userID uuid.UUID, unreadOnly bool) ([]domain.Notification, error) {
	return s.notifications.ListByUser(ctx, userID, unreadOnly)
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return s.notifications.MarkRead(ctx, id, userID)
}

type FeatureFlagService struct {
	repo *postgres.FeatureFlagRepository
}

func NewFeatureFlagService(repo *postgres.FeatureFlagRepository) *FeatureFlagService {
	return &FeatureFlagService{repo: repo}
}

func (s *FeatureFlagService) List(ctx context.Context, agencyID uuid.UUID) ([]domain.FeatureFlag, error) {
	return s.repo.List(ctx, agencyID)
}

func (s *FeatureFlagService) Upsert(ctx context.Context, flag *domain.FeatureFlag) error {
	return s.repo.Upsert(ctx, flag)
}

type RetentionService struct {
	db *postgres.DB
}

func NewRetentionService(db *postgres.DB) *RetentionService {
	return &RetentionService{db: db}
}

func (s *RetentionService) ListPolicies(ctx context.Context, agencyID uuid.UUID) ([]domain.RetentionPolicy, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, agency_id, entity_type, retention_years, disposition_action, is_active
		FROM retention_policies WHERE agency_id = $1 AND is_active = true
	`, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []domain.RetentionPolicy
	for rows.Next() {
		var p domain.RetentionPolicy
		if err := rows.Scan(&p.ID, &p.AgencyID, &p.EntityType, &p.RetentionYears, &p.DispositionAction, &p.IsActive); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

type AgencyService struct {
	agencies *postgres.AgencyRepository
}

func NewAgencyService(agencies *postgres.AgencyRepository) *AgencyService {
	return &AgencyService{agencies: agencies}
}

func (s *AgencyService) List(ctx context.Context) ([]domain.Agency, error) {
	return s.agencies.List(ctx)
}

func (s *AgencyService) GetPrograms(ctx context.Context, agencyID uuid.UUID) ([]domain.AgencyProgram, error) {
	return s.agencies.ListPrograms(ctx, agencyID)
}

func (s *AgencyService) IsProgramEnabledForAgency(ctx context.Context, agencyID, programID uuid.UUID) (bool, error) {
	return s.agencies.IsProgramEnabledForAgency(ctx, agencyID, programID)
}

type WorkflowService struct {
	workflow *postgres.WorkflowRepository
}

func NewWorkflowService(workflow *postgres.WorkflowRepository) *WorkflowService {
	return &WorkflowService{workflow: workflow}
}

func (s *WorkflowService) ListTransitions(ctx context.Context, agencyID uuid.UUID) ([]domain.WorkflowTransition, error) {
	return s.workflow.ListTransitions(ctx, agencyID)
}
