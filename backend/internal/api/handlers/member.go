package handlers

import (
	"net/http"
	"strconv"

	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"

	"github.com/gin-gonic/gin"
)

func ListOrganisationMembers(apiService api_service.IAPIService, c *gin.Context) {
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
	members, err := apiService.ListOrganisationMembers(userID, organisationID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, members)
}

func UpdateOrganisationMember(apiService api_service.IAPIService, c *gin.Context) {
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
	memberUserID, err := strconv.ParseInt(c.Param("memberUserID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var payload models.UpdateMember
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if utils.IsStructEmpty(&payload) {
		c.Status(http.StatusBadRequest)
		return
	}
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	err = apiService.UpdateOrganisationMember(payload, userID, organisationID, memberUserID)
	if err != nil {
		if err.Error() == "permission denied" {
			c.Status(http.StatusForbidden)
			return
		}
		if err.Error() == "cannot demote the last owner" {
			c.JSON(http.StatusConflict, gin.H{"error": "cannot demote the last owner"})
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.Status(http.StatusOK)
}

func RemoveOrganisationMember(apiService api_service.IAPIService, c *gin.Context) {
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
	memberUserID, err := strconv.ParseInt(c.Param("memberUserID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	err = apiService.RemoveOrganisationMember(userID, organisationID, memberUserID)
	if err != nil {
		if err.Error() == "permission denied" {
			c.Status(http.StatusForbidden)
			return
		}
		if err.Error() == "cannot remove the last owner" {
			c.JSON(http.StatusConflict, gin.H{"error": "cannot remove the last owner"})
			return
		}
		if err.Error() == "cannot remove yourself from the organisation" {
			c.JSON(http.StatusConflict, gin.H{"error": "cannot remove yourself"})
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.Status(http.StatusNoContent)
}
