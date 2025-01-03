UPDATE employee_history_cost_labels AS hcl
SET
    name = ?
WHERE
    hcl.id = ?
    AND hcl.organisation_id = get_current_organisation(?)
LIMIT 1;