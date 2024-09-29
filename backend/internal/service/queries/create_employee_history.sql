INSERT INTO go_employee_history
    (
     employee_id,
     hours_per_month,
     salary_per_month,
     salary_currency,
     vacation_days_per_year,
     from_date,
     to_date
    )
SELECT ?, ?, ?, ?, ?, ?, ?
FROM go_employees e
WHERE e.id = ?
  AND e.owner = ?
LIMIT 1;