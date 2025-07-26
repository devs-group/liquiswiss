package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListCurrencies(userID int64) ([]models.Currency, error) {
	currencies, err := a.dbService.ListCurrencies(userID)
	if err != nil {
		logger.Logger.Error(err)
		return currencies, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(currencies, "dive"); err != nil {
		logger.Logger.Error(err)
		return currencies, err
	}
	return currencies, nil
}

func (a *APIService) GetCurrency(currencyID int64) (*models.Currency, error) {
	currency, err := a.dbService.GetCurrency(currencyID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(currency); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return currency, nil
}

func (a *APIService) CreateCurrency(payload models.CreateCurrency) (*models.Currency, error) {
	currencyID, err := a.dbService.CreateCurrency(payload)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	currency, err := a.dbService.GetCurrency(currencyID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(currency); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return currency, nil
}

func (a *APIService) UpdateCurrency(payload models.UpdateCurrency, currencyID int64) (*models.Currency, error) {
	err := a.dbService.UpdateCurrency(payload, currencyID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	currency, err := a.dbService.GetCurrency(currencyID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(currency); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return currency, nil
}

func (a *APIService) CountCurrencies() (int64, error) {
	totalCount, err := a.dbService.CountCurrencies()
	if err != nil {
		return 0, err
	}
	return totalCount, nil
}
