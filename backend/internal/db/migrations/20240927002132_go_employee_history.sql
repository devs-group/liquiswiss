-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS go_employee_history (
    id SERIAL PRIMARY KEY,
    employee_id BIGINT UNSIGNED NOT NULL,
    hours_per_month SMALLINT UNSIGNED NOT NULL,
    salary_per_month BIGINT UNSIGNED NOT NULL,
    salary_currency BIGINT UNSIGNED NOT NULL,
    vacation_days_per_year SMALLINT UNSIGNED NOT NULL,
    from_date DATE NOT NULL,
    to_date DATE,

    CONSTRAINT FK_History_Employee FOREIGN KEY (employee_id) REFERENCES go_employees (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT CK_History_Dates CHECK (from_date <= to_date OR to_date IS NULL),
    CONSTRAINT FK_History_SalaryCurrency FOREIGN KEY (salary_currency) REFERENCES go_currencies (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT UQ_Employee_From_Date UNIQUE (employee_id, from_date)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS go_employee_history;
-- +goose StatementEnd
