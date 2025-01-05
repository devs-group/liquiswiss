-- +goose Up
-- +goose StatementBegin
CREATE VIEW ranked_employee_histories AS
WITH ranked_history AS (
    SELECT
        hc.id,
        hc.employee_id,
        hc.hours_per_month,
        hc.salary,
        hc.cycle,
        hc.currency_id,
        hc.vacation_days_per_year,
        hc.from_date,
        hc.to_date,
        IF(hc.from_date > CURDATE(), TRUE, FALSE) AS is_in_future,
        CAST(
            COALESCE(SUM(
                CASE
                    WHEN ehc.distribution_type = 'employee' AND ehc.amount_type = 'percentage'
                        THEN (hc.salary * ehc.amount / 100000)
                    WHEN ehc.distribution_type = 'employee' AND ehc.amount_type = 'fixed'
                        THEN ehc.amount
                    ELSE 0
                END
            ), 0) AS INTEGER
        ) AS employee_deductions,
        CAST(
            COALESCE(SUM(
                CASE
                    WHEN ehc.distribution_type = 'employer' AND ehc.amount_type = 'percentage'
                        THEN (hc.salary * ehc.amount / 100000)
                    WHEN ehc.distribution_type = 'employer' AND ehc.amount_type = 'fixed'
                        THEN ehc.amount
                    ELSE 0
                END
            ), 0) AS INTEGER
        ) AS employer_costs,
        hc.with_separate_costs,
        ROW_NUMBER() OVER (
            PARTITION BY employee_id
            ORDER BY
                CASE
                    WHEN from_date <= CURDATE() AND (to_date IS NULL OR to_date >= CURDATE()) THEN 1
                    WHEN from_date > CURDATE() THEN 2
                    ELSE 3
                END,
                from_date
        ) AS rn
    FROM employee_histories AS hc
    LEFT JOIN employee_history_costs ehc ON ehc.employee_history_id = hc.id
    WHERE to_date IS NULL OR to_date >= CURDATE()
    GROUP BY
        hc.id,
        hc.employee_id,
        hc.hours_per_month,
        hc.salary,
        hc.cycle,
        hc.vacation_days_per_year,
        hc.from_date,
        hc.to_date,
        hc.with_separate_costs
)
SELECT * FROM ranked_history;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS ranked_employee_histories;
-- +goose StatementEnd
