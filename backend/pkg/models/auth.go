package models

type Login struct {
	ID       int64  `json:"-"`
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required"`
}

type Registration struct {
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,min=8"`
}
