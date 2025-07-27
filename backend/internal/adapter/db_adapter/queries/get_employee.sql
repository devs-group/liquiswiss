SELECT
    e.id,
    e.name,
    rs.hours_per_month,
    IF(rs.with_separate_costs, rs.amount + rs.employer_costs, rs.amount) AS salary,
    rs.cycle,
    c.id,
    c.locale_code,
    c.description,
    c.code,
    rs.vacation_days_per_year,
    rs.from_date,
    rs.to_date,
    COALESCE(rs.is_in_future, false) AS is_in_future,
    COALESCE(rs.with_separate_costs, false) AS with_separate_costs,
    COALESCE(rs.is_termination, false) AS is_terminated,
    MAX(CASE WHEN s.is_termination = 1 THEN 1 ELSE 0 END) AS will_be_terminated,
    rs.id AS salary_id
FROM employees e
LEFT JOIN ranked_salaries rs ON e.id = rs.employee_id AND rs.rn = 1
LEFT JOIN currencies c ON rs.currency_id = c.id
LEFT JOIN salaries s ON s.employee_id = e.id AND s.from_date > CURDATE() AND s.is_termination = 1
WHERE e.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
LIMIT 1;