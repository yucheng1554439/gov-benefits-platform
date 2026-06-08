package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/govbenefits/platform/pkg/jwt"
)

const (
	ContextUserIDKey     = "user_id"
	ContextEmailKey      = "email"
	ContextRolesKey      = "roles"
	ContextAgencyIDKey   = "agency_id"
	ContextAgencyRoleKey = "agency_role"
	ContextClaimsKey     = "claims"
)

type AuthMiddleware struct {
	jwt *jwtpkg.Manager
}

func NewAuthMiddleware(jwt *jwtpkg.Manager) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearer(c.GetHeader("Authorization"))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session expired. Please sign in again."})
			return
		}
		claims, err := m.jwt.ValidateAccess(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session expired. Please sign in again."})
			return
		}
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextEmailKey, claims.Email)
		c.Set(ContextRolesKey, claims.Roles)
		c.Set(ContextAgencyIDKey, claims.AgencyID)
		c.Set(ContextAgencyRoleKey, claims.AgencyRole)
		c.Set(ContextClaimsKey, claims)
		c.Next()
	}
}

func extractBearer(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
