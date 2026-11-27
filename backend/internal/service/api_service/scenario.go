package api_service

import (
	"context"
	"errors"
	"liquiswiss/internal/events"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/reqctx"
)

// ErrBaseScenario is returned when a mutation is not allowed on the base scenario
var ErrBaseScenario = errors.New("base scenario cannot be deleted")

// scenarioScopedEntities only reach users who view the same scenario, everything else is shared by the organisation
var scenarioScopedEntities = map[string]bool{
	"employee":           true,
	"salary":             true,
	"salary_cost":        true,
	"transaction":        true,
	"bank_account":       true,
	"forecast_exclusion": true,
}

func (a *APIService) GetCurrentScenarioID(ctx context.Context, userID int64) (int64, error) {
	return a.dbService.GetCurrentScenarioID(userID)
}

func (a *APIService) ListScenarios(ctx context.Context, userID int64) ([]models.Scenario, error) {
	scenarios, err := a.dbService.ListScenarios(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return scenarios, nil
}

func (a *APIService) GetScenario(ctx context.Context, userID int64, scenarioID int64) (*models.Scenario, error) {
	scenario, err := a.dbService.GetScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return scenario, nil
}

// CreateScenario copies the source (default: current) scenario and switches the user into the copy
func (a *APIService) CreateScenario(ctx context.Context, payload models.CreateScenario, userID int64) (*models.Scenario, error) {
	sourceID := int64(0)
	if payload.SourceScenarioID != nil {
		sourceID = *payload.SourceScenarioID
	} else {
		currentID, err := a.dbService.GetCurrentScenarioID(userID)
		if err != nil {
			logger.Logger.Error(err)
			return nil, err
		}
		sourceID = currentID
	}

	scenarioID, err := a.dbService.CreateScenario(userID, payload.Name, sourceID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	a.notifyChange(ctx, userID, "scenario", events.ActionCreated, scenarioID)

	return a.SwitchScenario(ctx, userID, scenarioID)
}

func (a *APIService) UpdateScenario(ctx context.Context, payload models.UpdateScenario, userID int64, scenarioID int64) (*models.Scenario, error) {
	if _, err := a.dbService.GetScenario(userID, scenarioID); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	if err := a.dbService.UpdateScenario(payload, userID, scenarioID); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	scenario, err := a.dbService.GetScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	a.notifyChange(ctx, userID, "scenario", events.ActionUpdated, scenarioID)
	return scenario, nil
}

func (a *APIService) DeleteScenario(ctx context.Context, userID int64, scenarioID int64) error {
	scenario, err := a.dbService.GetScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	if scenario.IsBase {
		return ErrBaseScenario
	}
	if err := a.dbService.DeleteScenario(userID, scenarioID); err != nil {
		logger.Logger.Error(err)
		return err
	}
	// The user falls back to the base scenario, its forecast may be stale
	if scenario.IsCurrent {
		if _, err := a.CalculateForecast(ctx, userID); err != nil {
			logger.Logger.Error(err)
			return err
		}
	}
	a.notifyChange(ctx, userID, "scenario", events.ActionDeleted, scenarioID)
	return nil
}

func (a *APIService) SwitchScenario(ctx context.Context, userID int64, scenarioID int64) (*models.Scenario, error) {
	if err := a.dbService.SetUserCurrentScenario(userID, scenarioID); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Edits made by others while the user looked elsewhere are only in the forecast after a recalculation
	if _, err := a.CalculateForecast(ctx, userID); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	scenario, err := a.dbService.GetScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	a.notifyScenarioSwitch(ctx, userID, scenarioID)
	return scenario, nil
}

// GetScenarioForecast returns another scenario's stored forecast without switching or recalculating it
func (a *APIService) GetScenarioForecast(ctx context.Context, userID int64, scenarioID int64, limit int64) (*models.ScenarioForecast, error) {
	scenario, err := a.dbService.GetScenario(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	forecasts, err := a.dbService.ListScenarioForecasts(userID, scenarioID, limit)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	details, err := a.dbService.ListScenarioForecastDetails(userID, scenarioID, limit)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	bankAccounts, err := a.dbService.ListScenarioBankAccounts(userID, scenarioID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return &models.ScenarioForecast{
		Scenario:        *scenario,
		Forecasts:       forecasts,
		ForecastDetails: details,
		BankAccounts:    bankAccounts,
	}, nil
}

// notifyScenarioSwitch tells every open tab of this user to follow the switch (MCP or another tab)
func (a *APIService) notifyScenarioSwitch(ctx context.Context, userID int64, scenarioID int64) {
	if a.eventHub == nil {
		return
	}
	user, err := a.dbService.GetProfile(userID)
	if err != nil {
		logger.Logger.Warnf("events: could not resolve organisation for user %d: %v", userID, err)
		return
	}
	a.eventHub.Publish(events.Event{
		Entity:         "scenario",
		Action:         events.ActionSwitched,
		ID:             scenarioID,
		OrganisationID: user.CurrentOrganisationID,
		OnlyUserID:     userID,
		OriginUserID:   userID,
		OriginClientID: reqctx.ClientID(ctx),
	})
}
