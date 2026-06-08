package middleware

import "github.com/gin-gonic/gin"

func ContainsRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// ResolveWorkflowRole maps agency-scoped roles to workflow transition roles.
func ResolveWorkflowRole(c *gin.Context) string {
	switch GetAgencyRole(c) {
	case "worker":
		return "case_worker"
	case "admin", "supervisor", "case_worker", "citizen":
		return GetAgencyRole(c)
	}

	roles := GetRoles(c)
	for _, preferred := range []string{"admin", "supervisor", "case_worker", "citizen"} {
		if ContainsRole(roles, preferred) {
			return preferred
		}
	}
	if len(roles) > 0 {
		return roles[0]
	}
	return GetAgencyRole(c)
}