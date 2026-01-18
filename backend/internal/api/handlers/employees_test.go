package handlers_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
)

func TestCreateAndUpdateEmployee(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	// Preparations
	_, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "john@doe.com", "test", "Test Organisation",
	)
	assert.NoError(t, err)

	employee, err := CreateEmployee(apiService, user.ID, "Tom Riddle")
	assert.NoError(t, err)

	newName := "Peregrin Took"
	err = dbAdapter.UpdateEmployee(models.UpdateEmployee{
		Name: &newName,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	employee, err = dbAdapter.GetEmployee(user.ID, employee.ID)
	assert.NoError(t, err)

	assert.Equal(t, int64(1), employee.ID)
	assert.Equal(t, "Peregrin Took", employee.Name)
}

func TestListEmployees_NoSearch(t *testing.T) {
	conn, apiService, user := setupEmployeeDependencies(t)
	defer conn.Close()

	// Create multiple employees
	_, err := CreateEmployee(apiService, user.ID, "Alice Smith")
	require.NoError(t, err)
	_, err = CreateEmployee(apiService, user.ID, "Bob Johnson")
	require.NoError(t, err)
	_, err = CreateEmployee(apiService, user.ID, "Charlie Brown")
	require.NoError(t, err)

	// List without search
	employees, total, err := apiService.ListEmployees(user.ID, 1, 100, "name", "ASC", "", false)
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, employees, 3)
}

func TestListEmployees_WithSearch(t *testing.T) {
	conn, apiService, user := setupEmployeeDependencies(t)
	defer conn.Close()

	// Create multiple employees
	_, err := CreateEmployee(apiService, user.ID, "Alice Smith")
	require.NoError(t, err)
	_, err = CreateEmployee(apiService, user.ID, "Bob Johnson")
	require.NoError(t, err)
	_, err = CreateEmployee(apiService, user.ID, "Charlie Brown")
	require.NoError(t, err)

	// Search for "Alice"
	employees, total, err := apiService.ListEmployees(user.ID, 1, 100, "name", "ASC", "Alice", false)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, employees, 1)
	require.Equal(t, "Alice Smith", employees[0].Name)
}

func TestListEmployees_SearchCaseInsensitive(t *testing.T) {
	conn, apiService, user := setupEmployeeDependencies(t)
	defer conn.Close()

	// Create employee with mixed case
	_, err := CreateEmployee(apiService, user.ID, "Alice Smith")
	require.NoError(t, err)

	// Search with lowercase
	employees, total, err := apiService.ListEmployees(user.ID, 1, 100, "name", "ASC", "alice", false)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, employees, 1)
	require.Equal(t, "Alice Smith", employees[0].Name)

	// Search with uppercase
	employees, total, err = apiService.ListEmployees(user.ID, 1, 100, "name", "ASC", "ALICE", false)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, employees, 1)
}

func TestListEmployees_SearchNoResults(t *testing.T) {
	conn, apiService, user := setupEmployeeDependencies(t)
	defer conn.Close()

	_, err := CreateEmployee(apiService, user.ID, "Alice Smith")
	require.NoError(t, err)

	// Search for non-existent term
	employees, total, err := apiService.ListEmployees(user.ID, 1, 100, "name", "ASC", "nonexistent", false)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Len(t, employees, 0)
}

func setupEmployeeDependencies(t *testing.T) (*sql.DB, api_service.IAPIService, *models.User) {
	t.Helper()

	conn := SetupTestEnvironment(t)

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	_, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	require.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "employee.search@test.com", "test", "Employee Search Org",
	)
	require.NoError(t, err)

	return conn, apiService, user
}

func setupEmployeeDependenciesWithCurrency(t *testing.T) (*sql.DB, api_service.IAPIService, *models.User, *models.Currency) {
	t.Helper()

	conn := SetupTestEnvironment(t)

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	currency, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	require.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "employee.sort@test.com", "test", "Employee Sort Org",
	)
	require.NoError(t, err)

	return conn, apiService, user, currency
}

func TestListEmployees_SortByName(t *testing.T) {
	conn, apiService, user := setupEmployeeDependencies(t)
	defer conn.Close()

	_, err := CreateEmployee(apiService, user.ID, "Charlie")
	require.NoError(t, err)
	_, err = CreateEmployee(apiService, user.ID, "Alpha")
	require.NoError(t, err)
	_, err = CreateEmployee(apiService, user.ID, "Bravo")
	require.NoError(t, err)

	// ASC
	employees, _, err := apiService.ListEmployees(user.ID, 1, 100, "name", "ASC", "", false)
	require.NoError(t, err)
	require.Equal(t, "Alpha", employees[0].Name)
	require.Equal(t, "Bravo", employees[1].Name)
	require.Equal(t, "Charlie", employees[2].Name)

	// DESC
	employees, _, err = apiService.ListEmployees(user.ID, 1, 100, "name", "DESC", "", false)
	require.NoError(t, err)
	require.Equal(t, "Charlie", employees[0].Name)
	require.Equal(t, "Bravo", employees[1].Name)
	require.Equal(t, "Alpha", employees[2].Name)
}

func TestListEmployees_SortByHoursPerMonth(t *testing.T) {
	conn, apiService, user, currency := setupEmployeeDependenciesWithCurrency(t)
	defer conn.Close()

	empA, err := CreateEmployee(apiService, user.ID, "Medium Hours")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       100,
		Amount:              5000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 20,
		FromDate:            "2025-01-01",
	}, user.ID, empA.ID)
	require.NoError(t, err)

	empB, err := CreateEmployee(apiService, user.ID, "Few Hours")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       50,
		Amount:              2500_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 20,
		FromDate:            "2025-01-01",
	}, user.ID, empB.ID)
	require.NoError(t, err)

	empC, err := CreateEmployee(apiService, user.ID, "Many Hours")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              8000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 20,
		FromDate:            "2025-01-01",
	}, user.ID, empC.ID)
	require.NoError(t, err)

	// ASC
	employees, _, err := apiService.ListEmployees(user.ID, 1, 100, "hoursPerMonth", "ASC", "", false)
	require.NoError(t, err)
	require.Equal(t, "Few Hours", employees[0].Name)
	require.Equal(t, "Medium Hours", employees[1].Name)
	require.Equal(t, "Many Hours", employees[2].Name)

	// DESC
	employees, _, err = apiService.ListEmployees(user.ID, 1, 100, "hoursPerMonth", "DESC", "", false)
	require.NoError(t, err)
	require.Equal(t, "Many Hours", employees[0].Name)
	require.Equal(t, "Medium Hours", employees[1].Name)
	require.Equal(t, "Few Hours", employees[2].Name)
}

func TestListEmployees_SortBySalary(t *testing.T) {
	conn, apiService, user, currency := setupEmployeeDependenciesWithCurrency(t)
	defer conn.Close()

	empA, err := CreateEmployee(apiService, user.ID, "Medium Salary")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 20,
		FromDate:            "2025-01-01",
	}, user.ID, empA.ID)
	require.NoError(t, err)

	empB, err := CreateEmployee(apiService, user.ID, "Low Salary")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              3000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 20,
		FromDate:            "2025-01-01",
	}, user.ID, empB.ID)
	require.NoError(t, err)

	empC, err := CreateEmployee(apiService, user.ID, "High Salary")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              8000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 20,
		FromDate:            "2025-01-01",
	}, user.ID, empC.ID)
	require.NoError(t, err)

	// ASC
	employees, _, err := apiService.ListEmployees(user.ID, 1, 100, "salary", "ASC", "", false)
	require.NoError(t, err)
	require.Equal(t, "Low Salary", employees[0].Name)
	require.Equal(t, "Medium Salary", employees[1].Name)
	require.Equal(t, "High Salary", employees[2].Name)

	// DESC
	employees, _, err = apiService.ListEmployees(user.ID, 1, 100, "salary", "DESC", "", false)
	require.NoError(t, err)
	require.Equal(t, "High Salary", employees[0].Name)
	require.Equal(t, "Medium Salary", employees[1].Name)
	require.Equal(t, "Low Salary", employees[2].Name)
}

func TestListEmployees_SortByVacationDaysPerYear(t *testing.T) {
	conn, apiService, user, currency := setupEmployeeDependenciesWithCurrency(t)
	defer conn.Close()

	empA, err := CreateEmployee(apiService, user.ID, "Medium Vacation")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, user.ID, empA.ID)
	require.NoError(t, err)

	empB, err := CreateEmployee(apiService, user.ID, "Few Vacation")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 15,
		FromDate:            "2025-01-01",
	}, user.ID, empB.ID)
	require.NoError(t, err)

	empC, err := CreateEmployee(apiService, user.ID, "Many Vacation")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 30,
		FromDate:            "2025-01-01",
	}, user.ID, empC.ID)
	require.NoError(t, err)

	// ASC
	employees, _, err := apiService.ListEmployees(user.ID, 1, 100, "vacationDaysPerYear", "ASC", "", false)
	require.NoError(t, err)
	require.Equal(t, "Few Vacation", employees[0].Name)
	require.Equal(t, "Medium Vacation", employees[1].Name)
	require.Equal(t, "Many Vacation", employees[2].Name)

	// DESC
	employees, _, err = apiService.ListEmployees(user.ID, 1, 100, "vacationDaysPerYear", "DESC", "", false)
	require.NoError(t, err)
	require.Equal(t, "Many Vacation", employees[0].Name)
	require.Equal(t, "Medium Vacation", employees[1].Name)
	require.Equal(t, "Few Vacation", employees[2].Name)
}

func TestListEmployees_SortByFromDate(t *testing.T) {
	conn, apiService, user, currency := setupEmployeeDependenciesWithCurrency(t)
	defer conn.Close()

	empA, err := CreateEmployee(apiService, user.ID, "Middle Start")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 20,
		FromDate:            "2025-06-01",
	}, user.ID, empA.ID)
	require.NoError(t, err)

	empB, err := CreateEmployee(apiService, user.ID, "Early Start")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 20,
		FromDate:            "2025-01-01",
	}, user.ID, empB.ID)
	require.NoError(t, err)

	empC, err := CreateEmployee(apiService, user.ID, "Late Start")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 20,
		FromDate:            "2025-12-01",
	}, user.ID, empC.ID)
	require.NoError(t, err)

	// ASC
	employees, _, err := apiService.ListEmployees(user.ID, 1, 100, "fromDate", "ASC", "", false)
	require.NoError(t, err)
	require.Equal(t, "Early Start", employees[0].Name)
	require.Equal(t, "Middle Start", employees[1].Name)
	require.Equal(t, "Late Start", employees[2].Name)

	// DESC
	employees, _, err = apiService.ListEmployees(user.ID, 1, 100, "fromDate", "DESC", "", false)
	require.NoError(t, err)
	require.Equal(t, "Late Start", employees[0].Name)
	require.Equal(t, "Middle Start", employees[1].Name)
	require.Equal(t, "Early Start", employees[2].Name)
}

func TestListEmployees_SortByToDate(t *testing.T) {
	conn, apiService, user, currency := setupEmployeeDependenciesWithCurrency(t)
	defer conn.Close()

	// Set database time to before all toDate values so salaries are active
	err := SetDatabaseTime(conn, "2025-01-15")
	require.NoError(t, err)

	toDate1 := "2025-06-30"
	toDate2 := "2025-03-31"
	toDate3 := "2025-12-31"

	empA, err := CreateEmployee(apiService, user.ID, "Middle End")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 20,
		FromDate:            "2025-01-01",
		ToDate:              &toDate1,
	}, user.ID, empA.ID)
	require.NoError(t, err)

	empB, err := CreateEmployee(apiService, user.ID, "Early End")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 20,
		FromDate:            "2025-01-01",
		ToDate:              &toDate2,
	}, user.ID, empB.ID)
	require.NoError(t, err)

	empC, err := CreateEmployee(apiService, user.ID, "Late End")
	require.NoError(t, err)
	_, err = apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 20,
		FromDate:            "2025-01-01",
		ToDate:              &toDate3,
	}, user.ID, empC.ID)
	require.NoError(t, err)

	// ASC
	employees, _, err := apiService.ListEmployees(user.ID, 1, 100, "toDate", "ASC", "", false)
	require.NoError(t, err)
	require.Equal(t, "Early End", employees[0].Name)
	require.Equal(t, "Middle End", employees[1].Name)
	require.Equal(t, "Late End", employees[2].Name)

	// DESC
	employees, _, err = apiService.ListEmployees(user.ID, 1, 100, "toDate", "DESC", "", false)
	require.NoError(t, err)
	require.Equal(t, "Late End", employees[0].Name)
	require.Equal(t, "Middle End", employees[1].Name)
	require.Equal(t, "Early End", employees[2].Name)
}

func TestListEmployees_InvalidSortBy(t *testing.T) {
	conn, apiService, user := setupEmployeeDependencies(t)
	defer conn.Close()

	_, err := CreateEmployee(apiService, user.ID, "Test Employee")
	require.NoError(t, err)

	_, _, err = apiService.ListEmployees(user.ID, 1, 100, "invalidField", "ASC", "", false)
	require.Error(t, err)
}

func TestListEmployees_InvalidSortOrder(t *testing.T) {
	conn, apiService, user := setupEmployeeDependencies(t)
	defer conn.Close()

	_, err := CreateEmployee(apiService, user.ID, "Test Employee")
	require.NoError(t, err)

	_, _, err = apiService.ListEmployees(user.ID, 1, 100, "name", "INVALID", "", false)
	require.Error(t, err)
}
