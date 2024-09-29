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
    h.to_date
FROM go_employee_history h
JOIN go_employees e ON e.id = h.employee_id
JOIN go_currencies c ON h.salary_currency = c.id
WHERE h.id = ?   -- history ID
  AND e.owner = ? -- ensure the employee belongs to the current owner
LIMIT 1;