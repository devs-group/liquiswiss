package handlers

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"liquiswiss/internal/service"
	"liquiswiss/pkg/utils"
	"net/http"
)

func GetProfile(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	user, err := dbService.GetProfile(fmt.Sprint(userID))
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Kein Benutzer gefunden mit ID: %d", userID)})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	validator := utils.GetValidator()
	if err := validator.Struct(user); err != nil {
		// Return validation errors
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func GetAccessToken(dbService service.IDatabaseService, c *gin.Context) {
	// This does nothing it's simply for the user to get a refresh token
	c.Status(http.StatusNoContent)
}
