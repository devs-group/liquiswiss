-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS transactions_override (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NULL,
    amount BIGINT NULL,
    vat_included BOOL NULL,
    cycle ENUM('monthly', 'quarterly', 'biannually', 'yearly') NULL,
    type ENUM('single', 'repeating') NULL,
    start_date DATE NULL,
    end_date DATE NULL,
    deleted BOOL NULL,
    created_at TIMESTAMP NULL,

    vat_id BIGINT UNSIGNED NULL,
    category_id BIGINT UNSIGNED NULL,
    employee_id BIGINT UNSIGNED NULL,
    currency_id BIGINT UNSIGNED NULL,
    organisation_id BIGINT UNSIGNED NULL,
    scenario_id BIGINT UNSIGNED NOT NULL,
    parent_transaction_id BIGINT UNSIGNED NULL,

    CONSTRAINT fk_transactions_override_vat FOREIGN KEY (vat_id) REFERENCES vats (id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT fk_transactions_override_category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT fk_transactions_override_employee FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT fk_transactions_override_currency FOREIGN KEY (currency_id) REFERENCES currencies (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_transactions_override_organisation FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_transactions_override_scenario FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_transactions_override_transaction FOREIGN KEY (parent_transaction_id) REFERENCES transactions (id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions_override;
-- +goose StatementEnd
