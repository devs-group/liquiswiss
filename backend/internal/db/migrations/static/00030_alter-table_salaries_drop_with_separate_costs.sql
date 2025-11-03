-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS salaries
    DROP COLUMN IF EXISTS with_separate_costs;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS salaries
    ADD COLUMN IF NOT EXISTS with_separate_costs BOOL DEFAULT true AFTER to_date;
-- +goose StatementEnd
