package models

type EmployeeHistoryCostLabel struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type CreateEmployeeHistoryCostLabel struct {
	Name string `json:"name" validate:"required,max=255"`
}

type UpdateEmployeeHistoryCostLabel struct {
	Name *string `json:"name" validate:"required,max=255"`
}
