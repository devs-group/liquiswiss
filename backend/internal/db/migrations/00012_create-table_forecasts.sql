-- +goose Up
-- +goose StatementBegin
CREATE TABLE forecasts (
    id SERIAL PRIMARY KEY,
    month VARCHAR(7) NOT NULL,
    revenue BIGINT NOT NULL,
    expense BIGINT NOT NULL,
    cashflow BIGINT NOT NULL,
    deleted BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    organisation_id BIGINT UNSIGNED NOT NULL,

    FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE ON UPDATE CASCADE,

    UNIQUE INDEX IDX_ORGANISATION_MONTH (organisation_id, month)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS forecasts;
-- +goose StatementEnd
