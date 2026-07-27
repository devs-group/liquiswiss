package handlers

import (
	"errors"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListCategories(apiServce api_service.IAPIService, c *gin.Context) {
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

	// Action
	categories, totalCount, err := apiServce.ListCategories(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, models.ListResponse[models.Category]{
		Data:       categories,
		Pagination: models.CalculatePagination(page, limit, totalCount),
	})
}

func GetCategory(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	categoryID, err := strconv.ParseInt(c.Param("categoryID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	category, err := apiService.GetCategory(c.Request.Context(), userID, categoryID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, category)
}

func CreateCategory(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	var payload models.CreateCategory
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
	category, err := apiService.CreateCategory(c.Request.Context(), payload, &userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusCreated, category)
}

func UpdateCategory(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	categoryID, err := strconv.ParseInt(c.Param("categoryID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var payload models.UpdateCategory
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
	category, err := apiService.UpdateCategory(c.Request.Context(), payload, userID, categoryID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Poste
	c.JSON(http.StatusOK, category)
}

// ReassignCategory moves all transactions from one category to another so the
// source category can be deleted afterwards
func ReassignCategory(apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	categoryID, err := strconv.ParseInt(c.Param("categoryID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var payload models.ReassignCategory
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	affected, err := apiService.ReassignCategoryTransactions(c.Request.Context(), userID, categoryID, payload.TargetID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"affected": affected})
}

func DeleteCategory(apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	categoryID, err := strconv.ParseInt(c.Param("categoryID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	if err := apiService.DeleteCategory(c.Request.Context(), userID, categoryID); err != nil {
		switch {
		case errors.Is(err, api_service.ErrCategoryInUse):
			c.JSON(http.StatusConflict, gin.H{"error": "Diese Kategorie wird noch von Transaktionen verwendet und kann nicht gelöscht werden"})
		case errors.Is(err, api_service.ErrCategoryGlobal):
			c.JSON(http.StatusConflict, gin.H{"error": "System-Kategorien können nicht gelöscht werden"})
		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	// Poste
	c.Status(http.StatusNoContent)
}
