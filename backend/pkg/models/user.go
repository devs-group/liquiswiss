package models

type User struct {
	ID    int64  `db:"id" json:"id"`
	Name  string `db:"name" json:"name" validate:"max=100"`
	Email string `db:"email" json:"email" validate:"required,email"`
}

type UpdateUser struct {
	Name  *string `json:"name" validate:"omitempty,max=100"`
	Email *string `json:"email" validate:"omitempty,email"`
}

type UpdateUserPassword struct {
	Password string `json:"password" validate:"required,min=8"`
}
