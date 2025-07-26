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

func ListOrganisations(apiService api_service.IAPIService, c *gin.Context) {
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
	organisations, totalCount, err := apiService.ListOrganisations(userID, page, limit)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, models.ListResponse[models.Organisation]{
		Data:       organisations,
		Pagination: models.CalculatePagination(page, limit, totalCount),
	})
}

func GetOrganisation(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	organisationID, err := strconv.ParseInt(c.Param("organisationID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	organisation, err := apiService.GetOrganisation(userID, organisationID)
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
	c.JSON(http.StatusOK, organisation)
}

func CreateOrganisation(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	var payload models.CreateOrganisation
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		// Return validation errors
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	organisation, err := apiService.CreateOrganisation(payload, userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusCreated, organisation)
}

func UpdateOrganisation(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	organisationID, err := strconv.ParseInt(c.Param("organisationID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var payload models.UpdateOrganisation
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if utils.IsStructEmpty(&payload) {
		c.Status(http.StatusBadRequest)
		return
	}
	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	organisation, err := apiService.UpdateOrganisation(payload, organisationID, userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, organisation)
}
