-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    cycle ENUM('daily', 'weekly', 'monthly', 'quarterly', 'biannually', 'yearly'),
    type ENUM('single', 'repeating') NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    deleted BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    category_id BIGINT UNSIGNED NOT NULL,
    currency_id BIGINT UNSIGNED NOT NULL,
    employee_id BIGINT UNSIGNED,
    organisation_id BIGINT UNSIGNED,

    CONSTRAINT FK_Transaction_Category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_Transaction_Currency FOREIGN KEY (currency_id) REFERENCES currencies (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_Transaction_Employee FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_Transaction_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE ON UPDATE CASCADE,

    CONSTRAINT CHK_Cycle_Required CHECK (
        type != 'repeating' OR (type = 'repeating' AND cycle IS NOT NULL)
    )
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions;
-- +goose StatementEnd