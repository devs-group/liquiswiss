-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS go_users_2_organisations (
    user_id BIGINT UNSIGNED,
    organisation_id BIGINT UNSIGNED,
    role ENUM('owner', 'admin', 'editor', 'read-only') NOT NULL,

    PRIMARY KEY (user_id, organisation_id),
    FOREIGN KEY (user_id) REFERENCES go_users (id) ON DELETE CASCADE,
    FOREIGN KEY (organisation_id) REFERENCES go_organisations (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS go_users_2_organisations;
-- +goose StatementEnd