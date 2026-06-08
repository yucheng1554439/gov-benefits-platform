package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/govbenefits/platform/internal/featureflags"
)

func FeatureFlagGuard(flags *featureflags.Service, flagKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		agencyID := GetAgencyID(c)
		userID := GetUserID(c)
		enabled, err := flags.IsEnabled(c.Request.Context(), agencyID, flagKey, userID)
		if err != nil || !enabled {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "feature not available"})
			return
		}
		c.Next()
	}
}
