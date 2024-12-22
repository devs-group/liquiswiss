SELECT
    h.id,
    h.employee_id,
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
    COUNT(*) OVER() AS total_count
FROM employee_history h
JOIN employees e ON e.id = h.employee_id
JOIN currencies c ON h.currency_id = c.id
WHERE h.employee_id = ?
  AND e.organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)
ORDER BY h.from_date DESC -- Order by the most recent history first
LIMIT ? OFFSET ?