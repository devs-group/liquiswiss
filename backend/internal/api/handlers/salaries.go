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

func ListSalaries(dbService db_service.IDatabaseService, c *gin.Context) {
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

	salaries, totalCount, err := dbService.ListSalaries(userID, employeeID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Var(salaries, "dive"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	totalPages := totalCount / limit
	if totalCount%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.ListResponse[models.Salary]{
		Data:       salaries,
		Pagination: models.CalculatePagination(page, limit, totalCount),
	})
}

func GetSalary(dbService db_service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	salaryID, err := strconv.ParseInt(c.Param("salaryID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	salary, err := dbService.GetSalary(userID, salaryID)
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
	if err := validator.Struct(salary); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten"})
		return
	}

	c.JSON(http.StatusOK, salary)
}

func CreateSalary(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
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

	var payload models.CreateSalary
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	salaryID, previousSalaryID, nextSalaryID, err := dbService.CreateSalary(payload, userID, employeeID)
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
	err = dbService.RefreshSalaryCostDetails(userID, salaryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if previousSalaryID != nil {
		err = dbService.RefreshSalaryCostDetails(userID, *previousSalaryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if nextSalaryID != nil {
		err = dbService.RefreshSalaryCostDetails(userID, *nextSalaryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	salary, err := dbService.GetSalary(userID, salaryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Recalculate Forecast
	_, err = forecastService.CalculateForecast(userID)
	if err != nil {
		return
	}

	c.JSON(http.StatusCreated, salary)
}

func UpdateSalary(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	salaryID, err := strconv.ParseInt(c.Param("salaryID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	existingSalary, err := dbService.GetSalary(userID, salaryID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	var payload models.UpdateSalary
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if payload.HoursPerMonth == nil {
		hoursPerMonth := existingSalary.HoursPerMonth
		payload.HoursPerMonth = &hoursPerMonth
	}
	if payload.Amount == nil {
		salary := existingSalary.Amount
		payload.Amount = &salary
	}
	if payload.CurrencyID == nil {
		payload.CurrencyID = existingSalary.Currency.ID
	}
	if payload.VacationDaysPerYear == nil {
		vacationDaysPerYear := existingSalary.VacationDaysPerYear
		payload.VacationDaysPerYear = &vacationDaysPerYear
	}
	if payload.FromDate == nil {
		fromDate := existingSalary.FromDate.ToString()
		payload.FromDate = &fromDate
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	previousSalaryID, nextSalaryID, err := dbService.UpdateSalary(payload, existingSalary.EmployeeID, salaryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Refresh all cost details
	err = dbService.RefreshSalaryCostDetails(userID, salaryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if previousSalaryID != nil {
		err = dbService.RefreshSalaryCostDetails(userID, *previousSalaryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if nextSalaryID != nil {
		err = dbService.RefreshSalaryCostDetails(userID, *nextSalaryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	employee, err := dbService.GetSalary(userID, salaryID)
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

func DeleteSalary(dbService db_service.IDatabaseService, forecastService forecast_service.IForecastService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	salaryID, err := strconv.ParseInt(c.Param("salaryID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	existingSalary, err := dbService.GetSalary(userID, salaryID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	previousSalaryID, nextSalaryID, err := dbService.DeleteSalary(existingSalary, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Refresh all cost details
	err = dbService.RefreshSalaryCostDetails(userID, salaryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if previousSalaryID != nil {
		err = dbService.RefreshSalaryCostDetails(userID, *previousSalaryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if nextSalaryID != nil {
		err = dbService.RefreshSalaryCostDetails(userID, *nextSalaryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Recalculate Forecast
	_, err = forecastService.CalculateForecast(userID)
	if err != nil {
		return
	}

	c.Status(http.StatusNoContent)
}
