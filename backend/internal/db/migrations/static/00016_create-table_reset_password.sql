-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS reset_password (
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) NOT NULL,
    code VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT CK_Email_Not_Empty CHECK (email <> ''),

    CONSTRAINT UQ_Email UNIQUE (email)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reset_password;
-- +goose StatementEnd