package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) GetUserOrganisationSetting(userID int64) (*models.UserOrganisationSetting, error) {
	setting, err := a.dbService.GetUserOrganisationSetting(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// Create with defaults if not exists
	if setting == nil {
		_, err = a.dbService.CreateUserOrganisationSetting(userID)
		if err != nil {
			logger.Logger.Error(err)
			return nil, err
		}
		setting, err = a.dbService.GetUserOrganisationSetting(userID)
		if err != nil {
			logger.Logger.Error(err)
			return nil, err
		}
	}

	validator := utils.GetValidator()
	if err := validator.Struct(setting); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return setting, nil
}

func (a *APIService) UpdateUserOrganisationSetting(payload models.UpdateUserOrganisationSetting, userID int64) (*models.UserOrganisationSetting, error) {
	// Ensure setting exists (creates with defaults if not)
	_, err := a.GetUserOrganisationSetting(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	err = a.dbService.UpdateUserOrganisationSetting(payload, userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	setting, err := a.dbService.GetUserOrganisationSetting(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	validator := utils.GetValidator()
	if err := validator.Struct(setting); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return setting, nil
}
