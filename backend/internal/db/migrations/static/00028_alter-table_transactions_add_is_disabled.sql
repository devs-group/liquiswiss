-- +goose Up
-- +goose StatementBegin
ALTER TABLE transactions
    ADD COLUMN is_disabled TINYINT(1) NOT NULL DEFAULT 0 AFTER vat_included;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS transactions
    DROP COLUMN IF EXISTS is_disabled;
-- +goose StatementEnd
