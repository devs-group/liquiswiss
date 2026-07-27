-- +goose Up
-- +goose StatementBegin
-- MariaDB gives the FIRST TIMESTAMP column of a table an implicit
-- ON UPDATE CURRENT_TIMESTAMP. Any UPDATE on organisation_invitations
-- (e.g. resend setting last_sent_at) silently reset expires_at to NOW(),
-- making resent invitations instantly "expired". DATETIME has no such
-- implicit behavior. refresh_tokens has the same latent trap.
ALTER TABLE organisation_invitations
    MODIFY expires_at DATETIME NOT NULL;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE refresh_tokens
    MODIFY expires_at DATETIME NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE organisation_invitations
    MODIFY expires_at TIMESTAMP NOT NULL;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE refresh_tokens
    MODIFY expires_at TIMESTAMP NOT NULL;
-- +goose StatementEnd
