package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"testing"
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
