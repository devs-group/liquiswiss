-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_organisation_settings
    ADD COLUMN IF NOT EXISTS forecast_show_zero_rows BOOL NOT NULL DEFAULT false AFTER forecast_child_details;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_organisation_settings
    DROP COLUMN IF EXISTS forecast_show_zero_rows;
-- +goose StatementEnd
