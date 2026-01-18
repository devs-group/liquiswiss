package handlers_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
)

// TestListVats_ShowsSystemAndOwnOrg verifies that users see system VATs and
// their own organisation's VATs, but not other organisation's VATs
func TestListVats_ShowsSystemAndOwnOrg(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create org-specific VATs for User A
	vatA1, err := env.APIService.CreateVat(models.CreateVat{Value: 770}, env.UserA.ID)
	require.NoError(t, err)

	vatA2, err := env.APIService.CreateVat(models.CreateVat{Value: 250}, env.UserA.ID)
	require.NoError(t, err)

	// Create org-specific VAT for User B
	vatB1, err := env.APIService.CreateVat(models.CreateVat{Value: 380}, env.UserB.ID)
	require.NoError(t, err)

	// User A should see their own VATs (and any system VATs)
	vatsA, err := env.APIService.ListVats(env.UserA.ID)
	require.NoError(t, err)

	vatAIDs := make([]int64, 0)
	for _, v := range vatsA {
		vatAIDs = append(vatAIDs, v.ID)
	}
	require.Contains(t, vatAIDs, vatA1.ID)
	require.Contains(t, vatAIDs, vatA2.ID)
	require.NotContains(t, vatAIDs, vatB1.ID)

	// User B should see their own VATs (and any system VATs)
	vatsB, err := env.APIService.ListVats(env.UserB.ID)
	require.NoError(t, err)

	vatBIDs := make([]int64, 0)
	for _, v := range vatsB {
		vatBIDs = append(vatBIDs, v.ID)
	}
	require.Contains(t, vatBIDs, vatB1.ID)
	require.NotContains(t, vatBIDs, vatA1.ID)
	require.NotContains(t, vatBIDs, vatA2.ID)
}

// TestGetVat_CrossOrgIsolation verifies that a user cannot fetch
// a VAT belonging to another organisation (but can access system VATs)
func TestGetVat_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create org-specific VAT for User A
	vatA, err := env.APIService.CreateVat(models.CreateVat{Value: 770}, env.UserA.ID)
	require.NoError(t, err)

	// User A can get their own VAT
	fetchedVat, err := env.APIService.GetVat(env.UserA.ID, vatA.ID)
	require.NoError(t, err)
	require.Equal(t, vatA.ID, fetchedVat.ID)
	require.Equal(t, int64(770), fetchedVat.Value)

	// User B cannot get User A's VAT (should return sql.ErrNoRows)
	_, err = env.APIService.GetVat(env.UserB.ID, vatA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestUpdateVat_CrossOrgIsolation verifies that a user cannot update
// a VAT belonging to another organisation
func TestUpdateVat_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create org-specific VAT for User A
	vatA, err := env.APIService.CreateVat(models.CreateVat{Value: 770}, env.UserA.ID)
	require.NoError(t, err)

	// User A can update their own VAT
	newValue := int64(800)
	_, err = env.APIService.UpdateVat(models.UpdateVat{Value: &newValue}, env.UserA.ID, vatA.ID)
	require.NoError(t, err)

	// Verify the update worked
	updatedVat, err := env.APIService.GetVat(env.UserA.ID, vatA.ID)
	require.NoError(t, err)
	require.Equal(t, int64(800), updatedVat.Value)

	// User B attempts to update User A's VAT (should fail with ErrNoRows)
	maliciousValue := int64(100)
	_, err = env.APIService.UpdateVat(models.UpdateVat{Value: &maliciousValue}, env.UserB.ID, vatA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)

	// Verify VAT was NOT changed by User B
	vatAfterAttempt, err := env.APIService.GetVat(env.UserA.ID, vatA.ID)
	require.NoError(t, err)
	require.Equal(t, int64(800), vatAfterAttempt.Value)
	require.NotEqual(t, int64(100), vatAfterAttempt.Value)
}

// TestDeleteVat_CrossOrgIsolation verifies that a user cannot delete
// a VAT belonging to another organisation
func TestDeleteVat_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create org-specific VAT for User A
	vatA, err := env.APIService.CreateVat(models.CreateVat{Value: 770}, env.UserA.ID)
	require.NoError(t, err)

	// User B attempts to delete User A's VAT
	err = env.APIService.DeleteVat(env.UserB.ID, vatA.ID)
	// The delete should fail or affect 0 rows

	// Verify VAT still exists and was NOT deleted
	vatAfterDelete, err := env.APIService.GetVat(env.UserA.ID, vatA.ID)
	require.NoError(t, err)
	require.NotNil(t, vatAfterDelete)
	require.Equal(t, vatA.ID, vatAfterDelete.ID)

	// User A can successfully delete their own VAT
	err = env.APIService.DeleteVat(env.UserA.ID, vatA.ID)
	require.NoError(t, err)

	// Verify VAT is now deleted
	_, err = env.APIService.GetVat(env.UserA.ID, vatA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}
