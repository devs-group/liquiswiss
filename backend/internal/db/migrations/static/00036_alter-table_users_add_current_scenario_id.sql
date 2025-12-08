-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS current_scenario_id BIGINT UNSIGNED AFTER current_organisation_id,
    ADD CONSTRAINT FK_Current_Scenario FOREIGN KEY (current_scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE,
    ADD INDEX IF NOT EXISTS IDX_Current_Scenario (current_scenario_id);
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE users u
JOIN scenarios sc ON sc.organisation_id = u.current_organisation_id AND sc.is_default = true
SET u.current_scenario_id = sc.id
WHERE u.current_organisation_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP FOREIGN KEY IF EXISTS FK_Current_Scenario,
    DROP INDEX IF EXISTS IDX_Current_Scenario,
    DROP COLUMN IF EXISTS current_scenario_id;
-- +goose StatementEnd
