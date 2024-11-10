-- +goose Up
-- +goose StatementBegin
CREATE TABLE go_forecast_details (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner BIGINT UNSIGNED NOT NULL,
    month VARCHAR(7) NOT NULL,
    revenue JSON NOT NULL,
    expense JSON NOT NULL,
    forecast_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (forecast_id) REFERENCES go_forecasts(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (owner) REFERENCES go_users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    UNIQUE INDEX idx_forecast_detail (owner, month)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS go_forecast_details;
-- +goose StatementEnd
