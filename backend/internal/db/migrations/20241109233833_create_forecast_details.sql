-- +goose Up
-- +goose StatementBegin
CREATE TABLE forecast_details (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    month VARCHAR(7) NOT NULL,
    revenue JSON NOT NULL,
    expense JSON NOT NULL,
    deleted BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    forecast_id BIGINT UNSIGNED NOT NULL,
    organisation_id BIGINT UNSIGNED NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (forecast_id) REFERENCES forecasts(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE ON UPDATE CASCADE,
    UNIQUE INDEX idx_forecast_detail (organisation_id, month)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS forecast_details;
-- +goose StatementEnd
