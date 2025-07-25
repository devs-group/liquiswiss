-- +goose Up
-- +goose StatementBegin
CREATE VIEW ranked_salaries AS
WITH ranked_salary AS (
    SELECT
        s.id,
        s.employee_id,
        s.hours_per_month,
        s.amount,
        s.cycle,
        s.currency_id,
        s.vacation_days_per_year,
        s.from_date,
        s.to_date,
        IF(s.from_date > CURDATE(), TRUE, FALSE) AS is_in_future,
        CAST(
            COALESCE(SUM(
                CASE
                    WHEN sc.distribution_type = 'employee' AND sc.amount_type = 'percentage'
                        THEN (s.amount * sc.amount / 100000)
                    WHEN sc.distribution_type = 'employee' AND sc.amount_type = 'fixed'
                        THEN sc.amount
                    ELSE 0
                END
            ), 0) AS INTEGER
        ) AS employee_deductions,
        CAST(
            COALESCE(SUM(
                CASE
                    WHEN sc.distribution_type = 'employer' AND sc.amount_type = 'percentage'
                        THEN (s.amount * sc.amount / 100000)
                    WHEN sc.distribution_type = 'employer' AND sc.amount_type = 'fixed'
                        THEN sc.amount
                    ELSE 0
                END
            ), 0) AS INTEGER
        ) AS employer_costs,
        s.with_separate_costs,
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
    FROM salaries AS s
    LEFT JOIN salary_costs sc ON sc.salary_id = s.id
    WHERE to_date IS NULL OR to_date >= CURDATE()
    GROUP BY
        s.id,
        s.employee_id,
        s.hours_per_month,
        s.amount,
        s.cycle,
        s.vacation_days_per_year,
        s.from_date,
        s.to_date,
        s.with_separate_costs
)
SELECT * FROM ranked_salary;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS ranked_salaries;
-- +goose StatementEnd
