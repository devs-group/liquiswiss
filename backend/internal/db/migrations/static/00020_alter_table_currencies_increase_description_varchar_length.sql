-- +goose Up
-- +goose StatementBegin
ALTER TABLE currencies
MODIFY description VARCHAR(60) null;
-- +goose StatementEnd
