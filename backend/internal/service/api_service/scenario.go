package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListScenarios(userID, page, limit int64) ([]models.Scenario, int64, error) {
	scenarios, totalCount, err := a.dbService.ListScenarios(userID, page, limit)
	if err != nil {
		logger.Logger.Error(err)
		return scenarios, totalCount, err
	}

	validator := utils.GetValidator()
	if err := validator.Var(scenarios, "dive"); err != nil {
		logger.Logger.Error(err)
		return scenarios, totalCount, err
	}

	return scenarios, totalCount, nil
}

func (a *APIService) GetScenario(userID, scenarioID int64) (*models.Scenario, error) {
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
	scenarioID, err := a.dbService.CreateScenario(payload, userID)
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

func (a *APIService) UpdateScenario(payload models.UpdateScenario, userID, scenarioID int64) (*models.Scenario, error) {
	err := a.dbService.UpdateScenario(payload, userID, scenarioID)
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

func (a *APIService) DeleteScenario(userID, scenarioID int64) error {
	err := a.dbService.DeleteScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	return nil
}
