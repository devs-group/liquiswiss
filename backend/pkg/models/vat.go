package models

type Vat struct {
	ID             int64  `db:"id" json:"id"`
	Value          int64  `db:"value" json:"value"`
	FormattedValue string `db:"formatted_value" json:"formattedValue"`
	CanEdit        bool   `db:"can_edit" json:"canEdit"`
}

type CreateVat struct {
	Value int64 `json:"value" validate:"required,min=1"`
}

type UpdateVat struct {
	Value *int64 `json:"value" validate:"omitempty,min=1"`
}
