package handlers

import (
	"database/sql"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListEmployees(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	limit, err := strconv.ParseInt(c.Query("limit"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	page, err := strconv.ParseInt(c.Query("page"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	sortBy := c.DefaultQuery("sortBy", "name")
	sortOrder := c.DefaultQuery("sortOrder", "ASC")
	search := c.Query("search")

	// Action
	employees, totalCount, err := apiService.ListEmployees(userID, page, limit, sortBy, sortOrder, search)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, models.ListResponse[models.Employee]{
		Data:       employees,
		Pagination: models.CalculatePagination(page, limit, totalCount),
	})
}

func GetEmployee(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	employeeID, err := strconv.ParseInt(c.Param("employeeID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	employee, err := apiService.GetEmployee(userID, employeeID)
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
	c.JSON(http.StatusOK, employee)
}

func CreateEmployee(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	var payload models.CreateEmployee
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
	employee, err := apiService.CreateEmployee(payload, userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusCreated, employee)
}

func UpdateEmployee(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	employeeID, err := strconv.ParseInt(c.Param("employeeID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var payload models.UpdateEmployee
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
	employee, err := apiService.UpdateEmployee(payload, userID, employeeID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, employee)
}

func DeleteEmployee(apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	employeeID, err := strconv.ParseInt(c.Param("employeeID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	err = apiService.DeleteEmployee(userID, employeeID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.Status(http.StatusNoContent)
}

func GetEmployeesPagination(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
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

	// Action
	totalCount, err := apiService.CountEmployees(userID, page, limit)
	if err != nil {
		return
	}

	// Post
	c.JSON(http.StatusOK, models.PaginationResponse{
		Pagination: models.CalculatePagination(page, limit, totalCount),
	})
}
