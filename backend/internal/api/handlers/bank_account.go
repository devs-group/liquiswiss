package handlers

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"strconv"
)

func ListBankAccounts(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}

	// Action
	bankAccounts, err := apiService.ListBankAccounts(userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, bankAccounts)
}

func GetBankAccount(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}
	bankAccountID, err := strconv.ParseInt(c.Param("bankAccountID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	// Action
	bankAccount, err := apiService.GetBankAccount(userID, bankAccountID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Kein Bankkonto gefunden mit ID: %d", bankAccountID)})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Post
	c.JSON(http.StatusOK, bankAccount)
}

func CreateBankAccount(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	var payload models.CreateBankAccount
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
	bankAccount, err := apiService.CreateBankAccount(payload, userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusCreated, bankAccount)
}

func UpdateBankAccount(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	bankAccountID, err := strconv.ParseInt(c.Param("bankAccountID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var payload models.UpdateBankAccount
	if err := c.Bind(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Create
	bankAccount, err := apiService.UpdateBankAccount(payload, userID, bankAccountID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, bankAccount)
}

func DeleteBankAccount(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	bankAccountID, err := strconv.ParseInt(c.Param("bankAccountID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Create
	err = apiService.DeleteBankAccount(userID, bankAccountID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.Status(http.StatusNoContent)
}
