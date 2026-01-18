package models

import (
	"math"
)

func CalculatePagination(currentPage int64, limit int64, totalCount int64) Pagination {
	pagination := Pagination{
		CurrentPage: currentPage,
		TotalCount:  totalCount,
	}

	// Calculate total pages
	pagination.TotalPages = totalCount / limit
	if totalCount%limit != 0 {
		pagination.TotalPages++ // Add an extra page if there are remaining items
	}

	// Calculate the total remaining items after the current page
	itemsAfterCurrentPage := max(totalCount-(currentPage*limit),
		// Ensure it doesn't go negative
		0)
	pagination.TotalRemaining = itemsAfterCurrentPage

	return pagination
}

func CalculateAmountWithFiatRate(amount int64, rate float64) int64 {
	return int64(math.Round(float64(amount) / rate))
}

func GetFiatRateFromCurrency(fiatRates []FiatRate, base, target string) float64 {
	if base == target {
		return 1.0
	}

	for _, rate := range fiatRates {
		if rate.Base == base && rate.Target == target {
			return rate.Rate
		}
	}

	return 1.0
}
