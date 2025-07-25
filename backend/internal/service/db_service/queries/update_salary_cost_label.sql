UPDATE salary_cost_labels AS hcl
SET
    name = ?
WHERE
    hcl.id = ?
    AND hcl.organisation_id = get_current_user_organisation_id(?)
LIMIT 1;