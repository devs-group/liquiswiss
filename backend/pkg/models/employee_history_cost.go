package models

import "liquiswiss/pkg/types"

type EmployeeHistoryCost struct {
	ID                int64                     `db:"id" json:"id"`
	Label             *EmployeeHistoryCostLabel `db:"-" json:"label"`
	Cycle             string                    `db:"cycle" json:"cycle"`
	AmountType        string                    `db:"amount_type" json:"amountType"`
	Amount            uint64                    `db:"amount" json:"amount"`
	DistributionType  string                    `db:"distribution_type" json:"distributionType"`
	RelativeOffset    int64                     `db:"relative_offset" json:"relativeOffset"`
	TargetDate        *types.AsDate             `db:"target_date" json:"targetDate"`
	EmployeeHistoryID int64                     `db:"employee_history_id" json:"employeeHistoryID"`

	// Hidden values
	HistoryCycle    string        `db:"history_cycle" json:"-"`
	HistorySalary   uint64        `db:"history_salary" json:"-"`
	HistoryFromDate types.AsDate  `db:"history_from_date" json:"-"`
	HistoryToDate   *types.AsDate `db:"history_to_date" json:"-"`
	DBDate          types.AsDate  `db:"db_date" json:"-"`

	// Calculated values
	CalculatedAmount                uint64                      `db:"-" json:"calculatedAmount"`
	CalculatedPreviousExecutionDate *types.AsDate               `db:"-" json:"calculatedPreviousExecutionDate"`
	CalculatedNextExecutionDate     *types.AsDate               `db:"-" json:"calculatedNextExecutionDate"`
	CalculatedNextCost              uint64                      `db:"-" json:"calculatedNextCost"`
	CalculatedCostDetails           []EmployeeHistoryCostDetail `db:"-" json:"calculatedCostDetails"`
}

type CreateEmployeeHistoryCost struct {
	Cycle            string  `db:"cycle" json:"cycle" validate:"allowedCostCycles"`
	AmountType       string  `db:"amount_type" json:"amountType" validate:"allowedCostAmountTypes"`
	Amount           uint64  `db:"amount" json:"amount" validate:"gte=0"`
	DistributionType string  `db:"distribution_type" json:"distributionType"`
	RelativeOffset   int64   `db:"relative_offset" json:"relativeOffset" validate:"gt=0"`
	TargetDate       *string `db:"target_date" json:"targetDate"`
	LabelID          *int64  `db:"label_id" json:"labelID"`
}

type CopyEmployeeHistoryCosts struct {
	IDs []int64 `db:"-" json:"ids" validate:"required"`
}
