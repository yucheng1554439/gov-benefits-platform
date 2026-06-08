package events

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventCaseCreated       EventType = "case.created"
	EventCaseStatusChanged EventType = "case.status_changed"
	EventApplicationCreated EventType = "application.created"
	EventDocumentUploaded  EventType = "document.uploaded"
	EventEligibilityEvaluated EventType = "eligibility.evaluated"
	EventBenefitCalculated EventType = "benefit.calculated"
	EventFraudFlagged      EventType = "fraud.flagged"
	EventAppealFiled       EventType = "appeal.filed"
	EventAppealDecided     EventType = "appeal.decided"
	EventLetterGenerated   EventType = "letter.generated"
	EventSLABreached       EventType = "sla.breached"
)

type Event struct {
	ID        uuid.UUID      `json:"id"`
	Type      EventType      `json:"type"`
	AgencyID  uuid.UUID      `json:"agency_id"`
	ActorID   *uuid.UUID     `json:"actor_id,omitempty"`
	EntityID  uuid.UUID      `json:"entity_id"`
	Payload   map[string]any `json:"payload"`
	Timestamp time.Time      `json:"timestamp"`
}

type Handler func(ctx context.Context, event Event) error

type Bus struct {
	mu       sync.RWMutex
	handlers map[EventType][]Handler
	log      *slog.Logger
}

func NewBus(log *slog.Logger) *Bus {
	return &Bus{
		handlers: make(map[EventType][]Handler),
		log:      log,
	}
}

func (b *Bus) Subscribe(eventType EventType, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *Bus) SubscribeAll(handler Handler) {
	for _, t := range []EventType{
		EventCaseCreated, EventCaseStatusChanged, EventApplicationCreated,
		EventDocumentUploaded, EventEligibilityEvaluated, EventBenefitCalculated,
		EventFraudFlagged, EventAppealFiled, EventAppealDecided, EventLetterGenerated, EventSLABreached,
	} {
		b.Subscribe(t, handler)
	}
}

func (b *Bus) Publish(ctx context.Context, event Event) {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	b.mu.RLock()
	handlers := append([]Handler{}, b.handlers[event.Type]...)
	b.mu.RUnlock()

	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			b.log.Error("event handler failed",
				"event_type", event.Type,
				"event_id", event.ID,
				"error", err,
			)
		}
	}
}

func NewEvent(eventType EventType, agencyID, entityID uuid.UUID, actorID *uuid.UUID, payload map[string]any) Event {
	return Event{
		ID:        uuid.New(),
		Type:      eventType,
		AgencyID:  agencyID,
		ActorID:   actorID,
		EntityID:  entityID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
}
