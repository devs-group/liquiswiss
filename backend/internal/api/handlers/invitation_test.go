package handlers_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"liquiswiss/config"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/email_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func setupInvitationDependencies(t *testing.T) (*sql.DB, api_service.IAPIService, db_adapter.IDatabaseAdapter, *models.User, *models.Organisation) {
	t.Helper()

	conn := SetupTestEnvironment(t)

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	// SMTP host empty -> Send methods log warn and return nil; emails not actually delivered
	emailService := email_adapter.NewEmailAdapter(config.Config{})
	apiService := api_service.NewAPIService(dbAdapter, emailService)

	_, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	require.NoError(t, err)

	user, org, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "invitation.test@test.com", "test", "Invitation Test Org",
	)
	require.NoError(t, err)

	return conn, apiService, dbAdapter, user, org
}

func TestCreateInvitation_Success(t *testing.T) {
	conn, apiService, _, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	invitation, err := apiService.CreateOrganisationInvitation(context.Background(), models.CreateInvitation{
		Email: "newuser@test.com",
		Role:  "editor",
	}, user.ID, org.ID)

	// Empty SMTP host -> Send returns nil, invitation succeeds
	require.NoError(t, err)
	require.NotNil(t, invitation)
	require.Equal(t, "newuser@test.com", invitation.Email)
	require.Equal(t, "editor", invitation.Role)
	require.Equal(t, org.ID, invitation.OrganisationID)
}

func TestCreateInvitation_AlreadyMember(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create another user and add them to the organisation
	memberID, err := dbAdapter.CreateUser("member@test.com", "password")
	require.NoError(t, err)

	err = dbAdapter.AssignUserToOrganisation(memberID, org.ID, "editor", false)
	require.NoError(t, err)

	// Try to invite the existing member
	_, err = apiService.CreateOrganisationInvitation(context.Background(), models.CreateInvitation{
		Email: "member@test.com",
		Role:  "editor",
	}, user.ID, org.ID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "already a member")
}

func TestCreateInvitation_NonOwnerCannotInvite(t *testing.T) {
	conn, apiService, dbAdapter, _, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create a member with read-only role
	memberID, err := dbAdapter.CreateUser("readonly@test.com", "password")
	require.NoError(t, err)

	err = dbAdapter.AssignUserToOrganisation(memberID, org.ID, "read-only", false)
	require.NoError(t, err)

	err = dbAdapter.SetUserCurrentOrganisation(memberID, org.ID)
	require.NoError(t, err)

	// Try to invite as read-only member
	_, err = apiService.CreateOrganisationInvitation(context.Background(), models.CreateInvitation{
		Email: "newuser@test.com",
		Role:  "editor",
	}, memberID, org.ID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "permission denied")
}

func TestCreateInvitation_AdminCanInvite(t *testing.T) {
	conn, apiService, dbAdapter, _, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create a member with admin role
	adminID, err := dbAdapter.CreateUser("admin@test.com", "password")
	require.NoError(t, err)

	err = dbAdapter.AssignUserToOrganisation(adminID, org.ID, "admin", false)
	require.NoError(t, err)

	err = dbAdapter.SetUserCurrentOrganisation(adminID, org.ID)
	require.NoError(t, err)

	// Admin should be able to invite; empty SMTP host -> no email send, no error
	_, err = apiService.CreateOrganisationInvitation(context.Background(), models.CreateInvitation{
		Email: "newuser@test.com",
		Role:  "editor",
	}, adminID, org.ID)
	require.NoError(t, err)
}

func TestListInvitations_OwnerCanList(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create invitations directly in DB (bypassing email sending)
	token1 := "test-token-1"
	token2 := "test-token-2"
	expiresAt := time.Now().Add(utils.InvitationValidity)

	_, err := dbAdapter.CreateInvitation(org.ID, "invite1@test.com", "editor", token1, user.ID, expiresAt)
	require.NoError(t, err)

	_, err = dbAdapter.CreateInvitation(org.ID, "invite2@test.com", "read-only", token2, user.ID, expiresAt)
	require.NoError(t, err)

	// List invitations
	invitations, err := apiService.ListOrganisationInvitations(context.Background(), user.ID, org.ID)
	require.NoError(t, err)
	require.Len(t, invitations, 2)
}

func TestListInvitations_AdminCanList(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create an admin user
	adminID, err := dbAdapter.CreateUser("admin@test.com", "password")
	require.NoError(t, err)

	err = dbAdapter.AssignUserToOrganisation(adminID, org.ID, "admin", false)
	require.NoError(t, err)

	err = dbAdapter.SetUserCurrentOrganisation(adminID, org.ID)
	require.NoError(t, err)

	// Create invitation
	token := "test-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	_, err = dbAdapter.CreateInvitation(org.ID, "invite@test.com", "editor", token, user.ID, expiresAt)
	require.NoError(t, err)

	// Admin should be able to list
	invitations, err := apiService.ListOrganisationInvitations(context.Background(), adminID, org.ID)
	require.NoError(t, err)
	require.Len(t, invitations, 1)
}

func TestListInvitations_EditorCannotList(t *testing.T) {
	conn, apiService, dbAdapter, _, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create an editor user
	editorID, err := dbAdapter.CreateUser("editor@test.com", "password")
	require.NoError(t, err)

	err = dbAdapter.AssignUserToOrganisation(editorID, org.ID, "editor", false)
	require.NoError(t, err)

	err = dbAdapter.SetUserCurrentOrganisation(editorID, org.ID)
	require.NoError(t, err)

	// Editor should not be able to list
	_, err = apiService.ListOrganisationInvitations(context.Background(), editorID, org.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "permission denied")
}

func TestDeleteInvitation_Success(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create invitation directly in DB
	token := "test-token-delete"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	invitationID, err := dbAdapter.CreateInvitation(org.ID, "delete@test.com", "editor", token, user.ID, expiresAt)
	require.NoError(t, err)

	// Delete the invitation
	err = apiService.DeleteOrganisationInvitation(context.Background(), user.ID, org.ID, invitationID)
	require.NoError(t, err)

	// Verify it's deleted
	invitations, err := apiService.ListOrganisationInvitations(context.Background(), user.ID, org.ID)
	require.NoError(t, err)
	require.Len(t, invitations, 0)
}

func TestDeleteInvitation_NonOwnerCannotDelete(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create invitation
	token := "test-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	invitationID, err := dbAdapter.CreateInvitation(org.ID, "delete@test.com", "editor", token, user.ID, expiresAt)
	require.NoError(t, err)

	// Create an editor user
	editorID, err := dbAdapter.CreateUser("editor@test.com", "password")
	require.NoError(t, err)

	err = dbAdapter.AssignUserToOrganisation(editorID, org.ID, "editor", false)
	require.NoError(t, err)

	err = dbAdapter.SetUserCurrentOrganisation(editorID, org.ID)
	require.NoError(t, err)

	// Editor should not be able to delete
	err = apiService.DeleteOrganisationInvitation(context.Background(), editorID, org.ID, invitationID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "permission denied")
}

func TestResendOrganisationInvitation_AntiSpamWindow(t *testing.T) {
	// Force a long delay so any immediate resend hits the anti-spam guard.
	t.Setenv("INVITATION_RESEND_DELAY_MINUTES", "60")

	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create invitation. last_sent_at defaults to NOW() on insert.
	token := "resend-spam-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	invitationID, err := dbAdapter.CreateInvitation(org.ID, "spam@test.com", "editor", token, user.ID, expiresAt)
	require.NoError(t, err)

	beforeSent, err := dbAdapter.GetInvitationByID(org.ID, invitationID)
	require.NoError(t, err)

	// Resend immediately should be blocked by the anti-spam window.
	err = apiService.ResendOrganisationInvitation(context.Background(), user.ID, org.ID, invitationID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "warten", "expected German anti-spam message containing 'warten'")

	// last_sent_at must NOT have advanced when resend was blocked.
	afterBlocked, err := dbAdapter.GetInvitationByID(org.ID, invitationID)
	require.NoError(t, err)
	require.Equal(t, beforeSent.LastSentAt.Unix(), afterBlocked.LastSentAt.Unix(),
		"last_sent_at must not advance when resend is rate-limited")
}

func TestResendOrganisationInvitation_UpdatesLastSentAtAfterDelay(t *testing.T) {
	// 0 falls back to default; use a number that would normally block, then manually
	// rewind last_sent_at to simulate the delay having elapsed.
	t.Setenv("INVITATION_RESEND_DELAY_MINUTES", "10")

	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	token := "resend-allowed-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	invitationID, err := dbAdapter.CreateInvitation(org.ID, "allowed@test.com", "editor", token, user.ID, expiresAt)
	require.NoError(t, err)

	// Simulate that the last send was 30 minutes ago — past the 10-minute window.
	pastTime := time.Now().Add(-30 * time.Minute).UTC().Format("2006-01-02 15:04:05")
	_, err = conn.Exec("UPDATE organisation_invitations SET last_sent_at = ? WHERE id = ?", pastTime, invitationID)
	require.NoError(t, err)

	before, err := dbAdapter.GetInvitationByID(org.ID, invitationID)
	require.NoError(t, err)

	// Resend should succeed and advance last_sent_at to "now".
	err = apiService.ResendOrganisationInvitation(context.Background(), user.ID, org.ID, invitationID)
	require.NoError(t, err)

	after, err := dbAdapter.GetInvitationByID(org.ID, invitationID)
	require.NoError(t, err)
	require.WithinDuration(t, time.Now(), after.LastSentAt, 10*time.Second,
		"last_sent_at should advance to ~now after a successful resend")

	// Regression (MariaDB implicit ON UPDATE CURRENT_TIMESTAMP on the first
	// TIMESTAMP column): the resend UPDATE must NOT touch expires_at, else
	// every resent invitation is instantly expired. Fixed by migration 00043
	// (expires_at TIMESTAMP -> DATETIME).
	require.Equal(t, before.ExpiresAt.Unix(), after.ExpiresAt.Unix(),
		"expires_at must not change on resend")
	require.True(t, time.Now().Before(after.ExpiresAt),
		"resent invitation must still be valid")
}

func TestCheckInvitation_ValidToken(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create invitation
	token := "check-token-valid"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	_, err := dbAdapter.CreateInvitation(org.ID, "check@test.com", "editor", token, user.ID, expiresAt)
	require.NoError(t, err)

	// Check the invitation
	response, err := apiService.CheckInvitation(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, "check@test.com", response.Email)
	require.Equal(t, org.Name, response.OrganisationName)
	require.False(t, response.ExistingUser)
}

func TestCheckInvitation_ExpiredToken(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create expired invitation
	token := "expired-token"
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	_, err := dbAdapter.CreateInvitation(org.ID, "expired@test.com", "editor", token, user.ID, expiresAt)
	require.NoError(t, err)

	// Check the invitation - should fail
	_, err = apiService.CheckInvitation(context.Background(), token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

func TestCheckInvitation_InvalidToken(t *testing.T) {
	conn, apiService, _, _, _ := setupInvitationDependencies(t)
	defer conn.Close()

	// Check non-existent token
	_, err := apiService.CheckInvitation(context.Background(), "non-existent-token")
	require.Error(t, err)
}

func TestCheckInvitation_ExistingUser(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create another user (not in this org)
	_, err := dbAdapter.CreateUser("existing@test.com", "password")
	require.NoError(t, err)

	// Create invitation for that email
	token := "existing-user-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	_, err = dbAdapter.CreateInvitation(org.ID, "existing@test.com", "editor", token, user.ID, expiresAt)
	require.NoError(t, err)

	// Check the invitation
	response, err := apiService.CheckInvitation(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, "existing@test.com", response.Email)
	require.True(t, response.ExistingUser)
}

func TestAcceptInvitation_NewUser(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create invitation
	token := "accept-new-user-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	_, err := dbAdapter.CreateInvitation(org.ID, "newuser@test.com", "editor", token, user.ID, expiresAt)
	require.NoError(t, err)

	// Accept the invitation with password
	password := "SecurePassword123"
	acceptedUser, accessToken, _, refreshToken, _, err := apiService.AcceptInvitation(context.Background(), models.AcceptInvitation{
		Token:    token,
		Password: &password,
	}, "Test Device", 0)
	require.NoError(t, err)

	require.NotNil(t, acceptedUser)
	require.Equal(t, "newuser@test.com", acceptedUser.Email)
	require.NotNil(t, accessToken)
	require.NotNil(t, refreshToken)

	// Verify user is now a member of the organisation
	members, err := apiService.ListOrganisationMembers(context.Background(), user.ID, org.ID)
	require.NoError(t, err)

	found := false
	for _, member := range members {
		if member.Email == "newuser@test.com" {
			found = true
			require.Equal(t, "editor", member.Role)
			break
		}
	}
	require.True(t, found, "New user should be a member of the organisation")

	// Verify new user also got a personal default organisation, mirroring FinishRegistration.
	orgs, _, err := apiService.ListOrganisations(context.Background(), acceptedUser.ID, 1, 50)
	require.NoError(t, err)

	var defaultOrg *models.Organisation
	for i, o := range orgs {
		if o.IsDefault {
			defaultOrg = &orgs[i]
			break
		}
	}
	require.NotNil(t, defaultOrg, "Accepted new user should own a personal default organisation")
	require.NotEqual(t, org.ID, defaultOrg.ID, "Default org must not be the invited org")
	require.Equal(t, "Meine Organisation", defaultOrg.Name)
}

func TestAcceptInvitation_ExistingUser(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create an existing user (not in this org) with bcrypt-hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), 12)
	require.NoError(t, err)
	existingUserID, err := dbAdapter.CreateUser("existingaccept@test.com", string(hashedPassword))
	require.NoError(t, err)

	// Create a different org for them
	existingOrg, err := apiService.CreateOrganisation(context.Background(), models.CreateOrganisation{
		Name: "Existing User Org",
	}, existingUserID)
	require.NoError(t, err)

	err = dbAdapter.SetUserCurrentOrganisation(existingUserID, existingOrg.ID)
	require.NoError(t, err)

	// Create invitation for existing user's email
	token := "accept-existing-user-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	_, err = dbAdapter.CreateInvitation(org.ID, "existingaccept@test.com", "admin", token, user.ID, expiresAt)
	require.NoError(t, err)

	// Accept the invitation with correct password
	plainPassword := "password"
	acceptedUser, accessToken, _, refreshToken, _, err := apiService.AcceptInvitation(context.Background(), models.AcceptInvitation{
		Token:    token,
		Password: &plainPassword,
	}, "Test Device", 0)
	require.NoError(t, err)

	require.NotNil(t, acceptedUser)
	require.Equal(t, "existingaccept@test.com", acceptedUser.Email)
	require.NotNil(t, accessToken)
	require.NotNil(t, refreshToken)

	// Verify user is now a member of the new organisation
	members, err := apiService.ListOrganisationMembers(context.Background(), user.ID, org.ID)
	require.NoError(t, err)

	found := false
	for _, member := range members {
		if member.Email == "existingaccept@test.com" {
			found = true
			require.Equal(t, "admin", member.Role)
			break
		}
	}
	require.True(t, found, "Existing user should be a member of the new organisation")
}

func TestAcceptInvitation_ExistingUserWithoutPassword(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), 12)
	require.NoError(t, err)
	_, err = dbAdapter.CreateUser("existing-no-pw@test.com", string(hashedPassword))
	require.NoError(t, err)

	token := "existing-no-password-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	_, err = dbAdapter.CreateInvitation(org.ID, "existing-no-pw@test.com", "admin", token, user.ID, expiresAt)
	require.NoError(t, err)

	_, _, _, _, _, err = apiService.AcceptInvitation(context.Background(), models.AcceptInvitation{
		Token: token,
	}, "Test Device", 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid credentials")
}

func TestAcceptInvitation_ExistingUserWrongPassword(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correct-password"), 12)
	require.NoError(t, err)
	_, err = dbAdapter.CreateUser("existing-wrong-pw@test.com", string(hashedPassword))
	require.NoError(t, err)

	token := "existing-wrong-password-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	_, err = dbAdapter.CreateInvitation(org.ID, "existing-wrong-pw@test.com", "admin", token, user.ID, expiresAt)
	require.NoError(t, err)

	wrongPassword := "wrong-password"
	_, _, _, _, _, err = apiService.AcceptInvitation(context.Background(), models.AcceptInvitation{
		Token:    token,
		Password: &wrongPassword,
	}, "Test Device", 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid credentials")
}

func TestAcceptInvitation_AuthenticatedExistingUserSkipsPassword(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("irrelevant"), 12)
	require.NoError(t, err)
	existingUserID, err := dbAdapter.CreateUser("authuser@test.com", string(hashedPassword))
	require.NoError(t, err)

	token := "authenticated-accept-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	_, err = dbAdapter.CreateInvitation(org.ID, "authuser@test.com", "admin", token, user.ID, expiresAt)
	require.NoError(t, err)

	acceptedUser, accessToken, _, refreshToken, _, err := apiService.AcceptInvitation(context.Background(), models.AcceptInvitation{
		Token: token,
	}, "Test Device", existingUserID)
	require.NoError(t, err)
	require.NotNil(t, acceptedUser)
	require.Equal(t, "authuser@test.com", acceptedUser.Email)
	require.NotNil(t, accessToken)
	require.NotNil(t, refreshToken)
}

func TestAcceptInvitation_AuthenticatedDifferentUserRequiresPassword(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), 12)
	require.NoError(t, err)
	_, err = dbAdapter.CreateUser("invited@test.com", string(hashedPassword))
	require.NoError(t, err)

	token := "auth-mismatch-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	_, err = dbAdapter.CreateInvitation(org.ID, "invited@test.com", "admin", token, user.ID, expiresAt)
	require.NoError(t, err)

	// Different authenticated userID must not bypass password check.
	_, _, _, _, _, err = apiService.AcceptInvitation(context.Background(), models.AcceptInvitation{
		Token: token,
	}, "Test Device", user.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid credentials")
}

func TestAcceptInvitation_NewUserWithoutPassword(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create invitation
	token := "no-password-token"
	expiresAt := time.Now().Add(utils.InvitationValidity)
	_, err := dbAdapter.CreateInvitation(org.ID, "nopw@test.com", "editor", token, user.ID, expiresAt)
	require.NoError(t, err)

	// Try to accept without password - should fail for new user
	_, _, _, _, _, err = apiService.AcceptInvitation(context.Background(), models.AcceptInvitation{
		Token: token,
	}, "Test Device", 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "password is required")
}

func TestAcceptInvitation_ExpiredToken(t *testing.T) {
	conn, apiService, dbAdapter, user, org := setupInvitationDependencies(t)
	defer conn.Close()

	// Create expired invitation
	token := "accept-expired-token"
	expiresAt := time.Now().Add(-1 * time.Hour)
	_, err := dbAdapter.CreateInvitation(org.ID, "expired@test.com", "editor", token, user.ID, expiresAt)
	require.NoError(t, err)

	password := "SecurePassword123"
	_, _, _, _, _, err = apiService.AcceptInvitation(context.Background(), models.AcceptInvitation{
		Token:    token,
		Password: &password,
	}, "Test Device", 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

func TestAcceptInvitation_InvalidToken(t *testing.T) {
	conn, apiService, _, _, _ := setupInvitationDependencies(t)
	defer conn.Close()

	password := "SecurePassword123"
	_, _, _, _, _, err := apiService.AcceptInvitation(context.Background(), models.AcceptInvitation{
		Token:    "invalid-token",
		Password: &password,
	}, "Test Device", 0)
	require.Error(t, err)
}
