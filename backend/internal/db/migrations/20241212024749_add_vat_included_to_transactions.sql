-- +goose Up
-- +goose StatementBegin
ALTER TABLE transactions
    ADD COLUMN vat_included BOOL NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transactions
    DROP COLUMN vat_included;
-- +goose StatementEnd
