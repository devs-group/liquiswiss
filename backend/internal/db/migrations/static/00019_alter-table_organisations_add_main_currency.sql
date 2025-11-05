-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS organisations
ADD COLUMN IF NOT EXISTS main_currency_id BIGINT UNSIGNED AFTER name,
ADD CONSTRAINT FK_Organisation_Currency FOREIGN KEY (main_currency_id) REFERENCES currencies (id) ON DELETE RESTRICT ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS organisations
DROP CONSTRAINT IF EXISTS FK_Organisation_Currency,
DROP COLUMN IF EXISTS main_currency_id;
-- +goose StatementEnd