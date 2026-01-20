package handlers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
)

// TestListMembers_OnlyShowsOwnOrganisation verifies that users can only see
// members from their own organisation
func TestListMembers_OnlyShowsOwnOrganisation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// User A should only see Org A's members (just themselves)
	membersA, err := env.APIService.ListOrganisationMembers(env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)
	require.Len(t, membersA, 1)
	require.Equal(t, env.UserA.ID, membersA[0].UserID)

	// User B should only see Org B's members (just themselves)
	membersB, err := env.APIService.ListOrganisationMembers(env.UserB.ID, env.OrgB.ID)
	require.NoError(t, err)
	require.Len(t, membersB, 1)
	require.Equal(t, env.UserB.ID, membersB[0].UserID)
}

// TestListMembers_CannotAccessOtherOrganisation verifies that users cannot
// list members from organisations they don't belong to
func TestListMembers_CannotAccessOtherOrganisation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// User A tries to list Org B's members - should fail
	_, err := env.APIService.ListOrganisationMembers(env.UserA.ID, env.OrgB.ID)
	require.Error(t, err)

	// User B tries to list Org A's members - should fail
	_, err = env.APIService.ListOrganisationMembers(env.UserB.ID, env.OrgA.ID)
	require.Error(t, err)
}

// TestUpdateMember_CannotUpdateInOtherOrganisation verifies that users cannot
// update members in organisations they don't belong to
func TestUpdateMember_CannotUpdateInOtherOrganisation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// User B tries to update User A's role in Org A - should fail
	newRole := "read-only"
	err := env.APIService.UpdateOrganisationMember(models.UpdateMember{
		Role: &newRole,
	}, env.UserB.ID, env.OrgA.ID, env.UserA.ID)
	require.Error(t, err)

	// Verify User A's role is still owner
	membersA, err := env.APIService.ListOrganisationMembers(env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)
	require.Len(t, membersA, 1)
	require.Equal(t, "owner", membersA[0].Role)
}

// TestRemoveMember_CannotRemoveFromOtherOrganisation verifies that users cannot
// remove members from organisations they don't belong to
func TestRemoveMember_CannotRemoveFromOtherOrganisation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Add a member to Org A
	memberID, err := env.DBAdapter.CreateUser("orgamember@test.com", "password")
	require.NoError(t, err)
	err = env.DBAdapter.AssignUserToOrganisation(memberID, env.OrgA.ID, "editor", false)
	require.NoError(t, err)

	// User B tries to remove that member from Org A - should fail
	err = env.APIService.RemoveOrganisationMember(env.UserB.ID, env.OrgA.ID, memberID)
	require.Error(t, err)

	// Verify member still exists
	membersA, err := env.APIService.ListOrganisationMembers(env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)
	require.Len(t, membersA, 2)
}

// TestMember_CrossOrgMembershipIndependence verifies that a user who is a member
// of multiple organisations cannot use their permissions from one org in another
func TestMember_CrossOrgMembershipIndependence(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create a user who is owner of Org A and member of Org B
	sharedUserID, err := env.DBAdapter.CreateUser("shared@test.com", "password")
	require.NoError(t, err)

	// Make them an owner of a new org
	newOrg, err := env.APIService.CreateOrganisation(models.CreateOrganisation{
		Name: "Shared User Org",
	}, sharedUserID)
	require.NoError(t, err)

	// Also add them as a read-only member of Org A
	err = env.DBAdapter.AssignUserToOrganisation(sharedUserID, env.OrgA.ID, "read-only", false)
	require.NoError(t, err)

	// Set their current org to their own org where they're owner
	err = env.DBAdapter.SetUserCurrentOrganisation(sharedUserID, newOrg.ID)
	require.NoError(t, err)

	// They should NOT be able to update members in Org A (where they're only read-only)
	newRole := "admin"
	err = env.APIService.UpdateOrganisationMember(models.UpdateMember{
		Role: &newRole,
	}, sharedUserID, env.OrgA.ID, env.UserA.ID)
	require.Error(t, err)

	// But they CAN list members in Org A (any member can list)
	members, err := env.APIService.ListOrganisationMembers(sharedUserID, env.OrgA.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(members), 2) // Owner + shared user
}

// TestMember_UserInMultipleOrgs_CorrectMembershipPerOrg verifies that when a user
// is in multiple organisations, their membership is correctly tracked per org
func TestMember_UserInMultipleOrgs_CorrectMembershipPerOrg(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create a user who will be in both orgs with different roles
	dualUserID, err := env.DBAdapter.CreateUser("dual@test.com", "password")
	require.NoError(t, err)

	// Add to Org A as admin
	err = env.DBAdapter.AssignUserToOrganisation(dualUserID, env.OrgA.ID, "admin", false)
	require.NoError(t, err)

	// Add to Org B as read-only
	err = env.DBAdapter.AssignUserToOrganisation(dualUserID, env.OrgB.ID, "read-only", false)
	require.NoError(t, err)

	// Check their role in Org A
	membersA, err := env.APIService.ListOrganisationMembers(env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)
	for _, member := range membersA {
		if member.UserID == dualUserID {
			require.Equal(t, "admin", member.Role, "Should be admin in Org A")
			break
		}
	}

	// Check their role in Org B
	membersB, err := env.APIService.ListOrganisationMembers(env.UserB.ID, env.OrgB.ID)
	require.NoError(t, err)
	for _, member := range membersB {
		if member.UserID == dualUserID {
			require.Equal(t, "read-only", member.Role, "Should be read-only in Org B")
			break
		}
	}
}

// TestMember_PermissionsAreOrgSpecific verifies that member permissions are
// tracked separately for each organisation
func TestMember_PermissionsAreOrgSpecific(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create a user who will be in both orgs
	sharedUserID, err := env.DBAdapter.CreateUser("permtest@test.com", "password")
	require.NoError(t, err)

	// Add to Org A and B
	err = env.DBAdapter.AssignUserToOrganisation(sharedUserID, env.OrgA.ID, "editor", false)
	require.NoError(t, err)
	err = env.DBAdapter.AssignUserToOrganisation(sharedUserID, env.OrgB.ID, "editor", false)
	require.NoError(t, err)

	// Set different permissions in each org
	err = env.DBAdapter.UpsertMemberPermission(sharedUserID, env.OrgA.ID, true, true, true)
	require.NoError(t, err)
	err = env.DBAdapter.UpsertMemberPermission(sharedUserID, env.OrgB.ID, true, false, false)
	require.NoError(t, err)

	// Check permissions in Org A
	membersA, err := env.APIService.ListOrganisationMembers(env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)
	for _, member := range membersA {
		if member.UserID == sharedUserID {
			require.NotNil(t, member.Permission)
			require.True(t, member.Permission.CanView)
			require.True(t, member.Permission.CanEdit, "Should have edit permission in Org A")
			require.True(t, member.Permission.CanDelete, "Should have delete permission in Org A")
			break
		}
	}

	// Check permissions in Org B
	membersB, err := env.APIService.ListOrganisationMembers(env.UserB.ID, env.OrgB.ID)
	require.NoError(t, err)
	for _, member := range membersB {
		if member.UserID == sharedUserID {
			require.NotNil(t, member.Permission)
			require.True(t, member.Permission.CanView)
			require.False(t, member.Permission.CanEdit, "Should NOT have edit permission in Org B")
			require.False(t, member.Permission.CanDelete, "Should NOT have delete permission in Org B")
			break
		}
	}
}
