-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS transactions
    ADD COLUMN IF NOT EXISTS scenario_id BIGINT UNSIGNED AFTER organisation_id,
    ADD CONSTRAINT FK_Transaction_Scenario FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE,
    ADD INDEX IF NOT EXISTS IDX_Transaction_Scenario (scenario_id);
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE transactions t
JOIN scenarios sc ON sc.organisation_id = t.organisation_id AND sc.is_default = true
    SET t.scenario_id = sc.id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE transactions
    MODIFY COLUMN IF EXISTS scenario_id BIGINT UNSIGNED;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS transactions
    DROP FOREIGN KEY IF EXISTS FK_Transaction_Scenario,
    DROP INDEX IF EXISTS IDX_Transaction_Scenario,
    DROP COLUMN IF EXISTS scenario_id;
-- +goose StatementEnd