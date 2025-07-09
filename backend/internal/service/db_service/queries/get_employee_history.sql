SELECT
    h.id AS history_id,
    h.employee_id,
    h.hours_per_month,
    h.salary,
    h.cycle,
    c.id AS currency_id,
    c.locale_code,
    c.description,
    c.code,
    h.vacation_days_per_year,
    h.from_date,
    h.to_date,
    h.with_separate_costs,
    CURDATE() AS db_date
FROM employee_histories h
JOIN employees e ON e.id = h.employee_id
JOIN currencies c ON h.currency_id = c.id
LEFT JOIN employee_history_costs ehc ON ehc.employee_history_id = h.id
WHERE h.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
LIMIT 1;