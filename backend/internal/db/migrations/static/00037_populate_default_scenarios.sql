-- +goose Up
-- +goose StatementBegin
INSERT INTO scenarios (name, type, is_default, parent_scenario_id, organisation_id)
SELECT 'Standardszenario', 'horizontal', true, NULL, id
FROM organisations;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE employees e
INNER JOIN scenarios s ON s.organisation_id = e.organisation_id AND s.is_default = true
SET e.scenario_id = s.id
WHERE e.scenario_id IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE salaries sal
INNER JOIN employees e ON sal.employee_id = e.id
INNER JOIN scenarios s ON s.organisation_id = e.organisation_id AND s.is_default = true
SET sal.scenario_id = s.id
WHERE sal.scenario_id IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE salary_costs sc
INNER JOIN salaries sal ON sc.salary_id = sal.id
INNER JOIN employees e ON sal.employee_id = e.id
INNER JOIN scenarios s ON s.organisation_id = e.organisation_id AND s.is_default = true
SET sc.scenario_id = s.id
WHERE sc.scenario_id IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE transactions t
INNER JOIN scenarios s ON s.organisation_id = t.organisation_id AND s.is_default = true
SET t.scenario_id = s.id
WHERE t.scenario_id IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE employees MODIFY COLUMN scenario_id BIGINT UNSIGNED NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salaries MODIFY COLUMN scenario_id BIGINT UNSIGNED NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salary_costs MODIFY COLUMN scenario_id BIGINT UNSIGNED NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE transactions MODIFY COLUMN scenario_id BIGINT UNSIGNED NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transactions MODIFY COLUMN scenario_id BIGINT UNSIGNED NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salary_costs MODIFY COLUMN scenario_id BIGINT UNSIGNED NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE salaries MODIFY COLUMN scenario_id BIGINT UNSIGNED NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE employees MODIFY COLUMN scenario_id BIGINT UNSIGNED NULL;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE transactions SET scenario_id = NULL;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE salary_costs SET scenario_id = NULL;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE salaries SET scenario_id = NULL;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE employees SET scenario_id = NULL;
-- +goose StatementEnd

-- +goose StatementBegin
DELETE FROM scenarios WHERE parent_scenario_id IS NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
DELETE FROM scenarios WHERE is_default = true;
-- +goose StatementEnd
