package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/events"
)

type AuditListFilter struct {
	Action string
	Search string
	Offset int
	Limit  int
}

type AuditRepository struct {
	db *DB
}

func NewAuditRepository(db *DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO audit_logs (agency_id, actor_id, action, entity_type, entity_id, previous_state, new_state, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`, log.AgencyID, log.ActorID, log.Action, log.EntityType, log.EntityID, encodeJSON(log.PreviousState), encodeJSON(log.NewState), log.IPAddress).
		Scan(&log.ID, &log.CreatedAt)
}

func (r *AuditRepository) CreateFromEvent(ctx context.Context, event events.Event) error {
	log := &domain.AuditLog{
		AgencyID:   &event.AgencyID,
		ActorID:    event.ActorID,
		Action:     string(event.Type),
		EntityType: "event",
		EntityID:   &event.EntityID,
		NewState:   event.Payload,
	}
	return r.Create(ctx, log)
}

func (r *AuditRepository) List(ctx context.Context, agencyID string, filter AuditListFilter) ([]domain.AuditLog, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}

	args := []any{agencyID}
	where := []string{"(al.agency_id::text = $1 OR $1 = '')"}

	if filter.Action != "" {
		args = append(args, filter.Action)
		where = append(where, fmt.Sprintf("al.action = $%d", len(args)))
	}
	if filter.Search != "" {
		args = append(args, "%"+strings.ToLower(filter.Search)+"%")
		idx := len(args)
		where = append(where, fmt.Sprintf(`(
			LOWER(al.action) LIKE $%d OR
			LOWER(al.entity_type) LIKE $%d OR
			LOWER(COALESCE(up.first_name || ' ' || up.last_name, '')) LIKE $%d OR
			LOWER(al.new_state::text) LIKE $%d
		)`, idx, idx, idx, idx))
	}

	args = append(args, filter.Limit, filter.Offset)
	limitIdx := len(args) - 1
	offsetIdx := len(args)

	query := fmt.Sprintf(`
		SELECT al.id, al.agency_id, al.actor_id, al.action, al.entity_type, al.entity_id,
		       al.previous_state, al.new_state, COALESCE(al.ip_address,''), al.created_at,
		       COALESCE(up.first_name || ' ' || up.last_name, '')
		FROM audit_logs al
		LEFT JOIN user_profiles up ON up.user_id = al.actor_id
		WHERE %s
		ORDER BY al.created_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(where, " AND "), limitIdx, offsetIdx)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		var prev, newState []byte
		if err := rows.Scan(&l.ID, &l.AgencyID, &l.ActorID, &l.Action, &l.EntityType, &l.EntityID, &prev, &newState, &l.IPAddress, &l.CreatedAt, &l.ActorName); err != nil {
			return nil, err
		}
		l.PreviousState = decodeJSONMap(prev)
		l.NewState = decodeJSONMap(newState)
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (r *AuditRepository) Count(ctx context.Context, agencyID string, filter AuditListFilter) (int, error) {
	args := []any{agencyID}
	where := []string{"(al.agency_id::text = $1 OR $1 = '')"}

	if filter.Action != "" {
		args = append(args, filter.Action)
		where = append(where, fmt.Sprintf("al.action = $%d", len(args)))
	}
	if filter.Search != "" {
		args = append(args, "%"+strings.ToLower(filter.Search)+"%")
		idx := len(args)
		where = append(where, fmt.Sprintf(`(
			LOWER(al.action) LIKE $%d OR
			LOWER(al.entity_type) LIKE $%d OR
			LOWER(COALESCE(up.first_name || ' ' || up.last_name, '')) LIKE $%d OR
			LOWER(al.new_state::text) LIKE $%d
		)`, idx, idx, idx, idx))
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM audit_logs al
		LEFT JOIN user_profiles up ON up.user_id = al.actor_id
		WHERE %s
	`, strings.Join(where, " AND "))

	var count int
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}
