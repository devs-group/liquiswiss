SELECT
    hcl.id,
    hcl.name
FROM employee_history_cost_labels hcl
WHERE hcl.id = ?
  AND hcl.organisation_id = get_current_organisation(?)