package models

import (
	"liquiswiss/pkg/types"
)

type Transaction struct {
	ID        int64         `db:"id" json:"id"`
	Name      string        `db:"name" json:"name"`
	Amount    int64         `db:"amount" json:"amount"`
	Cycle     *string       `db:"cycle" json:"cycle"`
	Type      string        `db:"type" json:"type"`
	StartDate types.AsDate  `db:"start_date" json:"startDate"`
	EndDate   *types.AsDate `db:"end_date" json:"endDate"`
	Category  Category      `json:"category"`
	Currency  Currency      `json:"currency"`
}

type CreateTransaction struct {
	Name      string  `json:"name" validate:"required,max=255"`
	Amount    int64   `json:"amount" validate:"required"`
	Cycle     *string `json:"cycle" validate:"omitempty,allowedCycles"`
	Type      string  `json:"type" validate:"required,oneof='single' 'repeating',cycleRequiredIfRepeating"`
	StartDate string  `json:"startDate" validate:"required"`
	EndDate   *string `json:"endDate" validate:"omitempty,endDateGTEStartDate"`
	Category  int64   `json:"category" validate:"required"`
	Currency  int64   `json:"currency" validate:"required"`
}

type UpdateTransaction struct {
	Name      *string `json:"name" validate:"omitempty,max=255"`
	Amount    *int64  `json:"amount" validate:"omitempty"`
	Cycle     *string `json:"cycle" validate:"omitempty,allowedCycles"`
	Type      *string `json:"type" validate:"omitempty,oneof='single' 'repeating',cycleRequiredIfRepeating"`
	StartDate *string `json:"startDate" validate:"omitempty"`
	EndDate   *string `json:"endDate" validate:"omitempty,endDateGTEStartDate"`
	Category  *int64  `json:"category" validate:"omitempty"`
	Currency  *int64  `json:"currency" validate:"omitempty"`
}