-- +goose Up
-- +goose StatementBegin
CREATE TABLE go_bank_accounts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    amount BIGINT NOT NULL,
    currency BIGINT UNSIGNED NOT NULL,
    owner BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (id),

    CONSTRAINT FK_BankAccont_Currency FOREIGN KEY (currency) REFERENCES go_currencies (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_BankAccont_Owner FOREIGN KEY (owner) REFERENCES go_users (id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS go_bank_accounts;
-- +goose StatementEnd
