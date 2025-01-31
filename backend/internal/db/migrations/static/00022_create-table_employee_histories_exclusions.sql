-- +goose Up
-- +goose StatementBegin
CREATE TABLE employee_histories_exclusions (
    id SERIAL PRIMARY KEY,
    exclude_month VARCHAR(7) NOT NULL,

    employee_history_id BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_Exclusion_EmployeeHistory FOREIGN KEY (employee_history_id) REFERENCES employee_histories(id) ON DELETE CASCADE,

    CONSTRAINT CK_ExcludeMonth_Not_Empty CHECK (exclude_month <> ''),

    CONSTRAINT UQ_ExcludeMonth_EmployeeHistoryID UNIQUE (exclude_month, employee_history_id),

    INDEX IDX_ExcludeMonth_EmployeeHistoryID (exclude_month, employee_history_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS employee_histories_exclusions;
-- +goose StatementEnd