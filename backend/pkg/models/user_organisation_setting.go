package models

import (
	"encoding/json"
	"time"
)

type UserOrganisationSetting struct {
	ID                      int64           `db:"id" json:"id"`
	UserID                  int64           `db:"user_id" json:"userId"`
	OrganisationID          int64           `db:"organisation_id" json:"organisationId"`
	ForecastMonths          int             `db:"forecast_months" json:"forecastMonths"`
	ForecastPerformance     int             `db:"forecast_performance" json:"forecastPerformance"`
	ForecastRevenueDetails  bool            `db:"forecast_revenue_details" json:"forecastRevenueDetails"`
	ForecastExpenseDetails  bool            `db:"forecast_expense_details" json:"forecastExpenseDetails"`
	ForecastChildDetails    json.RawMessage `db:"forecast_child_details" json:"forecastChildDetails"`
	EmployeeDisplay         string          `db:"employee_display" json:"employeeDisplay"`
	EmployeeSortBy          string          `db:"employee_sort_by" json:"employeeSortBy"`
	EmployeeSortOrder       string          `db:"employee_sort_order" json:"employeeSortOrder"`
	EmployeeHideTerminated  bool            `db:"employee_hide_terminated" json:"employeeHideTerminated"`
	TransactionDisplay      string          `db:"transaction_display" json:"transactionDisplay"`
	TransactionSortBy       string          `db:"transaction_sort_by" json:"transactionSortBy"`
	TransactionSortOrder    string          `db:"transaction_sort_order" json:"transactionSortOrder"`
	TransactionHideDisabled bool            `db:"transaction_hide_disabled" json:"transactionHideDisabled"`
	BankAccountDisplay      string          `db:"bank_account_display" json:"bankAccountDisplay"`
	BankAccountSortBy       string          `db:"bank_account_sort_by" json:"bankAccountSortBy"`
	BankAccountSortOrder    string          `db:"bank_account_sort_order" json:"bankAccountSortOrder"`
	CreatedAt               time.Time       `db:"created_at" json:"createdAt"`
	UpdatedAt               time.Time       `db:"updated_at" json:"updatedAt"`
}

type UpdateUserOrganisationSetting struct {
	ForecastMonths          *int             `json:"forecastMonths" validate:"omitempty,min=1,max=60"`
	ForecastPerformance     *int             `json:"forecastPerformance" validate:"omitempty,min=0,max=200"`
	ForecastRevenueDetails  *bool            `json:"forecastRevenueDetails" validate:"omitempty"`
	ForecastExpenseDetails  *bool            `json:"forecastExpenseDetails" validate:"omitempty"`
	ForecastChildDetails    *json.RawMessage `json:"forecastChildDetails" validate:"omitempty"`
	EmployeeDisplay         *string          `json:"employeeDisplay" validate:"omitempty,oneof=grid list"`
	EmployeeSortBy          *string          `json:"employeeSortBy" validate:"omitempty"`
	EmployeeSortOrder       *string          `json:"employeeSortOrder" validate:"omitempty,oneof=ASC DESC"`
	EmployeeHideTerminated  *bool            `json:"employeeHideTerminated" validate:"omitempty"`
	TransactionDisplay      *string          `json:"transactionDisplay" validate:"omitempty,oneof=grid list"`
	TransactionSortBy       *string          `json:"transactionSortBy" validate:"omitempty"`
	TransactionSortOrder    *string          `json:"transactionSortOrder" validate:"omitempty,oneof=ASC DESC"`
	TransactionHideDisabled *bool            `json:"transactionHideDisabled" validate:"omitempty"`
	BankAccountDisplay      *string          `json:"bankAccountDisplay" validate:"omitempty,oneof=grid list"`
	BankAccountSortBy       *string          `json:"bankAccountSortBy" validate:"omitempty"`
	BankAccountSortOrder    *string          `json:"bankAccountSortOrder" validate:"omitempty,oneof=ASC DESC"`
}
