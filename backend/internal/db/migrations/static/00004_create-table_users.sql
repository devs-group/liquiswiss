-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(100) DEFAULT '',
    deleted BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    current_organisation_id BIGINT UNSIGNED,

    CONSTRAINT FK_Current_Organisation FOREIGN KEY (current_organisation_id) REFERENCES organisations (id) ON DELETE SET NULL,

    CONSTRAINT CK_Email_Not_Empty CHECK (email <> ''),
    CONSTRAINT CK_Password_Not_Empty CHECK (password <> ''),

    CONSTRAINT UQ_Email UNIQUE (email)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd