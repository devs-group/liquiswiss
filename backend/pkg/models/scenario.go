package models

import "time"

type Scenario struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	IsBase    bool      `db:"is_base" json:"isBase"`
	ParentID  *int64    `db:"parent_id" json:"parentId"`
	IsCurrent bool      `json:"isCurrent"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type CreateScenario struct {
	Name string `json:"name" validate:"required,max=100"`
	// Scenario to copy from, defaults to the current one
	SourceScenarioID *int64 `json:"sourceScenarioId" validate:"omitempty"`
}

type UpdateScenario struct {
	Name *string `json:"name" validate:"omitempty,max=100"`
}
