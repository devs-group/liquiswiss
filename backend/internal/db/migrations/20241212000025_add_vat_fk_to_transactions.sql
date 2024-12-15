-- +goose Up
-- +goose StatementBegin
ALTER TABLE transactions
    ADD COLUMN vat_id BIGINT UNSIGNED,
    ADD CONSTRAINT FK_Transaction_Vat FOREIGN KEY (vat_id) REFERENCES vats (id) ON DELETE SET NULL ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transactions
    DROP COLUMN vat_id,
    DROP CONSTRAINT FK_Transaction_Vat;
-- +goose StatementEnd
