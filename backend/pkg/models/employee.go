package models

import "liquiswiss/pkg/types"

type Employee struct {
	ID                  int64         `db:"id" json:"id"`
	Name                string        `db:"name" json:"name"`
	HoursPerMonth       uint16        `db:"hours_per_month" json:"hoursPerMonth"`
	VacationDaysPerYear uint16        `db:"vacation_days_per_year" json:"vacationDaysPerYear"`
	EntryDate           types.AsDate  `db:"entry_date" json:"entryDate"`
	ExitDate            *types.AsDate `db:"exit_date" json:"exitDate"`
}

type CreateEmployee struct {
	Name                string  `json:"name" validate:"required,max=100"`
	HoursPerMonth       uint16  `json:"hoursPerMonth" validate:"required,gt=0"`
	VacationDaysPerYear uint16  `json:"vacationDaysPerYear" validate:"required,gt=0"`
	EntryDate           string  `json:"entryDate" validate:"required"`
	ExitDate            *string `json:"exitDate" validate:"omitempty,exitDateGTEEntryDate"`
}

type UpdateEmployee struct {
	Name                *string `json:"name" validate:"omitempty,max=100"`
	HoursPerMonth       *uint16 `json:"hoursPerMonth" validate:"omitempty,gt=0"`
	VacationDaysPerYear *uint16 `json:"vacationDaysPerYear" validate:"omitempty,gt=0"`
	EntryDate           *string `json:"entryDate" validate:"omitempty"`
	ExitDate            *string `json:"exitDate" validate:"omitempty,exitDateGTEEntryDate"`
}
