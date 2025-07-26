-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    vat_included BOOL NOT NULL DEFAULT false,
    cycle ENUM('monthly', 'quarterly', 'biannually', 'yearly'),
    type ENUM('single', 'repeating') NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    deleted BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    vat_id BIGINT UNSIGNED,
    category_id BIGINT UNSIGNED,
    employee_id BIGINT UNSIGNED,
    currency_id BIGINT UNSIGNED NOT NULL,
    organisation_id BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_Transaction_Vat FOREIGN KEY (vat_id) REFERENCES vats (id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT FK_Transaction_Category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT FK_Transaction_Employee FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT FK_Transaction_Currency FOREIGN KEY (currency_id) REFERENCES currencies (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT FK_Transaction_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE ON UPDATE CASCADE,

    CONSTRAINT CK_Cycle_Required CHECK (type != 'repeating' OR (type = 'repeating' AND cycle IS NOT NULL))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions;
-- +goose StatementEnd