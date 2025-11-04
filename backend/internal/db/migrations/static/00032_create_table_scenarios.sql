-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS scenarios (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type ENUM('horizontal', 'vertical') NOT NULL,
    is_default BOOL NOT NULL DEFAULT false,
    parent_scenario_id BIGINT UNSIGNED,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    organisation_id BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_Scenario_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_Scenario_Parent FOREIGN KEY (parent_scenario_id) REFERENCES scenarios (id) ON DELETE RESTRICT ON UPDATE CASCADE,

    CONSTRAINT UQ_Organisation_Name UNIQUE (organisation_id, name)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_scenarios_parent ON scenarios(parent_scenario_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_scenarios_organisation ON scenarios(organisation_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS scenarios;
-- +goose StatementEnd
