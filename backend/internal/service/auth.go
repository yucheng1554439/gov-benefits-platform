package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	jwtpkg "github.com/govbenefits/platform/pkg/jwt"
	"github.com/govbenefits/platform/internal/repository/postgres"
	"golang.org/x/crypto/bcrypt"
)

// Demo password for seed users: Password123!
// bcrypt hash (cost 12): $2a$12$5pDrK0SKqeHdwLh4lkjlv.zZYeHOJegXzaHS/8K1wGgkHq7WDT1i.

type AuthService struct {
	users *postgres.UserRepository
	jwt   *jwtpkg.Manager
}

func NewAuthService(users *postgres.UserRepository, jwt *jwtpkg.Manager) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     string
	AgencyID  uuid.UUID
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    string    `json:"expires_at"`
	User         domain.AuthUser `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*TokenPair, error) {
	existing, err := s.users.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.users.Create(ctx, input.Email, string(hash))
	if err != nil {
		return nil, err
	}
	if err := s.users.CreateProfile(ctx, user.ID, input.FirstName, input.LastName, input.Phone, nil); err != nil {
		return nil, err
	}
	if err := s.users.AssignRole(ctx, user.ID, "citizen"); err != nil {
		return nil, err
	}
	if input.AgencyID != uuid.Nil {
		if err := s.users.LinkAgency(ctx, input.AgencyID, user.ID, "citizen", true); err != nil {
			return nil, err
		}
	}
	return s.issueTokens(ctx, user.ID)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	return s.issueTokens(ctx, user.ID)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.jwt.ValidateRefresh(refreshToken)
	if err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, claims.UserID)
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*domain.AuthUser, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}
	profile, _ := s.users.GetProfile(ctx, userID)
	roles, _ := s.users.GetRoles(ctx, userID)
	agency, _ := s.users.GetPrimaryAgency(ctx, userID)

	authUser := &domain.AuthUser{
		User:  *user,
		Roles: roles,
	}
	if profile != nil {
		authUser.Profile = profile
	}
	if agency != nil {
		authUser.AgencyID = agency.AgencyID
		authUser.AgencyRole = agency.AgencyRole
	}
	return authUser, nil
}

func (s *AuthService) issueTokens(ctx context.Context, userID uuid.UUID) (*TokenPair, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}
	roles, _ := s.users.GetRoles(ctx, userID)
	agency, _ := s.users.GetPrimaryAgency(ctx, userID)

	agencyID := uuid.Nil
	agencyRole := ""
	if agency != nil {
		agencyID = agency.AgencyID
		agencyRole = agency.AgencyRole
	}

	access, expiresAt, err := s.jwt.GenerateAccessToken(userID, user.Email, roles, agencyID, agencyRole)
	if err != nil {
		return nil, err
	}
	refresh, _, err := s.jwt.GenerateRefreshToken(userID, user.Email, roles, agencyID, agencyRole)
	if err != nil {
		return nil, err
	}

	authUser, _ := s.Me(ctx, userID)
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt.Format("2006-01-02T15:04:05Z07:00"),
		User:         *authUser,
	}, nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hash), err
}
