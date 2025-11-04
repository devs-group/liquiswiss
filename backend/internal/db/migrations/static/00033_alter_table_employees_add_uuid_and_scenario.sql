-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees ADD COLUMN uuid CHAR(36) NULL;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE employees SET uuid = UUID() WHERE uuid IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE employees MODIFY COLUMN uuid CHAR(36) NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE employees ADD COLUMN scenario_id BIGINT UNSIGNED NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE employees ADD CONSTRAINT FK_Employee_Scenario
    FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE RESTRICT ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_employees_scenario ON employees(scenario_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_employees_uuid ON employees(uuid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees DROP FOREIGN KEY FK_Employee_Scenario;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE employees DROP INDEX idx_employees_scenario;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE employees DROP INDEX idx_employees_uuid;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE employees DROP COLUMN scenario_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE employees DROP COLUMN uuid;
-- +goose StatementEnd
