package models

type Category struct {
	ID      int64  `db:"id" json:"id"`
	Name    string `db:"name" json:"name"`
	CanEdit bool   `db:"can_edit" json:"canEdit"`
}

type CreateCategory struct {
	Name string `json:"name" validate:"required,max=100"`
}

type UpdateCategory struct {
	Name *string `json:"name" validate:"omitempty,max=100"`
}
