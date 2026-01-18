package handlers_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
)

func TestListBankAccounts_NoSearch(t *testing.T) {
	conn, apiService, user, currency := setupBankAccountDependencies(t)
	defer conn.Close()

	// Create multiple bank accounts
	createBankAccount(t, apiService, user.ID, *currency.ID, "Main Account")
	createBankAccount(t, apiService, user.ID, *currency.ID, "Savings Account")
	createBankAccount(t, apiService, user.ID, *currency.ID, "Business Account")

	// List without search
	bankAccounts, total, err := apiService.ListBankAccounts(user.ID, 1, 100, "name", "ASC", "")
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, bankAccounts, 3)
}

func TestListBankAccounts_WithSearch(t *testing.T) {
	conn, apiService, user, currency := setupBankAccountDependencies(t)
	defer conn.Close()

	// Create multiple bank accounts
	createBankAccount(t, apiService, user.ID, *currency.ID, "Main Account")
	createBankAccount(t, apiService, user.ID, *currency.ID, "Savings Account")
	createBankAccount(t, apiService, user.ID, *currency.ID, "Business Account")

	// Search for "Savings"
	bankAccounts, total, err := apiService.ListBankAccounts(user.ID, 1, 100, "name", "ASC", "Savings")
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, bankAccounts, 1)
	require.Equal(t, "Savings Account", bankAccounts[0].Name)
}

func TestListBankAccounts_SearchCaseInsensitive(t *testing.T) {
	conn, apiService, user, currency := setupBankAccountDependencies(t)
	defer conn.Close()

	// Create bank account with mixed case
	createBankAccount(t, apiService, user.ID, *currency.ID, "Savings Account")

	// Search with lowercase
	bankAccounts, total, err := apiService.ListBankAccounts(user.ID, 1, 100, "name", "ASC", "savings")
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, bankAccounts, 1)
	require.Equal(t, "Savings Account", bankAccounts[0].Name)

	// Search with uppercase
	bankAccounts, total, err = apiService.ListBankAccounts(user.ID, 1, 100, "name", "ASC", "SAVINGS")
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, bankAccounts, 1)
}

func TestListBankAccounts_SearchNoResults(t *testing.T) {
	conn, apiService, user, currency := setupBankAccountDependencies(t)
	defer conn.Close()

	createBankAccount(t, apiService, user.ID, *currency.ID, "Main Account")

	// Search for non-existent term
	bankAccounts, total, err := apiService.ListBankAccounts(user.ID, 1, 100, "name", "ASC", "nonexistent")
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Len(t, bankAccounts, 0)
}

func setupBankAccountDependencies(t *testing.T) (*sql.DB, api_service.IAPIService, *models.User, *models.Currency) {
	t.Helper()

	conn := SetupTestEnvironment(t)

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	currency, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	require.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "bankaccount.search@test.com", "test", "Bank Account Search Org",
	)
	require.NoError(t, err)

	return conn, apiService, user, currency
}

func createBankAccount(t *testing.T, apiService api_service.IAPIService, userID int64, currencyID int64, name string) *models.BankAccount {
	t.Helper()

	bankAccount, err := apiService.CreateBankAccount(models.CreateBankAccount{
		Name:     name,
		Amount:   100000,
		Currency: currencyID,
	}, userID)
	require.NoError(t, err)

	return bankAccount
}
