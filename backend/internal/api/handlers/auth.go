package handlers

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/internal/service/sendgrid_service"
	"liquiswiss/pkg/auth"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
)

func CreateRegistration(dbService db_service.IDatabaseService, sendgridService sendgrid_service.ISendgridService, c *gin.Context) {
	var payload models.CreateRegistration
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Eingabe"})
		return
	}

	code := utils.GenerateUUID()

	registrationId, err := dbService.CreateRegistration(payload.Email, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Die Registrierung ist fehlgeschlagen"})
		return
	}

	// Make sure a registration can't exist if the email already exists for a user
	if registrationId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Registrierung mit dieser E-Mail ist nicht mehr möglich"})
		return
	}

	err = sendgridService.SendRegistrationMail(payload.Email, code)
	if err != nil {
		err := dbService.DeleteRegistration(registrationId, code)
		if err != nil {
			logger.Logger.Error("Couldn't delete registration: ", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Die Registrierung ist fehlgeschlagen"})
		return
	}

	c.Status(http.StatusOK)
}

// CheckRegistrationCode simply checks if the submitted data are valid
func CheckRegistrationCode(dbService db_service.IDatabaseService, c *gin.Context) {
	var payload models.CheckRegistrationCode
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Eingabe"})
		return
	}

	_, err := dbService.ValidateRegistration(payload.Email, payload.Code, utils.RegistrationCodeValidity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validierung der Registrierung ist fehlgeschlagen"})
		return
	}

	c.Status(http.StatusOK)
}

// FinishRegistration uses the (hopefully) valid data along with the password to create a full user
func FinishRegistration(dbService db_service.IDatabaseService, c *gin.Context) {
	var payload models.FinishRegistration
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Eingabe"})
		return
	}

	registrationId, err := dbService.ValidateRegistration(payload.Email, payload.Code, utils.RegistrationCodeValidity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validierung der Registrierung ist fehlgeschlagen"})
		return
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, err := dbService.CreateUser(payload.Email, string(encryptedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Die Registrierung ist fehlgeschlagen"})
		return
	}

	// Every new user gets an organisation assigned automatically
	organisationID, err := dbService.CreateOrganisation("Meine Organisation")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// We explicity set it as the default organisation which can only be deleted along with the users account
	err = dbService.AssignUserToOrganisation(userId, organisationID, "owner", true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set the new default organisation as the current one
	err = dbService.SetUserCurrentOrganisation(userId, organisationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete registration since we have the user now
	err = dbService.DeleteRegistration(registrationId, payload.Email)
	if err != nil {
		logger.Logger.Error("Error deleting registation: ", err)
		return
	}

	user, err := dbService.GetProfile(userId)
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

func Login(dbService db_service.IDatabaseService, c *gin.Context) {
	var payload models.Login
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
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

	user, err := dbService.GetProfile(loginUser.ID)
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

func Logout(dbService db_service.IDatabaseService, c *gin.Context) {
	clearRefreshTokenFromDatabase(dbService, c)
	auth.ClearAuthCookies(c)
	c.Status(http.StatusOK)
}

// ForgotPassword creates a password reset for the user
func ForgotPassword(dbService db_service.IDatabaseService, sendgridService sendgrid_service.ISendgridService, c *gin.Context) {
	var payload models.ForgotPassword
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Eingabe"})
		return
	}

	code := utils.GenerateUUID()

	hasCreated, err := dbService.CreateResetPassword(payload.Email, code, utils.ResetPasswordDelay)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zurücksetzen des Passworts ist fehlgeschlagen"})
		return
	}
	if hasCreated {
		err := sendgridService.SendPasswordResetMail(payload.Email, code)
		if err != nil {
			err := dbService.DeleteResetPassword(payload.Email)
			if err != nil {
				logger.Logger.Error("Couldn't delete reset password: ", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Zurücksetzen des Passworts ist fehlgeschlagen"})
			return
		}
	} else {
		logger.Logger.Infof("Skipped creating password reset")
	}

	c.Status(http.StatusOK)
}

// ResetPassword resets the users password
func ResetPassword(dbService db_service.IDatabaseService, c *gin.Context) {
	var payload models.ResetPassword
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Eingabe"})
		return
	}

	_, err := dbService.ValidateResetPassword(payload.Email, payload.Code, utils.ResetPasswordDelay)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zurücksetzen des Passworts ist fehlgeschlagen"})
		return
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zurücksetzen des Passworts ist fehlgeschlagen"})
		return
	}

	err = dbService.ResetPassword(string(encryptedPassword), payload.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zurücksetzen des Passworts ist fehlgeschlagen"})
		return
	}

	err = dbService.DeleteResetPassword(payload.Email)
	if err != nil {
		logger.Logger.Error("Couldn't delete reset password: ", err)
	}

	c.Status(http.StatusOK)
}

// CheckResetPasswordCode checks if the user can reset the password
func CheckResetPasswordCode(dbService db_service.IDatabaseService, c *gin.Context) {
	var payload models.CheckResetPasswordCode
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Eingabe"})
		return
	}

	_, err := dbService.ValidateResetPassword(payload.Email, payload.Code, utils.ResetPasswordDelay)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validierung des Zurücksetzen des Passworst ist fehlgeschlagen"})
		return
	}

	c.Status(http.StatusOK)
}

func clearRefreshTokenFromDatabase(dbService db_service.IDatabaseService, c *gin.Context) {
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
