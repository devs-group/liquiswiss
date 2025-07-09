-- +goose Up
-- +goose StatementBegin
INSERT INTO currencies (id, code, description, locale_code, deleted, created_at)
VALUES (1, 'CHF', 'Franken (Schweiz)', 'de-CH', 0, NOW())
ON DUPLICATE KEY UPDATE id = id;
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO categories (id, name, deleted, created_at, organisation_id)
VALUES (1, 'Generell', 0, NOW(), NULL)
ON DUPLICATE KEY UPDATE id = id;
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO vats (id, value, deleted, created_at, organisation_id)
VALUES (1, 810, 0, NOW(), NULL)
ON DUPLICATE KEY UPDATE id = id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- NO DOWN MIGRATION NEEDED
-- +goose StatementEnd
