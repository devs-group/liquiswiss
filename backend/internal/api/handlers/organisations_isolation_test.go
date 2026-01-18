package handlers_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
)

// TestListOrganisations_OnlyShowsOwnMemberships verifies that users only see
// organisations they are members of
func TestListOrganisations_OnlyShowsOwnMemberships(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// User A should only see Organisation A
	orgsA, totalA, err := env.APIService.ListOrganisations(env.UserA.ID, 1, 100)
	require.NoError(t, err)
	require.Equal(t, int64(1), totalA)
	require.Len(t, orgsA, 1)
	require.Equal(t, env.OrgA.ID, orgsA[0].ID)
	require.Equal(t, "Organisation A", orgsA[0].Name)

	// User B should only see Organisation B
	orgsB, totalB, err := env.APIService.ListOrganisations(env.UserB.ID, 1, 100)
	require.NoError(t, err)
	require.Equal(t, int64(1), totalB)
	require.Len(t, orgsB, 1)
	require.Equal(t, env.OrgB.ID, orgsB[0].ID)
	require.Equal(t, "Organisation B", orgsB[0].Name)

	// User A's organisation list should NOT contain User B's organisation
	for _, org := range orgsA {
		require.NotEqual(t, env.OrgB.ID, org.ID,
			"User A should not see User B's organisation")
	}

	// User B's organisation list should NOT contain User A's organisation
	for _, org := range orgsB {
		require.NotEqual(t, env.OrgA.ID, org.ID,
			"User B should not see User A's organisation")
	}
}

// TestGetOrganisation_MembershipRequired verifies that a user cannot access
// an organisation they are not a member of
func TestGetOrganisation_MembershipRequired(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// User A can get their own organisation
	orgA, err := env.APIService.GetOrganisation(env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)
	require.Equal(t, env.OrgA.ID, orgA.ID)
	require.Equal(t, "Organisation A", orgA.Name)

	// User A cannot get User B's organisation (should return sql.ErrNoRows)
	_, err = env.APIService.GetOrganisation(env.UserA.ID, env.OrgB.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)

	// User B can get their own organisation
	orgB, err := env.APIService.GetOrganisation(env.UserB.ID, env.OrgB.ID)
	require.NoError(t, err)
	require.Equal(t, env.OrgB.ID, orgB.ID)
	require.Equal(t, "Organisation B", orgB.Name)

	// User B cannot get User A's organisation (should return sql.ErrNoRows)
	_, err = env.APIService.GetOrganisation(env.UserB.ID, env.OrgA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestUpdateOrganisation_MembershipRequired verifies that a user cannot update
// an organisation they are not a member of
func TestUpdateOrganisation_MembershipRequired(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// User A can update their own organisation
	newNameA := "Organisation A Updated"
	_, err := env.APIService.UpdateOrganisation(models.UpdateOrganisation{
		Name: &newNameA,
	}, env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)

	// Verify the update worked
	updatedOrgA, err := env.APIService.GetOrganisation(env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)
	require.Equal(t, "Organisation A Updated", updatedOrgA.Name)

	// User B attempts to update User A's organisation (should fail with ErrNoRows)
	maliciousName := "Hacked By B"
	_, err = env.APIService.UpdateOrganisation(models.UpdateOrganisation{
		Name: &maliciousName,
	}, env.UserB.ID, env.OrgA.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)

	// Verify Organisation A was NOT changed by User B
	orgAfterAttempt, err := env.APIService.GetOrganisation(env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)
	require.Equal(t, "Organisation A Updated", orgAfterAttempt.Name)
	require.NotEqual(t, "Hacked By B", orgAfterAttempt.Name)
}

// TestUserSwitchOrganisation_DataChanges verifies that when a user with multiple
// organisations switches between them, they see data from the correct organisation
func TestUserSwitchOrganisation_DataChanges(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create a new organisation for User A (so they have 2 orgs)
	orgA2, err := env.APIService.CreateOrganisation(models.CreateOrganisation{
		Name: "Organisation A2",
	}, env.UserA.ID)
	require.NoError(t, err)

	// Create employees in each of User A's organisations
	// First, employee in org A (current)
	empOrgA, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee in Org A")
	require.NoError(t, err)

	// Switch to org A2
	err = env.APIService.SetUserCurrentOrganisation(models.UpdateUserCurrentOrganisation{
		OrganisationID: orgA2.ID,
	}, env.UserA.ID)
	require.NoError(t, err)

	// Create employee in org A2
	empOrgA2, err := CreateEmployee(env.APIService, env.UserA.ID, "Employee in Org A2")
	require.NoError(t, err)

	// List employees - should only see Org A2 employee
	employees, total, err := env.APIService.ListEmployees(env.UserA.ID, 1, 100, "name", "ASC", "", false)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, employees, 1)
	require.Equal(t, empOrgA2.ID, employees[0].ID)
	require.Equal(t, "Employee in Org A2", employees[0].Name)

	// The employee from Org A should NOT be visible
	for _, emp := range employees {
		require.NotEqual(t, empOrgA.ID, emp.ID,
			"Should not see employees from other organisation")
	}

	// Switch back to org A
	err = env.APIService.SetUserCurrentOrganisation(models.UpdateUserCurrentOrganisation{
		OrganisationID: env.OrgA.ID,
	}, env.UserA.ID)
	require.NoError(t, err)

	// List employees - should only see Org A employee
	employees, total, err = env.APIService.ListEmployees(env.UserA.ID, 1, 100, "name", "ASC", "", false)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, employees, 1)
	require.Equal(t, empOrgA.ID, employees[0].ID)
	require.Equal(t, "Employee in Org A", employees[0].Name)

	// The employee from Org A2 should NOT be visible
	for _, emp := range employees {
		require.NotEqual(t, empOrgA2.ID, emp.ID,
			"Should not see employees from other organisation")
	}

	// Trying to get the Org A2 employee by ID should fail (404)
	_, err = env.APIService.GetEmployee(env.UserA.ID, empOrgA2.ID)
	require.Error(t, err, "Should not be able to get employee from other organisation")
}
