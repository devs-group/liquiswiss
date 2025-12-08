package api_service

import (
	"errors"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListScenarios(userID int64) ([]models.Scenario, error) {
	scenarios, err := a.dbService.ListScenarios(userID)
	if err != nil {
		logger.Logger.Error(err)
		return scenarios, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(scenarios, "dive"); err != nil {
		logger.Logger.Error(err)
		return scenarios, err
	}
	return scenarios, nil
}

func (a *APIService) GetScenario(userID int64, scenarioID int64) (*models.Scenario, error) {
	scenario, err := a.dbService.GetScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(scenario); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return scenario, nil
}

func (a *APIService) CreateScenario(payload models.CreateScenario, userID int64) (*models.Scenario, error) {
	scenarioID, err := a.dbService.CreateScenario(payload, userID, false)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	organisation, err := a.GetCurrentOrganisation(userID)
	if err != nil {
		return nil, err
	}
	// Assign the user to the default scenario
	err = a.dbService.AssignUserToScenario(userID, organisation.ID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Set the default scenario as the current one for this organization
	err = a.dbService.SetUserCurrentScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	scenario, err := a.dbService.GetScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(scenario); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return scenario, nil
}

func (a *APIService) UpdateScenario(payload models.UpdateScenario, userID int64, scenarioID int64) (*models.Scenario, error) {
	_, err := a.dbService.GetScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.UpdateScenario(payload, userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	scenario, err := a.dbService.GetScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(scenario); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return scenario, nil
}

func (a *APIService) DeleteScenario(userID int64, scenarioID int64) error {
	scenario, err := a.dbService.GetScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	if scenario.IsDefault {
		err = errors.New("cannot delete default scenario")
		logger.Logger.Error(err.Error())
		return err
	}
	err = a.dbService.DeleteScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}
