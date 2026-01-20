package handlers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

// TestListInvitations_OnlyShowsOwnOrganisation verifies that users can only see
// invitations from their own organisation
func TestListInvitations_OnlyShowsOwnOrganisation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create invitations in both organisations
	tokenA := "org-a-invitation-token"
	tokenB := "org-b-invitation-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)

	_, err := env.DBAdapter.CreateInvitation(env.OrgA.ID, "inviteA@test.com", "editor", tokenA, env.UserA.ID, expiresAt)
	require.NoError(t, err)

	_, err = env.DBAdapter.CreateInvitation(env.OrgB.ID, "inviteB@test.com", "admin", tokenB, env.UserB.ID, expiresAt)
	require.NoError(t, err)

	// User A should only see Org A's invitations
	invitationsA, err := env.APIService.ListOrganisationInvitations(env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)
	require.Len(t, invitationsA, 1)
	require.Equal(t, "inviteA@test.com", invitationsA[0].Email)

	// User B should only see Org B's invitations
	invitationsB, err := env.APIService.ListOrganisationInvitations(env.UserB.ID, env.OrgB.ID)
	require.NoError(t, err)
	require.Len(t, invitationsB, 1)
	require.Equal(t, "inviteB@test.com", invitationsB[0].Email)
}

// TestListInvitations_CannotAccessOtherOrganisation verifies that users cannot
// list invitations from organisations they don't belong to
func TestListInvitations_CannotAccessOtherOrganisation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create invitation in Org A
	tokenA := "org-a-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	_, err := env.DBAdapter.CreateInvitation(env.OrgA.ID, "inviteA@test.com", "editor", tokenA, env.UserA.ID, expiresAt)
	require.NoError(t, err)

	// User B tries to list Org A's invitations - should fail
	_, err = env.APIService.ListOrganisationInvitations(env.UserB.ID, env.OrgA.ID)
	require.Error(t, err)

	// User A tries to list Org B's invitations - should fail
	_, err = env.APIService.ListOrganisationInvitations(env.UserA.ID, env.OrgB.ID)
	require.Error(t, err)
}

// TestCreateInvitation_CannotCreateInOtherOrganisation verifies that users cannot
// create invitations in organisations they don't belong to
func TestCreateInvitation_CannotCreateInOtherOrganisation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// User B tries to create invitation in Org A - should fail
	_, err := env.APIService.CreateOrganisationInvitation(models.CreateInvitation{
		Email: "malicious@test.com",
		Role:  "admin",
	}, env.UserB.ID, env.OrgA.ID)
	require.Error(t, err)

	// Verify no invitation was created
	invitations, err := env.APIService.ListOrganisationInvitations(env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)
	require.Len(t, invitations, 0)
}

// TestDeleteInvitation_CannotDeleteFromOtherOrganisation verifies that users cannot
// delete invitations from organisations they don't belong to
func TestDeleteInvitation_CannotDeleteFromOtherOrganisation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create invitation in Org A
	tokenA := "delete-test-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	invitationID, err := env.DBAdapter.CreateInvitation(env.OrgA.ID, "delete@test.com", "editor", tokenA, env.UserA.ID, expiresAt)
	require.NoError(t, err)

	// User B tries to delete Org A's invitation - should fail
	err = env.APIService.DeleteOrganisationInvitation(env.UserB.ID, env.OrgA.ID, invitationID)
	require.Error(t, err)

	// Verify invitation still exists
	invitations, err := env.APIService.ListOrganisationInvitations(env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)
	require.Len(t, invitations, 1)
}

// TestResendInvitation_CannotResendFromOtherOrganisation verifies that users cannot
// resend invitations from organisations they don't belong to
func TestResendInvitation_CannotResendFromOtherOrganisation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create invitation in Org A
	tokenA := "resend-test-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	invitationID, err := env.DBAdapter.CreateInvitation(env.OrgA.ID, "resend@test.com", "editor", tokenA, env.UserA.ID, expiresAt)
	require.NoError(t, err)

	// User B tries to resend Org A's invitation - should fail
	err = env.APIService.ResendOrganisationInvitation(env.UserB.ID, env.OrgA.ID, invitationID)
	require.Error(t, err)
}

// TestAcceptInvitation_JoinsCorrectOrganisation verifies that accepting an invitation
// adds the user to the correct organisation only
func TestAcceptInvitation_JoinsCorrectOrganisation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	// Create invitation in Org A
	token := "join-test-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	_, err := env.DBAdapter.CreateInvitation(env.OrgA.ID, "joiner@test.com", "editor", token, env.UserA.ID, expiresAt)
	require.NoError(t, err)

	// Accept the invitation
	password := "SecurePassword123"
	acceptedUser, _, _, _, _, err := env.APIService.AcceptInvitation(models.AcceptInvitation{
		Token:    token,
		Password: &password,
	}, "Test Device")
	require.NoError(t, err)
	require.NotNil(t, acceptedUser)

	// User should be a member of Org A
	membersA, err := env.APIService.ListOrganisationMembers(env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)

	found := false
	for _, member := range membersA {
		if member.Email == "joiner@test.com" {
			found = true
			require.Equal(t, "editor", member.Role)
			break
		}
	}
	require.True(t, found, "User should be a member of Org A")

	// User should NOT be a member of Org B
	membersB, err := env.APIService.ListOrganisationMembers(env.UserB.ID, env.OrgB.ID)
	require.NoError(t, err)

	for _, member := range membersB {
		require.NotEqual(t, "joiner@test.com", member.Email, "User should not be a member of Org B")
	}
}

// TestInvitation_SameEmailDifferentOrganisations verifies that the same email
// can have pending invitations in different organisations
func TestInvitation_SameEmailDifferentOrganisations(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	email := "multi-org@test.com"
	expiresAt := time.Now().Add(utils.InvitationValidity)

	// Create invitation in Org A
	tokenA := "multi-org-token-a"
	_, err := env.DBAdapter.CreateInvitation(env.OrgA.ID, email, "editor", tokenA, env.UserA.ID, expiresAt)
	require.NoError(t, err)

	// Create invitation in Org B with same email
	tokenB := "multi-org-token-b"
	_, err = env.DBAdapter.CreateInvitation(env.OrgB.ID, email, "admin", tokenB, env.UserB.ID, expiresAt)
	require.NoError(t, err)

	// Both organisations should have their own invitation
	invitationsA, err := env.APIService.ListOrganisationInvitations(env.UserA.ID, env.OrgA.ID)
	require.NoError(t, err)
	require.Len(t, invitationsA, 1)
	require.Equal(t, email, invitationsA[0].Email)
	require.Equal(t, "editor", invitationsA[0].Role)

	invitationsB, err := env.APIService.ListOrganisationInvitations(env.UserB.ID, env.OrgB.ID)
	require.NoError(t, err)
	require.Len(t, invitationsB, 1)
	require.Equal(t, email, invitationsB[0].Email)
	require.Equal(t, "admin", invitationsB[0].Role)
}
