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
    calculate_next_history_execution_date(
        'repeating',
        h.from_date,
        h.to_date,
        h.cycle,
        CURDATE()
    ) AS next_execution_date,
    calculate_salary_adjustments(
        h.salary,
        'employee',
        ehc.employee_history_id
    ) as employee_deductions,
    calculate_salary_adjustments(
        h.salary,
        'employer',
        ehc.employee_history_id
    ) as employer_costs,
    h.with_separate_costs
FROM employee_histories h
JOIN employees e ON e.id = h.employee_id
JOIN currencies c ON h.currency_id = c.id
LEFT JOIN employee_history_costs ehc ON ehc.employee_history_id = h.id
WHERE h.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
LIMIT 1;