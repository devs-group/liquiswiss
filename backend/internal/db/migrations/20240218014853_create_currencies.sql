-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS go_currencies (
    id SERIAL PRIMARY KEY,
    code VARCHAR(3) UNIQUE NOT NULL,
    description VARCHAR(30),
    locale_code VARCHAR(5) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS go_currencies;
-- +goose StatementEnd