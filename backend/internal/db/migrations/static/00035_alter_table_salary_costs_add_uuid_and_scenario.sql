-- +goose Up
-- +goose StatementBegin
ALTER TABLE salary_costs ADD COLUMN uuid CHAR(36) NULL;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE salary_costs SET uuid = UUID() WHERE uuid IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salary_costs MODIFY COLUMN uuid CHAR(36) NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salary_costs ADD COLUMN scenario_id BIGINT UNSIGNED NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salary_costs ADD CONSTRAINT FK_SalaryCost_Scenario
    FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE RESTRICT ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_salary_costs_scenario ON salary_costs(scenario_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_salary_costs_uuid ON salary_costs(uuid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE salary_costs DROP FOREIGN KEY FK_SalaryCost_Scenario;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salary_costs DROP INDEX idx_salary_costs_scenario;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salary_costs DROP INDEX idx_salary_costs_uuid;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salary_costs DROP COLUMN scenario_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salary_costs DROP COLUMN uuid;
-- +goose StatementEnd
