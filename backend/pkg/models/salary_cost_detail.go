package models

type SalaryCostDetail struct {
	ID           int64  `db:"id" json:"id"`
	Month        string `db:"month" json:"month"`
	Amount       uint64 `db:"amount" json:"amount"`
	Divider      uint   `db:"divider" json:"Divider"`
	IsExtraMonth bool   `db:"is_extra_month" json:"isExtraMonth"`
	CostID       int64  `db:"cost_id" json:"costID"`
}

type CreateSalaryCostDetail struct {
	Month        string `db:"month" json:"month"`
	Amount       uint64 `db:"amount" json:"amount"`
	Divider      uint   `db:"divider" json:"divider"`
	IsExtraMonth bool   `db:"is_extra_month" json:"isExtraMonth"`
	CostID       int64  `db:"cost_id" json:"costID"`
}
