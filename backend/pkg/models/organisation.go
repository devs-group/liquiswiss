package models

type Organisation struct {
	ID          int64  `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	MemberCount int64  `db:"member_count" json:"memberCount"`
	Role        string `db:"role" json:"role"`
	IsDefault   bool   `db:"is_default" json:"isDefault"`
}

type CreateOrganisation struct {
	Name string `json:"name" validate:"required,min=3,max=100"`
}

type UpdateOrganisation struct {
	Name *string `json:"name" validate:"omitempty,min=3,max=100"`
}
