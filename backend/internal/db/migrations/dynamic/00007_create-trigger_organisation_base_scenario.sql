-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER trg_organisation_base_scenario
    AFTER INSERT ON organisations
    FOR EACH ROW
    INSERT INTO scenarios (name, is_base, organisation_id) VALUES ('Grundszenario', true, NEW.id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_organisation_base_scenario;
-- +goose StatementEnd
