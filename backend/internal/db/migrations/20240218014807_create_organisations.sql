-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS organisations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL CHECK (name <> ''),
    member_count INT UNSIGNED NOT NULL DEFAULT 0,
    deleted BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS organisations;
-- +goose StatementEnd