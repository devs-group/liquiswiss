package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"liquiswiss/internal/service"
	"liquiswiss/pkg/auth"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
)

func Register(dbService service.IDatabaseService, c *gin.Context) {
	var payload models.Registration
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		// Return validation errors
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Eingabe"})
		return
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, err := dbService.RegisterUser(payload.Email, string(encryptedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Die Registrierung ist fehlgeschlagen"})
		return
	}

	user, err := dbService.GetProfile(fmt.Sprint(userId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	accessToken, accessExpirationTime, _, err := auth.GenerateAccessToken(*user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	refreshToken, tokenId, refreshExpirationTime, err := auth.GenerateRefreshToken(*user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Store the refresh token in the database
	deviceName := c.Request.UserAgent()
	err = dbService.StoreRefreshTokenID(userId, tokenId, refreshExpirationTime, deviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	accessTokenCookie := auth.GenerateCookie(utils.AccessTokenName, accessToken, accessExpirationTime)
	http.SetCookie(c.Writer, &accessTokenCookie)
	refreshTokenCookie := auth.GenerateCookie(utils.RefreshTokenName, refreshToken, refreshExpirationTime)
	http.SetCookie(c.Writer, &refreshTokenCookie)

	c.JSON(http.StatusOK, user)
}

func Login(dbService service.IDatabaseService, c *gin.Context) {
	var payload models.Login
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		// Return validation errors
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	loginUser, err := dbService.GetUserPasswordByEMail(payload.Email)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(loginUser.Password), []byte(payload.Password))
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	user, err := dbService.GetProfile(fmt.Sprint(loginUser.ID))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	accessToken, accessExpirationTime, _, err := auth.GenerateAccessToken(*user)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	refreshToken, tokenId, refreshExpirationTime, err := auth.GenerateRefreshToken(*user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Clear the current refresh token first from DB if some exists
	clearRefreshTokenFromDatabase(dbService, c)

	// Store the refresh token in the database
	deviceName := c.Request.UserAgent()
	err = dbService.StoreRefreshTokenID(loginUser.ID, tokenId, refreshExpirationTime, deviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	accessTokenCookie := auth.GenerateCookie(utils.AccessTokenName, accessToken, accessExpirationTime)
	http.SetCookie(c.Writer, &accessTokenCookie)
	refreshTokenCookie := auth.GenerateCookie(utils.RefreshTokenName, refreshToken, refreshExpirationTime)
	http.SetCookie(c.Writer, &refreshTokenCookie)

	c.JSON(http.StatusOK, user)
}

func Logout(dbService service.IDatabaseService, c *gin.Context) {
	clearRefreshTokenFromDatabase(dbService, c)
	auth.ClearAuthCookies(c)
	c.Status(http.StatusOK)
}

func clearRefreshTokenFromDatabase(dbService service.IDatabaseService, c *gin.Context) {
	refreshToken, err := c.Cookie(utils.RefreshTokenName)
	if err == nil && refreshToken != "" {
		// Verify the refresh token to get its claims (e.g., the tokenID and userID)
		refreshClaims, err2 := auth.VerifyToken(refreshToken)
		if err2 == nil {
			// Delete the refresh token from the database
			err2 = dbService.DeleteRefreshToken(refreshClaims.ID, refreshClaims.UserID)
			if err2 != nil {
				logger.Logger.Error("Failed to delete refresh token", err2)
			}
		}
	}
}
