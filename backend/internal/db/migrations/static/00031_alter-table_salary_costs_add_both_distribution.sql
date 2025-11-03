-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS salary_costs
MODIFY distribution_type ENUM('employer', 'employee', 'both') NOT NULL DEFAULT 'employee';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE salary_costs
SET distribution_type = 'employer'
WHERE distribution_type = 'both';
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE IF EXISTS salary_costs
MODIFY distribution_type ENUM('employer', 'employee') NOT NULL DEFAULT 'employee';
-- +goose StatementEnd
