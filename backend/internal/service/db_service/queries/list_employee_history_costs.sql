SELECT
    hc.id,
    COUNT(*) OVER() AS total_count
FROM employee_history_costs hc
JOIN employee_histories h ON h.id = hc.employee_history_id
JOIN employees e ON e.id = h.employee_id
LEFT JOIN employee_history_cost_labels hcl ON hcl.id = hc.label_id
WHERE hc.employee_history_id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
ORDER BY
    hcl.name,
    hc.distribution_type DESC
LIMIT ? OFFSET ?