package handlers_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

// TestListSalaries_CrossOrgIsolation verifies that users can only see salaries
// belonging to employees in their own organisation
func TestListSalaries_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employees for each organisation
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	employeeB, err := CreateEmployee(env.APIService, env.UserB.ID, "Employee B")
	require.NoError(t, err)

	// Create salaries for Employee A (User A's org)
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

	// Create salary for Employee B (User B's org)
	salaryB1, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              6000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserB.ID, employeeB.ID)
	require.NoError(t, err)

	// User A can list salaries for their own employee
	salariesA, totalA, err := env.APIService.ListSalaries(env.UserA.ID, employeeA.ID, 1, 100)
	require.NoError(t, err)
	require.Equal(t, int64(2), totalA)
	require.Len(t, salariesA, 2)

	salaryIDs := []int64{salariesA[0].ID, salariesA[1].ID}
	require.Contains(t, salaryIDs, salaryA1.ID)
	require.Contains(t, salaryIDs, salaryA2.ID)
	require.NotContains(t, salaryIDs, salaryB1.ID)

	// User B cannot list salaries for User A's employee (should return empty or error)
	salariesBAttempt, totalBAttempt, err := env.APIService.ListSalaries(env.UserB.ID, employeeA.ID, 1, 100)
	require.NoError(t, err)
	require.Equal(t, int64(0), totalBAttempt)
	require.Len(t, salariesBAttempt, 0)

	// User B can list salaries for their own employee
	salariesB, totalB, err := env.APIService.ListSalaries(env.UserB.ID, employeeB.ID, 1, 100)
	require.NoError(t, err)
	require.Equal(t, int64(1), totalB)
	require.Len(t, salariesB, 1)
	require.Equal(t, salaryB1.ID, salariesB[0].ID)
}

// TestGetSalary_CrossOrgIsolation verifies that a user cannot fetch
// a salary belonging to another organisation's employee
func TestGetSalary_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employee for User A
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	// Create salary for Employee A
	salaryA, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserA.ID, employeeA.ID)
	require.NoError(t, err)

	// User A can get their own employee's salary
	fetchedSalary, err := env.APIService.GetSalary(env.UserA.ID, salaryA.ID)
	require.NoError(t, err)
	require.Equal(t, salaryA.ID, fetchedSalary.ID)
	require.Equal(t, uint64(5000_00), fetchedSalary.Amount)

	// User B cannot get User A's employee's salary (should return sql.ErrNoRows)
	_, err = env.APIService.GetSalary(env.UserB.ID, salaryA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestCreateSalary_CrossOrgEmployee verifies that a user cannot create
// a salary for an employee belonging to another organisation
func TestCreateSalary_CrossOrgEmployee(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employee for User A
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	// User B attempts to create a salary for User A's employee
	_, err = env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              6000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserB.ID, employeeA.ID)
	// This should fail - the employee doesn't belong to User B's org
	require.Error(t, err)
}

// TestUpdateSalary_CrossOrgIsolation verifies that a user cannot update
// a salary belonging to another organisation's employee
func TestUpdateSalary_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employee for User A
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	// Create salary for Employee A
	salaryA, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserA.ID, employeeA.ID)
	require.NoError(t, err)

	// User A can update their own employee's salary
	newAmount := uint64(5500_00)
	_, err = env.APIService.UpdateSalary(models.UpdateSalary{
		Amount: &newAmount,
	}, env.UserA.ID, salaryA.ID)
	require.NoError(t, err)

	// Verify the update worked
	updatedSalary, err := env.APIService.GetSalary(env.UserA.ID, salaryA.ID)
	require.NoError(t, err)
	require.Equal(t, uint64(5500_00), updatedSalary.Amount)

	// User B attempts to update User A's employee's salary (should fail with ErrNoRows)
	maliciousAmount := uint64(1_00)
	_, err = env.APIService.UpdateSalary(models.UpdateSalary{
		Amount: &maliciousAmount,
	}, env.UserB.ID, salaryA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)

	// Verify salary was NOT changed by User B
	salaryAfterAttempt, err := env.APIService.GetSalary(env.UserA.ID, salaryA.ID)
	require.NoError(t, err)
	require.Equal(t, uint64(5500_00), salaryAfterAttempt.Amount)
	require.NotEqual(t, uint64(1_00), salaryAfterAttempt.Amount)
}

// TestDeleteSalary_CrossOrgIsolation verifies that a user cannot delete
// a salary belonging to another organisation's employee
func TestDeleteSalary_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employee for User A
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	// Create salary for Employee A
	salaryA, err := env.APIService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              5000_00,
		Cycle:               utils.CycleMonthly,
		CurrencyID:          *env.Currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
	}, env.UserA.ID, employeeA.ID)
	require.NoError(t, err)

	// User B attempts to delete User A's employee's salary
	err = env.APIService.DeleteSalary(env.UserB.ID, salaryA.ID)
	// This should fail with ErrNoRows since User B can't access this salary
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)

	// Verify salary still exists and was NOT deleted
	salaryAfterDelete, err := env.APIService.GetSalary(env.UserA.ID, salaryA.ID)
	require.NoError(t, err)
	require.NotNil(t, salaryAfterDelete)
	require.Equal(t, salaryA.ID, salaryAfterDelete.ID)

	// User A can successfully delete their own employee's salary
	err = env.APIService.DeleteSalary(env.UserA.ID, salaryA.ID)
	require.NoError(t, err)

	// Verify salary is now deleted
	_, err = env.APIService.GetSalary(env.UserA.ID, salaryA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}
