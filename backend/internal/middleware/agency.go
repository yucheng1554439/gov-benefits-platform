package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/repository/postgres"
)

type AgencyResolver struct {
	users *postgres.UserRepository
}

func NewAgencyResolver(users *postgres.UserRepository) *AgencyResolver {
	return &AgencyResolver{users: users}
}

func (m *AgencyResolver) ResolveAgency() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("X-Agency-ID")
		if header == "" {
			c.Next()
			return
		}

		agencyID, err := uuid.Parse(header)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid agency identifier."})
			return
		}

		userID := GetUserID(c)
		if userID == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session expired. Please sign in again."})
			return
		}

		membership, err := m.users.GetAgencyMembership(c.Request.Context(), userID, agencyID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to verify agency access."})
			return
		}
		if membership == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You do not have access to this agency."})
			return
		}

		c.Set(ContextAgencyIDKey, agencyID)
		c.Set(ContextAgencyRoleKey, membership.AgencyRole)
		c.Next()
	}
}
