package models

type User struct {
	ID    int64  `db:"id" json:"id"`
	Name  string `db:"name" json:"name" validate:"max=100"`
	Email string `db:"email" json:"email" validate:"required,email"`
}
