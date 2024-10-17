package utils

import (
	"github.com/google/uuid"
	"liquiswiss/pkg/models"
	"math"
	"os"
)

func IsProduction() bool {
	mode, _ := os.LookupEnv("GIN_MODE")
	return mode == "release"
}

func CalculatePagination(currentPage int64, limit int64, totalCount int64) models.Pagination {
	pagination := models.Pagination{
		CurrentPage: currentPage,
		TotalCount:  totalCount,
	}

	// Calculate total pages
	pagination.TotalPages = totalCount / limit
	if totalCount%limit != 0 {
		pagination.TotalPages++ // Add an extra page if there are remaining items
	}

	// Calculate the total remaining items after the current page
	itemsAfterCurrentPage := totalCount - (currentPage * limit)
	if itemsAfterCurrentPage < 0 {
		itemsAfterCurrentPage = 0 // Ensure it doesn't go negative
	}
	pagination.TotalRemaining = itemsAfterCurrentPage

	return pagination
}

// GenerateUUID generates a new UUID
func GenerateUUID() string {
	return uuid.New().String()
}

func CalculateAmountWithFiatRate(amount int64, rate float64) int64 {
	return int64(math.Round(float64(amount) / rate))
}

func GetFiatRateFromCurrency(fiatRates []models.FiatRate, currencyCode string) float64 {
	if currencyCode == "CHF" {
		return 1.0
	}

	for _, rate := range fiatRates {
		if rate.Target == currencyCode {
			return rate.Rate
		}
	}

	return 1.0
}
