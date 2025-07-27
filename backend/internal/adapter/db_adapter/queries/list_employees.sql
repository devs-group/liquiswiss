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
    COALESCE(rs.is_termination, false) AS is_termination,
    rs.id AS salary_id,
    COUNT(*) OVER () AS total_count
FROM employees e
LEFT JOIN ranked_salaries rs ON e.id = rs.employee_id AND rs.rn = 1
LEFT JOIN currencies c ON rs.currency_id = c.id
WHERE e.organisation_id = get_current_user_organisation_id(?)
ORDER BY
    %s IS NULL,
    %s %s,
    e.name ASC
LIMIT ? OFFSET ?;