package handlers_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
)

// TestListTransactions_CrossOrgIsolation verifies that users can only see transactions
// belonging to their own organisation
func TestListTransactions_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create categories for each org
	categoryA, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category A"}, &env.UserA.ID)
	require.NoError(t, err)
	categoryB, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category B"}, &env.UserB.ID)
	require.NoError(t, err)

	// Create transactions for User A's organisation
	txA1, err := env.APIService.CreateTransaction(models.CreateTransaction{
		Name:        "Transaction A1",
		Amount:      100_00,
		Type:        "single",
		StartDate:   "2025-01-01",
		Category:    categoryA.ID,
		Currency:    *env.Currency.ID,
		VatIncluded: false,
	}, env.UserA.ID)
	require.NoError(t, err)

	txA2, err := env.APIService.CreateTransaction(models.CreateTransaction{
		Name:        "Transaction A2",
		Amount:      200_00,
		Type:        "single",
		StartDate:   "2025-02-01",
		Category:    categoryA.ID,
		Currency:    *env.Currency.ID,
		VatIncluded: false,
	}, env.UserA.ID)
	require.NoError(t, err)

	// Create transactions for User B's organisation
	txB1, err := env.APIService.CreateTransaction(models.CreateTransaction{
		Name:        "Transaction B1",
		Amount:      300_00,
		Type:        "single",
		StartDate:   "2025-01-15",
		Category:    categoryB.ID,
		Currency:    *env.Currency.ID,
		VatIncluded: false,
	}, env.UserB.ID)
	require.NoError(t, err)

	// User A should only see their own transactions
	transactionsA, totalA, err := env.APIService.ListTransactions(env.UserA.ID, 1, 100, "name", "ASC", "", false)
	require.NoError(t, err)
	require.Equal(t, int64(2), totalA)
	require.Len(t, transactionsA, 2)

	txIDs := []int64{transactionsA[0].ID, transactionsA[1].ID}
	require.Contains(t, txIDs, txA1.ID)
	require.Contains(t, txIDs, txA2.ID)
	require.NotContains(t, txIDs, txB1.ID)

	// User B should only see their own transactions
	transactionsB, totalB, err := env.APIService.ListTransactions(env.UserB.ID, 1, 100, "name", "ASC", "", false)
	require.NoError(t, err)
	require.Equal(t, int64(1), totalB)
	require.Len(t, transactionsB, 1)
	require.Equal(t, txB1.ID, transactionsB[0].ID)
}

// TestGetTransaction_CrossOrgIsolation verifies that a user cannot fetch
// a transaction belonging to another organisation
func TestGetTransaction_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create category for User A
	categoryA, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category A"}, &env.UserA.ID)
	require.NoError(t, err)

	// Create transaction for User A's organisation
	txA, err := env.APIService.CreateTransaction(models.CreateTransaction{
		Name:        "Transaction A",
		Amount:      100_00,
		Type:        "single",
		StartDate:   "2025-01-01",
		Category:    categoryA.ID,
		Currency:    *env.Currency.ID,
		VatIncluded: false,
	}, env.UserA.ID)
	require.NoError(t, err)

	// User A can get their own transaction
	fetchedTx, err := env.APIService.GetTransaction(env.UserA.ID, txA.ID)
	require.NoError(t, err)
	require.Equal(t, txA.ID, fetchedTx.ID)
	require.Equal(t, "Transaction A", fetchedTx.Name)

	// User B cannot get User A's transaction (should return sql.ErrNoRows)
	_, err = env.APIService.GetTransaction(env.UserB.ID, txA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestUpdateTransaction_CrossOrgIsolation verifies that a user cannot update
// a transaction belonging to another organisation
func TestUpdateTransaction_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create category for User A
	categoryA, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category A"}, &env.UserA.ID)
	require.NoError(t, err)

	// Create transaction for User A's organisation
	txA, err := env.APIService.CreateTransaction(models.CreateTransaction{
		Name:        "Transaction A Original",
		Amount:      100_00,
		Type:        "single",
		StartDate:   "2025-01-01",
		Category:    categoryA.ID,
		Currency:    *env.Currency.ID,
		VatIncluded: false,
	}, env.UserA.ID)
	require.NoError(t, err)

	// User A can update their own transaction
	newNameA := "Transaction A Updated By A"
	_, err = env.APIService.UpdateTransaction(models.UpdateTransaction{
		Name: &newNameA,
	}, env.UserA.ID, txA.ID)
	require.NoError(t, err)

	// Verify the update worked
	updatedTx, err := env.APIService.GetTransaction(env.UserA.ID, txA.ID)
	require.NoError(t, err)
	require.Equal(t, "Transaction A Updated By A", updatedTx.Name)

	// User B attempts to update User A's transaction (should fail with ErrNoRows)
	maliciousName := "Hacked By B"
	_, err = env.APIService.UpdateTransaction(models.UpdateTransaction{
		Name: &maliciousName,
	}, env.UserB.ID, txA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)

	// Verify Transaction A's name was NOT changed by User B
	txAfterAttempt, err := env.APIService.GetTransaction(env.UserA.ID, txA.ID)
	require.NoError(t, err)
	require.Equal(t, "Transaction A Updated By A", txAfterAttempt.Name)
	require.NotEqual(t, "Hacked By B", txAfterAttempt.Name)
}

// TestDeleteTransaction_CrossOrgIsolation verifies that a user cannot delete
// a transaction belonging to another organisation
func TestDeleteTransaction_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create category for User A
	categoryA, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category A"}, &env.UserA.ID)
	require.NoError(t, err)

	// Create transaction for User A's organisation
	txA, err := env.APIService.CreateTransaction(models.CreateTransaction{
		Name:        "Transaction A To Delete",
		Amount:      100_00,
		Type:        "single",
		StartDate:   "2025-01-01",
		Category:    categoryA.ID,
		Currency:    *env.Currency.ID,
		VatIncluded: false,
	}, env.UserA.ID)
	require.NoError(t, err)

	// User B attempts to delete User A's transaction
	err = env.DBAdapter.DeleteTransaction(env.UserB.ID, txA.ID)
	// The delete should not return an error but should affect 0 rows

	// Verify Transaction A still exists and was NOT deleted
	txAfterDelete, err := env.APIService.GetTransaction(env.UserA.ID, txA.ID)
	require.NoError(t, err)
	require.NotNil(t, txAfterDelete)
	require.Equal(t, txA.ID, txAfterDelete.ID)

	// User A can successfully delete their own transaction
	err = env.DBAdapter.DeleteTransaction(env.UserA.ID, txA.ID)
	require.NoError(t, err)

	// Verify Transaction A is now deleted
	_, err = env.APIService.GetTransaction(env.UserA.ID, txA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestCreateTransaction_WithCrossOrgEmployee verifies that a user cannot create
// a transaction referencing an employee from another organisation
func TestCreateTransaction_WithCrossOrgEmployee(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employee for User A's organisation
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	// Create category for User B (to have a valid category)
	categoryB, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category B"}, &env.UserB.ID)
	require.NoError(t, err)

	// User B attempts to create a transaction referencing User A's employee
	employeeAID := employeeA.ID
	_, err = env.APIService.CreateTransaction(models.CreateTransaction{
		Name:        "Transaction with Cross-Org Employee",
		Amount:      100_00,
		Type:        "single",
		StartDate:   "2025-01-01",
		Category:    categoryB.ID,
		Currency:    *env.Currency.ID,
		VatIncluded: false,
		Employee:    &employeeAID,
	}, env.UserB.ID)
	// This should fail - either with an error or the employee reference should be ignored
	// The exact behavior depends on the implementation - we just verify isolation is maintained
	if err == nil {
		// If no error, verify the created transaction doesn't have the cross-org employee
		transactions, _, err := env.APIService.ListTransactions(env.UserB.ID, 1, 100, "name", "ASC", "", false)
		require.NoError(t, err)
		for _, tx := range transactions {
			if tx.Employee != nil {
				require.NotEqual(t, employeeA.ID, tx.Employee.ID,
					"Transaction should not reference employee from another organisation")
			}
		}
	}
	// If there's an error, that's also acceptable isolation behavior
}

// TestCreateTransaction_WithCrossOrgCategory verifies that users cannot use
// categories from another organisation
func TestCreateTransaction_WithCrossOrgCategory(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create org-specific categories
	categoryA, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category A"}, &env.UserA.ID)
	require.NoError(t, err)

	// User B attempts to use User A's category
	_, err = env.APIService.CreateTransaction(models.CreateTransaction{
		Name:        "Transaction with Cross-Org Category",
		Amount:      100_00,
		Type:        "single",
		StartDate:   "2025-01-01",
		Category:    categoryA.ID,
		Currency:    *env.Currency.ID,
		VatIncluded: false,
	}, env.UserB.ID)
	// This should fail or be rejected
	require.Error(t, err, "Should not be able to create transaction with category from another organisation")
}
