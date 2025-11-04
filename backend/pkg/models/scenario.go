package models

import "time"

type ScenarioType string

const (
	ScenarioTypeHorizontal ScenarioType = "horizontal"
	ScenarioTypeVertical   ScenarioType = "vertical"
)

type Scenario struct {
	ID               int64        `db:"id" json:"id"`
	Name             string       `db:"name" json:"name"`
	Type             ScenarioType `db:"type" json:"type"`
	IsDefault        bool         `db:"is_default" json:"isDefault"`
	ParentScenarioID *int64       `db:"parent_scenario_id" json:"parentScenarioId"`
	OrganisationID   int64        `db:"organisation_id" json:"organisationId"`
	CreatedAt        time.Time    `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time    `db:"updated_at" json:"updatedAt"`
}

type CreateScenario struct {
	Name             string       `json:"name" validate:"required,max=100"`
	Type             ScenarioType `json:"type" validate:"required,oneof=horizontal vertical"`
	ParentScenarioID *int64       `json:"parentScenarioId" validate:"required_if=Type vertical"`
}

type UpdateScenario struct {
	Name *string `json:"name" validate:"omitempty,max=100"`
}

type ScenarioListItem struct {
	ID               int64        `db:"id" json:"id"`
	Name             string       `db:"name" json:"name"`
	Type             ScenarioType `db:"type" json:"type"`
	IsDefault        bool         `db:"is_default" json:"isDefault"`
	ParentScenarioID *int64       `db:"parent_scenario_id" json:"parentScenarioId"`
	CreatedAt        time.Time    `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time    `db:"updated_at" json:"updatedAt"`
}
