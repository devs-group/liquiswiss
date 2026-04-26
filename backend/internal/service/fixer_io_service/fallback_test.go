package fixer_io_service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFallbackRates(t *testing.T) {
	data, err := loadFallbackRates()
	require.NoError(t, err)
	require.NotEmpty(t, data.SnapshotDate)
	require.GreaterOrEqual(t, len(data.Rates), 20)

	var foundCHFEUR bool
	for _, r := range data.Rates {
		require.Len(t, r.Base, 3)
		require.Len(t, r.Target, 3)
		require.Greater(t, r.Rate, 0.0)
		if r.Base == "CHF" && r.Target == "EUR" {
			foundCHFEUR = true
		}
	}
	require.True(t, foundCHFEUR, "expected CHF->EUR pair in fallback rates")
}
