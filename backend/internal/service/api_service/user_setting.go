package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) GetUserSetting(userID int64) (*models.UserSetting, error) {
	setting, err := a.dbService.GetUserSetting(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// Create with defaults if not exists
	if setting == nil {
		_, err = a.dbService.CreateUserSetting(userID)
		if err != nil {
			logger.Logger.Error(err)
			return nil, err
		}
		setting, err = a.dbService.GetUserSetting(userID)
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

func (a *APIService) UpdateUserSetting(payload models.UpdateUserSetting, userID int64) (*models.UserSetting, error) {
	// Ensure setting exists (creates with defaults if not)
	_, err := a.GetUserSetting(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	err = a.dbService.UpdateUserSetting(payload, userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	setting, err := a.dbService.GetUserSetting(userID)
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
