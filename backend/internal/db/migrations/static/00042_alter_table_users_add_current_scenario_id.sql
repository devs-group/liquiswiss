-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN current_scenario_id BIGINT UNSIGNED NULL;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE users ADD CONSTRAINT FK_User_Scenario FOREIGN KEY (current_scenario_id) REFERENCES scenarios (id);
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE users u
SET u.current_scenario_id = (
    SELECT s.id
    FROM scenarios s
    WHERE s.organisation_id = u.current_organisation_id
    AND s.is_default = true
    LIMIT 1
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP FOREIGN KEY IF EXISTS FK_User_Scenario;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN IF EXISTS current_scenario_id;
-- +goose StatementEnd
