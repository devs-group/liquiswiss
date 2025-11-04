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

func ListScenarios(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}

	// Action
	scenarios, err := apiService.ListScenarios(userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, scenarios)
}

func GetScenario(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	scenarioID, err := strconv.ParseInt(c.Param("scenarioID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	scenario, err := apiService.GetScenario(userID, scenarioID)
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
	c.JSON(http.StatusOK, scenario)
}

func GetDefaultScenario(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}

	// Action
	scenario, err := apiService.GetDefaultScenario(userID)
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
	c.JSON(http.StatusOK, scenario)
}

func CreateScenario(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
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
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	scenario, err := apiService.CreateScenario(payload, userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusCreated, scenario)
}

func UpdateScenario(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	scenarioID, err := strconv.ParseInt(c.Param("scenarioID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var payload models.UpdateScenario
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
	scenario, err := apiService.UpdateScenario(payload, userID, scenarioID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, scenario)
}

func DeleteScenario(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	scenarioID, err := strconv.ParseInt(c.Param("scenarioID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	err = apiService.DeleteScenario(userID, scenarioID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.Status(http.StatusNotFound)
			return
		case api_service.ErrCannotDeleteDefaultScenario:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	// Post
	c.Status(http.StatusNoContent)
}
