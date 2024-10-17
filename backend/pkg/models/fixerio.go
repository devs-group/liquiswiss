package models

import "time"

type FixerIOError struct {
	Code int    `json:"code"`
	Type string `json:"type"`
	Info string `json:"info"`
}

type FixerIOResponse struct {
	Success   bool                `json:"success"`
	Timestamp int64               `json:"timestamp"`
	Base      string              `json:"base"`
	Date      string              `json:"date"`
	Rates     *map[string]float64 `json:"rates"`
	Error     *FixerIOError       `json:"error"`
}

type FiatRate struct {
	ID        int64     `json:"id"`
	Base      string    `json:"base"`
	Target    string    `json:"target"`
	Rate      float64   `json:"rate"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateFiatRate struct {
	Base   string  `json:"base"`
	Target string  `json:"target"`
	Rate   float64 `json:"rate"`
}
