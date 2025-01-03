-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION calculate_salary_adjustments(
    salary BIGINT,
    distribution_type ENUM('employer', 'employee'),
    employee_history_id BIGINT
)
    RETURNS BIGINT
    DETERMINISTIC
BEGIN
    DECLARE adjustment BIGINT;

    SET adjustment = CAST(
        COALESCE((
            SELECT SUM(
               CASE
                   WHEN ehc.distribution_type = distribution_type AND ehc.amount_type = 'percentage'
                       THEN (salary * ehc.amount / 100000)
                   WHEN ehc.distribution_type = distribution_type AND ehc.amount_type = 'fixed'
                       THEN ehc.amount
                   ELSE 0
               END
            )
            FROM employee_history_costs ehc
            WHERE ehc.employee_history_id = employee_history_id
        ), 0) AS INTEGER
    );
    RETURN adjustment;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS calculate_salary_adjustments;
-- +goose StatementEnd
