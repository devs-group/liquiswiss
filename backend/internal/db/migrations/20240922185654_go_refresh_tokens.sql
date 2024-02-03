-- +goose Up
-- +goose StatementBegin
CREATE TABLE go_refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    token_id VARCHAR(36) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    device_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, token_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE go_refresh_tokens;
-- +goose StatementEnd
