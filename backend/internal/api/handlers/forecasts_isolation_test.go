package handlers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

// TestListForecasts_CrossOrgIsolation verifies that users can only see
// forecasts belonging to their own organisation
func TestListForecasts_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Set database time to 2025-01-01 so our test dates are in the future
	simulatedTime := "2025-01-01"
	err := SetDatabaseTime(env.Conn, simulatedTime)
	require.NoError(t, err)
	parsedTime, err := time.Parse(utils.InternalDateFormat, simulatedTime)
	require.NoError(t, err)
	utils.DefaultClock.SetFixedTime(&parsedTime)
	defer utils.DefaultClock.SetFixedTime(nil)

	// Create transactions for User A (to generate forecast data)
	categoryA, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category A"}, &env.UserA.ID)
	require.NoError(t, err)

	_, err = env.APIService.CreateTransaction(models.CreateTransaction{
		Name:        "Transaction A",
		Amount:      1000_00,
		Type:        "single",
		StartDate:   "2025-01-15",
		Category:    categoryA.ID,
		Currency:    *env.Currency.ID,
		VatIncluded: false,
	}, env.UserA.ID)
	require.NoError(t, err)

	// Create transactions for User B
	categoryB, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category B"}, &env.UserB.ID)
	require.NoError(t, err)

	_, err = env.APIService.CreateTransaction(models.CreateTransaction{
		Name:        "Transaction B",
		Amount:      2000_00,
		Type:        "single",
		StartDate:   "2025-01-20",
		Category:    categoryB.ID,
		Currency:    *env.Currency.ID,
		VatIncluded: false,
	}, env.UserB.ID)
	require.NoError(t, err)

	// Calculate forecasts for both users
	_, err = env.APIService.CalculateForecast(env.UserA.ID)
	require.NoError(t, err)

	_, err = env.APIService.CalculateForecast(env.UserB.ID)
	require.NoError(t, err)

	// User A should only see their own forecasts
	forecastsA, err := env.APIService.ListForecasts(env.UserA.ID, 12)
	require.NoError(t, err)
	require.NotEmpty(t, forecastsA)

	// Verify User A's forecast contains their transaction data
	hasUserAData := false
	for _, f := range forecastsA {
		if f.Data.Revenue > 0 {
			hasUserAData = true
			// User A's forecast should only include their own 1000_00 transaction
			require.LessOrEqual(t, f.Data.Revenue, int64(1000_00))
			break
		}
	}
	require.True(t, hasUserAData, "User A should have forecast data")

	// User B should only see their own forecasts
	forecastsB, err := env.APIService.ListForecasts(env.UserB.ID, 12)
	require.NoError(t, err)
	require.NotEmpty(t, forecastsB)

	// Verify User B's forecast contains their transaction data
	hasUserBData := false
	for _, f := range forecastsB {
		if f.Data.Revenue > 0 {
			hasUserBData = true
			// User B's forecast should only include their own 2000_00 transaction
			require.LessOrEqual(t, f.Data.Revenue, int64(2000_00))
			break
		}
	}
	require.True(t, hasUserBData, "User B should have forecast data")
}

// TestCalculateForecast_OnlyUsesOwnOrgData verifies that forecast calculation
// only uses data from the user's own organisation
func TestCalculateForecast_OnlyUsesOwnOrgData(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Set database time to 2025-01-01 so our test dates are in the future
	simulatedTime := "2025-01-01"
	err := SetDatabaseTime(env.Conn, simulatedTime)
	require.NoError(t, err)
	parsedTime, err := time.Parse(utils.InternalDateFormat, simulatedTime)
	require.NoError(t, err)
	utils.DefaultClock.SetFixedTime(&parsedTime)
	defer utils.DefaultClock.SetFixedTime(nil)

	// Create large transaction for User A
	categoryA, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category A"}, &env.UserA.ID)
	require.NoError(t, err)

	_, err = env.APIService.CreateTransaction(models.CreateTransaction{
		Name:        "Large Transaction A",
		Amount:      100000_00, // 100,000.00
		Type:        "single",
		StartDate:   "2025-02-01",
		Category:    categoryA.ID,
		Currency:    *env.Currency.ID,
		VatIncluded: false,
	}, env.UserA.ID)
	require.NoError(t, err)

	// Create small transaction for User B
	categoryB, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category B"}, &env.UserB.ID)
	require.NoError(t, err)

	_, err = env.APIService.CreateTransaction(models.CreateTransaction{
		Name:        "Small Transaction B",
		Amount:      100_00, // 100.00
		Type:        "single",
		StartDate:   "2025-02-01",
		Category:    categoryB.ID,
		Currency:    *env.Currency.ID,
		VatIncluded: false,
	}, env.UserB.ID)
	require.NoError(t, err)

	// Calculate forecasts
	_, err = env.APIService.CalculateForecast(env.UserA.ID)
	require.NoError(t, err)

	_, err = env.APIService.CalculateForecast(env.UserB.ID)
	require.NoError(t, err)

	// Get forecasts
	forecastsA, err := env.APIService.ListForecasts(env.UserA.ID, 12)
	require.NoError(t, err)

	forecastsB, err := env.APIService.ListForecasts(env.UserB.ID, 12)
	require.NoError(t, err)

	// Find the February 2025 forecast for both users
	var febForecastA, febForecastB *models.Forecast
	for i := range forecastsA {
		if forecastsA[i].Data.Month == "2025-02" {
			febForecastA = &forecastsA[i]
			break
		}
	}
	for i := range forecastsB {
		if forecastsB[i].Data.Month == "2025-02" {
			febForecastB = &forecastsB[i]
			break
		}
	}

	require.NotNil(t, febForecastA, "User A should have February 2025 forecast")
	require.NotNil(t, febForecastB, "User B should have February 2025 forecast")

	// User A's forecast should have the large amount, User B should have small amount
	require.Equal(t, int64(100000_00), febForecastA.Data.Revenue)
	require.Equal(t, int64(100_00), febForecastB.Data.Revenue)
}

// TestListForecastDetails_CrossOrgIsolation verifies that users can only see
// forecast details belonging to their own organisation
func TestListForecastDetails_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Set database time to 2024-12-01 so our test dates are in the future
	simulatedTime := "2024-12-01"
	err := SetDatabaseTime(env.Conn, simulatedTime)
	require.NoError(t, err)
	parsedTime, err := time.Parse(utils.InternalDateFormat, simulatedTime)
	require.NoError(t, err)
	utils.DefaultClock.SetFixedTime(&parsedTime)
	defer utils.DefaultClock.SetFixedTime(nil)

	// Create employee and salary for User A (to generate salary forecast details)
	empA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	_, err = env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserA.ID, empA.ID)
	require.NoError(t, err)
	_ = empA // Mark as used

	// Create employee and salary for User B
	empB, err := CreateEmployee(env.APIService, env.UserB.ID, "Employee B")
	require.NoError(t, err)

	_, err = env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              6000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserB.ID, empB.ID)
	require.NoError(t, err)
	_ = empB // Mark as used

	// Calculate forecasts for both users
	_, err = env.APIService.CalculateForecast(env.UserA.ID)
	require.NoError(t, err)

	_, err = env.APIService.CalculateForecast(env.UserB.ID)
	require.NoError(t, err)

	// Get forecast details
	detailsA, err := env.APIService.ListForecastDetails(env.UserA.ID, 12)
	require.NoError(t, err)

	detailsB, err := env.APIService.ListForecastDetails(env.UserB.ID, 12)
	require.NoError(t, err)

	// Verify that each user gets their own forecast details (not empty)
	require.NotEmpty(t, detailsA, "User A should have forecast details")
	require.NotEmpty(t, detailsB, "User B should have forecast details")

	// Helper to recursively find all related IDs from expense items
	var findRelatedIDs func(expenses []models.ForecastDetailRevenueExpense, table string) []int64
	findRelatedIDs = func(expenses []models.ForecastDetailRevenueExpense, table string) []int64 {
		var ids []int64
		for _, e := range expenses {
			if e.RelatedTable == table && e.RelatedID != 0 {
				ids = append(ids, e.RelatedID)
			}
			if len(e.Children) > 0 {
				ids = append(ids, findRelatedIDs(e.Children, table)...)
			}
		}
		return ids
	}

	// Collect salary IDs from User A's forecast details
	var userASalaryIDs []int64
	for _, detail := range detailsA {
		userASalaryIDs = append(userASalaryIDs, findRelatedIDs(detail.Expense, utils.SalariesTableName)...)
	}

	// Collect salary IDs from User B's forecast details
	var userBSalaryIDs []int64
	for _, detail := range detailsB {
		userBSalaryIDs = append(userBSalaryIDs, findRelatedIDs(detail.Expense, utils.SalariesTableName)...)
	}

	// At least one salary from each user should be in their forecast details
	require.NotEmpty(t, userASalaryIDs, "User A's forecast should include salary data")
	require.NotEmpty(t, userBSalaryIDs, "User B's forecast should include salary data")

	// User A's salary IDs should not appear in User B's data and vice versa
	for _, idA := range userASalaryIDs {
		for _, idB := range userBSalaryIDs {
			require.NotEqual(t, idA, idB,
				"Users should not share salary IDs in their forecast details")
		}
	}
}
