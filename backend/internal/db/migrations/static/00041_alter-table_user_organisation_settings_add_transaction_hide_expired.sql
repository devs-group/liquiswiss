-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_organisation_settings
    ADD COLUMN IF NOT EXISTS transaction_hide_expired BOOL NOT NULL DEFAULT false AFTER transaction_hide_disabled;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_organisation_settings
    DROP COLUMN IF EXISTS transaction_hide_expired;
-- +goose StatementEnd
