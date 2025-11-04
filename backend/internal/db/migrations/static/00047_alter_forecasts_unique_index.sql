-- +goose Up
-- Drop FK, then drop old unique index, create new one with scenario_id, recreate FK
-- +goose StatementBegin
ALTER TABLE forecasts DROP FOREIGN KEY IF EXISTS forecasts_ibfk_1;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts DROP INDEX IF EXISTS IDX_ORGANISATION_MONTH;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX IDX_SCENARIO_MONTH ON forecasts(scenario_id, month);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_forecasts_organisation ON forecasts(organisation_id);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts ADD CONSTRAINT forecasts_ibfk_1
    FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE forecasts DROP FOREIGN KEY IF EXISTS forecasts_ibfk_1;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts DROP INDEX IF EXISTS idx_forecasts_organisation;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts DROP INDEX IF EXISTS IDX_SCENARIO_MONTH;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts DROP INDEX IF EXISTS IDX_ORGANISATION_MONTH;
-- +goose StatementEnd

-- Delete duplicate forecasts keeping only the one from default scenario per organisation+month
-- +goose StatementBegin
DELETE f FROM forecasts f
LEFT JOIN scenarios s ON f.scenario_id = s.id
WHERE s.is_default = false;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX IDX_ORGANISATION_MONTH ON forecasts(organisation_id, month);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts ADD CONSTRAINT forecasts_ibfk_1
    FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE ON UPDATE CASCADE;
-- +goose StatementEnd
