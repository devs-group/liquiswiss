package models

import "liquiswiss/pkg/types"

type EmployeeHistory struct {
	ID                  int64         `db:"id" json:"id"`
	EmployeeID          int64         `db:"employee_id" json:"employeeID"`
	HoursPerMonth       uint16        `db:"hours_per_month" json:"hoursPerMonth"`
	SalaryPerMonth      uint64        `db:"salary_per_month" json:"salaryPerMonth"`
	Currency            Currency      `db:"currency_id" json:"currency"`
	VacationDaysPerYear uint16        `db:"vacation_days_per_year" json:"vacationDaysPerYear"`
	FromDate            types.AsDate  `db:"from_date" json:"fromDate"`
	ToDate              *types.AsDate `db:"to_date" json:"toDate"`
}

type CreateEmployeeHistory struct {
	HoursPerMonth       uint16  `json:"hoursPerMonth" validate:"gte=0"`
	SalaryPerMonth      uint64  `json:"salaryPerMonth" validate:"gte=0"`
	CurrencyID          int64   `json:"currencyID" validate:"required,gte=0"`
	VacationDaysPerYear uint16  `json:"vacationDaysPerYear" validate:"gte=0"`
	FromDate            string  `json:"fromDate" validate:"required"`
	ToDate              *string `json:"toDate" validate:"omitempty,fromDateGTEToDate"`
}

type UpdateEmployeeHistory struct {
	HoursPerMonth       *uint16 `json:"hoursPerMonth" validate:"omitempty,gte=0"`
	SalaryPerMonth      *uint64 `json:"salaryPerMonth" validate:"omitempty,gte=0"`
	CurrencyID          *int64  `json:"currencyID" validate:"omitempty,gte=0"`
	VacationDaysPerYear *uint16 `json:"vacationDaysPerYear" validate:"omitempty,gte=0"`
	FromDate            *string `json:"fromDate" validate:"omitempty"`
	ToDate              *string `json:"toDate" validate:"omitempty,fromDateGTEToDate"`
}
