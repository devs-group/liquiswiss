-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS fiat_rates (
    id SERIAL PRIMARY KEY,
    base CHAR(3) NOT NULL DEFAULT 'CHF',
    target CHAR(3) NOT NULL,
    rate DECIMAL(18, 6) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT UQ_BASE_TARGET UNIQUE (base, target),

    INDEX IDX_TARGET (target)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS fiat_rates
-- +goose StatementEnd
