package models

import "time"

type Scenario struct {
	ID               int64     `db:"id" json:"id" validate:"gt=0"`
	Name             string    `db:"name" json:"name" validate:"required,max=100"`
	IsDefault        bool      `db:"is_default" json:"isDefault"`
	CreatedAt        time.Time `db:"created_at" json:"createdAt"`
	ParentScenarioID *int64    `db:"parent_scenario_id" json:"parentScenarioId,omitempty"`
	OrganisationID   int64     `db:"organisation_id" json:"organisationId" validate:"gt=0"`
}

type CreateScenario struct {
	Name             string `json:"name" validate:"required,max=100"`
	IsDefault        bool   `json:"isDefault"`
	ParentScenarioID *int64 `json:"parentScenarioId" validate:"omitempty,gt=0"`
}

type UpdateScenario struct {
	Name             *string `json:"name" validate:"omitempty,max=100"`
	IsDefault        *bool   `json:"isDefault"`
	ParentScenarioID *int64  `json:"parentScenarioId" validate:"omitempty,gt=0"`
}
