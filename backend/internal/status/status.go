// Package status runs dependency checks in the background and keeps the last report for GET /api/status.
package status

import (
	"context"
	"fmt"
	"sync"
	"time"

	"liquiswiss/pkg/models"
)

const (
	StatusOK       = "ok"
	StatusDegraded = "degraded"
	StatusSkipped  = "skipped"
	StatusFailed   = "failed"

	checkInterval = time.Minute
	checkTimeout  = 3 * time.Second
	// The rate sync runs every 12h, so a day plus slack means at least one sync was missed
	maxFiatRateAge = 26 * time.Hour
)

type Check struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	DurationMs int64  `json:"durationMs"`
	Detail     string `json:"detail,omitempty"`
}

type Report struct {
	Status    string    `json:"status"`
	CheckedAt time.Time `json:"checkedAt"`
	Checks    []Check   `json:"checks"`
}

type Pinger interface {
	PingContext(ctx context.Context) error
}

type RateLister func(ctx context.Context) ([]models.FiatRate, error)

type Monitor struct {
	db         Pinger
	listRates  RateLister
	checkRates bool
	now        func() time.Time
	mu         sync.RWMutex
	report     Report
}

// NewMonitor skips the rate check when no Fixer.io key is set, local fallback rates never move
func NewMonitor(db Pinger, listRates RateLister, checkRates bool) *Monitor {
	return &Monitor{
		db:         db,
		listRates:  listRates,
		checkRates: checkRates,
		now:        time.Now,
		report:     Report{Status: StatusDegraded},
	}
}

// Start runs the first check right away, then every minute until ctx ends
func (m *Monitor) Start(ctx context.Context) {
	m.Run(ctx)
	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.Run(ctx)
			}
		}
	}()
}

func (m *Monitor) Run(ctx context.Context) {
	checks := []Check{m.checkDatabase(ctx), m.checkFiatRates(ctx)}
	overall := StatusOK
	for _, check := range checks {
		if check.Status == StatusFailed {
			overall = StatusDegraded
		}
	}
	m.mu.Lock()
	m.report = Report{Status: overall, CheckedAt: m.now().UTC(), Checks: checks}
	m.mu.Unlock()
}

func (m *Monitor) Report() Report {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.report
}

func (m *Monitor) checkDatabase(ctx context.Context) Check {
	start := m.now()
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()
	check := Check{Name: "database", Status: StatusOK}
	if err := m.db.PingContext(ctx); err != nil {
		check.Status = StatusFailed
		check.Detail = "unreachable"
	}
	check.DurationMs = m.now().Sub(start).Milliseconds()
	return check
}

func (m *Monitor) checkFiatRates(ctx context.Context) Check {
	check := Check{Name: "fiatRates", Status: StatusOK}
	if !m.checkRates {
		check.Status = StatusSkipped
		check.Detail = "no Fixer.io key configured"
		return check
	}
	start := m.now()
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()
	rates, err := m.listRates(ctx)
	check.DurationMs = m.now().Sub(start).Milliseconds()
	if err != nil {
		check.Status = StatusFailed
		check.Detail = "not readable"
		return check
	}
	var newest time.Time
	for _, rate := range rates {
		if rate.UpdatedAt.After(newest) {
			newest = rate.UpdatedAt
		}
	}
	if newest.IsZero() {
		check.Status = StatusFailed
		check.Detail = "no rates stored"
		return check
	}
	age := m.now().Sub(newest)
	check.Detail = fmt.Sprintf("updated %s ago", age.Round(time.Minute))
	if age > maxFiatRateAge {
		check.Status = StatusFailed
	}
	return check
}
