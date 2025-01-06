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
                        THEN
                        CASE
                            WHEN ehc.cycle IN ('once') THEN
                                (salary * ehc.amount / 100000)
                            WHEN ehc.cycle IN ('daily', 'weekly') THEN
                                (salary * ehc.amount / 100000) * ehc.relative_offset
                            WHEN ehc.cycle = 'monthly' THEN
                                CASE h.cycle
                                    WHEN 'monthly' THEN
                                        (salary * ehc.amount / 100000)
                                    WHEN 'quarterly' THEN
                                        (salary * ehc.amount / 100000) * 3
                                    WHEN 'biannually' THEN
                                        (salary * ehc.amount / 100000) * 6
                                    WHEN 'yearly' THEN
                                        (salary * ehc.amount / 100000) * 12
                                END
                            WHEN ehc.cycle = 'quarterly' THEN
                                CASE h.cycle
                                    WHEN 'monthly' THEN
                                        (salary * ehc.amount / 100000) / 3
                                    WHEN 'quarterly' THEN
                                        (salary * ehc.amount / 100000)
                                    WHEN 'biannually' THEN
                                        (salary * ehc.amount / 100000) * 2
                                    WHEN 'yearly' THEN
                                        (salary * ehc.amount / 100000) * 4
                                END
                            WHEN ehc.cycle = 'biannually' THEN
                                CASE h.cycle
                                    WHEN 'monthly' THEN
                                        (salary * ehc.amount / 100000) / 6
                                    WHEN 'quarterly' THEN
                                        (salary * ehc.amount / 100000) / 3
                                    WHEN 'biannually' THEN
                                        (salary * ehc.amount / 100000)
                                    WHEN 'yearly' THEN
                                        (salary * ehc.amount / 100000) * 2
                                END
                            WHEN ehc.cycle = 'yearly' THEN
                                CASE h.cycle
                                    WHEN 'monthly' THEN
                                        (salary * ehc.amount / 100000) / 12
                                    WHEN 'quarterly' THEN
                                        (salary * ehc.amount / 100000) / 4
                                    WHEN 'biannually' THEN
                                        (salary * ehc.amount / 100000) / 2
                                    WHEN 'yearly' THEN
                                        (salary * ehc.amount / 100000)
                                END
                        END
                    WHEN ehc.distribution_type = distribution_type AND ehc.amount_type = 'fixed'
                        THEN
                        CASE
                            WHEN ehc.cycle IN ('once') THEN
                                ehc.amount
                            WHEN ehc.cycle IN ('daily', 'weekly') THEN
                                ehc.amount * ehc.relative_offset
                            WHEN ehc.cycle = 'monthly' THEN
                                CASE h.cycle
                                    WHEN 'monthly' THEN
                                        ehc.amount
                                    WHEN 'quarterly' THEN
                                        ehc.amount * 3
                                    WHEN 'biannually' THEN
                                        ehc.amount * 6
                                    WHEN 'yearly' THEN
                                        ehc.amount * 12
                                END
                            WHEN ehc.cycle = 'quarterly' THEN
                                CASE h.cycle
                                    WHEN 'monthly' THEN
                                        ehc.amount / 3
                                    WHEN 'quarterly' THEN
                                        ehc.amount
                                    WHEN 'biannually' THEN
                                        ehc.amount * 2
                                    WHEN 'yearly' THEN
                                        ehc.amount * 4
                                END
                            WHEN ehc.cycle = 'biannually' THEN
                                CASE h.cycle
                                    WHEN 'monthly' THEN
                                        ehc.amount / 6
                                    WHEN 'quarterly' THEN
                                        ehc.amount / 3
                                    WHEN 'biannually' THEN
                                        ehc.amount
                                    WHEN 'yearly' THEN
                                        ehc.amount * 2
                                END
                            WHEN ehc.cycle = 'yearly' THEN
                                CASE h.cycle
                                    WHEN 'monthly' THEN
                                        ehc.amount / 12
                                    WHEN 'quarterly' THEN
                                        ehc.amount / 4
                                    WHEN 'biannually' THEN
                                        ehc.amount / 2
                                    WHEN 'yearly' THEN
                                        ehc.amount
                                END
                        END
                    ELSE 0
                END
            )
            FROM employee_history_costs ehc
            JOIN employee_histories h ON h.id = ehc.employee_history_id
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
