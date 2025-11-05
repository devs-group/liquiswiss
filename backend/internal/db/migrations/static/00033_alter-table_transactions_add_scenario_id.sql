-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS transactions
    ADD COLUMN IF NOT EXISTS scenario_id BIGINT UNSIGNED AFTER organisation_id;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE transactions AS t
JOIN scenarios AS s ON s.organisation_id = t.organisation_id AND s.is_default = TRUE
SET t.scenario_id = s.id
WHERE t.scenario_id IS NULL;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE transactions
    MODIFY COLUMN scenario_id BIGINT UNSIGNED NOT NULL,
    ADD CONSTRAINT fk_transactions_scenario FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE RESTRICT ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transactions
    DROP FOREIGN KEY IF EXISTS fk_transactions_scenario;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE IF EXISTS transactions
    DROP COLUMN IF EXISTS scenario_id;
-- +goose StatementEnd
