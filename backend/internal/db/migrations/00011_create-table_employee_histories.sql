-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS employee_histories (
    id SERIAL PRIMARY KEY,
    cycle ENUM('daily', 'weekly', 'monthly', 'quarterly', 'biannually', 'yearly') NOT NULL DEFAULT 'monthly',
    hours_per_month SMALLINT UNSIGNED NOT NULL,
    salary BIGINT UNSIGNED NOT NULL,
    vacation_days_per_year SMALLINT UNSIGNED NOT NULL,
    from_date DATE NOT NULL,
    to_date DATE,
    with_separate_costs BOOL DEFAULT true,
    deleted BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    employee_id BIGINT UNSIGNED NOT NULL,
    currency_id BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_History_Employee FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_History_Currency FOREIGN KEY (currency_id) REFERENCES currencies (id) ON DELETE RESTRICT ON UPDATE CASCADE,

    CONSTRAINT UQ_Employee_From_Date UNIQUE (employee_id, from_date),
    CONSTRAINT CK_History_Dates CHECK (from_date <= to_date OR to_date IS NULL)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS employee_histories;
-- +goose StatementEnd