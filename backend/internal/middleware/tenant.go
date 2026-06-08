package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TenantContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		if agencyID, exists := c.Get(ContextAgencyIDKey); exists {
			if id, ok := agencyID.(uuid.UUID); ok && id != uuid.Nil {
				c.Set("tenant_agency_id", id)
			}
		}
		if userID, exists := c.Get(ContextUserIDKey); exists {
			if id, ok := userID.(uuid.UUID); ok {
				c.Set("tenant_user_id", id)
			}
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) uuid.UUID {
	if v, ok := c.Get(ContextUserIDKey); ok {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

func GetAgencyID(c *gin.Context) uuid.UUID {
	if v, ok := c.Get(ContextAgencyIDKey); ok {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

func GetRoles(c *gin.Context) []string {
	if v, ok := c.Get(ContextRolesKey); ok {
		if roles, ok := v.([]string); ok {
			return roles
		}
	}
	return nil
}

func GetAgencyRole(c *gin.Context) string {
	if v, ok := c.Get(ContextAgencyRoleKey); ok {
		if role, ok := v.(string); ok {
			return role
		}
	}
	return ""
}
