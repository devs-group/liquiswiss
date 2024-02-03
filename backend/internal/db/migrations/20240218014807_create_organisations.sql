-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS go_organisations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL CHECK (name <> ''),
    member_count INT UNSIGNED NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS go_organisations;
-- +goose StatementEnd