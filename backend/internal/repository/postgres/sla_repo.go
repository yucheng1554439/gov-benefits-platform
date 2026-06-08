package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/jackc/pgx/v5"
)

type FeatureFlagRepository struct {
	db *DB
}

func NewFeatureFlagRepository(db *DB) *FeatureFlagRepository {
	return &FeatureFlagRepository{db: db}
}

func (r *FeatureFlagRepository) Get(ctx context.Context, agencyID uuid.UUID, flagKey string) (*domain.FeatureFlag, error) {
	var f domain.FeatureFlag
	var metadata []byte
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, agency_id, flag_key, is_enabled, rollout_pct, metadata, updated_at
		FROM feature_flags WHERE agency_id = $1 AND flag_key = $2
	`, agencyID, flagKey).Scan(&f.ID, &f.AgencyID, &f.FlagKey, &f.IsEnabled, &f.RolloutPct, &metadata, &f.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	f.Metadata = decodeJSONMap(metadata)
	return &f, nil
}

func (r *FeatureFlagRepository) List(ctx context.Context, agencyID uuid.UUID) ([]domain.FeatureFlag, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, agency_id, flag_key, is_enabled, rollout_pct, metadata, updated_at
		FROM feature_flags WHERE agency_id = $1 ORDER BY flag_key
	`, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []domain.FeatureFlag
	for rows.Next() {
		var f domain.FeatureFlag
		var metadata []byte
		if err := rows.Scan(&f.ID, &f.AgencyID, &f.FlagKey, &f.IsEnabled, &f.RolloutPct, &metadata, &f.UpdatedAt); err != nil {
			return nil, err
		}
		f.Metadata = decodeJSONMap(metadata)
		flags = append(flags, f)
	}
	return flags, rows.Err()
}

func (r *FeatureFlagRepository) Upsert(ctx context.Context, flag *domain.FeatureFlag) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO feature_flags (agency_id, flag_key, is_enabled, rollout_pct, metadata, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (agency_id, flag_key) DO UPDATE SET
			is_enabled = EXCLUDED.is_enabled,
			rollout_pct = EXCLUDED.rollout_pct,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
		RETURNING id, updated_at
	`, flag.AgencyID, flag.FlagKey, flag.IsEnabled, flag.RolloutPct, encodeJSON(flag.Metadata)).
		Scan(&flag.ID, &flag.UpdatedAt)
}

type SLARepository struct {
	db *DB
}

func NewSLARepository(db *DB) *SLARepository {
	return &SLARepository{db: db}
}

func (r *SLARepository) GetPolicy(ctx context.Context, agencyID, programID uuid.UUID) (*domain.SLAPolicy, error) {
	var p domain.SLAPolicy
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, agency_id, program_id, target_days, warning_threshold_pct, business_days_only
		FROM sla_policies WHERE agency_id = $1 AND program_id = $2
	`, agencyID, programID).Scan(&p.ID, &p.AgencyID, &p.ProgramID, &p.TargetDays, &p.WarningThresholdPct, &p.BusinessDaysOnly)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *SLARepository) CreateTracking(ctx context.Context, tracking *domain.CaseSLATracking) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO case_sla_tracking (case_id, sla_policy_id, due_at, status, elapsed_days)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (case_id) DO NOTHING
		RETURNING id
	`, tracking.CaseID, tracking.SLAPolicyID, tracking.DueAt, tracking.Status, tracking.ElapsedDays).
		Scan(&tracking.ID)
}

func (r *SLARepository) GetTracking(ctx context.Context, caseID uuid.UUID) (*domain.CaseSLATracking, error) {
	var t domain.CaseSLATracking
	var breachedAt *time.Time
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, case_id, sla_policy_id, due_at, status, elapsed_days, breached_at
		FROM case_sla_tracking WHERE case_id = $1
	`, caseID).Scan(&t.ID, &t.CaseID, &t.SLAPolicyID, &t.DueAt, &t.Status, &t.ElapsedDays, &breachedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	t.BreachedAt = breachedAt
	return &t, nil
}

func (r *SLARepository) EnsureTrackingForCase(ctx context.Context, caseID uuid.UUID) error {
	var agencyID, programID uuid.UUID
	var submittedAt time.Time
	err := r.db.Pool.QueryRow(ctx, `
		SELECT agency_id, program_id, submitted_at FROM cases WHERE id = $1
	`, caseID).Scan(&agencyID, &programID, &submittedAt)
	if err != nil {
		return err
	}

	policy, err := r.GetPolicy(ctx, agencyID, programID)
	if err != nil || policy == nil {
		return err
	}

	dueAt := submittedAt.AddDate(0, 0, policy.TargetDays)
	tracking := &domain.CaseSLATracking{
		CaseID:      caseID,
		SLAPolicyID: policy.ID,
		DueAt:       dueAt,
		Status:      "on_track",
		ElapsedDays: 0,
	}
	return r.CreateTracking(ctx, tracking)
}

func (r *SLARepository) UpdateTrackingStatus(ctx context.Context, caseID uuid.UUID) error {
	tracking, err := r.GetTracking(ctx, caseID)
	if err != nil || tracking == nil {
		return err
	}

	now := time.Now()
	status := "on_track"
	if now.After(tracking.DueAt) {
		status = "breached"
		_, err = r.db.Pool.Exec(ctx, `
			UPDATE case_sla_tracking SET status = $2, breached_at = COALESCE(breached_at, NOW()),
				elapsed_days = EXTRACT(DAY FROM NOW() - (due_at - ($3 || ' days')::interval))::int
			WHERE case_id = $1
		`, caseID, status, tracking.ElapsedDays)
		return err
	}

	_, err = r.db.Pool.Exec(ctx, `UPDATE case_sla_tracking SET status = $2 WHERE case_id = $1`, caseID, status)
	return err
}

func (r *SLARepository) ListBreached(ctx context.Context, agencyID uuid.UUID) ([]domain.CaseSLATracking, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT cst.id, cst.case_id, cst.sla_policy_id, cst.due_at, cst.status, cst.elapsed_days, cst.breached_at
		FROM case_sla_tracking cst
		JOIN cases c ON c.id = cst.case_id
		WHERE c.agency_id = $1 AND cst.status = 'breached'
	`, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trackings []domain.CaseSLATracking
	for rows.Next() {
		var t domain.CaseSLATracking
		var breachedAt *time.Time
		if err := rows.Scan(&t.ID, &t.CaseID, &t.SLAPolicyID, &t.DueAt, &t.Status, &t.ElapsedDays, &breachedAt); err != nil {
			return nil, err
		}
		t.BreachedAt = breachedAt
		trackings = append(trackings, t)
	}
	return trackings, rows.Err()
}

type WorkerRepository struct {
	db *DB
}

func NewWorkerRepository(db *DB) *WorkerRepository {
	return &WorkerRepository{db: db}
}

func (r *WorkerRepository) ListAvailable(ctx context.Context, agencyID uuid.UUID) ([]domain.WorkerProfile, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT user_id, agency_id, specializations, max_active_cases, current_case_count
		FROM worker_profiles
		WHERE agency_id = $1 AND current_case_count < max_active_cases
		ORDER BY current_case_count ASC
	`, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workers []domain.WorkerProfile
	for rows.Next() {
		var w domain.WorkerProfile
		if err := rows.Scan(&w.UserID, &w.AgencyID, &w.Specializations, &w.MaxActiveCases, &w.CurrentCaseCount); err != nil {
			return nil, err
		}
		workers = append(workers, w)
	}
	return workers, rows.Err()
}

func (r *WorkerRepository) IncrementCaseCount(ctx context.Context, workerID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE worker_profiles SET current_case_count = current_case_count + 1 WHERE user_id = $1
	`, workerID)
	return err
}

func (r *WorkerRepository) GetProgramCode(ctx context.Context, programID uuid.UUID) (string, error) {
	var code string
	err := r.db.Pool.QueryRow(ctx, `SELECT code FROM programs WHERE id = $1`, programID).Scan(&code)
	return code, err
}
