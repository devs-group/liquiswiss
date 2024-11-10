package models

import (
	"liquiswiss/pkg/types"
)

type Employee struct {
	ID                  int64         `db:"id" json:"id"`
	Name                string        `db:"name" json:"name"`
	HoursPerMonth       *uint16       `db:"-" json:"hoursPerMonth"`
	SalaryPerMonth      *uint64       `db:"-" json:"salaryPerMonth"`
	SalaryCurrency      *Currency     `db:"-" json:"salaryCurrency"`
	VacationDaysPerYear *uint16       `db:"-" json:"vacationDaysPerYear"`
	FromDate            *types.AsDate `db:"-" json:"fromDate"`
	ToDate              *types.AsDate `db:"-" json:"toDate"`
	IsInFuture          bool          `db:"-" json:"isInFuture"`
}

type CreateEmployee struct {
	Name string `json:"name" validate:"required,max=100"`
}

type UpdateEmployee struct {
	Name *string `json:"name" validate:"omitempty,max=100"`
}
