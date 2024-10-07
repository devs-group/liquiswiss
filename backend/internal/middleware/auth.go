package middleware

import (
	"liquiswiss/internal/service"
	"liquiswiss/pkg/auth"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

var databaseService service.IDatabaseService

func AuthMiddleware(c *gin.Context) {
	var accessClaims *auth.Claims

	// Get AccessToken and if possible verify it
	accessToken, err := c.Cookie(utils.AccessTokenName)
	if err == nil {
		accessClaims, err = auth.VerifyToken(accessToken)
		if err != nil {
			accessClaims = nil
		}
	}

	if accessClaims == nil {
		// If access token is invalid, check if a refresh token exists
		refreshToken, err := c.Cookie(utils.RefreshTokenName)
		if err != nil || refreshToken == "" {
			// Delete both tokens and abort
			auth.ClearAuthCookies(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Nicht angemeldet, Kein Refresh Token vorhanden", "logout": true})
			return
		}

		// Verify the refresh token
		refreshClaims, err := auth.VerifyToken(refreshToken)
		if err != nil {
			// If the refresh token is invalid, clear cookies and abort
			auth.ClearAuthCookies(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Nicht angemeldet, Refresh Token ungültig", "logout": true})
			return
		}

		// Check if the refresh token is valid in the database
		valid, err := databaseService.CheckRefreshToken(refreshClaims.ID, refreshClaims.UserID)
		if err != nil || !valid {
			// If the refresh token is not valid or not found, delete both tokens and abort
			auth.ClearAuthCookies(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Nicht angemeldet, Refresh Token ungültig oder deaktiviert", "logout": true})
			return
		}

		// Generate a new access token since the refresh token is valid
		newAccessToken, accessExpirationTime, newAccessClaims, err := auth.GenerateAccessToken(models.User{ID: refreshClaims.UserID})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Erstellen eines neuen Access-Tokens", "logout": true})
			return
		}

		// Set the new access token as cookie
		cookie := auth.GenerateCookie(utils.AccessTokenName, newAccessToken, accessExpirationTime)
		http.SetCookie(c.Writer, &cookie)

		// Proceed with the refreshed access token's claims
		accessClaims = newAccessClaims
	}

	exists, err := databaseService.CheckUserExistance(accessClaims.UserID)
	if err != nil {
		auth.ClearAuthCookies(c)
		// TODO: Report as exception to Sentry
		logger.Logger.Error("Error checking user existence", err)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Fehler beim Überprüfen der Benutzerexistenz", "logout": true})
		return
	}
	if !exists {
		// If the user no longer exists, delete both tokens and abort
		auth.ClearAuthCookies(c)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Nicht erlaubt, Benutzer existiert nicht mehr", "logout": true})
		return
	}

	// Pass the user ID to the next middleware or handler
	c.Set("userID", accessClaims.UserID)
	c.Next()
}

func InjectUserService(s service.IDatabaseService) {
	databaseService = s
}
