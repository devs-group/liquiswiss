-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS employee_history_cost_exclusions (
    id SERIAL PRIMARY KEY,
    exclude_month VARCHAR(7) NOT NULL,

    label_id BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_Exclusion_EmployeeHistoryCost FOREIGN KEY (label_id) REFERENCES employee_history_cost_labels(id) ON DELETE CASCADE,

    CONSTRAINT CK_ExcludeMonth_Not_Empty CHECK (exclude_month <> ''),

    CONSTRAINT UQ_ExcludeMonth_EmployeeHistoryCostLabelID UNIQUE (exclude_month, label_id),

    INDEX IDX_ExcludeMonth_EmployeeHistoryCostLabelID (exclude_month, label_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS employee_history_cost_exclusions;
-- +goose StatementEnd