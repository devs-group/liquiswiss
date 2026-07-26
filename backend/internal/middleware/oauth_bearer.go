package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"liquiswiss/config"
	"liquiswiss/pkg/auth"
)

const mcpAudience = "liquiswiss-mcp"

// OAuthBearerMiddleware protects the MCP endpoint. It only accepts JWT access
// tokens issued by the OAuth token endpoint (audience-bound), never web session
// cookies. 401 responses carry the resource metadata pointer required by the
// MCP authorization spec so clients can discover the authorization server.
func OAuthBearerMiddleware(c *gin.Context) {
	unauthorized := func() {
		cfg := config.GetConfig()
		base := strings.TrimSuffix(cfg.APIHost, "/")
		c.Header("WWW-Authenticate", `Bearer resource_metadata="`+base+`/.well-known/oauth-protected-resource"`)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}

	header := c.GetHeader("Authorization")
	tokenString, found := strings.CutPrefix(header, "Bearer ")
	if !found || tokenString == "" {
		unauthorized()
		return
	}

	claims, err := auth.VerifyToken(tokenString)
	if err != nil {
		unauthorized()
		return
	}
	if !slices.Contains(claims.Audience, mcpAudience) {
		unauthorized()
		return
	}

	exists, err := databaseService.CheckUserExistence(claims.UserID)
	if err != nil || !exists {
		unauthorized()
		return
	}

	c.Set("userID", claims.UserID)
	c.Next()
}
