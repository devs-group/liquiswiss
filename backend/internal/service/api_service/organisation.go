package api_service

import (
	"context"
	"errors"
	"liquiswiss/internal/events"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"slices"
)

func (a *APIService) ListOrganisations(ctx context.Context, userID int64, page int64, limit int64) ([]models.Organisation, int64, error) {
	organisations, totalCount, err := a.dbService.ListOrganisations(userID, page, limit)
	if err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(organisations, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	return organisations, totalCount, nil
}

func (a *APIService) GetOrganisation(ctx context.Context, userID int64, organisationID int64) (*models.Organisation, error) {
	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(organisation); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return organisation, nil
}

func (a *APIService) CreateOrganisation(ctx context.Context, payload models.CreateOrganisation, userID int64) (*models.Organisation, error) {
	organisationID, err := a.dbService.CreateOrganisation(payload.Name)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.AssignUserToOrganisation(userID, organisationID, "owner", false)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(organisation); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return organisation, nil
}

func (a *APIService) UpdateOrganisation(ctx context.Context, payload models.UpdateOrganisation, userID int64, organisationID int64) (*models.Organisation, error) {
	existingOrganisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Check if user is allowed to edit
	if !a.hasEditingPermission(existingOrganisation.Role) {
		err = errors.New("Permission denied")
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.UpdateOrganisation(payload, userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(organisation); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	a.notifyChange(ctx, userID, "organisation", events.ActionUpdated, organisationID)
	return organisation, err
}

func (a *APIService) hasEditingPermission(role string) bool {
	editingRoles := []string{"owner", "admin"}
	return slices.Contains(editingRoles, role)
}
