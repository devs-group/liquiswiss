-- +goose Up
-- +goose StatementBegin
ALTER TABLE go_transactions
    ADD COLUMN vat BIGINT UNSIGNED,
    ADD CONSTRAINT FK_Transaction_Vat FOREIGN KEY (vat) REFERENCES vats (id) ON DELETE SET NULL ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE go_transactions
    DROP COLUMN vat,
    DROP CONSTRAINT FK_Transaction_Vat;
-- +goose StatementEnd
