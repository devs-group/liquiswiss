package handlers_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
)

func setupMemberDependencies(t *testing.T) (*sql.DB, api_service.IAPIService, db_adapter.IDatabaseAdapter, *models.User, *models.Organisation) {
	t.Helper()

	conn := SetupTestEnvironment(t)

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	_, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	require.NoError(t, err)

	user, org, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "member.test@test.com", "test", "Member Test Org",
	)
	require.NoError(t, err)

	return conn, apiService, dbAdapter, user, org
}

func TestListMembers_ShowsOwner(t *testing.T) {
	conn, apiService, _, user, org := setupMemberDependencies(t)
	defer conn.Close()

	members, err := apiService.ListOrganisationMembers(user.ID, org.ID)
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, user.ID, members[0].UserID)
	require.Equal(t, "owner", members[0].Role)
}

func TestListMembers_ShowsAllMembers(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupMemberDependencies(t)
	defer conn.Close()

	// Add more members
	member1ID, err := dbAdapter.CreateUser("member1@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(member1ID, org.ID, "admin", false)
	require.NoError(t, err)

	member2ID, err := dbAdapter.CreateUser("member2@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(member2ID, org.ID, "editor", false)
	require.NoError(t, err)

	member3ID, err := dbAdapter.CreateUser("member3@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(member3ID, org.ID, "read-only", false)
	require.NoError(t, err)

	// List members
	members, err := apiService.ListOrganisationMembers(user.ID, org.ID)
	require.NoError(t, err)
	require.Len(t, members, 4)

	// Verify roles
	roleCount := map[string]int{}
	for _, member := range members {
		roleCount[member.Role]++
	}
	require.Equal(t, 1, roleCount["owner"])
	require.Equal(t, 1, roleCount["admin"])
	require.Equal(t, 1, roleCount["editor"])
	require.Equal(t, 1, roleCount["read-only"])
}

func TestListMembers_AnyMemberCanList(t *testing.T) {
	conn, apiService, dbAdapter, _, org := setupMemberDependencies(t)
	defer conn.Close()

	// Create a read-only member
	memberID, err := dbAdapter.CreateUser("readonly@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(memberID, org.ID, "read-only", false)
	require.NoError(t, err)
	err = dbAdapter.SetUserCurrentOrganisation(memberID, org.ID)
	require.NoError(t, err)

	// Read-only member should be able to list members
	members, err := apiService.ListOrganisationMembers(memberID, org.ID)
	require.NoError(t, err)
	require.Len(t, members, 2)
}

func TestUpdateMember_OwnerCanUpdateRole(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupMemberDependencies(t)
	defer conn.Close()

	// Create a member
	memberID, err := dbAdapter.CreateUser("update@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(memberID, org.ID, "editor", false)
	require.NoError(t, err)

	// Owner updates member role
	newRole := "admin"
	err = apiService.UpdateOrganisationMember(models.UpdateMember{
		Role: &newRole,
	}, user.ID, org.ID, memberID)
	require.NoError(t, err)

	// Verify role was updated
	members, err := apiService.ListOrganisationMembers(user.ID, org.ID)
	require.NoError(t, err)

	for _, member := range members {
		if member.UserID == memberID {
			require.Equal(t, "admin", member.Role)
			break
		}
	}
}

func TestUpdateMember_NonOwnerCannotUpdate(t *testing.T) {
	conn, apiService, dbAdapter, _, org := setupMemberDependencies(t)
	defer conn.Close()

	// Create an admin member
	adminID, err := dbAdapter.CreateUser("admin@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(adminID, org.ID, "admin", false)
	require.NoError(t, err)
	err = dbAdapter.SetUserCurrentOrganisation(adminID, org.ID)
	require.NoError(t, err)

	// Create another member
	memberID, err := dbAdapter.CreateUser("member@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(memberID, org.ID, "editor", false)
	require.NoError(t, err)

	// Admin tries to update member role - should fail
	newRole := "read-only"
	err = apiService.UpdateOrganisationMember(models.UpdateMember{
		Role: &newRole,
	}, adminID, org.ID, memberID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "permission denied")
}

func TestUpdateMember_CannotDemoteLastOwner(t *testing.T) {
	conn, apiService, _, user, org := setupMemberDependencies(t)
	defer conn.Close()

	// Try to demote the only owner
	newRole := "admin"
	err := apiService.UpdateOrganisationMember(models.UpdateMember{
		Role: &newRole,
	}, user.ID, org.ID, user.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot demote the last owner")
}

func TestUpdateMember_CanDemoteOwnerIfNotLast(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupMemberDependencies(t)
	defer conn.Close()

	// Create another owner
	owner2ID, err := dbAdapter.CreateUser("owner2@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(owner2ID, org.ID, "owner", false)
	require.NoError(t, err)

	// Now we can demote one owner
	newRole := "admin"
	err = apiService.UpdateOrganisationMember(models.UpdateMember{
		Role: &newRole,
	}, user.ID, org.ID, owner2ID)
	require.NoError(t, err)

	// Verify role was updated
	members, err := apiService.ListOrganisationMembers(user.ID, org.ID)
	require.NoError(t, err)

	for _, member := range members {
		if member.UserID == owner2ID {
			require.Equal(t, "admin", member.Role)
			break
		}
	}
}

func TestUpdateMember_UpdatePermissions(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupMemberDependencies(t)
	defer conn.Close()

	// Create a member
	memberID, err := dbAdapter.CreateUser("perms@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(memberID, org.ID, "read-only", false)
	require.NoError(t, err)

	// Update permissions
	canEdit := true
	canDelete := true
	err = apiService.UpdateOrganisationMember(models.UpdateMember{
		CanEdit:   &canEdit,
		CanDelete: &canDelete,
	}, user.ID, org.ID, memberID)
	require.NoError(t, err)

	// Verify permissions were updated
	members, err := apiService.ListOrganisationMembers(user.ID, org.ID)
	require.NoError(t, err)

	for _, member := range members {
		if member.UserID == memberID {
			require.NotNil(t, member.Permission)
			require.True(t, member.Permission.CanView)
			require.True(t, member.Permission.CanEdit)
			require.True(t, member.Permission.CanDelete)
			break
		}
	}
}

func TestRemoveMember_OwnerCanRemove(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupMemberDependencies(t)
	defer conn.Close()

	// Create a member
	memberID, err := dbAdapter.CreateUser("remove@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(memberID, org.ID, "editor", false)
	require.NoError(t, err)

	// Remove the member
	err = apiService.RemoveOrganisationMember(user.ID, org.ID, memberID)
	require.NoError(t, err)

	// Verify member was removed
	members, err := apiService.ListOrganisationMembers(user.ID, org.ID)
	require.NoError(t, err)
	require.Len(t, members, 1) // Only owner remains
}

func TestRemoveMember_NonOwnerCannotRemove(t *testing.T) {
	conn, apiService, dbAdapter, _, org := setupMemberDependencies(t)
	defer conn.Close()

	// Create an admin
	adminID, err := dbAdapter.CreateUser("admin@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(adminID, org.ID, "admin", false)
	require.NoError(t, err)
	err = dbAdapter.SetUserCurrentOrganisation(adminID, org.ID)
	require.NoError(t, err)

	// Create another member
	memberID, err := dbAdapter.CreateUser("member@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(memberID, org.ID, "editor", false)
	require.NoError(t, err)

	// Admin tries to remove member - should fail
	err = apiService.RemoveOrganisationMember(adminID, org.ID, memberID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "permission denied")
}

func TestRemoveMember_CannotRemoveLastOwner(t *testing.T) {
	conn, apiService, _, user, org := setupMemberDependencies(t)
	defer conn.Close()

	// Try to remove the only owner
	err := apiService.RemoveOrganisationMember(user.ID, org.ID, user.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot remove")
}

func TestRemoveMember_CannotRemoveSelf(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupMemberDependencies(t)
	defer conn.Close()

	// Create another owner so self-removal check is hit before last-owner check
	owner2ID, err := dbAdapter.CreateUser("owner2@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(owner2ID, org.ID, "owner", false)
	require.NoError(t, err)

	// Try to remove self
	err = apiService.RemoveOrganisationMember(user.ID, org.ID, user.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot remove yourself")
}

func TestRemoveMember_CanRemoveOwnerIfNotLast(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupMemberDependencies(t)
	defer conn.Close()

	// Create another owner
	owner2ID, err := dbAdapter.CreateUser("owner2@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(owner2ID, org.ID, "owner", false)
	require.NoError(t, err)

	// Remove the second owner
	err = apiService.RemoveOrganisationMember(user.ID, org.ID, owner2ID)
	require.NoError(t, err)

	// Verify they were removed
	members, err := apiService.ListOrganisationMembers(user.ID, org.ID)
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, user.ID, members[0].UserID)
}

func TestListMembers_IncludesPermissions(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupMemberDependencies(t)
	defer conn.Close()

	// Create a member with permissions
	memberID, err := dbAdapter.CreateUser("withperms@test.com", "password")
	require.NoError(t, err)
	err = dbAdapter.AssignUserToOrganisation(memberID, org.ID, "editor", false)
	require.NoError(t, err)

	// Add permissions
	err = dbAdapter.UpsertMemberPermission(memberID, org.ID, true, true, false)
	require.NoError(t, err)

	// List members
	members, err := apiService.ListOrganisationMembers(user.ID, org.ID)
	require.NoError(t, err)

	for _, member := range members {
		if member.UserID == memberID {
			require.NotNil(t, member.Permission)
			require.True(t, member.Permission.CanView)
			require.True(t, member.Permission.CanEdit)
			require.False(t, member.Permission.CanDelete)
			break
		}
	}
}
