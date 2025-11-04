-- +goose Up
-- +goose StatementBegin
ALTER TABLE transactions ADD COLUMN uuid CHAR(36) NULL;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE transactions SET uuid = UUID() WHERE uuid IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE transactions MODIFY COLUMN uuid CHAR(36) NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE transactions ADD COLUMN scenario_id BIGINT UNSIGNED NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE transactions ADD CONSTRAINT FK_Transaction_Scenario
    FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE RESTRICT ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_transactions_scenario ON transactions(scenario_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_transactions_uuid ON transactions(uuid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transactions DROP FOREIGN KEY FK_Transaction_Scenario;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE transactions DROP INDEX idx_transactions_scenario;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE transactions DROP INDEX idx_transactions_uuid;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE transactions DROP COLUMN scenario_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE transactions DROP COLUMN uuid;
-- +goose StatementEnd
