-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS vats (
    id SERIAL PRIMARY KEY,
    value BIGINT NOT NULL,
    deleted BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- Organisation ID empty = Coming from system for all users
    organisation_id BIGINT UNSIGNED,

    CONSTRAINT FK_Vat_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE ON UPDATE CASCADE,

    CONSTRAINT CHK_Value_Required CHECK (
       vats.value >= 0
    )
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS vats;
-- +goose StatementEnd