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
