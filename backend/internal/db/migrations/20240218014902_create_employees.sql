-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS go_employees (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    owner BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_Employee_Owner FOREIGN KEY (owner) REFERENCES go_users (id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS go_employees;
-- +goose StatementEnd