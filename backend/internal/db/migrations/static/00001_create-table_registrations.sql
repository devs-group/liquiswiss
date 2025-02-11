-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS registrations (
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) NOT NULL CHECK (email <> ''),
    code VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT CK_Email_Not_Empty CHECK (email <> '')
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS registrations;
-- +goose StatementEnd