package models

type Login struct {
	ID       int64  `json:"-"`
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required"`
}

type CreateRegistration struct {
	Email string `json:"email" validate:"required,email,max=100"`
}

type ForgotPassword struct {
	Email string `json:"email" validate:"required,email,max=100"`
}

type CheckResetPasswordCode struct {
	Email string `json:"email" validate:"required,email,max=100"`
	Code  string `json:"code" validate:"required"`
}

type ResetPassword struct {
	Email    string `json:"email" validate:"required,email,max=100"`
	Code     string `json:"code" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

type CheckRegistrationCode struct {
	Email string `json:"email" validate:"required,email,max=100"`
	Code  string `json:"code" validate:"required"`
}

type FinishRegistration struct {
	Email    string `json:"email" validate:"required,email,max=100"`
	Code     string `json:"code" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}
