package handlers_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
)

// TestListSalaryCostLabels_CrossOrgIsolation verifies that users can only see
// salary cost labels belonging to their own organisation
func TestListSalaryCostLabels_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create salary cost labels for User A's organisation
	labelA1, err := CreateSalaryCostLabel(env.APIService, env.UserA.ID, "Label A1")
	require.NoError(t, err)

	labelA2, err := CreateSalaryCostLabel(env.APIService, env.UserA.ID, "Label A2")
	require.NoError(t, err)

	// Create salary cost labels for User B's organisation
	labelB1, err := CreateSalaryCostLabel(env.APIService, env.UserB.ID, "Label B1")
	require.NoError(t, err)

	// User A should only see their own labels
	labelsA, totalA, err := env.APIService.ListSalaryCostLabels(env.UserA.ID, 1, 100)
	require.NoError(t, err)
	require.Equal(t, int64(2), totalA)
	require.Len(t, labelsA, 2)

	labelIDs := []int64{labelsA[0].ID, labelsA[1].ID}
	require.Contains(t, labelIDs, labelA1.ID)
	require.Contains(t, labelIDs, labelA2.ID)
	require.NotContains(t, labelIDs, labelB1.ID)

	// User B should only see their own labels
	labelsB, totalB, err := env.APIService.ListSalaryCostLabels(env.UserB.ID, 1, 100)
	require.NoError(t, err)
	require.Equal(t, int64(1), totalB)
	require.Len(t, labelsB, 1)
	require.Equal(t, labelB1.ID, labelsB[0].ID)
}

// TestGetSalaryCostLabel_CrossOrgIsolation verifies that a user cannot fetch
// a salary cost label belonging to another organisation
func TestGetSalaryCostLabel_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create salary cost label for User A
	labelA, err := CreateSalaryCostLabel(env.APIService, env.UserA.ID, "Label A")
	require.NoError(t, err)

	// User A can get their own label
	fetchedLabel, err := env.APIService.GetSalaryCostLabel(env.UserA.ID, labelA.ID)
	require.NoError(t, err)
	require.Equal(t, labelA.ID, fetchedLabel.ID)
	require.Equal(t, "Label A", fetchedLabel.Name)

	// User B cannot get User A's label (should return sql.ErrNoRows)
	_, err = env.APIService.GetSalaryCostLabel(env.UserB.ID, labelA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestUpdateSalaryCostLabel_CrossOrgIsolation verifies that a user cannot update
// a salary cost label belonging to another organisation
func TestUpdateSalaryCostLabel_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create salary cost label for User A
	labelA, err := CreateSalaryCostLabel(env.APIService, env.UserA.ID, "Label A Original")
	require.NoError(t, err)

	// User A can update their own label
	_, err = env.APIService.UpdateSalaryCostLabel(models.CreateSalaryCostLabel{
		Name: "Label A Updated By A",
	}, env.UserA.ID, labelA.ID)
	require.NoError(t, err)

	// Verify the update worked
	updatedLabel, err := env.APIService.GetSalaryCostLabel(env.UserA.ID, labelA.ID)
	require.NoError(t, err)
	require.Equal(t, "Label A Updated By A", updatedLabel.Name)

	// User B attempts to update User A's label (should fail with ErrNoRows)
	_, err = env.APIService.UpdateSalaryCostLabel(models.CreateSalaryCostLabel{
		Name: "Hacked By B",
	}, env.UserB.ID, labelA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)

	// Verify label was NOT changed by User B
	labelAfterAttempt, err := env.APIService.GetSalaryCostLabel(env.UserA.ID, labelA.ID)
	require.NoError(t, err)
	require.Equal(t, "Label A Updated By A", labelAfterAttempt.Name)
	require.NotEqual(t, "Hacked By B", labelAfterAttempt.Name)
}

// TestDeleteSalaryCostLabel_CrossOrgIsolation verifies that a user cannot delete
// a salary cost label belonging to another organisation
func TestDeleteSalaryCostLabel_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create salary cost label for User A
	labelA, err := CreateSalaryCostLabel(env.APIService, env.UserA.ID, "Label A To Delete")
	require.NoError(t, err)

	// User B attempts to delete User A's label
	err = env.APIService.DeleteSalaryCostLabel(env.UserB.ID, labelA.ID)
	// The delete should fail or affect 0 rows

	// Verify label still exists and was NOT deleted
	labelAfterDelete, err := env.APIService.GetSalaryCostLabel(env.UserA.ID, labelA.ID)
	require.NoError(t, err)
	require.NotNil(t, labelAfterDelete)
	require.Equal(t, labelA.ID, labelAfterDelete.ID)

	// User A can successfully delete their own label
	err = env.APIService.DeleteSalaryCostLabel(env.UserA.ID, labelA.ID)
	require.NoError(t, err)

	// Verify label is now deleted
	_, err = env.APIService.GetSalaryCostLabel(env.UserA.ID, labelA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestCreateSalaryCost_WithCrossOrgLabel verifies that a user cannot create
// a salary cost using a label from another organisation
func TestCreateSalaryCost_WithCrossOrgLabel(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create label for User A
	labelA, err := CreateSalaryCostLabel(env.APIService, env.UserA.ID, "Label A")
	require.NoError(t, err)

	// Create employee and salary for User B
	employeeB, err := CreateEmployee(env.APIService, env.UserB.ID, "Employee B")
	require.NoError(t, err)

	salaryB, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               "monthly",
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserB.ID, employeeB.ID)
	require.NoError(t, err)

	// User B attempts to create a salary cost using User A's label
	labelAID := labelA.ID
	_, err = env.APIService.CreateSalaryCost(models.CreateSalaryCost{
		Cycle:            "monthly",
		AmountType:       "fixed",
		Amount:           100_00,
		DistributionType: "employee",
		RelativeOffset:   1,
		LabelID:          &labelAID,
	}, env.UserB.ID, salaryB.ID)
	// This should fail - can't use a label from another org
	require.Error(t, err)
}
