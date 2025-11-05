package api_service

import (
	"golang.org/x/crypto/bcrypt"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) GetProfile(userID int64) (*models.User, error) {
	user, err := a.dbService.GetProfile(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(user); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return user, nil
}

func (a *APIService) UpdateProfile(payload models.UpdateUser, userID int64) (*models.User, error) {
	err := a.dbService.UpdateProfile(payload, userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	user, err := a.dbService.GetProfile(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(user); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return user, nil
}

func (a *APIService) UpdatePassword(payload models.UpdateUserPassword, userID int64) error {
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	err = a.dbService.UpdatePassword(userID, string(encryptedPassword))
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}

func (a *APIService) SetUserCurrentOrganisation(payload models.UpdateUserCurrentOrganisation, userID int64) error {
	err := a.dbService.SetUserCurrentOrganisation(userID, payload.OrganisationID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	defaultScenarioID, err := a.dbService.GetDefaultScenarioID(payload.OrganisationID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	err = a.dbService.SetUserCurrentScenario(userID, defaultScenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}

func (a *APIService) GetCurrentOrganisation(userID int64) (*models.Organisation, error) {
	user, err := a.dbService.GetProfile(userID)
	if err != nil {
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(user); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	organisation, err := a.dbService.GetOrganisation(userID, user.CurrentOrganisationID)
	if err != nil {
		return nil, err
	}
	if err := validator.Struct(organisation); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return organisation, nil
}
