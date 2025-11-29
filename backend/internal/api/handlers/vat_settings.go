package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
)

func GetVatSetting(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}

	// Action
	vatSetting, err := apiService.GetVatSetting(userID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.Status(http.StatusNotFound)
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	// Post
	c.JSON(http.StatusOK, vatSetting)
}

func CreateVatSetting(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	var payload models.CreateVatSetting
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
	vatSetting, err := apiService.CreateVatSetting(payload, userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusCreated, vatSetting)
}

func UpdateVatSetting(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	var payload models.UpdateVatSetting
	if err := c.Bind(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	vatSetting, err := apiService.UpdateVatSetting(payload, userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, vatSetting)
}

func DeleteVatSetting(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}

	// Action
	err := apiService.DeleteVatSetting(userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.Status(http.StatusNoContent)
}
