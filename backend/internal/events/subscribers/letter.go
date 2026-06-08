package subscribers

import (
	"context"
	"log/slog"

	"github.com/govbenefits/platform/internal/events"
	"github.com/govbenefits/platform/internal/jobs"
)

type LetterSubscriber struct {
	queue *jobs.Queue
	log   *slog.Logger
}

func NewLetterSubscriber(queue *jobs.Queue, log *slog.Logger) *LetterSubscriber {
	return &LetterSubscriber{queue: queue, log: log}
}

func (s *LetterSubscriber) Handle(ctx context.Context, event events.Event) error {
	if event.Type != events.EventCaseStatusChanged {
		return nil
	}
	toStatus, _ := event.Payload["to_status"].(string)
	if toStatus != "approved" && toStatus != "denied" {
		return nil
	}
	s.log.Debug("letter subscriber: enqueue letter job", "case_id", event.EntityID, "status", toStatus)
	return s.queue.Enqueue(ctx, jobs.Job{
		Type: jobs.JobGenerateLetter,
		Payload: map[string]any{
			"case_id":     event.EntityID.String(),
			"agency_id":   event.AgencyID.String(),
			"letter_type": letterTypeForStatus(toStatus),
		},
	})
}

func letterTypeForStatus(status string) string {
	switch status {
	case "approved":
		return "approval_notice"
	case "denied":
		return "denial_notice"
	default:
		return "status_update"
	}
}
