package api_service

import (
	"errors"

	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

var (
	ErrCannotDeleteDefaultScenario = errors.New("cannot delete default scenario")
)

func (a *APIService) ListScenarios(userID int64) ([]models.ScenarioListItem, error) {
	scenarios, err := a.dbService.ListScenarios(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(scenarios, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, err
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

func (a *APIService) GetDefaultScenario(userID int64) (*models.Scenario, error) {
	scenario, err := a.dbService.GetDefaultScenario(userID)
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

	// If horizontal scenario and has parent, copy data from parent
	if payload.Type == models.ScenarioTypeHorizontal && payload.ParentScenarioID != nil {
		err = a.dbService.CopyScenarioData(*payload.ParentScenarioID, scenarioID, userID)
		if err != nil {
			logger.Logger.Error(err)
			return nil, err
		}
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
	existingScenario, err := a.dbService.GetScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	if payload.Name == nil {
		name := existingScenario.Name
		payload.Name = &name
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
	existingScenario, err := a.dbService.GetScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	if existingScenario.IsDefault {
		return ErrCannotDeleteDefaultScenario
	}

	// Get all users currently using this scenario
	usersUsingScenario, err := a.dbService.GetUsersUsingScenario(scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	// Determine the target scenario to switch users to
	var targetScenarioID int64
	if existingScenario.ParentScenarioID != nil {
		// Switch to parent scenario
		targetScenarioID = *existingScenario.ParentScenarioID
	} else {
		// Switch to default scenario
		defaultScenario, err := a.dbService.GetDefaultScenario(userID)
		if err != nil {
			logger.Logger.Error(err)
			return err
		}
		targetScenarioID = defaultScenario.ID
	}

	// Switch all users using this scenario to the target scenario
	for _, affectedUserID := range usersUsingScenario {
		err = a.dbService.SetUserCurrentScenario(affectedUserID, targetScenarioID)
		if err != nil {
			logger.Logger.Error(err)
			return err
		}
	}

	// Now delete the scenario
	err = a.dbService.DeleteScenario(userID, existingScenario.ID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}
