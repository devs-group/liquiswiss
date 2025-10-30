package models

import "liquiswiss/pkg/types"

type Salary struct {
	ID                  int64         `db:"id" json:"id"`
	EmployeeID          int64         `db:"employee_id" json:"employeeID"`
	HoursPerMonth       uint16        `db:"hours_per_month" json:"hoursPerMonth"`
	Amount              uint64        `db:"amount" json:"amount"`
	Cycle               string        `db:"cycle" json:"cycle"`
	Currency            Currency      `db:"currency_id" json:"currency"`
	VacationDaysPerYear uint16        `db:"vacation_days_per_year" json:"vacationDaysPerYear"`
	FromDate            types.AsDate  `db:"from_date" json:"fromDate"`
	ToDate              *types.AsDate `db:"to_date" json:"toDate"`
	WithSeparateCosts   bool          `db:"with_separate_costs" json:"withSeparateCosts"`
	IsTermination       bool          `db:"is_termination" json:"isTermination"`
	IsDisabled          bool          `db:"is_disabled" json:"isDisabled"`

	// Hidden values
	DBDate types.AsDate `db:"db_date" json:"-"`

	// Calculated Values
	NextExecutionDate  *types.AsDate `db:"-" json:"nextExecutionDate"`
	EmployeeDeductions uint64        `db:"-" json:"employeeDeductions"`
	EmployerCosts      uint64        `db:"-" json:"employerCosts"`
}

type CreateSalary struct {
	HoursPerMonth       uint16  `json:"hoursPerMonth" validate:"gte=0"`
	Amount              uint64  `json:"amount" validate:"gte=0"`
	Cycle               string  `json:"cycle" validate:"allowedCycles"`
	CurrencyID          int64   `json:"currencyID" validate:"required,gte=0"`
	VacationDaysPerYear uint16  `json:"vacationDaysPerYear" validate:"gte=0"`
	FromDate            string  `json:"fromDate" validate:"required"`
	ToDate              *string `json:"toDate" validate:"omitempty,fromDateGTEToDate"`
	WithSeparateCosts   bool    `json:"withSeparateCosts"`
	IsTermination       bool    `json:"isTermination"`
}

type UpdateSalary struct {
	HoursPerMonth       *uint16 `json:"hoursPerMonth" validate:"omitempty,gte=0"`
	Amount              *uint64 `json:"amount" validate:"omitempty,gte=0"`
	Cycle               *string `json:"cycle" validate:"omitempty,allowedCycles"`
	CurrencyID          *int64  `json:"currencyID" validate:"omitempty,gte=0"`
	VacationDaysPerYear *uint16 `json:"vacationDaysPerYear" validate:"omitempty,gte=0"`
	FromDate            *string `json:"fromDate" validate:"omitempty"`
	ToDate              *string `json:"toDate" validate:"omitempty,fromDateGTEToDate"`
	WithSeparateCosts   *bool   `json:"withSeparateCosts" validate:"omitempty"`
	IsDisabled          *bool   `json:"isDisabled" validate:"omitempty"`
}
