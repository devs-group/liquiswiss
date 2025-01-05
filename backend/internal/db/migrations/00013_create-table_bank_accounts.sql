-- +goose Up
-- +goose StatementBegin
CREATE TABLE bank_accounts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    amount BIGINT NOT NULL,
    deleted BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    currency_id BIGINT UNSIGNED NOT NULL,
    organisation_id BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_BankAccount_Currency FOREIGN KEY (currency_id) REFERENCES currencies (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT FK_BankAccount_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bank_accounts;
-- +goose StatementEnd
