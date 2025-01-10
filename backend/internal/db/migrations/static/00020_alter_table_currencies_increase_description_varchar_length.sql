-- +goose Up
-- +goose StatementBegin
ALTER TABLE currencies
MODIFY description VARCHAR(60) null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS currencies
MODIFY description VARCHAR(30) null;
-- +goose StatementEnd
