-- +goose Up
-- +goose StatementBegin
ALTER TABLE employee_history
    ADD COLUMN cycle ENUM('daily', 'weekly', 'monthly', 'quarterly', 'biannually', 'yearly') NOT NULL DEFAULT 'monthly';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employee_history
    DROP COLUMN IF EXISTS cycle;
-- +goose StatementEnd
