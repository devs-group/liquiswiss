-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS salary_exclusions (
    id SERIAL PRIMARY KEY,
    exclude_month VARCHAR(7) NOT NULL,

    salary_id BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_Exclusion_Salary FOREIGN KEY (salary_id) REFERENCES salaries(id) ON DELETE CASCADE,

    CONSTRAINT CK_ExcludeMonth_Not_Empty CHECK (exclude_month <> ''),

    CONSTRAINT UQ_ExcludeMonth_SalaryID UNIQUE (exclude_month, salary_id),

    INDEX IDX_ExcludeMonth_SalaryID (exclude_month, salary_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS salary_exclusions;
-- +goose StatementEnd