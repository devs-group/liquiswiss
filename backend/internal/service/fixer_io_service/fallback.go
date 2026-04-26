package fixer_io_service

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed fallback_rates.json
var fallbackRatesJSON []byte

type fallbackRate struct {
	Base   string  `json:"base"`
	Target string  `json:"target"`
	Rate   float64 `json:"rate"`
}

type fallbackRatesFile struct {
	SnapshotDate string         `json:"snapshot_date"`
	Source       string         `json:"source"`
	Rates        []fallbackRate `json:"rates"`
}

func loadFallbackRates() (*fallbackRatesFile, error) {
	var data fallbackRatesFile
	if err := json.Unmarshal(fallbackRatesJSON, &data); err != nil {
		return nil, fmt.Errorf("parse fallback_rates.json: %w", err)
	}
	return &data, nil
}
