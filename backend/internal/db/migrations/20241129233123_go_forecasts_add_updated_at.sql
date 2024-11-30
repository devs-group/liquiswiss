-- +goose Up
-- +goose StatementBegin
ALTER TABLE go_forecasts
ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE go_forecasts
DROP COLUMN IF EXISTS updated_at;
-- +goose StatementEnd
