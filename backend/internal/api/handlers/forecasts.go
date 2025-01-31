package handlers

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/internal/service/forecast_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"strconv"
)

func ListForecasts(dbService db_service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	limit, err := strconv.ParseInt(c.Query("limit"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	forecasts, err := dbService.ListForecasts(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Var(forecasts, "dive"); err != nil {
		// Return validation errors
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, forecasts)
}

func ListForecastDetails(dbService db_service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	limit, err := strconv.ParseInt(c.Query("limit"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	forecastDetails, err := dbService.ListForecastDetails(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Var(forecastDetails, "dive"); err != nil {
		// Return validation errors
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, forecastDetails)
}

func CalculateForecasts(forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	// Recalculate Forecast
	foreCasts, err := forecastService.CalculateForecast(userID)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, foreCasts)
}

func ListForecastExclusions(dbService db_service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	relatedID, err := strconv.ParseInt(c.Query("relatedID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	relatedTable := c.Query("relatedTable")
	if relatedTable == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing relatedTable"})
		return
	}

	forecastExclusions, err := dbService.ListForecastExclusions(userID, relatedID, relatedTable)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Var(forecastExclusions, "dive"); err != nil {
		// Return validation errors
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, forecastExclusions)
}

func CreateForecastExclusion(dbService db_service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	var payload models.CreateForecastExclusion
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

	excludeID, err := dbService.CreateForecastExclusion(payload, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if excludeID == 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusCreated)
}

func DeleteForecastExclusion(dbService db_service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	var payload models.CreateForecastExclusion
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

	affected, err := dbService.DeleteForecastExclusion(payload, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if affected == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusNoContent)
}
