package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/jackc/pgx/v5"
)

type WorkflowRepository struct {
	db *DB
}

func NewWorkflowRepository(db *DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

func (r *WorkflowRepository) GetTransition(ctx context.Context, agencyID uuid.UUID, fromStatus, toStatus string) (*domain.WorkflowTransition, error) {
	var t domain.WorkflowTransition
	var agencyIDVal *uuid.UUID
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, agency_id, from_status, to_status, required_role
		FROM workflow_transitions
		WHERE (agency_id = $1 OR agency_id IS NULL) AND from_status = $2 AND to_status = $3
		ORDER BY agency_id NULLS LAST LIMIT 1
	`, agencyID, fromStatus, toStatus).Scan(&t.ID, &agencyIDVal, &t.FromStatus, &t.ToStatus, &t.RequiredRole)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	t.AgencyID = agencyIDVal
	return &t, nil
}

func (r *WorkflowRepository) RecordEvent(ctx context.Context, tx pgx.Tx, event *domain.WorkflowEvent) error {
	return tx.QueryRow(ctx, `
		INSERT INTO workflow_events (agency_id, case_id, from_status, to_status, actor_id, reason)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, event.AgencyID, event.CaseID, event.FromStatus, event.ToStatus, event.ActorID, event.Reason).
		Scan(&event.ID, &event.CreatedAt)
}

func (r *WorkflowRepository) ListEvents(ctx context.Context, tx pgx.Tx, caseID uuid.UUID) ([]domain.WorkflowEvent, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, agency_id, case_id, COALESCE(from_status,''), to_status, actor_id, COALESCE(reason,''), created_at
		FROM workflow_events WHERE case_id = $1 ORDER BY created_at DESC
	`, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.WorkflowEvent
	for rows.Next() {
		var e domain.WorkflowEvent
		if err := rows.Scan(&e.ID, &e.AgencyID, &e.CaseID, &e.FromStatus, &e.ToStatus, &e.ActorID, &e.Reason, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *WorkflowRepository) ListTransitions(ctx context.Context, agencyID uuid.UUID) ([]domain.WorkflowTransition, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, agency_id, from_status, to_status, required_role
		FROM workflow_transitions WHERE agency_id = $1 OR agency_id IS NULL
	`, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transitions []domain.WorkflowTransition
	for rows.Next() {
		var t domain.WorkflowTransition
		if err := rows.Scan(&t.ID, &t.AgencyID, &t.FromStatus, &t.ToStatus, &t.RequiredRole); err != nil {
			return nil, err
		}
		transitions = append(transitions, t)
	}
	return transitions, rows.Err()
}

func (r *WorkflowRepository) GetLatestEventTime(ctx context.Context, caseID uuid.UUID) (*time.Time, error) {
	var t time.Time
	err := r.db.Pool.QueryRow(ctx, `
		SELECT created_at FROM workflow_events WHERE case_id = $1 ORDER BY created_at DESC LIMIT 1
	`, caseID).Scan(&t)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}
