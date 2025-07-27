-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS salaries
    ADD COLUMN is_termination BOOL DEFAULT false AFTER with_separate_costs;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS salaries
    DROP COLUMN IF EXISTS is_termination;
-- +goose StatementEnd
