package handlers

import (
	"database/sql"
	"fmt"
	"liquiswiss/internal/service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListOrganisations(dbService service.IDatabaseService, c *gin.Context) {
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

	organisations, totalCount, err := dbService.ListOrganisations(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Var(organisations, "dive"); err != nil {
		// Return validation errors
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	totalPages := totalCount / limit
	if totalCount%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.ListResponse[models.Organisation]{
		Data:       organisations,
		Pagination: models.CalculatePagination(page, limit, totalCount),
	})
}

func GetOrganisation(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	organisation, err := dbService.GetOrganisation(userID, id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Keine Organisation gefunden mit ID: %s", id)})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	validator := utils.GetValidator()
	if err := validator.Struct(organisation); err != nil {
		// Return validation errors
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, organisation)
}

func CreateOrganisation(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	// TODO: Check if user already has organisation?

	var payload models.CreateOrganisation
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		// Return validation errors
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	organisationID, err := dbService.CreateOrganisation(payload.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = dbService.AssignUserToOrganisation(userID, organisationID, "owner", false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	organisation, err := dbService.GetOrganisation(userID, fmt.Sprint(organisationID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, organisation)
}

func UpdateOrganisation(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Es fehlt die ID"})
		return
	}

	existingOrganisation, err := dbService.GetOrganisation(userID, id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	// Check if user is allowed to edit
	if !hasEditingPermission(existingOrganisation.Role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Keine Berechtigung für diese Aktion"})
		return
	}

	var payload models.UpdateOrganisation
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if payload.Name == nil {
		startDate := existingOrganisation.Name
		payload.Name = &startDate
	}

	validator := utils.GetValidator()
	if err := validator.Struct(payload); err != nil {
		// Return validation errors
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	err = dbService.UpdateOrganisation(payload, userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	transaction, err := dbService.GetOrganisation(userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// hasEditingPermission is a simple method to check if the current user can edit the Organisation
func hasEditingPermission(role string) bool {
	editingRoles := []string{"owner", "admin"}
	for _, editingRole := range editingRoles {
		if role == editingRole {
			return true
		}
	}
	return false
}
