-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS go_users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) UNIQUE NOT NULL CHECK (email <> ''),
    password VARCHAR(255) NOT NULL CHECK (password <> ''),
    name VARCHAR(100) DEFAULT ''
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS go_users;
-- +goose StatementEnd