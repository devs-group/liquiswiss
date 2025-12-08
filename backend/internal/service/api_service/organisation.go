package api_service

import (
	"errors"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListOrganisations(userID int64, page int64, limit int64) ([]models.Organisation, int64, error) {
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

func (a *APIService) GetOrganisation(userID int64, organisationID int64) (*models.Organisation, error) {
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

func (a *APIService) CreateOrganisation(payload models.CreateOrganisation, userID int64, isDefault bool) (*models.Organisation, error) {
	organisationID, err := a.dbService.CreateOrganisation(payload.Name)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.AssignUserToOrganisation(userID, organisationID, "owner", isDefault)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// Switch to the new organisation so that CreateScenario uses the correct organisation_id
	err = a.dbService.SetUserCurrentOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// Create a default scenario for the new organisation
	scenarioID, err := a.dbService.CreateScenario(models.CreateScenario{
		Name: "Standardszenario",
	}, userID, true)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.AssignUserToScenario(userID, organisationID, scenarioID)
	if err != nil {
		return nil, err
	}
	err = a.dbService.SetUserCurrentScenario(userID, scenarioID)
	if err != nil {
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

func (a *APIService) UpdateOrganisation(payload models.UpdateOrganisation, userID int64, organisationID int64) (*models.Organisation, error) {
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
	return organisation, err
}

func (a *APIService) hasEditingPermission(role string) bool {
	editingRoles := []string{"owner", "admin"}
	for _, editingRole := range editingRoles {
		if role == editingRole {
			return true
		}
	}
	return false
}
