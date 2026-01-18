package handlers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
)

// TestGetVatSetting_CrossOrgIsolation verifies that users can only access
// their own organisation's VAT settings
func TestGetVatSetting_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create VAT setting for User A's organisation
	_, err := env.APIService.CreateVatSetting(models.CreateVatSetting{
		Enabled:                true,
		BillingDate:            "2025-01-15",
		TransactionMonthOffset: 1,
		Interval:               "quarterly",
	}, env.UserA.ID)
	require.NoError(t, err)

	// User A can get their own VAT setting
	vatSettingA, err := env.APIService.GetVatSetting(env.UserA.ID)
	require.NoError(t, err)
	require.NotNil(t, vatSettingA)
	require.True(t, vatSettingA.Enabled)
	require.Equal(t, "quarterly", vatSettingA.Interval)

	// User B gets their own (non-existent) VAT setting - should return nil
	vatSettingB, err := env.APIService.GetVatSetting(env.UserB.ID)
	require.NoError(t, err)
	require.Nil(t, vatSettingB) // No VAT setting for User B

	// Create VAT setting for User B with different settings
	_, err = env.APIService.CreateVatSetting(models.CreateVatSetting{
		Enabled:                false,
		BillingDate:            "2025-02-01",
		TransactionMonthOffset: 0,
		Interval:               "monthly",
	}, env.UserB.ID)
	require.NoError(t, err)

	// Verify User B gets their own settings, not User A's
	vatSettingB, err = env.APIService.GetVatSetting(env.UserB.ID)
	require.NoError(t, err)
	require.NotNil(t, vatSettingB)
	require.False(t, vatSettingB.Enabled)
	require.Equal(t, "monthly", vatSettingB.Interval)

	// Verify User A still gets their own settings
	vatSettingAVerify, err := env.APIService.GetVatSetting(env.UserA.ID)
	require.NoError(t, err)
	require.NotNil(t, vatSettingAVerify)
	require.True(t, vatSettingAVerify.Enabled)
	require.Equal(t, "quarterly", vatSettingAVerify.Interval)
}

// TestUpdateVatSetting_CrossOrgIsolation verifies that updating VAT settings
// only affects the user's own organisation
func TestUpdateVatSetting_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create VAT settings for both users
	_, err := env.APIService.CreateVatSetting(models.CreateVatSetting{
		Enabled:                true,
		BillingDate:            "2025-01-15",
		TransactionMonthOffset: 1,
		Interval:               "quarterly",
	}, env.UserA.ID)
	require.NoError(t, err)

	_, err = env.APIService.CreateVatSetting(models.CreateVatSetting{
		Enabled:                false,
		BillingDate:            "2025-02-01",
		TransactionMonthOffset: 0,
		Interval:               "monthly",
	}, env.UserB.ID)
	require.NoError(t, err)

	// User B updates their settings
	enabled := true
	_, err = env.APIService.UpdateVatSetting(models.UpdateVatSetting{
		Enabled: &enabled,
	}, env.UserB.ID)
	require.NoError(t, err)

	// Verify User A's settings were NOT affected
	vatSettingA, err := env.APIService.GetVatSetting(env.UserA.ID)
	require.NoError(t, err)
	require.True(t, vatSettingA.Enabled)
	require.Equal(t, "quarterly", vatSettingA.Interval)

	// Verify User B's settings were updated
	vatSettingB, err := env.APIService.GetVatSetting(env.UserB.ID)
	require.NoError(t, err)
	require.True(t, vatSettingB.Enabled) // Changed from false to true
	require.Equal(t, "monthly", vatSettingB.Interval) // Unchanged
}

// TestDeleteVatSetting_CrossOrgIsolation verifies that deleting VAT settings
// only affects the user's own organisation
func TestDeleteVatSetting_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create VAT settings for both users
	_, err := env.APIService.CreateVatSetting(models.CreateVatSetting{
		Enabled:                true,
		BillingDate:            "2025-01-15",
		TransactionMonthOffset: 1,
		Interval:               "quarterly",
	}, env.UserA.ID)
	require.NoError(t, err)

	_, err = env.APIService.CreateVatSetting(models.CreateVatSetting{
		Enabled:                false,
		BillingDate:            "2025-02-01",
		TransactionMonthOffset: 0,
		Interval:               "monthly",
	}, env.UserB.ID)
	require.NoError(t, err)

	// User B deletes their settings
	err = env.APIService.DeleteVatSetting(env.UserB.ID)
	require.NoError(t, err)

	// Verify User A's settings still exist
	vatSettingA, err := env.APIService.GetVatSetting(env.UserA.ID)
	require.NoError(t, err)
	require.NotNil(t, vatSettingA)
	require.True(t, vatSettingA.Enabled)

	// Verify User B's settings are deleted
	vatSettingB, err := env.APIService.GetVatSetting(env.UserB.ID)
	require.NoError(t, err)
	require.Nil(t, vatSettingB)
}
