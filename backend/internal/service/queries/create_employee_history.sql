INSERT INTO employee_history
    (
     employee_id,
     hours_per_month,
     salary_per_month,
     currency_id,
     vacation_days_per_year,
     from_date,
     to_date
    )
SELECT ?, ?, ?, ?, ?, ?, ?
FROM employees e
WHERE e.id = ?
  AND e.organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)
LIMIT 1;