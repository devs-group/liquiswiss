//go:generate mockgen -package=mocks -destination ../../mocks/fixer_io_service.go liquiswiss/internal/service/fixer_io_service IFixerIOService
package fixer_io_service

import (
	"context"
	"encoding/json"
	"liquiswiss/config"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	totalCurrenciesInRates, err := f.apiService.CountUniqueCurrenciesInFiatRates(context.Background())
	if err != nil {
		return false, err
	}
	totalCurrencies, err := f.apiService.CountCurrencies(context.Background())
	if err != nil {
		return false, err
	}
	if totalCurrenciesInRates < totalCurrencies {
		return true, nil
	}
	// Every deploy restarts the 12h cron, so stale rates are fetched on start as well
	rates, err := f.apiService.ListFiatRates(context.Background(), "CHF")
	if err != nil {
		return false, err
	}
	return ratesAreStale(rates, time.Now()), nil
}

func ratesAreStale(rates []models.FiatRate, now time.Time) bool {
	var newest time.Time
	for _, rate := range rates {
		if rate.UpdatedAt.After(newest) {
			newest = rate.UpdatedAt
		}
	}
	return now.Sub(newest) > 12*time.Hour
}

func (f *FixerIOService) fetchFromFallback() {
	if existing, err := f.apiService.ListFiatRates(context.Background(), "CHF"); err == nil && len(existing) > 0 {
		logger.Logger.Debug("Fixer.io API key not configured — fallback rates already loaded, skipping")
		return
	}

	logger.Logger.Info("Fixer.io API key not configured — loading fiat rates from local fallback")

	data, err := loadFallbackRates()
	if err != nil {
		logger.Logger.Errorf("Failed to load fallback rates: %v", err)
		return
	}

	for _, r := range data.Rates {
		err := f.apiService.UpsertFiatRate(context.Background(), models.CreateFiatRate{
			Base:   r.Base,
			Target: r.Target,
			Rate:   r.Rate,
		})
		if err != nil {
			logger.Logger.Errorf("Failed to upsert fallback rate %s->%s: %v", r.Base, r.Target, err)
			continue
		}
	}
	logger.Logger.Infof("Loaded %d fallback fiat rates (snapshot %s)", len(data.Rates), data.SnapshotDate)
}

// buildLatestURL accepts the API root or the full latest endpoint, FIXER_IO_URL has been set both ways
func buildLatestURL(baseURL, accessKey, base string, symbols []string) string {
	endpoint := strings.TrimSuffix(strings.TrimRight(baseURL, "/"), "/latest") + "/latest"
	query := url.Values{}
	query.Set("access_key", accessKey)
	query.Set("base", base)
	query.Set("symbols", strings.Join(symbols, ","))
	return endpoint + "?" + query.Encode()
}

func (f *FixerIOService) FetchFiatRates() {
	cfg := config.GetConfig()
	if cfg.FixerIOKey == "" {
		f.fetchFromFallback()
		return
	}

	if !utils.IsProduction() {
		fiatRates, err := f.apiService.ListFiatRates(context.Background(), "CHF")
		if err == nil && len(fiatRates) > 0 {
			logger.Logger.Debug("Skipping Fixer.io fetch because we are not on Production and already have some Fiat Rates")
			return
		}
	}

	logger.Logger.Infof("Running Fixer.io Cronjob")

	currencies, err := f.apiService.ListCurrencies(context.Background(), 0)
	if err != nil {
		logger.Logger.Errorf("Failed to load currencies: %v", err)
		return
	}

	symbols := make([]string, 0)
	for _, currency := range currencies {
		symbols = append(symbols, *currency.Code)
	}

	// One request with base EUR, every other pair is crossed locally to save Fixer.io quota
	requestURL := buildLatestURL(cfg.FixerIOURl, cfg.FixerIOKey, crossBase, symbols)
	// Never log the URL itself, it carries the access key
	logger.Logger.Infof("Requesting Fixer.io rates for base %s", crossBase)
	response, err := http.Get(requestURL)
	if err != nil {
		logger.Logger.Errorf("Failed to fetch fiat rates: %v", err)
		return
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
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
	if !exchangeData.Success || exchangeData.Rates == nil {
		logger.Logger.Errorf("Request failed: %v", exchangeData.Error)
		return
	}

	updated := 0
	for _, rate := range crossRates(*exchangeData.Rates, symbols) {
		if err := f.apiService.UpsertFiatRate(context.Background(), rate); err != nil {
			logger.Logger.Errorf("Failed to insert fiat rate for base %s to target %s: %v", rate.Base, rate.Target, err)
			continue
		}
		updated++
	}
	logger.Logger.Infof("Updated %d fiat rates from one Fixer.io request", updated)
}

const crossBase = "EUR"

// crossRates turns EUR based rates into every pair: A to B is EUR to B divided by EUR to A
func crossRates(eurRates map[string]float64, codes []string) []models.CreateFiatRate {
	rateFromEUR := func(code string) float64 {
		if code == crossBase {
			return 1
		}
		return eurRates[code]
	}
	rates := make([]models.CreateFiatRate, 0, len(codes)*len(codes))
	for _, base := range codes {
		baseRate := rateFromEUR(base)
		if baseRate <= 0 {
			continue
		}
		for _, target := range codes {
			targetRate := rateFromEUR(target)
			if targetRate <= 0 {
				continue
			}
			rates = append(rates, models.CreateFiatRate{Base: base, Target: target, Rate: targetRate / baseRate})
		}
	}
	return rates
}
