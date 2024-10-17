-- +goose Up
-- +goose StatementBegin
CREATE TABLE go_fiat_rates (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    base CHAR(3) NOT NULL DEFAULT 'CHF',
    target CHAR(3) NOT NULL,
    rate DECIMAL(18, 6) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY unique_exchange_rate (base, target),
    INDEX idx_target_currency (target)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS go_fiat_rates
-- +goose StatementEnd
