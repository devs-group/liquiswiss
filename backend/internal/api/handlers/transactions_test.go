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

func TestListTransactions_NoSearch(t *testing.T) {
	conn, apiService, _, user, category, currency := setupTransactionDependencies(t)
	defer conn.Close()

	// Create multiple transactions
	createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil, func(p *models.CreateTransaction) {
		p.Name = "Office Rent"
	})
	createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil, func(p *models.CreateTransaction) {
		p.Name = "Software License"
	})
	createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil, func(p *models.CreateTransaction) {
		p.Name = "Bank Fees"
	})

	// List without search
	transactions, total, err := apiService.ListTransactions(user.ID, 1, 100, "name", "ASC", "", false)
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, transactions, 3)
}

func TestListTransactions_WithSearch(t *testing.T) {
	conn, apiService, _, user, category, currency := setupTransactionDependencies(t)
	defer conn.Close()

	// Create multiple transactions
	createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil, func(p *models.CreateTransaction) {
		p.Name = "Office Rent"
	})
	createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil, func(p *models.CreateTransaction) {
		p.Name = "Software License"
	})
	createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil, func(p *models.CreateTransaction) {
		p.Name = "Bank Fees"
	})

	// Search for "Office"
	transactions, total, err := apiService.ListTransactions(user.ID, 1, 100, "name", "ASC", "Office", false)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, transactions, 1)
	require.Equal(t, "Office Rent", transactions[0].Name)
}

func TestListTransactions_SearchCaseInsensitive(t *testing.T) {
	conn, apiService, _, user, category, currency := setupTransactionDependencies(t)
	defer conn.Close()

	// Create transaction with mixed case
	createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil, func(p *models.CreateTransaction) {
		p.Name = "Office Rent"
	})

	// Search with lowercase
	transactions, total, err := apiService.ListTransactions(user.ID, 1, 100, "name", "ASC", "office", false)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, transactions, 1)
	require.Equal(t, "Office Rent", transactions[0].Name)

	// Search with uppercase
	transactions, total, err = apiService.ListTransactions(user.ID, 1, 100, "name", "ASC", "OFFICE", false)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, transactions, 1)
}

func TestListTransactions_SearchNoResults(t *testing.T) {
	conn, apiService, _, user, category, currency := setupTransactionDependencies(t)
	defer conn.Close()

	createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil, func(p *models.CreateTransaction) {
		p.Name = "Office Rent"
	})

	// Search for non-existent term
	transactions, total, err := apiService.ListTransactions(user.ID, 1, 100, "name", "ASC", "nonexistent", false)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Len(t, transactions, 0)
}

func TestListTransactions_HideDisabled(t *testing.T) {
	conn, apiService, _, user, category, currency := setupTransactionDependencies(t)
	defer conn.Close()

	// Create enabled transactions
	createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil, func(p *models.CreateTransaction) {
		p.Name = "Active Transaction 1"
	})
	createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil, func(p *models.CreateTransaction) {
		p.Name = "Active Transaction 2"
	})

	// Create a disabled transaction
	disabledTx := createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil, func(p *models.CreateTransaction) {
		p.Name = "Disabled Transaction"
	})

	// Disable the transaction
	isDisabled := true
	_, err := apiService.UpdateTransaction(models.UpdateTransaction{
		IsDisabled: &isDisabled,
	}, user.ID, disabledTx.ID)
	require.NoError(t, err)

	// With hideDisabled=false, should see all 3 transactions
	transactions, total, err := apiService.ListTransactions(user.ID, 1, 100, "name", "ASC", "", false)
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, transactions, 3)

	// With hideDisabled=true, should only see 2 active transactions
	transactions, total, err = apiService.ListTransactions(user.ID, 1, 100, "name", "ASC", "", true)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, transactions, 2)

	// Verify the disabled transaction is not in the list
	for _, tx := range transactions {
		require.NotEqual(t, "Disabled Transaction", tx.Name)
		require.False(t, tx.IsDisabled)
	}
}

func TestListTransactions_HideDisabledWithSearch(t *testing.T) {
	conn, apiService, _, user, category, currency := setupTransactionDependencies(t)
	defer conn.Close()

	// Create transactions with similar names
	createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil, func(p *models.CreateTransaction) {
		p.Name = "Rent Payment Active"
	})

	disabledTx := createTransaction(t, apiService, user.ID, category.ID, *currency.ID, nil, func(p *models.CreateTransaction) {
		p.Name = "Rent Payment Disabled"
	})

	// Disable one transaction
	isDisabled := true
	_, err := apiService.UpdateTransaction(models.UpdateTransaction{
		IsDisabled: &isDisabled,
	}, user.ID, disabledTx.ID)
	require.NoError(t, err)

	// Search for "Rent" with hideDisabled=false - should find 2
	transactions, total, err := apiService.ListTransactions(user.ID, 1, 100, "name", "ASC", "Rent", false)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, transactions, 2)

	// Search for "Rent" with hideDisabled=true - should find only 1
	transactions, total, err = apiService.ListTransactions(user.ID, 1, 100, "name", "ASC", "Rent", true)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, transactions, 1)
	require.Equal(t, "Rent Payment Active", transactions[0].Name)
}

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

func TestUpdateTransaction_DisableKeepsExistingFields(t *testing.T) {
	conn, apiService, dbAdapter, user, category, currency := setupTransactionDependencies(t)
	defer conn.Close()

	initialEndDate := "2025-09-01"
	cycle := utils.CycleMonthly
	employee, err := CreateEmployee(apiService, user.ID, "Transaction Owner")
	require.NoError(t, err)
	employeeID := employee.ID

	vat, err := apiService.CreateVat(models.CreateVat{Value: 770}, user.ID)
	require.NoError(t, err)
	vatID := vat.ID

	transaction := createTransaction(
		t,
		apiService,
		user.ID,
		category.ID,
		*currency.ID,
		&initialEndDate,
		func(payload *models.CreateTransaction) {
			payload.Type = "repeating"
			payload.Cycle = &cycle
			payload.Employee = &employeeID
			payload.Vat = &vatID
			payload.VatIncluded = true
		},
	)
	assertDateEquals(t, transaction.EndDate, initialEndDate)
	require.NotNil(t, transaction.Cycle)
	require.Equal(t, cycle, *transaction.Cycle)
	require.NotNil(t, transaction.Employee)
	require.Equal(t, employeeID, transaction.Employee.ID)
	require.NotNil(t, transaction.Vat)
	require.Equal(t, vatID, transaction.Vat.ID)

	isDisabled := true
	updated, err := apiService.UpdateTransaction(models.UpdateTransaction{
		IsDisabled: &isDisabled,
	}, user.ID, transaction.ID)
	require.NoError(t, err)
	require.True(t, updated.IsDisabled)
	assertDateEquals(t, updated.EndDate, initialEndDate)
	require.NotNil(t, updated.Cycle)
	require.Equal(t, cycle, *updated.Cycle)
	require.NotNil(t, updated.Employee)
	require.Equal(t, employeeID, updated.Employee.ID)
	require.NotNil(t, updated.Vat)
	require.Equal(t, vatID, updated.Vat.ID)
	require.True(t, updated.VatIncluded)

	stored, err := dbAdapter.GetTransaction(user.ID, transaction.ID)
	require.NoError(t, err)
	require.True(t, stored.IsDisabled)
	assertDateEquals(t, stored.EndDate, initialEndDate)
	require.NotNil(t, stored.Cycle)
	require.Equal(t, cycle, *stored.Cycle)
	require.NotNil(t, stored.Employee)
	require.Equal(t, employeeID, stored.Employee.ID)
	require.NotNil(t, stored.Vat)
	require.Equal(t, vatID, stored.Vat.ID)
	require.True(t, stored.VatIncluded)
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

func createTransaction(
	t *testing.T,
	apiService api_service.IAPIService,
	userID int64,
	categoryID int64,
	currencyID int64,
	endDate *string,
	options ...func(*models.CreateTransaction),
) *models.Transaction {
	t.Helper()

	payload := models.CreateTransaction{
		Name:        "Retainer",
		Amount:      120_00,
		Type:        "single",
		StartDate:   "2025-01-01",
		EndDate:     endDate,
		Category:    categoryID,
		Currency:    currencyID,
		VatIncluded: false,
	}

	for _, option := range options {
		option(&payload)
	}

	transaction, err := apiService.CreateTransaction(payload, userID)
	require.NoError(t, err)

	return transaction
}

func assertDateEquals(t *testing.T, actual *types.AsDate, expected string) {
	t.Helper()

	require.NotNil(t, actual)
	require.Equal(t, expected, time.Time(*actual).Format(utils.InternalDateFormat))
}
