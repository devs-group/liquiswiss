package status

import (
	"context"
	"errors"
	"testing"
	"time"

	"liquiswiss/pkg/models"

	"github.com/stretchr/testify/require"
)

type fakePinger struct{ err error }

func (f fakePinger) PingContext(context.Context) error { return f.err }

func ratesUpdated(at time.Time) RateLister {
	return func(context.Context) ([]models.FiatRate, error) {
		return []models.FiatRate{{UpdatedAt: at.Add(-time.Hour)}, {UpdatedAt: at}}, nil
	}
}

func checkByName(report Report, name string) Check {
	for _, check := range report.Checks {
		if check.Name == name {
			return check
		}
	}
	return Check{}
}

func TestMonitor(t *testing.T) {
	now := time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)

	t.Run("all ok", func(t *testing.T) {
		m := NewMonitor(fakePinger{}, ratesUpdated(now.Add(-3*time.Hour)), true)
		m.now = func() time.Time { return now }
		m.Run(context.Background())
		report := m.Report()
		require.Equal(t, StatusOK, report.Status)
		require.Equal(t, StatusOK, checkByName(report, "database").Status)
		require.Equal(t, "updated 3h0m0s ago", checkByName(report, "fiatRates").Detail)
	})

	t.Run("database down", func(t *testing.T) {
		m := NewMonitor(fakePinger{err: errors.New("dial tcp: refused")}, ratesUpdated(now), true)
		m.now = func() time.Time { return now }
		m.Run(context.Background())
		report := m.Report()
		require.Equal(t, StatusDegraded, report.Status)
		require.Equal(t, StatusFailed, checkByName(report, "database").Status)
		// The raw error never leaves the process
		require.Equal(t, "unreachable", checkByName(report, "database").Detail)
	})

	t.Run("stale rates", func(t *testing.T) {
		m := NewMonitor(fakePinger{}, ratesUpdated(now.Add(-30*time.Hour)), true)
		m.now = func() time.Time { return now }
		m.Run(context.Background())
		require.Equal(t, StatusDegraded, m.Report().Status)
		require.Equal(t, StatusFailed, checkByName(m.Report(), "fiatRates").Status)
	})

	t.Run("rates skipped without key", func(t *testing.T) {
		m := NewMonitor(fakePinger{}, ratesUpdated(now.Add(-300*time.Hour)), false)
		m.now = func() time.Time { return now }
		m.Run(context.Background())
		require.Equal(t, StatusOK, m.Report().Status)
		require.Equal(t, StatusSkipped, checkByName(m.Report(), "fiatRates").Status)
	})

	t.Run("before the first run", func(t *testing.T) {
		m := NewMonitor(fakePinger{}, ratesUpdated(now), true)
		require.Equal(t, StatusDegraded, m.Report().Status)
	})
}
