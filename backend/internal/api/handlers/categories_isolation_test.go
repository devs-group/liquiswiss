package handlers_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
)

// TestListCategories_ShowsSystemAndOwnOrg verifies that users see system categories and
// their own organisation's categories, but not other organisation's categories
func TestListCategories_ShowsSystemAndOwnOrg(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create org-specific categories for User A
	catA1, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category A1"}, &env.UserA.ID)
	require.NoError(t, err)

	catA2, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category A2"}, &env.UserA.ID)
	require.NoError(t, err)

	// Create org-specific category for User B
	catB1, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category B1"}, &env.UserB.ID)
	require.NoError(t, err)

	// User A should see their own categories (and any pre-existing system categories)
	catsA, totalA, err := env.APIService.ListCategories(env.UserA.ID, 1, 100)
	require.NoError(t, err)
	require.GreaterOrEqual(t, totalA, int64(2)) // At least 2 own

	catAIDs := make([]int64, 0)
	for _, c := range catsA {
		catAIDs = append(catAIDs, c.ID)
	}
	require.Contains(t, catAIDs, catA1.ID)
	require.Contains(t, catAIDs, catA2.ID)
	require.NotContains(t, catAIDs, catB1.ID)

	// User B should see their own categories (and any pre-existing system categories)
	catsB, totalB, err := env.APIService.ListCategories(env.UserB.ID, 1, 100)
	require.NoError(t, err)
	require.GreaterOrEqual(t, totalB, int64(1)) // At least 1 own

	catBIDs := make([]int64, 0)
	for _, c := range catsB {
		catBIDs = append(catBIDs, c.ID)
	}
	require.Contains(t, catBIDs, catB1.ID)
	require.NotContains(t, catBIDs, catA1.ID)
	require.NotContains(t, catBIDs, catA2.ID)
}

// TestGetCategory_CrossOrgIsolation verifies that a user cannot fetch
// a category belonging to another organisation
func TestGetCategory_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create org-specific category for User A
	catA, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category A"}, &env.UserA.ID)
	require.NoError(t, err)

	// User A can get their own category
	fetchedCat, err := env.APIService.GetCategory(env.UserA.ID, catA.ID)
	require.NoError(t, err)
	require.Equal(t, catA.ID, fetchedCat.ID)
	require.Equal(t, "Category A", fetchedCat.Name)

	// User B cannot get User A's category (should return sql.ErrNoRows)
	_, err = env.APIService.GetCategory(env.UserB.ID, catA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestUpdateCategory_CrossOrgIsolation verifies that a user cannot update
// a category belonging to another organisation
func TestUpdateCategory_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create org-specific category for User A
	catA, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Category A Original"}, &env.UserA.ID)
	require.NoError(t, err)

	// User A can update their own category
	newNameA := "Category A Updated"
	_, err = env.APIService.UpdateCategory(models.UpdateCategory{Name: &newNameA}, env.UserA.ID, catA.ID)
	require.NoError(t, err)

	// Verify the update worked
	updatedCat, err := env.APIService.GetCategory(env.UserA.ID, catA.ID)
	require.NoError(t, err)
	require.Equal(t, "Category A Updated", updatedCat.Name)

	// User B attempts to update User A's category (should fail with ErrNoRows)
	maliciousName := "Hacked By B"
	_, err = env.APIService.UpdateCategory(models.UpdateCategory{Name: &maliciousName}, env.UserB.ID, catA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)

	// Verify category was NOT changed by User B
	catAfterAttempt, err := env.APIService.GetCategory(env.UserA.ID, catA.ID)
	require.NoError(t, err)
	require.Equal(t, "Category A Updated", catAfterAttempt.Name)
	require.NotEqual(t, "Hacked By B", catAfterAttempt.Name)
}

// TestDeleteCategory_SentinelErrors verifies that blocked deletes return the
// typed errors the REST handler maps to 409 Conflict
func TestDeleteCategory_SentinelErrors(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Global preset (no owner organisation) must not be deletable
	globalCat, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Global Preset"}, nil)
	require.NoError(t, err)
	err = env.APIService.DeleteCategory(env.UserA.ID, globalCat.ID)
	require.ErrorIs(t, err, api_service.ErrCategoryGlobal)

	// Own category still used by a transaction must not be deletable
	ownCat, err := env.APIService.CreateCategory(models.CreateCategory{Name: "In Use"}, &env.UserA.ID)
	require.NoError(t, err)
	transaction, err := env.APIService.CreateTransaction(models.CreateTransaction{
		Name:        "Uses Category",
		Amount:      100_00,
		Type:        "single",
		StartDate:   "2025-01-01",
		Category:    ownCat.ID,
		Currency:    *env.Currency.ID,
		VatIncluded: false,
	}, env.UserA.ID)
	require.NoError(t, err)
	err = env.APIService.DeleteCategory(env.UserA.ID, ownCat.ID)
	require.ErrorIs(t, err, api_service.ErrCategoryInUse)

	// After removing the transaction the category can be deleted
	err = env.APIService.DeleteTransaction(env.UserA.ID, transaction.ID)
	require.NoError(t, err)
	err = env.APIService.DeleteCategory(env.UserA.ID, ownCat.ID)
	require.NoError(t, err)
}

// TestReassignCategoryTransactions verifies bulk relinking of transactions to
// another category, including cross-org isolation of both source and target
func TestReassignCategoryTransactions(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	source, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Reassign Source"}, &env.UserA.ID)
	require.NoError(t, err)
	target, err := env.APIService.CreateCategory(models.CreateCategory{Name: "Reassign Target"}, &env.UserA.ID)
	require.NoError(t, err)

	for i := range 2 {
		_, err = env.APIService.CreateTransaction(models.CreateTransaction{
			Name:        "Reassign TX",
			Amount:      int64(100_00 + i),
			Type:        "single",
			StartDate:   "2025-01-01",
			Category:    source.ID,
			Currency:    *env.Currency.ID,
			VatIncluded: false,
		}, env.UserA.ID)
		require.NoError(t, err)
	}

	// User B must not be able to touch User A's categories
	_, err = env.APIService.ReassignCategoryTransactions(env.UserB.ID, source.ID, target.ID)
	require.Error(t, err)

	// Same source and target is rejected
	_, err = env.APIService.ReassignCategoryTransactions(env.UserA.ID, source.ID, source.ID)
	require.Error(t, err)

	// Happy path: both transactions move, then the source can be deleted
	affected, err := env.APIService.ReassignCategoryTransactions(env.UserA.ID, source.ID, target.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), affected)

	count, err := env.DBAdapter.CountTransactionsWithCategory(env.UserA.ID, source.ID)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)
	count, err = env.DBAdapter.CountTransactionsWithCategory(env.UserA.ID, target.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)

	require.NoError(t, env.APIService.DeleteCategory(env.UserA.ID, source.ID))
}
