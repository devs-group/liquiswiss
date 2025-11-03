SELECT
    s.id AS salary_id,
    s.employee_id,
    s.hours_per_month,
    s.amount,
    s.cycle,
    c.id AS currency_id,
    c.locale_code,
    c.description,
    c.code,
    s.vacation_days_per_year,
    s.from_date,
    s.to_date,
    s.is_termination,
    s.is_disabled,
    CURDATE() AS db_date
FROM salaries s
JOIN employees e ON e.id = s.employee_id
JOIN currencies c ON s.currency_id = c.id
LEFT JOIN salary_costs sc ON sc.salary_id = s.id
WHERE s.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
LIMIT 1;
