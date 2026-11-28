package fixer_io_service

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

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
