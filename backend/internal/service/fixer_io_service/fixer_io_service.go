//go:generate mockgen -package=mocks -destination ../../mocks/fixer_io_service.go liquiswiss/internal/service/fixer_io_service IFixerIOService
package fixer_io_service

import (
	"encoding/json"
	"fmt"
	"liquiswiss/config"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"strings"
)

type IFixerIOService interface {
	FetchFiatRates()
	RequiresInitialFetch() (bool, error)
}

type FixerIOService struct {
	apiService api_service.IAPIService
}

func NewFixerIOService(s *api_service.IAPIService) IFixerIOService {
	return &FixerIOService{
		apiService: *s,
	}
}

func (f *FixerIOService) RequiresInitialFetch() (bool, error) {
	totalCurrenciesInRates, err := f.apiService.CountUniqueCurrenciesInFiatRates()
	if err != nil {
		return false, err
	}
	totalCurrencies, err := f.apiService.CountCurrencies()
	if err != nil {
		return false, err
	}
	return totalCurrenciesInRates < totalCurrencies, nil
}

func (f *FixerIOService) FetchFiatRates() {
	if !utils.IsProduction() {
		fiatRates, err := f.apiService.ListFiatRates("CHF")
		if err == nil && len(fiatRates) > 0 {
			logger.Logger.Debug("Skipping Fiat Rates because we are not on Production and already have some Fiat Rates")
			return
		}
	}

	logger.Logger.Infof("Running Fixer.io Cronjob")

	cfg := config.GetConfig()
	currencies, err := f.apiService.ListCurrencies(0)
	if err != nil {
		logger.Logger.Errorf("Failed to load currencies: %v", err)
		return
	}

	symbols := make([]string, 0)
	for _, currency := range currencies {
		symbols = append(symbols, *currency.Code)
	}

	for _, currency := range currencies {
		baseCurrency := *currency.Code

		url := fmt.Sprintf("%s/latest?access_key=%s&base=%s&symbols=%s", cfg.FixerIOURl, cfg.FixerIOKey, baseCurrency, strings.Join(symbols, ","))
		logger.Logger.Infof("Sending request to %s", url)
		response, err := http.Get(url)
		if err != nil {
			logger.Logger.Errorf("Failed to fetch fiat rates: %v", err)
			return
		}
		defer func() {
			err := response.Body.Close()
			if err != nil {
				logger.Logger.Errorf("Error closing response body: %v", err)
			}
		}()
		if response.StatusCode != http.StatusOK {
			logger.Logger.Errorf("Non-OK HTTP status: %d", response.StatusCode)
			return
		}
		var exchangeData models.FixerIOResponse
		if err := json.NewDecoder(response.Body).Decode(&exchangeData); err != nil {
			logger.Logger.Errorf("Failed to decode JSON: %v", err)
			return
		}

		if !exchangeData.Success {
			logger.Logger.Errorf("Request failed: %v", exchangeData.Error)
			return
		}

		for targetCurrency, rate := range *exchangeData.Rates {
			err = f.apiService.UpsertFiatRate(models.CreateFiatRate{
				Base:   baseCurrency,
				Target: targetCurrency,
				Rate:   rate,
			})
			if err != nil {
				logger.Logger.Errorf("Failed to insert fiat rate for base %s to target %s: %v", baseCurrency, targetCurrency, err)
				continue
			}
			logger.Logger.Infof("Successfully updated fiat rate for base %s to target %s with %f", baseCurrency, targetCurrency, rate)
		}
	}
}
