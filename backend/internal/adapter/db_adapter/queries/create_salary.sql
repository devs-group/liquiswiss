INSERT INTO salaries
    (
     employee_id,
     hours_per_month,
     amount,
     cycle,
     currency_id,
     vacation_days_per_year,
     from_date,
     to_date,
     with_separate_costs
    )
SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?
FROM employees e
WHERE e.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
LIMIT 1;