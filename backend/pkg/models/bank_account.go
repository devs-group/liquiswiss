package models

type BankAccount struct {
	ID       int64    `db:"id" json:"id"`
	Name     string   `db:"name" json:"name"`
	Amount   int64    `db:"amount" json:"amount"`
	Currency Currency `json:"currency"`
}

type CreateBankAccount struct {
	Name     string `json:"name" validate:"required,max=100"`
	Amount   int64  `json:"amount" validate:"required"`
	Currency int64  `json:"currency" validate:"required"`
}

type UpdateBankAccount struct {
	Name     *string `json:"name" validate:"omitempty,max=100"`
	Amount   *int64  `json:"amount" validate:"omitempty"`
	Currency *int64  `json:"currency" validate:"omitempty"`
}
