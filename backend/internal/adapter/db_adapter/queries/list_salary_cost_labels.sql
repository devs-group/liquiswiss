SELECT
    hcl.id,
    hcl.name,
    COUNT(*) OVER() AS total_count
FROM salary_cost_labels hcl
WHERE hcl.organisation_id = get_current_user_organisation_id(?)
ORDER BY hcl.name
LIMIT ? OFFSET ?