-- +goose Up
-- +goose StatementBegin
ALTER TABLE go_transactions
    ADD COLUMN vat_included BOOL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE go_transactions
    DROP COLUMN vat_included;
-- +goose StatementEnd
