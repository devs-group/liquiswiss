package models

type Forecast struct {
	Month    string `db:"month" json:"month"`
	Revenue  int64  `db:"revenue" json:"revenue"`
	Expense  int64  `db:"expense" json:"expense"`
	Cashflow int64  `db:"cashflow" json:"cashflow"`
}

type CreateForecast struct {
	Month    string `json:"month" validate:"required,max=7"`
	Revenue  int64  `json:"revenue" validate:"required"`
	Expense  int64  `json:"expense" validate:"required"`
	Cashflow int64  `json:"cashflow" validate:"required"`
}
