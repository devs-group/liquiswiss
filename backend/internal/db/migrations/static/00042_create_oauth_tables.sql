-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS oauth_clients (
    id SERIAL PRIMARY KEY,
    client_id VARCHAR(36) NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    redirect_uris JSON NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (client_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS oauth_auth_codes (
    id SERIAL PRIMARY KEY,
    code_hash CHAR(64) NOT NULL,
    client_id VARCHAR(36) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    code_challenge VARCHAR(128) NOT NULL,
    redirect_uri VARCHAR(1024) NOT NULL,
    resource VARCHAR(1024) DEFAULT NULL,
    expires_at DATETIME NOT NULL,
    used_at DATETIME NULL DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (code_hash),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS oauth_refresh_tokens (
    id SERIAL PRIMARY KEY,
    token_hash CHAR(64) NOT NULL,
    client_id VARCHAR(36) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    expires_at DATETIME NOT NULL,
    revoked_at DATETIME NULL DEFAULT NULL,
    rotated_from CHAR(64) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (token_hash),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS oauth_refresh_tokens;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS oauth_auth_codes;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS oauth_clients;
-- +goose StatementEnd
