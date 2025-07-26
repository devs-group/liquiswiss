package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"testing"
)

func TestCurrencyOrderAndOrganisationDependency(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	// Preparations
	currencyCHF, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)
	currencyVND, err := CreateCurrency(apiService, "VND", "Vietnamese Dong", "vi-VN")
	assert.NoError(t, err)
	currencyAED, err := CreateCurrency(apiService, "AED", "United Arab Emirates Dirham", "ar-AE")
	assert.NoError(t, err)

	user, organisation, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "john@doe.com", "test", "Test Organisation",
	)
	assert.NoError(t, err)

	// An organisation without a set currency will fall back to the system currency
	// Currently defined in 00004_create-function_get_default_system_currency_id.sql
	assert.Equal(t, *currencyCHF.Code, *organisation.Currency.Code)

	currencies, err := dbAdapter.ListCurrencies(user.ID)
	assert.NoError(t, err)

	// The first currency in the list should be the company currency
	assert.Equal(t, *currencyCHF.Code, *currencies[0].Code)
	// Followed by the currencies sorted by Name ASC
	assert.Equal(t, *currencyAED.Code, *currencies[1].Code)
	assert.Equal(t, *currencyVND.Code, *currencies[2].Code)

	// We change the currency to VND
	err = dbAdapter.UpdateOrganisation(models.UpdateOrganisation{
		Name:       nil,
		CurrencyID: currencyVND.ID,
	}, user.ID, organisation.ID)
	if err != nil {
		return
	}

	currencies, err = dbAdapter.ListCurrencies(user.ID)
	assert.NoError(t, err)

	// The first currency should now be the newly set one, in this case VND
	assert.Equal(t, currencyVND.Code, currencies[0].Code)
	// Followed by the currencies sorted by Name ASC
	assert.Equal(t, currencyAED.Code, currencies[1].Code)
	assert.Equal(t, currencyCHF.Code, currencies[2].Code)
}
