package utils

import (
	"github.com/google/uuid"
	"liquiswiss/pkg/models"
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
