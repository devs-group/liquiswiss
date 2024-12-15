-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS currencies (
    id SERIAL PRIMARY KEY,
    code VARCHAR(3) UNIQUE NOT NULL,
    description VARCHAR(30),
    locale_code VARCHAR(5) NOT NULL,
    deleted BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS currencies;
-- +goose StatementEnd