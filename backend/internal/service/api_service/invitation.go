package api_service

import (
	"database/sql"
	"errors"
	"fmt"
	"liquiswiss/config"
	"liquiswiss/internal/events"
	"liquiswiss/pkg/auth"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (a *APIService) ListOrganisationInvitations(userID int64, organisationID int64) ([]models.Invitation, error) {
	// Check if user belongs to the organisation
	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// Only owner and admin can view invitations
	if !a.hasInvitingPermission(organisation.Role) {
		err = errors.New("permission denied")
		logger.Logger.Error(err)
		return nil, err
	}

	invitations, err := a.dbService.ListInvitations(organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	return invitations, nil
}

func (a *APIService) CreateOrganisationInvitation(payload models.CreateInvitation, userID int64, organisationID int64) (*models.Invitation, error) {
	// Check if user belongs to the organisation
	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// Only owner and admin can invite
	if !a.hasInvitingPermission(organisation.Role) {
		err = errors.New("permission denied")
		logger.Logger.Error(err)
		return nil, err
	}

	// Check if the email is already a member
	existingUserID, err := a.dbService.GetUserIDByEmail(payload.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Logger.Error(err)
		return nil, err
	}
	if existingUserID > 0 {
		inOrg, err := a.dbService.CheckUserInOrganisation(existingUserID, organisationID)
		if err != nil {
			logger.Logger.Error(err)
			return nil, err
		}
		if inOrg {
			err = errors.New("user is already a member of this organisation")
			logger.Logger.Error(err)
			return nil, err
		}
	}

	// Generate token
	token := uuid.New().String()
	expiresAt := time.Now().Add(config.GetConfig().InvitationValidity)

	// Create invitation
	invitationID, err := a.dbService.CreateInvitation(organisationID, payload.Email, payload.Role, token, userID, expiresAt)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// Get inviter name and org name for email
	inviter, err := a.dbService.GetProfile(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	inviterName := inviter.Name
	if inviterName == "" {
		inviterName = inviter.Email
	}

	// Send invitation email
	err = a.emailAdapter.SendInvitationMail(payload.Email, token, organisation.Name, inviterName)
	if err != nil {
		logger.Logger.Error(err)
		// Delete the invitation if email fails
		_ = a.dbService.DeleteInvitation(organisationID, invitationID)
		return nil, err
	}

	// Get and return the created invitation
	invitation, err := a.dbService.GetInvitationByID(organisationID, invitationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	a.notifyOrganisationChange(organisationID, "invitation", events.ActionCreated, invitationID)
	return invitation, nil
}

func (a *APIService) DeleteOrganisationInvitation(userID int64, organisationID int64, invitationID int64) error {
	// Check if user belongs to the organisation
	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	// Only owner and admin can delete invitations
	if !a.hasInvitingPermission(organisation.Role) {
		err = errors.New("permission denied")
		logger.Logger.Error(err)
		return err
	}

	// Delete the invitation
	err = a.dbService.DeleteInvitation(organisationID, invitationID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	a.notifyOrganisationChange(organisationID, "invitation", events.ActionDeleted, invitationID)
	return nil
}

func (a *APIService) ResendOrganisationInvitation(userID int64, organisationID int64, invitationID int64) error {
	// Check if user belongs to the organisation
	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	// Only owner and admin can resend invitations
	if !a.hasInvitingPermission(organisation.Role) {
		err = errors.New("permission denied")
		logger.Logger.Error(err)
		return err
	}

	// Get the invitation
	invitation, err := a.dbService.GetInvitationByID(organisationID, invitationID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	// Anti-spam: block resend within configured window
	delay := config.GetConfig().InvitationResendDelay
	if elapsed := time.Since(invitation.LastSentAt); elapsed < delay {
		remaining := delay - elapsed
		minutes := int(remaining.Minutes())
		if minutes < 1 {
			minutes = 1
		}
		return fmt.Errorf("Bitte warten Sie noch %d Minute(n) bevor Sie diese Einladung erneut senden", minutes)
	}

	// Get inviter name for email
	inviter, err := a.dbService.GetProfile(userID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	inviterName := inviter.Name
	if inviterName == "" {
		inviterName = inviter.Email
	}

	// Resend email
	err = a.emailAdapter.SendInvitationMail(invitation.Email, invitation.Token, organisation.Name, inviterName)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	if err := a.dbService.UpdateInvitationLastSentAt(organisationID, invitationID); err != nil {
		logger.Logger.Error(err)
	}

	a.notifyOrganisationChange(organisationID, "invitation", events.ActionUpdated, invitationID)
	return nil
}

func (a *APIService) CheckInvitation(token string) (*models.CheckInvitationResponse, error) {
	invitation, err := a.dbService.GetInvitationByToken(token)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// Check if expired
	if time.Now().After(invitation.ExpiresAt) {
		err = errors.New("invitation has expired")
		logger.Logger.Error(err)
		return nil, err
	}

	// Get organisation name
	orgName, err := a.dbService.GetOrganisationName(invitation.OrganisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// Check if user already exists
	existingUserID, err := a.dbService.GetUserIDByEmail(invitation.Email)
	existingUser := err == nil && existingUserID > 0

	return &models.CheckInvitationResponse{
		Email:            invitation.Email,
		OrganisationName: orgName,
		InvitedByName:    invitation.InvitedByName,
		ExistingUser:     existingUser,
	}, nil
}

func (a *APIService) DeclineMyInvitation(userID int64, invitationID int64) error {
	profile, err := a.dbService.GetProfile(userID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	invitations, err := a.dbService.ListPendingInvitationsByEmail(profile.Email)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	var target *models.UserPendingInvitation
	for i := range invitations {
		if invitations[i].ID == invitationID {
			target = &invitations[i]
			break
		}
	}
	if target == nil {
		return errors.New("invitation not found")
	}

	if err := a.dbService.DeleteInvitation(target.OrganisationID, target.ID); err != nil {
		logger.Logger.Error(err)
		return err
	}
	a.notifyOrganisationChange(target.OrganisationID, "invitation", events.ActionDeleted, target.ID)
	return nil
}

func (a *APIService) ListMyPendingInvitations(userID int64) ([]models.UserPendingInvitation, error) {
	profile, err := a.dbService.GetProfile(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	invitations, err := a.dbService.ListPendingInvitationsByEmail(profile.Email)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	return invitations, nil
}

func (a *APIService) AcceptInvitation(payload models.AcceptInvitation, deviceName string, authenticatedUserID int64) (*models.User, *string, *time.Time, *string, *time.Time, error) {
	// Get invitation by token
	invitation, err := a.dbService.GetInvitationByToken(payload.Token)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	// Check if expired
	if time.Now().After(invitation.ExpiresAt) {
		err = errors.New("invitation has expired")
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	var userID int64

	// Check if user already exists
	existingUserID, err := a.dbService.GetUserIDByEmail(invitation.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	if existingUserID > 0 {
		// If the caller is already authenticated as the invited user, skip the
		// password challenge since we already trust the session.
		if authenticatedUserID > 0 && authenticatedUserID == existingUserID {
			userID = existingUserID
		} else {
			// Otherwise require password to prevent anyone with the invitation
			// token from hijacking the account.
			if payload.Password == nil || *payload.Password == "" {
				return nil, nil, nil, nil, nil, errors.New("invalid credentials")
			}

			loginUser, err := a.dbService.GetUserPasswordByEMail(invitation.Email)
			if err != nil {
				logger.Logger.Error(err)
				return nil, nil, nil, nil, nil, errors.New("invalid credentials")
			}
			if err := bcrypt.CompareHashAndPassword([]byte(loginUser.Password), []byte(*payload.Password)); err != nil {
				return nil, nil, nil, nil, nil, errors.New("invalid credentials")
			}
			userID = existingUserID
		}
	} else {
		// New user - password is required
		if payload.Password == nil || *payload.Password == "" {
			err = errors.New("password is required for new users")
			logger.Logger.Error(err)
			return nil, nil, nil, nil, nil, err
		}

		encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(*payload.Password), 12)
		if err != nil {
			logger.Logger.Error(err)
			return nil, nil, nil, nil, nil, err
		}

		userID, err = a.dbService.CreateUser(invitation.Email, string(encryptedPassword))
		if err != nil {
			logger.Logger.Error(err)
			return nil, nil, nil, nil, nil, err
		}

		// Create user settings
		_, err = a.dbService.CreateUserSetting(userID)
		if err != nil {
			logger.Logger.Error(err)
			return nil, nil, nil, nil, nil, err
		}

		// Every new user gets a personal default organisation, mirroring FinishRegistration.
		// The invited organisation is assigned afterwards and set as the current one.
		defaultOrgID, err := a.dbService.CreateOrganisation("Meine Organisation")
		if err != nil {
			logger.Logger.Error(err)
			return nil, nil, nil, nil, nil, err
		}
		err = a.dbService.AssignUserToOrganisation(userID, defaultOrgID, "owner", true)
		if err != nil {
			logger.Logger.Error(err)
			return nil, nil, nil, nil, nil, err
		}
	}

	// Assign user to organisation
	err = a.dbService.AssignUserToOrganisation(userID, invitation.OrganisationID, invitation.Role, false)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	// Create default permissions based on role
	canView, canEdit, canDelete := getPermissionsForRole(invitation.Role)
	err = a.dbService.UpsertMemberPermission(userID, invitation.OrganisationID, canView, canEdit, canDelete)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	// Set the organisation as current for the user
	err = a.dbService.SetUserCurrentOrganisation(userID, invitation.OrganisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	// Delete the invitation
	err = a.dbService.DeleteInvitationByToken(payload.Token)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	// Accepted: invitation gone, new member joined
	a.notifyOrganisationChange(invitation.OrganisationID, "invitation", events.ActionDeleted, invitation.ID)
	a.notifyOrganisationChange(invitation.OrganisationID, "member", events.ActionCreated, userID)

	// Get user profile
	user, err := a.dbService.GetProfile(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	// Generate tokens
	accessToken, accessExpirationTime, _, err := auth.GenerateAccessToken(*user)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	refreshToken, tokenId, refreshExpirationTime, err := auth.GenerateRefreshToken(*user)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	// Store the refresh token in the database
	err = a.dbService.StoreRefreshTokenID(userID, tokenId, refreshExpirationTime, deviceName)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	return user, &accessToken, &accessExpirationTime, &refreshToken, &refreshExpirationTime, nil
}

func (a *APIService) hasInvitingPermission(role string) bool {
	return role == "owner" || role == "admin"
}

func getPermissionsForRole(role string) (canView, canEdit, canDelete bool) {
	switch role {
	case "admin":
		return true, true, true
	case "editor":
		return true, true, false
	case "read-only":
		return true, false, false
	default:
		return true, false, false
	}
}
