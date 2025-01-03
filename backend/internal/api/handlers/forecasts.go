package handlers

import (
	"github.com/gin-gonic/gin"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/internal/service/forecast_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"strconv"
	"time"
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

	c.JSON(http.StatusOK, map[string]interface{}{
		"forecast": foreCasts,
	})
}

func initForecastMapKey(forecastMap map[string]map[string]int64, monthKey string) {
	forecastMap[monthKey] = make(map[string]int64)
	forecastMap[monthKey]["revenue"] = 0
	forecastMap[monthKey]["expense"] = 0
}

func initForecastDetailMapKey(forecastDetailMap map[string]map[string]map[string]int64, monthKey, subkey string) {
	forecastDetailMap[monthKey] = make(map[string]map[string]int64)
	forecastDetailMap[monthKey]["revenue"] = make(map[string]int64)
	forecastDetailMap[monthKey]["expense"] = make(map[string]int64)
	forecastDetailMap[monthKey]["revenue"][subkey] = 0
	forecastDetailMap[monthKey]["expense"][subkey] = 0
}

func addForecastDetail(detailMap map[string]*models.ForecastDetails, monthKey, category string, amount int64) {
	// Make sure the map is prepared
	if detailMap[monthKey] == nil {
		detailMap[monthKey] = &models.ForecastDetails{
			Revenue: make(map[string]int64),
			Expense: make(map[string]int64),
		}
	}
	if amount > 0 {
		// Make sure the inner map is prepared
		if detailMap[monthKey].Revenue[category] == 0 {
			detailMap[monthKey].Revenue[category] = 0
		}

		detailMap[monthKey].Revenue[category] += amount
	} else if amount < 0 {
		// Make sure the inner map is prepared
		if detailMap[monthKey].Expense[category] == 0 {
			detailMap[monthKey].Expense[category] = 0
		}

		detailMap[monthKey].Expense[category] += amount
	}
}

func getYearMonth(date time.Time) string {
	return date.Format("2006-01")
}
