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

func ListSalaries(apiService api_service.IAPIService, c *gin.Context) {
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

	// Action
	salaries, totalCount, err := apiService.ListSalaries(userID, employeeID, page, limit)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, models.ListResponse[models.Salary]{
		Data:       salaries,
		Pagination: models.CalculatePagination(page, limit, totalCount),
	})
}

func GetSalary(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
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

	// Action
	salary, err := apiService.GetSalary(userID, salaryID)
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
	c.JSON(http.StatusOK, salary)
}

func CreateSalary(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
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

	// Action
	salary, err := apiService.CreateSalary(payload, employeeID, userID)
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
	c.JSON(http.StatusCreated, salary)
}

func UpdateSalary(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	salaryID, err := strconv.ParseInt(c.Param("salaryID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var payload models.UpdateSalary
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
	salary, err := apiService.UpdateSalary(payload, userID, salaryID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, salary)
}

func DeleteSalary(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	salaryID, err := strconv.ParseInt(c.Param("salaryID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	err = apiService.DeleteSalary(userID, salaryID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.Status(http.StatusNoContent)
}
