package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/pkg/models"
	"testing"
)

func TestCreateAndUpdateEmployee(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbService := db_service.NewDatabaseService(conn)

	// Preparations
	_, err := CreateCurrency(dbService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		dbService, "John Doe", "test", "Test Organisation",
	)
	assert.NoError(t, err)

	employee, err := CreateEmployee(dbService, user.ID, "Tom Riddle")
	assert.NoError(t, err)

	newName := "Peregrin Took"
	err = dbService.UpdateEmployee(models.UpdateEmployee{
		Name: &newName,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	employee, err = dbService.GetEmployee(user.ID, employee.ID)
	assert.NoError(t, err)

	assert.Equal(t, int64(1), employee.ID)
	assert.Equal(t, "Peregrin Took", employee.Name)
}
