SELECT
    h.id,
    h.employee_id,
    h.hours_per_month,
    h.salary_per_month,
    c.id,
    c.locale_code,
    c.description,
    c.code,
    h.vacation_days_per_year,
    h.from_date,
    h.to_date,
    COUNT(*) OVER() AS total_count
FROM go_employee_history h
JOIN go_employees e ON e.id = h.employee_id
JOIN go_currencies c ON h.salary_currency = c.id
WHERE h.employee_id = ?  -- Filter by the specific employee_id
  AND e.owner = ?        -- Ensure the employee belongs to the current user (owner)
ORDER BY h.from_date DESC -- Order by the most recent history first
LIMIT ? OFFSET ?;        -- Pagination: limit the number of rows returned and apply offset