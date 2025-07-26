SELECT
    hcl.id,
    hcl.name
FROM salary_cost_labels hcl
WHERE hcl.id = ?
  AND hcl.organisation_id = get_current_user_organisation_id(?)