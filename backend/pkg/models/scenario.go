package models

import "liquiswiss/pkg/types"

type Scenario struct {
	ID               int64        `db:"id" json:"id"`
	Name             string       `db:"name" json:"name"`
	IsDefault        bool         `db:"is_default" json:"isDefault"`
	CreatedAt        types.AsDate `db:"created_at" json:"createdAt"`
	ParentScenarioID *int64       `db:"parent_scenario_id" json:"parentScenarioID"`
}

type CreateScenario struct {
	Name             string `json:"name" validate:"required,max=100"`
	ParentScenarioID *int64 `json:"parentScenarioID" validate:"omitempty"`
}

type UpdateScenario struct {
	Name             *string `json:"name" validate:"omitempty,max=100"`
	ParentScenarioID *int64  `json:"parentScenarioID" validate:"omitempty"`
}
