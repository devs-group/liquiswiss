-- +goose Up
-- +goose StatementBegin
ALTER TABLE forecasts ADD COLUMN IF NOT EXISTS scenario_id BIGINT UNSIGNED NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts ADD CONSTRAINT FK_Forecast_Scenario
    FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE RESTRICT ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_forecasts_scenario ON forecasts(scenario_id);
-- +goose StatementEnd

-- Populate existing forecasts with default scenario for their organisation
-- +goose StatementBegin
UPDATE forecasts f
INNER JOIN scenarios s ON s.organisation_id = f.organisation_id AND s.is_default = true
SET f.scenario_id = s.id
WHERE f.scenario_id IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts MODIFY COLUMN scenario_id BIGINT UNSIGNED NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE forecasts DROP FOREIGN KEY IF EXISTS FK_Forecast_Scenario;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts DROP INDEX IF EXISTS idx_forecasts_scenario;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts DROP COLUMN IF EXISTS scenario_id;
-- +goose StatementEnd
