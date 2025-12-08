package handlers_test

import (
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAndUpdateUser(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	// Preparations
	_, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	user, organisation, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "john@doe.com", "test", "Test Organisation",
	)
	assert.NoError(t, err)

	userName := "Jane Doe"
	err = dbAdapter.UpdateProfile(models.UpdateUser{
		Name: &userName,
	}, user.ID)
	assert.NoError(t, err)

	user, err = dbAdapter.GetProfile(user.ID)
	assert.NoError(t, err)

	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "Jane Doe", user.Name)
	assert.Equal(t, organisation.ID, user.CurrentOrganisationID)
}

func TestSwitchOrganisationRestoresScenarioContext(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	// Create currency
	_, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	// Create user with first organization and scenario
	user, org1, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "user@test.com", "password", "Organization 1",
	)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotNil(t, org1)

	// Get the default scenario for org1
	scenarios1, err := apiService.ListScenarios(user.ID)
	assert.NoError(t, err)
	assert.Len(t, scenarios1, 1)
	scenario1Default := scenarios1[0]
	assert.Equal(t, "Standardszenario", scenario1Default.Name)
	assert.True(t, scenario1Default.IsDefault)

	// Create a second scenario for org1
	scenario1Custom, err := apiService.CreateScenario(models.CreateScenario{
		Name: "Custom Scenario Org1",
	}, user.ID)
	assert.NoError(t, err)

	// Verify current scenario changed
	user, err = dbAdapter.GetProfile(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, scenario1Custom.ID, user.CurrentScenarioID)

	// Create second organization for the same user
	org2, err := apiService.CreateOrganisation(models.CreateOrganisation{
		Name: "Organization 2",
	}, user.ID, false)
	assert.NoError(t, err)

	// Verify user switched to org2 and got the default scenario
	user, err = dbAdapter.GetProfile(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, org2.ID, user.CurrentOrganisationID)

	scenarios2, err := apiService.ListScenarios(user.ID)
	assert.NoError(t, err)
	assert.Len(t, scenarios2, 1)
	scenario2Default := scenarios2[0]
	assert.Equal(t, "Standardszenario", scenario2Default.Name)
	assert.Equal(t, scenario2Default.ID, user.CurrentScenarioID)

	// Create a custom scenario for org2
	scenario2Custom, err := apiService.CreateScenario(models.CreateScenario{
		Name: "Custom Scenario Org2",
	}, user.ID)
	assert.NoError(t, err)

	// Verify current scenario changed in org2
	user, err = dbAdapter.GetProfile(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, scenario2Custom.ID, user.CurrentScenarioID)

	// Switch back to org1
	err = apiService.SetUserCurrentOrganisation(models.UpdateUserCurrentOrganisation{
		OrganisationID: org1.ID,
	}, user.ID)
	assert.NoError(t, err)

	// Verify user is back in org1 AND the custom scenario is restored
	user, err = dbAdapter.GetProfile(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, org1.ID, user.CurrentOrganisationID)
	assert.Equal(t, scenario1Custom.ID, user.CurrentScenarioID, "Should restore the last-used scenario (Custom Scenario Org1) when switching back to org1")

	// Switch to org2 again
	err = apiService.SetUserCurrentOrganisation(models.UpdateUserCurrentOrganisation{
		OrganisationID: org2.ID,
	}, user.ID)
	assert.NoError(t, err)

	// Verify user is back in org2 AND the custom scenario is restored
	user, err = dbAdapter.GetProfile(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, org2.ID, user.CurrentOrganisationID)
	assert.Equal(t, scenario2Custom.ID, user.CurrentScenarioID, "Should restore the last-used scenario (Custom Scenario Org2) when switching back to org2")
}
