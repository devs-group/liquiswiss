-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS employee_overrides (
    id SERIAL PRIMARY KEY,
    uuid CHAR(36) NOT NULL,
    scenario_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NULL,
    deleted BOOL NULL,
    organisation_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT FK_EmployeeOverride_Scenario FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_EmployeeOverride_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE ON UPDATE CASCADE,

    UNIQUE KEY UQ_EmployeeOverride_UUID_Scenario (uuid, scenario_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_employee_overrides_scenario ON employee_overrides(scenario_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_employee_overrides_uuid ON employee_overrides(uuid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS employee_overrides;
-- +goose StatementEnd
