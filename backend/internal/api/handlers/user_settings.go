package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func GetUserSetting(apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}

	setting, err := apiService.GetUserSetting(userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, setting)
}

func UpdateUserSetting(apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}

	var payload models.UpdateUserSetting
	if err := c.Bind(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	setting, err := apiService.UpdateUserSetting(payload, userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, setting)
}

func GetUserOrganisationSetting(apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}

	setting, err := apiService.GetUserOrganisationSetting(userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, setting)
}

func UpdateUserOrganisationSetting(apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}

	var payload models.UpdateUserOrganisationSetting
	if err := c.Bind(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	setting, err := apiService.UpdateUserOrganisationSetting(payload, userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, setting)
}
