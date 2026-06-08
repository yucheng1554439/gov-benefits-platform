package subscribers

import (
	"context"
	"log/slog"

	"github.com/govbenefits/platform/internal/events"
	"github.com/govbenefits/platform/internal/repository/postgres"
)

type AuditSubscriber struct {
	repo *postgres.AuditRepository
	log  *slog.Logger
}

func NewAuditSubscriber(repo *postgres.AuditRepository, log *slog.Logger) *AuditSubscriber {
	return &AuditSubscriber{repo: repo, log: log}
}

func (s *AuditSubscriber) Handle(ctx context.Context, event events.Event) error {
	s.log.Debug("audit subscriber", "event_type", event.Type, "entity_id", event.EntityID)
	return s.repo.CreateFromEvent(ctx, event)
}
