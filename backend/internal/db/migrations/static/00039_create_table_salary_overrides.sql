-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS salary_overrides (
    id SERIAL PRIMARY KEY,
    uuid CHAR(36) NOT NULL,
    scenario_id BIGINT UNSIGNED NOT NULL,
    employee_id BIGINT UNSIGNED NULL,
    cycle ENUM('monthly', 'quarterly', 'biannually', 'yearly') NULL,
    hours_per_month SMALLINT UNSIGNED NULL,
    amount BIGINT UNSIGNED NULL,
    vacation_days_per_year SMALLINT UNSIGNED NULL,
    from_date DATE NULL,
    to_date DATE NULL,
    is_termination BOOL NULL,
    is_disabled BOOL NULL,
    deleted BOOL NULL,
    currency_id BIGINT UNSIGNED NULL,
    organisation_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT FK_SalaryOverride_Scenario FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_SalaryOverride_Employee FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_SalaryOverride_Currency FOREIGN KEY (currency_id) REFERENCES currencies (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT FK_SalaryOverride_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE ON UPDATE CASCADE,

    UNIQUE KEY UQ_SalaryOverride_UUID_Scenario (uuid, scenario_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_salary_overrides_scenario ON salary_overrides(scenario_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_salary_overrides_uuid ON salary_overrides(uuid);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_salary_overrides_employee ON salary_overrides(employee_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS salary_overrides;
-- +goose StatementEnd
