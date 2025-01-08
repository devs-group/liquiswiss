package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/pkg/models"
	"testing"
)

func TestCurrencyOrderAndOrganisationDependency(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbService := db_service.NewDatabaseService(conn)

	// Preparations
	currencyCHF, err := CreateCurrency(dbService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)
	currencyVND, err := CreateCurrency(dbService, "VND", "Vietnamese Dong", "vi-VN")
	assert.NoError(t, err)
	currencyAED, err := CreateCurrency(dbService, "AED", "United Arab Emirates Dirham", "ar-AE")
	assert.NoError(t, err)

	user, organisation, err := CreateUserWithOrganisation(
		dbService, "John Doe", "test", "Test Organisation",
	)
	assert.NoError(t, err)

	// An organisation without a set currency will fall back to the system currency
	// Currently defined in 00009_create-function_get_default_system_currency_id.sql
	assert.Equal(t, *currencyCHF.Code, *organisation.Currency.Code)

	currencies, _, err := dbService.ListCurrencies(user.ID, 1, 250)
	assert.NoError(t, err)

	// The first currency in the list should be the company currency
	assert.Equal(t, *currencyCHF.Code, *currencies[0].Code)
	// Followed by the currencies sorted by Name ASC
	assert.Equal(t, *currencyAED.Code, *currencies[1].Code)
	assert.Equal(t, *currencyVND.Code, *currencies[2].Code)

	// We change the currency to VND
	err = dbService.UpdateOrganisation(models.UpdateOrganisation{
		Name:       nil,
		CurrencyID: currencyVND.ID,
	}, user.ID, organisation.ID)
	if err != nil {
		return
	}

	currencies, _, err = dbService.ListCurrencies(user.ID, 1, 250)
	assert.NoError(t, err)

	// The first currency should now be the newly set one, in this case VND
	assert.Equal(t, currencyVND.Code, currencies[0].Code)
	// Followed by the currencies sorted by Name ASC
	assert.Equal(t, currencyAED.Code, currencies[1].Code)
	assert.Equal(t, currencyCHF.Code, currencies[2].Code)
}
