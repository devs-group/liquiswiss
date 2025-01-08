package handlers

import (
	"github.com/gin-gonic/gin"
	"liquiswiss/internal/service/db_service"
	"net/http"
)

func ListFiatRates(dbService db_service.IDatabaseService, c *gin.Context) {
	base := c.Param("base")
	if base == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Basiswährung fehlt"})
		return
	}

	fiatRates, err := dbService.ListFiatRates(base)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fiatRates)
}

func GetFiatRate(dbService db_service.IDatabaseService, c *gin.Context) {
	base := c.Param("base")
	if base == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Basiswährung fehlt"})
		return
	}

	target := c.Param("target")
	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zielwährung fehlt"})
		return
	}

	fiatRate, err := dbService.GetFiatRate(base, target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Konnte Währungskombination nicht finden"})
		return
	}

	c.JSON(http.StatusOK, fiatRate)
}
