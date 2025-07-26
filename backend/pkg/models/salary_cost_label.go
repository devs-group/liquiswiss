package models

type SalaryCostLabel struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type CreateSalaryCostLabel struct {
	Name string `json:"name" validate:"required,max=255"`
}

type UpdateSalaryCostLabel struct {
	Name *string `json:"name" validate:"required,max=255"`
}
