-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS scenarios (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    is_default BOOL DEFAULT false,
    deleted BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    parent_scenario_id BIGINT UNSIGNED,
    organisation_id BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_Scenario_Parent FOREIGN KEY (parent_scenario_id) REFERENCES scenarios (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT FK_Scenario_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS scenarios;
-- +goose StatementEnd
