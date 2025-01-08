SELECT
    hcl.id,
    hcl.name,
    COUNT(*) OVER() AS total_count
FROM employee_history_cost_labels hcl
WHERE hcl.organisation_id = get_current_user_organisation_id(?)
ORDER BY hcl.name
LIMIT ? OFFSET ?