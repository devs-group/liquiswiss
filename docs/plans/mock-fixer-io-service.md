# Mock Fixer.io Service

## Goal

Make currency exchange rate fetching mockable for local development. Remove dependency on Fixer.io API for running the application locally.

## Motivation

- Remove external API dependency for local development
- Open source friendly - no API keys required to run locally
- Faster startup (no HTTP calls to external service)
- Consistent test data

## Current State

- `IFixerIOService` interface in `backend/internal/service/fixer_io_service/fixer_io_service.go`
- Fetches rates from Fixer.io API on startup (cronjob every 12 hours)
- Already skips fetch in non-production if rates exist in DB
- Stores rates in `fiat_rates` table via `apiService.UpsertFiatRate()`

## Architecture

### Approach: Fixture-Based Mock

Instead of calling Fixer.io API, load exchange rates from a fixture file or seed them directly into the database.

### New Implementation: `MockFixerIOService`

```go
// backend/internal/service/fixer_io_service/mock.go
package fixer_io_service

type MockFixerIOService struct {
    apiService api_service.IAPIService
}

func NewMockFixerIOService(s *api_service.IAPIService) IFixerIOService {
    return &MockFixerIOService{
        apiService: *s,
    }
}

func (m *MockFixerIOService) RequiresInitialFetch() (bool, error) {
    // Same logic as real service
    totalCurrenciesInRates, err := m.apiService.CountUniqueCurrenciesInFiatRates()
    if err != nil {
        return false, err
    }
    totalCurrencies, err := m.apiService.CountCurrencies()
    if err != nil {
        return false, err
    }
    return totalCurrenciesInRates < totalCurrencies, nil
}

func (m *MockFixerIOService) FetchFiatRates() {
    // Load from fixture and insert into DB
    rates := loadFixtureRates()
    for _, rate := range rates {
        m.apiService.UpsertFiatRate(rate)
    }
}
```

### Fixture File

Create a JSON fixture with realistic exchange rates:

```json
// backend/internal/service/fixer_io_service/fixtures/fiat_rates.json
{
  "rates": [
    { "base": "CHF", "target": "EUR", "rate": 1.04 },
    { "base": "CHF", "target": "USD", "rate": 1.12 },
    { "base": "CHF", "target": "GBP", "rate": 0.88 },
    { "base": "EUR", "target": "CHF", "rate": 0.96 },
    { "base": "EUR", "target": "USD", "rate": 1.08 },
    { "base": "EUR", "target": "GBP", "rate": 0.85 },
    { "base": "USD", "target": "CHF", "rate": 0.89 },
    { "base": "USD", "target": "EUR", "rate": 0.93 },
    { "base": "USD", "target": "GBP", "rate": 0.79 },
    { "base": "GBP", "target": "CHF", "rate": 1.14 },
    { "base": "GBP", "target": "EUR", "rate": 1.18 },
    { "base": "GBP", "target": "USD", "rate": 1.27 }
  ]
}
```

### Configuration

Add to `backend/config/config.go`:

```go
type Config struct {
    // Existing...

    // Fixer.io settings
    FixerIOURl string `env:"FIXER_IO_URL" envDefault:"http://data.fixer.io/api"`
    FixerIOKey string `env:"FIXER_IO_KEY"`

    // Use mock for local development
    UseMockFixerIO bool `env:"USE_MOCK_FIXER_IO" envDefault:"false"`
}
```

### Factory Function

```go
// backend/internal/service/fixer_io_service/factory.go
func NewFixerIOServiceWithConfig(s *api_service.IAPIService, useMock bool) IFixerIOService {
    if useMock {
        return NewMockFixerIOService(s)
    }
    return NewFixerIOService(s)
}
```

### Alternative: Dynamic Migration Fixture

Instead of a separate fixture file, add rates to the dynamic migrations (similar to E2E test data):

```sql
-- backend/internal/db/migrations/dynamic/00002_seed_fiat_rates.sql
-- +goose Up
-- +goose StatementBegin

-- Only insert if no rates exist (for local dev)
INSERT IGNORE INTO fiat_rates (base_currency, target_currency, rate, updated_at)
SELECT 'CHF', 'EUR', 1.04, NOW() FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM fiat_rates WHERE base_currency = 'CHF' AND target_currency = 'EUR');

-- ... more rates

-- +goose StatementEnd

-- +goose Down
-- Dynamic migrations don't need down
```

## Files to Create

| File | Purpose |
|------|---------|
| `backend/internal/service/fixer_io_service/mock.go` | Mock implementation |
| `backend/internal/service/fixer_io_service/factory.go` | Factory function |
| `backend/internal/service/fixer_io_service/fixtures/fiat_rates.json` | Fixture data |

## Files to Modify

| File | Changes |
|------|---------|
| `backend/config/config.go` | Add `UseMockFixerIO` config |
| `backend/main.go` | Use factory to create service |
| `backend/.env.example` | Add `USE_MOCK_FIXER_IO` |

## Implementation Steps

1. Create fixture file with realistic exchange rates
2. Create `MockFixerIOService` that loads from fixture
3. Add factory function
4. Add config option `USE_MOCK_FIXER_IO`
5. Update `main.go` to use factory
6. Update `.env.example`

## Generating Fixture Data

Option 1: **Manual** - Use approximate current rates

Option 2: **One-time fetch** - Run app once with real API, then export:
```sql
SELECT CONCAT('{ "base": "', base_currency, '", "target": "', target_currency, '", "rate": ', rate, ' },')
FROM fiat_rates;
```

Option 3: **Script** - Create a one-time script to fetch and save:
```go
// cmd/fetch-rates/main.go
// Fetches from Fixer.io and saves to fixtures/fiat_rates.json
```

## Environment Examples

### Local Development (.env)
```
USE_MOCK_FIXER_IO=true
# No FIXER_IO_KEY needed
```

### Production (.env)
```
USE_MOCK_FIXER_IO=false
FIXER_IO_KEY=your-api-key
```

## Testing

- Mock service should pass same interface tests as real service
- Verify rates are correctly loaded from fixture
- Verify `RequiresInitialFetch()` works correctly

## Notes

- The existing code already skips fetching in non-production if rates exist
- This enhancement makes it explicit with a config flag
- Fixture data can be updated periodically if needed
- Consider adding a CLI command to refresh fixture from live API
