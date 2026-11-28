package fixer_io_service

import (
	"liquiswiss/pkg/models"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCrossRates(t *testing.T) {
	eurRates := map[string]float64{"CHF": 0.9155, "USD": 1.0950}
	rates := crossRates(eurRates, []string{"CHF", "EUR", "USD", "XXX"})

	byPair := map[string]float64{}
	for _, rate := range rates {
		byPair[rate.Base+"->"+rate.Target] = rate.Rate
	}
	// XXX has no EUR rate, so it is skipped as base and as target
	require.Len(t, rates, 9)
	require.InDelta(t, 1.0, byPair["CHF->CHF"], 1e-9)
	require.InDelta(t, 0.9155, byPair["EUR->CHF"], 1e-9)
	require.InDelta(t, 1/0.9155, byPair["CHF->EUR"], 1e-9)
	require.InDelta(t, 0.9155/1.0950, byPair["USD->CHF"], 1e-9)
	require.InDelta(t, 1.0950/0.9155, byPair["CHF->USD"], 1e-9)
}

func TestRatesAreStale(t *testing.T) {
	now := time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)
	fresh := []models.FiatRate{
		{UpdatedAt: now.Add(-30 * time.Hour)},
		{UpdatedAt: now.Add(-2 * time.Hour)},
	}
	stale := []models.FiatRate{{UpdatedAt: now.Add(-13 * time.Hour)}}

	require.False(t, ratesAreStale(fresh, now))
	require.True(t, ratesAreStale(stale, now))
	require.True(t, ratesAreStale(nil, now))
}

func TestBuildLatestURL(t *testing.T) {
	symbols := []string{"CHF", "EUR", "USD"}
	for _, baseURL := range []string{
		"https://data.fixer.io/api/latest",
		"https://data.fixer.io/api/latest/",
		"https://data.fixer.io/api",
		"https://data.fixer.io/api/",
	} {
		t.Run(baseURL, func(t *testing.T) {
			parsed, err := url.Parse(buildLatestURL(baseURL, "secret", "USD", symbols))
			require.NoError(t, err)
			require.Equal(t, "/api/latest", parsed.Path)
			require.Equal(t, "secret", parsed.Query().Get("access_key"))
			require.Equal(t, "USD", parsed.Query().Get("base"))
			require.Equal(t, "CHF,EUR,USD", parsed.Query().Get("symbols"))
		})
	}
}
