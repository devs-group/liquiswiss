package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"liquiswiss/internal/service"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListEmployees(dbService service.IDatabaseService, c *gin.Context) {
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
	page, err := strconv.ParseInt(c.Query("page"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sortBy := c.DefaultQuery("sortBy", "name")
	sortOrder := c.DefaultQuery("sortOrder", "ASC")

	employees, totalCount, err := dbService.ListEmployees(userID, page, limit, sortBy, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Var(employees, "dive"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	totalPages := totalCount / limit
	if totalCount%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.ListResponse[models.Employee]{
		Data:       employees,
		Pagination: models.CalculatePagination(page, limit, totalCount),
	})
}

func ListEmployeeHistory(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	employeeID := c.Param("employeeID")
	if employeeID == "" {
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

func GetEmployeesPagination(dbService service.IDatabaseService, c *gin.Context) {
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
	page, err := strconv.ParseInt(c.Query("page"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	totalCount, err := dbService.CountEmployees(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := totalCount / limit
	if totalCount%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.PaginationResponse{
		Pagination: models.CalculatePagination(page, limit, totalCount),
	})
}

func GetEmployee(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	employeeID := c.Param("employeeID")
	if employeeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	employee, err := dbService.GetEmployee(userID, employeeID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Kein Mitarbeiter gefunden mit ID: %s", employeeID)})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	validator := utils.GetValidator()
	if err := validator.Struct(employee); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, employee)
}

func GetEmployeeHistory(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	historyID := c.Param("historyID")
	if historyID == "" {
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

func CreateEmployee(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	var payload models.CreateEmployee
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	employeeID, err := dbService.CreateEmployee(payload, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	employee, err := dbService.GetEmployee(userID, fmt.Sprint(employeeID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, employee)
}

func CreateEmployeeHistory(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	employeeID := c.Param("employeeID")
	if employeeID == "" {
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

	historyID, err := dbService.CreateEmployeeHistory(payload, userID, employeeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return
		}
		logger.Logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Beim Erstellen des Eintrags ist ein Fehler aufgetreten"})
		return
	}

	employeeHistory, err := dbService.GetEmployeeHistory(userID, fmt.Sprint(historyID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	CalculateForecasts(dbService, c)

	c.JSON(http.StatusCreated, employeeHistory)
}

func UpdateEmployee(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	employeeID := c.Param("employeeID")
	if employeeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	existingEmployee, err := dbService.GetEmployee(userID, employeeID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	var payload models.UpdateEmployee
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if payload.Name == nil {
		name := existingEmployee.Name
		payload.Name = &name
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	err = dbService.UpdateEmployee(payload, userID, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	employee, err := dbService.GetEmployee(userID, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, employee)
}

func UpdateEmployeeHistory(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	historyID := c.Param("historyID")
	if historyID == "" {
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
	if payload.SalaryPerMonth == nil {
		salaryPerMonth := existingEmployeeHistory.SalaryPerMonth
		payload.SalaryPerMonth = &salaryPerMonth
	}
	if payload.SalaryCurrency == nil {
		salaryCurrency := existingEmployeeHistory.SalaryCurrency
		payload.SalaryCurrency = salaryCurrency.ID
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

	err = dbService.UpdateEmployeeHistory(payload, existingEmployeeHistory.EmployeeID, historyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	employee, err := dbService.GetEmployeeHistory(userID, historyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	CalculateForecasts(dbService, c)

	c.JSON(http.StatusOK, employee)
}

func DeleteEmployee(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	employeeID := c.Param("employeeID")
	if employeeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	existingEmployee, err := dbService.GetEmployee(userID, employeeID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	err = dbService.DeleteEmployee(existingEmployee.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteEmployeeHistory(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	historyID := c.Param("historyID")
	if historyID == "" {
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

	CalculateForecasts(dbService, c)

	c.Status(http.StatusNoContent)
}
