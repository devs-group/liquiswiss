-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS vats (
    id SERIAL PRIMARY KEY,
    value BIGINT NOT NULL,
    -- Owner empty = Coming from system for all users
    owner BIGINT UNSIGNED,
    organisation BIGINT UNSIGNED,

    CONSTRAINT FK_Vat_Owner FOREIGN KEY (owner) REFERENCES go_users (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_Vat_Organisation FOREIGN KEY (organisation) REFERENCES go_organisations (id) ON DELETE CASCADE ON UPDATE CASCADE,

    CONSTRAINT CHK_Value_Required CHECK (
       vats.value > 0
    )
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS vats;
-- +goose StatementEnd