package handlers

import (
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListCurrencies(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}

	// Action
	currencies, err := apiService.ListCurrencies(userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, currencies)
}

func GetCurrency(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	currencyID, err := strconv.ParseInt(c.Param("currencyID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	currency, err := apiService.GetCurrency(currencyID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, currency)
}

func CreateCurrency(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	var payload models.CreateCurrency
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
	currency, err := apiService.CreateCurrency(payload)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusCreated, currency)
}

func UpdateCurrency(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	currencyID, err := strconv.ParseInt(c.Param("currencyID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var payload models.UpdateCurrency
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
	currency, err := apiService.UpdateCurrency(payload, currencyID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, currency)
}
