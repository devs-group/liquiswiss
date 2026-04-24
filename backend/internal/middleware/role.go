package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Role hierarchy: owner > admin > editor > read-only
const (
	RoleOwner    = "owner"
	RoleAdmin    = "admin"
	RoleEditor   = "editor"
	RoleReadOnly = "read-only"
)

func roleRank(role string) int {
	switch role {
	case RoleOwner:
		return 3
	case RoleAdmin:
		return 2
	case RoleEditor:
		return 1
	case RoleReadOnly:
		return 0
	default:
		return -1
	}
}

// RequireMinRole blocks requests whose authenticated user has a role in their
// current organisation below minRole. AuthMiddleware must run first so userID
// is set in the context.
func RequireMinRole(minRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("userID")
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Nicht angemeldet"})
			return
		}

		role, err := databaseService.GetCurrentUserRole(userID)
		if err != nil || role == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Keine Berechtigung für diese Organisation"})
			return
		}

		if roleRank(role) < roleRank(minRole) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Ihre Rolle erlaubt diese Aktion nicht"})
			return
		}

		c.Set("currentOrgRole", role)
		c.Next()
	}
}
