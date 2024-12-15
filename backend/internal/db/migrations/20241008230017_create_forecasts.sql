-- +goose Up
-- +goose StatementBegin
CREATE TABLE forecasts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    month VARCHAR(7) NOT NULL,
    revenue BIGINT NOT NULL,
    expense BIGINT NOT NULL,
    cashflow BIGINT NOT NULL,
    deleted BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    organisation_id BIGINT UNSIGNED NOT NULL,

    PRIMARY KEY (id),
    KEY FK_Forecast_Organisation (organisation_id),
    FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE ON UPDATE CASCADE,
    UNIQUE INDEX idx_organisation_forecast (organisation_id, month)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS forecasts;
-- +goose StatementEnd
