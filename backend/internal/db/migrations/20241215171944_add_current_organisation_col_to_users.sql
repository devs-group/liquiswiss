-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD COLUMN current_organisation BIGINT UNSIGNED DEFAULT NULL,
ADD CONSTRAINT FK_Current_Organisation FOREIGN KEY (current_organisation) REFERENCES organisations (id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
DROP COLUMN current_organisation,
DROP CONSTRAINT FK_Current_Organisation;
-- +goose StatementEnd
