WITH ranked_history AS (
    SELECT
        employee_id,
        hours_per_month,
        salary,
        cycle,
        currency_id,
        vacation_days_per_year,
        from_date,
        to_date,
        IF(from_date > CURDATE(), TRUE, FALSE) AS is_in_future,
        ROW_NUMBER() OVER (
            PARTITION BY employee_id
            ORDER BY
                CASE
                    WHEN from_date <= CURDATE() AND (to_date IS NULL OR to_date >= CURDATE()) THEN 1
                    WHEN from_date > CURDATE() THEN 2 -- Next, prioritize future entries
                    ELSE 3
                    END,
                from_date -- Sort ASC from_date for ties
        ) AS rn
    FROM employee_history
    WHERE to_date IS NULL OR to_date >= CURDATE()
)
SELECT
    e.id,
    e.name,
    h.hours_per_month,
    h.salary,
    h.cycle,
    c.id,
    c.locale_code,
    c.description,
    c.code,
    h.vacation_days_per_year,
    h.from_date,
    h.to_date,
    h.is_in_future,
    COUNT(*) OVER () AS total_count
FROM employees e
LEFT JOIN ranked_history h ON e.id = h.employee_id AND h.rn = 1
LEFT JOIN currencies c ON h.currency_id = c.id
WHERE e.organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)
ORDER BY
    %s IS NULL,
    %s %s,
    e.name ASC
LIMIT ? OFFSET ?;