-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS member_permissions (
    id SERIAL PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    organisation_id BIGINT UNSIGNED NOT NULL,
    entity_type VARCHAR(50) DEFAULT NULL,
    can_view BOOL DEFAULT true,
    can_edit BOOL DEFAULT false,
    can_delete BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE,
    UNIQUE (user_id, organisation_id, entity_type)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS member_permissions;
-- +goose StatementEnd
