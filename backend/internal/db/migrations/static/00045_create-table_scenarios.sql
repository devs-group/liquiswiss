-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS scenarios (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    is_base BOOL NOT NULL DEFAULT false,
    parent_id BIGINT UNSIGNED NULL,
    organisation_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT FK_Scenario_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_Scenario_Parent FOREIGN KEY (parent_id) REFERENCES scenarios (id) ON DELETE SET NULL ON UPDATE CASCADE,

    INDEX IDX_Scenario_Organisation_Base (organisation_id, is_base)
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO scenarios (name, is_base, organisation_id)
SELECT 'Grundszenario', true, o.id
FROM organisations o
WHERE NOT EXISTS (SELECT 1 FROM scenarios s WHERE s.organisation_id = o.id AND s.is_base);
-- +goose StatementEnd

-- employees
-- +goose StatementBegin
ALTER TABLE employees ADD COLUMN IF NOT EXISTS scenario_id BIGINT UNSIGNED NULL AFTER organisation_id;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE employees e JOIN scenarios s ON s.organisation_id = e.organisation_id AND s.is_base SET e.scenario_id = s.id;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE employees
    MODIFY scenario_id BIGINT UNSIGNED NOT NULL,
    ADD CONSTRAINT FK_Employee_Scenario FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE;
-- +goose StatementEnd

-- transactions
-- +goose StatementBegin
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS scenario_id BIGINT UNSIGNED NULL AFTER organisation_id;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE transactions t JOIN scenarios s ON s.organisation_id = t.organisation_id AND s.is_base SET t.scenario_id = s.id;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE transactions
    MODIFY scenario_id BIGINT UNSIGNED NOT NULL,
    ADD CONSTRAINT FK_Transaction_Scenario FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE;
-- +goose StatementEnd

-- bank_accounts
-- +goose StatementBegin
ALTER TABLE bank_accounts ADD COLUMN IF NOT EXISTS scenario_id BIGINT UNSIGNED NULL AFTER organisation_id;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE bank_accounts b JOIN scenarios s ON s.organisation_id = b.organisation_id AND s.is_base SET b.scenario_id = s.id;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE bank_accounts
    MODIFY scenario_id BIGINT UNSIGNED NOT NULL,
    ADD CONSTRAINT FK_BankAccount_Scenario FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE;
-- +goose StatementEnd

-- salary_cost_exclusions hang on shared labels, so they need their own scenario column
-- +goose StatementBegin
ALTER TABLE salary_cost_exclusions ADD COLUMN IF NOT EXISTS scenario_id BIGINT UNSIGNED NULL;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE salary_cost_exclusions sce
    JOIN salary_cost_labels scl ON scl.id = sce.label_id
    JOIN scenarios s ON s.organisation_id = scl.organisation_id AND s.is_base
SET sce.scenario_id = s.id;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE salary_cost_exclusions
    ADD INDEX IF NOT EXISTS IDX_SalaryCostExclusion_Label (label_id),
    DROP INDEX UQ_ExcludeMonth_SalaryCostLabelID,
    MODIFY scenario_id BIGINT UNSIGNED NOT NULL,
    ADD CONSTRAINT FK_SalaryCostExclusion_Scenario FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE,
    ADD CONSTRAINT UQ_ExcludeMonth_Label_Scenario UNIQUE (exclude_month, label_id, scenario_id);
-- +goose StatementEnd

-- forecasts are recalculated, so old rows are simply moved to the base scenario
-- +goose StatementBegin
ALTER TABLE forecasts ADD COLUMN IF NOT EXISTS scenario_id BIGINT UNSIGNED NULL AFTER organisation_id;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE forecasts f JOIN scenarios s ON s.organisation_id = f.organisation_id AND s.is_base SET f.scenario_id = s.id;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE forecasts
    ADD INDEX IF NOT EXISTS IDX_Forecast_Organisation (organisation_id),
    DROP INDEX IDX_ORGANISATION_MONTH,
    MODIFY scenario_id BIGINT UNSIGNED NOT NULL,
    ADD CONSTRAINT FK_Forecast_Scenario FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE,
    ADD UNIQUE INDEX IDX_SCENARIO_MONTH (scenario_id, month);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecast_details ADD COLUMN IF NOT EXISTS scenario_id BIGINT UNSIGNED NULL AFTER organisation_id;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE forecast_details f JOIN scenarios s ON s.organisation_id = f.organisation_id AND s.is_base SET f.scenario_id = s.id;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE forecast_details
    ADD INDEX IF NOT EXISTS IDX_ForecastDetail_Organisation (organisation_id),
    DROP INDEX idx_forecast_detail,
    MODIFY scenario_id BIGINT UNSIGNED NOT NULL,
    ADD CONSTRAINT FK_ForecastDetail_Scenario FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE,
    ADD UNIQUE INDEX idx_forecast_detail_scenario (scenario_id, month);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE user_organisation_settings
    ADD COLUMN IF NOT EXISTS current_scenario_id BIGINT UNSIGNED NULL AFTER organisation_id,
    ADD CONSTRAINT FK_UserOrgSetting_Scenario FOREIGN KEY (current_scenario_id) REFERENCES scenarios (id) ON DELETE SET NULL ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- Every step tolerates a database without scenarios, the tests run all Down migrations unversioned
-- Rolling back drops every non-base scenario, its data goes with it through ON DELETE CASCADE
-- +goose StatementBegin
BEGIN NOT ATOMIC
    IF EXISTS (SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'scenarios') THEN
        DELETE FROM scenarios WHERE NOT is_base;
    END IF;
END;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE user_organisation_settings
    DROP FOREIGN KEY IF EXISTS FK_UserOrgSetting_Scenario,
    DROP COLUMN IF EXISTS current_scenario_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecast_details
    DROP FOREIGN KEY IF EXISTS FK_ForecastDetail_Scenario,
    DROP INDEX IF EXISTS idx_forecast_detail_scenario,
    DROP COLUMN IF EXISTS scenario_id,
    ADD UNIQUE INDEX IF NOT EXISTS idx_forecast_detail (organisation_id, month),
    DROP INDEX IF EXISTS IDX_ForecastDetail_Organisation;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts
    DROP FOREIGN KEY IF EXISTS FK_Forecast_Scenario,
    DROP INDEX IF EXISTS IDX_SCENARIO_MONTH,
    DROP COLUMN IF EXISTS scenario_id,
    ADD UNIQUE INDEX IF NOT EXISTS IDX_ORGANISATION_MONTH (organisation_id, month),
    DROP INDEX IF EXISTS IDX_Forecast_Organisation;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salary_cost_exclusions
    DROP FOREIGN KEY IF EXISTS FK_SalaryCostExclusion_Scenario,
    DROP INDEX IF EXISTS UQ_ExcludeMonth_Label_Scenario,
    DROP COLUMN IF EXISTS scenario_id,
    ADD UNIQUE INDEX IF NOT EXISTS UQ_ExcludeMonth_SalaryCostLabelID (exclude_month, label_id);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE bank_accounts DROP FOREIGN KEY IF EXISTS FK_BankAccount_Scenario, DROP COLUMN IF EXISTS scenario_id;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE transactions DROP FOREIGN KEY IF EXISTS FK_Transaction_Scenario, DROP COLUMN IF EXISTS scenario_id;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE employees DROP FOREIGN KEY IF EXISTS FK_Employee_Scenario, DROP COLUMN IF EXISTS scenario_id;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS scenarios;
-- +goose StatementEnd
