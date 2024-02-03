-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS go_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS go_categories;
-- +goose StatementEnd