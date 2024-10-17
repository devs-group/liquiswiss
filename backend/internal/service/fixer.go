package service

import (
	"encoding/json"
	"fmt"
	"liquiswiss/config"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"strings"
)

type IFixerIOService interface {
	FetchFiatRates()
}

type FixerIOService struct {
	dbService IDatabaseService
}

func NewFixerIOService(s IDatabaseService) IFixerIOService {
	return &FixerIOService{
		dbService: s,
	}
}

func (f *FixerIOService) FetchFiatRates() {
	logger.Logger.Infof("Running Fixer.io Cronjob")

	cfg := config.GetConfig()
	currencies, _, err := f.dbService.ListCurrencies(1, 1000)
	if err != nil {
		logger.Logger.Errorf("Failed to load currencies: %v", err)
		return
	}
	symbols := make([]string, 0)
	for _, currency := range currencies {
		if *currency.Code == utils.BaseCurrency {
			continue
		}
		symbols = append(symbols, *currency.Code)
	}
	url := fmt.Sprintf("%s/latest?access_key=%s&base=CHF&symbols=%s", cfg.FixerIOURl, cfg.FixerIOKey, strings.Join(symbols, ","))
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
		err = f.dbService.UpsertFiatRate(models.CreateFiatRate{
			Base:   utils.BaseCurrency,
			Target: targetCurrency,
			Rate:   rate,
		})
		if err != nil {
			logger.Logger.Errorf("Failed to insert fiat rate for %s: %v", targetCurrency, err)
			continue
		}
		logger.Logger.Infof("Successfully updated fiat rate for %s with %f", targetCurrency, rate)
	}
}
