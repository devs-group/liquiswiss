package models

import "liquiswiss/pkg/types"

type EmployeeHistoryCost struct {
	ID                    int64                     `db:"id" json:"id"`
	Label                 *EmployeeHistoryCostLabel `db:"-" json:"label"`
	Cycle                 string                    `db:"cycle" json:"cycle"`
	AmountType            string                    `db:"amount_type" json:"amountType"`
	Amount                uint64                    `db:"amount" json:"amount"`
	DistributionType      string                    `db:"distribution_type" json:"distributionType"`
	CalculatedAmount      uint64                    `db:"calculated_amount" json:"calculatedAmount"`
	RelativeOffset        int64                     `db:"relative_offset" json:"relativeOffset"`
	TargetDate            *types.AsDate             `db:"target_date" json:"targetDate"`
	PreviousExecutionDate *types.AsDate             `db:"previous_execution_date" json:"previousExecutionDate"`
	NextExecutionDate     *types.AsDate             `db:"next_execution_date" json:"nextExecutionDate"`
	NextCost              uint64                    `db:"next_cost" json:"nextCost"`
	EmployeeHistoryID     int64                     `db:"employee_history_id" json:"employeeHistoryID"`
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
