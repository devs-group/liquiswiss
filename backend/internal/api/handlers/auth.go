package handlers

import (
	"github.com/gin-gonic/gin"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/auth"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
)

func Login(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	var payload models.Login
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	deviceName := c.Request.UserAgent()
	existingRefreshToken, err := c.Cookie(utils.RefreshTokenName)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Action
	user, accessToken, accessExpirationTime, refreshToken, refreshExpirationTime, err := apiService.Login(payload, deviceName, existingRefreshToken)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	accessTokenCookie := auth.GenerateCookie(utils.AccessTokenName, *accessToken, *accessExpirationTime)
	http.SetCookie(c.Writer, &accessTokenCookie)
	refreshTokenCookie := auth.GenerateCookie(utils.RefreshTokenName, *refreshToken, *refreshExpirationTime)
	http.SetCookie(c.Writer, &refreshTokenCookie)

	c.JSON(http.StatusOK, user)
}

func Logout(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	existingRefreshToken, err := c.Cookie(utils.RefreshTokenName)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Action
	apiService.Logout(existingRefreshToken)

	// Post
	auth.ClearAuthCookies(c)
	c.Status(http.StatusOK)
}

func ForgotPassword(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	var payload models.ForgotPassword
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	code := utils.GenerateUUID()
	err := apiService.ForgotPassword(payload, code)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.Status(http.StatusOK)
}

func ResetPassword(apiService api_service.IAPIService, c *gin.Context) {
	// Post
	var payload models.ResetPassword
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	err := apiService.ResetPassword(payload)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.Status(http.StatusOK)
}

func CheckResetPasswordCode(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	var payload models.CheckResetPasswordCode
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	err := apiService.CheckResetPasswordCode(payload)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.Status(http.StatusOK)
}

func CreateRegistration(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	var payload models.CreateRegistration
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	code := utils.GenerateUUID()
	_, err := apiService.CreateRegistration(payload, code)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.Status(http.StatusOK)
}

func CheckRegistrationCode(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	var payload models.CheckRegistrationCode
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	_, err := apiService.CheckRegistrationCode(payload, utils.RegistrationCodeValidity)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Post
	c.Status(http.StatusOK)
}

func FinishRegistration(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	var payload models.FinishRegistration
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	deviceName := c.Request.UserAgent()

	// Action
	user, accessToken, accessExpirationTime, refreshToken, refreshExpirationTime, err := apiService.FinishRegistration(payload, deviceName, utils.RegistrationCodeValidity)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	accessTokenCookie := auth.GenerateCookie(utils.AccessTokenName, *accessToken, *accessExpirationTime)
	http.SetCookie(c.Writer, &accessTokenCookie)
	refreshTokenCookie := auth.GenerateCookie(utils.RefreshTokenName, *refreshToken, *refreshExpirationTime)
	http.SetCookie(c.Writer, &refreshTokenCookie)

	c.JSON(http.StatusOK, user)
}
