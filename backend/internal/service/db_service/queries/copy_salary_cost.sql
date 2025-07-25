INSERT INTO salary_costs (
    cycle,
    amount_type,
    amount,
    distribution_type,
    relative_offset,
    target_date,
    label_id,
    salary_id
)
SELECT
    hc.cycle,
    hc.amount_type,
    hc.amount,
    hc.distribution_type,
    hc.relative_offset,
    hc.target_date,
    hc.label_id,
    ?
FROM salary_costs as hc
JOIN salaries AS h ON h.id = hc.salary_id
JOIN employees AS e ON e.id = h.employee_id
WHERE hc.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)