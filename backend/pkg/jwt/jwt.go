package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	Roles      []string  `json:"roles"`
	AgencyID   uuid.UUID `json:"agency_id"`
	AgencyRole string    `json:"agency_role"`
	TokenType  string    `json:"token_type"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *Manager) GenerateAccessToken(userID uuid.UUID, email string, roles []string, agencyID uuid.UUID, agencyRole string) (string, time.Time, error) {
	return m.generate(userID, email, roles, agencyID, agencyRole, "access", m.accessTTL)
}

func (m *Manager) GenerateRefreshToken(userID uuid.UUID, email string, roles []string, agencyID uuid.UUID, agencyRole string) (string, time.Time, error) {
	return m.generate(userID, email, roles, agencyID, agencyRole, "refresh", m.refreshTTL)
}

func (m *Manager) generate(userID uuid.UUID, email string, roles []string, agencyID uuid.UUID, agencyRole, tokenType string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)
	claims := Claims{
		UserID:     userID,
		Email:      email,
		Roles:      roles,
		AgencyID:   agencyID,
		AgencyRole: agencyRole,
		TokenType:  tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

func (m *Manager) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

func (m *Manager) ValidateAccess(tokenString string) (*Claims, error) {
	claims, err := m.Validate(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type: expected access")
	}
	return claims, nil
}

func (m *Manager) ValidateRefresh(tokenString string) (*Claims, error) {
	claims, err := m.Validate(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type: expected refresh")
	}
	return claims, nil
}
