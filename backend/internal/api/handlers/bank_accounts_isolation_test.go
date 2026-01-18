package handlers_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
)

// TestListBankAccounts_CrossOrgIsolation verifies that users can only see bank accounts
// belonging to their own organisation
func TestListBankAccounts_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create bank accounts for User A's organisation
	baA1, err := env.APIService.CreateBankAccount(models.CreateBankAccount{
		Name:     "Account A1",
		Amount:   100000,
		Currency: *env.Currency.ID,
	}, env.UserA.ID)
	require.NoError(t, err)

	baA2, err := env.APIService.CreateBankAccount(models.CreateBankAccount{
		Name:     "Account A2",
		Amount:   200000,
		Currency: *env.Currency.ID,
	}, env.UserA.ID)
	require.NoError(t, err)

	// Create bank accounts for User B's organisation
	baB1, err := env.APIService.CreateBankAccount(models.CreateBankAccount{
		Name:     "Account B1",
		Amount:   300000,
		Currency: *env.Currency.ID,
	}, env.UserB.ID)
	require.NoError(t, err)

	// User A should only see their own bank accounts
	accountsA, totalA, err := env.APIService.ListBankAccounts(env.UserA.ID, 1, 100, "name", "ASC", "")
	require.NoError(t, err)
	require.Equal(t, int64(2), totalA)
	require.Len(t, accountsA, 2)

	accountIDs := []int64{accountsA[0].ID, accountsA[1].ID}
	require.Contains(t, accountIDs, baA1.ID)
	require.Contains(t, accountIDs, baA2.ID)
	require.NotContains(t, accountIDs, baB1.ID)

	// User B should only see their own bank accounts
	accountsB, totalB, err := env.APIService.ListBankAccounts(env.UserB.ID, 1, 100, "name", "ASC", "")
	require.NoError(t, err)
	require.Equal(t, int64(1), totalB)
	require.Len(t, accountsB, 1)
	require.Equal(t, baB1.ID, accountsB[0].ID)
}

// TestGetBankAccount_CrossOrgIsolation verifies that a user cannot fetch
// a bank account belonging to another organisation
func TestGetBankAccount_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create bank account for User A's organisation
	baA, err := env.APIService.CreateBankAccount(models.CreateBankAccount{
		Name:     "Account A",
		Amount:   100000,
		Currency: *env.Currency.ID,
	}, env.UserA.ID)
	require.NoError(t, err)

	// User A can get their own bank account
	fetchedBA, err := env.APIService.GetBankAccount(env.UserA.ID, baA.ID)
	require.NoError(t, err)
	require.Equal(t, baA.ID, fetchedBA.ID)
	require.Equal(t, "Account A", fetchedBA.Name)

	// User B cannot get User A's bank account (should return sql.ErrNoRows)
	_, err = env.APIService.GetBankAccount(env.UserB.ID, baA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestUpdateBankAccount_CrossOrgIsolation verifies that a user cannot update
// a bank account belonging to another organisation
func TestUpdateBankAccount_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create bank account for User A's organisation
	baA, err := env.APIService.CreateBankAccount(models.CreateBankAccount{
		Name:     "Account A Original",
		Amount:   100000,
		Currency: *env.Currency.ID,
	}, env.UserA.ID)
	require.NoError(t, err)

	// User A can update their own bank account
	newNameA := "Account A Updated By A"
	_, err = env.APIService.UpdateBankAccount(models.UpdateBankAccount{
		Name: &newNameA,
	}, env.UserA.ID, baA.ID)
	require.NoError(t, err)

	// Verify the update worked
	updatedBA, err := env.APIService.GetBankAccount(env.UserA.ID, baA.ID)
	require.NoError(t, err)
	require.Equal(t, "Account A Updated By A", updatedBA.Name)

	// User B attempts to update User A's bank account (should fail with ErrNoRows)
	maliciousName := "Hacked By B"
	_, err = env.APIService.UpdateBankAccount(models.UpdateBankAccount{
		Name: &maliciousName,
	}, env.UserB.ID, baA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)

	// Verify Bank Account A's name was NOT changed by User B
	baAfterAttempt, err := env.APIService.GetBankAccount(env.UserA.ID, baA.ID)
	require.NoError(t, err)
	require.Equal(t, "Account A Updated By A", baAfterAttempt.Name)
	require.NotEqual(t, "Hacked By B", baAfterAttempt.Name)
}

// TestDeleteBankAccount_CrossOrgIsolation verifies that a user cannot delete
// a bank account belonging to another organisation
func TestDeleteBankAccount_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create bank account for User A's organisation
	baA, err := env.APIService.CreateBankAccount(models.CreateBankAccount{
		Name:     "Account A To Delete",
		Amount:   100000,
		Currency: *env.Currency.ID,
	}, env.UserA.ID)
	require.NoError(t, err)

	// User B attempts to delete User A's bank account
	err = env.DBAdapter.DeleteBankAccount(env.UserB.ID, baA.ID)
	// The delete should not return an error but should affect 0 rows

	// Verify Bank Account A still exists and was NOT deleted
	baAfterDelete, err := env.APIService.GetBankAccount(env.UserA.ID, baA.ID)
	require.NoError(t, err)
	require.NotNil(t, baAfterDelete)
	require.Equal(t, baA.ID, baAfterDelete.ID)

	// User A can successfully delete their own bank account
	err = env.DBAdapter.DeleteBankAccount(env.UserA.ID, baA.ID)
	require.NoError(t, err)

	// Verify Bank Account A is now deleted
	_, err = env.APIService.GetBankAccount(env.UserA.ID, baA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}
