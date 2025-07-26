package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListFiatRates(base string) ([]models.FiatRate, error) {
	fiatRates, err := a.dbService.ListFiatRates(base)
	if err != nil {
		logger.Logger.Error(err)
		return fiatRates, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(fiatRates, "dive"); err != nil {
		logger.Logger.Error(err)
		return fiatRates, err
	}
	return fiatRates, nil
}

func (a *APIService) GetFiatRate(base, target string) (*models.FiatRate, error) {
	fiatRate, err := a.dbService.GetFiatRate(base, target)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(fiatRate); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return fiatRate, nil
}

func (a *APIService) UpsertFiatRate(payload models.CreateFiatRate) error {
	err := a.dbService.UpsertFiatRate(payload)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}

func (a *APIService) CountUniqueCurrenciesInFiatRates() (int64, error) {
	totalCount, err := a.dbService.CountUniqueCurrenciesInFiatRates()
	if err != nil {
		logger.Logger.Error(err)
		return 0, err
	}
	return totalCount, nil
}
