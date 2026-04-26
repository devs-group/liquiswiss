-- +goose Up
-- +goose StatementBegin
ALTER TABLE organisation_invitations
    ADD COLUMN last_sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER created_at;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE organisation_invitations
    DROP COLUMN last_sent_at;
-- +goose StatementEnd
