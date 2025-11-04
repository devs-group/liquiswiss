-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS transaction_overrides (
    id SERIAL PRIMARY KEY,
    uuid CHAR(36) NOT NULL,
    scenario_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NULL,
    amount BIGINT NULL,
    vat_included BOOL NULL,
    is_disabled BOOL NULL,
    cycle ENUM('monthly', 'quarterly', 'biannually', 'yearly') NULL,
    type ENUM('single', 'repeating') NULL,
    start_date DATE NULL,
    end_date DATE NULL,
    deleted BOOL NULL,
    vat_id BIGINT UNSIGNED NULL,
    category_id BIGINT UNSIGNED NULL,
    employee_id BIGINT UNSIGNED NULL,
    currency_id BIGINT UNSIGNED NULL,
    organisation_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT FK_TransactionOverride_Scenario FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_TransactionOverride_Vat FOREIGN KEY (vat_id) REFERENCES vats (id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT FK_TransactionOverride_Category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT FK_TransactionOverride_Employee FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT FK_TransactionOverride_Currency FOREIGN KEY (currency_id) REFERENCES currencies (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT FK_TransactionOverride_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE ON UPDATE CASCADE,

    UNIQUE KEY UQ_TransactionOverride_UUID_Scenario (uuid, scenario_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_transaction_overrides_scenario ON transaction_overrides(scenario_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_transaction_overrides_uuid ON transaction_overrides(uuid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transaction_overrides;
-- +goose StatementEnd
