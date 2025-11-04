-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS salary_cost_overrides (
    id SERIAL PRIMARY KEY,
    uuid CHAR(36) NOT NULL,
    scenario_id BIGINT UNSIGNED NOT NULL,
    salary_id BIGINT UNSIGNED NULL,
    cycle ENUM('once', 'monthly', 'quarterly', 'biannually', 'yearly') NULL,
    amount_type ENUM('fixed', 'percentage') NULL,
    amount BIGINT UNSIGNED NULL,
    distribution_type ENUM('employer', 'employee', 'both') NULL,
    relative_offset BIGINT NULL,
    target_date DATE NULL,
    label_id BIGINT UNSIGNED NULL,
    organisation_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT FK_SalaryCostOverride_Scenario FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_SalaryCostOverride_Salary FOREIGN KEY (salary_id) REFERENCES salaries (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_SalaryCostOverride_Label FOREIGN KEY (label_id) REFERENCES salary_cost_labels (id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT FK_SalaryCostOverride_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE ON UPDATE CASCADE,

    UNIQUE KEY UQ_SalaryCostOverride_UUID_Scenario (uuid, scenario_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_salary_cost_overrides_scenario ON salary_cost_overrides(scenario_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_salary_cost_overrides_uuid ON salary_cost_overrides(uuid);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_salary_cost_overrides_salary ON salary_cost_overrides(salary_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS salary_cost_overrides;
-- +goose StatementEnd
