-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS auto_create_default_scenario
AFTER INSERT ON organisations
FOR EACH ROW
BEGIN
    INSERT INTO scenarios (name, type, is_default, parent_scenario_id, organisation_id)
    VALUES ('Standardszenario', 'horizontal', true, NULL, NEW.id);
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS auto_create_default_scenario;
-- +goose StatementEnd
