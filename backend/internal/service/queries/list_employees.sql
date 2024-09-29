WITH ranked_history AS (
    SELECT
        employee_id,
        hours_per_month,
        salary_per_month,
        salary_currency,
        vacation_days_per_year,
        from_date,
        to_date,
        ROW_NUMBER() OVER (PARTITION BY employee_id ORDER BY from_date DESC) AS rn
    FROM go_employee_history
    WHERE from_date <= CURDATE()
      AND (to_date IS NULL OR to_date >= CURDATE())
)
SELECT
    e.id,
    e.name,
    h.hours_per_month,
    h.salary_per_month,
    c.id,
    c.locale_code,
    c.description,
    c.code,
    h.vacation_days_per_year,
    h.from_date,
    h.to_date,
    COUNT(*) OVER () AS total_count
FROM go_employees e
LEFT JOIN ranked_history h ON e.id = h.employee_id AND h.rn = 1
LEFT JOIN go_currencies c ON h.salary_currency = c.id
WHERE e.owner = ? -- Filter by the owner
LIMIT ? OFFSET ?;