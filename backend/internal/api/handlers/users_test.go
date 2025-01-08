package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/pkg/models"
	"testing"
)

func TestCreateAndUpdateUser(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbService := db_service.NewDatabaseService(conn)

	// Preparations
	_, err := CreateCurrency(dbService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	user, organisation, err := CreateUserWithOrganisation(
		dbService, "John Doe", "test", "Test Organisation",
	)
	assert.NoError(t, err)

	userName := "Jane Doe"
	err = dbService.UpdateProfile(models.UpdateUser{
		Name: &userName,
	}, user.ID)
	assert.NoError(t, err)

	user, err = dbService.GetProfile(user.ID)
	assert.NoError(t, err)

	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "Jane Doe", user.Name)
	assert.Equal(t, organisation.ID, user.CurrentOrganisationID)
}
