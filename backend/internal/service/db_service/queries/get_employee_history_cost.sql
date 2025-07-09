SELECT
    hc.id,
    hcl.id AS label_id,
    hcl.name AS label_name,
    hc.cycle,
    hc.amount_type,
    hc.amount,
    hc.distribution_type,
    hc.relative_offset,
    hc.target_date,
    hc.employee_history_id,
    h.cycle AS history_cycle,
    h.salary AS history_salary,
    h.from_date AS history_from_date,
    h.to_date AS history_to_date,
    CURDATE() AS db_date
FROM employee_history_costs hc
JOIN employee_histories h ON h.id = hc.employee_history_id
JOIN employees e ON e.id = h.employee_id
LEFT JOIN employee_history_cost_labels hcl ON hcl.id = hc.label_id
WHERE hc.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)