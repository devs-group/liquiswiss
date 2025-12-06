-- +goose Up
-- +goose StatementBegin
INSERT INTO scenarios (name, is_default, organisation_id)
SELECT 'Standardszenario', true, id FROM organisations;
-- +goose StatementEnd
