-- +goose Up
-- +goose StatementBegin
CREATE VIEW ranked_salaries AS
WITH current_month AS (
    SELECT DATE_FORMAT(CURDATE(), '%Y-%m') AS month
),
latest_cost_details AS (
    SELECT
        scd.cost_id,
        scd.amount / scd.divider AS calculated_amount,
        ROW_NUMBER() OVER (
            PARTITION BY scd.cost_id
            ORDER BY
                CASE WHEN scd.month >= (SELECT month FROM current_month) THEN 0 ELSE 1 END,
                ABS(TIMESTAMPDIFF(MONTH, STR_TO_DATE(CONCAT(scd.month, '-01'), '%Y-%m-%d'), CURDATE()))
        ) AS rn
    FROM salary_cost_details scd
),
ranked_salary AS (
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
                    WHEN sc.distribution_type IN ('employee', 'both')
                        THEN COALESCE(lcd.calculated_amount, 0)
                    ELSE 0
                END
            ), 0) AS INTEGER
        ) AS employee_deductions,
        CAST(
            COALESCE(SUM(
                CASE
                    WHEN sc.distribution_type IN ('employer', 'both')
                        THEN COALESCE(lcd.calculated_amount, 0)
                    ELSE 0
                END
            ), 0) AS INTEGER
        ) AS employer_costs,
        s.is_termination,
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
    LEFT JOIN latest_cost_details lcd ON lcd.cost_id = sc.id AND lcd.rn = 1
    WHERE (s.to_date IS NULL OR s.to_date >= CURDATE())
      AND s.is_disabled = 0
    GROUP BY
        s.id,
        s.employee_id,
        s.hours_per_month,
        s.amount,
        s.cycle,
        s.vacation_days_per_year,
        s.from_date,
        s.to_date,
        s.is_termination
)
SELECT * FROM ranked_salary;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS ranked_salaries;
-- +goose StatementEnd
