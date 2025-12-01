-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS vat_settings (
    id SERIAL PRIMARY KEY,
    organisation_id BIGINT UNSIGNED NOT NULL,
    enabled BOOL DEFAULT false,
    billing_date DATE NOT NULL COMMENT 'Rechnungszeitpunkt - reference date for VAT calculation period',
    transaction_month_offset INT NOT NULL DEFAULT 0 COMMENT 'Months after billing date when VAT payment appears in forecast (0 = same month, 1 = 1 month later, etc.)',
    `interval` ENUM('monthly', 'quarterly', 'biannually', 'yearly') NOT NULL DEFAULT 'quarterly',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE,
    UNIQUE KEY unique_org_vat_setting (organisation_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS vat_settings;
-- +goose StatementEnd
