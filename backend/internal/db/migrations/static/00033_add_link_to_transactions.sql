-- +goose Up
-- +goose StatementBegin
ALTER TABLE transactions ADD COLUMN link VARCHAR(2048) NULL AFTER name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transactions DROP COLUMN link;
-- +goose StatementEnd
