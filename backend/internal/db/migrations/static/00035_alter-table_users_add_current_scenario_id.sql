-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS users
    ADD COLUMN IF NOT EXISTS current_scenario_id BIGINT UNSIGNED NULL AFTER current_organisation_id;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE users AS u
LEFT JOIN scenarios AS s ON s.organisation_id = u.current_organisation_id AND s.is_default = TRUE
SET u.current_scenario_id = s.id
WHERE u.current_organisation_id IS NOT NULL
  AND u.current_scenario_id IS NULL;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE users
    ADD CONSTRAINT fk_users_current_scenario FOREIGN KEY (current_scenario_id) REFERENCES scenarios (id) ON DELETE SET NULL ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS users
    DROP FOREIGN KEY IF EXISTS fk_users_current_scenario;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE IF EXISTS users
    DROP COLUMN IF EXISTS current_scenario_id;
-- +goose StatementEnd
