package handlers_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
)

// TestGetUserSetting_AutoCreation verifies that user settings are auto-created
// with defaults when requested for the first time
func TestGetUserSetting_AutoCreation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// User A gets their settings (should auto-create)
	settingA, err := env.APIService.GetUserSetting(env.UserA.ID)
	require.NoError(t, err)
	require.NotNil(t, settingA)
	require.Equal(t, env.UserA.ID, settingA.UserID)
	require.Equal(t, "settings/profile", settingA.SettingsTab)
	require.False(t, settingA.SkipOrganisationSwitchQuestion)
}

// TestGetUserSetting_Isolation verifies that user settings are isolated per user
func TestGetUserSetting_Isolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Get settings for both users (auto-creates with defaults)
	settingA, err := env.APIService.GetUserSetting(env.UserA.ID)
	require.NoError(t, err)
	require.NotNil(t, settingA)

	settingB, err := env.APIService.GetUserSetting(env.UserB.ID)
	require.NoError(t, err)
	require.NotNil(t, settingB)

	// Verify they are different records
	require.NotEqual(t, settingA.ID, settingB.ID)
	require.Equal(t, env.UserA.ID, settingA.UserID)
	require.Equal(t, env.UserB.ID, settingB.UserID)
}

// TestUpdateUserSetting_Isolation verifies that updating user settings only
// affects the correct user
func TestUpdateUserSetting_Isolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Get settings for both users (auto-creates)
	_, err := env.APIService.GetUserSetting(env.UserA.ID)
	require.NoError(t, err)
	_, err = env.APIService.GetUserSetting(env.UserB.ID)
	require.NoError(t, err)

	// User A updates their settings
	newTab := "settings/automation"
	skipQuestion := true
	_, err = env.APIService.UpdateUserSetting(models.UpdateUserSetting{
		SettingsTab:                    &newTab,
		SkipOrganisationSwitchQuestion: &skipQuestion,
	}, env.UserA.ID)
	require.NoError(t, err)

	// Verify User A's settings were updated
	settingA, err := env.APIService.GetUserSetting(env.UserA.ID)
	require.NoError(t, err)
	require.Equal(t, "settings/automation", settingA.SettingsTab)
	require.True(t, settingA.SkipOrganisationSwitchQuestion)

	// Verify User B's settings were NOT affected
	settingB, err := env.APIService.GetUserSetting(env.UserB.ID)
	require.NoError(t, err)
	require.Equal(t, "settings/profile", settingB.SettingsTab)
	require.False(t, settingB.SkipOrganisationSwitchQuestion)
}

// TestGetUserOrganisationSetting_AutoCreation verifies that organisation settings
// are auto-created with defaults when requested for the first time
func TestGetUserOrganisationSetting_AutoCreation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// User A gets their organisation settings (should auto-create)
	setting, err := env.APIService.GetUserOrganisationSetting(env.UserA.ID)
	require.NoError(t, err)
	require.NotNil(t, setting)
	require.Equal(t, env.UserA.ID, setting.UserID)

	// Verify defaults
	require.Equal(t, 13, setting.ForecastMonths)
	require.Equal(t, 100, setting.ForecastPerformance)
	require.False(t, setting.ForecastRevenueDetails)
	require.False(t, setting.ForecastExpenseDetails)
	require.Equal(t, "grid", setting.EmployeeDisplay)
	require.Equal(t, "name", setting.EmployeeSortBy)
	require.Equal(t, "ASC", setting.EmployeeSortOrder)
	require.True(t, setting.EmployeeHideTerminated)
	require.Equal(t, "grid", setting.TransactionDisplay)
	require.True(t, setting.TransactionHideDisabled)
	require.Equal(t, "grid", setting.BankAccountDisplay)
}

// TestGetUserOrganisationSetting_CrossOrgIsolation verifies that organisation
// settings are isolated per user and organisation
func TestGetUserOrganisationSetting_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Get organisation settings for both users
	settingA, err := env.APIService.GetUserOrganisationSetting(env.UserA.ID)
	require.NoError(t, err)
	require.NotNil(t, settingA)

	settingB, err := env.APIService.GetUserOrganisationSetting(env.UserB.ID)
	require.NoError(t, err)
	require.NotNil(t, settingB)

	// Verify they are different records with different organisations
	require.NotEqual(t, settingA.ID, settingB.ID)
	require.NotEqual(t, settingA.OrganisationID, settingB.OrganisationID)
}

// TestUpdateUserOrganisationSetting_CrossOrgIsolation verifies that updating
// organisation settings only affects the correct user's organisation
func TestUpdateUserOrganisationSetting_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Get settings for both users (auto-creates)
	_, err := env.APIService.GetUserOrganisationSetting(env.UserA.ID)
	require.NoError(t, err)
	_, err = env.APIService.GetUserOrganisationSetting(env.UserB.ID)
	require.NoError(t, err)

	// User A updates their organisation settings
	newMonths := 24
	newPerformance := 80
	newDisplay := "list"
	childDetails := json.RawMessage(`["child1","child2"]`)
	_, err = env.APIService.UpdateUserOrganisationSetting(models.UpdateUserOrganisationSetting{
		ForecastMonths:       &newMonths,
		ForecastPerformance:  &newPerformance,
		EmployeeDisplay:      &newDisplay,
		ForecastChildDetails: &childDetails,
	}, env.UserA.ID)
	require.NoError(t, err)

	// Verify User A's settings were updated
	settingA, err := env.APIService.GetUserOrganisationSetting(env.UserA.ID)
	require.NoError(t, err)
	require.Equal(t, 24, settingA.ForecastMonths)
	require.Equal(t, 80, settingA.ForecastPerformance)
	require.Equal(t, "list", settingA.EmployeeDisplay)
	require.JSONEq(t, `["child1","child2"]`, string(settingA.ForecastChildDetails))

	// Verify User B's settings were NOT affected
	settingB, err := env.APIService.GetUserOrganisationSetting(env.UserB.ID)
	require.NoError(t, err)
	require.Equal(t, 13, settingB.ForecastMonths)
	require.Equal(t, 100, settingB.ForecastPerformance)
	require.Equal(t, "grid", settingB.EmployeeDisplay)
}

// TestUpdateUserOrganisationSetting_AllFields verifies that all fields can be
// updated correctly
func TestUpdateUserOrganisationSetting_AllFields(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Get settings (auto-creates)
	_, err := env.APIService.GetUserOrganisationSetting(env.UserA.ID)
	require.NoError(t, err)

	// Update all fields
	months := 36
	performance := 150
	revenueDetails := true
	expenseDetails := true
	childDetails := json.RawMessage(`["detail1"]`)
	empDisplay := "list"
	empSortBy := "email"
	empSortOrder := "DESC"
	empHideTerminated := false
	txDisplay := "list"
	txSortBy := "amount"
	txSortOrder := "DESC"
	txHideDisabled := false
	baDisplay := "list"
	baSortBy := "balance"
	baSortOrder := "DESC"

	_, err = env.APIService.UpdateUserOrganisationSetting(models.UpdateUserOrganisationSetting{
		ForecastMonths:          &months,
		ForecastPerformance:     &performance,
		ForecastRevenueDetails:  &revenueDetails,
		ForecastExpenseDetails:  &expenseDetails,
		ForecastChildDetails:    &childDetails,
		EmployeeDisplay:         &empDisplay,
		EmployeeSortBy:          &empSortBy,
		EmployeeSortOrder:       &empSortOrder,
		EmployeeHideTerminated:  &empHideTerminated,
		TransactionDisplay:      &txDisplay,
		TransactionSortBy:       &txSortBy,
		TransactionSortOrder:    &txSortOrder,
		TransactionHideDisabled: &txHideDisabled,
		BankAccountDisplay:      &baDisplay,
		BankAccountSortBy:       &baSortBy,
		BankAccountSortOrder:    &baSortOrder,
	}, env.UserA.ID)
	require.NoError(t, err)

	// Verify all fields were updated
	setting, err := env.APIService.GetUserOrganisationSetting(env.UserA.ID)
	require.NoError(t, err)
	require.Equal(t, 36, setting.ForecastMonths)
	require.Equal(t, 150, setting.ForecastPerformance)
	require.True(t, setting.ForecastRevenueDetails)
	require.True(t, setting.ForecastExpenseDetails)
	require.JSONEq(t, `["detail1"]`, string(setting.ForecastChildDetails))
	require.Equal(t, "list", setting.EmployeeDisplay)
	require.Equal(t, "email", setting.EmployeeSortBy)
	require.Equal(t, "DESC", setting.EmployeeSortOrder)
	require.False(t, setting.EmployeeHideTerminated)
	require.Equal(t, "list", setting.TransactionDisplay)
	require.Equal(t, "amount", setting.TransactionSortBy)
	require.Equal(t, "DESC", setting.TransactionSortOrder)
	require.False(t, setting.TransactionHideDisabled)
	require.Equal(t, "list", setting.BankAccountDisplay)
	require.Equal(t, "balance", setting.BankAccountSortBy)
	require.Equal(t, "DESC", setting.BankAccountSortOrder)
}
