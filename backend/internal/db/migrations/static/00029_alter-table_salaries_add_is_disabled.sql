-- +goose Up
-- +goose StatementBegin
ALTER TABLE salaries
    ADD COLUMN is_disabled TINYINT(1) NOT NULL DEFAULT 0 AFTER is_termination;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS salaries
    DROP COLUMN IF EXISTS is_disabled;
-- +goose StatementEnd
