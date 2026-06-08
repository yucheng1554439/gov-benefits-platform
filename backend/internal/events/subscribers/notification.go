package subscribers

import (
	"context"
	"log/slog"

	"github.com/govbenefits/platform/internal/events"
	"github.com/govbenefits/platform/internal/repository/postgres"
)

type NotificationSubscriber struct {
	repo *postgres.NotificationRepository
	log  *slog.Logger
}

func NewNotificationSubscriber(repo *postgres.NotificationRepository, log *slog.Logger) *NotificationSubscriber {
	return &NotificationSubscriber{repo: repo, log: log}
}

func (s *NotificationSubscriber) Handle(ctx context.Context, event events.Event) error {
	s.log.Debug("notification subscriber", "event_type", event.Type)
	return s.repo.CreateFromEvent(ctx, event)
}
