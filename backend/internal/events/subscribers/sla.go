package subscribers

import (
	"context"
	"log/slog"

	"github.com/govbenefits/platform/internal/events"
	"github.com/govbenefits/platform/internal/repository/postgres"
)

type SLASubscriber struct {
	repo *postgres.SLARepository
	log  *slog.Logger
}

func NewSLASubscriber(repo *postgres.SLARepository, log *slog.Logger) *SLASubscriber {
	return &SLASubscriber{repo: repo, log: log}
}

func (s *SLASubscriber) Handle(ctx context.Context, event events.Event) error {
	switch event.Type {
	case events.EventCaseCreated:
		caseID := event.EntityID
		if cid, ok := event.Payload["case_id"].(string); ok {
			_ = cid
		}
		s.log.Debug("sla subscriber: tracking case", "case_id", caseID)
		return s.repo.EnsureTrackingForCase(ctx, caseID)
	case events.EventCaseStatusChanged:
		s.log.Debug("sla subscriber: status changed", "entity_id", event.EntityID)
		return s.repo.UpdateTrackingStatus(ctx, event.EntityID)
	default:
		return nil
	}
}
