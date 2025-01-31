-- +goose Up
-- +goose StatementBegin
CREATE TABLE transaction_exclusions (
    id SERIAL PRIMARY KEY,
    exclude_month VARCHAR(7) NOT NULL,

    transaction_id BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_Exclusion_Transaction FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE ON UPDATE CASCADE,

    CONSTRAINT CK_ExcludeMonth_Not_Empty CHECK (exclude_month <> ''),

    CONSTRAINT UQ_ExcludeMonth_TransactionID UNIQUE (exclude_month, transaction_id),

    INDEX IDX_ExcludeMonth_TransactionID (exclude_month, transaction_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transaction_exclusions;
-- +goose StatementEnd