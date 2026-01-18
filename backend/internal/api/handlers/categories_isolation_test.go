package handlers_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

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
