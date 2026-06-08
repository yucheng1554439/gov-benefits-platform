package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	var u domain.User
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash) VALUES ($1, $2)
		RETURNING id, email, password_hash, status, created_at, updated_at
	`, email, passwordHash).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, status, created_at, updated_at FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, status, created_at, updated_at FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	var p domain.UserProfile
	var address []byte
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, first_name, last_name, COALESCE(phone,''), COALESCE(ssn_hash,''), address, created_at
		FROM user_profiles WHERE user_id = $1
	`, userID).Scan(&p.ID, &p.UserID, &p.FirstName, &p.LastName, &p.Phone, &p.SSNHash, &address, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get profile: %w", err)
	}
	p.Address = decodeJSONMap(address)
	return &p, nil
}

func (r *UserRepository) CreateProfile(ctx context.Context, userID uuid.UUID, firstName, lastName, phone string, address map[string]any) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO user_profiles (user_id, first_name, last_name, phone, address)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, firstName, lastName, phone, encodeJSON(address))
	return err
}

func (r *UserRepository) GetRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT r.name FROM roles r JOIN user_roles ur ON ur.role_id = r.id WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *UserRepository) AssignRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE name = $2
		ON CONFLICT DO NOTHING
	`, userID, roleName)
	return err
}

func (r *UserRepository) GetPrimaryAgency(ctx context.Context, userID uuid.UUID) (*domain.AgencyUser, error) {
	var au domain.AgencyUser
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, agency_id, user_id, agency_role, is_primary
		FROM agency_users WHERE user_id = $1 AND is_primary = true LIMIT 1
	`, userID).Scan(&au.ID, &au.AgencyID, &au.UserID, &au.AgencyRole, &au.IsPrimary)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &au, nil
}

func (r *UserRepository) LinkAgency(ctx context.Context, agencyID, userID uuid.UUID, agencyRole string, isPrimary bool) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO agency_users (agency_id, user_id, agency_role, is_primary)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (agency_id, user_id) DO NOTHING
	`, agencyID, userID, agencyRole, isPrimary)
	return err
}

func (r *UserRepository) GetAgencyMembership(ctx context.Context, userID, agencyID uuid.UUID) (*domain.AgencyUser, error) {
	var au domain.AgencyUser
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, agency_id, user_id, agency_role, is_primary
		FROM agency_users WHERE user_id = $1 AND agency_id = $2
	`, userID, agencyID).Scan(&au.ID, &au.AgencyID, &au.UserID, &au.AgencyRole, &au.IsPrimary)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &au, nil
}

func (r *UserRepository) GetDisplayNames(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	names := make(map[uuid.UUID]string)
	if len(userIDs) == 0 {
		return names, nil
	}
	rows, err := r.db.Pool.Query(ctx, `
		SELECT user_id, COALESCE(first_name || ' ' || last_name, '')
		FROM user_profiles WHERE user_id = ANY($1)
	`, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		names[id] = name
	}
	return names, rows.Err()
}
