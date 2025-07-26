package handlers

import (
	"github.com/gin-gonic/gin"
	"liquiswiss/internal/service/api_service"
	"net/http"
)

func ListFiatRates(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	base := c.Param("base")
	if base == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	fiatRates, err := apiService.ListFiatRates(base)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, fiatRates)
}

func GetFiatRate(apiService api_service.IAPIService, c *gin.Context) {
	// Pre
	base := c.Param("base")
	if base == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	target := c.Param("target")
	if target == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// Action
	fiatRate, err := apiService.GetFiatRate(base, target)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Post
	c.JSON(http.StatusOK, fiatRate)
}
