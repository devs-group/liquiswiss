package handlers

import (
	"github.com/gin-gonic/gin"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/pkg/utils"
	"net/http"
)

func ListFiatRates(dbService db_service.IDatabaseService, c *gin.Context) {
	fiatRates, err := dbService.ListFiatRates(utils.BaseCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fiatRates)
}

func GetFiatRate(dbService db_service.IDatabaseService, c *gin.Context) {
	targetCurrency := c.Param("currency")
	if targetCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target Currency is missing"})
		return
	}

	fiatRate, err := dbService.GetFiatRate(utils.BaseCurrency, targetCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Konnte Währung nicht finden"})
		return
	}

	c.JSON(http.StatusOK, fiatRate)
}
