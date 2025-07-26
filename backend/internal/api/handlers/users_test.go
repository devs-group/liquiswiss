package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"testing"
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
