package models

import "time"

type ForecastData struct {
	Month    string `db:"month" json:"month"`
	Revenue  int64  `db:"revenue" json:"revenue"`
	Expense  int64  `db:"expense" json:"expense"`
	Cashflow int64  `db:"cashflow" json:"cashflow"`
}

type Forecast struct {
	UpdatedAt *time.Time   `json:"updatedAt"`
	Data      ForecastData `json:"data"`
}

type ForecastDetailRevenueExpense struct {
	Name   string `json:"name"`
	Amount int64  `json:"amount,omitempty"` // Only set for leaf nodes
	// Recursive
	Children []ForecastDetailRevenueExpense `json:"children,omitempty"`
}

type ForecastDetails struct {
	Revenue    map[string]interface{} `db:"revenue" json:"revenue"`
	Expense    map[string]interface{} `db:"expense" json:"expense"`
	ForecastID int64                  `db:"forecast_id" json:"forecastID"`
}

type ForecastDatabaseDetails struct {
	Month      string                         `db:"month" json:"month"`
	Revenue    []ForecastDetailRevenueExpense `db:"revenue" json:"revenue"`
	Expense    []ForecastDetailRevenueExpense `db:"expense" json:"expense"`
	ForecastID *int64                         `db:"forecast_id" json:"forecastID"`
}

type CreateForecast struct {
	Month    string `json:"month" validate:"required,max=7"`
	Revenue  int64  `json:"revenue" validate:"required"`
	Expense  int64  `json:"expense" validate:"required"`
	Cashflow int64  `json:"cashflow" validate:"required"`
}

type CreateForecastDetail struct {
	Month      string                         `json:"month" validate:"required,max=7"`
	Revenue    []ForecastDetailRevenueExpense `json:"revenue" validate:"required"`
	Expense    []ForecastDetailRevenueExpense `json:"expense" validate:"required"`
	ForecastID int64                          `json:"forecast_id" validate:"required"`
}
