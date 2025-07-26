SELECT
    hc.id,
    COUNT(*) OVER() AS total_count
FROM salary_costs hc
JOIN salaries h ON h.id = hc.salary_id
JOIN employees e ON e.id = h.employee_id
LEFT JOIN salary_cost_labels hcl ON hcl.id = hc.label_id
WHERE hc.salary_id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
ORDER BY
    hcl.name,
    hc.distribution_type DESC
LIMIT ? OFFSET ?