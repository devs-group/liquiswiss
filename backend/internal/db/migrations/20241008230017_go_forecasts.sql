-- +goose Up
-- +goose StatementBegin
CREATE TABLE go_forecasts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner BIGINT UNSIGNED NOT NULL,
    month VARCHAR(7) NOT NULL,
    revenue BIGINT NOT NULL,
    expense BIGINT NOT NULL,
    cashflow BIGINT NOT NULL,
    PRIMARY KEY (id),
    KEY FK_Forecast_Owner (owner),
    FOREIGN KEY (owner) REFERENCES go_users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    UNIQUE INDEX idx_owner_forecast (owner, month)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS go_forecasts;
-- +goose StatementEnd
