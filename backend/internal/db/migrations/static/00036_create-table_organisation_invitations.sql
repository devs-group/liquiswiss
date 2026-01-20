-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS organisation_invitations (
    id SERIAL PRIMARY KEY,
    organisation_id BIGINT UNSIGNED NOT NULL,
    email VARCHAR(100) NOT NULL,
    role ENUM('admin', 'editor', 'read-only') NOT NULL DEFAULT 'read-only',
    token VARCHAR(100) NOT NULL UNIQUE,
    invited_by BIGINT UNSIGNED NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE,
    FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (organisation_id, email)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS organisation_invitations;
-- +goose StatementEnd
