package api_service

import (
	"database/sql"
	"errors"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
)

func (a *APIService) ListOrganisationMembers(userID int64, organisationID int64) ([]models.OrganisationMember, error) {
	// Check if user belongs to the organisation
	_, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	members, err := a.dbService.ListMembers(organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// Attach permissions to members
	for i := range members {
		permission, err := a.dbService.GetMemberPermission(members[i].UserID, organisationID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Logger.Error(err)
			return nil, err
		}
		if permission != nil {
			members[i].Permission = permission
		}
	}

	return members, nil
}

func (a *APIService) UpdateOrganisationMember(payload models.UpdateMember, userID int64, organisationID int64, memberUserID int64) error {
	// Check if user belongs to the organisation
	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	// Only owner can update members
	if organisation.Role != "owner" {
		err = errors.New("permission denied")
		logger.Logger.Error(err)
		return err
	}

	// Get the member
	member, err := a.dbService.GetMember(organisationID, memberUserID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	// Cannot update owner role to something else (would need to transfer ownership first)
	if member.Role == "owner" && payload.Role != nil && *payload.Role != "owner" {
		// Check if this is the last owner
		ownerCount, err := a.dbService.CountOwners(organisationID)
		if err != nil {
			logger.Logger.Error(err)
			return err
		}
		if ownerCount <= 1 {
			err = errors.New("cannot demote the last owner")
			logger.Logger.Error(err)
			return err
		}
	}

	// Update role if provided
	if payload.Role != nil {
		err = a.dbService.UpdateMemberRole(organisationID, memberUserID, *payload.Role)
		if err != nil {
			logger.Logger.Error(err)
			return err
		}
	}

	// Update permissions if provided
	if payload.CanView != nil || payload.CanEdit != nil || payload.CanDelete != nil {
		// Get current permissions or use defaults
		currentPerm, err := a.dbService.GetMemberPermission(memberUserID, organisationID)
		canView := true
		canEdit := false
		canDelete := false

		if err == nil && currentPerm != nil {
			canView = currentPerm.CanView
			canEdit = currentPerm.CanEdit
			canDelete = currentPerm.CanDelete
		}

		// Override with provided values
		if payload.CanView != nil {
			canView = *payload.CanView
		}
		if payload.CanEdit != nil {
			canEdit = *payload.CanEdit
		}
		if payload.CanDelete != nil {
			canDelete = *payload.CanDelete
		}

		err = a.dbService.UpsertMemberPermission(memberUserID, organisationID, canView, canEdit, canDelete)
		if err != nil {
			logger.Logger.Error(err)
			return err
		}
	}

	return nil
}

func (a *APIService) RemoveOrganisationMember(userID int64, organisationID int64, memberUserID int64) error {
	// Check if user belongs to the organisation
	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	// Only owner can remove members
	if organisation.Role != "owner" {
		err = errors.New("permission denied")
		logger.Logger.Error(err)
		return err
	}

	// Get the member to check their role
	member, err := a.dbService.GetMember(organisationID, memberUserID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	// Cannot remove the last owner
	if member.Role == "owner" {
		ownerCount, err := a.dbService.CountOwners(organisationID)
		if err != nil {
			logger.Logger.Error(err)
			return err
		}
		if ownerCount <= 1 {
			err = errors.New("cannot remove the last owner")
			logger.Logger.Error(err)
			return err
		}
	}

	// Cannot remove yourself (use leave organisation instead)
	if userID == memberUserID {
		err = errors.New("cannot remove yourself from the organisation")
		logger.Logger.Error(err)
		return err
	}

	// Delete member permissions
	err = a.dbService.DeleteMemberPermissions(memberUserID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	// Remove member from organisation
	err = a.dbService.DeleteMember(organisationID, memberUserID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	return nil
}
