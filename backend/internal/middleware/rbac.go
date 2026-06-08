package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		userRoles, _ := c.Get(ContextRolesKey)
		roleList, _ := userRoles.([]string)
		for _, r := range roleList {
			if _, ok := allowed[r]; ok {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}

func RequireAgencyRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		agencyRole, _ := c.Get(ContextAgencyRoleKey)
		role, _ := agencyRole.(string)
		if _, ok := allowed[role]; ok {
			c.Next()
			return
		}
		userRoles, _ := c.Get(ContextRolesKey)
		if roleList, ok := userRoles.([]string); ok {
			for _, r := range roleList {
				if r == "admin" {
					c.Next()
					return
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient agency role"})
	}
}
