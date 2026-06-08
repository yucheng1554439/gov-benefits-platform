package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	settingAgencyID = "app.current_agency_id"
	settingUserID   = "app.current_user_id"
)

func SetTenantContext(ctx context.Context, tx pgx.Tx, agencyID, userID uuid.UUID) error {
	if _, err := tx.Exec(ctx, fmt.Sprintf(`SET LOCAL %s = '%s'`, settingAgencyID, agencyID)); err != nil {
		return fmt.Errorf("set agency context: %w", err)
	}
	if _, err := tx.Exec(ctx, fmt.Sprintf(`SET LOCAL %s = '%s'`, settingUserID, userID)); err != nil {
		return fmt.Errorf("set user context: %w", err)
	}
	return nil
}

func WithTenant(ctx context.Context, pool *DB, agencyID, userID uuid.UUID, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := pool.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := SetTenantContext(ctx, tx, agencyID, userID); err != nil {
		return err
	}
	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
