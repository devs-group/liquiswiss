package handlers_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"liquiswiss/pkg/utils"
)

func TestUpdateTransaction_SetEndDate(t *testing.T) {
	conn, apiService, dbAdapter, user, category, currency := setupTransactionDependencies(t)
	defer conn.Close()

	transaction := createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil)

	endDate := "2025-06-30"
	updated, err := apiService.UpdateTransaction(models.UpdateTransaction{
		EndDate: &endDate,
	}, user.ID, transaction.ID)
	require.NoError(t, err)

	assertDateEquals(t, updated.EndDate, endDate)

	stored, err := dbAdapter.GetTransaction(user.ID, transaction.ID)
	require.NoError(t, err)
	assertDateEquals(t, stored.EndDate, endDate)
}

func TestUpdateTransaction_RemoveEndDate(t *testing.T) {
	conn, apiService, dbAdapter, user, category, currency := setupTransactionDependencies(t)
	defer conn.Close()

	initialEndDate := "2025-07-15"
	transaction := createTransaction(t, apiService, user.ID, category.ID, *currency.ID, &initialEndDate)
	assertDateEquals(t, transaction.EndDate, initialEndDate)

	updated, err := apiService.UpdateTransaction(models.UpdateTransaction{}, user.ID, transaction.ID)
	require.NoError(t, err)
	require.Nil(t, updated.EndDate)

	stored, err := dbAdapter.GetTransaction(user.ID, transaction.ID)
	require.NoError(t, err)
	require.Nil(t, stored.EndDate)
}

func TestUpdateTransaction_DisableKeepsEndDate(t *testing.T) {
	conn, apiService, dbAdapter, user, category, currency := setupTransactionDependencies(t)
	defer conn.Close()

	initialEndDate := "2025-09-01"
	transaction := createTransaction(t, apiService, user.ID, category.ID, *currency.ID, &initialEndDate)
	assertDateEquals(t, transaction.EndDate, initialEndDate)

	isDisabled := true
	updated, err := apiService.UpdateTransaction(models.UpdateTransaction{
		IsDisabled: &isDisabled,
	}, user.ID, transaction.ID)
	require.NoError(t, err)
	require.True(t, updated.IsDisabled)
	assertDateEquals(t, updated.EndDate, initialEndDate)

	stored, err := dbAdapter.GetTransaction(user.ID, transaction.ID)
	require.NoError(t, err)
	require.True(t, stored.IsDisabled)
	assertDateEquals(t, stored.EndDate, initialEndDate)
}

func setupTransactionDependencies(t *testing.T) (*sql.DB, api_service.IAPIService, db_adapter.IDatabaseAdapter, *models.User, *models.Category, *models.Currency) {
	t.Helper()

	conn := SetupTestEnvironment(t)

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	currency, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	require.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "john.transaction@doe.com", "test", "Transaction Org",
	)
	require.NoError(t, err)

	category, err := apiService.CreateCategory(models.CreateCategory{Name: "Services"}, &user.ID)
	require.NoError(t, err)

	return conn, apiService, dbAdapter, user, category, currency
}

func createTransaction(t *testing.T, apiService api_service.IAPIService, userID int64, categoryID int64, currencyID int64, endDate *string) *models.Transaction {
	t.Helper()

	transaction, err := apiService.CreateTransaction(models.CreateTransaction{
		Name:        "Retainer",
		Amount:      120_00,
		Type:        "single",
		StartDate:   "2025-01-01",
		EndDate:     endDate,
		Category:    categoryID,
		Currency:    currencyID,
		VatIncluded: false,
	}, userID)
	require.NoError(t, err)

	return transaction
}

func assertDateEquals(t *testing.T, actual *types.AsDate, expected string) {
	t.Helper()

	require.NotNil(t, actual)
	require.Equal(t, expected, time.Time(*actual).Format(utils.InternalDateFormat))
}
