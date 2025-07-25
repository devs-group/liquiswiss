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

func ListSalaryCosts(dbService db_service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	salaryID, err := strconv.ParseInt(c.Param("salaryID"), 10, 64)
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

	salaryCosts, totalCount, err := dbService.ListSalaryCosts(userID, salaryID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Var(salaryCosts, "dive"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	totalPages := totalCount / limit
	if totalCount%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.ListResponse[models.SalaryCost]{
		Data:       salaryCosts,
		Pagination: models.CalculatePagination(page, limit, totalCount),
	})
}

func GetSalaryCost(dbService db_service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	salaryCostID, err := strconv.ParseInt(c.Param("salaryCostID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die Lohnkosten ID"})
		return
	}

	salaryCost, err := dbService.GetSalaryCost(userID, salaryCostID)
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
	if err := validator.Struct(salaryCost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten"})
		return
	}

	c.JSON(http.StatusOK, salaryCost)
}

func CreateSalaryCost(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	salaryID, err := strconv.ParseInt(c.Param("salaryID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die Lohn ID"})
		return
	}

	var payload models.CreateSalaryCost
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	salaryCostID, err := dbService.CreateSalaryCost(payload, userID, salaryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return
		}
		logger.Logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Beim Erstellen des Eintrags ist ein Fehler aufgetreten"})
		return
	}

	err = dbService.CalculateSalaryCostDetails(salaryCostID, userID)
	if err != nil {
		return
	}

	salaryCost, err := dbService.GetSalaryCost(userID, salaryCostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Recalculate Forecast
	_, err = forecastService.CalculateForecast(userID)
	if err != nil {
		return
	}

	c.JSON(http.StatusCreated, salaryCost)
}

func CopySalaryCosts(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	salaryID, err := strconv.ParseInt(c.Param("salaryID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die Lohn ID"})
		return
	}

	var payload models.CopySalaryCosts
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	err = dbService.CopySalaryCosts(payload, userID, salaryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return
		}
		logger.Logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Beim Kopieren der Lohnkosten ist ein Fehler aufgetreten"})
		return
	}

	err = dbService.RefreshSalaryCostDetails(userID, salaryID)
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

func UpdateSalaryCost(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	salaryCostID, err := strconv.ParseInt(c.Param("salaryCostID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	_, err = dbService.GetSalaryCost(userID, salaryCostID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	var payload models.CreateSalaryCost
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	err = dbService.UpdateSalaryCost(payload, userID, salaryCostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = dbService.CalculateSalaryCostDetails(salaryCostID, userID)
	if err != nil {
		return
	}

	salaryCost, err := dbService.GetSalaryCost(userID, salaryCostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Recalculate Forecast
	_, err = forecastService.CalculateForecast(userID)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, salaryCost)
}

func DeleteSalaryCost(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	salaryCostID, err := strconv.ParseInt(c.Param("salaryCostID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	existingSalaryCost, err := dbService.GetSalaryCost(userID, salaryCostID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	err = dbService.DeleteSalaryCost(existingSalaryCost.ID, userID)
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
