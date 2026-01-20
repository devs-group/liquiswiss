package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/auth"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"

	"github.com/gin-gonic/gin"
)

func ListOrganisationInvitations(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	organisationID, err := strconv.ParseInt(c.Param("organisationID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	invitations, err := apiService.ListOrganisationInvitations(userID, organisationID)
	if err != nil {
		if err.Error() == "permission denied" {
			c.Status(http.StatusForbidden)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, invitations)
}

func CreateOrganisationInvitation(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	organisationID, err := strconv.ParseInt(c.Param("organisationID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var payload models.CreateInvitation
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
	invitation, err := apiService.CreateOrganisationInvitation(payload, userID, organisationID)
	if err != nil {
		if err.Error() == "permission denied" {
			c.Status(http.StatusForbidden)
			return
		}
		if err.Error() == "user is already a member of this organisation" {
			c.JSON(http.StatusConflict, gin.H{"error": "user is already a member"})
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusCreated, invitation)
}

func DeleteOrganisationInvitation(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	organisationID, err := strconv.ParseInt(c.Param("organisationID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	invitationID, err := strconv.ParseInt(c.Param("invitationID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	err = apiService.DeleteOrganisationInvitation(userID, organisationID, invitationID)
	if err != nil {
		if err.Error() == "permission denied" {
			c.Status(http.StatusForbidden)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.Status(http.StatusNoContent)
}

func ResendOrganisationInvitation(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	organisationID, err := strconv.ParseInt(c.Param("organisationID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	invitationID, err := strconv.ParseInt(c.Param("invitationID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	err = apiService.ResendOrganisationInvitation(userID, organisationID, invitationID)
	if err != nil {
		if err.Error() == "permission denied" {
			c.Status(http.StatusForbidden)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.Status(http.StatusOK)
}

func CheckInvitation(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	token := c.Query("token")
	if token == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	response, err := apiService.CheckInvitation(token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return
		}
		if err.Error() == "invitation has expired" {
			c.JSON(http.StatusGone, gin.H{"error": "invitation has expired"})
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, response)
}

func AcceptInvitation(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	var payload models.AcceptInvitation
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
	user, accessToken, accessExpirationTime, refreshToken, refreshExpirationTime, err := apiService.AcceptInvitation(payload, deviceName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return
		}
		if err.Error() == "invitation has expired" {
			c.JSON(http.StatusGone, gin.H{"error": "invitation has expired"})
			return
		}
		if err.Error() == "password is required for new users" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
			return
		}
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
