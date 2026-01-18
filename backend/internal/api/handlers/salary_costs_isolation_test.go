package handlers_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

// TestListSalaryCosts_CrossOrgIsolation verifies that users can only see salary costs
// belonging to salaries in their own organisation
func TestListSalaryCosts_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employees for each organisation
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	employeeB, err := CreateEmployee(env.APIService, env.UserB.ID, "Employee B")
	require.NoError(t, err)

	// Create salaries for each employee
	salaryA, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserA.ID, employeeA.ID)
	require.NoError(t, err)

	salaryB, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              6000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserB.ID, employeeB.ID)
	require.NoError(t, err)

	// Create salary costs for Salary A
	costA1, err := env.APIService.CreateSalaryCost(models.CreateSalaryCost{
		Cycle:            utils.CycleMonthly,
		AmountType:       "fixed",
		Amount:           100_00,
		DistributionType: "employee",
		RelativeOffset:   1,
	}, env.UserA.ID, salaryA.ID)
	require.NoError(t, err)

	costA2, err := env.APIService.CreateSalaryCost(models.CreateSalaryCost{
		Cycle:            utils.CycleMonthly,
		AmountType:       "fixed",
		Amount:           200_00,
		DistributionType: "employer",
		RelativeOffset:   1,
	}, env.UserA.ID, salaryA.ID)
	require.NoError(t, err)

	// Create salary cost for Salary B
	costB1, err := env.APIService.CreateSalaryCost(models.CreateSalaryCost{
		Cycle:            utils.CycleMonthly,
		AmountType:       "fixed",
		Amount:           150_00,
		DistributionType: "employee",
		RelativeOffset:   1,
	}, env.UserB.ID, salaryB.ID)
	require.NoError(t, err)

	// User A can list salary costs for their own salary
	costsA, totalA, err := env.APIService.ListSalaryCosts(env.UserA.ID, salaryA.ID, 1, 100, false)
	require.NoError(t, err)
	require.Equal(t, int64(2), totalA)
	require.Len(t, costsA, 2)

	costIDs := []int64{costsA[0].ID, costsA[1].ID}
	require.Contains(t, costIDs, costA1.ID)
	require.Contains(t, costIDs, costA2.ID)
	require.NotContains(t, costIDs, costB1.ID)

	// User B cannot list salary costs for User A's salary (should return empty)
	costsBAttempt, totalBAttempt, err := env.APIService.ListSalaryCosts(env.UserB.ID, salaryA.ID, 1, 100, false)
	require.NoError(t, err)
	require.Equal(t, int64(0), totalBAttempt)
	require.Len(t, costsBAttempt, 0)

	// User B can list salary costs for their own salary
	costsB, totalB, err := env.APIService.ListSalaryCosts(env.UserB.ID, salaryB.ID, 1, 100, false)
	require.NoError(t, err)
	require.Equal(t, int64(1), totalB)
	require.Len(t, costsB, 1)
	require.Equal(t, costB1.ID, costsB[0].ID)
}

// TestGetSalaryCost_CrossOrgIsolation verifies that a user cannot fetch
// a salary cost belonging to another organisation
func TestGetSalaryCost_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employee and salary for User A
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	salaryA, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserA.ID, employeeA.ID)
	require.NoError(t, err)

	// Create salary cost for Salary A
	costA, err := env.APIService.CreateSalaryCost(models.CreateSalaryCost{
		Cycle:            utils.CycleMonthly,
		AmountType:       "fixed",
		Amount:           100_00,
		DistributionType: "employee",
		RelativeOffset:   1,
	}, env.UserA.ID, salaryA.ID)
	require.NoError(t, err)

	// User A can get their own salary cost
	fetchedCost, err := env.APIService.GetSalaryCost(env.UserA.ID, costA.ID, false)
	require.NoError(t, err)
	require.Equal(t, costA.ID, fetchedCost.ID)
	require.Equal(t, uint64(100_00), fetchedCost.Amount)

	// User B cannot get User A's salary cost (should return sql.ErrNoRows)
	_, err = env.APIService.GetSalaryCost(env.UserB.ID, costA.ID, false)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestCreateSalaryCost_CrossOrgSalary verifies that a user cannot create
// a salary cost for a salary belonging to another organisation
func TestCreateSalaryCost_CrossOrgSalary(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employee and salary for User A
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	salaryA, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserA.ID, employeeA.ID)
	require.NoError(t, err)

	// User B attempts to create a salary cost for User A's salary
	_, err = env.APIService.CreateSalaryCost(models.CreateSalaryCost{
		Cycle:            utils.CycleMonthly,
		AmountType:       "fixed",
		Amount:           100_00,
		DistributionType: "employee",
		RelativeOffset:   1,
	}, env.UserB.ID, salaryA.ID)
	// This should fail - the salary doesn't belong to User B's org
	require.Error(t, err)
}

// TestUpdateSalaryCost_CrossOrgIsolation verifies that a user cannot update
// a salary cost belonging to another organisation
func TestUpdateSalaryCost_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employee and salary for User A
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	salaryA, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserA.ID, employeeA.ID)
	require.NoError(t, err)

	// Create salary cost for Salary A
	costA, err := env.APIService.CreateSalaryCost(models.CreateSalaryCost{
		Cycle:            utils.CycleMonthly,
		AmountType:       "fixed",
		Amount:           100_00,
		DistributionType: "employee",
		RelativeOffset:   1,
	}, env.UserA.ID, salaryA.ID)
	require.NoError(t, err)

	// User A can update their own salary cost
	_, err = env.APIService.UpdateSalaryCost(models.CreateSalaryCost{
		Cycle:            utils.CycleMonthly,
		AmountType:       "fixed",
		Amount:           150_00,
		DistributionType: "employee",
		RelativeOffset:   1,
	}, env.UserA.ID, costA.ID)
	require.NoError(t, err)

	// Verify the update worked
	updatedCost, err := env.APIService.GetSalaryCost(env.UserA.ID, costA.ID, false)
	require.NoError(t, err)
	require.Equal(t, uint64(150_00), updatedCost.Amount)

	// User B attempts to update User A's salary cost
	_, err = env.APIService.UpdateSalaryCost(models.CreateSalaryCost{
		Cycle:            utils.CycleMonthly,
		AmountType:       "fixed",
		Amount:           1_00,
		DistributionType: "employee",
		RelativeOffset:   1,
	}, env.UserB.ID, costA.ID)
	// This should fail - can't update costs from another org
	require.Error(t, err)

	// Verify salary cost was NOT changed by User B
	costAfterAttempt, err := env.APIService.GetSalaryCost(env.UserA.ID, costA.ID, false)
	require.NoError(t, err)
	require.Equal(t, uint64(150_00), costAfterAttempt.Amount)
	require.NotEqual(t, uint64(1_00), costAfterAttempt.Amount)
}

// TestDeleteSalaryCost_CrossOrgIsolation verifies that a user cannot delete
// a salary cost belonging to another organisation
func TestDeleteSalaryCost_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employee and salary for User A
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	salaryA, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserA.ID, employeeA.ID)
	require.NoError(t, err)

	// Create salary cost for Salary A
	costA, err := env.APIService.CreateSalaryCost(models.CreateSalaryCost{
		Cycle:            utils.CycleMonthly,
		AmountType:       "fixed",
		Amount:           100_00,
		DistributionType: "employee",
		RelativeOffset:   1,
	}, env.UserA.ID, salaryA.ID)
	require.NoError(t, err)

	// User B attempts to delete User A's salary cost
	err = env.APIService.DeleteSalaryCost(env.UserB.ID, costA.ID)
	// This should fail - can't delete costs from another org
	require.Error(t, err)

	// Verify salary cost still exists and was NOT deleted
	costAfterDelete, err := env.APIService.GetSalaryCost(env.UserA.ID, costA.ID, false)
	require.NoError(t, err)
	require.NotNil(t, costAfterDelete)
	require.Equal(t, costA.ID, costAfterDelete.ID)

	// User A can successfully delete their own salary cost
	err = env.APIService.DeleteSalaryCost(env.UserA.ID, costA.ID)
	require.NoError(t, err)

	// Verify salary cost is now deleted
	_, err = env.APIService.GetSalaryCost(env.UserA.ID, costA.ID, false)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestCopySalaryCosts_CrossOrgIsolation verifies that a user cannot copy
// salary costs from a salary belonging to another organisation
func TestCopySalaryCosts_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employee and salaries for User A
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	salaryA1, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
		ToDate:              utils.StringAsPointer("2025-06-30"),
	}, env.UserA.ID, employeeA.ID)
	require.NoError(t, err)

	salaryA2, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5500_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-07-01",
	}, env.UserA.ID, employeeA.ID)
	require.NoError(t, err)

	// Create salary cost for Salary A1
	_, err = env.APIService.CreateSalaryCost(models.CreateSalaryCost{
		Cycle:            utils.CycleMonthly,
		AmountType:       "fixed",
		Amount:           100_00,
		DistributionType: "employee",
		RelativeOffset:   1,
	}, env.UserA.ID, salaryA1.ID)
	require.NoError(t, err)

	// Create employee and salary for User B
	employeeB, err := CreateEmployee(env.APIService, env.UserB.ID, "Employee B")
	require.NoError(t, err)

	salaryB, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              6000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserB.ID, employeeB.ID)
	require.NoError(t, err)

	// User B attempts to copy costs from User A's salary to their own
	sourceSalaryA1ID := salaryA1.ID
	err = env.APIService.CopySalaryCosts(models.CopySalaryCosts{
		SourceSalaryID: &sourceSalaryA1ID,
	}, env.UserB.ID, salaryB.ID)
	// This should fail - can't copy costs from another org's salary
	require.Error(t, err)

	// Verify no costs were copied to User B's salary
	costsB, totalB, err := env.APIService.ListSalaryCosts(env.UserB.ID, salaryB.ID, 1, 100, false)
	require.NoError(t, err)
	require.Equal(t, int64(0), totalB)
	require.Len(t, costsB, 0)

	// User A can successfully copy costs between their own salaries
	err = env.APIService.CopySalaryCosts(models.CopySalaryCosts{
		SourceSalaryID: &sourceSalaryA1ID,
	}, env.UserA.ID, salaryA2.ID)
	require.NoError(t, err)

	// Verify costs were copied to Salary A2
	costsA2, totalA2, err := env.APIService.ListSalaryCosts(env.UserA.ID, salaryA2.ID, 1, 100, false)
	require.NoError(t, err)
	require.Equal(t, int64(1), totalA2)
	require.Len(t, costsA2, 1)
}
