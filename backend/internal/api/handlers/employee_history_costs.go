package handlers

import (
	"database/sql"
	"errors"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/internal/service/forecast_service"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListEmployeeHistoryCosts(dbService db_service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	historyID, err := strconv.ParseInt(c.Param("historyID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die Lohn ID"})
		return
	}

	limit, err := strconv.ParseInt(c.Query("limit"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	page, err := strconv.ParseInt(c.Query("page"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	historyCosts, totalCount, err := dbService.ListEmployeeHistoryCosts(userID, historyID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Var(historyCosts, "dive"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	totalPages := totalCount / limit
	if totalCount%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.ListResponse[models.EmployeeHistoryCost]{
		Data:       historyCosts,
		Pagination: models.CalculatePagination(page, limit, totalCount),
	})
}

func GetEmployeeHistoryCost(dbService db_service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	historyCostID, err := strconv.ParseInt(c.Param("historyCostID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die Lohnkosten ID"})
		return
	}

	historyCost, err := dbService.GetEmployeeHistoryCost(userID, historyCostID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.Status(http.StatusNotFound)
			return
		default:
			logger.Logger.Error(err)
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	validator := utils.GetValidator()
	if err := validator.Struct(historyCost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten"})
		return
	}

	c.JSON(http.StatusOK, historyCost)
}

func CreateEmployeeHistoryCost(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	historyID, err := strconv.ParseInt(c.Param("historyID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die Lohn ID"})
		return
	}

	var payload models.CreateEmployeeHistoryCost
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	historyCostID, err := dbService.CreateEmployeeHistoryCost(payload, userID, historyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return
		}
		logger.Logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Beim Erstellen des Eintrags ist ein Fehler aufgetreten"})
		return
	}

	err = dbService.CalculateEmployeeHistoryCostDetails(historyCostID, userID)
	if err != nil {
		return
	}

	historyCost, err := dbService.GetEmployeeHistoryCost(userID, historyCostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Recalculate Forecast
	_, err = forecastService.CalculateForecast(userID)
	if err != nil {
		return
	}

	c.JSON(http.StatusCreated, historyCost)
}

func CopyEmployeeHistoryCosts(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	historyID, err := strconv.ParseInt(c.Param("historyID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die Lohn ID"})
		return
	}

	var payload models.CopyEmployeeHistoryCosts
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	err = dbService.CopyEmployeeHistoryCosts(payload, userID, historyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return
		}
		logger.Logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Beim Kopieren der Lohnkosten ist ein Fehler aufgetreten"})
		return
	}

	err = dbService.RefreshCostDetails(userID, historyID)
	if err != nil {
		return
	}

	// Recalculate Forecast
	_, err = forecastService.CalculateForecast(userID)
	if err != nil {
		return
	}

	c.Status(http.StatusCreated)
}

func UpdateEmployeeHistoryCost(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	historyCostID, err := strconv.ParseInt(c.Param("historyCostID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	_, err = dbService.GetEmployeeHistoryCost(userID, historyCostID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	var payload models.CreateEmployeeHistoryCost
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	err = dbService.UpdateEmployeeHistoryCost(payload, userID, historyCostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = dbService.CalculateEmployeeHistoryCostDetails(historyCostID, userID)
	if err != nil {
		return
	}

	historyCost, err := dbService.GetEmployeeHistoryCost(userID, historyCostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Recalculate Forecast
	_, err = forecastService.CalculateForecast(userID)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, historyCost)
}

func DeleteEmployeeHistoryCost(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	historyCostID, err := strconv.ParseInt(c.Param("historyCostID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	existingEmployeeHistory, err := dbService.GetEmployeeHistoryCost(userID, historyCostID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	err = dbService.DeleteEmployeeHistoryCost(existingEmployeeHistory.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Recalculate Forecast
	_, err = forecastService.CalculateForecast(userID)
	if err != nil {
		return
	}

	c.Status(http.StatusNoContent)
}
