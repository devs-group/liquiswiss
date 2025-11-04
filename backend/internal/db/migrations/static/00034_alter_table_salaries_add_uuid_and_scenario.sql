-- +goose Up
-- +goose StatementBegin
ALTER TABLE salaries ADD COLUMN uuid CHAR(36) NULL;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE salaries SET uuid = UUID() WHERE uuid IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salaries MODIFY COLUMN uuid CHAR(36) NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salaries ADD COLUMN scenario_id BIGINT UNSIGNED NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salaries ADD CONSTRAINT FK_Salary_Scenario
    FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE RESTRICT ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_salaries_scenario ON salaries(scenario_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_salaries_uuid ON salaries(uuid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE salaries DROP FOREIGN KEY FK_Salary_Scenario;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salaries DROP INDEX idx_salaries_scenario;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salaries DROP INDEX idx_salaries_uuid;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salaries DROP COLUMN scenario_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salaries DROP COLUMN uuid;
-- +goose StatementEnd
