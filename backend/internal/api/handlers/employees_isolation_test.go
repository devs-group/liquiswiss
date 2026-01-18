package handlers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
)

// TestListEmployees_CrossOrgIsolation verifies that users can only see employees
// belonging to their own organisation
func TestListEmployees_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employees for User A's organisation
	employeeA1, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A1")
	require.NoError(t, err)
	employeeA2, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A2")
	require.NoError(t, err)

	// Create employees for User B's organisation
	employeeB1, err := CreateEmployee(env.APIService, env.UserB.ID, "Employee B1")
	require.NoError(t, err)

	// User A should only see their own employees
	employeesA, totalA, err := env.APIService.ListEmployees(env.UserA.ID, 1, 100, "name", "ASC", "")
	require.NoError(t, err)
	require.Equal(t, int64(2), totalA)
	require.Len(t, employeesA, 2)

	employeeIDs := []int64{employeesA[0].ID, employeesA[1].ID}
	require.Contains(t, employeeIDs, employeeA1.ID)
	require.Contains(t, employeeIDs, employeeA2.ID)
	require.NotContains(t, employeeIDs, employeeB1.ID)

	// User B should only see their own employees
	employeesB, totalB, err := env.APIService.ListEmployees(env.UserB.ID, 1, 100, "name", "ASC", "")
	require.NoError(t, err)
	require.Equal(t, int64(1), totalB)
	require.Len(t, employeesB, 1)
	require.Equal(t, employeeB1.ID, employeesB[0].ID)
}

// TestGetEmployee_CrossOrgIsolation verifies that a user cannot fetch
// an employee belonging to another organisation
func TestGetEmployee_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employee for User A's organisation
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A")
	require.NoError(t, err)

	// User A can get their own employee
	fetchedEmployee, err := env.APIService.GetEmployee(env.UserA.ID, employeeA.ID)
	require.NoError(t, err)
	require.Equal(t, employeeA.ID, fetchedEmployee.ID)
	require.Equal(t, "Employee A", fetchedEmployee.Name)

	// User B cannot get User A's employee (should return an error)
	_, err = env.APIService.GetEmployee(env.UserB.ID, employeeA.ID)
	require.Error(t, err)
}

// TestUpdateEmployee_CrossOrgIsolation verifies that a user cannot update
// an employee belonging to another organisation
func TestUpdateEmployee_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employee for User A's organisation
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A Original")
	require.NoError(t, err)

	// User A can update their own employee
	newNameA := "Employee A Updated By A"
	err = env.DBAdapter.UpdateEmployee(models.UpdateEmployee{
		Name: &newNameA,
	}, env.UserA.ID, employeeA.ID)
	require.NoError(t, err)

	// Verify the update worked
	updatedEmployee, err := env.APIService.GetEmployee(env.UserA.ID, employeeA.ID)
	require.NoError(t, err)
	require.Equal(t, "Employee A Updated By A", updatedEmployee.Name)

	// User B attempts to update User A's employee
	maliciousName := "Hacked By B"
	err = env.DBAdapter.UpdateEmployee(models.UpdateEmployee{
		Name: &maliciousName,
	}, env.UserB.ID, employeeA.ID)
	// The update should succeed but affect 0 rows (no error returned by current implementation)
	// We need to verify the employee was NOT actually updated

	// Verify Employee A's name was NOT changed by User B
	employeeAfterAttempt, err := env.APIService.GetEmployee(env.UserA.ID, employeeA.ID)
	require.NoError(t, err)
	require.Equal(t, "Employee A Updated By A", employeeAfterAttempt.Name)
	require.NotEqual(t, "Hacked By B", employeeAfterAttempt.Name)
}

// TestDeleteEmployee_CrossOrgIsolation verifies that a user cannot delete
// an employee belonging to another organisation
func TestDeleteEmployee_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create employee for User A's organisation
	employeeA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee A To Delete")
	require.NoError(t, err)

	// User B attempts to delete User A's employee
	err = env.DBAdapter.DeleteEmployee(env.UserB.ID, employeeA.ID)
	// The delete should not return an error but should affect 0 rows

	// Verify Employee A still exists and was NOT deleted
	employeeAfterDelete, err := env.APIService.GetEmployee(env.UserA.ID, employeeA.ID)
	require.NoError(t, err)
	require.NotNil(t, employeeAfterDelete)
	require.Equal(t, employeeA.ID, employeeAfterDelete.ID)

	// User A can successfully delete their own employee
	err = env.DBAdapter.DeleteEmployee(env.UserA.ID, employeeA.ID)
	require.NoError(t, err)

	// Verify Employee A is now deleted
	_, err = env.APIService.GetEmployee(env.UserA.ID, employeeA.ID)
	require.Error(t, err)
}
