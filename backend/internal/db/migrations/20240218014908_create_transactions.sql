-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS go_transactions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    cycle ENUM('daily', 'weekly', 'monthly', 'quarterly', 'biannually', 'yearly'),
    type ENUM('single', 'repeating') NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    category BIGINT UNSIGNED NOT NULL,
    currency BIGINT UNSIGNED NOT NULL,
    employee BIGINT UNSIGNED,
    owner BIGINT UNSIGNED NOT NULL,
    organisation BIGINT UNSIGNED,

    CONSTRAINT FK_Transaction_Category FOREIGN KEY (category) REFERENCES go_categories (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_Transaction_Currency FOREIGN KEY (currency) REFERENCES go_currencies (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_Transaction_Employee FOREIGN KEY (employee) REFERENCES go_employees (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_Transaction_Owner FOREIGN KEY (owner) REFERENCES go_users (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_Transaction_Organisation FOREIGN KEY (organisation) REFERENCES go_organisations (id) ON DELETE SET NULL ON UPDATE CASCADE,

    CONSTRAINT CHK_Cycle_Required CHECK (
        type != 'repeating' OR (type = 'repeating' AND cycle IS NOT NULL)
    )
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS go_transactions;
-- +goose StatementEnd