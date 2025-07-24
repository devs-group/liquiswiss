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

func ListEmployeeHistory(dbService db_service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	employeeID, err := strconv.ParseInt(c.Param("employeeID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
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

	employeeHistories, totalCount, err := dbService.ListEmployeeHistory(userID, employeeID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Var(employeeHistories, "dive"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	totalPages := totalCount / limit
	if totalCount%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.ListResponse[models.EmployeeHistory]{
		Data:       employeeHistories,
		Pagination: models.CalculatePagination(page, limit, totalCount),
	})
}

func GetEmployeeHistory(dbService db_service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	historyID, err := strconv.ParseInt(c.Param("historyID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	employeeHistory, err := dbService.GetEmployeeHistory(userID, historyID)
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
	if err := validator.Struct(employeeHistory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten"})
		return
	}

	c.JSON(http.StatusOK, employeeHistory)
}

func CreateEmployeeHistory(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	employeeID, err := strconv.ParseInt(c.Param("employeeID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	var payload models.CreateEmployeeHistory
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	historyID, previousHistoryID, nextHistoryID, err := dbService.CreateEmployeeHistory(payload, userID, employeeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return
		}
		logger.Logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Beim Erstellen des Eintrags ist ein Fehler aufgetreten"})
		return
	}

	// Refresh all cost details
	err = dbService.RefreshCostDetails(userID, historyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if previousHistoryID != nil {
		err = dbService.RefreshCostDetails(userID, *previousHistoryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if nextHistoryID != nil {
		err = dbService.RefreshCostDetails(userID, *nextHistoryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	employeeHistory, err := dbService.GetEmployeeHistory(userID, historyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Recalculate Forecast
	_, err = forecastService.CalculateForecast(userID)
	if err != nil {
		return
	}

	c.JSON(http.StatusCreated, employeeHistory)
}

func UpdateEmployeeHistory(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	historyID, err := strconv.ParseInt(c.Param("historyID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	existingEmployeeHistory, err := dbService.GetEmployeeHistory(userID, historyID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	var payload models.UpdateEmployeeHistory
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if payload.HoursPerMonth == nil {
		hoursPerMonth := existingEmployeeHistory.HoursPerMonth
		payload.HoursPerMonth = &hoursPerMonth
	}
	if payload.Salary == nil {
		salary := existingEmployeeHistory.Salary
		payload.Salary = &salary
	}
	if payload.CurrencyID == nil {
		payload.CurrencyID = existingEmployeeHistory.Currency.ID
	}
	if payload.VacationDaysPerYear == nil {
		vacationDaysPerYear := existingEmployeeHistory.VacationDaysPerYear
		payload.VacationDaysPerYear = &vacationDaysPerYear
	}
	if payload.FromDate == nil {
		fromDate := existingEmployeeHistory.FromDate.ToString()
		payload.FromDate = &fromDate
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	previousHistoryID, nextHistoryID, err := dbService.UpdateEmployeeHistory(payload, existingEmployeeHistory.EmployeeID, historyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Refresh all cost details
	err = dbService.RefreshCostDetails(userID, historyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if previousHistoryID != nil {
		err = dbService.RefreshCostDetails(userID, *previousHistoryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if nextHistoryID != nil {
		err = dbService.RefreshCostDetails(userID, *nextHistoryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	employee, err := dbService.GetEmployeeHistory(userID, historyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Recalculate Forecast
	_, err = forecastService.CalculateForecast(userID)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, employee)
}

func DeleteEmployeeHistory(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	historyID, err := strconv.ParseInt(c.Param("historyID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	existingEmployeeHistory, err := dbService.GetEmployeeHistory(userID, historyID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	err = dbService.DeleteEmployeeHistory(existingEmployeeHistory, userID)
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
