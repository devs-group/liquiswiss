package models

import "liquiswiss/pkg/types"

type SalaryCost struct {
	ID                int64            `db:"id" json:"id"`
	Label             *SalaryCostLabel `db:"-" json:"label"`
	Cycle             string           `db:"cycle" json:"cycle"`
	AmountType        string           `db:"amount_type" json:"amountType"`
	Amount            uint64           `db:"amount" json:"amount"`
	DistributionType  string           `db:"distribution_type" json:"distributionType" validate:"allowedCostDistributionTypes"`
	RelativeOffset    int64            `db:"relative_offset" json:"relativeOffset"`
	TargetDate        *types.AsDate    `db:"target_date" json:"targetDate"`
	BaseSalaryCostIDs []int64          `db:"-" json:"baseSalaryCostIDs"`
	SalaryID          int64            `db:"salary_id" json:"salaryID"`

	// Hidden values
	SalaryCycle    string        `db:"salary_cycle" json:"-"`
	SalaryAmount   uint64        `db:"salary_amount" json:"-"`
	SalaryFromDate types.AsDate  `db:"salary_from_date" json:"-"`
	SalaryToDate   *types.AsDate `db:"salary_to_date" json:"-"`
	DBDate         types.AsDate  `db:"db_date" json:"-"`

	// Calculated values
	CalculatedAmount                uint64             `db:"-" json:"calculatedAmount"`
	CalculatedPreviousExecutionDate *types.AsDate      `db:"-" json:"calculatedPreviousExecutionDate"`
	CalculatedNextExecutionDate     *types.AsDate      `db:"-" json:"calculatedNextExecutionDate"`
	CalculatedNextCost              uint64             `db:"-" json:"calculatedNextCost"`
	CalculatedCostDetails           []SalaryCostDetail `db:"-" json:"calculatedCostDetails"`
}

type CreateSalaryCost struct {
	Cycle             string  `db:"cycle" json:"cycle" validate:"allowedCostCycles"`
	AmountType        string  `db:"amount_type" json:"amountType" validate:"allowedCostAmountTypes"`
	Amount            uint64  `db:"amount" json:"amount" validate:"gte=0"`
	DistributionType  string  `db:"distribution_type" json:"distributionType" validate:"allowedCostDistributionTypes"`
	RelativeOffset    int64   `db:"relative_offset" json:"relativeOffset" validate:"gt=0"`
	TargetDate        *string `db:"target_date" json:"targetDate"`
	LabelID           *int64  `db:"label_id" json:"labelID"`
	BaseSalaryCostIDs []int64 `db:"-" json:"baseSalaryCostIDs" validate:"omitempty,dive,gt=0"`
}

type CopySalaryCosts struct {
	IDs            []int64 `db:"-" json:"ids" validate:"omitempty,dive,gt=0"`
	SourceSalaryID *int64  `db:"-" json:"sourceSalaryID" validate:"omitempty,gt=0"`
}
