package handlers_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
)

func TestCreateHorizontalScenarioWithoutCopying(t *testing.T) {
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

	// Create horizontal scenario without parent
	scenario, err := apiService.CreateScenario(models.CreateScenario{
		Name: "Scenario A",
		Type: models.ScenarioTypeHorizontal,
	}, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, scenario)
	assert.Equal(t, "Scenario A", scenario.Name)
	assert.Equal(t, models.ScenarioTypeHorizontal, scenario.Type)
	assert.False(t, scenario.IsDefault)
	assert.Nil(t, scenario.ParentScenarioID)
}

func TestCreateHorizontalScenarioWithCopying(t *testing.T) {
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

	// Get default scenario
	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)

	// Create employee in default scenario
	_, err = CreateEmployee(apiService, user.ID, "John Doe")
	assert.NoError(t, err)

	// Create another employee in default scenario
	_, err = CreateEmployee(apiService, user.ID, "Jane Smith")
	assert.NoError(t, err)

	// Create horizontal scenario with copying from default
	scenario, err := apiService.CreateScenario(models.CreateScenario{
		Name:             "Scenario B",
		Type:             models.ScenarioTypeHorizontal,
		ParentScenarioID: &defaultScenario.ID,
	}, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, scenario)

	// Verify employees in default scenario (user is still in default scenario)
	defaultEmployees, _, err := dbAdapter.ListEmployees(user.ID, 1, 100, "name", "ASC")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(defaultEmployees))
	assert.Equal(t, defaultScenario.ID, defaultEmployees[0].ScenarioID)

	// Switch to new scenario
	err = apiService.SetUserCurrentScenario(models.UpdateUserCurrentScenario{
		ScenarioID: scenario.ID,
	}, user.ID)
	assert.NoError(t, err)

	// Verify employees were copied to new scenario
	copiedEmployees, _, err := dbAdapter.ListEmployees(user.ID, 1, 100, "name", "ASC")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(copiedEmployees))
	assert.Equal(t, scenario.ID, copiedEmployees[0].ScenarioID)

	// Verify UUIDs are different between default and copied employees
	for _, copiedEmp := range copiedEmployees {
		for _, defaultEmp := range defaultEmployees {
			if copiedEmp.Name == defaultEmp.Name {
				assert.NotEqual(t, defaultEmp.UUID, copiedEmp.UUID)
			}
		}
	}
}

func TestCreateVerticalScenario(t *testing.T) {
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

	// Get default scenario (horizontal)
	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)

	// Create vertical scenario with default as parent
	verticalScenario, err := apiService.CreateScenario(models.CreateScenario{
		Name:             "Vertical Scenario",
		Type:             models.ScenarioTypeVertical,
		ParentScenarioID: &defaultScenario.ID,
	}, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, verticalScenario)
	assert.Equal(t, "Vertical Scenario", verticalScenario.Name)
	assert.Equal(t, models.ScenarioTypeVertical, verticalScenario.Type)
	assert.Equal(t, defaultScenario.ID, *verticalScenario.ParentScenarioID)
	assert.False(t, verticalScenario.IsDefault)
}

func TestSwitchScenario(t *testing.T) {
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

	// Get default scenario
	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)

	// Verify user's current scenario is default
	assert.NotNil(t, user.CurrentScenarioID)
	assert.Equal(t, defaultScenario.ID, *user.CurrentScenarioID)

	// Create a new horizontal scenario
	newScenario, err := apiService.CreateScenario(models.CreateScenario{
		Name: "Scenario C",
		Type: models.ScenarioTypeHorizontal,
	}, user.ID)
	assert.NoError(t, err)

	// Switch to new scenario
	err = apiService.SetUserCurrentScenario(models.UpdateUserCurrentScenario{
		ScenarioID: newScenario.ID,
	}, user.ID)
	assert.NoError(t, err)

	// Verify scenario was switched
	updatedUser, err := dbAdapter.GetProfile(user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser.CurrentScenarioID)
	assert.Equal(t, newScenario.ID, *updatedUser.CurrentScenarioID)

	// Switch back to default
	err = apiService.SetUserCurrentScenario(models.UpdateUserCurrentScenario{
		ScenarioID: defaultScenario.ID,
	}, user.ID)
	assert.NoError(t, err)

	// Verify switched back
	updatedUser, err = dbAdapter.GetProfile(user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser.CurrentScenarioID)
	assert.Equal(t, defaultScenario.ID, *updatedUser.CurrentScenarioID)
}

func TestDataIsolationBetweenHorizontalScenarios(t *testing.T) {
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

	// Get default scenario
	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)

	// Create employee in default scenario
	_, err = CreateEmployee(apiService, user.ID, "Employee in Default")
	assert.NoError(t, err)

	// Create horizontal scenario (empty)
	horizontalScenario, err := apiService.CreateScenario(models.CreateScenario{
		Name: "Empty Scenario",
		Type: models.ScenarioTypeHorizontal,
	}, user.ID)
	assert.NoError(t, err)

	// List employees - should see employee in default scenario
	employees, _, err := dbAdapter.ListEmployees(user.ID, 1, 100, "name", "ASC")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(employees))
	assert.Equal(t, defaultScenario.ID, employees[0].ScenarioID)

	// Switch to horizontal scenario
	err = apiService.SetUserCurrentScenario(models.UpdateUserCurrentScenario{
		ScenarioID: horizontalScenario.ID,
	}, user.ID)
	assert.NoError(t, err)

	// Create employee in horizontal scenario
	_, err = CreateEmployee(apiService, user.ID, "Employee in Horizontal")
	assert.NoError(t, err)

	// List employees in horizontal scenario - should only see 1 employee
	horizontalEmployees, _, err := dbAdapter.ListEmployees(user.ID, 1, 100, "name", "ASC")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(horizontalEmployees))
	assert.Equal(t, horizontalScenario.ID, horizontalEmployees[0].ScenarioID)
	assert.Equal(t, "Employee in Horizontal", horizontalEmployees[0].Name)

	// Switch back to default scenario
	err = apiService.SetUserCurrentScenario(models.UpdateUserCurrentScenario{
		ScenarioID: defaultScenario.ID,
	}, user.ID)
	assert.NoError(t, err)

	// List employees in default scenario - should only see 1 employee
	defaultEmployees, _, err := dbAdapter.ListEmployees(user.ID, 1, 100, "name", "ASC")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(defaultEmployees))
	assert.Equal(t, defaultScenario.ID, defaultEmployees[0].ScenarioID)
	assert.Equal(t, "Employee in Default", defaultEmployees[0].Name)
}

func TestListScenarios(t *testing.T) {
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

	// List scenarios - should have default scenario
	scenarios, err := apiService.ListScenarios(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(scenarios))
	assert.True(t, scenarios[0].IsDefault)

	// Create more scenarios
	_, err = apiService.CreateScenario(models.CreateScenario{
		Name: "Scenario 1",
		Type: models.ScenarioTypeHorizontal,
	}, user.ID)
	assert.NoError(t, err)

	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)

	_, err = apiService.CreateScenario(models.CreateScenario{
		Name:             "Scenario 2",
		Type:             models.ScenarioTypeVertical,
		ParentScenarioID: &defaultScenario.ID,
	}, user.ID)
	assert.NoError(t, err)

	// List scenarios - should have 3 now
	scenarios, err = apiService.ListScenarios(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(scenarios))

	// Verify default scenario is first
	assert.True(t, scenarios[0].IsDefault)
}

func TestRenameDefaultScenario(t *testing.T) {
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

	// Get default scenario
	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)
	assert.True(t, defaultScenario.IsDefault)

	// Rename default scenario
	newName := "My Custom Default Scenario"
	updatedScenario, err := apiService.UpdateScenario(models.UpdateScenario{
		Name: &newName,
	}, user.ID, defaultScenario.ID)
	assert.NoError(t, err)
	assert.Equal(t, "My Custom Default Scenario", updatedScenario.Name)
	assert.True(t, updatedScenario.IsDefault, "Default flag should remain true")
}

func TestRenameVerticalScenario(t *testing.T) {
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

	// Get default scenario
	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)

	// Create vertical scenario (child)
	verticalScenario, err := apiService.CreateScenario(models.CreateScenario{
		Name:             "Original Vertical Name",
		Type:             models.ScenarioTypeVertical,
		ParentScenarioID: &defaultScenario.ID,
	}, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.ScenarioTypeVertical, verticalScenario.Type)

	// Rename vertical scenario
	newName := "Renamed Vertical Scenario"
	updatedScenario, err := apiService.UpdateScenario(models.UpdateScenario{
		Name: &newName,
	}, user.ID, verticalScenario.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Renamed Vertical Scenario", updatedScenario.Name)
	assert.Equal(t, models.ScenarioTypeVertical, updatedScenario.Type)
	assert.Equal(t, defaultScenario.ID, *updatedScenario.ParentScenarioID)
}

func TestRenameHorizontalNonDefaultScenario(t *testing.T) {
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

	// Create horizontal non-default scenario
	scenario, err := apiService.CreateScenario(models.CreateScenario{
		Name: "Original Horizontal Name",
		Type: models.ScenarioTypeHorizontal,
	}, user.ID)
	assert.NoError(t, err)
	assert.False(t, scenario.IsDefault)
	assert.Equal(t, models.ScenarioTypeHorizontal, scenario.Type)

	// Rename horizontal scenario
	newName := "Renamed Horizontal Scenario"
	updatedScenario, err := apiService.UpdateScenario(models.UpdateScenario{
		Name: &newName,
	}, user.ID, scenario.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Renamed Horizontal Scenario", updatedScenario.Name)
	assert.False(t, updatedScenario.IsDefault)
	assert.Equal(t, models.ScenarioTypeHorizontal, updatedScenario.Type)
}

func TestDeleteVerticalScenario(t *testing.T) {
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

	// Get default scenario
	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)

	// Create vertical scenario (child)
	verticalScenario, err := apiService.CreateScenario(models.CreateScenario{
		Name:             "Vertical To Delete",
		Type:             models.ScenarioTypeVertical,
		ParentScenarioID: &defaultScenario.ID,
	}, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.ScenarioTypeVertical, verticalScenario.Type)
	assert.False(t, verticalScenario.IsDefault)

	// Verify we have 2 scenarios now
	scenariosBeforeDelete, err := apiService.ListScenarios(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(scenariosBeforeDelete))

	// Delete vertical scenario
	err = apiService.DeleteScenario(user.ID, verticalScenario.ID)
	assert.NoError(t, err)

	// Verify scenario is deleted
	scenariosAfterDelete, err := apiService.ListScenarios(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(scenariosAfterDelete)) // Only default remains
	assert.True(t, scenariosAfterDelete[0].IsDefault)

	// Try to get deleted scenario - should fail
	_, err = dbAdapter.GetScenario(user.ID, verticalScenario.ID)
	assert.Error(t, err)
}

func TestDeleteHorizontalNonDefaultScenario(t *testing.T) {
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

	// Create horizontal non-default scenario
	scenario, err := apiService.CreateScenario(models.CreateScenario{
		Name: "Horizontal To Delete",
		Type: models.ScenarioTypeHorizontal,
	}, user.ID)
	assert.NoError(t, err)
	assert.False(t, scenario.IsDefault)
	assert.Equal(t, models.ScenarioTypeHorizontal, scenario.Type)

	// Verify we have 2 scenarios now
	scenariosBeforeDelete, err := apiService.ListScenarios(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(scenariosBeforeDelete))

	// Delete horizontal scenario
	err = apiService.DeleteScenario(user.ID, scenario.ID)
	assert.NoError(t, err)

	// Verify scenario is deleted
	scenariosAfterDelete, err := apiService.ListScenarios(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(scenariosAfterDelete)) // Only default remains
	assert.True(t, scenariosAfterDelete[0].IsDefault)

	// Try to get deleted scenario - should fail
	_, err = dbAdapter.GetScenario(user.ID, scenario.ID)
	assert.Error(t, err)
}

func TestCannotDeleteDefaultScenario(t *testing.T) {
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

	// Get default scenario
	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)

	// Try to delete default scenario - should fail
	err = apiService.DeleteScenario(user.ID, defaultScenario.ID)
	assert.Error(t, err)

	// Verify default scenario still exists
	scenarios, err := apiService.ListScenarios(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(scenarios))
	assert.True(t, scenarios[0].IsDefault)
}

func TestForecastIsolationBetweenScenarios(t *testing.T) {
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

	// Get default scenario
	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)

	// Create forecast in default scenario
	defaultForecastID, err := dbAdapter.UpsertForecast(models.CreateForecast{
		Month:    "2025-12",
		Revenue:  10000,
		Expense:  5000,
		Cashflow: 5000,
	}, user.ID)
	assert.NoError(t, err)
	assert.NotZero(t, defaultForecastID)

	// List forecasts in default scenario - should see the forecast
	defaultForecasts, err := dbAdapter.ListForecasts(user.ID, 12)
	assert.NoError(t, err)
	foundDefault := false
	for _, f := range defaultForecasts {
		if f.Data.Month == "2025-12" && f.UpdatedAt != nil && f.Data.Revenue == 10000 {
			foundDefault = true
			break
		}
	}
	assert.True(t, foundDefault, "Forecast should exist in default scenario")

	// Create horizontal scenario (empty)
	horizontalScenario, err := apiService.CreateScenario(models.CreateScenario{
		Name: "Scenario B",
		Type: models.ScenarioTypeHorizontal,
	}, user.ID)
	assert.NoError(t, err)

	// Switch to horizontal scenario
	err = apiService.SetUserCurrentScenario(models.UpdateUserCurrentScenario{
		ScenarioID: horizontalScenario.ID,
	}, user.ID)
	assert.NoError(t, err)

	// List forecasts in horizontal scenario - should NOT see default scenario's forecast
	horizontalForecasts, err := dbAdapter.ListForecasts(user.ID, 12)
	assert.NoError(t, err)
	foundInHorizontal := false
	for _, f := range horizontalForecasts {
		if f.Data.Month == "2025-12" && f.UpdatedAt != nil && f.Data.Revenue == 10000 {
			foundInHorizontal = true
			break
		}
	}
	assert.False(t, foundInHorizontal, "Default scenario's forecast should NOT appear in horizontal scenario")

	// Create different forecast in horizontal scenario
	horizontalForecastID, err := dbAdapter.UpsertForecast(models.CreateForecast{
		Month:    "2025-12",
		Revenue:  20000,
		Expense:  8000,
		Cashflow: 12000,
	}, user.ID)
	assert.NoError(t, err)
	assert.NotZero(t, horizontalForecastID)
	assert.NotEqual(t, defaultForecastID, horizontalForecastID, "Forecast IDs should be different")

	// Verify horizontal scenario has its own forecast
	horizontalForecastsAfter, err := dbAdapter.ListForecasts(user.ID, 12)
	assert.NoError(t, err)
	foundHorizontalForecast := false
	for _, f := range horizontalForecastsAfter {
		if f.Data.Month == "2025-12" && f.UpdatedAt != nil && f.Data.Revenue == 20000 {
			foundHorizontalForecast = true
			break
		}
	}
	assert.True(t, foundHorizontalForecast, "Horizontal scenario should have its own forecast")

	// Switch back to default scenario
	err = apiService.SetUserCurrentScenario(models.UpdateUserCurrentScenario{
		ScenarioID: defaultScenario.ID,
	}, user.ID)
	assert.NoError(t, err)

	// Verify default scenario still has original forecast, not horizontal's
	defaultForecastsAfter, err := dbAdapter.ListForecasts(user.ID, 12)
	assert.NoError(t, err)
	foundOriginalDefault := false
	foundHorizontalInDefault := false
	for _, f := range defaultForecastsAfter {
		if f.Data.Month == "2025-12" && f.UpdatedAt != nil {
			if f.Data.Revenue == 10000 {
				foundOriginalDefault = true
			}
			if f.Data.Revenue == 20000 {
				foundHorizontalInDefault = true
			}
		}
	}
	assert.True(t, foundOriginalDefault, "Default scenario should still have original forecast")
	assert.False(t, foundHorizontalInDefault, "Default scenario should NOT have horizontal scenario's forecast")
}

func TestForecastDetailsIsolationBetweenScenarios(t *testing.T) {
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

	// Get default scenario
	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)

	// Create forecast with details in default scenario
	defaultForecastID, err := dbAdapter.UpsertForecast(models.CreateForecast{
		Month:    "2026-01",
		Revenue:  15000,
		Expense:  7000,
		Cashflow: 8000,
	}, user.ID)
	assert.NoError(t, err)

	// Create forecast details for default scenario
	defaultDetailID, err := dbAdapter.UpsertForecastDetail(models.CreateForecastDetail{
		Month: "2026-01",
		Revenue: []models.ForecastDetailRevenueExpense{
			{Name: "Product Sales", Amount: 15000},
		},
		Expense: []models.ForecastDetailRevenueExpense{
			{Name: "Office Rent", Amount: 7000},
		},
	}, user.ID, defaultForecastID)
	assert.NoError(t, err)
	assert.NotZero(t, defaultDetailID)

	// List forecast details in default scenario
	defaultDetails, err := dbAdapter.ListForecastDetails(user.ID, 12)
	assert.NoError(t, err)
	foundDefaultDetail := false
	for _, d := range defaultDetails {
		if d.Month == "2026-01" && len(d.Revenue) > 0 && d.Revenue[0].Name == "Product Sales" {
			foundDefaultDetail = true
			break
		}
	}
	assert.True(t, foundDefaultDetail, "Forecast details should exist in default scenario")

	// Create horizontal scenario
	horizontalScenario, err := apiService.CreateScenario(models.CreateScenario{
		Name: "Scenario C",
		Type: models.ScenarioTypeHorizontal,
	}, user.ID)
	assert.NoError(t, err)

	// Switch to horizontal scenario
	err = apiService.SetUserCurrentScenario(models.UpdateUserCurrentScenario{
		ScenarioID: horizontalScenario.ID,
	}, user.ID)
	assert.NoError(t, err)

	// List forecast details in horizontal scenario - should NOT see default's details
	horizontalDetails, err := dbAdapter.ListForecastDetails(user.ID, 12)
	assert.NoError(t, err)
	foundInHorizontal := false
	for _, d := range horizontalDetails {
		if d.Month == "2026-01" && len(d.Revenue) > 0 {
			foundInHorizontal = true
			break
		}
	}
	assert.False(t, foundInHorizontal, "Default scenario's forecast details should NOT appear in horizontal scenario")

	// Create different forecast and details in horizontal scenario
	horizontalForecastID, err := dbAdapter.UpsertForecast(models.CreateForecast{
		Month:    "2026-01",
		Revenue:  25000,
		Expense:  12000,
		Cashflow: 13000,
	}, user.ID)
	assert.NoError(t, err)

	horizontalDetailID, err := dbAdapter.UpsertForecastDetail(models.CreateForecastDetail{
		Month: "2026-01",
		Revenue: []models.ForecastDetailRevenueExpense{
			{Name: "Service Revenue", Amount: 25000},
		},
		Expense: []models.ForecastDetailRevenueExpense{
			{Name: "Marketing", Amount: 12000},
		},
	}, user.ID, horizontalForecastID)
	assert.NoError(t, err)
	assert.NotZero(t, horizontalDetailID)
	assert.NotEqual(t, defaultDetailID, horizontalDetailID, "Forecast detail IDs should be different")

	// Verify horizontal scenario has its own details
	horizontalDetailsAfter, err := dbAdapter.ListForecastDetails(user.ID, 12)
	assert.NoError(t, err)
	foundHorizontalDetail := false
	for _, d := range horizontalDetailsAfter {
		if d.Month == "2026-01" && len(d.Revenue) > 0 && d.Revenue[0].Name == "Service Revenue" {
			foundHorizontalDetail = true
			break
		}
	}
	assert.True(t, foundHorizontalDetail, "Horizontal scenario should have its own forecast details")

	// Switch back to default scenario
	err = apiService.SetUserCurrentScenario(models.UpdateUserCurrentScenario{
		ScenarioID: defaultScenario.ID,
	}, user.ID)
	assert.NoError(t, err)

	// Verify default scenario still has original details
	defaultDetailsAfter, err := dbAdapter.ListForecastDetails(user.ID, 12)
	assert.NoError(t, err)
	foundOriginalDefaultDetail := false
	foundHorizontalInDefault := false
	for _, d := range defaultDetailsAfter {
		if d.Month == "2026-01" && len(d.Revenue) > 0 {
			if d.Revenue[0].Name == "Product Sales" {
				foundOriginalDefaultDetail = true
			}
			if d.Revenue[0].Name == "Service Revenue" {
				foundHorizontalInDefault = true
			}
		}
	}
	assert.True(t, foundOriginalDefaultDetail, "Default scenario should still have original details")
	assert.False(t, foundHorizontalInDefault, "Default scenario should NOT have horizontal scenario's details")
}

func TestForecastsNotCopiedToNewScenario(t *testing.T) {
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

	// Get default scenario
	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)

	// Create multiple forecasts in default scenario
	for i := 1; i <= 3; i++ {
		month := fmt.Sprintf("2026-%02d", i)
		_, err := dbAdapter.UpsertForecast(models.CreateForecast{
			Month:    month,
			Revenue:  int64(i * 10000),
			Expense:  int64(i * 5000),
			Cashflow: int64(i * 5000),
		}, user.ID)
		assert.NoError(t, err)
	}

	// Verify forecasts exist in default scenario
	defaultForecasts, err := dbAdapter.ListForecasts(user.ID, 12)
	assert.NoError(t, err)
	forecastCount := 0
	for _, f := range defaultForecasts {
		if f.UpdatedAt != nil && f.Data.Revenue > 0 {
			forecastCount++
		}
	}
	assert.Equal(t, 3, forecastCount, "Default scenario should have 3 forecasts")

	// Create horizontal scenario with copy data option
	horizontalScenario, err := apiService.CreateScenario(models.CreateScenario{
		Name:             "Scenario with Copy",
		Type:             models.ScenarioTypeHorizontal,
		ParentScenarioID: &defaultScenario.ID, // Copy from default
	}, user.ID)
	assert.NoError(t, err)

	// Switch to new scenario
	err = apiService.SetUserCurrentScenario(models.UpdateUserCurrentScenario{
		ScenarioID: horizontalScenario.ID,
	}, user.ID)
	assert.NoError(t, err)

	// Verify forecasts are NOT copied (only employees, salaries, transactions should be copied)
	horizontalForecasts, err := dbAdapter.ListForecasts(user.ID, 12)
	assert.NoError(t, err)
	horizontalForecastCount := 0
	for _, f := range horizontalForecasts {
		if f.UpdatedAt != nil && f.Data.Revenue > 0 {
			horizontalForecastCount++
		}
	}
	assert.Equal(t, 0, horizontalForecastCount, "Forecasts should NOT be copied to new scenario")
}

func TestDeleteActiveScenarioWithoutParent(t *testing.T) {
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

	// Get default scenario
	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)

	// Create a horizontal scenario without parent
	horizontalScenario, err := apiService.CreateScenario(models.CreateScenario{
		Name: "Horizontal Scenario",
		Type: models.ScenarioTypeHorizontal,
	}, user.ID)
	assert.NoError(t, err)
	assert.False(t, horizontalScenario.IsDefault)
	assert.Nil(t, horizontalScenario.ParentScenarioID)

	// Switch user to horizontal scenario
	err = apiService.SetUserCurrentScenario(models.UpdateUserCurrentScenario{
		ScenarioID: horizontalScenario.ID,
	}, user.ID)
	assert.NoError(t, err)

	// Verify user is on horizontal scenario
	updatedUser, err := dbAdapter.GetProfile(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, horizontalScenario.ID, *updatedUser.CurrentScenarioID)

	// Delete the horizontal scenario while it's active (should auto-switch to default)
	err = apiService.DeleteScenario(user.ID, horizontalScenario.ID)
	assert.NoError(t, err)

	// Verify user was auto-switched to default scenario
	updatedUser, err = dbAdapter.GetProfile(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, defaultScenario.ID, *updatedUser.CurrentScenarioID)

	// Verify scenario is deleted
	_, err = dbAdapter.GetScenario(user.ID, horizontalScenario.ID)
	assert.Error(t, err)
}

func TestDeleteActiveScenarioWithParent(t *testing.T) {
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

	// Get default scenario
	defaultScenario, err := dbAdapter.GetDefaultScenario(user.ID)
	assert.NoError(t, err)

	// Create a vertical scenario (child of default)
	verticalScenario, err := apiService.CreateScenario(models.CreateScenario{
		Name:             "Vertical Scenario",
		Type:             models.ScenarioTypeVertical,
		ParentScenarioID: &defaultScenario.ID,
	}, user.ID)
	assert.NoError(t, err)
	assert.False(t, verticalScenario.IsDefault)
	assert.NotNil(t, verticalScenario.ParentScenarioID)
	assert.Equal(t, defaultScenario.ID, *verticalScenario.ParentScenarioID)

	// Switch user to vertical scenario
	err = apiService.SetUserCurrentScenario(models.UpdateUserCurrentScenario{
		ScenarioID: verticalScenario.ID,
	}, user.ID)
	assert.NoError(t, err)

	// Verify user is on vertical scenario
	updatedUser, err := dbAdapter.GetProfile(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, verticalScenario.ID, *updatedUser.CurrentScenarioID)

	// Delete the vertical scenario while it's active (should auto-switch to parent)
	err = apiService.DeleteScenario(user.ID, verticalScenario.ID)
	assert.NoError(t, err)

	// Verify user was auto-switched to parent (default) scenario
	updatedUser, err = dbAdapter.GetProfile(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, defaultScenario.ID, *updatedUser.CurrentScenarioID)

	// Verify scenario is deleted
	_, err = dbAdapter.GetScenario(user.ID, verticalScenario.ID)
	assert.Error(t, err)
}
