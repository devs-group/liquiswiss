package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func scenarioIDParam(c *gin.Context) (int64, bool) {
	scenarioID, err := strconv.ParseInt(c.Param("scenarioID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return 0, false
	}
	return scenarioID, true
}

func scenarioError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		c.JSON(http.StatusNotFound, gin.H{"error": "Szenario nicht gefunden"})
	case errors.Is(err, api_service.ErrBaseScenario):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Das Grundszenario kann nicht gelöscht werden"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func ListScenarios(apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}

	scenarios, err := apiService.ListScenarios(c.Request.Context(), userID)
	if err != nil {
		scenarioError(c, err)
		return
	}

	c.JSON(http.StatusOK, scenarios)
}

func GetScenario(apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	scenarioID, ok := scenarioIDParam(c)
	if !ok {
		return
	}

	scenario, err := apiService.GetScenario(c.Request.Context(), userID, scenarioID)
	if err != nil {
		scenarioError(c, err)
		return
	}

	c.JSON(http.StatusOK, scenario)
}

func CreateScenario(apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	var payload models.CreateScenario
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if err := utils.GetValidator().Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	scenario, err := apiService.CreateScenario(c.Request.Context(), payload, userID)
	if err != nil {
		scenarioError(c, err)
		return
	}

	c.JSON(http.StatusCreated, scenario)
}

func UpdateScenario(apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	scenarioID, ok := scenarioIDParam(c)
	if !ok {
		return
	}
	var payload models.UpdateScenario
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if err := utils.GetValidator().Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	scenario, err := apiService.UpdateScenario(c.Request.Context(), payload, userID, scenarioID)
	if err != nil {
		scenarioError(c, err)
		return
	}

	c.JSON(http.StatusOK, scenario)
}

func DeleteScenario(apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	scenarioID, ok := scenarioIDParam(c)
	if !ok {
		return
	}

	if err := apiService.DeleteScenario(c.Request.Context(), userID, scenarioID); err != nil {
		scenarioError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func SwitchScenario(apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	scenarioID, ok := scenarioIDParam(c)
	if !ok {
		return
	}

	scenario, err := apiService.SwitchScenario(c.Request.Context(), userID, scenarioID)
	if err != nil {
		scenarioError(c, err)
		return
	}

	c.JSON(http.StatusOK, scenario)
}
