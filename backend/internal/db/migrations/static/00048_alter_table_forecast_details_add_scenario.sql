-- +goose Up
-- Drop FK, then drop old unique index, create new one with scenario_id, recreate FK
-- +goose StatementBegin
ALTER TABLE forecast_details DROP FOREIGN KEY IF EXISTS forecast_details_ibfk_2;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecast_details DROP INDEX IF EXISTS idx_forecast_detail;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecast_details DROP INDEX IF EXISTS idx_forecast_details_organisation;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX idx_forecast_detail ON forecast_details(scenario_id, month);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_forecast_details_organisation ON forecast_details(organisation_id);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecast_details ADD CONSTRAINT forecast_details_ibfk_2
    FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE forecast_details DROP FOREIGN KEY IF EXISTS forecast_details_ibfk_2;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecast_details DROP INDEX IF EXISTS idx_forecast_details_organisation;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecast_details DROP INDEX IF EXISTS idx_forecast_detail;
-- +goose StatementEnd

-- Delete duplicate forecast_details keeping only the one from default scenario per organisation+month
-- +goose StatementBegin
DELETE fd FROM forecast_details fd
LEFT JOIN scenarios s ON fd.scenario_id = s.id
WHERE s.is_default = false;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX idx_forecast_detail ON forecast_details(organisation_id, month);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecast_details ADD CONSTRAINT forecast_details_ibfk_2
    FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE ON UPDATE CASCADE;
-- +goose StatementEnd
