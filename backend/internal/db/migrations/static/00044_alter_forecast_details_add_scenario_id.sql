-- +goose Up
-- +goose StatementBegin
ALTER TABLE forecast_details ADD COLUMN IF NOT EXISTS scenario_id BIGINT UNSIGNED NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecast_details ADD CONSTRAINT FK_ForecastDetail_Scenario
    FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE RESTRICT ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_forecast_details_scenario ON forecast_details(scenario_id);
-- +goose StatementEnd

-- Populate existing forecast_details with default scenario for their organisation
-- +goose StatementBegin
UPDATE forecast_details fd
INNER JOIN scenarios s ON s.organisation_id = fd.organisation_id AND s.is_default = true
SET fd.scenario_id = s.id
WHERE fd.scenario_id IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecast_details MODIFY COLUMN scenario_id BIGINT UNSIGNED NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE forecast_details DROP FOREIGN KEY IF EXISTS FK_ForecastDetail_Scenario;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecast_details DROP INDEX IF EXISTS idx_forecast_details_scenario;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecast_details DROP COLUMN IF EXISTS scenario_id;
-- +goose StatementEnd
