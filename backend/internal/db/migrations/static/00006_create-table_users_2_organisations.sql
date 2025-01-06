-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users_2_organisations (
    user_id BIGINT UNSIGNED,
    organisation_id BIGINT UNSIGNED,
    role ENUM('owner', 'admin', 'editor', 'read-only') NOT NULL,
    is_default BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (user_id, organisation_id),

    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users_2_organisations;
-- +goose StatementEnd