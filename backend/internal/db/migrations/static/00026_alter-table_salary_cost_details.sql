-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS salary_cost_details
    ADD COLUMN is_extra_month BOOL DEFAULT false AFTER divider;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS salary_cost_details
    DROP COLUMN IF EXISTS is_extra_month;
-- +goose StatementEnd
