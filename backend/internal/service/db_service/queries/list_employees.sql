SELECT
    e.id,
    e.name,
    h.hours_per_month,
    IF(h.with_separate_costs, h.salary + h.employer_costs, h.salary) AS salary,
    h.cycle,
    c.id,
    c.locale_code,
    c.description,
    c.code,
    h.vacation_days_per_year,
    h.from_date,
    h.to_date,
    COALESCE(h.is_in_future, false) AS is_in_future,
    COALESCE(h.with_separate_costs, false) AS with_separate_costs,
    h.id AS history_id,
    COUNT(*) OVER () AS total_count
FROM employees e
LEFT JOIN ranked_employee_histories h ON e.id = h.employee_id AND h.rn = 1
LEFT JOIN currencies c ON h.currency_id = c.id
WHERE e.organisation_id = get_current_organisation(?)
ORDER BY
    %s IS NULL,
    %s %s,
    e.name ASC
LIMIT ? OFFSET ?;