package models

type EmployeeHistoryCostDetail struct {
	ID      int64  `db:"id" json:"id"`
	Month   string `db:"month" json:"month"`
	Amount  uint64 `db:"amount" json:"amount"`
	Divider uint   `db:"Divider" json:"Divider"`
	CostID  int64  `db:"cost_id" json:"costID"`
}

type CreateEmployeeHistoryCostDetail struct {
	Month   string `db:"month" json:"month"`
	Amount  uint64 `db:"amount" json:"amount"`
	Divider uint   `db:"divider" json:"divider"`
	CostID  int64  `db:"cost_id" json:"costID"`
}
