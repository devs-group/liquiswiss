-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS go_employees (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    hours_per_month SMALLINT UNSIGNED NOT NULL,
    vacation_days_per_year SMALLINT UNSIGNED NOT NULL,
    entry_date DATE,
    exit_date DATE,
    owner BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_Employee_Owner FOREIGN KEY (owner) REFERENCES go_users (id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS go_employees;
-- +goose StatementEnd