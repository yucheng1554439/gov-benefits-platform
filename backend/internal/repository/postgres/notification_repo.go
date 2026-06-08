package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/events"
)

type NotificationRepository struct {
	db *DB
}

func NewNotificationRepository(db *DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO notifications (agency_id, user_id, channel, event_type, title, body)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, is_read, created_at
	`, n.AgencyID, n.UserID, n.Channel, n.EventType, n.Title, n.Body).
		Scan(&n.ID, &n.IsRead, &n.CreatedAt)
}

func (r *NotificationRepository) CreateFromEvent(ctx context.Context, event events.Event) error {
	userID := uuid.Nil
	if event.ActorID != nil {
		userID = *event.ActorID
	}
	if uid, ok := event.Payload["citizen_id"].(string); ok {
		if parsed, err := uuid.Parse(uid); err == nil {
			userID = parsed
		}
	}
	if uid, ok := event.Payload["user_id"].(string); ok {
		if parsed, err := uuid.Parse(uid); err == nil {
			userID = parsed
		}
	}
	if userID == uuid.Nil {
		userID = event.EntityID
	}
	n := &domain.Notification{
		AgencyID:  event.AgencyID,
		UserID:    userID,
		Channel:   "in_app",
		EventType: string(event.Type),
		Title:     fmt.Sprintf("Event: %s", event.Type),
		Body:      fmt.Sprintf("Entity %s updated", event.EntityID),
	}
	return r.Create(ctx, n)
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, unreadOnly bool) ([]domain.Notification, error) {
	query := `
		SELECT id, agency_id, user_id, channel, event_type, title, COALESCE(body,''), is_read, created_at
		FROM notifications WHERE user_id = $1
	`
	if unreadOnly {
		query += " AND is_read = false"
	}
	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.AgencyID, &n.UserID, &n.Channel, &n.EventType, &n.Title, &n.Body, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (r *NotificationRepository) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}
